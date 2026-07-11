package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// Repository handles payment persistence and workflows.
type Repository struct {
	db         *sql.DB
	encryptKey []byte
	gateways   map[string]Gateway
}

// NewRepository creates a payment repository pre-registered with MockGateway.
// Use WithGateway to register additional production providers.
func NewRepository(db *sql.DB, appSecret string) *Repository {
	r := &Repository{
		db:         db,
		encryptKey: deriveKey(appSecret),
		gateways:   make(map[string]Gateway),
	}
	mock := &MockGateway{}
	r.gateways[mock.Provider()] = mock
	return r
}

// WithGateway registers an additional Gateway provider. Call this from the router
// when wiring up a real production provider (e.g. Midtrans, Xendit).
func (r *Repository) WithGateway(gw Gateway) *Repository {
	r.gateways[gw.Provider()] = gw
	return r
}

// LoadGatewayConfig returns branch payment gateway config.
func (r *Repository) LoadGatewayConfig(ctx context.Context, branchID, orgID uint64) (*GatewayConfig, error) {
	if !r.branchBelongsToOrg(ctx, branchID, orgID) {
		return nil, sql.ErrNoRows
	}
	const q = `
		SELECT id, branch_id, provider, is_active, config_json, created_at, updated_at
		FROM payment_gateway_configs
		WHERE branch_id = ?
		ORDER BY is_active DESC, id DESC
		LIMIT 1`
	cfg, err := r.scanGatewayConfig(r.db.QueryRowContext(ctx, q, branchID))
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// UpsertGatewayConfig creates or updates branch payment gateway config.
func (r *Repository) UpsertGatewayConfig(ctx context.Context, branchID, orgID uint64, req *UpsertGatewayConfigRequest) (*GatewayConfig, error) {
	if !r.branchBelongsToOrg(ctx, branchID, orgID) {
		return nil, sql.ErrNoRows
	}

	apiKeyEncrypted, err := encryptString(r.encryptKey, strings.TrimSpace(req.APIKey))
	if err != nil {
		return nil, err
	}
	webhookSecretEncrypted, err := encryptString(r.encryptKey, strings.TrimSpace(req.WebhookSecret))
	if err != nil {
		return nil, err
	}

	cfgPayload := map[string]string{
		"merchant_id":              strings.TrimSpace(req.MerchantID),
		"api_key_encrypted":        apiKeyEncrypted,
		"webhook_secret_encrypted": webhookSecretEncrypted,
	}
	cfgJSON, err := json.Marshal(cfgPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal gateway config: %w", err)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if isActive {
		if _, err := tx.ExecContext(ctx,
			`UPDATE payment_gateway_configs SET is_active = 0 WHERE branch_id = ?`, branchID,
		); err != nil {
			return nil, fmt.Errorf("deactivate old config: %w", err)
		}
	}

	var existingID uint64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM payment_gateway_configs WHERE branch_id = ? AND provider = ? LIMIT 1 FOR UPDATE`,
		branchID, req.Provider,
	).Scan(&existingID)

	switch err {
	case nil:
		_, err = tx.ExecContext(ctx,
			`UPDATE payment_gateway_configs SET is_active = ?, config_json = ? WHERE id = ?`,
			isActive, string(cfgJSON), existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("update gateway config: %w", err)
		}
	case sql.ErrNoRows:
		_, err = tx.ExecContext(ctx, `
			INSERT INTO payment_gateway_configs (branch_id, provider, is_active, config_json)
			VALUES (?, ?, ?, ?)
		`, branchID, req.Provider, isActive, string(cfgJSON))
		if err != nil {
			return nil, fmt.Errorf("insert gateway config: %w", err)
		}
	default:
		return nil, fmt.Errorf("load gateway config: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return r.LoadGatewayConfig(ctx, branchID, orgID)
}

// CreateInvoice creates or reuses a pending QRIS invoice for an order.
func (r *Repository) CreateInvoice(ctx context.Context, orderID, orgID uint64, req *CreateInvoiceRequest) (*Payment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var branchID uint64
	var orderCode string
	var totalAmount string
	var orderStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT o.branch_id, o.order_code, CAST(o.total_amount AS CHAR), o.status
		FROM orders o
		JOIN branches b ON b.id = o.branch_id
		WHERE o.id = ? AND b.organization_id = ?
		LIMIT 1
		FOR UPDATE
	`, orderID, orgID).Scan(&branchID, &orderCode, &totalAmount, &orderStatus); err != nil {
		return nil, err
	}
	if orderStatus == "cancelled" || orderStatus == "completed" {
		return nil, fmt.Errorf("pesanan tidak dapat dibayar pada status saat ini")
	}

	cfg, err := r.loadActiveGatewayConfigTx(ctx, tx, branchID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("konfigurasi gateway QRIS belum aktif untuk cabang ini")
		}
		return nil, err
	}

	gw, ok := r.gateways[cfg.Provider]
	if !ok {
		return nil, fmt.Errorf("provider gateway '%s' tidak dikenali; daftarkan provider sebelum digunakan", cfg.Provider)
	}

	existing, err := r.loadLatestOrderPaymentTx(ctx, tx, orderID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		switch existing.Status {
		case "paid":
			return nil, fmt.Errorf("pesanan ini sudah dibayar")
		case "pending":
			if existing.ExpiryAt == nil || existing.ExpiryAt.After(time.Now()) {
				if err := tx.Commit(); err != nil {
					return nil, fmt.Errorf("commit tx: %w", err)
				}
				return existing, nil
			}
		}
	}

	expiryMinutes := 15
	if req != nil && req.ExpiryMinutes >= 5 && req.ExpiryMinutes <= 120 {
		expiryMinutes = req.ExpiryMinutes
	}
	now := time.Now()
	expiryAt := now.Add(time.Duration(expiryMinutes) * time.Minute)

	invoiceRes, err := gw.CreateInvoice(ctx, &GatewayInvoiceRequest{
		MerchantID: cfg.Config["merchant_id"],
		OrderCode:  orderCode,
		Amount:     totalAmount,
		ExpiryAt:   expiryAt,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuat invoice dari gateway: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO payments
		  (order_id, gateway_config_id, gateway_invoice_id, payment_method, amount, status, qr_code_url, qr_string, expiry_at)
		VALUES (?, ?, ?, 'qris', ?, 'pending', ?, ?, ?)
	`, orderID, cfg.ID, invoiceRes.InvoiceID, totalAmount, invoiceRes.QRCodeURL, invoiceRes.QRString, invoiceRes.ExpiryAt)
	if err != nil {
		if strings.Contains(err.Error(), "1062") {
			return nil, fmt.Errorf("invoice gateway sudah terdaftar")
		}
		return nil, fmt.Errorf("insert payment: %w", err)
	}
	paymentID, _ := res.LastInsertId()

	if orderStatus == "pending" {
		_, _ = tx.ExecContext(ctx, `UPDATE orders SET status = 'confirmed' WHERE id = ?`, orderID)
	}

	p, err := r.findPaymentByIDTx(ctx, tx, uint64(paymentID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return p, nil
}

// ListReconciliation returns payments for reconciliation monitoring.
func (r *Repository) ListReconciliation(
	ctx context.Context,
	orgID uint64,
	branchID *uint64,
	status *string,
	dateFrom *time.Time,
	dateTo *time.Time,
) ([]Payment, error) {
	q := `
		SELECT p.id, p.order_id, p.gateway_config_id, p.gateway_invoice_id, p.payment_method,
		       CAST(p.amount AS CHAR), p.status, p.paid_at, p.qr_code_url, p.qr_string,
		       p.expiry_at, p.created_at, p.updated_at
		FROM payments p
		JOIN orders o ON o.id = p.order_id
		JOIN branches b ON b.id = o.branch_id
		WHERE b.organization_id = ?`
	args := []any{orgID}

	if branchID != nil {
		q += " AND o.branch_id = ?"
		args = append(args, *branchID)
	}
	if status != nil {
		q += " AND p.status = ?"
		args = append(args, *status)
	}
	if dateFrom != nil {
		q += " AND p.created_at >= ?"
		args = append(args, dateFrom.Format(time.DateTime))
	}
	if dateTo != nil {
		q += " AND p.created_at <= ?"
		args = append(args, dateTo.Format(time.DateTime))
	}
	q += " ORDER BY p.created_at DESC LIMIT 200"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list reconciliation: %w", err)
	}
	defer rows.Close()

	var out []Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	if out == nil {
		out = []Payment{}
	}
	return out, rows.Err()
}

// ProcessWebhook applies a normalized webhook event to update payment status.
func (r *Repository) ProcessWebhook(ctx context.Context, branchID uint64, event *GatewayWebhookEvent, rawPayload []byte) (*Payment, error) {
	if _, ok := validPaymentStatuses[event.Status]; !ok {
		return nil, fmt.Errorf("status pembayaran tidak valid")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	p, err := r.findPaymentByInvoiceIDTx(ctx, tx, event.InvoiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice pembayaran tidak ditemukan")
		}
		return nil, err
	}

	var paymentBranchID uint64
	if err := tx.QueryRowContext(ctx, `
		SELECT o.branch_id
		FROM orders o
		WHERE o.id = ?
		LIMIT 1`, p.OrderID,
	).Scan(&paymentBranchID); err != nil {
		return nil, fmt.Errorf("load payment branch: %w", err)
	}
	if paymentBranchID != branchID {
		slog.Warn("payment webhook branch mismatch", "invoice_id", event.InvoiceID, "expected_branch_id", paymentBranchID, "received_branch_id", branchID)
		return nil, fmt.Errorf("branch webhook tidak cocok")
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO payment_webhook_logs (payment_id, gateway_event_id, event_type, raw_payload, is_processed)
		VALUES (?, ?, ?, ?, 0)
	`, p.ID, event.EventID, event.EventType, string(rawPayload)); err != nil {
		if strings.Contains(err.Error(), "1062") {
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			return p, nil
		}
		return nil, fmt.Errorf("insert webhook log: %w", err)
	}

	newStatus := event.Status
	if p.Status == "paid" {
		newStatus = "paid"
	}
	var paidAt any
	if newStatus == "paid" {
		if event.PaidAt != "" {
			parsed, parseErr := time.Parse(time.RFC3339, event.PaidAt)
			if parseErr == nil {
				paidAt = parsed
			}
		}
		if paidAt == nil {
			now := time.Now()
			paidAt = now
		}
	}

	if paidAt != nil {
		_, err = tx.ExecContext(ctx,
			`UPDATE payments SET status = ?, paid_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			newStatus, paidAt, p.ID,
		)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE payments SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			newStatus, p.ID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("update payment status: %w", err)
	}

	if newStatus == "paid" {
		_, _ = tx.ExecContext(ctx,
			`UPDATE orders SET status = 'confirmed' WHERE id = ? AND status = 'pending'`,
			p.OrderID,
		)
	}

	_, _ = tx.ExecContext(ctx, `
		UPDATE payment_webhook_logs
		SET is_processed = 1, processed_at = CURRENT_TIMESTAMP
		WHERE gateway_event_id = ?
	`, event.EventID)

	updated, err := r.findPaymentByIDTx(ctx, tx, p.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return updated, nil
}

// NormalizeWebhook loads the active gateway config for the branch, selects the
// appropriate Gateway, and delegates signature verification and payload parsing to it.
// Returns the normalized GatewayWebhookEvent or an error if validation fails.
func (r *Repository) NormalizeWebhook(ctx context.Context, branchID uint64, rawPayload []byte, signature string) (*GatewayWebhookEvent, error) {
	cfg, err := r.loadActiveGatewayConfig(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi gateway tidak ditemukan untuk cabang ini")
	}
	gw, ok := r.gateways[cfg.Provider]
	if !ok {
		return nil, fmt.Errorf("provider gateway '%s' tidak dikenali", cfg.Provider)
	}
	webhookSecret := strings.TrimSpace(cfg.Config["webhook_secret"])
	event, err := gw.NormalizeWebhook(rawPayload, signature, webhookSecret)
	if err != nil {
		slog.Warn("webhook normalization failed", "branch_id", branchID, "provider", cfg.Provider, "error", err)
		return nil, err
	}
	return event, nil
}

func (r *Repository) loadActiveGatewayConfig(ctx context.Context, branchID uint64) (*GatewayConfig, error) {
	const q = `
		SELECT id, branch_id, provider, is_active, config_json, created_at, updated_at
		FROM payment_gateway_configs
		WHERE branch_id = ? AND is_active = 1
		ORDER BY id DESC
		LIMIT 1`
	return r.scanGatewayConfig(r.db.QueryRowContext(ctx, q, branchID))
}

func (r *Repository) loadActiveGatewayConfigTx(ctx context.Context, tx *sql.Tx, branchID uint64) (*GatewayConfig, error) {
	const q = `
		SELECT id, branch_id, provider, is_active, config_json, created_at, updated_at
		FROM payment_gateway_configs
		WHERE branch_id = ? AND is_active = 1
		ORDER BY id DESC
		LIMIT 1
		FOR UPDATE`
	return r.scanGatewayConfig(tx.QueryRowContext(ctx, q, branchID))
}

func (r *Repository) scanGatewayConfig(row interface{ Scan(...any) error }) (*GatewayConfig, error) {
	var cfg GatewayConfig
	var raw string
	if err := row.Scan(&cfg.ID, &cfg.BranchID, &cfg.Provider, &cfg.IsActive, &raw, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
		return nil, err
	}
	payload := map[string]string{}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return nil, fmt.Errorf("decode config json: %w", err)
		}
	}
	apiKey, _ := decryptString(r.encryptKey, payload["api_key_encrypted"])
	webhookSecret, _ := decryptString(r.encryptKey, payload["webhook_secret_encrypted"])
	cfg.Config = map[string]string{
		"merchant_id":    payload["merchant_id"],
		"api_key":        apiKey,
		"webhook_secret": webhookSecret,
		"api_key_masked": maskSecret(apiKey),
		"webhook_masked": maskSecret(webhookSecret),
	}
	return &cfg, nil
}

func (r *Repository) loadLatestOrderPaymentTx(ctx context.Context, tx *sql.Tx, orderID uint64) (*Payment, error) {
	const q = `
		SELECT id, order_id, gateway_config_id, gateway_invoice_id, payment_method,
		       CAST(amount AS CHAR), status, paid_at, qr_code_url, qr_string,
		       expiry_at, created_at, updated_at
		FROM payments
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 1
		FOR UPDATE`
	return scanPayment(tx.QueryRowContext(ctx, q, orderID))
}

func (r *Repository) findPaymentByIDTx(ctx context.Context, tx *sql.Tx, id uint64) (*Payment, error) {
	const q = `
		SELECT id, order_id, gateway_config_id, gateway_invoice_id, payment_method,
		       CAST(amount AS CHAR), status, paid_at, qr_code_url, qr_string,
		       expiry_at, created_at, updated_at
		FROM payments
		WHERE id = ?
		LIMIT 1`
	return scanPayment(tx.QueryRowContext(ctx, q, id))
}

func (r *Repository) findPaymentByInvoiceIDTx(ctx context.Context, tx *sql.Tx, invoiceID string) (*Payment, error) {
	const q = `
		SELECT id, order_id, gateway_config_id, gateway_invoice_id, payment_method,
		       CAST(amount AS CHAR), status, paid_at, qr_code_url, qr_string,
		       expiry_at, created_at, updated_at
		FROM payments
		WHERE gateway_invoice_id = ?
		LIMIT 1
		FOR UPDATE`
	return scanPayment(tx.QueryRowContext(ctx, q, invoiceID))
}

func scanPayment(row interface{ Scan(...any) error }) (*Payment, error) {
	var p Payment
	var gatewayConfigID sql.NullInt64
	var paidAt sql.NullTime
	var qrCodeURL sql.NullString
	var qrString sql.NullString
	var expiryAt sql.NullTime
	if err := row.Scan(
		&p.ID, &p.OrderID, &gatewayConfigID, &p.GatewayInvoiceID, &p.PaymentMethod,
		&p.Amount, &p.Status, &paidAt, &qrCodeURL, &qrString,
		&expiryAt, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if gatewayConfigID.Valid {
		v := uint64(gatewayConfigID.Int64)
		p.GatewayConfigID = &v
	}
	if paidAt.Valid {
		p.PaidAt = &paidAt.Time
	}
	if qrCodeURL.Valid {
		v := qrCodeURL.String
		p.QRCodeURL = &v
	}
	if qrString.Valid {
		v := qrString.String
		p.QRString = &v
	}
	if expiryAt.Valid {
		p.ExpiryAt = &expiryAt.Time
	}
	return &p, nil
}

func (r *Repository) branchBelongsToOrg(ctx context.Context, branchID, orgID uint64) bool {
	var foundOrgID uint64
	err := r.db.QueryRowContext(ctx,
		`SELECT organization_id FROM branches WHERE id = ? LIMIT 1`, branchID,
	).Scan(&foundOrgID)
	return err == nil && foundOrgID == orgID
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
}

func parseUintPointer(raw string) *uint64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}

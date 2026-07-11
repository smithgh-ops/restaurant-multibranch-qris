package order

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/kds"
)

// Repository handles DB operations for orders.
type Repository struct {
	db      *sql.DB
	kdsRepo *kds.Repository
	kdsHub  *kds.Hub
}

// NewRepository creates a new order repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// WithKDS attaches optional KDS dependencies to order writes.
func (r *Repository) WithKDS(repo *kds.Repository, hub *kds.Hub) *Repository {
	r.kdsRepo = repo
	r.kdsHub = hub
	return r
}

// ── Scan helpers ──────────────────────────────────────────────────────────────

func scanOrder(row interface{ Scan(...any) error }) (*Order, error) {
	o := &Order{}
	var tableID sql.NullInt64
	var notes sql.NullString
	var createdBy sql.NullInt64
	err := row.Scan(
		&o.ID, &o.BranchID, &tableID, &o.OrderCode, &o.OrderType, &o.Status,
		&notes, &o.Subtotal, &o.TaxAmount, &o.ServiceCharge, &o.TotalAmount,
		&createdBy, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if tableID.Valid {
		v := uint64(tableID.Int64)
		o.TableID = &v
	}
	if notes.Valid {
		o.Notes = &notes.String
	}
	if createdBy.Valid {
		v := uint64(createdBy.Int64)
		o.CreatedBy = &v
	}
	return o, nil
}

func scanOrderItem(row interface{ Scan(...any) error }) (*OrderItem, error) {
	it := &OrderItem{}
	var notes sql.NullString
	err := row.Scan(&it.ID, &it.OrderID, &it.MenuItemID, &it.ItemName, &it.UnitPrice, &it.Quantity, &it.Subtotal, &notes)
	if err != nil {
		return nil, err
	}
	if notes.Valid {
		it.Notes = &notes.String
	}
	return it, nil
}

// ── Read ──────────────────────────────────────────────────────────────────────

// FindByID returns an order with its items, scoped to org via branch.
func (r *Repository) FindByID(ctx context.Context, id, orgID uint64) (*Order, error) {
	const q = `
		SELECT o.id, o.branch_id, o.table_id, o.order_code, o.order_type, o.status,
		       o.notes, CAST(o.subtotal AS CHAR), CAST(o.tax_amount AS CHAR),
		       CAST(o.service_charge AS CHAR), CAST(o.total_amount AS CHAR),
		       o.created_by, o.created_at, o.updated_at
		FROM orders o
		JOIN branches b ON b.id = o.branch_id
		WHERE o.id = ? AND b.organization_id = ?
		LIMIT 1`
	o, err := scanOrder(r.db.QueryRowContext(ctx, q, id, orgID))
	if err != nil {
		return nil, err
	}
	items, err := r.loadItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

// List returns orders filtered by branch and optional status.
func (r *Repository) List(ctx context.Context, orgID uint64, branchID *uint64, status *string) ([]Order, error) {
	q := `
		SELECT o.id, o.branch_id, o.table_id, o.order_code, o.order_type, o.status,
		       o.notes, CAST(o.subtotal AS CHAR), CAST(o.tax_amount AS CHAR),
		       CAST(o.service_charge AS CHAR), CAST(o.total_amount AS CHAR),
		       o.created_by, o.created_at, o.updated_at
		FROM orders o
		JOIN branches b ON b.id = o.branch_id
		WHERE b.organization_id = ?`
	args := []any{orgID}

	if branchID != nil {
		q += " AND o.branch_id = ?"
		args = append(args, *branchID)
	}
	if status != nil {
		q += " AND o.status = ?"
		args = append(args, *status)
	}
	q += " ORDER BY o.created_at DESC"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, *o)
	}
	if orders == nil {
		orders = []Order{}
	}
	return orders, rows.Err()
}

// loadItems loads all order items for a given order ID.
func (r *Repository) loadItems(ctx context.Context, orderID uint64) ([]OrderItem, error) {
	const q = `
		SELECT id, order_id, menu_item_id, item_name,
		       CAST(unit_price AS CHAR), quantity, CAST(subtotal AS CHAR), notes
		FROM order_items WHERE order_id = ?`
	rows, err := r.db.QueryContext(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("load items: %w", err)
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		it, err := scanOrderItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, *it)
	}
	if items == nil {
		items = []OrderItem{}
	}
	return items, rows.Err()
}

// ── Write ─────────────────────────────────────────────────────────────────────

// Create inserts a new order with its items inside a transaction.
func (r *Repository) Create(ctx context.Context, orgID uint64, req *CreateOrderRequest, createdByUserID *uint64) (*Order, error) {
	// Verify branch belongs to org
	var branchOrgID uint64
	if err := r.db.QueryRowContext(ctx,
		`SELECT organization_id FROM branches WHERE id = ? AND is_active = 1 LIMIT 1`, req.BranchID,
	).Scan(&branchOrgID); err != nil || branchOrgID != orgID {
		return nil, fmt.Errorf("cabang tidak ditemukan atau tidak aktif")
	}

	// Validate table belongs to branch (if provided)
	if req.TableID != nil {
		var tableActive bool
		if err := r.db.QueryRowContext(ctx,
			`SELECT is_active FROM restaurant_tables WHERE id = ? AND branch_id = ? LIMIT 1`,
			*req.TableID, req.BranchID,
		).Scan(&tableActive); err != nil || !tableActive {
			return nil, fmt.Errorf("meja tidak ditemukan atau tidak aktif")
		}
	}

	// Resolve order type
	orderType := req.OrderType
	if orderType == "" {
		orderType = "dine_in"
	}

	// Fetch prices for all requested items and compute totals
	type itemLine struct {
		menuItemID uint64
		name       string
		unitPrice  float64
		quantity   uint16
		notes      *string
		subtotal   float64
	}
	lines := make([]itemLine, 0, len(req.Items))
	var subtotal float64

	for _, ri := range req.Items {
		var itemName string
		var basePrice float64
		var itemOrgID uint64

		if err := r.db.QueryRowContext(ctx,
			`SELECT name, CAST(base_price AS DECIMAL(12,2)), organization_id
			 FROM menu_items WHERE id = ? AND is_active = 1 LIMIT 1`, ri.MenuItemID,
		).Scan(&itemName, &basePrice, &itemOrgID); err != nil || itemOrgID != orgID {
			return nil, fmt.Errorf("item menu id=%d tidak ditemukan atau tidak aktif", ri.MenuItemID)
		}

		// Check if branch has a price override
		var priceOverride sql.NullFloat64
		_ = r.db.QueryRowContext(ctx,
			`SELECT CAST(price_override AS DECIMAL(12,2))
			 FROM menu_branch_settings WHERE menu_item_id = ? AND branch_id = ? LIMIT 1`,
			ri.MenuItemID, req.BranchID,
		).Scan(&priceOverride)

		unitPrice := basePrice
		if priceOverride.Valid {
			unitPrice = priceOverride.Float64
		}

		lineSubtotal := unitPrice * float64(ri.Quantity)
		subtotal += lineSubtotal
		lines = append(lines, itemLine{
			menuItemID: ri.MenuItemID,
			name:       itemName,
			unitPrice:  unitPrice,
			quantity:   ri.Quantity,
			notes:      ri.Notes,
			subtotal:   lineSubtotal,
		})
	}

	// Tax and service charge are fixed at 0 for now.
	// Per-branch tax/service-charge configuration is planned for Phase 6 (Laporan & Analitik).
	taxAmount := 0.0
	serviceCharge := 0.0
	totalAmount := subtotal + taxAmount + serviceCharge

	// Generate unique order code
	orderCode, err := r.generateOrderCode(ctx)
	if err != nil {
		return nil, err
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("create order: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	res, err := tx.ExecContext(ctx, `
		INSERT INTO orders
		  (branch_id, table_id, order_code, order_type, status, notes,
		   subtotal, tax_amount, service_charge, total_amount, created_by)
		VALUES (?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?, ?)`,
		req.BranchID, req.TableID, orderCode, orderType, req.Notes,
		subtotal, taxAmount, serviceCharge, totalAmount, createdByUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("create order: insert: %w", err)
	}
	orderID, _ := res.LastInsertId()

	for _, l := range lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_items
			  (order_id, menu_item_id, item_name, unit_price, quantity, subtotal, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			orderID, l.menuItemID, l.name, l.unitPrice, l.quantity, l.subtotal, l.notes,
		); err != nil {
			return nil, fmt.Errorf("create order: insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create order: commit: %w", err)
	}

	return r.FindByID(ctx, uint64(orderID), orgID)
}

// UpdateStatus changes the status of an order.
func (r *Repository) UpdateStatus(ctx context.Context, id, orgID uint64, status string) (*Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("update status: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var branchID uint64
	if err := tx.QueryRowContext(ctx, `
		SELECT o.branch_id
		FROM orders o
		JOIN branches b ON b.id = o.branch_id
		WHERE o.id = ? AND b.organization_id = ?
		LIMIT 1
		FOR UPDATE`,
		id, orgID,
	).Scan(&branchID); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("update status: load order: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE orders
		SET status = ?
		WHERE id = ?`,
		status, id,
	); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	createdTickets := make([]kds.KitchenTicket, 0)
	updatedTickets := make([]kds.KitchenTicket, 0)
	if r.kdsRepo != nil {
		switch status {
		case "confirmed":
			createdTickets, err = r.kdsRepo.EnsureTicketsForOrderTx(ctx, tx, id)
		case "cancelled":
			updatedTickets, err = r.kdsRepo.CancelTicketsForOrderTx(ctx, tx, id)
		}
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("update status: commit: %w", err)
	}

	o, err := r.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if r.kdsHub != nil {
		for i := range createdTickets {
			ticket := createdTickets[i]
			_ = r.kdsHub.Broadcast(branchID, kds.RealtimeEvent{Type: "ticket.created", Ticket: &ticket})
		}
		for i := range updatedTickets {
			ticket := updatedTickets[i]
			_ = r.kdsHub.Broadcast(branchID, kds.RealtimeEvent{Type: "ticket.updated", Ticket: &ticket})
		}
	}

	return o, nil
}

// generateOrderCode produces a unique order code like ORD-20240711-A1B2C3.
// It retries up to 5 times to avoid the rare collision when two orders are
// created simultaneously on the same date with the same 3-byte random suffix
// (probability ≈ 1/16^6 per attempt ≈ negligible in practice).
func (r *Repository) generateOrderCode(ctx context.Context) (string, error) {
	datePart := time.Now().Format("20060102")
	for range 5 {
		b := make([]byte, 3)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("generate order code: %w", err)
		}
		code := fmt.Sprintf("ORD-%s-%s", datePart, strings.ToUpper(hex.EncodeToString(b)))

		var exists int
		if err := r.db.QueryRowContext(ctx,
			`SELECT COUNT(1) FROM orders WHERE order_code = ?`, code,
		).Scan(&exists); err != nil {
			return "", fmt.Errorf("generate order code: check: %w", err)
		}
		if exists == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("gagal membuat kode pesanan unik")
}

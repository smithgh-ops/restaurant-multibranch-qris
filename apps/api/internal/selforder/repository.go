package selforder

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Repository handles DB operations for QR table tokens and public self-ordering.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a selforder repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── QR Token management (authenticated) ──────────────────────────────────────

// GenerateToken creates a new active QR token for the given table, deactivating
// any previous token for that table. The table must belong to orgID's branch.
func (r *Repository) GenerateToken(ctx context.Context, tableID, branchID, orgID uint64) (*QRTableToken, error) {
	if !r.branchBelongsToOrg(ctx, branchID, orgID) {
		return nil, sql.ErrNoRows
	}
	// Verify table belongs to branch.
	var foundBranchID uint64
	if err := r.db.QueryRowContext(ctx,
		`SELECT branch_id FROM restaurant_tables WHERE id = ? LIMIT 1`, tableID,
	).Scan(&foundBranchID); err != nil || foundBranchID != branchID {
		return nil, sql.ErrNoRows
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(raw)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Deactivate any existing tokens for this table.
	if _, err := tx.ExecContext(ctx,
		`UPDATE qr_table_tokens SET is_active = 0 WHERE table_id = ?`, tableID,
	); err != nil {
		return nil, fmt.Errorf("deactivate old tokens: %w", err)
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO qr_table_tokens (table_id, branch_id, token, is_active) VALUES (?, ?, ?, 1)`,
		tableID, branchID, token,
	)
	if err != nil {
		return nil, fmt.Errorf("insert token: %w", err)
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return r.findTokenByID(ctx, uint64(id))
}

// GetToken returns the current active QR token for a table, or sql.ErrNoRows.
func (r *Repository) GetToken(ctx context.Context, tableID, branchID, orgID uint64) (*QRTableToken, error) {
	if !r.branchBelongsToOrg(ctx, branchID, orgID) {
		return nil, sql.ErrNoRows
	}
	const q = `
		SELECT id, table_id, branch_id, token, is_active, expires_at, created_at, updated_at
		FROM qr_table_tokens
		WHERE table_id = ? AND branch_id = ? AND is_active = 1
		ORDER BY id DESC
		LIMIT 1`
	return scanToken(r.db.QueryRowContext(ctx, q, tableID, branchID))
}

// ── Public self-order API (unauthenticated) ───────────────────────────────────

// GetMenuByToken returns branch, table, and available menu for a valid QR token.
func (r *Repository) GetMenuByToken(ctx context.Context, token string) (*PublicMenuResponse, error) {
	var tokenRow QRTableToken
	var expiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, table_id, branch_id, token, is_active, expires_at, created_at, updated_at
		FROM qr_table_tokens
		WHERE token = ? AND is_active = 1
		LIMIT 1
	`, token).Scan(
		&tokenRow.ID, &tokenRow.TableID, &tokenRow.BranchID, &tokenRow.Token,
		&tokenRow.IsActive, &expiresAt, &tokenRow.CreatedAt, &tokenRow.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("load token: %w", err)
	}
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return nil, sql.ErrNoRows
	}

	// Branch info.
	var branch PublicBranch
	var address sql.NullString
	if err := r.db.QueryRowContext(ctx,
		`SELECT id, name, address FROM branches WHERE id = ? LIMIT 1`, tokenRow.BranchID,
	).Scan(&branch.ID, &branch.Name, &address); err != nil {
		return nil, fmt.Errorf("load branch: %w", err)
	}
	if address.Valid {
		branch.Address = address.String
	}

	// Table info.
	var table PublicTable
	if err := r.db.QueryRowContext(ctx,
		`SELECT id, table_number, capacity FROM restaurant_tables WHERE id = ? LIMIT 1`, tokenRow.TableID,
	).Scan(&table.ID, &table.TableNumber, &table.Capacity); err != nil {
		return nil, fmt.Errorf("load table: %w", err)
	}

	// Menu items with per-branch price override.
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			mi.id,
			mi.category_id,
			mi.name,
			mi.description,
			COALESCE(mbs.price_override, mi.base_price) AS effective_price,
			mi.image_url
		FROM menu_items mi
		JOIN menu_categories mc ON mc.id = mi.category_id
		LEFT JOIN menu_branch_settings mbs
			ON mbs.menu_item_id = mi.id AND mbs.branch_id = ?
		WHERE mi.is_active = 1
		  AND mc.is_active = 1
		  AND (mbs.id IS NULL OR mbs.is_available = 1)
		ORDER BY mc.sort_order, mc.name, mi.name
	`, tokenRow.BranchID)
	if err != nil {
		return nil, fmt.Errorf("list menu: %w", err)
	}
	defer rows.Close()

	catMap := map[uint64]*PublicMenuCategory{}
	catOrder := []uint64{}

	for rows.Next() {
		var item PublicMenuItem
		var desc, imageURL sql.NullString
		var price string
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Name, &desc, &price, &imageURL); err != nil {
			return nil, fmt.Errorf("scan menu item: %w", err)
		}
		item.Price = price
		if desc.Valid {
			item.Description = &desc.String
		}
		if imageURL.Valid && imageURL.String != "" {
			item.ImageURL = &imageURL.String
		}
		if _, exists := catMap[item.CategoryID]; !exists {
			catMap[item.CategoryID] = &PublicMenuCategory{ID: item.CategoryID}
			catOrder = append(catOrder, item.CategoryID)
		}
		catMap[item.CategoryID].Items = append(catMap[item.CategoryID].Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load category names.
	for catID := range catMap {
		var name string
		if err := r.db.QueryRowContext(ctx,
			`SELECT name FROM menu_categories WHERE id = ? LIMIT 1`, catID,
		).Scan(&name); err == nil {
			catMap[catID].Name = name
		}
	}

	categories := make([]PublicMenuCategory, 0, len(catOrder))
	for _, id := range catOrder {
		categories = append(categories, *catMap[id])
	}

	return &PublicMenuResponse{
		Branch:     branch,
		Table:      table,
		Categories: categories,
	}, nil
}

// CreateOrderByToken creates a new order from a self-order QR session (no auth).
// The table's branch must have the QR token active. Returns a slim order summary.
func (r *Repository) CreateOrderByToken(ctx context.Context, token string, req *CreateSelfOrderRequest) (*SelfOrderResponse, error) {
	var tableID, branchID uint64
	var expiresAt sql.NullTime
	var isActive bool
	err := r.db.QueryRowContext(ctx, `
		SELECT table_id, branch_id, is_active, expires_at
		FROM qr_table_tokens
		WHERE token = ?
		LIMIT 1
	`, token).Scan(&tableID, &branchID, &isActive, &expiresAt)
	if err == sql.ErrNoRows || !isActive {
		return nil, fmt.Errorf("token QR tidak valid atau tidak aktif")
	}
	if err != nil {
		return nil, fmt.Errorf("load token: %w", err)
	}
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token QR sudah kadaluarsa")
	}

	// Verify branch is active.
	var branchActive bool
	if err := r.db.QueryRowContext(ctx,
		`SELECT is_active FROM branches WHERE id = ? LIMIT 1`, branchID,
	).Scan(&branchActive); err != nil || !branchActive {
		return nil, fmt.Errorf("cabang tidak tersedia")
	}

	// Build order code.
	now := time.Now()
	randBytes := make([]byte, 3)
	rand.Read(randBytes) //nolint:errcheck
	orderCode := fmt.Sprintf("ORD-%s-%s", now.Format("20060102"), strings.ToUpper(hex.EncodeToString(randBytes)))

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Compute total from items.
	var totalAmount float64
	type resolvedItem struct {
		menuItemID uint64
		name       string
		unitPrice  float64
		quantity   uint8
		notes      *string
	}
	items := make([]resolvedItem, 0, len(req.Items))
	for _, it := range req.Items {
		var name string
		var basePrice, priceOverride sql.NullString
		var isActive bool
		if err := tx.QueryRowContext(ctx, `
			SELECT mi.name, mi.base_price, mbs.price_override, mi.is_active
			FROM menu_items mi
			LEFT JOIN menu_branch_settings mbs
				ON mbs.menu_item_id = mi.id AND mbs.branch_id = ?
			WHERE mi.id = ?
			LIMIT 1
		`, branchID, it.MenuItemID).Scan(&name, &basePrice, &priceOverride, &isActive); err == sql.ErrNoRows {
			return nil, fmt.Errorf("item menu %d tidak ditemukan", it.MenuItemID)
		} else if err != nil {
			return nil, fmt.Errorf("load menu item: %w", err)
		}
		if !isActive {
			return nil, fmt.Errorf("item menu '%s' tidak tersedia", name)
		}
		priceStr := basePrice.String
		if priceOverride.Valid && priceOverride.String != "" {
			priceStr = priceOverride.String
		}
		var price float64
		fmt.Sscanf(priceStr, "%f", &price)
		totalAmount += price * float64(it.Quantity)
		items = append(items, resolvedItem{
			menuItemID: it.MenuItemID,
			name:       name,
			unitPrice:  price,
			quantity:   it.Quantity,
			notes:      it.Notes,
		})
	}

	var notes *string
	if req.Notes != nil && strings.TrimSpace(*req.Notes) != "" {
		n := strings.TrimSpace(*req.Notes)
		notes = &n
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO orders
		  (branch_id, table_id, order_code, order_type, status, notes, subtotal, tax_amount, service_charge, total_amount)
		VALUES (?, ?, ?, 'dine_in', 'pending', ?, ?, 0, 0, ?)
	`, branchID, tableID, orderCode, notes, totalAmount, totalAmount)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}
	orderID, _ := res.LastInsertId()

	for _, it := range items {
		subtotal := it.unitPrice * float64(it.quantity)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, menu_item_id, item_name, unit_price, quantity, subtotal, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, orderID, it.menuItemID, it.name, it.unitPrice, it.quantity, subtotal, it.notes); err != nil {
			return nil, fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &SelfOrderResponse{
		OrderID:   uint64(orderID),
		OrderCode: orderCode,
		Status:    "pending",
		Total:     fmt.Sprintf("%.2f", totalAmount),
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (r *Repository) findTokenByID(ctx context.Context, id uint64) (*QRTableToken, error) {
	const q = `
		SELECT id, table_id, branch_id, token, is_active, expires_at, created_at, updated_at
		FROM qr_table_tokens WHERE id = ? LIMIT 1`
	return scanToken(r.db.QueryRowContext(ctx, q, id))
}

func scanToken(row interface{ Scan(...any) error }) (*QRTableToken, error) {
	t := &QRTableToken{}
	var expiresAt sql.NullTime
	if err := row.Scan(&t.ID, &t.TableID, &t.BranchID, &t.Token, &t.IsActive, &expiresAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	if expiresAt.Valid {
		t.ExpiresAt = &expiresAt.Time
	}
	return t, nil
}

func (r *Repository) branchBelongsToOrg(ctx context.Context, branchID, orgID uint64) bool {
	var foundOrgID uint64
	err := r.db.QueryRowContext(ctx,
		`SELECT organization_id FROM branches WHERE id = ? LIMIT 1`, branchID,
	).Scan(&foundOrgID)
	return err == nil && foundOrgID == orgID
}

package menu

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Repository handles database operations for menu categories and items.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new menu repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ── Categories ────────────────────────────────────────────────────────────────

func scanCategory(row interface{ Scan(...any) error }) (*Category, error) {
	c := &Category{}
	var desc sql.NullString
	err := row.Scan(&c.ID, &c.OrganizationID, &c.Name, &desc, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		c.Description = &desc.String
	}
	return c, nil
}

// ListCategories returns all categories for an organization.
func (r *Repository) ListCategories(ctx context.Context, orgID uint64) ([]Category, error) {
	const q = `
		SELECT id, organization_id, name, description, sort_order, is_active, created_at, updated_at
		FROM menu_categories
		WHERE organization_id = ?
		ORDER BY sort_order, name`
	rows, err := r.db.QueryContext(ctx, q, orgID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cats = append(cats, *c)
	}
	if cats == nil {
		cats = []Category{}
	}
	return cats, rows.Err()
}

// FindCategoryByID returns a category scoped to the organization.
func (r *Repository) FindCategoryByID(ctx context.Context, id, orgID uint64) (*Category, error) {
	const q = `
		SELECT id, organization_id, name, description, sort_order, is_active, created_at, updated_at
		FROM menu_categories
		WHERE id = ? AND organization_id = ? LIMIT 1`
	return scanCategory(r.db.QueryRowContext(ctx, q, id, orgID))
}

// CreateCategory inserts a new category.
func (r *Repository) CreateCategory(ctx context.Context, orgID uint64, req *CreateCategoryRequest) (*Category, error) {
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO menu_categories (organization_id, name, description, sort_order)
		VALUES (?, ?, ?, ?)`,
		orgID, req.Name, req.Description, sortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.FindCategoryByID(ctx, uint64(id), orgID)
}

// UpdateCategory applies partial updates to a category.
func (r *Repository) UpdateCategory(ctx context.Context, id, orgID uint64, req *UpdateCategoryRequest) (*Category, error) {
	clauses := []string{}
	args := []any{}

	if req.Name != nil {
		clauses = append(clauses, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		clauses = append(clauses, "description = ?")
		args = append(args, *req.Description)
	}
	if req.SortOrder != nil {
		clauses = append(clauses, "sort_order = ?")
		args = append(args, *req.SortOrder)
	}
	if req.IsActive != nil {
		clauses = append(clauses, "is_active = ?")
		args = append(args, *req.IsActive)
	}

	if len(clauses) == 0 {
		return r.FindCategoryByID(ctx, id, orgID)
	}

	query := "UPDATE menu_categories SET " + strings.Join(clauses, ", ") + " WHERE id = ? AND organization_id = ?"
	args = append(args, id, orgID)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return r.FindCategoryByID(ctx, id, orgID)
}

// ── Items ─────────────────────────────────────────────────────────────────────

func scanItem(row interface{ Scan(...any) error }) (*Item, error) {
	it := &Item{}
	var desc, imageURL sql.NullString
	err := row.Scan(
		&it.ID, &it.OrganizationID, &it.CategoryID, &it.Name, &desc,
		&it.BasePrice, &imageURL, &it.IsActive, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		it.Description = &desc.String
	}
	if imageURL.Valid {
		it.ImageURL = &imageURL.String
	}
	return it, nil
}

// ListItems returns menu items with optional filtering.
func (r *Repository) ListItems(ctx context.Context, orgID uint64, categoryID *uint64, branchID *uint64, activeOnly bool) ([]Item, error) {
	q := `
		SELECT id, organization_id, category_id, name, description,
		       CAST(base_price AS CHAR), image_url, is_active, created_at, updated_at
		FROM menu_items
		WHERE organization_id = ?`
	args := []any{orgID}

	if categoryID != nil {
		q += " AND category_id = ?"
		args = append(args, *categoryID)
	}
	if activeOnly {
		q += " AND is_active = 1"
	}
	// branch filter: only items that are available for that branch (or have no override)
	if branchID != nil {
		q += ` AND (
			NOT EXISTS (SELECT 1 FROM menu_branch_settings mbs WHERE mbs.menu_item_id = menu_items.id AND mbs.branch_id = ?)
			OR EXISTS  (SELECT 1 FROM menu_branch_settings mbs WHERE mbs.menu_item_id = menu_items.id AND mbs.branch_id = ? AND mbs.is_available = 1)
		)`
		args = append(args, *branchID, *branchID)
	}
	q += " ORDER BY name"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, *it)
	}
	if items == nil {
		items = []Item{}
	}
	return items, rows.Err()
}

// FindItemByID returns a menu item scoped to the organization.
func (r *Repository) FindItemByID(ctx context.Context, id, orgID uint64) (*Item, error) {
	const q = `
		SELECT id, organization_id, category_id, name, description,
		       CAST(base_price AS CHAR), image_url, is_active, created_at, updated_at
		FROM menu_items
		WHERE id = ? AND organization_id = ? LIMIT 1`
	return scanItem(r.db.QueryRowContext(ctx, q, id, orgID))
}

// CreateItem inserts a new menu item.
func (r *Repository) CreateItem(ctx context.Context, orgID uint64, req *CreateItemRequest) (*Item, error) {
	// Verify category belongs to org
	var catOrgID uint64
	err := r.db.QueryRowContext(ctx,
		`SELECT organization_id FROM menu_categories WHERE id = ? LIMIT 1`, req.CategoryID,
	).Scan(&catOrgID)
	if err == sql.ErrNoRows || catOrgID != orgID {
		return nil, fmt.Errorf("kategori tidak ditemukan dalam organisasi ini")
	}
	if err != nil {
		return nil, fmt.Errorf("create item: verify category: %w", err)
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO menu_items (organization_id, category_id, name, description, base_price, image_url)
		VALUES (?, ?, ?, ?, ?, ?)`,
		orgID, req.CategoryID, req.Name, req.Description, req.BasePrice, req.ImageURL,
	)
	if err != nil {
		return nil, fmt.Errorf("create item: insert: %w", err)
	}
	id, _ := res.LastInsertId()
	return r.FindItemByID(ctx, uint64(id), orgID)
}

// UpdateItem applies partial updates to a menu item.
func (r *Repository) UpdateItem(ctx context.Context, id, orgID uint64, req *UpdateItemRequest) (*Item, error) {
	clauses := []string{}
	args := []any{}

	if req.CategoryID != nil {
		// Verify category belongs to org
		var catOrgID uint64
		err := r.db.QueryRowContext(ctx,
			`SELECT organization_id FROM menu_categories WHERE id = ? LIMIT 1`, *req.CategoryID,
		).Scan(&catOrgID)
		if err == sql.ErrNoRows || catOrgID != orgID {
			return nil, fmt.Errorf("kategori tidak ditemukan dalam organisasi ini")
		}
		if err != nil {
			return nil, fmt.Errorf("update item: verify category: %w", err)
		}
		clauses = append(clauses, "category_id = ?")
		args = append(args, *req.CategoryID)
	}
	if req.Name != nil {
		clauses = append(clauses, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		clauses = append(clauses, "description = ?")
		args = append(args, *req.Description)
	}
	if req.BasePrice != nil {
		clauses = append(clauses, "base_price = ?")
		args = append(args, *req.BasePrice)
	}
	if req.ImageURL != nil {
		clauses = append(clauses, "image_url = ?")
		args = append(args, *req.ImageURL)
	}
	if req.IsActive != nil {
		clauses = append(clauses, "is_active = ?")
		args = append(args, *req.IsActive)
	}

	if len(clauses) == 0 {
		return r.FindItemByID(ctx, id, orgID)
	}

	query := "UPDATE menu_items SET " + strings.Join(clauses, ", ") + " WHERE id = ? AND organization_id = ?"
	args = append(args, id, orgID)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("update item: %w", err)
	}
	return r.FindItemByID(ctx, id, orgID)
}

// UpsertBranchSetting inserts or updates the per-branch setting for an item.
func (r *Repository) UpsertBranchSetting(ctx context.Context, itemID, orgID, branchID uint64, req *UpsertBranchSettingRequest) (*BranchSetting, error) {
	// Verify item belongs to org
	var itemOrgID uint64
	if err := r.db.QueryRowContext(ctx, `SELECT organization_id FROM menu_items WHERE id = ? LIMIT 1`, itemID).Scan(&itemOrgID); err != nil || itemOrgID != orgID {
		return nil, fmt.Errorf("item tidak ditemukan dalam organisasi ini")
	}
	// Verify branch belongs to org
	var branchOrgID uint64
	if err := r.db.QueryRowContext(ctx, `SELECT organization_id FROM branches WHERE id = ? LIMIT 1`, branchID).Scan(&branchOrgID); err != nil || branchOrgID != orgID {
		return nil, fmt.Errorf("cabang tidak ditemukan dalam organisasi ini")
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO menu_branch_settings (menu_item_id, branch_id, price_override, is_available)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE price_override = VALUES(price_override), is_available = VALUES(is_available)`,
		itemID, branchID, req.PriceOverride, req.IsAvailable,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert branch setting: %w", err)
	}

	bs := &BranchSetting{}
	var priceOverride sql.NullString
	err = r.db.QueryRowContext(ctx, `
		SELECT id, menu_item_id, branch_id, price_override, is_available, updated_at
		FROM menu_branch_settings
		WHERE menu_item_id = ? AND branch_id = ? LIMIT 1`,
		itemID, branchID,
	).Scan(&bs.ID, &bs.MenuItemID, &bs.BranchID, &priceOverride, &bs.IsAvailable, &bs.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert branch setting: read back: %w", err)
	}
	if priceOverride.Valid {
		bs.PriceOverride = &priceOverride.String
	}
	return bs, nil
}

// GetBranchSettings returns all branch settings for a menu item.
func (r *Repository) GetBranchSettings(ctx context.Context, itemID, orgID uint64) ([]BranchSetting, error) {
	// Verify item belongs to org
	var itemOrgID uint64
	if err := r.db.QueryRowContext(ctx, `SELECT organization_id FROM menu_items WHERE id = ? LIMIT 1`, itemID).Scan(&itemOrgID); err != nil || itemOrgID != orgID {
		return nil, fmt.Errorf("item tidak ditemukan dalam organisasi ini")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, menu_item_id, branch_id, price_override, is_available, updated_at
		FROM menu_branch_settings
		WHERE menu_item_id = ?`,
		itemID,
	)
	if err != nil {
		return nil, fmt.Errorf("get branch settings: %w", err)
	}
	defer rows.Close()

	var settings []BranchSetting
	for rows.Next() {
		bs := BranchSetting{}
		var priceOverride sql.NullString
		if err := rows.Scan(&bs.ID, &bs.MenuItemID, &bs.BranchID, &priceOverride, &bs.IsAvailable, &bs.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan branch setting: %w", err)
		}
		if priceOverride.Valid {
			bs.PriceOverride = &priceOverride.String
		}
		settings = append(settings, bs)
	}
	if settings == nil {
		settings = []BranchSetting{}
	}
	return settings, rows.Err()
}

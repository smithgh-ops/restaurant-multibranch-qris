package branch

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository handles database operations for branches.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new branch repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func scanBranch(row interface{ Scan(...any) error }) (*Branch, error) {
	b := &Branch{}
	var addr, phone sql.NullString
	err := row.Scan(
		&b.ID, &b.OrganizationID, &b.Name, &b.Slug,
		&addr, &phone, &b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if addr.Valid {
		b.Address = &addr.String
	}
	if phone.Valid {
		b.Phone = &phone.String
	}
	return b, nil
}

// List returns all branches for an organization.
func (r *Repository) List(ctx context.Context, orgID uint64) ([]Branch, error) {
	const q = `
		SELECT id, organization_id, name, slug, address, phone, is_active, created_at, updated_at
		FROM branches
		WHERE organization_id = ?
		ORDER BY name`
	rows, err := r.db.QueryContext(ctx, q, orgID)
	if err != nil {
		return nil, fmt.Errorf("branch list: %w", err)
	}
	defer rows.Close()

	var branches []Branch
	for rows.Next() {
		b, err := scanBranch(rows)
		if err != nil {
			return nil, fmt.Errorf("branch scan: %w", err)
		}
		branches = append(branches, *b)
	}
	if branches == nil {
		branches = []Branch{}
	}
	return branches, rows.Err()
}

// FindByID returns a branch by id, scoped to an organization.
func (r *Repository) FindByID(ctx context.Context, id, orgID uint64) (*Branch, error) {
	const q = `
		SELECT id, organization_id, name, slug, address, phone, is_active, created_at, updated_at
		FROM branches
		WHERE id = ? AND organization_id = ? LIMIT 1`
	return scanBranch(r.db.QueryRowContext(ctx, q, id, orgID))
}

// Create inserts a new branch and returns it.
func (r *Repository) Create(ctx context.Context, orgID uint64, req *CreateBranchRequest) (*Branch, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("branch create: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	res, err := tx.ExecContext(ctx, `
		INSERT INTO branches (organization_id, name, slug, address, phone)
		VALUES (?, ?, ?, ?, ?)`,
		orgID, req.Name, req.Slug, req.Address, req.Phone,
	)
	if err != nil {
		return nil, fmt.Errorf("branch create: insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("branch create: last id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("branch create: commit: %w", err)
	}

	return r.FindByID(ctx, uint64(id), orgID)
}

// Update applies partial updates to a branch.
func (r *Repository) Update(ctx context.Context, id, orgID uint64, req *UpdateBranchRequest) (*Branch, error) {
	// Build dynamic SET clause
	setClauses := []string{}
	args := []any{}

	if req.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Slug != nil {
		setClauses = append(setClauses, "slug = ?")
		args = append(args, *req.Slug)
	}
	if req.Address != nil {
		setClauses = append(setClauses, "address = ?")
		args = append(args, *req.Address)
	}
	if req.Phone != nil {
		setClauses = append(setClauses, "phone = ?")
		args = append(args, *req.Phone)
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, "is_active = ?")
		args = append(args, *req.IsActive)
	}

	if len(setClauses) == 0 {
		return r.FindByID(ctx, id, orgID)
	}

	query := "UPDATE branches SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = ? AND organization_id = ?"
	args = append(args, id, orgID)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("branch update: %w", err)
	}

	return r.FindByID(ctx, id, orgID)
}

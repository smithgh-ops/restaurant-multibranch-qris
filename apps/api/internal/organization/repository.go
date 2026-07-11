package organization

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Repository handles database operations for organizations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new organization repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindByID returns the organization with the given id.
func (r *Repository) FindByID(ctx context.Context, id uint64) (*Organization, error) {
	const q = `
		SELECT id, name, slug, logo_url, is_active, created_at, updated_at
		FROM organizations
		WHERE id = ? LIMIT 1`
	org := &Organization{}
	var logoURL sql.NullString
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&org.ID, &org.Name, &org.Slug, &logoURL,
		&org.IsActive, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("organization find: %w", err)
	}
	if logoURL.Valid {
		org.LogoURL = &logoURL.String
	}
	return org, nil
}

// Update applies partial updates (name, slug) to an organization.
func (r *Repository) Update(ctx context.Context, id uint64, name, slug *string) (*Organization, error) {
	setClauses := []string{}
	args := []any{}

	if name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *name)
	}
	if slug != nil {
		setClauses = append(setClauses, "slug = ?")
		args = append(args, *slug)
	}
	if len(setClauses) == 0 {
		return r.FindByID(ctx, id)
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	q := fmt.Sprintf("UPDATE organizations SET %s WHERE id = ?",
		strings.Join(setClauses, ", "))
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("organization update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return r.FindByID(ctx, id)
}

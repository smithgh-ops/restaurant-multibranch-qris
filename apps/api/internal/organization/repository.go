package organization

import (
	"context"
	"database/sql"
	"fmt"
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

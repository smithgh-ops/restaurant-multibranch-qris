package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository handles all database operations for the auth module.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new auth repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindUserByEmail returns the user with matching email, or sql.ErrNoRows.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
		SELECT id, organization_id, name, email, password_hash, is_active
		FROM users
		WHERE email = ? LIMIT 1`
	u := &User{}
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.OrganizationID, &u.Name, &u.Email, &u.PasswordHash, &u.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindUserByID returns the user with matching id.
func (r *Repository) FindUserByID(ctx context.Context, id uint64) (*User, error) {
	const q = `
		SELECT id, organization_id, name, email, is_active
		FROM users
		WHERE id = ? LIMIT 1`
	u := &User{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.OrganizationID, &u.Name, &u.Email, &u.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindPasswordHashByID returns only the password_hash for the given user id.
func (r *Repository) FindPasswordHashByID(ctx context.Context, id uint64) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = ? LIMIT 1`, id).Scan(&hash)
	return hash, err
}

// UpdatePassword sets a new password_hash for the given user.
func (r *Repository) UpdatePassword(ctx context.Context, id uint64, newHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = NOW() WHERE id = ?`,
		newHash, id,
	)
	return err
}

// GetUserRoles returns all role assignments for a user.
func (r *Repository) GetUserRoles(ctx context.Context, userID uint64) ([]UserRole, error) {
	const q = `
		SELECT ubr.branch_id, ubr.role_id, ro.name
		FROM user_branch_roles ubr
		JOIN roles ro ON ro.id = ubr.role_id
		WHERE ubr.user_id = ?`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}
	defer rows.Close()

	var roles []UserRole
	for rows.Next() {
		var ur UserRole
		var branchID sql.NullInt64
		if err := rows.Scan(&branchID, &ur.RoleID, &ur.RoleName); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		if branchID.Valid {
			v := uint64(branchID.Int64)
			ur.BranchID = &v
		}
		roles = append(roles, ur)
	}
	return roles, rows.Err()
}

// UpdateLastLogin sets the last_login_at timestamp for a user.
func (r *Repository) UpdateLastLogin(ctx context.Context, userID uint64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = ? WHERE id = ?`,
		time.Now().UTC(), userID,
	)
	return err
}

// StoreRefreshToken inserts a new hashed refresh token record.
func (r *Repository) StoreRefreshToken(ctx context.Context, userID uint64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

// FindRefreshToken looks up a (non-revoked, non-expired) refresh token by its hash.
func (r *Repository) FindRefreshToken(ctx context.Context, hash string) (*RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = ? LIMIT 1`
	rt := &RefreshToken{}
	var revokedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, q, hash).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &revokedAt,
	)
	if err != nil {
		return nil, err
	}
	if revokedAt.Valid {
		rt.RevokedAt = &revokedAt.Time
	}
	return rt, nil
}

// RevokeRefreshToken marks a token as revoked by its hash.
func (r *Repository) RevokeRefreshToken(ctx context.Context, hash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ?`,
		time.Now().UTC(), hash,
	)
	return err
}

// RevokeAllUserRefreshTokens revokes every active refresh token for a user.
func (r *Repository) RevokeAllUserRefreshTokens(ctx context.Context, userID uint64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`,
		time.Now().UTC(), userID,
	)
	return err
}

// DeleteExpiredTokens purges tokens that have passed their expiry (housekeeping).
func (r *Repository) DeleteExpiredTokens(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < ?`, time.Now().UTC(),
	)
	return err
}

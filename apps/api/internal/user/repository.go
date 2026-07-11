package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Repository handles database operations for user management.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListRoles returns all available roles.
func (r *Repository) ListRoles(ctx context.Context) ([]Role, error) {
	const q = `SELECT id, name, description FROM roles ORDER BY id`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var ro Role
		var desc sql.NullString
		if err := rows.Scan(&ro.ID, &ro.Name, &desc); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		if desc.Valid {
			ro.Description = &desc.String
		}
		roles = append(roles, ro)
	}
	return roles, rows.Err()
}

// List returns all users in the organization with their roles.
// Optionally filtered by branch_id (only users with a role in that branch).
func (r *Repository) List(ctx context.Context, orgID uint64, branchID *uint64) ([]UserWithRoles, error) {
	q := `SELECT DISTINCT u.id, u.organization_id, u.name, u.email, u.is_active, u.created_at, u.updated_at
		FROM users u`
	args := []any{}

	if branchID != nil {
		q += ` JOIN user_branch_roles ubr ON ubr.user_id = u.id AND ubr.branch_id = ?`
		args = append(args, *branchID)
	}

	q += ` WHERE u.organization_id = ? ORDER BY u.name`
	args = append(args, orgID)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []UserWithRoles
	idSet := []uint64{}
	userMap := map[uint64]*UserWithRoles{}
	for rows.Next() {
		u := UserWithRoles{}
		if err := rows.Scan(&u.ID, &u.OrganizationID, &u.Name, &u.Email, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		u.Roles = []UserRole{}
		users = append(users, u)
		idSet = append(idSet, u.ID)
		userMap[u.ID] = &users[len(users)-1]
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(idSet) == 0 {
		return users, nil
	}

	// Load roles for all users in one query.
	placeholders := make([]string, len(idSet))
	roleArgs := make([]any, len(idSet))
	for i, id := range idSet {
		placeholders[i] = "?"
		roleArgs[i] = id
	}
	roleQ := fmt.Sprintf(`
		SELECT ubr.user_id, ubr.role_id, ro.name, ubr.branch_id
		FROM user_branch_roles ubr
		JOIN roles ro ON ro.id = ubr.role_id
		WHERE ubr.user_id IN (%s)`, strings.Join(placeholders, ","))

	roleRows, err := r.db.QueryContext(ctx, roleQ, roleArgs...)
	if err != nil {
		return nil, fmt.Errorf("list user roles: %w", err)
	}
	defer roleRows.Close()

	for roleRows.Next() {
		var userID uint64
		var ur UserRole
		var branchIDNull sql.NullInt64
		if err := roleRows.Scan(&userID, &ur.RoleID, &ur.RoleName, &branchIDNull); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		if branchIDNull.Valid {
			v := uint64(branchIDNull.Int64)
			ur.BranchID = &v
		}
		if u, ok := userMap[userID]; ok {
			u.Roles = append(u.Roles, ur)
		}
	}
	return users, roleRows.Err()
}

// FindByID returns a single user with their roles, scoped to the organization.
func (r *Repository) FindByID(ctx context.Context, id, orgID uint64) (*UserWithRoles, error) {
	const q = `
		SELECT id, organization_id, name, email, is_active, created_at, updated_at
		FROM users
		WHERE id = ? AND organization_id = ? LIMIT 1`
	u := &UserWithRoles{}
	err := r.db.QueryRowContext(ctx, q, id, orgID).Scan(
		&u.ID, &u.OrganizationID, &u.Name, &u.Email, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Roles = []UserRole{}

	const rq = `
		SELECT ubr.role_id, ro.name, ubr.branch_id
		FROM user_branch_roles ubr
		JOIN roles ro ON ro.id = ubr.role_id
		WHERE ubr.user_id = ?`
	rows, err := r.db.QueryContext(ctx, rq, id)
	if err != nil {
		return nil, fmt.Errorf("find user roles: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ur UserRole
		var branchIDNull sql.NullInt64
		if err := rows.Scan(&ur.RoleID, &ur.RoleName, &branchIDNull); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		if branchIDNull.Valid {
			v := uint64(branchIDNull.Int64)
			ur.BranchID = &v
		}
		u.Roles = append(u.Roles, ur)
	}
	return u, rows.Err()
}

// Create inserts a new user and their initial role assignments inside a transaction.
func (r *Repository) Create(ctx context.Context, orgID uint64, req *CreateUserRequest, passwordHash string) (*UserWithRoles, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("create user begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	res, err := tx.ExecContext(ctx,
		`INSERT INTO users (organization_id, name, email, password_hash, is_active) VALUES (?, ?, ?, ?, 1)`,
		orgID, req.Name, req.Email, passwordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	newID, _ := res.LastInsertId()

	for _, ra := range req.Roles {
		if ra.BranchID != nil {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO user_branch_roles (user_id, branch_id, role_id) VALUES (?, ?, ?)`,
				newID, *ra.BranchID, ra.RoleID,
			)
		} else {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO user_branch_roles (user_id, branch_id, role_id) VALUES (?, NULL, ?)`,
				newID, ra.RoleID,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("insert user role: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("create user commit: %w", err)
	}

	return r.FindByID(ctx, uint64(newID), orgID)
}

// Update applies partial updates to a user.
func (r *Repository) Update(ctx context.Context, id, orgID uint64, req *UpdateUserRequest) (*UserWithRoles, error) {
	setClauses := []string{}
	args := []any{}

	if req.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Email != nil {
		setClauses = append(setClauses, "email = ?")
		args = append(args, *req.Email)
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, "is_active = ?")
		args = append(args, *req.IsActive)
	}
	if len(setClauses) == 0 {
		return r.FindByID(ctx, id, orgID)
	}

	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now().UTC())
	args = append(args, id, orgID)

	q := fmt.Sprintf("UPDATE users SET %s WHERE id = ? AND organization_id = ?",
		strings.Join(setClauses, ", "))
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return r.FindByID(ctx, id, orgID)
}

// Deactivate marks a user as inactive (soft delete).
func (r *Repository) Deactivate(ctx context.Context, id, orgID uint64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active = 0, updated_at = ? WHERE id = ? AND organization_id = ?`,
		time.Now().UTC(), id, orgID,
	)
	if err != nil {
		return fmt.Errorf("deactivate user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetRoles replaces all role assignments for a user inside a transaction.
func (r *Repository) SetRoles(ctx context.Context, id, orgID uint64, roles []RoleAssignment) (*UserWithRoles, error) {
	// Verify user belongs to org.
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE id = ? AND organization_id = ?`, id, orgID).Scan(&count)
	if err != nil || count == 0 {
		return nil, sql.ErrNoRows
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("set roles begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_branch_roles WHERE user_id = ?`, id); err != nil {
		return nil, fmt.Errorf("delete old roles: %w", err)
	}

	for _, ra := range roles {
		if ra.BranchID != nil {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO user_branch_roles (user_id, branch_id, role_id) VALUES (?, ?, ?)`,
				id, *ra.BranchID, ra.RoleID,
			)
		} else {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO user_branch_roles (user_id, branch_id, role_id) VALUES (?, NULL, ?)`,
				id, ra.RoleID,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("insert role: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("set roles commit: %w", err)
	}

	return r.FindByID(ctx, id, orgID)
}

// IsOrgAdmin checks if a user has org_admin or super_admin role in the given organization.
func (r *Repository) IsOrgAdmin(ctx context.Context, userID, orgID uint64) (bool, error) {
	const q = `
		SELECT COUNT(*)
		FROM user_branch_roles ubr
		JOIN roles ro ON ro.id = ubr.role_id
		JOIN users u ON u.id = ubr.user_id
		WHERE ubr.user_id = ?
		  AND u.organization_id = ?
		  AND ro.name IN ('org_admin', 'super_admin')`
	var count int
	if err := r.db.QueryRowContext(ctx, q, userID, orgID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

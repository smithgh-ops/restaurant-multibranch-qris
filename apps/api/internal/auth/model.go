package auth

import "time"

// User is a slim representation of a user row for auth operations.
type User struct {
	ID             uint64  `json:"id"`
	OrganizationID uint64  `json:"organization_id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	PasswordHash   string  `json:"-"`
	IsActive       bool    `json:"is_active"`
	LastLoginAt    *string `json:"last_login_at,omitempty"`
}

// UserRole carries a single role assignment for a user.
type UserRole struct {
	RoleID   uint64  `json:"role_id"`
	RoleName string  `json:"role_name"`
	BranchID *uint64 `json:"branch_id,omitempty"` // nil = global
}

// RefreshToken is a stored (hashed) refresh token record.
type RefreshToken struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	RevokedAt *time.Time `json:"-"`
}

// LoginRequest is the payload for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenPair is returned on successful login or refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	RefreshToken string `json:"refresh_token,omitempty"`
}

// MeResponse is returned by GET /auth/me.
type MeResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID uint64     `json:"organization_id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Roles          []UserRole `json:"roles"`
}

// ChangePasswordRequest is the payload for PATCH /auth/me/password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required,min=6"`
}

package user

import "time"

// UserRole carries a single role assignment for a user.
type UserRole struct {
	RoleID   uint64  `json:"role_id"`
	RoleName string  `json:"role_name"`
	BranchID *uint64 `json:"branch_id,omitempty"` // nil = global/org-level
}

// UserWithRoles is a full user record with role assignments.
type UserWithRoles struct {
	ID             uint64     `json:"id"`
	OrganizationID uint64     `json:"organization_id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Roles          []UserRole `json:"roles"`
}

// Role is a role record from the roles table.
type Role struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// RoleAssignment represents a single role + optional branch assignment.
type RoleAssignment struct {
	RoleID   uint64  `json:"role_id"  binding:"required"`
	BranchID *uint64 `json:"branch_id"` // nil = global
}

// CreateUserRequest is the payload for POST /api/v1/users.
type CreateUserRequest struct {
	Name     string           `json:"name"     binding:"required"`
	Email    string           `json:"email"    binding:"required,email"`
	Password string           `json:"password" binding:"required,min=6"`
	Roles    []RoleAssignment `json:"roles"`
}

// UpdateUserRequest is the payload for PATCH /api/v1/users/:id.
type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	IsActive *bool   `json:"is_active"`
}

// UpdateRolesRequest is the payload for PUT /api/v1/users/:id/roles.
type UpdateRolesRequest struct {
	Roles []RoleAssignment `json:"roles" binding:"required"`
}

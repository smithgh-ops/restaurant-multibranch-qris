package branch

import "time"

// Branch represents a physical restaurant location.
type Branch struct {
	ID             uint64    `json:"id"`
	OrganizationID uint64    `json:"organization_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Address        *string   `json:"address,omitempty"`
	Phone          *string   `json:"phone,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateBranchRequest is the payload for POST /api/v1/branches.
type CreateBranchRequest struct {
	Name    string  `json:"name"    binding:"required,min=2,max=100"`
	Slug    string  `json:"slug"    binding:"required,min=2,max=100"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
}

// UpdateBranchRequest is the payload for PATCH /api/v1/branches/:branch_id.
type UpdateBranchRequest struct {
	Name     *string `json:"name"`
	Slug     *string `json:"slug"`
	Address  *string `json:"address"`
	Phone    *string `json:"phone"`
	IsActive *bool   `json:"is_active"`
}

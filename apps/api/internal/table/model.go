package table

import "time"

// DiningArea represents a seating zone within a branch.
type DiningArea struct {
	ID          uint64    `json:"id"`
	BranchID    uint64    `json:"branch_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Table represents a physical table inside a dining area.
type Table struct {
	ID           uint64    `json:"id"`
	DiningAreaID uint64    `json:"dining_area_id"`
	BranchID     uint64    `json:"branch_id"`
	TableNumber  string    `json:"table_number"`
	Capacity     uint8     `json:"capacity"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateDiningAreaRequest is the payload for POST .../dining-areas.
type CreateDiningAreaRequest struct {
	Name        string  `json:"name"        binding:"required,min=1,max=100"`
	Description *string `json:"description"`
}

// CreateTableRequest is the payload for POST .../tables.
type CreateTableRequest struct {
	DiningAreaID uint64 `json:"dining_area_id" binding:"required"`
	TableNumber  string `json:"table_number"   binding:"required,min=1,max=20"`
	Capacity     uint8  `json:"capacity"`
}

// UpdateTableRequest is the payload for PATCH .../tables/:id.
type UpdateTableRequest struct {
	TableNumber *string `json:"table_number"`
	Capacity    *uint8  `json:"capacity"`
	IsActive    *bool   `json:"is_active"`
}

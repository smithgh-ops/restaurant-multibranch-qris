package menu

import "time"

// Category is a menu category scoped to an organization.
type Category struct {
	ID             uint64    `json:"id"`
	OrganizationID uint64    `json:"organization_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	SortOrder      int       `json:"sort_order"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Item is a menu item scoped to an organization.
type Item struct {
	ID             uint64    `json:"id"`
	OrganizationID uint64    `json:"organization_id"`
	CategoryID     uint64    `json:"category_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	BasePrice      string    `json:"base_price"` // kept as string for exact decimal representation
	ImageURL       *string   `json:"image_url,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BranchSetting holds per-branch price and availability overrides for an item.
type BranchSetting struct {
	ID            uint64    `json:"id"`
	MenuItemID    uint64    `json:"menu_item_id"`
	BranchID      uint64    `json:"branch_id"`
	PriceOverride *string   `json:"price_override,omitempty"`
	IsAvailable   bool      `json:"is_available"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateCategoryRequest is the payload for POST /api/v1/menu/categories.
type CreateCategoryRequest struct {
	Name        string  `json:"name"        binding:"required,min=2,max=100"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

// UpdateCategoryRequest is the payload for PATCH /api/v1/menu/categories/:id.
type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

// CreateItemRequest is the payload for POST /api/v1/menu/items.
type CreateItemRequest struct {
	CategoryID  uint64  `json:"category_id"  binding:"required"`
	Name        string  `json:"name"         binding:"required,min=2,max=150"`
	Description *string `json:"description"`
	BasePrice   string  `json:"base_price"   binding:"required"`
	ImageURL    *string `json:"image_url"`
}

// UpdateItemRequest is the payload for PATCH /api/v1/menu/items/:id.
type UpdateItemRequest struct {
	CategoryID  *uint64 `json:"category_id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	BasePrice   *string `json:"base_price"`
	ImageURL    *string `json:"image_url"`
	IsActive    *bool   `json:"is_active"`
}

// UpsertBranchSettingRequest is the payload for PUT /api/v1/menu/items/:id/branches/:branch_id.
type UpsertBranchSettingRequest struct {
	PriceOverride *string `json:"price_override"`
	IsAvailable   bool    `json:"is_available"`
}

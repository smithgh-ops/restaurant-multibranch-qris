package selforder

import "time"

// QRTableToken represents a QR code token linked to a specific restaurant table.
type QRTableToken struct {
	ID        uint64     `json:"id"`
	TableID   uint64     `json:"table_id"`
	BranchID  uint64     `json:"branch_id"`
	Token     string     `json:"token"`
	IsActive  bool       `json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// PublicBranch is the branch info exposed to unauthenticated self-order sessions.
type PublicBranch struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address,omitempty"`
}

// PublicTable is the table info exposed to unauthenticated self-order sessions.
type PublicTable struct {
	ID          uint64 `json:"id"`
	TableNumber string `json:"table_number"`
	Capacity    uint8  `json:"capacity"`
}

// PublicMenuItem is a menu item visible to a self-order session.
type PublicMenuItem struct {
	ID          uint64  `json:"id"`
	CategoryID  uint64  `json:"category_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Price       string  `json:"price"` // effective price (override or base)
	ImageURL    *string `json:"image_url,omitempty"`
}

// PublicMenuCategory groups menu items for a self-order session.
type PublicMenuCategory struct {
	ID    uint64           `json:"id"`
	Name  string           `json:"name"`
	Items []PublicMenuItem `json:"items"`
}

// PublicMenuResponse is the full payload returned by GET /api/v1/public/table/:token.
type PublicMenuResponse struct {
	Branch     PublicBranch         `json:"branch"`
	Table      PublicTable          `json:"table"`
	Categories []PublicMenuCategory `json:"categories"`
}

// SelfOrderItemRequest is a single line in a self-order checkout.
type SelfOrderItemRequest struct {
	MenuItemID uint64  `json:"menu_item_id" binding:"required"`
	Quantity   uint8   `json:"quantity"     binding:"required,min=1"`
	Notes      *string `json:"notes"`
}

// CreateSelfOrderRequest is the payload for POST /api/v1/public/table/:token/orders.
type CreateSelfOrderRequest struct {
	Notes *string                `json:"notes"`
	Items []SelfOrderItemRequest `json:"items" binding:"required,min=1"`
}

// SelfOrderResponse is the slim response returned after a successful self-order.
type SelfOrderResponse struct {
	OrderID   uint64 `json:"order_id"`
	OrderCode string `json:"order_code"`
	Status    string `json:"status"`
	Total     string `json:"total_amount"`
}

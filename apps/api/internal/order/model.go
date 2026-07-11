package order

import "time"

// Order represents a customer order at a branch.
type Order struct {
	ID            uint64      `json:"id"`
	BranchID      uint64      `json:"branch_id"`
	TableID       *uint64     `json:"table_id,omitempty"`
	OrderCode     string      `json:"order_code"`
	OrderType     string      `json:"order_type"`
	Status        string      `json:"status"`
	Notes         *string     `json:"notes,omitempty"`
	Subtotal      string      `json:"subtotal"`
	TaxAmount     string      `json:"tax_amount"`
	ServiceCharge string      `json:"service_charge"`
	TotalAmount   string      `json:"total_amount"`
	CreatedBy     *uint64     `json:"created_by,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Items         []OrderItem `json:"items,omitempty"`
}

// OrderItem represents a line item inside an order.
type OrderItem struct {
	ID         uint64  `json:"id"`
	OrderID    uint64  `json:"order_id"`
	MenuItemID uint64  `json:"menu_item_id"`
	ItemName   string  `json:"item_name"`
	UnitPrice  string  `json:"unit_price"`
	Quantity   uint16  `json:"quantity"`
	Subtotal   string  `json:"subtotal"`
	Notes      *string `json:"notes,omitempty"`
}

// CreateOrderRequest is the payload for POST /api/v1/orders.
type CreateOrderRequest struct {
	BranchID  uint64               `json:"branch_id"  binding:"required"`
	TableID   *uint64              `json:"table_id"`
	OrderType string               `json:"order_type"`
	Notes     *string              `json:"notes"`
	Items     []CreateOrderItemReq `json:"items"      binding:"required,min=1,dive"`
}

// CreateOrderItemReq is a single line item inside CreateOrderRequest.
type CreateOrderItemReq struct {
	MenuItemID uint64  `json:"menu_item_id" binding:"required"`
	Quantity   uint16  `json:"quantity"     binding:"required,min=1"`
	Notes      *string `json:"notes"`
}

// UpdateOrderStatusRequest is the payload for PATCH /api/v1/orders/:id/status.
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// validStatuses contains the allowed order status transitions.
var validStatuses = map[string]struct{}{
	"pending":    {},
	"confirmed":  {},
	"preparing":  {},
	"ready":      {},
	"completed":  {},
	"cancelled":  {},
}

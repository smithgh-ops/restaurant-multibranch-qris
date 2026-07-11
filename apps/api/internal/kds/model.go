package kds

import "time"

// KitchenStation represents a kitchen processing station in a branch.
type KitchenStation struct {
	ID          uint64    `json:"id"`
	BranchID    uint64    `json:"branch_id"`
	Name        string    `json:"name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description *string   `json:"description,omitempty"`
}

// KitchenTicket represents a single kitchen work item derived from an order item.
type KitchenTicket struct {
	ID          uint64     `json:"id"`
	OrderID     uint64     `json:"order_id"`
	StationID   *uint64    `json:"station_id,omitempty"`
	StationName *string    `json:"station_name,omitempty"`
	OrderItemID uint64     `json:"order_item_id"`
	BranchID    uint64     `json:"branch_id"`
	TableID     *uint64    `json:"table_id,omitempty"`
	OrderCode   string     `json:"order_code"`
	ItemName    string     `json:"item_name"`
	Quantity    uint16     `json:"quantity"`
	Notes       *string    `json:"notes,omitempty"`
	Status      string     `json:"status"`
	Priority    uint8      `json:"priority"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TicketListFilter struct {
	StationID *uint64
	Status    *string
}

type CreateStationRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type RealtimeEvent struct {
	Type   string         `json:"type"`
	Ticket *KitchenTicket `json:"ticket,omitempty"`
}

var validTicketStatuses = map[string]struct{}{
	"queued":      {},
	"in_progress": {},
	"done":        {},
	"cancelled":   {},
}

package payment

import "time"

// GatewayConfig stores branch-level QRIS provider configuration.
type GatewayConfig struct {
	ID        uint64            `json:"id"`
	BranchID  uint64            `json:"branch_id"`
	Provider  string            `json:"provider"`
	IsActive  bool              `json:"is_active"`
	Config    map[string]string `json:"config"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Payment represents a QRIS payment transaction.
type Payment struct {
	ID               uint64     `json:"id"`
	OrderID          uint64     `json:"order_id"`
	GatewayConfigID  *uint64    `json:"gateway_config_id,omitempty"`
	GatewayInvoiceID string     `json:"gateway_invoice_id"`
	PaymentMethod    string     `json:"payment_method"`
	Amount           string     `json:"amount"`
	Status           string     `json:"status"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	QRCodeURL        *string    `json:"qr_code_url,omitempty"`
	QRString         *string    `json:"qr_string,omitempty"`
	ExpiryAt         *time.Time `json:"expiry_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// UpsertGatewayConfigRequest is payload for branch gateway config.
type UpsertGatewayConfigRequest struct {
	Provider      string `json:"provider" binding:"required"`
	IsActive      *bool  `json:"is_active"`
	MerchantID    string `json:"merchant_id" binding:"required"`
	APIKey        string `json:"api_key"`
	WebhookSecret string `json:"webhook_secret"`
}

// CreateInvoiceRequest optionally customizes invoice expiry.
type CreateInvoiceRequest struct {
	ExpiryMinutes int `json:"expiry_minutes"`
}

// WebhookRequest is the legacy normalized mock QRIS webhook payload.
// Kept for backwards-compatibility with existing API documentation.
// New webhook flows go through Gateway.NormalizeWebhook instead.
type WebhookRequest struct {
	Provider  string `json:"provider" binding:"required"`
	BranchID  uint64 `json:"branch_id" binding:"required"`
	EventID   string `json:"event_id" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	InvoiceID string `json:"invoice_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
	PaidAt    string `json:"paid_at"`
}

var validPaymentStatuses = map[string]struct{}{
	"pending": {},
	"paid":    {},
	"failed":  {},
	"expired": {},
}

package payment

import (
	"context"
	"time"
)

// Gateway is the interface that all QRIS payment gateway providers must implement.
// To add a new provider (e.g. Midtrans, Xendit), create a new struct that satisfies
// this interface and register it via NewRepository(...).WithGateway(yourProvider).
type Gateway interface {
	// Provider returns the provider identifier stored in payment_gateway_configs.provider.
	Provider() string

	// CreateInvoice creates a new QRIS invoice with the given request and returns
	// the invoice details (QR string, QR code URL, assigned invoice ID, expiry time).
	CreateInvoice(ctx context.Context, req *GatewayInvoiceRequest) (*GatewayInvoiceResponse, error)

	// NormalizeWebhook validates the provider-specific webhook signature using the
	// given webhook secret, then parses rawPayload into a normalized WebhookEvent.
	// Returns an error if the signature is invalid or the payload cannot be parsed.
	NormalizeWebhook(rawPayload []byte, signature string, webhookSecret string) (*GatewayWebhookEvent, error)
}

// GatewayInvoiceRequest contains the data needed to create a QRIS invoice.
type GatewayInvoiceRequest struct {
	// MerchantID is the merchant identifier from the branch gateway config.
	MerchantID string
	// OrderCode is the human-readable order reference (e.g. "ORD-20250711-AB12CD").
	OrderCode string
	// Amount is the invoice amount formatted as a decimal string (e.g. "150000.00").
	Amount string
	// ExpiryAt is the desired expiry time of the invoice.
	ExpiryAt time.Time
}

// GatewayInvoiceResponse contains the result of a successful invoice creation.
type GatewayInvoiceResponse struct {
	// InvoiceID is the unique identifier assigned by the provider (stored as gateway_invoice_id).
	InvoiceID string
	// QRString is the raw QRIS data string that a payment app scans.
	QRString string
	// QRCodeURL is a URL to a rendered QR code image (may be empty if the provider
	// does not generate one; the application will fall back to QRString).
	QRCodeURL string
	// ExpiryAt is the actual expiry time confirmed by the provider.
	ExpiryAt time.Time
}

// GatewayWebhookEvent is a normalised representation of a provider webhook notification.
// Each provider's NormalizeWebhook implementation translates its native format into this struct.
type GatewayWebhookEvent struct {
	// EventID is the idempotency key for this webhook delivery (stored as gateway_event_id).
	EventID string
	// EventType is the provider-specific event name (e.g. "payment.success").
	EventType string
	// InvoiceID matches the InvoiceID returned by CreateInvoice (gateway_invoice_id).
	InvoiceID string
	// Status is one of: "pending", "paid", "failed", "expired".
	Status string
	// PaidAt is an optional RFC 3339 timestamp when the payment was confirmed (may be "").
	PaidAt string
}

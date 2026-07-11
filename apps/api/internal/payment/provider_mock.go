package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// MockGateway is a built-in QRIS provider for development and testing.
// It generates invoice IDs and QR strings locally without calling any external API.
// The webhook signature scheme uses HMAC-SHA256 over the raw JSON body.
//
// To replace this with a real provider (e.g. Midtrans or Xendit), implement the
// Gateway interface and register your provider with Repository.WithGateway().
type MockGateway struct{}

// Provider returns the provider identifier used in payment_gateway_configs.
func (m *MockGateway) Provider() string { return "qris_mock" }

// CreateInvoice generates a local QRIS invoice without calling an external API.
func (m *MockGateway) CreateInvoice(_ context.Context, req *GatewayInvoiceRequest) (*GatewayInvoiceResponse, error) {
	now := time.Now()
	invoiceID := fmt.Sprintf("INV-%s-%s",
		req.OrderCode,
		strings.ToUpper(now.Format("20060102150405")),
	)
	qrString := fmt.Sprintf("QRIS:%s:%s:%s", req.MerchantID, invoiceID, req.Amount)
	// NOTE: qrserver.com is a third-party QR code rendering service used here for
	// development/demo purposes. In production, generate QR images locally or via
	// a self-hosted service to avoid sending invoice data to an external party.
	qrCodeURL := "https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=" + url.QueryEscape(qrString)
	return &GatewayInvoiceResponse{
		InvoiceID: invoiceID,
		QRString:  qrString,
		QRCodeURL: qrCodeURL,
		ExpiryAt:  req.ExpiryAt,
	}, nil
}

// mockWebhookPayload is the raw JSON shape that MockGateway's webhook endpoint expects.
// Real providers send their own formats; each provider's NormalizeWebhook handles that.
type mockWebhookPayload struct {
	Provider  string `json:"provider"`
	BranchID  uint64 `json:"branch_id"`
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	InvoiceID string `json:"invoice_id"`
	Status    string `json:"status"`
	PaidAt    string `json:"paid_at"`
}

// NormalizeWebhook validates the HMAC-SHA256 signature (header X-QRIS-Signature) and
// parses the mock webhook payload into a GatewayWebhookEvent.
func (m *MockGateway) NormalizeWebhook(rawPayload []byte, signature string, webhookSecret string) (*GatewayWebhookEvent, error) {
	if webhookSecret == "" {
		return nil, fmt.Errorf("webhook secret tidak dikonfigurasi untuk provider ini")
	}

	// Validate HMAC-SHA256 signature.
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(rawPayload) //nolint:errcheck
	expected := hex.EncodeToString(mac.Sum(nil))

	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return nil, fmt.Errorf("internal: invalid expected signature hex")
	}
	sigBytes, err := hex.DecodeString(strings.TrimSpace(signature))
	if err != nil || !hmac.Equal(sigBytes, expectedBytes) {
		return nil, fmt.Errorf("signature webhook tidak valid")
	}

	var p mockWebhookPayload
	if err := json.Unmarshal(rawPayload, &p); err != nil {
		return nil, fmt.Errorf("payload webhook tidak dapat di-parse: %w", err)
	}
	if p.EventID == "" || p.InvoiceID == "" || p.Status == "" {
		return nil, fmt.Errorf("payload webhook tidak lengkap: event_id, invoice_id, dan status wajib ada")
	}

	return &GatewayWebhookEvent{
		EventID:   p.EventID,
		EventType: p.EventType,
		InvoiceID: p.InvoiceID,
		Status:    p.Status,
		PaidAt:    p.PaidAt,
	}, nil
}

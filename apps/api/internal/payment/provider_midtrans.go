package payment

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MidtransGateway implements the Gateway interface using the Midtrans Core API.
//
// To register this provider with the payment repository:
//
//	gw := payment.NewMidtransGateway(serverKey, isProduction)
//	paymentRepo.WithGateway(gw)
//
// Configure the branch payment gateway via the dashboard (provider = "midtrans").
// Store the Midtrans Server Key in the api_key and webhook_secret fields when
// calling PUT /api/v1/branches/:id/payment-gateway-config.
type MidtransGateway struct {
	serverKey    string
	isProduction bool
	httpClient   *http.Client
	// baseURLOverride is empty in normal use; set only in tests to point at an httptest.Server.
	baseURLOverride string
}

// NewMidtransGateway creates a Midtrans QRIS gateway provider.
//   - serverKey: Midtrans Server Key (e.g. "SB-Mid-server-...").
//   - isProduction: false → sandbox, true → production.
func NewMidtransGateway(serverKey string, isProduction bool) *MidtransGateway {
	return &MidtransGateway{
		serverKey:    serverKey,
		isProduction: isProduction,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// Provider returns the identifier stored in payment_gateway_configs.provider.
func (g *MidtransGateway) Provider() string { return "midtrans" }

func (g *MidtransGateway) baseURL() string {
	if g.baseURLOverride != "" {
		return g.baseURLOverride
	}
	if g.isProduction {
		return "https://api.midtrans.com"
	}
	return "https://api.sandbox.midtrans.com"
}

// CreateInvoice creates a QRIS charge via Midtrans Core API POST /v2/charge.
// The GatewayInvoiceRequest.Amount must be a decimal string (e.g. "150000.00").
func (g *MidtransGateway) CreateInvoice(ctx context.Context, req *GatewayInvoiceRequest) (*GatewayInvoiceResponse, error) {
	grossAmount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return nil, fmt.Errorf("midtrans: jumlah tidak valid %q: %w", req.Amount, err)
	}

	reqBody := midtransChargeRequest{
		PaymentType: "qris",
		TransactionDetails: midtransTransactionDetails{
			OrderID:     req.OrderCode,
			GrossAmount: int64(grossAmount),
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("midtrans: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.baseURL()+"/v2/charge", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("midtrans: build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	// Basic auth: base64(serverKey + ":") per Midtrans documentation.
	creds := base64.StdEncoding.EncodeToString([]byte(g.serverKey + ":"))
	httpReq.Header.Set("Authorization", "Basic "+creds)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans: HTTP error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("midtrans: read response: %w", err)
	}

	var result midtransChargeResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("midtrans: parse response: %w", err)
	}

	// Midtrans returns "201" for newly created transactions.
	if result.StatusCode != "201" && result.StatusCode != "200" {
		return nil, fmt.Errorf("midtrans: charge gagal (status_code=%s): %s",
			result.StatusCode, result.StatusMessage)
	}
	if result.TransactionID == "" {
		return nil, fmt.Errorf("midtrans: response tidak mengandung transaction_id")
	}

	// Locate QR code image URL from the actions array.
	var qrCodeURL string
	for _, action := range result.Actions {
		if action.Name == "generate-qr-code" {
			qrCodeURL = action.URL
			break
		}
	}

	// Prefer the expiry time returned by Midtrans; fall back to caller's value.
	expiryAt := req.ExpiryAt
	if result.ExpiryTime != "" {
		// Midtrans format: "2006-01-02 15:04:05" in WIB (UTC+7).
		// Go uses time.Local; adjust if the server runs in a different timezone.
		if parsed, parseErr := time.ParseInLocation("2006-01-02 15:04:05",
			result.ExpiryTime, time.Local); parseErr == nil {
			expiryAt = parsed
		}
	}

	return &GatewayInvoiceResponse{
		InvoiceID: result.TransactionID,
		QRString:  result.QRString,
		QRCodeURL: qrCodeURL,
		ExpiryAt:  expiryAt,
	}, nil
}

// NormalizeWebhook validates the Midtrans notification signature and returns a
// normalised GatewayWebhookEvent.
//
// Midtrans embeds the signature in the JSON body (field "signature_key") rather
// than in an HTTP header, so the `signature` parameter is ignored — the signature
// is read directly from rawPayload.
//
// webhookSecret is the Midtrans Server Key used to recompute the expected signature:
//
//	SHA-512( order_id + status_code + gross_amount + serverKey )
//
// If webhookSecret is empty, the gateway's own serverKey is used as the fallback.
func (g *MidtransGateway) NormalizeWebhook(rawPayload []byte, _ string, webhookSecret string) (*GatewayWebhookEvent, error) {
	var notif midtransNotification
	if err := json.Unmarshal(rawPayload, &notif); err != nil {
		return nil, fmt.Errorf("midtrans: parse notification: %w", err)
	}

	if notif.TransactionID == "" || notif.OrderID == "" {
		return nil, fmt.Errorf("midtrans: notification tidak lengkap: transaction_id dan order_id wajib ada")
	}

	// Use the configured webhook secret (Server Key) for validation.
	key := strings.TrimSpace(webhookSecret)
	if key == "" {
		key = g.serverKey
	}

	// Recompute expected signature.
	raw := notif.OrderID + notif.StatusCode + notif.GrossAmount + key
	sum := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(sum[:])

	if !strings.EqualFold(strings.TrimSpace(notif.SignatureKey), expected) {
		return nil, fmt.Errorf("midtrans: signature webhook tidak valid")
	}

	status := normalizeMidtransStatus(notif.TransactionStatus, notif.FraudStatus)

	// Parse settlement time as RFC 3339 for the caller.
	paidAt := ""
	if status == "paid" && notif.SettlementTime != "" {
		if parsed, parseErr := time.ParseInLocation("2006-01-02 15:04:05",
			notif.SettlementTime, time.Local); parseErr == nil {
			paidAt = parsed.UTC().Format(time.RFC3339)
		}
	}

	return &GatewayWebhookEvent{
		EventID:   notif.TransactionID,
		EventType: notif.TransactionStatus,
		// Midtrans uses transaction_id as both the event ID and the invoice ID.
		InvoiceID: notif.TransactionID,
		Status:    status,
		PaidAt:    paidAt,
	}, nil
}

// normalizeMidtransStatus maps Midtrans transaction_status values to the
// platform's canonical payment statuses: "pending", "paid", "failed", "expired".
func normalizeMidtransStatus(transactionStatus, fraudStatus string) string {
	switch transactionStatus {
	case "settlement", "capture":
		if strings.EqualFold(fraudStatus, "deny") {
			return "failed"
		}
		return "paid"
	case "pending":
		return "pending"
	case "cancel", "deny", "failure":
		return "failed"
	case "expire":
		return "expired"
	default:
		return "pending"
	}
}

// ── Midtrans API data types ───────────────────────────────────────────────────

type midtransChargeRequest struct {
	PaymentType        string                      `json:"payment_type"`
	TransactionDetails midtransTransactionDetails  `json:"transaction_details"`
}

type midtransTransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type midtransChargeResponse struct {
	StatusCode    string            `json:"status_code"`
	StatusMessage string            `json:"status_message"`
	TransactionID string            `json:"transaction_id"`
	OrderID       string            `json:"order_id"`
	GrossAmount   string            `json:"gross_amount"`
	PaymentType   string            `json:"payment_type"`
	ExpiryTime    string            `json:"expiry_time"`
	QRString      string            `json:"qr_string"`
	Actions       []midtransAction  `json:"actions"`
}

type midtransAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

// midtransNotification is the webhook notification body sent by Midtrans.
// Reference: https://docs.midtrans.com/docs/handling-notifications
type midtransNotification struct {
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	StatusCode        string `json:"status_code"`
	SettlementTime    string `json:"settlement_time"`
	SignatureKey      string `json:"signature_key"`
}

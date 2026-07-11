package payment

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestMidtransGateway creates a MidtransGateway pointed at a test server URL.
func newTestMidtransGateway(serverKey, baseURL string) *MidtransGateway {
	gw := NewMidtransGateway(serverKey, false)
	gw.baseURLOverride = baseURL
	return gw
}

func TestMidtransGateway_Provider(t *testing.T) {
	gw := NewMidtransGateway("SB-Mid-server-test", false)
	if got := gw.Provider(); got != "midtrans" {
		t.Fatalf("Provider() = %q, want %q", got, "midtrans")
	}
}

func TestMidtransGateway_BaseURL(t *testing.T) {
	sandbox := NewMidtransGateway("key", false)
	if sandbox.baseURL() != "https://api.sandbox.midtrans.com" {
		t.Errorf("sandbox baseURL = %q", sandbox.baseURL())
	}
	prod := NewMidtransGateway("key", true)
	if prod.baseURL() != "https://api.midtrans.com" {
		t.Errorf("production baseURL = %q", prod.baseURL())
	}
}

func TestMidtransGateway_CreateInvoice_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/charge" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		user, _, ok := r.BasicAuth()
		if !ok || user != "SB-Mid-server-test" {
			t.Errorf("unexpected or missing Basic auth user: %q", user)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(midtransChargeResponse{ //nolint:errcheck
			StatusCode:    "201",
			StatusMessage: "Created",
			TransactionID: "txn-abc123",
			OrderID:       "ORD-TEST-001",
			GrossAmount:   "150000.00",
			PaymentType:   "qris",
			ExpiryTime:    "2030-01-01 12:00:00",
			QRString:      "00020101021226...",
			Actions: []midtransAction{
				{Name: "generate-qr-code", Method: "GET", URL: "https://qr.example.com/abc"},
			},
		})
	}))
	defer srv.Close()

	gw := newTestMidtransGateway("SB-Mid-server-test", srv.URL)
	res, err := gw.CreateInvoice(context.Background(), &GatewayInvoiceRequest{
		MerchantID: "M001",
		OrderCode:  "ORD-TEST-001",
		Amount:     "150000.00",
		ExpiryAt:   time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("CreateInvoice error: %v", err)
	}
	if res.InvoiceID != "txn-abc123" {
		t.Errorf("InvoiceID = %q, want %q", res.InvoiceID, "txn-abc123")
	}
	if res.QRString != "00020101021226..." {
		t.Errorf("QRString = %q, want %q", res.QRString, "00020101021226...")
	}
	if res.QRCodeURL != "https://qr.example.com/abc" {
		t.Errorf("QRCodeURL = %q, want %q", res.QRCodeURL, "https://qr.example.com/abc")
	}
}

func TestMidtransGateway_CreateInvoice_InvalidAmount(t *testing.T) {
	gw := NewMidtransGateway("key", false)
	_, err := gw.CreateInvoice(context.Background(), &GatewayInvoiceRequest{
		OrderCode: "ORD-1",
		Amount:    "not-a-number",
	})
	if err == nil {
		t.Fatal("expected error for invalid amount, got nil")
	}
}

func TestMidtransGateway_CreateInvoice_GatewayError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(midtransChargeResponse{ //nolint:errcheck
			StatusCode:    "401",
			StatusMessage: "Access denied due to unauthorized transaction",
		})
	}))
	defer srv.Close()

	gw := newTestMidtransGateway("wrong-key", srv.URL)
	_, err := gw.CreateInvoice(context.Background(), &GatewayInvoiceRequest{
		OrderCode: "ORD-1",
		Amount:    "50000.00",
	})
	if err == nil {
		t.Fatal("expected error from gateway, got nil")
	}
}

func TestMidtransGateway_NormalizeWebhook_ValidSignature_Settlement(t *testing.T) {
	serverKey := "SB-Mid-server-test"
	notif := midtransNotification{
		TransactionID:     "txn-abc123",
		OrderID:           "ORD-TEST-001",
		GrossAmount:       "150000.00",
		PaymentType:       "qris",
		TransactionStatus: "settlement",
		FraudStatus:       "accept",
		StatusCode:        "200",
		SettlementTime:    "2030-01-01 10:00:00",
	}
	notif.SignatureKey = midtransSignature(notif.OrderID, notif.StatusCode, notif.GrossAmount, serverKey)

	payload, _ := json.Marshal(notif)
	gw := NewMidtransGateway(serverKey, false)
	event, err := gw.NormalizeWebhook(payload, "", serverKey)
	if err != nil {
		t.Fatalf("NormalizeWebhook error: %v", err)
	}
	if event.Status != "paid" {
		t.Errorf("Status = %q, want %q", event.Status, "paid")
	}
	if event.InvoiceID != "txn-abc123" {
		t.Errorf("InvoiceID = %q, want %q", event.InvoiceID, "txn-abc123")
	}
	if event.EventType != "settlement" {
		t.Errorf("EventType = %q, want %q", event.EventType, "settlement")
	}
	if event.PaidAt == "" {
		t.Error("PaidAt should be non-empty for settled transaction")
	}
}

func TestMidtransGateway_NormalizeWebhook_ValidSignature_Expire(t *testing.T) {
	serverKey := "SB-Mid-server-test"
	notif := midtransNotification{
		TransactionID:     "txn-xyz",
		OrderID:           "ORD-002",
		GrossAmount:       "25000.00",
		TransactionStatus: "expire",
		StatusCode:        "407",
	}
	notif.SignatureKey = midtransSignature(notif.OrderID, notif.StatusCode, notif.GrossAmount, serverKey)

	payload, _ := json.Marshal(notif)
	gw := NewMidtransGateway(serverKey, false)
	event, err := gw.NormalizeWebhook(payload, "", serverKey)
	if err != nil {
		t.Fatalf("NormalizeWebhook error: %v", err)
	}
	if event.Status != "expired" {
		t.Errorf("Status = %q, want %q", event.Status, "expired")
	}
}

func TestMidtransGateway_NormalizeWebhook_InvalidSignature(t *testing.T) {
	serverKey := "SB-Mid-server-test"
	notif := midtransNotification{
		TransactionID:     "txn-abc123",
		OrderID:           "ORD-TEST-001",
		GrossAmount:       "150000.00",
		StatusCode:        "200",
		TransactionStatus: "settlement",
		SignatureKey:      "invalidsignature",
	}
	payload, _ := json.Marshal(notif)

	gw := NewMidtransGateway(serverKey, false)
	_, err := gw.NormalizeWebhook(payload, "", serverKey)
	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
	}
}

func TestMidtransGateway_NormalizeWebhook_MalformedPayload(t *testing.T) {
	gw := NewMidtransGateway("key", false)
	_, err := gw.NormalizeWebhook([]byte(`{invalid json`), "", "key")
	if err == nil {
		t.Fatal("expected error for malformed payload, got nil")
	}
}

func TestMidtransGateway_NormalizeWebhook_MissingFields(t *testing.T) {
	gw := NewMidtransGateway("key", false)
	payload, _ := json.Marshal(midtransNotification{
		TransactionID: "",
		OrderID:       "",
	})
	_, err := gw.NormalizeWebhook(payload, "", "key")
	if err == nil {
		t.Fatal("expected error for missing required fields, got nil")
	}
}

func TestNormalizeMidtransStatus(t *testing.T) {
	cases := []struct {
		txStatus    string
		fraudStatus string
		want        string
	}{
		{"settlement", "accept", "paid"},
		{"settlement", "", "paid"},
		{"capture", "accept", "paid"},
		{"capture", "deny", "failed"},
		{"pending", "", "pending"},
		{"cancel", "", "failed"},
		{"deny", "", "failed"},
		{"failure", "", "failed"},
		{"expire", "", "expired"},
		{"unknown_status", "", "pending"},
	}
	for _, tc := range cases {
		got := normalizeMidtransStatus(tc.txStatus, tc.fraudStatus)
		if got != tc.want {
			t.Errorf("normalizeMidtransStatus(%q, %q) = %q, want %q",
				tc.txStatus, tc.fraudStatus, got, tc.want)
		}
	}
}

// midtransSignature computes the expected Midtrans webhook signature:
// SHA-512( orderID + statusCode + grossAmount + serverKey ).
func midtransSignature(orderID, statusCode, grossAmount, serverKey string) string {
	raw := orderID + statusCode + grossAmount + serverKey
	sum := sha512.Sum512([]byte(raw))
	return hex.EncodeToString(sum[:])
}

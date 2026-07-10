package auth_test

import (
	"testing"
	"time"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/auth"
)

const testSecret = "test-secret-key-for-unit-tests"

// TestGenerateAndParseAccessToken verifies the round-trip of access token generation and parsing.
func TestGenerateAndParseAccessToken(t *testing.T) {
	ttl := 15 * time.Minute
	token, err := auth.GenerateAccessToken(testSecret, ttl, 1, 42, "admin@example.com", "Admin Test")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := auth.ParseAccessToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("UserID: got %d, want 1", claims.UserID)
	}
	if claims.OrganizationID != 42 {
		t.Errorf("OrganizationID: got %d, want 42", claims.OrganizationID)
	}
	if claims.Email != "admin@example.com" {
		t.Errorf("Email: got %q, want %q", claims.Email, "admin@example.com")
	}
	if claims.Name != "Admin Test" {
		t.Errorf("Name: got %q, want %q", claims.Name, "Admin Test")
	}
}

// TestExpiredAccessToken ensures expired tokens are rejected.
func TestExpiredAccessToken(t *testing.T) {
	token, err := auth.GenerateAccessToken(testSecret, -1*time.Second, 1, 1, "x@x.com", "X")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	_, err = auth.ParseAccessToken(testSecret, token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

// TestWrongSecretRejected ensures tokens signed with a different secret are rejected.
func TestWrongSecretRejected(t *testing.T) {
	token, err := auth.GenerateAccessToken(testSecret, time.Minute, 1, 1, "x@x.com", "X")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	_, err = auth.ParseAccessToken("wrong-secret", token)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

// TestGenerateRefreshToken verifies that unique tokens are generated each time.
func TestGenerateRefreshToken(t *testing.T) {
	plain1, hash1, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	plain2, hash2, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken (2nd): %v", err)
	}

	if plain1 == plain2 {
		t.Error("expected unique plain tokens")
	}
	if hash1 == hash2 {
		t.Error("expected unique hashes")
	}

	// Hash must be deterministic for the same plain token
	if auth.HashToken(plain1) != hash1 {
		t.Error("HashToken(plain1) should equal hash1")
	}
}

// TestHashToken ensures the same input always produces the same hash.
func TestHashToken(t *testing.T) {
	h1 := auth.HashToken("hello")
	h2 := auth.HashToken("hello")
	if h1 != h2 {
		t.Errorf("HashToken is not deterministic: %q != %q", h1, h2)
	}

	h3 := auth.HashToken("world")
	if h1 == h3 {
		t.Error("different inputs should not produce the same hash")
	}
}

// TestMalformedTokenRejected ensures arbitrary strings are rejected.
func TestMalformedTokenRejected(t *testing.T) {
	_, err := auth.ParseAccessToken(testSecret, "not.a.jwt")
	if err == nil {
		t.Fatal("expected error for malformed token, got nil")
	}
}

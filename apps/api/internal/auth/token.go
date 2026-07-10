package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload embedded in access tokens.
type Claims struct {
	UserID         uint64 `json:"uid"`
	OrganizationID uint64 `json:"oid"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateAccessToken issues a short-lived signed JWT.
func GenerateAccessToken(secret string, ttl time.Duration, userID, orgID uint64, email, name string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          email,
		Name:           name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseAccessToken validates the JWT and returns its claims.
func ParseAccessToken(secret, tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// GenerateRefreshToken creates a cryptographically random opaque token (32 bytes hex).
func GenerateRefreshToken() (plain, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	plain = hex.EncodeToString(b)
	hashed = HashToken(plain)
	return plain, hashed, nil
}

// HashToken returns the SHA-256 hex digest of the opaque token for safe storage.
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

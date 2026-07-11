package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func computeSignature(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload) //nolint:errcheck
	return hex.EncodeToString(h.Sum(nil))
}

package app

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a cryptographically random 16-byte hex string (32 chars),
// used as the primary key for Vouchers created via CreateVoucherUseCase.
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

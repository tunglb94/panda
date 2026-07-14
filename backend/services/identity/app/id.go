package app

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a cryptographically random 16-byte hex string (32 chars),
// used as the primary key for entities this package creates (User,
// OTPChallenge). Mirrors identity/infrastructure/jwt's generateID — kept as
// a local copy rather than a shared export so this package doesn't reach
// into jwt's internals for an unrelated concern.
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

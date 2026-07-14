package localstore

import (
	"crypto/rand"
	"encoding/hex"
)

// randomToken returns a random 16-byte hex string, used to make saved
// filenames unique without pulling in a UUID dependency this module
// doesn't otherwise need.
func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

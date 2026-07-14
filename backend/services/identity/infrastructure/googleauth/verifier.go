// Package googleauth verifies Google Sign-In ID tokens server-side.
package googleauth

import "context"

// Identity is what a verified Google ID token tells us about the account.
type Identity struct {
	Sub           string // Google's stable per-account subject ID
	Email         string
	EmailVerified bool
	Name          string
}

// Verifier checks a Google-issued ID token and returns the identity it
// asserts. Implementations must reject tokens whose audience does not match
// the configured OAuth Client ID.
type Verifier interface {
	Verify(ctx context.Context, idToken string) (*Identity, error)
}

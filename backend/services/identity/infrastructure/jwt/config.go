// Package jwt provides HMAC-SHA256 (HS256) token generation and validation
// for the FAIRRIDE Identity service. The implementation uses only the Go
// standard library — no external JWT dependency is required.
package jwt

import (
	"time"

	domainerrors "github.com/fairride/shared/errors"
)

const (
	// minSecretLen is the minimum byte length accepted for HMAC secrets.
	// 32 bytes = 256 bits, matching the output size of SHA-256.
	minSecretLen = 32

	DefaultAccessTokenTTL  = 15 * time.Minute
	DefaultRefreshTokenTTL = 7 * 24 * time.Hour
)

// Config holds the secrets and TTLs used by TokenService.
// AccessSecret and RefreshSecret must be distinct high-entropy values.
// Set them from environment variables — never hard-code in source.
type Config struct {
	// AccessSecret is the HMAC key for signing access tokens.
	AccessSecret string
	// RefreshSecret is the HMAC key for signing refresh tokens.
	// Using a separate secret ensures a leaked access key cannot forge refresh tokens.
	RefreshSecret string
	// AccessTokenTTL is the validity window for access tokens. Default: 15 minutes.
	AccessTokenTTL time.Duration
	// RefreshTokenTTL is the validity window for refresh tokens. Default: 7 days.
	RefreshTokenTTL time.Duration
}

// DefaultConfig returns a Config with sensible TTL defaults.
// Secrets are intentionally left empty — callers MUST set them before calling Validate.
func DefaultConfig() Config {
	return Config{
		AccessTokenTTL:  DefaultAccessTokenTTL,
		RefreshTokenTTL: DefaultRefreshTokenTTL,
	}
}

// Validate returns an error if the Config is unusable.
func (c Config) Validate() error {
	if len(c.AccessSecret) < minSecretLen {
		return domainerrors.InvalidArgument("jwt: access secret must be at least 32 bytes")
	}
	if len(c.RefreshSecret) < minSecretLen {
		return domainerrors.InvalidArgument("jwt: refresh secret must be at least 32 bytes")
	}
	if c.AccessSecret == c.RefreshSecret {
		return domainerrors.InvalidArgument("jwt: access secret and refresh secret must differ")
	}
	if c.AccessTokenTTL <= 0 {
		return domainerrors.InvalidArgument("jwt: access token TTL must be positive")
	}
	if c.RefreshTokenTTL <= 0 {
		return domainerrors.InvalidArgument("jwt: refresh token TTL must be positive")
	}
	if c.RefreshTokenTTL <= c.AccessTokenTTL {
		return domainerrors.InvalidArgument("jwt: refresh token TTL must exceed access token TTL")
	}
	return nil
}

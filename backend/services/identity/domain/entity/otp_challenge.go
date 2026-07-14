package entity

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// MaxOTPAttempts is how many wrong codes a single challenge tolerates before
// it locks out (the caller must request a fresh OTP). Deliberately small —
// a 6-digit code has 1e6 combinations, so this is what actually prevents
// brute-forcing within the 5-minute TTL, not the code space itself.
const MaxOTPAttempts = 5

// OTPTTL is how long a requested code remains valid.
const OTPTTL = 5 * time.Minute

// OTPResendCooldown is the minimum time between two OTP requests for the
// same phone number — prevents a caller from hammering RequestOTP (and
// the downstream SMS/Zalo provider) in a tight loop.
const OTPResendCooldown = 60 * time.Second

// OTPChallenge is a single phone-verification attempt. The code itself is
// never persisted — only a salted SHA-256 hash — so a database leak does
// not expose usable OTP codes (Security requirement: "Không lưu OTP
// plaintext").
type OTPChallenge struct {
	ID          string
	PhoneNumber string
	CodeHash    string
	Purpose     string // "login" today; reserved for future purposes (e.g. "phone_link")
	ExpiresAt   time.Time
	Attempts    int
	Consumed    bool
	CreatedAt   time.Time
}

// NewOTPChallenge creates a challenge for phoneNumber holding the hash of
// code (computed via HashOTPCode). now is issuance time; ExpiresAt = now + OTPTTL.
func NewOTPChallenge(id, phoneNumber, code string, purpose string, now time.Time) (*OTPChallenge, error) {
	if id == "" {
		return nil, errors.InvalidArgument("otp challenge id must not be empty")
	}
	if strings.TrimSpace(phoneNumber) == "" {
		return nil, errors.InvalidArgument("phone number must not be empty")
	}
	if len(code) != 6 {
		return nil, errors.InvalidArgument("otp code must be 6 digits")
	}
	if purpose == "" {
		purpose = "login"
	}
	return &OTPChallenge{
		ID:          id,
		PhoneNumber: phoneNumber,
		CodeHash:    HashOTPCode(phoneNumber, code),
		Purpose:     purpose,
		ExpiresAt:   now.Add(OTPTTL),
		Attempts:    0,
		Consumed:    false,
		CreatedAt:   now,
	}, nil
}

// ReconstituteOTPChallenge rebuilds an OTPChallenge from a persistence record. No validation.
func ReconstituteOTPChallenge(id, phoneNumber, codeHash, purpose string, expiresAt time.Time, attempts int, consumed bool, createdAt time.Time) *OTPChallenge {
	return &OTPChallenge{
		ID:          id,
		PhoneNumber: phoneNumber,
		CodeHash:    codeHash,
		Purpose:     purpose,
		ExpiresAt:   expiresAt,
		Attempts:    attempts,
		Consumed:    consumed,
		CreatedAt:   createdAt,
	}
}

// HashOTPCode computes the salted hash stored in CodeHash. Salting with the
// phone number is cheap defense-in-depth: it stops a precomputed table of
// hash(code) from working across every phone number in the table at once.
func HashOTPCode(phoneNumber, code string) string {
	sum := sha256.Sum256([]byte(phoneNumber + ":" + code))
	return hex.EncodeToString(sum[:])
}

// CooldownRemaining reports how much longer the caller must wait before a
// new OTP request for this challenge's phone number is allowed. Zero (or
// negative) means the cooldown has elapsed.
func (c *OTPChallenge) CooldownRemaining(now time.Time) time.Duration {
	return c.CreatedAt.Add(OTPResendCooldown).Sub(now)
}

// Verify checks code against the stored hash. On any failure it returns a
// CodeUnauthenticated error and — for a plain mismatch — the caller must
// still persist the incremented Attempts (Verify mutates c.Attempts itself
// so a single Save afterwards captures it). Once Attempts reaches
// MaxOTPAttempts or the challenge is expired/consumed, every subsequent
// call fails without a real comparison.
func (c *OTPChallenge) Verify(code string, now time.Time) error {
	if c.Consumed {
		return errors.Unauthenticated("otp already used")
	}
	if now.After(c.ExpiresAt) {
		return errors.Unauthenticated("otp expired")
	}
	if c.Attempts >= MaxOTPAttempts {
		return errors.Unauthenticated("too many incorrect attempts — request a new code")
	}
	want := HashOTPCode(c.PhoneNumber, code)
	if subtle.ConstantTimeCompare([]byte(want), []byte(c.CodeHash)) != 1 {
		c.Attempts++
		return errors.Unauthenticated("incorrect code")
	}
	c.Consumed = true
	return nil
}

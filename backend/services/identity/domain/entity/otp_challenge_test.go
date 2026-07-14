package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/identity/domain/entity"
)

var otpTestNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func TestNewOTPChallenge_HashesCodeNotPlaintext(t *testing.T) {
	c, err := entity.NewOTPChallenge("id-1", "+84901234567", "123456", "login", otpTestNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.CodeHash == "123456" {
		t.Fatal("code hash must not equal the plaintext code")
	}
	if c.CodeHash != entity.HashOTPCode("+84901234567", "123456") {
		t.Fatal("code hash must match HashOTPCode(phone, code)")
	}
	if !c.ExpiresAt.Equal(otpTestNow.Add(entity.OTPTTL)) {
		t.Fatalf("expected expiry %v, got %v", otpTestNow.Add(entity.OTPTTL), c.ExpiresAt)
	}
}

func TestNewOTPChallenge_RejectsBadInput(t *testing.T) {
	if _, err := entity.NewOTPChallenge("", "+84901234567", "123456", "login", otpTestNow); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, err := entity.NewOTPChallenge("id-1", "", "123456", "login", otpTestNow); err == nil {
		t.Fatal("expected error for empty phone")
	}
	if _, err := entity.NewOTPChallenge("id-1", "+84901234567", "42", "login", otpTestNow); err == nil {
		t.Fatal("expected error for non-6-digit code")
	}
}

func TestOTPChallenge_Verify_CorrectCode(t *testing.T) {
	c, _ := entity.NewOTPChallenge("id-1", "+84901234567", "123456", "login", otpTestNow)
	if err := c.Verify("123456", otpTestNow); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !c.Consumed {
		t.Fatal("challenge should be marked consumed after a correct verify")
	}
}

func TestOTPChallenge_Verify_WrongCodeIncrementsAttempts(t *testing.T) {
	c, _ := entity.NewOTPChallenge("id-1", "+84901234567", "123456", "login", otpTestNow)
	if err := c.Verify("000000", otpTestNow); err == nil {
		t.Fatal("expected error for wrong code")
	}
	if c.Attempts != 1 {
		t.Fatalf("expected Attempts=1, got %d", c.Attempts)
	}
	if c.Consumed {
		t.Fatal("a wrong code must not consume the challenge")
	}
}

func TestOTPChallenge_Verify_LocksOutAfterMaxAttempts(t *testing.T) {
	c, _ := entity.NewOTPChallenge("id-1", "+84901234567", "123456", "login", otpTestNow)
	for i := 0; i < entity.MaxOTPAttempts; i++ {
		_ = c.Verify("000000", otpTestNow)
	}
	// Even the correct code must now be rejected — the challenge is locked.
	if err := c.Verify("123456", otpTestNow); err == nil {
		t.Fatal("expected lockout error after max attempts, got nil")
	}
}

func TestOTPChallenge_Verify_ExpiredRejected(t *testing.T) {
	c, _ := entity.NewOTPChallenge("id-1", "+84901234567", "123456", "login", otpTestNow)
	afterExpiry := otpTestNow.Add(entity.OTPTTL + time.Second)
	if err := c.Verify("123456", afterExpiry); err == nil {
		t.Fatal("expected expiry error")
	}
}

func TestOTPChallenge_Verify_AlreadyConsumedRejected(t *testing.T) {
	c, _ := entity.NewOTPChallenge("id-1", "+84901234567", "123456", "login", otpTestNow)
	if err := c.Verify("123456", otpTestNow); err != nil {
		t.Fatalf("unexpected error on first verify: %v", err)
	}
	if err := c.Verify("123456", otpTestNow); err == nil {
		t.Fatal("expected error verifying an already-consumed challenge")
	}
}

func TestOTPChallenge_CooldownRemaining(t *testing.T) {
	c, _ := entity.NewOTPChallenge("id-1", "+84901234567", "123456", "login", otpTestNow)
	if remaining := c.CooldownRemaining(otpTestNow); remaining <= 0 {
		t.Fatal("expected a positive cooldown immediately after creation")
	}
	after := otpTestNow.Add(entity.OTPResendCooldown + time.Second)
	if remaining := c.CooldownRemaining(after); remaining > 0 {
		t.Fatalf("expected cooldown elapsed, got %v remaining", remaining)
	}
}

package jwt

import (
	"strings"
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
)

// testConfig returns a valid Config for tests.
// Secrets are exactly 32 bytes of ASCII to satisfy the minimum length.
func testConfig() Config {
	return Config{
		AccessSecret:    "test-access-secret-32-bytes-long",
		RefreshSecret:   "test-refresh-secret-32-bytes-lon",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
}

func mustService(t *testing.T, cfg Config) *TokenService {
	t.Helper()
	svc, err := NewTokenService(cfg)
	if err != nil {
		t.Fatalf("NewTokenService: %v", err)
	}
	return svc
}

var testNow = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

// ─── Config.Validate ─────────────────────────────────────────────────────────

func TestConfig_DefaultTTLs(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.AccessTokenTTL != DefaultAccessTokenTTL {
		t.Errorf("AccessTokenTTL: got %v, want %v", cfg.AccessTokenTTL, DefaultAccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != DefaultRefreshTokenTTL {
		t.Errorf("RefreshTokenTTL: got %v, want %v", cfg.RefreshTokenTTL, DefaultRefreshTokenTTL)
	}
}

func TestConfig_Validate_Valid(t *testing.T) {
	if err := testConfig().Validate(); err != nil {
		t.Errorf("expected valid config, got: %v", err)
	}
}

func TestConfig_Validate_EmptyAccessSecret(t *testing.T) {
	cfg := testConfig()
	cfg.AccessSecret = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestConfig_Validate_ShortSecret(t *testing.T) {
	cfg := testConfig()
	cfg.AccessSecret = "short"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for short secret, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestConfig_Validate_SameSecrets(t *testing.T) {
	cfg := testConfig()
	cfg.RefreshSecret = cfg.AccessSecret
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when secrets are identical, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestConfig_Validate_ZeroAccessTTL(t *testing.T) {
	cfg := testConfig()
	cfg.AccessTokenTTL = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for zero access TTL, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestConfig_Validate_ZeroRefreshTTL(t *testing.T) {
	cfg := testConfig()
	cfg.RefreshTokenTTL = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for zero refresh TTL, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

func TestConfig_Validate_RefreshNotGreaterThanAccess(t *testing.T) {
	cfg := testConfig()
	cfg.RefreshTokenTTL = cfg.AccessTokenTTL // equal, not greater
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when refresh TTL ≤ access TTL, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

// ─── NewTokenService ─────────────────────────────────────────────────────────

func TestNewTokenService_ValidConfig(t *testing.T) {
	svc, err := NewTokenService(testConfig())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestNewTokenService_InvalidConfig(t *testing.T) {
	cfg := testConfig()
	cfg.AccessSecret = ""
	_, err := NewTokenService(cfg)
	if err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected CodeInvalidArgument, got %v", domainerrors.GetCode(err))
	}
}

// ─── GenerateAccessToken ─────────────────────────────────────────────────────

func TestGenerateAccessToken_ValidStructure(t *testing.T) {
	svc := mustService(t, testConfig())
	token, err := svc.GenerateAccessToken("user-1", "rider", "role-rider", testNow)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 dot-separated parts, got %d", len(parts))
	}
	for i, p := range parts {
		if p == "" {
			t.Errorf("part %d is empty", i)
		}
	}
}

func TestGenerateAccessToken_UniquePerCall(t *testing.T) {
	svc := mustService(t, testConfig())
	t1, _ := svc.GenerateAccessToken("user-1", "rider", "role-rider", testNow)
	t2, _ := svc.GenerateAccessToken("user-1", "rider", "role-rider", testNow)
	// JTI is random — same input must produce different tokens.
	if t1 == t2 {
		t.Error("expected unique tokens per call, got identical tokens")
	}
}

// ─── ValidateAccessToken ─────────────────────────────────────────────────────

func TestValidateAccessToken_Valid(t *testing.T) {
	svc := mustService(t, testConfig())
	token, err := svc.GenerateAccessToken("user-1", "rider", "role-rider", time.Now())
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken: %v", err)
	}
	if claims == nil {
		t.Fatal("expected non-nil claims")
	}
}

func TestValidateAccessToken_Claims(t *testing.T) {
	svc := mustService(t, testConfig())
	now := time.Now().UTC()
	token, _ := svc.GenerateAccessToken("user-42", "driver", "role-driver", now)
	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken: %v", err)
	}
	if claims.UserID != "user-42" {
		t.Errorf("UserID: got %q, want %q", claims.UserID, "user-42")
	}
	if claims.UserType != "driver" {
		t.Errorf("UserType: got %q, want %q", claims.UserType, "driver")
	}
	if claims.RoleID != "role-driver" {
		t.Errorf("RoleID: got %q, want %q", claims.RoleID, "role-driver")
	}
	if claims.TokenID == "" {
		t.Error("TokenID must not be empty")
	}
	if claims.ExpiresAt.IsZero() {
		t.Error("ExpiresAt must not be zero")
	}
	if claims.IssuedAt.IsZero() {
		t.Error("IssuedAt must not be zero")
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	svc := mustService(t, testConfig())
	// Issue the token anchored far enough in the past that exp is already passed.
	pastNow := time.Now().Add(-testConfig().AccessTokenTTL - time.Second)
	token, _ := svc.GenerateAccessToken("user-1", "rider", "role-1", pastNow)

	_, err := svc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeUnauthenticated) {
		t.Errorf("expected CodeUnauthenticated, got %v", domainerrors.GetCode(err))
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	svc := mustService(t, testConfig())
	token, _ := svc.GenerateAccessToken("user-1", "rider", "role-1", time.Now())

	// Build a second service with a different access secret.
	altCfg := testConfig()
	altCfg.AccessSecret = "different-access-secret-32-bytes"
	altSvc := mustService(t, altCfg)

	_, err := altSvc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeUnauthenticated) {
		t.Errorf("expected CodeUnauthenticated, got %v", domainerrors.GetCode(err))
	}
}

func TestValidateAccessToken_Malformed_Empty(t *testing.T) {
	svc := mustService(t, testConfig())
	_, err := svc.ValidateAccessToken("")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeUnauthenticated) {
		t.Errorf("expected CodeUnauthenticated, got %v", domainerrors.GetCode(err))
	}
}

func TestValidateAccessToken_Malformed_TwoParts(t *testing.T) {
	svc := mustService(t, testConfig())
	_, err := svc.ValidateAccessToken("header.payload")
	if err == nil {
		t.Fatal("expected error for two-part token, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeUnauthenticated) {
		t.Errorf("expected CodeUnauthenticated, got %v", domainerrors.GetCode(err))
	}
}

func TestValidateAccessToken_RefreshTokenRejected(t *testing.T) {
	svc := mustService(t, testConfig())
	// A refresh token signed with the refresh secret must not validate as an access token.
	rt, _ := svc.GenerateRefreshToken("user-1", "driver", "role-1", time.Now())
	_, err := svc.ValidateAccessToken(rt.Token)
	if err == nil {
		t.Fatal("expected error when using refresh token as access token, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeUnauthenticated) {
		t.Errorf("expected CodeUnauthenticated, got %v", domainerrors.GetCode(err))
	}
}

// ─── GenerateRefreshToken ─────────────────────────────────────────────────────

func TestGenerateRefreshToken_ValidStructure(t *testing.T) {
	svc := mustService(t, testConfig())
	rt, err := svc.GenerateRefreshToken("user-1", "driver", "role-1", testNow)
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if rt.Token == "" {
		t.Error("Token must not be empty")
	}
	if rt.TokenID == "" {
		t.Error("TokenID must not be empty")
	}
	if rt.UserID != "user-1" {
		t.Errorf("UserID: got %q, want %q", rt.UserID, "user-1")
	}
	if rt.Family == "" {
		t.Error("Family must not be empty")
	}
	wantExpiry := testNow.Add(testConfig().RefreshTokenTTL)
	if !rt.ExpiresAt.Equal(wantExpiry) {
		t.Errorf("ExpiresAt: got %v, want %v", rt.ExpiresAt, wantExpiry)
	}
}

func TestGenerateRefreshToken_UniquePerCall(t *testing.T) {
	svc := mustService(t, testConfig())
	r1, _ := svc.GenerateRefreshToken("user-1", "driver", "role-1", testNow)
	r2, _ := svc.GenerateRefreshToken("user-1", "driver", "role-1", testNow)
	if r1.Token == r2.Token {
		t.Error("expected unique tokens per call, got identical tokens")
	}
	if r1.TokenID == r2.TokenID {
		t.Error("expected unique TokenIDs per call")
	}
	if r1.Family == r2.Family {
		t.Error("expected unique Family IDs per call")
	}
}

// ─── ValidateRefreshToken ─────────────────────────────────────────────────────

func TestValidateRefreshToken_Valid(t *testing.T) {
	svc := mustService(t, testConfig())
	rt, err := svc.GenerateRefreshToken("user-1", "driver", "role-1", time.Now())
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	claims, err := svc.ValidateRefreshToken(rt.Token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken: %v", err)
	}
	if claims == nil {
		t.Fatal("expected non-nil claims")
	}
}

func TestValidateRefreshToken_Claims(t *testing.T) {
	svc := mustService(t, testConfig())
	now := time.Now().UTC()
	rt, _ := svc.GenerateRefreshToken("user-99", "driver", "role-1", now)
	claims, err := svc.ValidateRefreshToken(rt.Token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken: %v", err)
	}
	if claims.UserID != "user-99" {
		t.Errorf("UserID: got %q, want %q", claims.UserID, "user-99")
	}
	if claims.UserType != "driver" {
		t.Errorf("UserType: got %q, want %q", claims.UserType, "driver")
	}
	if claims.RoleID != "role-1" {
		t.Errorf("RoleID: got %q, want %q", claims.RoleID, "role-1")
	}
	if claims.TokenID != rt.TokenID {
		t.Errorf("TokenID: got %q, want %q", claims.TokenID, rt.TokenID)
	}
	if claims.Family != rt.Family {
		t.Errorf("Family: got %q, want %q", claims.Family, rt.Family)
	}
	if claims.ExpiresAt.IsZero() {
		t.Error("ExpiresAt must not be zero")
	}
	if claims.IssuedAt.IsZero() {
		t.Error("IssuedAt must not be zero")
	}
}

func TestValidateRefreshToken_Expired(t *testing.T) {
	svc := mustService(t, testConfig())
	pastNow := time.Now().Add(-testConfig().RefreshTokenTTL - time.Second)
	rt, _ := svc.GenerateRefreshToken("user-1", "driver", "role-1", pastNow)

	_, err := svc.ValidateRefreshToken(rt.Token)
	if err == nil {
		t.Fatal("expected error for expired refresh token, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeUnauthenticated) {
		t.Errorf("expected CodeUnauthenticated, got %v", domainerrors.GetCode(err))
	}
}

func TestValidateRefreshToken_WrongSecret(t *testing.T) {
	svc := mustService(t, testConfig())
	rt, _ := svc.GenerateRefreshToken("user-1", "driver", "role-1", time.Now())

	altCfg := testConfig()
	altCfg.RefreshSecret = "different-refresh-secret-32-byte"
	altSvc := mustService(t, altCfg)

	_, err := altSvc.ValidateRefreshToken(rt.Token)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeUnauthenticated) {
		t.Errorf("expected CodeUnauthenticated, got %v", domainerrors.GetCode(err))
	}
}

func TestValidateRefreshToken_AccessTokenRejected(t *testing.T) {
	svc := mustService(t, testConfig())
	// An access token (signed with access secret) must not validate as a refresh token.
	// The secrets differ, so signature will fail before kind check, which is fine.
	accessToken, _ := svc.GenerateAccessToken("user-1", "rider", "role-1", time.Now())
	_, err := svc.ValidateRefreshToken(accessToken)
	if err == nil {
		t.Fatal("expected error when using access token as refresh token, got nil")
	}
	if !domainerrors.IsCode(err, domainerrors.CodeUnauthenticated) {
		t.Errorf("expected CodeUnauthenticated, got %v", domainerrors.GetCode(err))
	}
}

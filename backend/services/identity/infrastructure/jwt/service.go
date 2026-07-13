package jwt

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	domainerrors "github.com/fairride/shared/errors"
)

// ─── Public types ─────────────────────────────────────────────────────────────

// AccessClaims holds the verified claims extracted from a valid access token.
type AccessClaims struct {
	UserID    string
	UserType  string // mirrors entity.UserType value
	RoleID    string
	TokenID   string
	ExpiresAt time.Time
	IssuedAt  time.Time
}

// RefreshClaims holds the verified claims extracted from a valid refresh token.
// UserType/RoleID are carried so a refresh can re-mint an access token
// without a database round-trip (they mirror whatever was passed to
// GenerateRefreshToken at issuance — the same values the original access
// token was minted with).
type RefreshClaims struct {
	UserID    string
	UserType  string
	RoleID    string
	TokenID   string
	Family    string // token family ID for future rotation tracking
	ExpiresAt time.Time
	IssuedAt  time.Time
}

// RefreshToken is the value object returned by GenerateRefreshToken.
// Token is the signed JWT string to hand to the client.
// TokenID, UserID, Family, and ExpiresAt are stored server-side for revocation.
type RefreshToken struct {
	Token     string
	TokenID   string
	UserID    string
	Family    string
	ExpiresAt time.Time
}

// ─── TokenService ─────────────────────────────────────────────────────────────

// TokenService generates and validates HS256 JWTs for the Identity service.
type TokenService struct {
	cfg           Config
	accessSecret  []byte
	refreshSecret []byte
}

// NewTokenService validates cfg and returns a ready-to-use TokenService.
// Returns CodeInvalidArgument if cfg fails validation.
func NewTokenService(cfg Config) (*TokenService, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &TokenService{
		cfg:           cfg,
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
	}, nil
}

// GenerateAccessToken creates a signed HS256 access token for the given user.
// now is the issuance time; expiry is now + Config.AccessTokenTTL.
func (s *TokenService) GenerateAccessToken(userID, userType, roleID string, now time.Time) (string, error) {
	jti, err := generateID()
	if err != nil {
		return "", err
	}
	p := accessPayload{
		Sub:    userID,
		Exp:    now.Add(s.cfg.AccessTokenTTL).Unix(),
		Iat:    now.Unix(),
		Jti:    jti,
		TknTyp: tokenKindAccess,
		UType:  userType,
		RID:    roleID,
	}
	return encodeToken(p, s.accessSecret)
}

// GenerateRefreshToken creates a signed HS256 refresh token for the given user.
// now is the issuance time; expiry is now + Config.RefreshTokenTTL. userType/
// roleID are embedded in the token so a later refresh can re-mint an access
// token directly from RefreshClaims, with no database lookup.
func (s *TokenService) GenerateRefreshToken(userID, userType, roleID string, now time.Time) (*RefreshToken, error) {
	jti, err := generateID()
	if err != nil {
		return nil, err
	}
	family, err := generateID()
	if err != nil {
		return nil, err
	}
	expiresAt := now.Add(s.cfg.RefreshTokenTTL)
	p := refreshPayload{
		Sub:    userID,
		Exp:    expiresAt.Unix(),
		Iat:    now.Unix(),
		Jti:    jti,
		TknTyp: tokenKindRefresh,
		Fam:    family,
		UType:  userType,
		RID:    roleID,
	}
	token, err := encodeToken(p, s.refreshSecret)
	if err != nil {
		return nil, err
	}
	return &RefreshToken{
		Token:     token,
		TokenID:   jti,
		UserID:    userID,
		Family:    family,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateAccessToken verifies signature, algorithm, token kind, and expiry.
// Returns CodeUnauthenticated on any failure — callers must not distinguish
// between expired and tampered tokens (information leak).
func (s *TokenService) ValidateAccessToken(token string) (*AccessClaims, error) {
	payload, err := verifyToken(token, s.accessSecret)
	if err != nil {
		return nil, err
	}
	var p accessPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, domainerrors.Unauthenticated("malformed access token")
	}
	if p.TknTyp != tokenKindAccess {
		return nil, domainerrors.Unauthenticated("token is not an access token")
	}
	if p.Exp <= time.Now().Unix() {
		return nil, domainerrors.Unauthenticated("access token has expired")
	}
	return &AccessClaims{
		UserID:    p.Sub,
		UserType:  p.UType,
		RoleID:    p.RID,
		TokenID:   p.Jti,
		ExpiresAt: time.Unix(p.Exp, 0).UTC(),
		IssuedAt:  time.Unix(p.Iat, 0).UTC(),
	}, nil
}

// ValidateRefreshToken verifies signature, algorithm, token kind, and expiry.
// Returns CodeUnauthenticated on any failure.
func (s *TokenService) ValidateRefreshToken(token string) (*RefreshClaims, error) {
	payload, err := verifyToken(token, s.refreshSecret)
	if err != nil {
		return nil, err
	}
	var p refreshPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, domainerrors.Unauthenticated("malformed refresh token")
	}
	if p.TknTyp != tokenKindRefresh {
		return nil, domainerrors.Unauthenticated("token is not a refresh token")
	}
	if p.Exp <= time.Now().Unix() {
		return nil, domainerrors.Unauthenticated("refresh token has expired")
	}
	return &RefreshClaims{
		UserID:    p.Sub,
		UserType:  p.UType,
		RoleID:    p.RID,
		TokenID:   p.Jti,
		Family:    p.Fam,
		ExpiresAt: time.Unix(p.Exp, 0).UTC(),
		IssuedAt:  time.Unix(p.Iat, 0).UTC(),
	}, nil
}

// ─── Internal JWT primitives ──────────────────────────────────────────────────

const (
	tokenKindAccess  = "access"
	tokenKindRefresh = "refresh"
)

// jwtHeader is marshaled as the JWT header segment.
type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// accessPayload is the JWT payload for access tokens.
type accessPayload struct {
	Sub    string `json:"sub"`   // subject = userID
	Exp    int64  `json:"exp"`   // expiry (Unix seconds)
	Iat    int64  `json:"iat"`   // issued-at (Unix seconds)
	Jti    string `json:"jti"`   // JWT ID (unique per token)
	TknTyp string `json:"tkt"`   // token kind discriminator
	UType  string `json:"utype"` // user type (rider, driver, …)
	RID    string `json:"rid"`   // role ID
}

// refreshPayload is the JWT payload for refresh tokens.
type refreshPayload struct {
	Sub    string `json:"sub"` // subject = userID
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
	Jti    string `json:"jti"`
	TknTyp string `json:"tkt"`   // token kind discriminator
	Fam    string `json:"fam"`   // token family ID for rotation
	UType  string `json:"utype"` // user type — so a refresh can re-mint an access token with no DB lookup
	RID    string `json:"rid"`   // role ID — same reason
}

// encodeToken serialises payload to JSON and produces a signed HS256 JWT string.
func encodeToken(payload any, secret []byte) (string, error) {
	hdrJSON, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", domainerrors.Internal("marshal jwt header").WithMeta("error", err.Error())
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", domainerrors.Internal("marshal jwt payload").WithMeta("error", err.Error())
	}

	hdrEnc := b64Encode(hdrJSON)
	payEnc := b64Encode(payloadJSON)
	sigInput := hdrEnc + "." + payEnc

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	sig := b64Encode(mac.Sum(nil))

	return sigInput + "." + sig, nil
}

// verifyToken splits the JWT, checks the HS256 signature using constant-time
// comparison, and returns the raw payload JSON. It does NOT check expiry.
func verifyToken(token string, secret []byte) ([]byte, error) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return nil, domainerrors.Unauthenticated("malformed token")
	}

	hdrJSON, err := b64Decode(parts[0])
	if err != nil {
		return nil, domainerrors.Unauthenticated("malformed token")
	}
	var hdr jwtHeader
	if err := json.Unmarshal(hdrJSON, &hdr); err != nil || hdr.Alg != "HS256" {
		return nil, domainerrors.Unauthenticated("unsupported token algorithm")
	}

	gotSig, err := b64Decode(parts[2])
	if err != nil {
		return nil, domainerrors.Unauthenticated("malformed token")
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(mac.Sum(nil), gotSig) {
		return nil, domainerrors.Unauthenticated("invalid token signature")
	}

	payloadJSON, err := b64Decode(parts[1])
	if err != nil {
		return nil, domainerrors.Unauthenticated("malformed token")
	}
	return payloadJSON, nil
}

// generateID returns a cryptographically random 16-byte hex string (32 chars).
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", domainerrors.Internal("generate token id").WithMeta("error", err.Error())
	}
	return hex.EncodeToString(b), nil
}

func b64Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func b64Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

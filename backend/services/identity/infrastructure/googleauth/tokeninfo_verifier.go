package googleauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	domainerrors "github.com/fairride/shared/errors"
)

// tokenInfoURL is Google's token-verification endpoint. It validates the
// JWT's signature/expiry itself and returns 4xx for anything invalid — using
// it (rather than fetching Google's JWKS and verifying the RS256 signature
// locally) avoids pulling in an RSA/JWKS library, matching this codebase's
// existing preference for stdlib-only crypto (see identity/infrastructure/jwt).
// Trade-off: this makes Verify a network call to Google per login instead of
// an offline check — acceptable at Panda's current scale; see plan's Known
// Gaps for the offline-JWKS alternative if this ever needs to change.
const tokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"

// TokenInfoVerifier implements Verifier via Google's tokeninfo endpoint.
type TokenInfoVerifier struct {
	clientID   string
	httpClient *http.Client
}

// NewTokenInfoVerifier constructs a TokenInfoVerifier. clientID is the
// OAuth 2.0 Web/Android/iOS Client ID configured in Google Cloud Console —
// every verified token's "aud" claim must match it exactly.
func NewTokenInfoVerifier(clientID string) *TokenInfoVerifier {
	return &TokenInfoVerifier{
		clientID:   clientID,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type tokenInfoResponse struct {
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Error         string `json:"error_description"`
}

func (v *TokenInfoVerifier) Verify(ctx context.Context, idToken string) (*Identity, error) {
	if idToken == "" {
		return nil, domainerrors.InvalidArgument("id_token is required")
	}
	if v.clientID == "" {
		return nil, domainerrors.Unavailable("google sign-in is not configured")
	}

	reqURL := tokenInfoURL + "?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, domainerrors.Internal("build google tokeninfo request").WithMeta("error", err.Error())
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, domainerrors.Unavailable("google tokeninfo request failed").WithMeta("error", err.Error())
	}
	defer resp.Body.Close()

	var body tokenInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, domainerrors.Unavailable("malformed google tokeninfo response")
	}

	if resp.StatusCode != http.StatusOK {
		msg := body.Error
		if msg == "" {
			msg = "invalid google id token"
		}
		return nil, domainerrors.Unauthenticated(msg)
	}
	if body.Aud != v.clientID {
		return nil, domainerrors.Unauthenticated("google id token was not issued for this app")
	}
	if body.Sub == "" {
		return nil, domainerrors.Unauthenticated("google id token missing subject")
	}

	return &Identity{
		Sub:           body.Sub,
		Email:         body.Email,
		EmailVerified: body.EmailVerified == "true",
		Name:          body.Name,
	}, nil
}

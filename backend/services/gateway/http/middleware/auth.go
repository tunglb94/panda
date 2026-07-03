package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fairride/identity/infrastructure/jwt"
)

type contextKey string

// ClaimsKey is the context key under which Auth middleware stores JWT claims.
const ClaimsKey contextKey = "claims"

// Auth validates Bearer JWTs on every request and stores the claims in the context.
// Requests without a valid token receive a 401 response and the chain is halted.
func Auth(svc *jwt.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := extractBearer(r)
			if !ok {
				writeAuthError(w, "missing or invalid authorization header")
				return
			}
			claims, err := svc.ValidateAccessToken(token)
			if err != nil {
				writeAuthError(w, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext retrieves JWT claims stored by the Auth middleware.
// Returns false if the middleware was not applied or the token was invalid.
func ClaimsFromContext(ctx context.Context) (*jwt.AccessClaims, bool) {
	c, ok := ctx.Value(ClaimsKey).(*jwt.AccessClaims)
	return c, ok
}

func extractBearer(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

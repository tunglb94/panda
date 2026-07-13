package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fairride/gateway/http/middleware"
	"github.com/fairride/identity/infrastructure/jwt"
)

func newTestTokenService(t *testing.T) *jwt.TokenService {
	t.Helper()
	svc, err := jwt.NewTokenService(jwt.Config{
		AccessSecret:    "test-access-secret-long-enough-32ch",
		RefreshSecret:   "test-refresh-secret-long-enough-32c",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewTokenService: %v", err)
	}
	return svc
}

func TestAuth_NoHeader(t *testing.T) {
	svc := newTestTokenService(t)
	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuth_InvalidHeaderFormat(t *testing.T) {
	svc := newTestTokenService(t)
	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	cases := []string{"Basic abc123", "token", "Bearer"}
	for _, h := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", h)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("header %q: want 401, got %d", h, w.Code)
		}
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	svc := newTestTokenService(t)
	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuth_ValidToken_ClaimsInContext(t *testing.T) {
	svc := newTestTokenService(t)

	token, err := svc.GenerateAccessToken("user-123", "rider", "role-1", time.Now())
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	var gotUserID string
	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r.Context())
		if !ok {
			t.Error("claims not in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		gotUserID = claims.UserID
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if gotUserID != "user-123" {
		t.Fatalf("want user-123, got %q", gotUserID)
	}
}

func TestRequireAdmin_RejectsNonAdmin(t *testing.T) {
	svc := newTestTokenService(t)
	token, _ := svc.GenerateAccessToken("d-1", "driver", "r-1", time.Now())

	handler := middleware.Auth(svc)(middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("non-admin should be rejected, want 401, got %d", w.Code)
	}
}

func TestRequireAdmin_AllowsAdmin(t *testing.T) {
	svc := newTestTokenService(t)
	token, _ := svc.GenerateAccessToken("a-1", "admin", "r-1", time.Now())

	handler := middleware.Auth(svc)(middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("admin should be allowed, want 200, got %d", w.Code)
	}
}

func TestAuth_BearerCaseInsensitive(t *testing.T) {
	svc := newTestTokenService(t)
	token, _ := svc.GenerateAccessToken("u-1", "rider", "r-1", time.Now())

	handler := middleware.Auth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, prefix := range []string{"Bearer", "bearer", "BEARER"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", prefix+" "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("prefix %q: want 200, got %d", prefix, w.Code)
		}
	}
}

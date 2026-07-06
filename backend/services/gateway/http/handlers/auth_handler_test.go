package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	driverentity "github.com/fairride/driver/domain/entity"
	"github.com/fairride/gateway/http/handlers"
	identityentity "github.com/fairride/identity/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

// ���── stubs ────────────────────────────────────────────────────────────────────

type stubUserFinder struct {
	user *identityentity.User
	err  error
}

func (s *stubUserFinder) FindByPhone(_ context.Context, _ string) (*identityentity.User, error) {
	return s.user, s.err
}

type stubDriverFinder struct {
	driver *driverentity.DriverProfile
	err    error
}

func (s *stubDriverFinder) FindByUserID(_ context.Context, _ string) (*driverentity.DriverProfile, error) {
	return s.driver, s.err
}

// ─── fixtures ─────────────────────────────────────────────────────────────────

func testUser() *identityentity.User {
	return identityentity.ReconstituteUser(
		"user-001", "+1234567890", "Test Driver", "",
		identityentity.TypeDriver, identityentity.StatusActive,
		"role-001", time.Now(), time.Now(),
	)
}

func testDriver() *driverentity.DriverProfile {
	return driverentity.ReconstituteDriverProfile(
		"drv-001", "user-001", "LIC-001",
		driverentity.VehicleTypeCar,
		"", "", "", "PLT-001",
		driverentity.OnlineStatusOffline,
		driverentity.VerificationStatusVerified,
		time.Now(), time.Now(),
	)
}

func newAuthHandler(t *testing.T) (*handlers.AuthHandler, *stubUserFinder, *stubDriverFinder) {
	t.Helper()
	uf := &stubUserFinder{}
	df := &stubDriverFinder{}
	return handlers.NewAuthHandler(uf, df, newTestTokenService(t)), uf, df
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	h, uf, df := newAuthHandler(t)
	uf.user = testUser()
	df.driver = testDriver()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"phone":"+1234567890"}`))
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["access_token"] == "" {
		t.Error("expected non-empty access_token")
	}
	if resp["driver_id"] != "drv-001" {
		t.Errorf("driver_id = %q, want drv-001", resp["driver_id"])
	}
	if resp["user_id"] != "user-001" {
		t.Errorf("user_id = %q, want user-001", resp["user_id"])
	}
}

func TestLogin_MissingPhone(t *testing.T) {
	h, _, _ := newAuthHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestLogin_BlankPhone(t *testing.T) {
	h, _, _ := newAuthHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"phone":"   "}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	h, uf, _ := newAuthHandler(t)
	uf.err = domainerrors.NotFound("user not found")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"phone":"+1"}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestLogin_DriverNotFound(t *testing.T) {
	h, uf, df := newAuthHandler(t)
	uf.user = testUser()
	df.err = domainerrors.NotFound("driver not found")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"phone":"+1"}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestLogin_DBNotConfigured(t *testing.T) {
	h := handlers.NewAuthHandler(nil, nil, newTestTokenService(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"phone":"+1"}`))
	w := httptest.NewRecorder()
	h.Login(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

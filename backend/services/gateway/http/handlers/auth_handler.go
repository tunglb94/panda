package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	driverentity "github.com/fairride/driver/domain/entity"
	identityentity "github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/infrastructure/jwt"
	domainerrors "github.com/fairride/shared/errors"
)

// userFinder abstracts identity.UserRepository for unit-testability.
type userFinder interface {
	FindByPhone(ctx context.Context, phone string) (*identityentity.User, error)
}

// driverFinder abstracts driver.DriverRepository for unit-testability.
type driverFinder interface {
	FindByUserID(ctx context.Context, userID string) (*driverentity.DriverProfile, error)
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	users    userFinder
	drivers  driverFinder
	tokenSvc *jwt.TokenService
}

// NewAuthHandler constructs an AuthHandler.
// Passing nil for users or drivers makes Login return 503.
func NewAuthHandler(users userFinder, drivers driverFinder, tokenSvc *jwt.TokenService) *AuthHandler {
	return &AuthHandler{users: users, drivers: drivers, tokenSvc: tokenSvc}
}

type loginRequest struct {
	Phone string `json:"phone"`
}

// Login handles POST /api/v1/auth/login.
// Looks up the user by phone, resolves the associated driver profile, and issues
// a JWT whose sub is the driver's DriverID. Downstream handlers can therefore use
// claims.UserID as the driver identifier without an additional lookup.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.users == nil || h.drivers == nil {
		writeJSON(w, http.StatusServiceUnavailable,
			errorResponse{Error: "authentication service not configured"})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" {
		writeBadRequest(w, "phone is required")
		return
	}

	user, err := h.users.FindByPhone(r.Context(), req.Phone)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	driver, err := h.drivers.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	now := time.Now()
	token, err := h.tokenSvc.GenerateAccessToken(driver.DriverID, string(user.Type), user.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	refreshTok, err := h.tokenSvc.GenerateRefreshToken(driver.DriverID, string(user.Type), user.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  token,
		"refresh_token": refreshTok.Token,
		"driver_id":     driver.DriverID,
		"user_id":       user.ID,
	})
}

// RiderLogin handles POST /api/v1/auth/rider/login.
// Looks up the user by phone, verifies they hold the Rider user type, and issues
// a JWT whose sub is the user's ID. No driver profile is required.
func (h *AuthHandler) RiderLogin(w http.ResponseWriter, r *http.Request) {
	if h.users == nil {
		writeJSON(w, http.StatusServiceUnavailable,
			errorResponse{Error: "authentication service not configured"})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" {
		writeBadRequest(w, "phone is required")
		return
	}

	user, err := h.users.FindByPhone(r.Context(), req.Phone)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	if user.Type != identityentity.TypeRider {
		writeDomainError(w, domainerrors.NotFound("rider not found"))
		return
	}

	now := time.Now()
	token, err := h.tokenSvc.GenerateAccessToken(user.ID, string(user.Type), user.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	refreshTok, err := h.tokenSvc.GenerateRefreshToken(user.ID, string(user.Type), user.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  token,
		"refresh_token": refreshTok.Token,
		"rider_id":      user.ID,
	})
}

// AdminLogin handles POST /api/v1/auth/admin/login. Same phone-lookup
// pattern as RiderLogin, gated on identityentity.TypeAdmin instead —
// backs the RequireAdmin-gated KYC review dashboard endpoints (Phần 11).
// No admin-provisioning UI exists in this phase; an admin User row must
// already exist (e.g. seeded) for this to succeed.
func (h *AuthHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if h.users == nil {
		writeJSON(w, http.StatusServiceUnavailable,
			errorResponse{Error: "authentication service not configured"})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" {
		writeBadRequest(w, "phone is required")
		return
	}

	user, err := h.users.FindByPhone(r.Context(), req.Phone)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	if user.Type != identityentity.TypeAdmin {
		writeDomainError(w, domainerrors.NotFound("admin not found"))
		return
	}

	now := time.Now()
	token, err := h.tokenSvc.GenerateAccessToken(user.ID, string(user.Type), user.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	refreshTok, err := h.tokenSvc.GenerateRefreshToken(user.ID, string(user.Type), user.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  token,
		"refresh_token": refreshTok.Token,
		"admin_id":      user.ID,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh handles POST /api/v1/auth/refresh — exchanges a still-valid
// refresh token for a new access token (and a new refresh token, extending
// the session) without requiring the phone-lookup login flow again. Access
// tokens are short-lived (15 min, see identity/infrastructure/jwt/config.go)
// specifically so this endpoint has to exist — every authenticated request
// otherwise starts failing with 401 the moment the access token expires,
// with no way for the client to recover short of a full re-login.
//
// UserType/RoleID are read directly from the refresh token's own claims
// (embedded at issuance) rather than re-queried from identity/driver
// repositories — this endpoint has no dependency on either, by design.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		writeBadRequest(w, "refresh_token is required")
		return
	}

	claims, err := h.tokenSvc.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	now := time.Now()
	accessTok, err := h.tokenSvc.GenerateAccessToken(claims.UserID, claims.UserType, claims.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	refreshTok, err := h.tokenSvc.GenerateRefreshToken(claims.UserID, claims.UserType, claims.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessTok,
		"refresh_token": refreshTok.Token,
	})
}

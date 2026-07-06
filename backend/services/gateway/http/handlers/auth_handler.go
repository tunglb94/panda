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

	token, err := h.tokenSvc.GenerateAccessToken(driver.DriverID, string(user.Type), user.RoleID, time.Now())
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token": token,
		"driver_id":    driver.DriverID,
		"user_id":      user.ID,
	})
}

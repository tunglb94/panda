package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/fairride/gateway/http/middleware"
	identityapp "github.com/fairride/identity/app"
	identityentity "github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/infrastructure/jwt"
	domainerrors "github.com/fairride/shared/errors"
)

// OTPAuthHandler implements the phone-OTP and Google Sign-In login
// endpoints — the "no office visit" self-registration path. Kept separate
// from AuthHandler (which owns the older phone-lookup-only endpoints, left
// unchanged for backward compatibility) since this is a distinct concern
// with its own dependency set.
type OTPAuthHandler struct {
	requestOTP    *identityapp.RequestOTPUseCase
	verifyOTP     *identityapp.VerifyOTPUseCase
	googleLogin   *identityapp.GoogleLoginUseCase
	users         userByIDFinder
	recordLogin   *identityapp.RecordLoginUseCase
	upsertDevice  *identityapp.UpsertDeviceUseCase
	tokenSvc      *jwt.TokenService
	isDevelopment bool
}

// NewOTPAuthHandler constructs an OTPAuthHandler. Any nil use case makes the
// corresponding endpoint return 503 — same graceful-degrade pattern as every
// other optional dependency in this gateway (e.g. GOOGLE_CLIENT_ID unset).
// recordLogin/upsertDevice are best-effort telemetry — nil simply skips
// them (see recordLoginAttempt/upsertLoginDevice), never fails a login.
func NewOTPAuthHandler(
	requestOTP *identityapp.RequestOTPUseCase,
	verifyOTP *identityapp.VerifyOTPUseCase,
	googleLogin *identityapp.GoogleLoginUseCase,
	users userByIDFinder,
	recordLogin *identityapp.RecordLoginUseCase,
	upsertDevice *identityapp.UpsertDeviceUseCase,
	tokenSvc *jwt.TokenService,
	isDevelopment bool,
) *OTPAuthHandler {
	return &OTPAuthHandler{
		requestOTP:    requestOTP,
		verifyOTP:     verifyOTP,
		googleLogin:   googleLogin,
		users:         users,
		recordLogin:   recordLogin,
		upsertDevice:  upsertDevice,
		tokenSvc:      tokenSvc,
		isDevelopment: isDevelopment,
	}
}

// deviceRequest is embedded in the OTP-verify and Google-login request
// bodies — every field is optional, sent best-effort by the client (see
// apps/*/lib/core/device/device_info.dart). Absent entirely, login still
// works exactly as before this phase.
type deviceRequest struct {
	DeviceID   string `json:"device_id"`
	Platform   string `json:"platform"`
	Model      string `json:"model"`
	AppVersion string `json:"app_version"`
	FCMToken   string `json:"fcm_token"`
}

// recordLoginAttempt appends a login-history row, best-effort — a failure
// here (or a nil recordLogin use case, e.g. DB_URL unset) must never fail
// the login response itself.
func (h *OTPAuthHandler) recordLoginAttempt(r *http.Request, userID string, dev deviceRequest, method identityentity.LoginMethod, success bool) {
	if h.recordLogin == nil {
		return
	}
	_ = h.recordLogin.Execute(r.Context(), identityapp.RecordLoginInput{
		UserID:      userID,
		IP:          clientIP(r),
		DeviceID:    dev.DeviceID,
		Platform:    dev.Platform,
		LoginMethod: method,
		Success:     success,
	})
}

// upsertLoginDevice registers/refreshes the device on a successful login —
// best-effort, same rationale as recordLoginAttempt.
func (h *OTPAuthHandler) upsertLoginDevice(r *http.Request, userID string, dev deviceRequest) {
	if h.upsertDevice == nil {
		return
	}
	_ = h.upsertDevice.Execute(r.Context(), identityapp.UpsertDeviceInput{
		UserID:     userID,
		DeviceID:   dev.DeviceID,
		Platform:   dev.Platform,
		Model:      dev.Model,
		AppVersion: dev.AppVersion,
		FCMToken:   dev.FCMToken,
	})
}

// clientIP prefers the leftmost X-Forwarded-For entry (this gateway
// typically sits behind a Cloudflare tunnel — see apps/*'s AppConfig
// pointing at a trycloudflare.com URL), falling back to the raw
// RemoteAddr (host part only, port stripped) for direct/local connections.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if idx := strings.Index(fwd, ","); idx >= 0 {
			return strings.TrimSpace(fwd[:idx])
		}
		return strings.TrimSpace(fwd)
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx >= 0 {
		host = host[:idx]
	}
	return host
}

type otpRequestRequest struct {
	Phone    string `json:"phone"`
	UserType string `json:"user_type"`
}

// RequestOTP handles POST /api/v1/auth/otp/request.
func (h *OTPAuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	if h.requestOTP == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "otp service not configured"})
		return
	}
	var req otpRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" {
		writeBadRequest(w, "phone is required")
		return
	}

	result, err := h.requestOTP.Execute(r.Context(), req.Phone)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	body := map[string]any{
		"expires_in": int(result.ExpiresIn.Seconds()),
	}
	// Never expose the real code outside development — see plan's OTP dev
	// visibility decision.
	if h.isDevelopment {
		body["debug_otp_code"] = result.Code
	}
	writeJSON(w, http.StatusOK, body)
}

type otpVerifyRequest struct {
	Phone    string `json:"phone"`
	Code     string `json:"code"`
	UserType string `json:"user_type"`
	deviceRequest
}

// VerifyOTP handles POST /api/v1/auth/otp/verify.
func (h *OTPAuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if h.verifyOTP == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "otp service not configured"})
		return
	}
	var req otpVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	userType, err := parseUserType(req.UserType)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	result, err := h.verifyOTP.Execute(r.Context(), identityapp.VerifyOTPInput{
		PhoneNumber: req.Phone,
		Code:        req.Code,
		UserType:    userType,
	})
	if err != nil {
		h.recordLoginAttempt(r, "", req.deviceRequest, identityentity.LoginMethodOTP, false)
		writeDomainError(w, err)
		return
	}
	h.recordLoginAttempt(r, result.User.ID, req.deviceRequest, identityentity.LoginMethodOTP, true)
	h.upsertLoginDevice(r, result.User.ID, req.deviceRequest)
	h.writeLoginTokens(w, result.User, result.IsNewUser)
}

type googleLoginRequest struct {
	IDToken  string `json:"id_token"`
	UserType string `json:"user_type"`
	deviceRequest
}

// GoogleLogin handles POST /api/v1/auth/google.
func (h *OTPAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.googleLogin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "google sign-in is not configured"})
		return
	}
	var req googleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	userType, err := parseUserType(req.UserType)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	result, err := h.googleLogin.Execute(r.Context(), identityapp.GoogleLoginInput{
		IDToken:  req.IDToken,
		UserType: userType,
	})
	if err != nil {
		h.recordLoginAttempt(r, "", req.deviceRequest, identityentity.LoginMethodGoogle, false)
		writeDomainError(w, err)
		return
	}
	h.recordLoginAttempt(r, result.User.ID, req.deviceRequest, identityentity.LoginMethodGoogle, true)
	h.upsertLoginDevice(r, result.User.ID, req.deviceRequest)
	h.writeLoginTokens(w, result.User, result.IsNewUser)
}

func (h *OTPAuthHandler) writeLoginTokens(w http.ResponseWriter, user *identityentity.User, isNewUser bool) {
	now := time.Now()
	accessTok, err := h.tokenSvc.GenerateAccessToken(user.ID, string(user.Type), user.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	refreshTok, err := h.tokenSvc.GenerateRefreshToken(user.ID, string(user.Type), user.RoleID, now)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  accessTok,
		"refresh_token": refreshTok.Token,
		"user_id":       user.ID,
		"user_type":     string(user.Type),
		"is_new_user":   isNewUser,
	})
}

// Me handles GET /api/v1/auth/me — returns the calling account's identity
// + capability flags. Flutter uses this to decide, at Splash, whether the
// Driver app's gate is satisfied (driver_enabled) independent of the
// separate Driver KYC approval check (see plan's Startup Flow phase).
func (h *OTPAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if h.users == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "user service not configured"})
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	user, err := h.users.FindByID(r.Context(), claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":        user.ID,
		"phone":          user.PhoneNumber,
		"email":          user.Email,
		"google_linked":  user.GoogleSub != "",
		"type":           string(user.Type),
		"driver_enabled": user.DriverEnabled,
	})
}

func parseUserType(raw string) (identityentity.UserType, error) {
	switch identityentity.UserType(strings.TrimSpace(raw)) {
	case identityentity.TypeRider:
		return identityentity.TypeRider, nil
	case identityentity.TypeDriver:
		return identityentity.TypeDriver, nil
	default:
		return "", domainerrors.InvalidArgument(`user_type must be "rider" or "driver"`)
	}
}

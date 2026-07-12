package handlers

import (
	"context"
	"net/http"
	"strings"

	driverentity "github.com/fairride/driver/domain/entity"
	"github.com/fairride/gateway/http/middleware"
	identityentity "github.com/fairride/identity/domain/entity"
	notificationapp "github.com/fairride/notification/app"
	reviewentity "github.com/fairride/review/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

// userByIDFinder abstracts identity.UserRepository.FindByID — a subset
// distinct from AuthHandler's userFinder (which only needs FindByPhone).
type userByIDFinder interface {
	FindByID(ctx context.Context, id string) (*identityentity.User, error)
}

// driverByIDFinder abstracts driver.DriverRepository.FindByID.
type driverByIDFinder interface {
	FindByID(ctx context.Context, driverID string) (*driverentity.DriverProfile, error)
}

// averageRatingReader is the review service's new read-only aggregate use
// case (app.GetAverageRatingUseCase), consumed directly as a Go dependency
// — see that use case's doc comment for why this isn't a new reviewpb RPC.
type averageRatingReader interface {
	Execute(ctx context.Context, rateeID string, raterRole reviewentity.Role) (avg float64, count int32, err error)
}

// CallHandler exposes Phone Call (Part 1) and Contact Card (Part 4) over HTTP.
type CallHandler struct {
	trips      TripStatusClient
	users      userByIDFinder
	drivers    driverByIDFinder
	ratings    averageRatingReader
	recordCall *notificationapp.RecordCallUseCase
}

func NewCallHandler(trips TripStatusClient, users userByIDFinder, drivers driverByIDFinder, ratings averageRatingReader, recordCall *notificationapp.RecordCallUseCase) *CallHandler {
	return &CallHandler{trips: trips, users: users, drivers: drivers, ratings: ratings, recordCall: recordCall}
}

func (h *CallHandler) configured() bool {
	return h != nil && h.trips != nil && h.users != nil
}

// otherParty resolves who requesterID should be contacting for a trip, and
// whether that party is the trip's driver (vs its rider).
func otherParty(trip notificationapp.TripSnapshot, requesterID string) (calleeID string, calleeIsDriver bool) {
	if requesterID == trip.RiderID {
		return trip.DriverID, true
	}
	return trip.RiderID, false
}

// loadTripAndAuthorize fetches the trip via tripReaderAdapter and verifies
// requesterID is a participant with a counterpart already assigned.
func (h *CallHandler) loadTripAndAuthorize(ctx context.Context, tripID, requesterID string) (notificationapp.TripSnapshot, error) {
	trip, err := (tripReaderAdapter{client: h.trips}).GetTrip(ctx, tripID)
	if err != nil {
		return notificationapp.TripSnapshot{}, err
	}
	if requesterID != trip.RiderID && requesterID != trip.DriverID {
		return notificationapp.TripSnapshot{}, domainerrors.PermissionDenied("requester is not a participant of this trip")
	}
	if trip.DriverID == "" {
		return notificationapp.TripSnapshot{}, domainerrors.PreconditionFailed("trip has no driver assigned yet")
	}
	return trip, nil
}

// GetContact handles GET /api/v1/rides/{tripID}/contact (Part 4 — Contact Card).
// Never returns a real phone number, only masked_phone for display.
func (h *CallHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "contact service not configured"})
		return
	}
	tripID := r.PathValue("tripID")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	trip, err := h.loadTripAndAuthorize(r.Context(), tripID, claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	calleeID, calleeIsDriver := otherParty(trip, claims.UserID)

	body, err := h.contactBody(r.Context(), calleeID, calleeIsDriver)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	body["trip_id"] = tripID
	writeJSON(w, http.StatusOK, body)
}

func (h *CallHandler) contactBody(ctx context.Context, calleeID string, calleeIsDriver bool) (map[string]any, error) {
	identityUserID := calleeID
	body := map[string]any{}
	if calleeIsDriver {
		if h.drivers == nil {
			return nil, domainerrors.Unavailable("driver service not configured")
		}
		profile, err := h.drivers.FindByID(ctx, calleeID)
		if err != nil {
			return nil, err
		}
		identityUserID = profile.UserID
		body["vehicle_type"] = string(profile.VehicleType)
		body["vehicle_brand"] = profile.VehicleBrand
		body["vehicle_model"] = profile.VehicleModel
		body["plate_number"] = profile.PlateNumber
	}

	user, err := h.users.FindByID(ctx, identityUserID)
	if err != nil {
		return nil, err
	}
	body["name"] = user.Name
	body["masked_phone"] = maskPhone(user.PhoneNumber)

	if h.ratings != nil {
		role := reviewentity.RoleRider
		if !calleeIsDriver {
			role = reviewentity.RoleDriver
		}
		if avg, count, ratingErr := h.ratings.Execute(ctx, calleeID, role); ratingErr == nil {
			body["rating"] = avg
			body["rating_count"] = count
		}
	}
	return body, nil
}

// Call handles POST /api/v1/rides/{tripID}/call (Part 1 — Phone Call).
// Returns the real phone number ONLY as the response of this explicit
// call-intent action, for the client to immediately open with url_launcher's
// tel: scheme — never persisted, never rendered as text.
func (h *CallHandler) Call(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "call service not configured"})
		return
	}
	tripID := r.PathValue("tripID")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	trip, err := h.loadTripAndAuthorize(r.Context(), tripID, claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if notificationapp.IsTripStatusClosed(trip.Status) {
		writeDomainError(w, domainerrors.PreconditionFailed("trip has ended — calling is no longer available"))
		return
	}
	calleeID, calleeIsDriver := otherParty(trip, claims.UserID)

	identityUserID := calleeID
	if calleeIsDriver {
		if h.drivers == nil {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "driver service not configured"})
			return
		}
		profile, profErr := h.drivers.FindByID(r.Context(), calleeID)
		if profErr != nil {
			writeDomainError(w, profErr)
			return
		}
		identityUserID = profile.UserID
	}
	user, err := h.users.FindByID(r.Context(), identityUserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	if h.recordCall != nil {
		_, _ = h.recordCall.Execute(r.Context(), tripID, claims.UserID, calleeID)
	}
	writeJSON(w, http.StatusOK, map[string]string{"phone": user.PhoneNumber})
}

// maskPhone hides every digit except the first 3 and last 3 (e.g.
// "0901234123" -> "090****123") for display in the Contact Card. Never used
// for the actual dialer action — see Call, which returns the real number.
func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	n := len(phone)
	if n <= 6 {
		return strings.Repeat("*", n)
	}
	return phone[:3] + strings.Repeat("*", n-6) + phone[n-3:]
}

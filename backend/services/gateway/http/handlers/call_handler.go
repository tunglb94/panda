package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/fairride/booking/grpc/bookingpb"
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

// driverVerificationReader/vehicleVerificationReader are the driver
// service's KYC read use cases (app.GetDriverVerificationUseCase /
// app.GetVehicleVerificationUseCase), consumed directly as Go dependencies
// — same in-process pattern as averageRatingReader above.
type driverVerificationReader interface {
	Execute(ctx context.Context, driverID string) (*driverentity.DriverVerification, error)
}

type vehicleVerificationReader interface {
	Execute(ctx context.Context, driverID string) (*driverentity.VehicleVerification, error)
}

// CallHandler exposes Phone Call (Part 1) and Contact Card (Part 4, now
// enriched with KYC verified badge / join date / trip count — Phần 8 of the
// Driver KYC spec) over HTTP.
type CallHandler struct {
	trips      TripStatusClient
	users      userByIDFinder
	drivers    driverByIDFinder
	ratings    averageRatingReader
	recordCall *notificationapp.RecordCallUseCase

	driverVerifications  driverVerificationReader
	vehicleVerifications vehicleVerificationReader
	// booking is reused (not a new dependency surface) purely to count a
	// driver's completed trips for the Contact Card — read-only, no Ride
	// Lifecycle code touched.
	booking BookingClient
}

func NewCallHandler(
	trips TripStatusClient,
	users userByIDFinder,
	drivers driverByIDFinder,
	ratings averageRatingReader,
	recordCall *notificationapp.RecordCallUseCase,
	driverVerifications driverVerificationReader,
	vehicleVerifications vehicleVerificationReader,
	booking BookingClient,
) *CallHandler {
	return &CallHandler{
		trips: trips, users: users, drivers: drivers, ratings: ratings, recordCall: recordCall,
		driverVerifications: driverVerifications, vehicleVerifications: vehicleVerifications, booking: booking,
	}
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
		body["plate_number"] = maskPlate(profile.PlateNumber)
		body["joined_at"] = profile.CreatedAt.UTC().Format(time.RFC3339)
		body["is_verified"] = h.isDriverKYCVerified(ctx, calleeID)
		body["trip_count"] = h.countCompletedTrips(ctx, calleeID)
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

// isDriverKYCVerified reports whether calleeID's Driver KYC AND Vehicle
// Verification are both Approved (Phần 8's "✓ Đã xác minh" badge) — a
// stricter, additive check on top of DriverProfile's own legacy
// verification_status, never that field alone.
func (h *CallHandler) isDriverKYCVerified(ctx context.Context, driverID string) bool {
	if h.driverVerifications == nil || h.vehicleVerifications == nil {
		return false
	}
	dv, err := h.driverVerifications.Execute(ctx, driverID)
	if err != nil || !dv.IsApproved() {
		return false
	}
	vv, err := h.vehicleVerifications.Execute(ctx, driverID)
	if err != nil || !vv.IsApproved() {
		return false
	}
	return true
}

// countCompletedTrips reuses the existing BookingClient.ListDriverTrips (no
// new Trip-service surface, no Ride Lifecycle code touched) to count a
// driver's completed/settled trips for the Contact Card's "Số chuyến".
// Best-effort: any failure counts as 0 rather than failing the whole
// Contact Card response.
func (h *CallHandler) countCompletedTrips(ctx context.Context, driverID string) int {
	if h.booking == nil {
		return 0
	}
	resp, err := h.booking.ListDriverTrips(ctx, &bookingpb.ListTripsRequest{PartyId: driverID})
	if err != nil {
		return 0
	}
	count := 0
	for _, t := range resp.GetTrips() {
		switch t.GetStatus() {
		case "completed", "settled":
			count++
		}
	}
	return count
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

// maskPlate hides the middle of a plate number (e.g. "59-X1 123.45" ->
// "59****45") for the Contact Card (Phần 10 of the Driver KYC Hardening
// spec — "Biển số (mask)"). Short/malformed plates are masked entirely
// rather than risk showing the whole thing.
func maskPlate(plate string) string {
	plate = strings.TrimSpace(plate)
	n := len(plate)
	if n <= 4 {
		return strings.Repeat("*", n)
	}
	return plate[:2] + strings.Repeat("*", n-4) + plate[n-2:]
}

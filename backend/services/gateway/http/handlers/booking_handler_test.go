package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fairride/booking/grpc/bookingpb"
	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/gateway/http/middleware"
	"github.com/fairride/identity/infrastructure/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─── Stub client ─────────────────────────────────────────────────────────────

type stubBookingClient struct {
	bookRide              func(ctx context.Context, in *bookingpb.BookRideRequest, opts ...grpc.CallOption) (*bookingpb.BookRideResponse, error)
	acceptDispatchOffer   func(ctx context.Context, in *bookingpb.AcceptDispatchOfferRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	rejectDispatchOffer   func(ctx context.Context, in *bookingpb.RejectDispatchOfferRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	startTrip             func(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	finishTrip            func(ctx context.Context, in *bookingpb.FinishTripRequest, opts ...grpc.CallOption) (*bookingpb.FinishedTripResponse, error)
	getBookingDetails     func(ctx context.Context, in *bookingpb.GetBookingDetailsRequest, opts ...grpc.CallOption) (*bookingpb.BookingDetailsResponse, error)
	getDriverCurrentOffer func(ctx context.Context, in *bookingpb.GetDriverCurrentOfferRequest, opts ...grpc.CallOption) (*bookingpb.GetDriverCurrentOfferResponse, error)
	cancelRide            func(ctx context.Context, in *bookingpb.CancelRideRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
}

func (s *stubBookingClient) BookRide(ctx context.Context, in *bookingpb.BookRideRequest, opts ...grpc.CallOption) (*bookingpb.BookRideResponse, error) {
	return s.bookRide(ctx, in, opts...)
}
func (s *stubBookingClient) AcceptDispatchOffer(ctx context.Context, in *bookingpb.AcceptDispatchOfferRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
	return s.acceptDispatchOffer(ctx, in, opts...)
}
func (s *stubBookingClient) RejectDispatchOffer(ctx context.Context, in *bookingpb.RejectDispatchOfferRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
	return s.rejectDispatchOffer(ctx, in, opts...)
}
func (s *stubBookingClient) ArriveAtPickup(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
	return &bookingpb.BookingActionResponse{TripId: in.GetTripId(), Status: "driver_arrived"}, nil
}
func (s *stubBookingClient) StartTrip(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
	return s.startTrip(ctx, in, opts...)
}
func (s *stubBookingClient) FinishTrip(ctx context.Context, in *bookingpb.FinishTripRequest, opts ...grpc.CallOption) (*bookingpb.FinishedTripResponse, error) {
	return s.finishTrip(ctx, in, opts...)
}
func (s *stubBookingClient) GetBookingDetails(ctx context.Context, in *bookingpb.GetBookingDetailsRequest, opts ...grpc.CallOption) (*bookingpb.BookingDetailsResponse, error) {
	return s.getBookingDetails(ctx, in, opts...)
}
func (s *stubBookingClient) GetDriverCurrentOffer(ctx context.Context, in *bookingpb.GetDriverCurrentOfferRequest, opts ...grpc.CallOption) (*bookingpb.GetDriverCurrentOfferResponse, error) {
	if s.getDriverCurrentOffer != nil {
		return s.getDriverCurrentOffer(ctx, in, opts...)
	}
	return &bookingpb.GetDriverCurrentOfferResponse{HasOffer: false}, nil
}
func (s *stubBookingClient) CancelRide(ctx context.Context, in *bookingpb.CancelRideRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
	if s.cancelRide != nil {
		return s.cancelRide(ctx, in, opts...)
	}
	return &bookingpb.BookingActionResponse{TripId: in.GetTripId(), Status: "cancelled"}, nil
}
func (s *stubBookingClient) PayRide(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.FinishedTripResponse, error) {
	return &bookingpb.FinishedTripResponse{TripId: in.GetTripId(), Status: "settled"}, nil
}
func (s *stubBookingClient) ListRiderTrips(_ context.Context, _ *bookingpb.ListTripsRequest, _ ...grpc.CallOption) (*bookingpb.TripListResponse, error) {
	return &bookingpb.TripListResponse{}, nil
}
func (s *stubBookingClient) ListDriverTrips(_ context.Context, _ *bookingpb.ListTripsRequest, _ ...grpc.CallOption) (*bookingpb.TripListResponse, error) {
	return &bookingpb.TripListResponse{}, nil
}

// ─── Test helpers ─────────────────────────────────────────────────────────────

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

// withClaims injects fake JWT claims into the request context, mimicking the Auth middleware.
func withClaims(r *http.Request, userID, userType string) *http.Request {
	claims := &jwt.AccessClaims{UserID: userID, UserType: userType, RoleID: "role-1"}
	ctx := context.WithValue(r.Context(), middleware.ClaimsKey, claims)
	return r.WithContext(ctx)
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewBuffer(b)
}

// ─── BookRide ─────────────────────────────────────────────────────────────────

func TestBookRide_Success(t *testing.T) {
	stub := &stubBookingClient{
		bookRide: func(_ context.Context, in *bookingpb.BookRideRequest, _ ...grpc.CallOption) (*bookingpb.BookRideResponse, error) {
			if in.GetRiderId() != "rider-1" {
				return nil, status.Error(codes.InvalidArgument, "bad rider")
			}
			return &bookingpb.BookRideResponse{TripId: "trip-99", Status: "searching"}, nil
		},
	}
	h := handlers.NewBookingHandler(stub)

	body := jsonBody(t, map[string]any{
		"pickup_address":  "A",
		"dropoff_address": "B",
		"pickup_lat":      1.0,
		"pickup_lon":      2.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides", body)
	req = withClaims(req, "rider-1", "rider")
	w := httptest.NewRecorder()

	h.BookRide(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["trip_id"] != "trip-99" {
		t.Fatalf("want trip-99, got %q", resp["trip_id"])
	}
	if resp["status"] != "searching" {
		t.Fatalf("want searching, got %q", resp["status"])
	}
}

func TestBookRide_MissingPickup(t *testing.T) {
	h := handlers.NewBookingHandler(&stubBookingClient{})
	body := jsonBody(t, map[string]any{"dropoff_address": "B"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides", body)
	req = withClaims(req, "rider-1", "rider")
	w := httptest.NewRecorder()
	h.BookRide(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestBookRide_GRPCError_MapsToHTTP(t *testing.T) {
	stub := &stubBookingClient{
		bookRide: func(_ context.Context, _ *bookingpb.BookRideRequest, _ ...grpc.CallOption) (*bookingpb.BookRideResponse, error) {
			return nil, status.Error(codes.Internal, "upstream down")
		},
	}
	h := handlers.NewBookingHandler(stub)
	body := jsonBody(t, map[string]any{"pickup_address": "A", "dropoff_address": "B"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides", body)
	req = withClaims(req, "rider-1", "rider")
	w := httptest.NewRecorder()
	h.BookRide(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

// ─── GetBooking ───────────────────────────────────────────────────────────────

func TestGetBooking_Success(t *testing.T) {
	stub := &stubBookingClient{
		getBookingDetails: func(_ context.Context, in *bookingpb.GetBookingDetailsRequest, _ ...grpc.CallOption) (*bookingpb.BookingDetailsResponse, error) {
			return &bookingpb.BookingDetailsResponse{
				TripId:         in.GetTripId(),
				TripStatus:     "in_progress",
				RiderId:        "rider-1",
				DriverId:       "driver-1",
				PickupAddress:  "A",
				DropoffAddress: "B",
				DispatchStatus: "accepted",
			}, nil
		},
	}
	h := handlers.NewBookingHandler(stub)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rides/trip-1", nil)
	req = withClaims(req, "rider-1", "rider")
	// Simulate path value from ServeMux
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.GetBooking(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetBooking_NotFound(t *testing.T) {
	stub := &stubBookingClient{
		getBookingDetails: func(_ context.Context, _ *bookingpb.GetBookingDetailsRequest, _ ...grpc.CallOption) (*bookingpb.BookingDetailsResponse, error) {
			return nil, status.Error(codes.NotFound, "trip not found")
		},
	}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rides/nope", nil)
	req = withClaims(req, "rider-1", "rider")
	req.SetPathValue("tripID", "nope")
	w := httptest.NewRecorder()
	h.GetBooking(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

// ─── AcceptDispatchOffer ──────────────────────────────────────────────────────

func TestAcceptDispatchOffer_Success(t *testing.T) {
	stub := &stubBookingClient{
		acceptDispatchOffer: func(_ context.Context, in *bookingpb.AcceptDispatchOfferRequest, _ ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
			return &bookingpb.BookingActionResponse{TripId: in.GetTripId(), Status: "driver_assigned"}, nil
		},
	}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/accept", strings.NewReader("{}"))
	req = withClaims(req, "driver-1", "driver")
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.AcceptDispatchOffer(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestAcceptDispatchOffer_PreconditionFailed(t *testing.T) {
	stub := &stubBookingClient{
		acceptDispatchOffer: func(_ context.Context, _ *bookingpb.AcceptDispatchOfferRequest, _ ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
			return nil, status.Error(codes.FailedPrecondition, "offer already taken")
		},
	}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/accept", nil)
	req = withClaims(req, "driver-1", "driver")
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.AcceptDispatchOffer(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", w.Code)
	}
}

// ─── RejectDispatchOffer ──────────────────────────────────────────────────────

func TestRejectDispatchOffer_Success(t *testing.T) {
	stub := &stubBookingClient{
		rejectDispatchOffer: func(_ context.Context, in *bookingpb.RejectDispatchOfferRequest, _ ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
			return &bookingpb.BookingActionResponse{TripId: in.GetTripId(), Status: "searching"}, nil
		},
	}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/reject", strings.NewReader("{}"))
	req = withClaims(req, "driver-1", "driver")
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.RejectDispatchOffer(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

// ─── StartTrip ───────────────────────────────────────────────────────────────

func TestStartTrip_Success(t *testing.T) {
	stub := &stubBookingClient{
		startTrip: func(_ context.Context, in *bookingpb.StartTripRequest, _ ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
			return &bookingpb.BookingActionResponse{TripId: in.GetTripId(), Status: "in_progress"}, nil
		},
	}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/start", nil)
	req = withClaims(req, "driver-1", "driver")
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.StartTrip(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestStartTrip_NotFound(t *testing.T) {
	stub := &stubBookingClient{
		startTrip: func(_ context.Context, _ *bookingpb.StartTripRequest, _ ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
			return nil, status.Error(codes.NotFound, "trip not found")
		},
	}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/nope/start", nil)
	req = withClaims(req, "driver-1", "driver")
	req.SetPathValue("tripID", "nope")
	w := httptest.NewRecorder()
	h.StartTrip(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

// ─── FinishTrip ───────────────────────────────────────────────────────────────

func TestFinishTrip_Success(t *testing.T) {
	stub := &stubBookingClient{
		finishTrip: func(_ context.Context, in *bookingpb.FinishTripRequest, _ ...grpc.CallOption) (*bookingpb.FinishedTripResponse, error) {
			return &bookingpb.FinishedTripResponse{
				TripId:      in.GetTripId(),
				Status:      "completed",
				FinalFare:   1500,
				Currency:    "USD",
				VehicleType: in.GetVehicleType(),
				DistanceKm:  in.GetDistanceKm(),
				DurationMin: in.GetDurationMin(),
			}, nil
		},
	}
	h := handlers.NewBookingHandler(stub)
	body := jsonBody(t, map[string]any{
		"vehicle_type": "car",
		"distance_km":  5.0,
		"duration_min": 15.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/finish", body)
	req = withClaims(req, "driver-1", "driver")
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.FinishTrip(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "completed" {
		t.Fatalf("want completed, got %q", resp["status"])
	}
}

func TestFinishTrip_MissingVehicleType(t *testing.T) {
	h := handlers.NewBookingHandler(&stubBookingClient{})
	body := jsonBody(t, map[string]any{"distance_km": 5.0, "duration_min": 15.0})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/finish", body)
	req = withClaims(req, "driver-1", "driver")
	req.SetPathValue("tripID", "trip-1")
	w := httptest.NewRecorder()
	h.FinishTrip(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

// ─── Error mapping ────────────────────────────────────────────────────────────

func TestGRPCErrorMapping(t *testing.T) {
	cases := []struct {
		grpcCode codes.Code
		wantHTTP int
	}{
		{codes.NotFound, http.StatusNotFound},
		{codes.InvalidArgument, http.StatusBadRequest},
		{codes.Unauthenticated, http.StatusUnauthorized},
		{codes.PermissionDenied, http.StatusForbidden},
		{codes.FailedPrecondition, http.StatusUnprocessableEntity},
		{codes.AlreadyExists, http.StatusConflict},
		{codes.Internal, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		stub := &stubBookingClient{
			startTrip: func(_ context.Context, _ *bookingpb.StartTripRequest, _ ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
				return nil, status.Error(tc.grpcCode, "err")
			},
		}
		h := handlers.NewBookingHandler(stub)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/t/start", nil)
		req = withClaims(req, "driver-1", "driver")
		req.SetPathValue("tripID", "t")
		w := httptest.NewRecorder()
		h.StartTrip(w, req)
		if w.Code != tc.wantHTTP {
			t.Errorf("grpc code %v: want HTTP %d, got %d", tc.grpcCode, tc.wantHTTP, w.Code)
		}
	}
}

// ─── GetDriverOffer ──────────────────────────────────────────────────────────

func TestGetDriverOffer_HasOffer(t *testing.T) {
	expires := time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC)
	stub := &stubBookingClient{
		getDriverCurrentOffer: func(_ context.Context, _ *bookingpb.GetDriverCurrentOfferRequest, _ ...grpc.CallOption) (*bookingpb.GetDriverCurrentOfferResponse, error) {
			return &bookingpb.GetDriverCurrentOfferResponse{
				HasOffer:       true,
				TripId:         "trip1",
				PickupAddress:  "123 Main St",
				DropoffAddress: "456 Elm Ave",
				OfferExpiresAt: timestamppb.New(expires),
			}, nil
		},
	}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/driver/current-offer", nil)
	req = withClaims(req, "driver-1", "driver")
	w := httptest.NewRecorder()
	h.GetDriverOffer(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["has_offer"] != true {
		t.Errorf("has_offer = %v, want true", body["has_offer"])
	}
	if body["trip_id"] != "trip1" {
		t.Errorf("trip_id = %v, want trip1", body["trip_id"])
	}
	if body["pickup_address"] != "123 Main St" {
		t.Errorf("pickup_address = %v, want 123 Main St", body["pickup_address"])
	}
}

func TestGetDriverOffer_NoOffer(t *testing.T) {
	stub := &stubBookingClient{}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/driver/current-offer", nil)
	req = withClaims(req, "driver-1", "driver")
	w := httptest.NewRecorder()
	h.GetDriverOffer(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["has_offer"] != false {
		t.Errorf("has_offer = %v, want false", body["has_offer"])
	}
}

func TestGetDriverOffer_MissingClaims(t *testing.T) {
	stub := &stubBookingClient{}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/driver/current-offer", nil)
	// no withClaims
	w := httptest.NewRecorder()
	h.GetDriverOffer(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ─── CancelRide ───────────────────────────────────────────────────────────────

func TestCancelRide_Success(t *testing.T) {
	stub := &stubBookingClient{
		cancelRide: func(_ context.Context, in *bookingpb.CancelRideRequest, _ ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
			if in.GetTripId() != "trip-1" {
				return nil, status.Error(codes.NotFound, "trip not found")
			}
			return &bookingpb.BookingActionResponse{TripId: "trip-1", Status: "cancelled"}, nil
		},
	}
	h := handlers.NewBookingHandler(stub)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/cancel", nil)
	req.SetPathValue("tripID", "trip-1")
	req = withClaims(req, "rider-1", "rider")
	w := httptest.NewRecorder()

	h.CancelRide(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "cancelled" {
		t.Errorf("status = %q, want cancelled", resp["status"])
	}
}

func TestCancelRide_MissingTripID(t *testing.T) {
	h := handlers.NewBookingHandler(&stubBookingClient{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides//cancel", nil)
	req = withClaims(req, "rider-1", "rider")
	w := httptest.NewRecorder()
	h.CancelRide(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

func TestCancelRide_GRPCError(t *testing.T) {
	stub := &stubBookingClient{
		cancelRide: func(_ context.Context, _ *bookingpb.CancelRideRequest, _ ...grpc.CallOption) (*bookingpb.BookingActionResponse, error) {
			return nil, status.Error(codes.Internal, "trip service down")
		},
	}
	h := handlers.NewBookingHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rides/trip-1/cancel", nil)
	req.SetPathValue("tripID", "trip-1")
	req = withClaims(req, "rider-1", "rider")
	w := httptest.NewRecorder()
	h.CancelRide(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", w.Code)
	}
}

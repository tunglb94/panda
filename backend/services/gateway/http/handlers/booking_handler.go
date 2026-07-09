package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fairride/booking/grpc/bookingpb"
	"github.com/fairride/gateway/http/middleware"
	"google.golang.org/grpc"
)

// BookingClient is the subset of bookingpb.BookingServiceClient used by the gateway.
// Defining it locally keeps the handler unit-testable without a live gRPC connection.
type BookingClient interface {
	BookRide(ctx context.Context, in *bookingpb.BookRideRequest, opts ...grpc.CallOption) (*bookingpb.BookRideResponse, error)
	AcceptDispatchOffer(ctx context.Context, in *bookingpb.AcceptDispatchOfferRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	RejectDispatchOffer(ctx context.Context, in *bookingpb.RejectDispatchOfferRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	StartTrip(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	FinishTrip(ctx context.Context, in *bookingpb.FinishTripRequest, opts ...grpc.CallOption) (*bookingpb.FinishedTripResponse, error)
	GetBookingDetails(ctx context.Context, in *bookingpb.GetBookingDetailsRequest, opts ...grpc.CallOption) (*bookingpb.BookingDetailsResponse, error)
	GetDriverCurrentOffer(ctx context.Context, in *bookingpb.GetDriverCurrentOfferRequest, opts ...grpc.CallOption) (*bookingpb.GetDriverCurrentOfferResponse, error)
	CancelRide(ctx context.Context, in *bookingpb.CancelRideRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	PayRide(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.FinishedTripResponse, error)
}

// BookingHandler exposes booking operations over HTTP.
type BookingHandler struct {
	client BookingClient
}

func NewBookingHandler(client BookingClient) *BookingHandler {
	return &BookingHandler{client: client}
}

// ─── POST /api/v1/rides ───────────────────────────────────────────────────────

type bookRideRequest struct {
	PickupAddress  string  `json:"pickup_address"`
	DropoffAddress string  `json:"dropoff_address"`
	PickupLat      float64 `json:"pickup_lat"`
	PickupLon      float64 `json:"pickup_lon"`
}

func (h *BookingHandler) BookRide(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req bookRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.PickupAddress == "" {
		writeBadRequest(w, "pickup_address is required")
		return
	}
	if req.DropoffAddress == "" {
		writeBadRequest(w, "dropoff_address is required")
		return
	}
	resp, err := h.client.BookRide(r.Context(), &bookingpb.BookRideRequest{
		RiderId:        claims.UserID,
		PickupAddress:  req.PickupAddress,
		DropoffAddress: req.DropoffAddress,
		PickupLat:      req.PickupLat,
		PickupLon:      req.PickupLon,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{
		"trip_id": resp.GetTripId(),
		"status":  resp.GetStatus(),
	})
}

// ─── GET /api/v1/rides/{tripID} ───────────────────────────────────────────────

func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	resp, err := h.client.GetBookingDetails(r.Context(), &bookingpb.GetBookingDetailsRequest{
		TripId: tripID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trip_id":         resp.GetTripId(),
		"trip_status":     resp.GetTripStatus(),
		"rider_id":        resp.GetRiderId(),
		"driver_id":       resp.GetDriverId(),
		"pickup_address":  resp.GetPickupAddress(),
		"dropoff_address": resp.GetDropoffAddress(),
		"dispatch_status": resp.GetDispatchStatus(),
		"final_fare":      resp.GetFinalFare(),
		"currency":        resp.GetCurrency(),
	})
}

// ─── POST /api/v1/rides/{tripID}/accept ──────────────────────────────────────

func (h *BookingHandler) AcceptDispatchOffer(w http.ResponseWriter, r *http.Request) {
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.AcceptDispatchOffer(r.Context(), &bookingpb.AcceptDispatchOfferRequest{
		TripId:   tripID,
		DriverId: claims.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trip_id": resp.GetTripId(),
		"status":  resp.GetStatus(),
	})
}

// ─── POST /api/v1/rides/{tripID}/reject ──────────────────────────────────────

func (h *BookingHandler) RejectDispatchOffer(w http.ResponseWriter, r *http.Request) {
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.RejectDispatchOffer(r.Context(), &bookingpb.RejectDispatchOfferRequest{
		TripId:   tripID,
		DriverId: claims.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trip_id": resp.GetTripId(),
		"status":  resp.GetStatus(),
	})
}

// ─── POST /api/v1/rides/{tripID}/start ───────────────────────────────────────

func (h *BookingHandler) StartTrip(w http.ResponseWriter, r *http.Request) {
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	resp, err := h.client.StartTrip(r.Context(), &bookingpb.StartTripRequest{
		TripId: tripID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trip_id": resp.GetTripId(),
		"status":  resp.GetStatus(),
	})
}

// ─── POST /api/v1/rides/{tripID}/finish ──────────────────────────────────────

type finishTripRequest struct {
	VehicleType string  `json:"vehicle_type"`
	DistanceKM  float64 `json:"distance_km"`
	DurationMin float64 `json:"duration_min"`
}

func (h *BookingHandler) FinishTrip(w http.ResponseWriter, r *http.Request) {
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	var req finishTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.VehicleType == "" {
		writeBadRequest(w, "vehicle_type is required")
		return
	}
	resp, err := h.client.FinishTrip(r.Context(), &bookingpb.FinishTripRequest{
		TripId:      tripID,
		VehicleType: req.VehicleType,
		DistanceKm:  req.DistanceKM,
		DurationMin: req.DurationMin,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trip_id":      resp.GetTripId(),
		"status":       resp.GetStatus(),
		"final_fare":   resp.GetFinalFare(),
		"currency":     resp.GetCurrency(),
		"vehicle_type": resp.GetVehicleType(),
		"distance_km":  resp.GetDistanceKm(),
		"duration_min": resp.GetDurationMin(),
	})
}

// ─── POST /api/v1/rides/{tripID}/cancel ──────────────────────────────────────

func (h *BookingHandler) CancelRide(w http.ResponseWriter, r *http.Request) {
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.CancelRide(r.Context(), &bookingpb.CancelRideRequest{
		TripId:  tripID,
		RiderId: claims.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trip_id": resp.GetTripId(),
		"status":  resp.GetStatus(),
	})
}

// ─── POST /api/v1/rides/{tripID}/pay ─────────────────────────────────────────

type payRideRequest struct {
	PaymentMethod string `json:"payment_method"`
}

func (h *BookingHandler) PayRide(w http.ResponseWriter, r *http.Request) {
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	var req payRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	resp, err := h.client.PayRide(r.Context(), &bookingpb.StartTripRequest{
		TripId: tripID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trip_id":    resp.GetTripId(),
		"status":     resp.GetStatus(),
		"final_fare": resp.GetFinalFare(),
		"currency":   resp.GetCurrency(),
	})
}

// ─── GET /api/v1/driver/current-offer ────────────────────────────────────────

func (h *BookingHandler) GetDriverOffer(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.GetDriverCurrentOffer(r.Context(), &bookingpb.GetDriverCurrentOfferRequest{
		DriverId: claims.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	if !resp.GetHasOffer() {
		writeJSON(w, http.StatusOK, map[string]any{"has_offer": false})
		return
	}
	body := map[string]any{
		"has_offer":       true,
		"trip_id":         resp.GetTripId(),
		"pickup_address":  resp.GetPickupAddress(),
		"dropoff_address": resp.GetDropoffAddress(),
	}
	if ts := resp.GetOfferExpiresAt(); ts != nil {
		body["offer_expires_at"] = ts.AsTime().UTC().Format("2006-01-02T15:04:05Z")
	}
	writeJSON(w, http.StatusOK, body)
}

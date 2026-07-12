package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/fairride/booking/grpc/bookingpb"
	"github.com/fairride/gateway/http/middleware"
	"github.com/fairride/trip/grpc/trippb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// BookingClient is the subset of bookingpb.BookingServiceClient used by the gateway.
// Defining it locally keeps the handler unit-testable without a live gRPC connection.
type BookingClient interface {
	BookRide(ctx context.Context, in *bookingpb.BookRideRequest, opts ...grpc.CallOption) (*bookingpb.BookRideResponse, error)
	AcceptDispatchOffer(ctx context.Context, in *bookingpb.AcceptDispatchOfferRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	RejectDispatchOffer(ctx context.Context, in *bookingpb.RejectDispatchOfferRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	ArriveAtPickup(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	StartTrip(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	FinishTrip(ctx context.Context, in *bookingpb.FinishTripRequest, opts ...grpc.CallOption) (*bookingpb.FinishedTripResponse, error)
	GetBookingDetails(ctx context.Context, in *bookingpb.GetBookingDetailsRequest, opts ...grpc.CallOption) (*bookingpb.BookingDetailsResponse, error)
	GetDriverCurrentOffer(ctx context.Context, in *bookingpb.GetDriverCurrentOfferRequest, opts ...grpc.CallOption) (*bookingpb.GetDriverCurrentOfferResponse, error)
	CancelRide(ctx context.Context, in *bookingpb.CancelRideRequest, opts ...grpc.CallOption) (*bookingpb.BookingActionResponse, error)
	PayRide(ctx context.Context, in *bookingpb.StartTripRequest, opts ...grpc.CallOption) (*bookingpb.FinishedTripResponse, error)
	ListRiderTrips(ctx context.Context, in *bookingpb.ListTripsRequest, opts ...grpc.CallOption) (*bookingpb.TripListResponse, error)
	ListDriverTrips(ctx context.Context, in *bookingpb.ListTripsRequest, opts ...grpc.CallOption) (*bookingpb.TripListResponse, error)
}

// TripStatusClient is the subset of trippb.TripServiceClient used to enrich
// a booking status poll with delivery fields that bookingpb.BookingDetailsResponse
// doesn't carry (Booking's proto was never extended with trip_type/delivery_id/
// delivery_status). Optional — nil-safe, see GetBooking.
type TripStatusClient interface {
	GetTrip(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error)
}

// BookingHandler exposes booking operations over HTTP.
type BookingHandler struct {
	client     BookingClient
	tripClient TripStatusClient
	notifier   *TripEventNotifier
}

func NewBookingHandler(client BookingClient, tripClient TripStatusClient) *BookingHandler {
	return &BookingHandler{client: client, tripClient: tripClient}
}

// SetNotifier wires the Communication Module's best-effort trip-event
// notifications (Part 3). Additive and optional — nil (the default) means
// no notifications fire, identical to this handler's pre-Communication-Module
// behavior.
func (h *BookingHandler) SetNotifier(n *TripEventNotifier) {
	h.notifier = n
}

// ─── POST /api/v1/rides ───────────────────────────────────────────────────────

type bookRideRequest struct {
	PickupAddress  string  `json:"pickup_address"`
	DropoffAddress string  `json:"dropoff_address"`
	PickupLat      float64 `json:"pickup_lat"`
	PickupLon      float64 `json:"pickup_lon"`

	// Delivery fields — all optional. Omitted/empty means a plain ride
	// booking (TripType defaults to "ride" server-side when unset).
	TripType string `json:"trip_type,omitempty"`

	// ServiceType is one of the Vehicle/Service Catalog's 4 tiers
	// (bike/bike_plus/car/car_xl) — optional, empty means no service-type
	// filter (matches every caller written before this catalog existed).
	// Carried to Booking as gRPC metadata rather than a new
	// BookRideRequest field — see this handler's BookRide for why.
	ServiceType string `json:"service_type,omitempty"`

	PickupContactName  string  `json:"pickup_contact_name,omitempty"`
	PickupContactPhone string  `json:"pickup_contact_phone,omitempty"`
	ReceiverName       string  `json:"receiver_name,omitempty"`
	ReceiverPhone      string  `json:"receiver_phone,omitempty"`
	PackageNote        string  `json:"package_note,omitempty"`
	PackageValue       int64   `json:"package_value,omitempty"`
	PackageWeight      float64 `json:"package_weight,omitempty"`
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
	// service_type is NOT a field on BookRideRequest — adding one would
	// require regenerating the message's compiled descriptor, which this
	// environment has no protoc/buf toolchain for. Carried as outgoing
	// gRPC metadata instead (read back by booking/grpc/handler.go's
	// BookRide) — needs no schema change and is equally wire-safe.
	ctx := r.Context()
	if req.ServiceType != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-service-type", req.ServiceType)
	}
	resp, err := h.client.BookRide(ctx, &bookingpb.BookRideRequest{
		RiderId:            claims.UserID,
		PickupAddress:      req.PickupAddress,
		DropoffAddress:     req.DropoffAddress,
		PickupLat:          req.PickupLat,
		PickupLon:          req.PickupLon,
		TripType:           req.TripType,
		PickupContactName:  req.PickupContactName,
		PickupContactPhone: req.PickupContactPhone,
		ReceiverName:       req.ReceiverName,
		ReceiverPhone:      req.ReceiverPhone,
		PackageNote:        req.PackageNote,
		PackageValue:       req.PackageValue,
		PackageWeight:      req.PackageWeight,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{
		"trip_id":     resp.GetTripId(),
		"status":      resp.GetStatus(),
		"delivery_id": resp.GetDeliveryId(),
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
	body := map[string]any{
		"trip_id":         resp.GetTripId(),
		"trip_status":     resp.GetTripStatus(),
		"rider_id":        resp.GetRiderId(),
		"driver_id":       resp.GetDriverId(),
		"pickup_address":  resp.GetPickupAddress(),
		"dropoff_address": resp.GetDropoffAddress(),
		"dispatch_status": resp.GetDispatchStatus(),
		"final_fare":      resp.GetFinalFare(),
		"currency":        resp.GetCurrency(),
	}
	// bookingpb.BookingDetailsResponse has no trip_type/delivery_id/
	// delivery_status fields (Booking's proto was never extended for
	// delivery) — best-effort enrich from Trip service directly when
	// available. Never fails the request: a delivery-status lookup error
	// just means those 3 keys stay absent from the response.
	if h.tripClient != nil {
		if tripResp, tripErr := h.tripClient.GetTrip(r.Context(), &trippb.GetTripRequest{TripId: tripID}); tripErr == nil {
			body["trip_type"] = tripResp.GetTrip().GetTripType()
			body["delivery_id"] = tripResp.GetTrip().GetDeliveryId()
			body["delivery_status"] = tripResp.GetTrip().GetDeliveryStatus()
		}
	}
	writeJSON(w, http.StatusOK, body)
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
	if h.notifier != nil {
		h.notifier.Notify(r.Context(), tripID, "accepted")
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
	if h.notifier != nil {
		h.notifier.Notify(r.Context(), tripID, "started")
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
	if h.notifier != nil {
		h.notifier.Notify(r.Context(), tripID, "finished")
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
	if h.notifier != nil {
		h.notifier.Notify(r.Context(), tripID, "cancelled")
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

// ─── POST /api/v1/rides/{tripID}/arrive ──────────────────────────────────────

func (h *BookingHandler) ArriveAtPickup(w http.ResponseWriter, r *http.Request) {
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	resp, err := h.client.ArriveAtPickup(r.Context(), &bookingpb.StartTripRequest{
		TripId: tripID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	if h.notifier != nil {
		h.notifier.Notify(r.Context(), tripID, "arrived")
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trip_id": resp.GetTripId(),
		"status":  resp.GetStatus(),
	})
}

// listTripItem is the JSON shape of one entry in /api/v1/rider/trips and
// /api/v1/driver/trips.
type listTripItem struct {
	TripID         string `json:"trip_id"`
	Status         string `json:"status"`
	PickupAddress  string `json:"pickup_address"`
	DropoffAddress string `json:"dropoff_address"`
	FinalFare      int64  `json:"final_fare"`
	Currency       string `json:"currency"`
	CreatedAt      string `json:"created_at"`
	// Best-effort — see enrichTripDetails. Empty means "ride" (the
	// default) or that the enrichment lookup for this one trip failed/was
	// skipped.
	TripType string `json:"trip_type,omitempty"`
	// Empty for a Ride trip or when the enrichment lookup failed/was
	// skipped.
	DeliveryStatus string `json:"delivery_status,omitempty"`
}

// tripDetails is the per-trip enrichment fetched from the Trip service
// directly for a list of trips — see enrichTripDetails.
type tripDetails struct {
	TripType       string
	DeliveryStatus string
}

// enrichTripDetails best-effort fetches trip_type/delivery_status for each
// id via the Trip service directly (bookingpb.TripSummaryProto has neither
// field — see TripStatusClient's doc comment on GetBooking for the same gap
// on a single trip). Concurrent since a history page can list dozens of
// trips; an individual lookup failure just leaves that trip's fields empty,
// never fails the whole list.
func (h *BookingHandler) enrichTripDetails(ctx context.Context, tripIDs []string) map[string]tripDetails {
	result := make(map[string]tripDetails, len(tripIDs))
	if h.tripClient == nil {
		return result
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, id := range tripIDs {
		wg.Add(1)
		go func(tripID string) {
			defer wg.Done()
			resp, err := h.tripClient.GetTrip(ctx, &trippb.GetTripRequest{TripId: tripID})
			if err != nil {
				return
			}
			mu.Lock()
			result[tripID] = tripDetails{
				TripType:       resp.GetTrip().GetTripType(),
				DeliveryStatus: resp.GetTrip().GetDeliveryStatus(),
			}
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	return result
}

// ─── GET /api/v1/rider/trips ──────────────────────────────────────────────────

func (h *BookingHandler) ListRiderTrips(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.ListRiderTrips(r.Context(), &bookingpb.ListTripsRequest{
		PartyId: claims.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	trips := resp.GetTrips()
	tripIDs := make([]string, len(trips))
	for i, t := range trips {
		tripIDs[i] = t.GetTripId()
	}
	details := h.enrichTripDetails(r.Context(), tripIDs)
	items := make([]listTripItem, len(trips))
	for i, t := range trips {
		var createdAt string
		if ts := t.GetCreatedAt(); ts != nil {
			createdAt = ts.AsTime().UTC().Format("2006-01-02T15:04:05Z")
		}
		items[i] = listTripItem{
			TripID:         t.GetTripId(),
			Status:         t.GetStatus(),
			PickupAddress:  t.GetPickupAddress(),
			DropoffAddress: t.GetDropoffAddress(),
			FinalFare:      t.GetFinalFare(),
			Currency:       t.GetCurrency(),
			CreatedAt:      createdAt,
			TripType:       details[t.GetTripId()].TripType,
			DeliveryStatus: details[t.GetTripId()].DeliveryStatus,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"trips": items})
}

// ─── GET /api/v1/driver/trips ─────────────────────────────────────────────────

func (h *BookingHandler) ListDriverTrips(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.ListDriverTrips(r.Context(), &bookingpb.ListTripsRequest{
		PartyId: claims.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	trips := resp.GetTrips()
	tripIDs := make([]string, len(trips))
	for i, t := range trips {
		tripIDs[i] = t.GetTripId()
	}
	details := h.enrichTripDetails(r.Context(), tripIDs)
	items := make([]listTripItem, len(trips))
	for i, t := range trips {
		var createdAt string
		if ts := t.GetCreatedAt(); ts != nil {
			createdAt = ts.AsTime().UTC().Format("2006-01-02T15:04:05Z")
		}
		items[i] = listTripItem{
			TripID:         t.GetTripId(),
			Status:         t.GetStatus(),
			PickupAddress:  t.GetPickupAddress(),
			DropoffAddress: t.GetDropoffAddress(),
			FinalFare:      t.GetFinalFare(),
			Currency:       t.GetCurrency(),
			CreatedAt:      createdAt,
			TripType:       details[t.GetTripId()].TripType,
			DeliveryStatus: details[t.GetTripId()].DeliveryStatus,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"trips": items})
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
	// GetDriverCurrentOfferResponse has no trip_type field — same
	// best-effort Trip-service enrichment as GetBooking (see
	// TripStatusClient's doc comment), so the driver app can render a
	// distinct Delivery offer card without guessing.
	if h.tripClient != nil {
		if tripResp, tripErr := h.tripClient.GetTrip(r.Context(), &trippb.GetTripRequest{TripId: resp.GetTripId()}); tripErr == nil {
			body["trip_type"] = tripResp.GetTrip().GetTripType()
		}
	}
	writeJSON(w, http.StatusOK, body)
}

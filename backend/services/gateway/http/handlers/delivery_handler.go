package handlers

import (
	"context"
	"net/http"

	notificationapp "github.com/fairride/notification/app"
	"github.com/fairride/trip/grpc/trippb"
	"google.golang.org/grpc"
)

// DeliveryTripClient is the subset of trippb.TripServiceClient used for the
// three delivery-lifecycle actions. These RPCs live on the Trip service
// itself (see PickupParcelUseCase/StartDeliveryUseCase/CompleteDeliveryUseCase),
// not Booking — Booking's proto never gained equivalent delivery RPCs, so the
// gateway talks to Trip directly here, mirroring how it already dials
// Driver/Dispatch/Review directly for their own concerns.
type DeliveryTripClient interface {
	PickupParcel(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error)
	StartDelivery(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error)
	CompleteDelivery(ctx context.Context, in *trippb.GetTripRequest, opts ...grpc.CallOption) (*trippb.TripResponse, error)
}

// DeliveryHandler exposes the driver-side delivery-lifecycle actions
// (confirm pickup / start delivery / complete delivery) over HTTP. Nil-safe:
// if the Trip service address wasn't configured at boot, every handler
// returns 503 rather than panicking, matching the gateway's existing
// pattern for optional service dependencies (see AvailabilityHandler,
// LocationHandler, DriverProfileHandler, RatingHandler).
type DeliveryHandler struct {
	client   DeliveryTripClient
	notifier *TripEventNotifier
}

func NewDeliveryHandler(client DeliveryTripClient) *DeliveryHandler {
	return &DeliveryHandler{client: client}
}

// SetNotifier wires the Communication Module's best-effort trip-event
// notifications (Part 3). Additive and optional — see BookingHandler.SetNotifier.
func (h *DeliveryHandler) SetNotifier(n *TripEventNotifier) {
	h.notifier = n
}

// notifySnapshot builds a notificationapp.TripSnapshot straight from a
// trippb.TripResponse that this handler's own RPCs already returned,
// avoiding a second GetTrip round-trip.
func notifySnapshotFromTripResponse(resp *trippb.TripResponse) notificationapp.TripSnapshot {
	t := resp.GetTrip()
	tripType := t.GetTripType()
	if tripType == "" {
		tripType = "ride"
	}
	return notificationapp.TripSnapshot{
		TripID:   t.GetTripId(),
		RiderID:  t.GetRiderId(),
		DriverID: t.GetDriverId(),
		Status:   t.GetStatus(),
		TripType: tripType,
	}
}

func (h *DeliveryHandler) unavailable(w http.ResponseWriter) bool {
	if h.client != nil {
		return false
	}
	writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "delivery service unavailable"})
	return true
}

// ─── POST /api/v1/rides/{tripID}/pickup-parcel ───────────────────────────────

func (h *DeliveryHandler) PickupParcel(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	resp, err := h.client.PickupParcel(r.Context(), &trippb.GetTripRequest{TripId: tripID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	if h.notifier != nil {
		h.notifier.NotifyFromSnapshot(r.Context(), notifySnapshotFromTripResponse(resp), "pickup_parcel")
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trip_id":         resp.GetTrip().GetTripId(),
		"status":          resp.GetTrip().GetStatus(),
		"delivery_status": resp.GetTrip().GetDeliveryStatus(),
	})
}

// ─── POST /api/v1/rides/{tripID}/start-delivery ──────────────────────────────

func (h *DeliveryHandler) StartDelivery(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	resp, err := h.client.StartDelivery(r.Context(), &trippb.GetTripRequest{TripId: tripID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	if h.notifier != nil {
		h.notifier.NotifyFromSnapshot(r.Context(), notifySnapshotFromTripResponse(resp), "start_delivery")
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trip_id":         resp.GetTrip().GetTripId(),
		"status":          resp.GetTrip().GetStatus(),
		"delivery_status": resp.GetTrip().GetDeliveryStatus(),
	})
}

// ─── POST /api/v1/rides/{tripID}/complete-delivery ───────────────────────────

func (h *DeliveryHandler) CompleteDelivery(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	tripID := r.PathValue("tripID")
	if tripID == "" {
		writeBadRequest(w, "trip_id is required")
		return
	}
	resp, err := h.client.CompleteDelivery(r.Context(), &trippb.GetTripRequest{TripId: tripID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	if h.notifier != nil {
		h.notifier.NotifyFromSnapshot(r.Context(), notifySnapshotFromTripResponse(resp), "complete_delivery")
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"trip_id":         resp.GetTrip().GetTripId(),
		"status":          resp.GetTrip().GetStatus(),
		"delivery_status": resp.GetTrip().GetDeliveryStatus(),
	})
}

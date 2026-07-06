package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/fairride/driver/grpc/driverpb"
	"github.com/fairride/gateway/http/middleware"
	"google.golang.org/grpc"
)

// AvailabilityClient is the subset of driverpb.DriverAvailabilityServiceClient used by the gateway.
type AvailabilityClient interface {
	GoOnline(ctx context.Context, in *driverpb.GoOnlineRequest, opts ...grpc.CallOption) (*driverpb.AvailabilityResponse, error)
	GoOffline(ctx context.Context, in *driverpb.GoOfflineRequest, opts ...grpc.CallOption) (*driverpb.AvailabilityResponse, error)
	GetAvailability(ctx context.Context, in *driverpb.GetAvailabilityRequest, opts ...grpc.CallOption) (*driverpb.AvailabilityResponse, error)
}

// AvailabilityHandler exposes driver availability operations over HTTP.
type AvailabilityHandler struct {
	client AvailabilityClient
}

// NewAvailabilityHandler constructs an AvailabilityHandler.
// Passing nil for client makes all endpoints return 503.
func NewAvailabilityHandler(client AvailabilityClient) *AvailabilityHandler {
	return &AvailabilityHandler{client: client}
}

// GoOnline handles POST /api/v1/driver/go-online.
func (h *AvailabilityHandler) GoOnline(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "availability service not configured"})
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.GoOnline(r.Context(), &driverpb.GoOnlineRequest{DriverId: claims.UserID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAvailabilityMap(resp))
}

// GoOffline handles POST /api/v1/driver/go-offline.
func (h *AvailabilityHandler) GoOffline(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "availability service not configured"})
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.GoOffline(r.Context(), &driverpb.GoOfflineRequest{DriverId: claims.UserID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAvailabilityMap(resp))
}

// GetAvailability handles GET /api/v1/driver/availability.
func (h *AvailabilityHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "availability service not configured"})
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	resp, err := h.client.GetAvailability(r.Context(), &driverpb.GetAvailabilityRequest{DriverId: claims.UserID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAvailabilityMap(resp))
}

func toAvailabilityMap(resp *driverpb.AvailabilityResponse) map[string]any {
	m := map[string]any{
		"driver_id": resp.GetDriverId(),
		"is_online": resp.GetIsOnline(),
	}
	if ts := resp.GetLastSeen(); ts != nil {
		m["last_seen"] = ts.AsTime().UTC().Format(time.RFC3339)
	}
	return m
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fairride/dispatch/grpc/dispatchpb"
	"github.com/fairride/gateway/http/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// DispatchLocationClient is the subset of dispatchpb.DispatchServiceClient used
// by the gateway for driver location operations.
type DispatchLocationClient interface {
	UpdateDriverLocation(ctx context.Context, in *dispatchpb.UpdateDriverLocationRequest, opts ...grpc.CallOption) (*dispatchpb.UpdateDriverLocationResponse, error)
	GetDriverLocation(ctx context.Context, in *dispatchpb.GetDriverLocationRequest, opts ...grpc.CallOption) (*dispatchpb.GetDriverLocationResponse, error)
}

// LocationHandler exposes driver location operations over HTTP.
type LocationHandler struct {
	client DispatchLocationClient
}

// NewLocationHandler constructs a LocationHandler.
// Passing nil for client makes all endpoints return 503.
func NewLocationHandler(client DispatchLocationClient) *LocationHandler {
	return &LocationHandler{client: client}
}

type updateLocationRequest struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
	// ServiceType is optional (Vehicle/Service Catalog refactor) — one of
	// the 4 product tiers ("bike"/"bike_plus"/"car"/"car_xl") or empty.
	// Omitted/empty is fully backward compatible: dispatch treats "not
	// reported" as "no service-type filter" for this driver, exactly as
	// before this field existed. Carried to Dispatch as gRPC metadata
	// rather than a new UpdateDriverLocationRequest field — avoids
	// touching the generated proto message at all.
	ServiceType string `json:"service_type"`
	// RideEnabled/DeliveryEnabled are the driver's current trip-type
	// capability. Both are pointers so "omitted" is distinguishable from
	// "explicitly false" — an omitted RideEnabled defaults to true and an
	// omitted DeliveryEnabled defaults to false (migration 008's DB column
	// defaults), so a client that predates this field keeps matching Ride
	// jobs exactly as before and is simply not yet Delivery-eligible.
	RideEnabled     *bool `json:"ride_enabled"`
	DeliveryEnabled *bool `json:"delivery_enabled"`
}

// UpdateLocation handles POST /api/v1/driver/location.
// The driver ID is taken from the JWT claims; the body supplies lat/lon.
func (h *LocationHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "dispatch service not configured"})
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req updateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	ctx := r.Context()
	if req.ServiceType != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-service-type", req.ServiceType)
	}
	rideEnabled := req.RideEnabled == nil || *req.RideEnabled             // default true
	deliveryEnabled := req.DeliveryEnabled != nil && *req.DeliveryEnabled // default false
	ctx = metadata.AppendToOutgoingContext(ctx, "x-ride-enabled", boolMetadataValue(rideEnabled))
	ctx = metadata.AppendToOutgoingContext(ctx, "x-delivery-enabled", boolMetadataValue(deliveryEnabled))
	if _, err := h.client.UpdateDriverLocation(ctx, &dispatchpb.UpdateDriverLocationRequest{
		DriverId: claims.UserID,
		Lat:      req.Lat,
		Lon:      req.Lon,
	}); err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func boolMetadataValue(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// GetLocation handles GET /api/v1/driver/{driverID}/location.
// Returns the driver's last known coordinates and whether they are still active.
func (h *LocationHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "dispatch service not configured"})
		return
	}
	driverID := r.PathValue("driverID")
	if driverID == "" {
		writeBadRequest(w, "driver_id is required")
		return
	}
	resp, err := h.client.GetDriverLocation(r.Context(), &dispatchpb.GetDriverLocationRequest{
		DriverId: driverID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"driver_id": driverID,
		"lat":       resp.GetLat(),
		"lon":       resp.GetLon(),
		"is_active": resp.GetIsActive(),
	})
}

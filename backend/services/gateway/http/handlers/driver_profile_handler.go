package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/fairride/driver/grpc/driverpb"
	"google.golang.org/grpc"
)

// DriverProfileClient is the subset of driverpb.DriverProfileServiceClient used by the gateway.
type DriverProfileClient interface {
	GetDriverProfile(ctx context.Context, in *driverpb.GetDriverProfileRequest, opts ...grpc.CallOption) (*driverpb.GetDriverProfileResponse, error)
}

// DriverProfileHandler exposes driver profile reads over HTTP.
type DriverProfileHandler struct {
	client DriverProfileClient
}

// NewDriverProfileHandler constructs a DriverProfileHandler.
// Passing nil for client makes all endpoints return 503.
func NewDriverProfileHandler(client DriverProfileClient) *DriverProfileHandler {
	return &DriverProfileHandler{client: client}
}

// GetDriverProfile handles GET /api/v1/drivers/{driverID}/profile.
func (h *DriverProfileHandler) GetDriverProfile(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "driver service not configured"})
		return
	}
	driverID := r.PathValue("driverID")
	if driverID == "" {
		writeBadRequest(w, "driverID is required")
		return
	}
	resp, err := h.client.GetDriverProfile(r.Context(), &driverpb.GetDriverProfileRequest{
		DriverId: driverID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	p := resp.GetProfile()
	body := map[string]any{
		"driver_id":           p.GetDriverId(),
		"vehicle_type":        p.GetVehicleType(),
		"vehicle_brand":       p.GetVehicleBrand(),
		"vehicle_model":       p.GetVehicleModel(),
		"vehicle_color":       p.GetVehicleColor(),
		"plate_number":        p.GetPlateNumber(),
		"verification_status": p.GetVerificationStatus(),
	}
	if ts := p.GetCreatedAt(); ts != nil {
		body["created_at"] = ts.AsTime().UTC().Format(time.RFC3339)
	}
	writeJSON(w, http.StatusOK, body)
}

package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/grpc/driverpb"
)

// AvailabilityHandler implements driverpb.DriverAvailabilityServiceServer.
type AvailabilityHandler struct {
	driverpb.UnimplementedDriverAvailabilityServiceServer
	goOnline       *app.GoOnlineUseCase
	goOffline      *app.GoOfflineUseCase
	heartbeat      *app.HeartbeatUseCase
	getAvailability *app.GetAvailabilityUseCase
}

// NewAvailabilityHandler constructs an AvailabilityHandler wired to all four use cases.
func NewAvailabilityHandler(
	goOnline *app.GoOnlineUseCase,
	goOffline *app.GoOfflineUseCase,
	heartbeat *app.HeartbeatUseCase,
	getAvailability *app.GetAvailabilityUseCase,
) *AvailabilityHandler {
	if goOnline == nil {
		panic("availability grpc: GoOnlineUseCase must not be nil")
	}
	if goOffline == nil {
		panic("availability grpc: GoOfflineUseCase must not be nil")
	}
	if heartbeat == nil {
		panic("availability grpc: HeartbeatUseCase must not be nil")
	}
	if getAvailability == nil {
		panic("availability grpc: GetAvailabilityUseCase must not be nil")
	}
	return &AvailabilityHandler{
		goOnline:        goOnline,
		goOffline:       goOffline,
		heartbeat:       heartbeat,
		getAvailability: getAvailability,
	}
}

// GoOnline implements DriverAvailabilityServiceServer.GoOnline.
func (h *AvailabilityHandler) GoOnline(ctx context.Context, req *driverpb.GoOnlineRequest) (*driverpb.AvailabilityResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	state, err := h.goOnline.Execute(ctx, req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return availabilityToProto(state), nil
}

// GoOffline implements DriverAvailabilityServiceServer.GoOffline.
func (h *AvailabilityHandler) GoOffline(ctx context.Context, req *driverpb.GoOfflineRequest) (*driverpb.AvailabilityResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	state, err := h.goOffline.Execute(ctx, req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return availabilityToProto(state), nil
}

// Heartbeat implements DriverAvailabilityServiceServer.Heartbeat.
func (h *AvailabilityHandler) Heartbeat(ctx context.Context, req *driverpb.HeartbeatRequest) (*driverpb.AvailabilityResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	state, err := h.heartbeat.Execute(ctx, req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return availabilityToProto(state), nil
}

// GetAvailability implements DriverAvailabilityServiceServer.GetAvailability.
func (h *AvailabilityHandler) GetAvailability(ctx context.Context, req *driverpb.GetAvailabilityRequest) (*driverpb.AvailabilityResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	state, err := h.getAvailability.Execute(ctx, req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return availabilityToProto(state), nil
}

// ─── private helper ───────────────────────────────────────────────────────────

func availabilityToProto(s *entity.AvailabilityState) *driverpb.AvailabilityResponse {
	resp := &driverpb.AvailabilityResponse{
		DriverId: s.DriverID,
		IsOnline: s.IsOnline,
	}
	if !s.LastSeen.IsZero() {
		resp.LastSeen = timestamppb.New(s.LastSeen)
	}
	return resp
}

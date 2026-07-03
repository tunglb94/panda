// Package grpc contains the gRPC handler for the Dispatch service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/dispatch/app"
	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/grpc/dispatchpb"
	domainerrors "github.com/fairride/shared/errors"
)

// Handler implements dispatchpb.DispatchServiceServer.
type Handler struct {
	dispatchpb.UnimplementedDispatchServiceServer
	requestDispatch     *app.RequestDispatchUseCase
	acceptTrip          *app.AcceptTripUseCase
	rejectTrip          *app.RejectTripUseCase
	updateDriverLocation *app.UpdateDriverLocationUseCase
	getDispatchStatus   *app.GetDispatchStatusUseCase
}

// NewHandler wires all five dispatch use cases into a gRPC handler.
func NewHandler(
	requestDispatch *app.RequestDispatchUseCase,
	acceptTrip *app.AcceptTripUseCase,
	rejectTrip *app.RejectTripUseCase,
	updateDriverLocation *app.UpdateDriverLocationUseCase,
	getDispatchStatus *app.GetDispatchStatusUseCase,
) *Handler {
	if requestDispatch == nil {
		panic("dispatch grpc: RequestDispatchUseCase must not be nil")
	}
	if acceptTrip == nil {
		panic("dispatch grpc: AcceptTripUseCase must not be nil")
	}
	if rejectTrip == nil {
		panic("dispatch grpc: RejectTripUseCase must not be nil")
	}
	if updateDriverLocation == nil {
		panic("dispatch grpc: UpdateDriverLocationUseCase must not be nil")
	}
	if getDispatchStatus == nil {
		panic("dispatch grpc: GetDispatchStatusUseCase must not be nil")
	}
	return &Handler{
		requestDispatch:      requestDispatch,
		acceptTrip:           acceptTrip,
		rejectTrip:           rejectTrip,
		updateDriverLocation: updateDriverLocation,
		getDispatchStatus:    getDispatchStatus,
	}
}

// RequestDispatch implements DispatchServiceServer.RequestDispatch.
func (h *Handler) RequestDispatch(ctx context.Context, req *dispatchpb.RequestDispatchRequest) (*dispatchpb.DispatchResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if req.GetRiderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "rider_id is required")
	}
	job, err := h.requestDispatch.Execute(ctx, app.RequestDispatchInput{
		TripID:          req.GetTripId(),
		RiderID:         req.GetRiderId(),
		PickupLat:       req.GetPickupLat(),
		PickupLon:       req.GetPickupLon(),
		OfferTimeoutSec: int(req.GetOfferTimeoutSec()),
		MaxAttempts:     int(req.GetMaxAttempts()),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &dispatchpb.DispatchResponse{Job: toProto(job)}, nil
}

// AcceptTrip implements DispatchServiceServer.AcceptTrip.
func (h *Handler) AcceptTrip(ctx context.Context, req *dispatchpb.AcceptTripRequest) (*dispatchpb.DispatchResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	job, err := h.acceptTrip.Execute(ctx, req.GetTripId(), req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &dispatchpb.DispatchResponse{Job: toProto(job)}, nil
}

// RejectTrip implements DispatchServiceServer.RejectTrip.
func (h *Handler) RejectTrip(ctx context.Context, req *dispatchpb.RejectTripRequest) (*dispatchpb.DispatchResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	job, err := h.rejectTrip.Execute(ctx, req.GetTripId(), req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &dispatchpb.DispatchResponse{Job: toProto(job)}, nil
}

// UpdateDriverLocation implements DispatchServiceServer.UpdateDriverLocation.
func (h *Handler) UpdateDriverLocation(ctx context.Context, req *dispatchpb.UpdateDriverLocationRequest) (*dispatchpb.UpdateDriverLocationResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	if err := h.updateDriverLocation.Execute(ctx, req.GetDriverId(), req.GetLat(), req.GetLon()); err != nil {
		return nil, toGRPCError(err)
	}
	return &dispatchpb.UpdateDriverLocationResponse{}, nil
}

// GetDispatchStatus implements DispatchServiceServer.GetDispatchStatus.
func (h *Handler) GetDispatchStatus(ctx context.Context, req *dispatchpb.GetDispatchStatusRequest) (*dispatchpb.DispatchResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	job, err := h.getDispatchStatus.Execute(ctx, req.GetTripId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &dispatchpb.DispatchResponse{Job: toProto(job)}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func toProto(j *entity.DispatchJob) *dispatchpb.DispatchJobProto {
	p := &dispatchpb.DispatchJobProto{
		TripId:           j.TripID,
		RiderId:          j.RiderID,
		Status:           string(j.Status),
		CurrentDriverId:  j.CurrentDriverID,
		AssignedDriverId: j.AssignedDriverID,
		AttemptCount:     int32(j.AttemptCount),
		MaxAttempts:      int32(j.MaxAttempts),
		CreatedAt:        timestamppb.New(j.CreatedAt),
		UpdatedAt:        timestamppb.New(j.UpdatedAt),
	}
	if !j.OfferExpiresAt.IsZero() {
		p.OfferExpiresAt = timestamppb.New(j.OfferExpiresAt)
	}
	return p
}

func toGRPCError(err error) error {
	code := domainerrors.GetCode(err)
	var grpcCode codes.Code
	switch code {
	case domainerrors.CodeNotFound:
		grpcCode = codes.NotFound
	case domainerrors.CodeInvalidArgument:
		grpcCode = codes.InvalidArgument
	case domainerrors.CodeAlreadyExists:
		grpcCode = codes.AlreadyExists
	case domainerrors.CodePreconditionFailed:
		grpcCode = codes.FailedPrecondition
	case domainerrors.CodeUnauthenticated:
		grpcCode = codes.Unauthenticated
	case domainerrors.CodePermissionDenied:
		grpcCode = codes.PermissionDenied
	case domainerrors.CodeUnavailable:
		grpcCode = codes.Unavailable
	default:
		grpcCode = codes.Internal
	}
	return status.Error(grpcCode, err.Error())
}

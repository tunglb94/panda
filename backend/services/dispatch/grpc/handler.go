// Package grpc contains the gRPC handler for the Dispatch service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
	requestDispatch      *app.RequestDispatchUseCase
	acceptTrip           *app.AcceptTripUseCase
	rejectTrip           *app.RejectTripUseCase
	updateDriverLocation *app.UpdateDriverLocationUseCase
	getDispatchStatus    *app.GetDispatchStatusUseCase
	getDriverOffer       *app.GetDriverOfferUseCase
	getDriverLocation    *app.GetDriverLocationUseCase
}

// NewHandler wires all seven dispatch use cases into a gRPC handler.
func NewHandler(
	requestDispatch *app.RequestDispatchUseCase,
	acceptTrip *app.AcceptTripUseCase,
	rejectTrip *app.RejectTripUseCase,
	updateDriverLocation *app.UpdateDriverLocationUseCase,
	getDispatchStatus *app.GetDispatchStatusUseCase,
	getDriverOffer *app.GetDriverOfferUseCase,
	getDriverLocation *app.GetDriverLocationUseCase,
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
	if getDriverOffer == nil {
		panic("dispatch grpc: GetDriverOfferUseCase must not be nil")
	}
	if getDriverLocation == nil {
		panic("dispatch grpc: GetDriverLocationUseCase must not be nil")
	}
	return &Handler{
		requestDispatch:      requestDispatch,
		acceptTrip:           acceptTrip,
		rejectTrip:           rejectTrip,
		updateDriverLocation: updateDriverLocation,
		getDispatchStatus:    getDispatchStatus,
		getDriverOffer:       getDriverOffer,
		getDriverLocation:    getDriverLocation,
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
	tripType := entity.TripTypeRide
	if req.GetTripType() == string(entity.TripTypeDelivery) {
		tripType = entity.TripTypeDelivery
	}
	job, err := h.requestDispatch.Execute(ctx, app.RequestDispatchInput{
		TripID:          req.GetTripId(),
		RiderID:         req.GetRiderId(),
		PickupLat:       req.GetPickupLat(),
		PickupLon:       req.GetPickupLon(),
		OfferTimeoutSec: int(req.GetOfferTimeoutSec()),
		MaxAttempts:     int(req.GetMaxAttempts()),
		TripType:        tripType,
		// The wire field is still named "vehicle_type" (added during an
		// earlier Delivery phase, before this catalog existed) — its VALUE
		// now carries a ServiceType. Renaming the proto field itself would
		// need a compiled-descriptor change this environment can't safely
		// make (see delivery_fields_smoke_test.go); reinterpreting an
		// existing string field is wire-compatible and needs none.
		ServiceType: entity.ServiceType(req.GetVehicleType()),
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
//
// service_type/ride_enabled/delivery_enabled are NOT fields on
// UpdateDriverLocationRequest — adding one would require regenerating the
// message's compiled descriptor, which this environment has no protoc/buf
// toolchain for. Instead the gateway carries them as outgoing gRPC
// metadata ("x-service-type"/"x-ride-enabled"/"x-delivery-enabled", see
// gateway/http/handlers/location_handler.go), which needs no schema change
// at all and is equally wire-safe. Absent metadata (older gateway build, or
// a driver who hasn't reported yet) resolves to
// serviceType=""/rideEnabled=true/deliveryEnabled=false — matching
// migration 008's DB column defaults — fully backward compatible.
func (h *Handler) UpdateDriverLocation(ctx context.Context, req *dispatchpb.UpdateDriverLocationRequest) (*dispatchpb.UpdateDriverLocationResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	var serviceType entity.ServiceType
	rideEnabled, deliveryEnabled := true, false
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-service-type"); len(vals) > 0 {
			serviceType = entity.ServiceType(vals[0])
		}
		if vals := md.Get("x-ride-enabled"); len(vals) > 0 {
			rideEnabled = vals[0] == "1"
		}
		if vals := md.Get("x-delivery-enabled"); len(vals) > 0 {
			deliveryEnabled = vals[0] == "1"
		}
	}
	if err := h.updateDriverLocation.Execute(ctx, req.GetDriverId(), req.GetLat(), req.GetLon(), serviceType, rideEnabled, deliveryEnabled); err != nil {
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

// GetDriverOffer implements DispatchServiceServer.GetDriverOffer.
func (h *Handler) GetDriverOffer(ctx context.Context, req *dispatchpb.GetDriverOfferRequest) (*dispatchpb.GetDriverOfferResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	job, err := h.getDriverOffer.Execute(ctx, req.GetDriverId())
	if err != nil {
		code := domainerrors.GetCode(err)
		if code == domainerrors.CodeNotFound {
			return &dispatchpb.GetDriverOfferResponse{HasOffer: false}, nil
		}
		return nil, toGRPCError(err)
	}
	resp := &dispatchpb.GetDriverOfferResponse{
		HasOffer: true,
		TripId:   job.TripID,
		JobId:    job.JobID,
	}
	if !job.OfferExpiresAt.IsZero() {
		resp.OfferExpiresAt = timestamppb.New(job.OfferExpiresAt)
	}
	return resp, nil
}

// GetDriverLocation implements DispatchServiceServer.GetDriverLocation.
func (h *Handler) GetDriverLocation(ctx context.Context, req *dispatchpb.GetDriverLocationRequest) (*dispatchpb.GetDriverLocationResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	result, err := h.getDriverLocation.Execute(ctx, req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &dispatchpb.GetDriverLocationResponse{
		Lat:      result.Lat,
		Lon:      result.Lon,
		IsActive: result.IsActive,
	}, nil
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

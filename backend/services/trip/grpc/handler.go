// Package grpc contains the gRPC handler for the Trip service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/trip/app"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/grpc/trippb"
	domainerrors "github.com/fairride/shared/errors"
)

// Handler implements trippb.TripServiceServer.
type Handler struct {
	trippb.UnimplementedTripServiceServer
	createTrip      *app.CreateTripUseCase
	cancelTrip      *app.CancelTripUseCase
	getTrip         *app.GetTripUseCase
	startTrip       *app.StartTripUseCase
	completeTrip    *app.CompleteTripUseCase
	initiatePayment *app.InitiatePaymentUseCase
	payTrip         *app.PayTripUseCase
}

// NewHandler wires all trip use cases into a gRPC handler.
func NewHandler(
	createTrip *app.CreateTripUseCase,
	cancelTrip *app.CancelTripUseCase,
	getTrip *app.GetTripUseCase,
	startTrip *app.StartTripUseCase,
	completeTrip *app.CompleteTripUseCase,
	initiatePayment *app.InitiatePaymentUseCase,
	payTrip *app.PayTripUseCase,
) *Handler {
	if createTrip == nil {
		panic("trip grpc: CreateTripUseCase must not be nil")
	}
	if cancelTrip == nil {
		panic("trip grpc: CancelTripUseCase must not be nil")
	}
	if getTrip == nil {
		panic("trip grpc: GetTripUseCase must not be nil")
	}
	if startTrip == nil {
		panic("trip grpc: StartTripUseCase must not be nil")
	}
	if completeTrip == nil {
		panic("trip grpc: CompleteTripUseCase must not be nil")
	}
	if initiatePayment == nil {
		panic("trip grpc: InitiatePaymentUseCase must not be nil")
	}
	if payTrip == nil {
		panic("trip grpc: PayTripUseCase must not be nil")
	}
	return &Handler{
		createTrip:      createTrip,
		cancelTrip:      cancelTrip,
		getTrip:         getTrip,
		startTrip:       startTrip,
		completeTrip:    completeTrip,
		initiatePayment: initiatePayment,
		payTrip:         payTrip,
	}
}

// CreateTrip implements TripServiceServer.CreateTrip.
func (h *Handler) CreateTrip(ctx context.Context, req *trippb.CreateTripRequest) (*trippb.TripResponse, error) {
	if req.GetRiderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "rider_id is required")
	}
	if req.GetPickupAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "pickup_address is required")
	}
	if req.GetDropoffAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "dropoff_address is required")
	}
	trip, err := h.createTrip.Execute(ctx, app.CreateTripInput{
		RiderID:        req.GetRiderId(),
		PickupAddress:  req.GetPickupAddress(),
		DropoffAddress: req.GetDropoffAddress(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProto(trip)}, nil
}

// CancelTrip implements TripServiceServer.CancelTrip.
func (h *Handler) CancelTrip(ctx context.Context, req *trippb.CancelTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	trip, err := h.cancelTrip.Execute(ctx, app.CancelTripInput{
		TripID: req.GetTripId(),
		Reason: req.GetReason(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProto(trip)}, nil
}

// GetTrip implements TripServiceServer.GetTrip.
func (h *Handler) GetTrip(ctx context.Context, req *trippb.GetTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	trip, err := h.getTrip.Execute(ctx, req.GetTripId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProto(trip)}, nil
}

// StartTrip implements TripServiceServer.StartTrip.
func (h *Handler) StartTrip(ctx context.Context, req *trippb.StartTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	trip, err := h.startTrip.Execute(ctx, app.StartTripInput{TripID: req.GetTripId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProto(trip)}, nil
}

// CompleteTrip implements TripServiceServer.CompleteTrip.
func (h *Handler) CompleteTrip(ctx context.Context, req *trippb.CompleteTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if req.GetFareCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "fare_currency is required")
	}
	trip, err := h.completeTrip.Execute(ctx, app.CompleteTripInput{
		TripID:         req.GetTripId(),
		FinalFareTotal: req.GetFinalFareTotal(),
		FareCurrency:   req.GetFareCurrency(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProto(trip)}, nil
}

// InitiatePayment implements TripServiceServer.InitiatePayment.
func (h *Handler) InitiatePayment(ctx context.Context, req *trippb.GetTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	trip, err := h.initiatePayment.Execute(ctx, req.GetTripId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProto(trip)}, nil
}

// PayTrip implements TripServiceServer.PayTrip.
// req.TripId = trip ID, req.Reason = payment method ("cash"|"wallet").
func (h *Handler) PayTrip(ctx context.Context, req *trippb.CancelTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	trip, err := h.payTrip.Execute(ctx, app.PayTripInput{
		TripID:        req.GetTripId(),
		PaymentMethod: req.GetReason(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProto(trip)}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func toProto(t *entity.Trip) *trippb.TripProto {
	return &trippb.TripProto{
		TripId:             t.TripID,
		RiderId:            t.RiderID,
		DriverId:           t.DriverID,
		Status:             string(t.Status),
		PickupAddress:      t.PickupAddress,
		DropoffAddress:     t.DropoffAddress,
		CancellationReason: t.CancellationReason,
		CreatedAt:          timestamppb.New(t.CreatedAt),
		UpdatedAt:          timestamppb.New(t.UpdatedAt),
		FinalFareTotal:     t.FinalFareTotal,
		FareCurrency:       t.FareCurrency,
	}
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

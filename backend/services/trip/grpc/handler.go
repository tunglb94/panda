// Package grpc contains the gRPC handler for the Trip service.
package grpc

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/app"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/grpc/trippb"
)

// Handler implements trippb.TripServiceServer.
type Handler struct {
	trippb.UnimplementedTripServiceServer
	createTrip        *app.CreateTripUseCase
	cancelTrip        *app.CancelTripUseCase
	getTrip           *app.GetTripUseCase
	markDriverArrived *app.MarkDriverArrivedUseCase
	startTrip         *app.StartTripUseCase
	completeTrip      *app.CompleteTripUseCase
	initiatePayment   *app.InitiatePaymentUseCase
	payTrip           *app.PayTripUseCase
	listTripsByRider  *app.ListTripsByRiderUseCase
	listTripsByDriver *app.ListTripsByDriverUseCase
	// Delivery V1 Phase 4 (docs/business/DELIVERY_V1_DESIGN.md) — additive.
	pickupParcel     *app.PickupParcelUseCase
	startDelivery    *app.StartDeliveryUseCase
	completeDelivery *app.CompleteDeliveryUseCase
	// Production hardening P0-1 — additive.
	acceptDelivery *app.AcceptDeliveryUseCase
}

// NewHandler wires all trip use cases into a gRPC handler.
func NewHandler(
	createTrip *app.CreateTripUseCase,
	cancelTrip *app.CancelTripUseCase,
	getTrip *app.GetTripUseCase,
	markDriverArrived *app.MarkDriverArrivedUseCase,
	startTrip *app.StartTripUseCase,
	completeTrip *app.CompleteTripUseCase,
	initiatePayment *app.InitiatePaymentUseCase,
	payTrip *app.PayTripUseCase,
	listTripsByRider *app.ListTripsByRiderUseCase,
	listTripsByDriver *app.ListTripsByDriverUseCase,
	pickupParcel *app.PickupParcelUseCase,
	startDelivery *app.StartDeliveryUseCase,
	completeDelivery *app.CompleteDeliveryUseCase,
	acceptDelivery *app.AcceptDeliveryUseCase,
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
	if markDriverArrived == nil {
		panic("trip grpc: MarkDriverArrivedUseCase must not be nil")
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
	if listTripsByRider == nil {
		panic("trip grpc: ListTripsByRiderUseCase must not be nil")
	}
	if listTripsByDriver == nil {
		panic("trip grpc: ListTripsByDriverUseCase must not be nil")
	}
	if pickupParcel == nil {
		panic("trip grpc: PickupParcelUseCase must not be nil")
	}
	if startDelivery == nil {
		panic("trip grpc: StartDeliveryUseCase must not be nil")
	}
	if completeDelivery == nil {
		panic("trip grpc: CompleteDeliveryUseCase must not be nil")
	}
	if acceptDelivery == nil {
		panic("trip grpc: AcceptDeliveryUseCase must not be nil")
	}
	return &Handler{
		createTrip:        createTrip,
		cancelTrip:        cancelTrip,
		getTrip:           getTrip,
		markDriverArrived: markDriverArrived,
		startTrip:         startTrip,
		completeTrip:      completeTrip,
		initiatePayment:   initiatePayment,
		payTrip:           payTrip,
		listTripsByRider:  listTripsByRider,
		listTripsByDriver: listTripsByDriver,
		pickupParcel:      pickupParcel,
		startDelivery:     startDelivery,
		completeDelivery:  completeDelivery,
		acceptDelivery:    acceptDelivery,
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
	tripType := entity.TripTypeRide
	if req.GetTripType() == string(entity.TripTypeDelivery) {
		tripType = entity.TripTypeDelivery
	}
	trip, err := h.createTrip.Execute(ctx, app.CreateTripInput{
		RiderID:            req.GetRiderId(),
		PickupAddress:      req.GetPickupAddress(),
		DropoffAddress:     req.GetDropoffAddress(),
		TripType:           tripType,
		PickupContactName:  req.GetPickupContactName(),
		PickupContactPhone: req.GetPickupContactPhone(),
		ReceiverName:       req.GetReceiverName(),
		ReceiverPhone:      req.GetReceiverPhone(),
		PackageNote:        req.GetPackageNote(),
		PackageValue:       req.GetPackageValue(),
		PackageWeightKg:    req.GetPackageWeight(),
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
	setTripFinancialsHeader(ctx, trip)
	return &trippb.TripResponse{Trip: toProto(trip)}, nil
}

// setTripFinancialsHeader exposes PaymentMethod and commission detail as
// OUTGOING response header metadata — TripProto has no field slots for
// either (no protoc/buf toolchain available to add them). Callers that need
// durable, non-invented commission numbers (Gateway's SettlementEngine) read
// this instead of Trip's own FinalFareTotal * a self-chosen rate.
func setTripFinancialsHeader(ctx context.Context, trip *entity.Trip) {
	md := metadata.Pairs("x-payment-method", trip.PaymentMethod)
	if trip.HasCommissionDetail {
		md.Append("x-has-commission-detail", "true")
		md.Append("x-commission-cents", strconv.FormatInt(trip.CommissionCents, 10))
		md.Append("x-driver-income-cents", strconv.FormatInt(trip.DriverIncomeCents, 10))
		md.Append("x-voucher-discount-cents", strconv.FormatInt(trip.VoucherDiscountCents, 10))
		md.Append("x-commission-rate", strconv.FormatFloat(trip.CommissionRate, 'f', -1, 64))
	}
	_ = grpc.SetHeader(ctx, md)
}

// MarkDriverArrived implements TripServiceServer.MarkDriverArrived.
func (h *Handler) MarkDriverArrived(ctx context.Context, req *trippb.GetTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	trip, err := h.markDriverArrived.Execute(ctx, req.GetTripId())
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
	// Commission detail has no field on CompleteTripRequest — carried as
	// incoming metadata from Booking's TripAdapter.CompleteTrip (same
	// proto-extension constraint as x-service-type in booking/grpc/handler.go).
	var fin entity.CompleteFinancials
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-has-commission-detail"); len(vals) > 0 && vals[0] == "true" {
			fin.HasCommissionDetail = true
			if v := md.Get("x-commission-cents"); len(v) > 0 {
				fin.CommissionCents, _ = strconv.ParseInt(v[0], 10, 64)
			}
			if v := md.Get("x-driver-income-cents"); len(v) > 0 {
				fin.DriverIncomeCents, _ = strconv.ParseInt(v[0], 10, 64)
			}
			if v := md.Get("x-voucher-discount-cents"); len(v) > 0 {
				fin.VoucherDiscountCents, _ = strconv.ParseInt(v[0], 10, 64)
			}
			if v := md.Get("x-commission-rate"); len(v) > 0 {
				fin.CommissionRate, _ = strconv.ParseFloat(v[0], 64)
			}
		}
	}
	trip, err := h.completeTrip.Execute(ctx, app.CompleteTripInput{
		TripID:         req.GetTripId(),
		FinalFareTotal: req.GetFinalFareTotal(),
		FareCurrency:   req.GetFareCurrency(),
		Financials:     fin,
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

// ListTripsByRider implements TripServiceServer.ListTripsByRider.
func (h *Handler) ListTripsByRider(ctx context.Context, req *trippb.ListTripsRequest) (*trippb.TripsResponse, error) {
	if req.GetPartyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "party_id (rider_id) is required")
	}
	trips, err := h.listTripsByRider.Execute(ctx, req.GetPartyId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	protos := make([]*trippb.TripProto, len(trips))
	for i, t := range trips {
		protos[i] = toProto(t)
	}
	return &trippb.TripsResponse{Trips: protos}, nil
}

// ListTripsByDriver implements TripServiceServer.ListTripsByDriver.
func (h *Handler) ListTripsByDriver(ctx context.Context, req *trippb.ListTripsRequest) (*trippb.TripsResponse, error) {
	if req.GetPartyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "party_id (driver_id) is required")
	}
	trips, err := h.listTripsByDriver.Execute(ctx, req.GetPartyId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	protos := make([]*trippb.TripProto, len(trips))
	for i, t := range trips {
		protos[i] = toProto(t)
	}
	return &trippb.TripsResponse{Trips: protos}, nil
}

// PickupParcel implements TripServiceServer.PickupParcel. Delivery V1
// Phase 4 (docs/business/DELIVERY_V1_DESIGN.md) — returns InvalidArgument
// for a Ride trip (via the use case's loadDeliveryTrip guard).
func (h *Handler) PickupParcel(ctx context.Context, req *trippb.GetTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	delivery, err := h.pickupParcel.Execute(ctx, app.PickupParcelInput{TripID: req.GetTripId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	trip, err := h.getTrip.Execute(ctx, req.GetTripId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProtoWithDeliveryStatus(trip, delivery.Status)}, nil
}

// StartDelivery implements TripServiceServer.StartDelivery.
func (h *Handler) StartDelivery(ctx context.Context, req *trippb.GetTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	delivery, err := h.startDelivery.Execute(ctx, app.StartDeliveryInput{TripID: req.GetTripId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	trip, err := h.getTrip.Execute(ctx, req.GetTripId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProtoWithDeliveryStatus(trip, delivery.Status)}, nil
}

// CompleteDelivery implements TripServiceServer.CompleteDelivery.
func (h *Handler) CompleteDelivery(ctx context.Context, req *trippb.GetTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	delivery, err := h.completeDelivery.Execute(ctx, app.CompleteDeliveryInput{TripID: req.GetTripId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	trip, err := h.getTrip.Execute(ctx, req.GetTripId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &trippb.TripResponse{Trip: toProtoWithDeliveryStatus(trip, delivery.Status)}, nil
}

// AcceptDelivery implements TripServiceServer.AcceptDelivery. Production
// hardening P0-1. Unlike PickupParcel/StartDelivery/CompleteDelivery, a
// Ride trip is not an error here — Booking calls this unconditionally
// after every accept, so a nil *entity.Delivery (nothing to do) falls back
// to the trip's plain proto with delivery_status left at its empty zero
// value, same as every non-delivery-lifecycle RPC.
func (h *Handler) AcceptDelivery(ctx context.Context, req *trippb.GetTripRequest) (*trippb.TripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	delivery, err := h.acceptDelivery.Execute(ctx, app.AcceptDeliveryInput{TripID: req.GetTripId()})
	if err != nil {
		return nil, toGRPCError(err)
	}
	trip, err := h.getTrip.Execute(ctx, req.GetTripId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if delivery == nil {
		return &trippb.TripResponse{Trip: toProto(trip)}, nil
	}
	return &trippb.TripResponse{Trip: toProtoWithDeliveryStatus(trip, delivery.Status)}, nil
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
		TripType:           string(t.TripType),
		DeliveryId:         t.DeliveryID,
	}
}

// toProtoWithDeliveryStatus is toProto plus the given DeliveryStatus.
// Delivery V1 Phase 4 — used only by the 3 new delivery-lifecycle RPCs,
// which already have the just-mutated *entity.Delivery in hand; existing
// RPCs (GetTrip, CreateTrip, ...) keep calling plain toProto unchanged and
// leave delivery_status at its empty zero value, per
// docs/business/DELIVERY_V1_DESIGN.md's additive-only principle.
func toProtoWithDeliveryStatus(t *entity.Trip, deliveryStatus entity.DeliveryStatus) *trippb.TripProto {
	p := toProto(t)
	p.DeliveryStatus = string(deliveryStatus)
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

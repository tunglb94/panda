// Package grpc contains the gRPC handler for the Booking service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/booking/app"
	"github.com/fairride/booking/grpc/bookingpb"
)

// Handler implements bookingpb.BookingServiceServer.
type Handler struct {
	bookingpb.UnimplementedBookingServiceServer
	bookRide       *app.BookRideUseCase
	acceptOffer    *app.AcceptDispatchOfferUseCase
	rejectOffer    *app.RejectDispatchOfferUseCase
	startTrip      *app.StartTripUseCase
	finishTrip     *app.FinishTripUseCase
	getDetails     *app.GetBookingDetailsUseCase
	getDriverOffer *app.GetDriverCurrentOfferUseCase
}

// NewHandler wires all seven booking use cases into a gRPC handler.
func NewHandler(
	bookRide *app.BookRideUseCase,
	acceptOffer *app.AcceptDispatchOfferUseCase,
	rejectOffer *app.RejectDispatchOfferUseCase,
	startTrip *app.StartTripUseCase,
	finishTrip *app.FinishTripUseCase,
	getDetails *app.GetBookingDetailsUseCase,
	getDriverOffer *app.GetDriverCurrentOfferUseCase,
) *Handler {
	return &Handler{
		bookRide:       bookRide,
		acceptOffer:    acceptOffer,
		rejectOffer:    rejectOffer,
		startTrip:      startTrip,
		finishTrip:     finishTrip,
		getDetails:     getDetails,
		getDriverOffer: getDriverOffer,
	}
}

// BookRide implements BookingServiceServer.BookRide.
func (h *Handler) BookRide(ctx context.Context, req *bookingpb.BookRideRequest) (*bookingpb.BookRideResponse, error) {
	if req.GetRiderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "rider_id is required")
	}
	if req.GetPickupAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "pickup_address is required")
	}
	if req.GetDropoffAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "dropoff_address is required")
	}
	result, err := h.bookRide.Execute(ctx, app.BookRideInput{
		RiderID:        req.GetRiderId(),
		PickupAddress:  req.GetPickupAddress(),
		DropoffAddress: req.GetDropoffAddress(),
		PickupLat:      req.GetPickupLat(),
		PickupLon:      req.GetPickupLon(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.BookRideResponse{TripId: result.TripID, Status: result.Status}, nil
}

// AcceptDispatchOffer implements BookingServiceServer.AcceptDispatchOffer.
func (h *Handler) AcceptDispatchOffer(ctx context.Context, req *bookingpb.AcceptDispatchOfferRequest) (*bookingpb.BookingActionResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	if err := h.acceptOffer.Execute(ctx, req.GetTripId(), req.GetDriverId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.BookingActionResponse{TripId: req.GetTripId(), Status: "driver_assigned"}, nil
}

// RejectDispatchOffer implements BookingServiceServer.RejectDispatchOffer.
func (h *Handler) RejectDispatchOffer(ctx context.Context, req *bookingpb.RejectDispatchOfferRequest) (*bookingpb.BookingActionResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	if err := h.rejectOffer.Execute(ctx, req.GetTripId(), req.GetDriverId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.BookingActionResponse{TripId: req.GetTripId(), Status: "searching"}, nil
}

// StartTrip implements BookingServiceServer.StartTrip.
func (h *Handler) StartTrip(ctx context.Context, req *bookingpb.StartTripRequest) (*bookingpb.BookingActionResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if err := h.startTrip.Execute(ctx, req.GetTripId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.BookingActionResponse{TripId: req.GetTripId(), Status: "in_progress"}, nil
}

// FinishTrip implements BookingServiceServer.FinishTrip.
func (h *Handler) FinishTrip(ctx context.Context, req *bookingpb.FinishTripRequest) (*bookingpb.FinishedTripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if req.GetVehicleType() == "" {
		return nil, status.Error(codes.InvalidArgument, "vehicle_type is required")
	}
	result, err := h.finishTrip.Execute(ctx, app.FinishTripInput{
		TripID:      req.GetTripId(),
		VehicleType: req.GetVehicleType(),
		DistanceKM:  req.GetDistanceKm(),
		DurationMin: req.GetDurationMin(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.FinishedTripResponse{
		TripId:      result.TripID,
		Status:      result.Status,
		FinalFare:   result.FinalFare,
		Currency:    result.Currency,
		VehicleType: result.VehicleType,
		DistanceKm:  result.DistanceKM,
		DurationMin: result.DurationMin,
	}, nil
}

// GetBookingDetails implements BookingServiceServer.GetBookingDetails.
func (h *Handler) GetBookingDetails(ctx context.Context, req *bookingpb.GetBookingDetailsRequest) (*bookingpb.BookingDetailsResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	details, err := h.getDetails.Execute(ctx, req.GetTripId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.BookingDetailsResponse{
		TripId:         details.TripID,
		TripStatus:     details.TripStatus,
		RiderId:        details.RiderID,
		DriverId:       details.DriverID,
		PickupAddress:  details.PickupAddress,
		DropoffAddress: details.DropoffAddress,
		DispatchStatus: details.DispatchStatus,
		FinalFare:      details.FinalFare,
		Currency:       details.Currency,
	}, nil
}

// GetDriverCurrentOffer implements BookingServiceServer.GetDriverCurrentOffer.
func (h *Handler) GetDriverCurrentOffer(ctx context.Context, req *bookingpb.GetDriverCurrentOfferRequest) (*bookingpb.GetDriverCurrentOfferResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	offer, err := h.getDriverOffer.Execute(ctx, req.GetDriverId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if offer == nil {
		return &bookingpb.GetDriverCurrentOfferResponse{HasOffer: false}, nil
	}
	resp := &bookingpb.GetDriverCurrentOfferResponse{
		HasOffer:       true,
		TripId:         offer.TripID,
		PickupAddress:  offer.PickupAddress,
		DropoffAddress: offer.DropoffAddress,
	}
	if !offer.OfferExpiresAt.IsZero() {
		resp.OfferExpiresAt = timestamppb.New(offer.OfferExpiresAt)
	}
	return resp, nil
}

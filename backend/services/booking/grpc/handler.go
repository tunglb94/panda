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
	bookRide        *app.BookRideUseCase
	acceptOffer     *app.AcceptDispatchOfferUseCase
	rejectOffer     *app.RejectDispatchOfferUseCase
	arriveAtPickup  *app.ArriveAtPickupUseCase
	startTrip       *app.StartTripUseCase
	finishTrip      *app.FinishTripUseCase
	getDetails      *app.GetBookingDetailsUseCase
	getDriverOffer  *app.GetDriverCurrentOfferUseCase
	cancelRide      *app.CancelRideUseCase
	payRide         *app.PayRideUseCase
	listRiderTrips  *app.ListRiderTripsUseCase
	listDriverTrips *app.ListDriverTripsUseCase
}

// NewHandler wires all booking use cases into a gRPC handler.
func NewHandler(
	bookRide *app.BookRideUseCase,
	acceptOffer *app.AcceptDispatchOfferUseCase,
	rejectOffer *app.RejectDispatchOfferUseCase,
	arriveAtPickup *app.ArriveAtPickupUseCase,
	startTrip *app.StartTripUseCase,
	finishTrip *app.FinishTripUseCase,
	getDetails *app.GetBookingDetailsUseCase,
	getDriverOffer *app.GetDriverCurrentOfferUseCase,
	cancelRide *app.CancelRideUseCase,
	payRide *app.PayRideUseCase,
	listRiderTrips *app.ListRiderTripsUseCase,
	listDriverTrips *app.ListDriverTripsUseCase,
) *Handler {
	return &Handler{
		bookRide:        bookRide,
		acceptOffer:     acceptOffer,
		rejectOffer:     rejectOffer,
		arriveAtPickup:  arriveAtPickup,
		startTrip:       startTrip,
		finishTrip:      finishTrip,
		getDetails:      getDetails,
		getDriverOffer:  getDriverOffer,
		cancelRide:      cancelRide,
		payRide:         payRide,
		listRiderTrips:  listRiderTrips,
		listDriverTrips: listDriverTrips,
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

// CancelRide implements BookingServiceServer.CancelRide.
func (h *Handler) CancelRide(ctx context.Context, req *bookingpb.CancelRideRequest) (*bookingpb.BookingActionResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if err := h.cancelRide.Execute(ctx, req.GetTripId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.BookingActionResponse{TripId: req.GetTripId(), Status: "cancelled"}, nil
}

// PayRide implements BookingServiceServer.PayRide.
func (h *Handler) PayRide(ctx context.Context, req *bookingpb.StartTripRequest) (*bookingpb.FinishedTripResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	result, err := h.payRide.Execute(ctx, app.PayRideInput{TripID: req.GetTripId()})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.FinishedTripResponse{
		TripId:    result.TripID,
		Status:    result.Status,
		FinalFare: result.FinalFare,
		Currency:  result.Currency,
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

// ArriveAtPickup implements BookingServiceServer.ArriveAtPickup.
func (h *Handler) ArriveAtPickup(ctx context.Context, req *bookingpb.StartTripRequest) (*bookingpb.BookingActionResponse, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if err := h.arriveAtPickup.Execute(ctx, req.GetTripId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.BookingActionResponse{TripId: req.GetTripId(), Status: "driver_arrived"}, nil
}

// ListRiderTrips implements BookingServiceServer.ListRiderTrips.
func (h *Handler) ListRiderTrips(ctx context.Context, req *bookingpb.ListTripsRequest) (*bookingpb.TripListResponse, error) {
	if req.GetPartyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "party_id is required")
	}
	trips, err := h.listRiderTrips.Execute(ctx, req.GetPartyId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.TripListResponse{Trips: toTripSummaryProtos(trips)}, nil
}

// ListDriverTrips implements BookingServiceServer.ListDriverTrips.
func (h *Handler) ListDriverTrips(ctx context.Context, req *bookingpb.ListTripsRequest) (*bookingpb.TripListResponse, error) {
	if req.GetPartyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "party_id is required")
	}
	trips, err := h.listDriverTrips.Execute(ctx, req.GetPartyId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &bookingpb.TripListResponse{Trips: toTripSummaryProtos(trips)}, nil
}

func toTripSummaryProtos(trips []app.TripSummary) []*bookingpb.TripSummaryProto {
	out := make([]*bookingpb.TripSummaryProto, len(trips))
	for i, t := range trips {
		out[i] = &bookingpb.TripSummaryProto{
			TripId:        t.TripID,
			Status:        t.Status,
			PickupAddress: t.PickupAddress,
			DropoffAddress: t.DropoffAddress,
			FinalFare:     t.FinalFare,
			Currency:      t.Currency,
			CreatedAt:     timestamppb.New(t.CreatedAt),
		}
	}
	return out
}

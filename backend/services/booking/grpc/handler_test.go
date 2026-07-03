package grpc_test

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fairride/booking/app"
	bookinggrpc "github.com/fairride/booking/grpc"
	"github.com/fairride/booking/grpc/bookingpb"
)

// ─── stubs ───────────────────────────────────────────────────────────────────

type stubTrip struct {
	trips     map[string]*app.TripInfo
	nextID    string
	startErr  error
	createErr error
}

func newStubTrip() *stubTrip {
	return &stubTrip{trips: make(map[string]*app.TripInfo), nextID: "trip-001"}
}

func (s *stubTrip) CreateTrip(_ context.Context, riderID, pickup, dropoff string) (string, error) {
	if s.createErr != nil {
		return "", s.createErr
	}
	id := s.nextID
	s.trips[id] = &app.TripInfo{TripID: id, RiderID: riderID, Status: "pending",
		PickupAddress: pickup, DropoffAddress: dropoff}
	return id, nil
}

func (s *stubTrip) StartTrip(_ context.Context, tripID string) error {
	if s.startErr != nil {
		return s.startErr
	}
	t, ok := s.trips[tripID]
	if !ok {
		return errors.New("not found")
	}
	t.Status = "in_progress"
	return nil
}

func (s *stubTrip) CompleteTrip(_ context.Context, tripID string, fare int64, currency string) (*app.TripInfo, error) {
	t, ok := s.trips[tripID]
	if !ok {
		return nil, errors.New("not found")
	}
	t.Status = "completed"
	t.FinalFareTotal = fare
	t.FareCurrency = currency
	return t, nil
}

func (s *stubTrip) GetTrip(_ context.Context, tripID string) (*app.TripInfo, error) {
	t, ok := s.trips[tripID]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

type stubDispatch struct {
	jobs       map[string]*app.DispatchInfo
	acceptErr  error
	rejectErr  error
	requestErr error
}

func newStubDispatch() *stubDispatch {
	return &stubDispatch{jobs: make(map[string]*app.DispatchInfo)}
}

func (s *stubDispatch) RequestDispatch(_ context.Context, tripID, _ string, _, _ float64) error {
	if s.requestErr != nil {
		return s.requestErr
	}
	s.jobs[tripID] = &app.DispatchInfo{TripID: tripID, Status: "searching"}
	return nil
}

func (s *stubDispatch) AcceptTrip(_ context.Context, tripID, driverID string) error {
	if s.acceptErr != nil {
		return s.acceptErr
	}
	j, ok := s.jobs[tripID]
	if !ok {
		return errors.New("not found")
	}
	j.Status = "assigned"
	j.AssignedDriverID = driverID
	return nil
}

func (s *stubDispatch) RejectTrip(_ context.Context, tripID, _ string) error {
	if s.rejectErr != nil {
		return s.rejectErr
	}
	if _, ok := s.jobs[tripID]; !ok {
		return errors.New("not found")
	}
	return nil
}

func (s *stubDispatch) GetDispatchStatus(_ context.Context, tripID string) (*app.DispatchInfo, error) {
	j, ok := s.jobs[tripID]
	if !ok {
		return nil, errors.New("not found")
	}
	return j, nil
}

type stubPricing struct {
	fare     int64
	currency string
	err      error
}

func (s *stubPricing) CalculateFinalFare(_ context.Context, _ string, _, _ float64) (*app.FareInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &app.FareInfo{Total: s.fare, CurrencyCode: s.currency}, nil
}

// ─── helper ──────────────────────────────────────────────────────────────────

func newHandler(trip *stubTrip, dispatch *stubDispatch, pricing *stubPricing) *bookinggrpc.Handler {
	return bookinggrpc.NewHandler(
		app.NewBookRideUseCase(trip, dispatch),
		app.NewAcceptDispatchOfferUseCase(dispatch),
		app.NewRejectDispatchOfferUseCase(dispatch),
		app.NewStartTripUseCase(trip),
		app.NewFinishTripUseCase(pricing, trip),
		app.NewGetBookingDetailsUseCase(trip, dispatch),
	)
}

func defaultPricing() *stubPricing { return &stubPricing{fare: 325, currency: "USD"} }

// ─── BookRide ─────────────────────────────────────────────────────────────────

func TestBookRide_OK(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	resp, err := h.BookRide(context.Background(), &bookingpb.BookRideRequest{
		RiderId:        "r1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
		PickupLat:      10.77,
		PickupLon:      106.69,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetTripId() == "" {
		t.Error("expected non-empty trip_id")
	}
	if resp.GetStatus() != "searching" {
		t.Errorf("status = %q, want searching", resp.GetStatus())
	}
}

func TestBookRide_MissingRiderID(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.BookRide(context.Background(), &bookingpb.BookRideRequest{
		PickupAddress:  "pickup",
		DropoffAddress: "dropoff",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestBookRide_MissingPickup(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.BookRide(context.Background(), &bookingpb.BookRideRequest{
		RiderId:        "r1",
		DropoffAddress: "dropoff",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

// ─── AcceptDispatchOffer ──────────────────────────────────────────────────────

func TestAcceptDispatchOffer_OK(t *testing.T) {
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "searching"}

	h := newHandler(newStubTrip(), dispatch, defaultPricing())
	resp, err := h.AcceptDispatchOffer(context.Background(), &bookingpb.AcceptDispatchOfferRequest{
		TripId: "t1", DriverId: "d1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetStatus() != "driver_assigned" {
		t.Errorf("status = %q, want driver_assigned", resp.GetStatus())
	}
}

func TestAcceptDispatchOffer_MissingTripID(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.AcceptDispatchOffer(context.Background(), &bookingpb.AcceptDispatchOfferRequest{DriverId: "d1"})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

// ─── RejectDispatchOffer ──────────────────────────────────────────────────────

func TestRejectDispatchOffer_OK(t *testing.T) {
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "searching"}

	h := newHandler(newStubTrip(), dispatch, defaultPricing())
	resp, err := h.RejectDispatchOffer(context.Background(), &bookingpb.RejectDispatchOfferRequest{
		TripId: "t1", DriverId: "d1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetStatus() != "searching" {
		t.Errorf("status = %q, want searching", resp.GetStatus())
	}
}

func TestRejectDispatchOffer_MissingDriverID(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.RejectDispatchOffer(context.Background(), &bookingpb.RejectDispatchOfferRequest{TripId: "t1"})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

// ─── StartTrip ───────────────────────────────────────────────────────────────

func TestStartTrip_OK(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "driver_assigned"}

	h := newHandler(trip, newStubDispatch(), defaultPricing())
	resp, err := h.StartTrip(context.Background(), &bookingpb.StartTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetStatus() != "in_progress" {
		t.Errorf("status = %q, want in_progress", resp.GetStatus())
	}
}

func TestStartTrip_MissingTripID(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.StartTrip(context.Background(), &bookingpb.StartTripRequest{})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

// ─── FinishTrip ───────────────────────────────────────────────────────────────

func TestFinishTrip_OK(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}

	h := newHandler(trip, newStubDispatch(), &stubPricing{fare: 325, currency: "USD"})
	resp, err := h.FinishTrip(context.Background(), &bookingpb.FinishTripRequest{
		TripId:      "t1",
		VehicleType: "car",
		DistanceKm:  5.0,
		DurationMin: 15.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetStatus() != "completed" {
		t.Errorf("status = %q, want completed", resp.GetStatus())
	}
	if resp.GetFinalFare() != 325 {
		t.Errorf("final_fare = %d, want 325", resp.GetFinalFare())
	}
	if resp.GetCurrency() != "USD" {
		t.Errorf("currency = %q, want USD", resp.GetCurrency())
	}
}

func TestFinishTrip_MissingTripID(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.FinishTrip(context.Background(), &bookingpb.FinishTripRequest{VehicleType: "car"})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestFinishTrip_MissingVehicleType(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.FinishTrip(context.Background(), &bookingpb.FinishTripRequest{TripId: "t1"})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

// ─── GetBookingDetails ────────────────────────────────────────────────────────

func TestGetBookingDetails_OK(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{
		TripID: "t1", Status: "completed", RiderID: "r1", DriverID: "d1",
		PickupAddress: "pickup", DropoffAddress: "dropoff",
		FinalFareTotal: 325, FareCurrency: "USD",
	}
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "assigned"}

	h := newHandler(trip, dispatch, defaultPricing())
	resp, err := h.GetBookingDetails(context.Background(), &bookingpb.GetBookingDetailsRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetTripStatus() != "completed" {
		t.Errorf("trip_status = %q, want completed", resp.GetTripStatus())
	}
	if resp.GetDispatchStatus() != "assigned" {
		t.Errorf("dispatch_status = %q, want assigned", resp.GetDispatchStatus())
	}
	if resp.GetFinalFare() != 325 {
		t.Errorf("final_fare = %d, want 325", resp.GetFinalFare())
	}
}

func TestGetBookingDetails_MissingTripID(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.GetBookingDetails(context.Background(), &bookingpb.GetBookingDetailsRequest{})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

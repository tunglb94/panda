package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func (s *stubTrip) CreateTrip(_ context.Context, in app.CreateTripParams) (*app.CreateTripResult, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	id := s.nextID
	s.trips[id] = &app.TripInfo{TripID: id, RiderID: in.RiderID, Status: "pending",
		PickupAddress: in.PickupAddress, DropoffAddress: in.DropoffAddress}
	result := &app.CreateTripResult{TripID: id}
	if in.TripType == "delivery" {
		result.DeliveryID = "delivery-001"
	}
	return result, nil
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

func (s *stubTrip) CancelTrip(_ context.Context, tripID, _ string) error {
	if t, ok := s.trips[tripID]; ok {
		t.Status = "cancelled"
	}
	return nil
}

func (s *stubTrip) InitiatePayment(_ context.Context, tripID string) error {
	if t, ok := s.trips[tripID]; ok {
		t.Status = "payment_pending"
	}
	return nil
}

func (s *stubTrip) PayTrip(_ context.Context, tripID, _ string) (*app.TripInfo, error) {
	t, ok := s.trips[tripID]
	if !ok {
		return nil, errors.New("not found")
	}
	t.Status = "settled"
	return t, nil
}

func (s *stubTrip) MarkDriverArrived(_ context.Context, tripID string) error {
	if t, ok := s.trips[tripID]; ok {
		t.Status = "driver_arrived"
	}
	return nil
}

func (s *stubTrip) ListByRider(_ context.Context, _ string) ([]app.TripSummary, error) {
	return nil, nil
}

func (s *stubTrip) ListByDriver(_ context.Context, _ string) ([]app.TripSummary, error) {
	return nil, nil
}

func (s *stubTrip) AcceptDelivery(_ context.Context, _ string) error {
	return nil
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

func (s *stubDispatch) RequestDispatch(_ context.Context, tripID, _, _, _ string, _, _ float64) error {
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

func (s *stubDispatch) GetDriverOffer(_ context.Context, _ string) (*app.DriverOfferInfo, error) {
	return nil, nil
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
		app.NewAcceptDispatchOfferUseCase(dispatch, trip),
		app.NewRejectDispatchOfferUseCase(dispatch),
		app.NewArriveAtPickupUseCase(trip),
		app.NewStartTripUseCase(trip),
		app.NewFinishTripUseCase(pricing, trip),
		app.NewGetBookingDetailsUseCase(trip, dispatch),
		app.NewGetDriverCurrentOfferUseCase(dispatch, trip),
		app.NewCancelRideUseCase(trip),
		app.NewPayRideUseCase(trip),
		app.NewListRiderTripsUseCase(trip),
		app.NewListDriverTripsUseCase(trip),
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

// TestBookRide_Delivery_OK is an end-to-end (proto request -> handler ->
// use case -> proto response) check that Delivery V1 Phase 2's new
// BookRideRequest fields flow through the gRPC handler correctly and
// DeliveryId comes back populated. Delivery V1 Phase 2
// (docs/business/DELIVERY_V1_DESIGN.md).
func TestBookRide_Delivery_OK(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	resp, err := h.BookRide(context.Background(), &bookingpb.BookRideRequest{
		RiderId:            "r1",
		PickupAddress:      "123 Main St",
		DropoffAddress:     "456 Elm Ave",
		PickupLat:          10.77,
		PickupLon:          106.69,
		TripType:           "delivery",
		PickupContactName:  "Nguyen Van A",
		PickupContactPhone: "0912345678",
		ReceiverName:       "Tran Thi B",
		ReceiverPhone:      "0987654321",
		PackageNote:        "handle with care",
		PackageValue:       500000,
		PackageWeight:      1.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetStatus() != "searching" {
		t.Errorf("status = %q, want searching", resp.GetStatus())
	}
	if resp.GetDeliveryId() != "delivery-001" {
		t.Errorf("delivery_id = %q, want delivery-001", resp.GetDeliveryId())
	}
}

func TestBookRide_RideOnly_DeliveryIdStaysEmpty(t *testing.T) {
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
	if resp.GetDeliveryId() != "" {
		t.Errorf("delivery_id = %q, want empty for a Ride booking", resp.GetDeliveryId())
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
	if resp.GetStatus() != "payment_pending" {
		t.Errorf("status = %q, want payment_pending", resp.GetStatus())
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

// ─── GetDriverCurrentOffer ────────────────────────────────────────────────────

// stubDispatchWithOffer overrides GetDriverOffer to return a fixed offer.
type stubDispatchWithOffer struct {
	*stubDispatch
	offer *app.DriverOfferInfo
}

func (s *stubDispatchWithOffer) GetDriverOffer(_ context.Context, _ string) (*app.DriverOfferInfo, error) {
	return s.offer, nil
}

func newHandlerWithOfferDispatch(trip *stubTrip, dispatch app.DispatchClient, pricing *stubPricing) *bookinggrpc.Handler {
	return bookinggrpc.NewHandler(
		app.NewBookRideUseCase(trip, dispatch),
		app.NewAcceptDispatchOfferUseCase(dispatch, trip),
		app.NewRejectDispatchOfferUseCase(dispatch),
		app.NewArriveAtPickupUseCase(trip),
		app.NewStartTripUseCase(trip),
		app.NewFinishTripUseCase(pricing, trip),
		app.NewGetBookingDetailsUseCase(trip, dispatch),
		app.NewGetDriverCurrentOfferUseCase(dispatch, trip),
		app.NewCancelRideUseCase(trip),
		app.NewPayRideUseCase(trip),
		app.NewListRiderTripsUseCase(trip),
		app.NewListDriverTripsUseCase(trip),
	)
}

func TestGetDriverCurrentOffer_HasOffer(t *testing.T) {
	trip := newStubTrip()
	trip.trips["trip1"] = &app.TripInfo{
		TripID:         "trip1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
		Status:         "searching",
	}
	dispatch := &stubDispatchWithOffer{
		stubDispatch: newStubDispatch(),
		offer: &app.DriverOfferInfo{
			TripID:         "trip1",
			OfferExpiresAt: time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC),
		},
	}
	h := newHandlerWithOfferDispatch(trip, dispatch, defaultPricing())

	resp, err := h.GetDriverCurrentOffer(context.Background(), &bookingpb.GetDriverCurrentOfferRequest{
		DriverId: "d1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetHasOffer() {
		t.Error("has_offer = false, want true")
	}
	if resp.GetTripId() != "trip1" {
		t.Errorf("trip_id = %q, want trip1", resp.GetTripId())
	}
	if resp.GetPickupAddress() != "123 Main St" {
		t.Errorf("pickup_address = %q, want 123 Main St", resp.GetPickupAddress())
	}
	if resp.GetDropoffAddress() != "456 Elm Ave" {
		t.Errorf("dropoff_address = %q, want 456 Elm Ave", resp.GetDropoffAddress())
	}
	if resp.GetOfferExpiresAt() == nil {
		t.Error("offer_expires_at is nil, want non-nil")
	}
}

func TestGetDriverCurrentOffer_NoOffer(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	resp, err := h.GetDriverCurrentOffer(context.Background(), &bookingpb.GetDriverCurrentOfferRequest{
		DriverId: "d99",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetHasOffer() {
		t.Error("has_offer = true, want false when no offer exists")
	}
}

func TestGetDriverCurrentOffer_MissingDriverID(t *testing.T) {
	h := newHandler(newStubTrip(), newStubDispatch(), defaultPricing())
	_, err := h.GetDriverCurrentOffer(context.Background(), &bookingpb.GetDriverCurrentOfferRequest{})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

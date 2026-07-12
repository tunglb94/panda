package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fairride/booking/app"
)

// ─── stub clients ────────────────────────────────────────────────────────────

type stubTrip struct {
	trips               map[string]*app.TripInfo
	nextID              string
	nextDeliveryID      string
	createErr           error
	startErr            error
	cancelErr           error
	cancelled           []string // IDs cancelled via CancelTrip
	acceptDeliveryErr   error
	acceptDeliveryCalls []string // IDs passed to AcceptDelivery
	// lastCreateParams records the params passed to the most recent
	// CreateTrip call, for asserting Booking correctly forwards Delivery
	// fields (Delivery V1 Phase 2, docs/business/DELIVERY_V1_DESIGN.md).
	lastCreateParams app.CreateTripParams
}

func newStubTrip() *stubTrip {
	return &stubTrip{trips: make(map[string]*app.TripInfo), nextID: "trip-001", nextDeliveryID: "delivery-001"}
}

func (s *stubTrip) CreateTrip(_ context.Context, in app.CreateTripParams) (*app.CreateTripResult, error) {
	s.lastCreateParams = in
	if s.createErr != nil {
		return nil, s.createErr
	}
	id := s.nextID
	s.trips[id] = &app.TripInfo{
		TripID:         id,
		RiderID:        in.RiderID,
		Status:         "pending",
		PickupAddress:  in.PickupAddress,
		DropoffAddress: in.DropoffAddress,
	}
	result := &app.CreateTripResult{TripID: id}
	if in.TripType == "delivery" {
		result.DeliveryID = s.nextDeliveryID
	}
	return result, nil
}

func (s *stubTrip) StartTrip(_ context.Context, tripID string) error {
	if s.startErr != nil {
		return s.startErr
	}
	t, ok := s.trips[tripID]
	if !ok {
		return errors.New("trip not found")
	}
	t.Status = "in_progress"
	return nil
}

func (s *stubTrip) CompleteTrip(_ context.Context, tripID string, fare int64, currency string) (*app.TripInfo, error) {
	t, ok := s.trips[tripID]
	if !ok {
		return nil, errors.New("trip not found")
	}
	t.Status = "completed"
	t.FinalFareTotal = fare
	t.FareCurrency = currency
	return t, nil
}

func (s *stubTrip) GetTrip(_ context.Context, tripID string) (*app.TripInfo, error) {
	t, ok := s.trips[tripID]
	if !ok {
		return nil, errors.New("trip not found")
	}
	return t, nil
}

func (s *stubTrip) CancelTrip(_ context.Context, tripID, _ string) error {
	if s.cancelErr != nil {
		return s.cancelErr
	}
	s.cancelled = append(s.cancelled, tripID)
	if t, ok := s.trips[tripID]; ok {
		t.Status = "cancelled"
	}
	return nil
}

func (s *stubTrip) InitiatePayment(_ context.Context, tripID string) error {
	t, ok := s.trips[tripID]
	if !ok {
		return errors.New("trip not found")
	}
	t.Status = "payment_pending"
	return nil
}

func (s *stubTrip) PayTrip(_ context.Context, tripID, method string) (*app.TripInfo, error) {
	t, ok := s.trips[tripID]
	if !ok {
		return nil, errors.New("trip not found")
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

func (s *stubTrip) AcceptDelivery(_ context.Context, tripID string) error {
	s.acceptDeliveryCalls = append(s.acceptDeliveryCalls, tripID)
	return s.acceptDeliveryErr
}

type stubDispatch struct {
	jobs       map[string]*app.DispatchInfo
	acceptErr  error
	rejectErr  error
	requestErr error
	// lastTripType records the tripType passed to the most recent
	// RequestDispatch call, for asserting Booking correctly forwards it
	// (Delivery V1 Phase 3, docs/business/DELIVERY_V1_DESIGN.md).
	lastTripType string
}

func newStubDispatch() *stubDispatch {
	return &stubDispatch{jobs: make(map[string]*app.DispatchInfo)}
}

func (s *stubDispatch) RequestDispatch(_ context.Context, tripID, _, tripType, _ string, _, _ float64) error {
	s.lastTripType = tripType
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
		return errors.New("dispatch job not found")
	}
	j.Status = "assigned"
	j.AssignedDriverID = driverID
	return nil
}

func (s *stubDispatch) RejectTrip(_ context.Context, tripID, _ string) error {
	if s.rejectErr != nil {
		return s.rejectErr
	}
	j, ok := s.jobs[tripID]
	if !ok {
		return errors.New("dispatch job not found")
	}
	j.Status = "searching"
	return nil
}

func (s *stubDispatch) GetDispatchStatus(_ context.Context, tripID string) (*app.DispatchInfo, error) {
	j, ok := s.jobs[tripID]
	if !ok {
		return nil, errors.New("job not found")
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

func newStubPricing(fare int64, currency string) *stubPricing {
	return &stubPricing{fare: fare, currency: currency}
}

func (s *stubPricing) CalculateFinalFare(_ context.Context, _ string, _, _ float64) (*app.FareInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &app.FareInfo{Total: s.fare, CurrencyCode: s.currency}, nil
}

// ─── BookRide ─────────────────────────────────────────────────────────────────

func TestBookRide_Success(t *testing.T) {
	trip := newStubTrip()
	dispatch := newStubDispatch()
	uc := app.NewBookRideUseCase(trip, dispatch)

	result, err := uc.Execute(context.Background(), app.BookRideInput{
		RiderID:        "r1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
		PickupLat:      10.0,
		PickupLon:      106.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TripID != "trip-001" {
		t.Errorf("TripID = %q, want trip-001", result.TripID)
	}
	if result.Status != "searching" {
		t.Errorf("Status = %q, want searching", result.Status)
	}
	// Verify dispatch was triggered
	di, _ := dispatch.GetDispatchStatus(context.Background(), result.TripID)
	if di.Status != "searching" {
		t.Errorf("dispatch status = %q, want searching", di.Status)
	}
}

func TestBookRide_CreateTripError(t *testing.T) {
	trip := newStubTrip()
	trip.createErr = errors.New("trip service unavailable")
	uc := app.NewBookRideUseCase(trip, newStubDispatch())

	_, err := uc.Execute(context.Background(), app.BookRideInput{
		RiderID:        "r1",
		PickupAddress:  "pickup",
		DropoffAddress: "dropoff",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── BookRide — Delivery (Delivery V1 Phase 2, docs/business/DELIVERY_V1_DESIGN.md) ──
//
// BookRideUseCase does not itself validate receiver/pickup-contact/phone —
// that logic lives once, in the Trip service's entity.NewDelivery (see
// backend/services/trip/app/app_test.go's TestCreateTrip_Delivery_* tests
// for that real validation coverage), reached here only through the
// TripClient port. These tests instead verify Booking's actual job: the
// same BookRideInput/Execute pipeline used for Ride correctly forwards
// every Delivery field to TripClient.CreateTrip, returns DeliveryID when
// present, and propagates a TripClient validation failure unchanged rather
// than swallowing or duplicating it.

func validDeliveryBookRideInput() app.BookRideInput {
	return app.BookRideInput{
		RiderID:            "r1",
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
		PackageWeightKg:    1.5,
	}
}

func TestBookRide_RideBookingPass(t *testing.T) {
	// Same pipeline as Delivery, TripType left at its zero value — the
	// literal "Ride booking hoạt động y chang hiện tại" requirement.
	trip := newStubTrip()
	dispatch := newStubDispatch()
	uc := app.NewBookRideUseCase(trip, dispatch)

	result, err := uc.Execute(context.Background(), app.BookRideInput{
		RiderID:        "r1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
		PickupLat:      10.0,
		PickupLon:      106.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DeliveryID != "" {
		t.Errorf("DeliveryID = %q, want empty for a Ride booking", result.DeliveryID)
	}
	if trip.lastCreateParams.TripType != "" {
		t.Errorf("forwarded TripType = %q, want empty", trip.lastCreateParams.TripType)
	}
}

func TestBookRide_DeliveryBookingPass(t *testing.T) {
	trip := newStubTrip()
	dispatch := newStubDispatch()
	uc := app.NewBookRideUseCase(trip, dispatch)

	result, err := uc.Execute(context.Background(), validDeliveryBookRideInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "searching" {
		t.Errorf("Status = %q, want searching", result.Status)
	}
	if result.DeliveryID != "delivery-001" {
		t.Errorf("DeliveryID = %q, want delivery-001", result.DeliveryID)
	}
	// Every delivery field must have been forwarded to TripClient.CreateTrip unchanged.
	got := trip.lastCreateParams
	if got.TripType != "delivery" {
		t.Errorf("forwarded TripType = %q, want delivery", got.TripType)
	}
	if got.PickupContactName != "Nguyen Van A" || got.PickupContactPhone != "0912345678" {
		t.Errorf("pickup contact not forwarded correctly: %+v", got)
	}
	if got.ReceiverName != "Tran Thi B" || got.ReceiverPhone != "0987654321" {
		t.Errorf("receiver not forwarded correctly: %+v", got)
	}
	if got.PackageNote != "handle with care" || got.PackageValue != 500000 || got.PackageWeightKg != 1.5 {
		t.Errorf("package fields not forwarded correctly: %+v", got)
	}
	// Delivery V1 Phase 3: TripType must also reach Dispatch, via the same
	// RequestDispatch call Ride uses (docs/business/DELIVERY_V1_DESIGN.md Phần 9).
	if dispatch.lastTripType != "delivery" {
		t.Errorf("forwarded TripType to Dispatch = %q, want delivery", dispatch.lastTripType)
	}
	// Delivery must reuse the exact same dispatch pipeline as Ride.
	di, _ := dispatch.GetDispatchStatus(context.Background(), result.TripID)
	if di.Status != "searching" {
		t.Errorf("dispatch status = %q, want searching", di.Status)
	}
}

func TestBookRide_Delivery_MissingReceiverFails(t *testing.T) {
	trip := newStubTrip()
	trip.createErr = errors.New("[INVALID_ARGUMENT] receiver name must not be empty")
	uc := app.NewBookRideUseCase(trip, newStubDispatch())

	in := validDeliveryBookRideInput()
	in.ReceiverName = ""
	_, err := uc.Execute(context.Background(), in)
	if err == nil {
		t.Fatal("expected error for missing receiver to propagate from TripClient")
	}
}

func TestBookRide_Delivery_MissingPickupContactFails(t *testing.T) {
	trip := newStubTrip()
	trip.createErr = errors.New("[INVALID_ARGUMENT] sender name must not be empty")
	uc := app.NewBookRideUseCase(trip, newStubDispatch())

	in := validDeliveryBookRideInput()
	in.PickupContactName = ""
	_, err := uc.Execute(context.Background(), in)
	if err == nil {
		t.Fatal("expected error for missing pickup contact to propagate from TripClient")
	}
}

func TestBookRide_Delivery_InvalidPhoneFails(t *testing.T) {
	trip := newStubTrip()
	trip.createErr = errors.New("[INVALID_ARGUMENT] receiver phone is not a valid phone number")
	uc := app.NewBookRideUseCase(trip, newStubDispatch())

	in := validDeliveryBookRideInput()
	in.ReceiverPhone = "not-a-phone"
	_, err := uc.Execute(context.Background(), in)
	if err == nil {
		t.Fatal("expected error for invalid phone to propagate from TripClient")
	}
}

func TestBookRide_DispatchError(t *testing.T) {
	trip := newStubTrip()
	dispatch := newStubDispatch()
	dispatch.requestErr = errors.New("dispatch unavailable")
	uc := app.NewBookRideUseCase(trip, dispatch)

	_, err := uc.Execute(context.Background(), app.BookRideInput{
		RiderID:        "r1",
		PickupAddress:  "pickup",
		DropoffAddress: "dropoff",
	})
	if err == nil {
		t.Fatal("expected error when dispatch fails")
	}
}

// ─── AcceptDispatchOffer ──────────────────────────────────────────────────────

func TestAcceptDispatchOffer_Success(t *testing.T) {
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "searching"}
	trip := newStubTrip()

	uc := app.NewAcceptDispatchOfferUseCase(dispatch, trip)
	if err := uc.Execute(context.Background(), "t1", "d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	di, _ := dispatch.GetDispatchStatus(context.Background(), "t1")
	if di.AssignedDriverID != "d1" {
		t.Errorf("AssignedDriverID = %q, want d1", di.AssignedDriverID)
	}
	// P0-1: a successful accept must also tell Trip to accept the linked
	// Delivery, if any.
	if len(trip.acceptDeliveryCalls) != 1 || trip.acceptDeliveryCalls[0] != "t1" {
		t.Errorf("AcceptDelivery calls = %v, want exactly one call for t1", trip.acceptDeliveryCalls)
	}
}

func TestAcceptDispatchOffer_DispatchError(t *testing.T) {
	dispatch := newStubDispatch()
	dispatch.acceptErr = errors.New("offer expired")
	trip := newStubTrip()

	uc := app.NewAcceptDispatchOfferUseCase(dispatch, trip)
	if err := uc.Execute(context.Background(), "t1", "d1"); err == nil {
		t.Fatal("expected error, got nil")
	}
	// Dispatch failed first — AcceptDelivery must never have been reached.
	if len(trip.acceptDeliveryCalls) != 0 {
		t.Errorf("AcceptDelivery must not be called when dispatch.AcceptTrip fails, got calls: %v", trip.acceptDeliveryCalls)
	}
}

func TestAcceptDispatchOffer_AcceptDeliveryErrorPropagates(t *testing.T) {
	// P0-1: if Trip rejects the Delivery-acceptance step (e.g. a genuine
	// state conflict), the driver's accept must surface that error — never
	// silently succeed and leave the delivery stuck.
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "searching"}
	trip := newStubTrip()
	trip.acceptDeliveryErr = errors.New("delivery service unavailable")

	uc := app.NewAcceptDispatchOfferUseCase(dispatch, trip)
	if err := uc.Execute(context.Background(), "t1", "d1"); err == nil {
		t.Fatal("expected AcceptDelivery's error to propagate, got nil")
	}
}

func TestAcceptDispatchOffer_NilTripClientIsSafe(t *testing.T) {
	// Nil-safe: callers that don't care about Delivery keep working
	// unchanged (also covers the production wiring's other call sites
	// that predate this fix, if any still exist).
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "searching"}

	uc := app.NewAcceptDispatchOfferUseCase(dispatch, nil)
	if err := uc.Execute(context.Background(), "t1", "d1"); err != nil {
		t.Fatalf("unexpected error with nil trip client: %v", err)
	}
}

// ─── RejectDispatchOffer ──────────────────────────────────────────────────────

func TestRejectDispatchOffer_Success(t *testing.T) {
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "searching"}

	uc := app.NewRejectDispatchOfferUseCase(dispatch)
	if err := uc.Execute(context.Background(), "t1", "d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRejectDispatchOffer_Error(t *testing.T) {
	dispatch := newStubDispatch()
	dispatch.rejectErr = errors.New("already assigned")

	uc := app.NewRejectDispatchOfferUseCase(dispatch)
	if err := uc.Execute(context.Background(), "t1", "d1"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── StartTrip ───────────────────────────────────────────────────────────────

func TestStartTrip_Success(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "driver_assigned"}

	uc := app.NewStartTripUseCase(trip)
	if err := uc.Execute(context.Background(), "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ti, _ := trip.GetTrip(context.Background(), "t1")
	if ti.Status != "in_progress" {
		t.Errorf("status = %q, want in_progress", ti.Status)
	}
}

func TestStartTrip_Error(t *testing.T) {
	trip := newStubTrip()
	trip.startErr = errors.New("wrong status")

	uc := app.NewStartTripUseCase(trip)
	if err := uc.Execute(context.Background(), "t1"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ─── FinishTrip ───────────────────────────────────────────────────────────────

func TestFinishTrip_Success(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}

	pricing := newStubPricing(325, "USD")
	uc := app.NewFinishTripUseCase(pricing, trip)

	result, err := uc.Execute(context.Background(), app.FinishTripInput{
		TripID:      "t1",
		VehicleType: "car",
		DistanceKM:  5.0,
		DurationMin: 15.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "payment_pending" {
		t.Errorf("status = %q, want payment_pending", result.Status)
	}
	if result.FinalFare != 325 {
		t.Errorf("FinalFare = %d, want 325", result.FinalFare)
	}
	if result.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", result.Currency)
	}
	if result.VehicleType != "car" {
		t.Errorf("VehicleType = %q, want car", result.VehicleType)
	}
	if result.DistanceKM != 5.0 {
		t.Errorf("DistanceKM = %f, want 5.0", result.DistanceKM)
	}
}

func TestFinishTrip_PricingError(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}

	pricing := &stubPricing{err: errors.New("pricing unavailable")}
	uc := app.NewFinishTripUseCase(pricing, trip)

	_, err := uc.Execute(context.Background(), app.FinishTripInput{
		TripID:      "t1",
		VehicleType: "car",
		DistanceKM:  5.0,
		DurationMin: 15.0,
	})
	if err == nil {
		t.Fatal("expected error when pricing fails")
	}
}

func TestFinishTrip_TripNotFound(t *testing.T) {
	pricing := newStubPricing(325, "USD")
	trip := newStubTrip() // empty trips map

	uc := app.NewFinishTripUseCase(pricing, trip)
	_, err := uc.Execute(context.Background(), app.FinishTripInput{
		TripID:      "missing",
		VehicleType: "car",
		DistanceKM:  5.0,
		DurationMin: 15.0,
	})
	if err == nil {
		t.Fatal("expected error for missing trip")
	}
}

// ─── GetBookingDetails ────────────────────────────────────────────────────────

func TestGetBookingDetails_WithDispatch(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{
		TripID:         "t1",
		RiderID:        "r1",
		DriverID:       "d1",
		Status:         "in_progress",
		PickupAddress:  "pickup",
		DropoffAddress: "dropoff",
	}
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "assigned"}

	uc := app.NewGetBookingDetailsUseCase(trip, dispatch)
	details, err := uc.Execute(context.Background(), "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.TripStatus != "in_progress" {
		t.Errorf("TripStatus = %q, want in_progress", details.TripStatus)
	}
	if details.DispatchStatus != "assigned" {
		t.Errorf("DispatchStatus = %q, want assigned", details.DispatchStatus)
	}
	if details.DriverID != "d1" {
		t.Errorf("DriverID = %q, want d1", details.DriverID)
	}
}

func TestGetBookingDetails_NoDispatch(t *testing.T) {
	// Trip exists but dispatch job has not been created yet
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "pending"}
	dispatch := newStubDispatch() // no jobs

	uc := app.NewGetBookingDetailsUseCase(trip, dispatch)
	details, err := uc.Execute(context.Background(), "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.DispatchStatus != "unknown" {
		t.Errorf("DispatchStatus = %q, want unknown", details.DispatchStatus)
	}
}

func TestGetBookingDetails_TripNotFound(t *testing.T) {
	trip := newStubTrip()
	uc := app.NewGetBookingDetailsUseCase(trip, newStubDispatch())

	_, err := uc.Execute(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing trip")
	}
}

// ─── Full booking flow ─────────────────────────────────────────────────────────

// TestFullBookingFlow exercises all steps of the booking flow end-to-end using stubs.
func TestFullBookingFlow(t *testing.T) {
	tripClient := newStubTrip()
	dispatchClient := newStubDispatch()
	pricingClient := newStubPricing(325, "USD")

	ctx := context.Background()

	// Step 1: Rider books a ride
	bookUC := app.NewBookRideUseCase(tripClient, dispatchClient)
	bookResult, err := bookUC.Execute(ctx, app.BookRideInput{
		RiderID:        "rider-1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
		PickupLat:      10.77,
		PickupLon:      106.69,
	})
	if err != nil {
		t.Fatalf("step 1 BookRide: %v", err)
	}
	tripID := bookResult.TripID
	if bookResult.Status != "searching" {
		t.Errorf("step 1: status = %q, want searching", bookResult.Status)
	}

	// Step 2: Simulate dispatch offering trip to driver (dispatch already set to searching)
	dispatchClient.jobs[tripID].Status = "searching"

	// Step 3: Driver accepts offer
	acceptUC := app.NewAcceptDispatchOfferUseCase(dispatchClient, tripClient)
	if err := acceptUC.Execute(ctx, tripID, "driver-1"); err != nil {
		t.Fatalf("step 3 AcceptDispatchOffer: %v", err)
	}
	// Update trip to simulate dispatch service setting driver_assigned
	tripClient.trips[tripID].Status = "driver_assigned"
	tripClient.trips[tripID].DriverID = "driver-1"

	// Step 4: Driver starts the trip
	startUC := app.NewStartTripUseCase(tripClient)
	if err := startUC.Execute(ctx, tripID); err != nil {
		t.Fatalf("step 4 StartTrip: %v", err)
	}
	ti, _ := tripClient.GetTrip(ctx, tripID)
	if ti.Status != "in_progress" {
		t.Errorf("step 4: status = %q, want in_progress", ti.Status)
	}

	// Step 5: Driver finishes the trip (fare calculated, trip → payment_pending)
	finishUC := app.NewFinishTripUseCase(pricingClient, tripClient)
	result, err := finishUC.Execute(ctx, app.FinishTripInput{
		TripID:      tripID,
		VehicleType: "car",
		DistanceKM:  5.0,
		DurationMin: 15.0,
	})
	if err != nil {
		t.Fatalf("step 5 FinishTrip: %v", err)
	}
	if result.Status != "payment_pending" {
		t.Errorf("step 5: status = %q, want payment_pending", result.Status)
	}
	if result.FinalFare != 325 {
		t.Errorf("step 5: FinalFare = %d, want 325", result.FinalFare)
	}

	// Step 6: Rider pays (payment_pending → settled)
	payUC := app.NewPayRideUseCase(tripClient)
	payResult, err := payUC.Execute(ctx, app.PayRideInput{
		TripID:        tripID,
		PaymentMethod: "cash",
	})
	if err != nil {
		t.Fatalf("step 6 PayRide: %v", err)
	}
	if payResult.Status != "settled" {
		t.Errorf("step 6: status = %q, want settled", payResult.Status)
	}

	// Step 7: Verify final booking details
	getUC := app.NewGetBookingDetailsUseCase(tripClient, dispatchClient)
	details, err := getUC.Execute(ctx, tripID)
	if err != nil {
		t.Fatalf("step 7 GetBookingDetails: %v", err)
	}
	if details.TripStatus != "settled" {
		t.Errorf("step 7: TripStatus = %q, want settled", details.TripStatus)
	}
	if details.FinalFare != 325 {
		t.Errorf("step 7: FinalFare = %d, want 325", details.FinalFare)
	}
}

// ─── Saga Compensation ────────────────────────────────────────────────────────

// TestBookRide_DispatchError_CompensatesTrip verifies that when RequestDispatch fails
// after a trip has been created, CancelTrip is called to prevent orphaned trips.
func TestBookRide_DispatchError_CompensatesTrip(t *testing.T) {
	trip := newStubTrip()
	dispatch := newStubDispatch()
	dispatch.requestErr = errors.New("dispatch unavailable")
	uc := app.NewBookRideUseCase(trip, dispatch)

	_, err := uc.Execute(context.Background(), app.BookRideInput{
		RiderID:        "r1",
		PickupAddress:  "pickup",
		DropoffAddress: "dropoff",
	})
	if err == nil {
		t.Fatal("expected error when dispatch fails")
	}
	// CancelTrip must have been called for the created trip
	if len(trip.cancelled) != 1 {
		t.Errorf("CancelTrip calls = %d, want 1 (saga compensation)", len(trip.cancelled))
	}
	if len(trip.cancelled) == 1 && trip.cancelled[0] != trip.nextID {
		t.Errorf("cancelled trip = %q, want %q", trip.cancelled[0], trip.nextID)
	}
}

// ─── Idempotency ──────────────────────────────────────────────────────────────

func TestBookRide_DuplicateIdempotentRequest(t *testing.T) {
	trip := newStubTrip()
	dispatch := newStubDispatch()
	store := app.NewMemoryIdempotencyStore()
	uc := app.NewBookRideUseCase(trip, dispatch).WithIdempotency(store)

	in := app.BookRideInput{
		RiderID:        "r1",
		PickupAddress:  "pickup",
		DropoffAddress: "dropoff",
		IdempotencyKey: "key-abc",
	}

	// First call succeeds.
	if _, err := uc.Execute(context.Background(), in); err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	// Second call with same key must return AlreadyExists.
	_, err := uc.Execute(context.Background(), in)
	if err == nil {
		t.Fatal("second call: expected AlreadyExists error, got nil")
	}
}

func TestAcceptDispatchOffer_DuplicateIdempotentRequest(t *testing.T) {
	dispatch := newStubDispatch()
	dispatch.jobs["t1"] = &app.DispatchInfo{TripID: "t1", Status: "searching"}
	store := app.NewMemoryIdempotencyStore()
	uc := app.NewAcceptDispatchOfferUseCase(dispatch, nil).WithIdempotency(store)

	// First accept succeeds.
	if err := uc.Execute(context.Background(), "t1", "d1"); err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	// Second accept with same tripID must return AlreadyExists.
	if err := uc.Execute(context.Background(), "t1", "d1"); err == nil {
		t.Fatal("second call: expected AlreadyExists error, got nil")
	}
}

func TestFinishTrip_DuplicateIdempotentRequest(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "in_progress"}
	pricing := newStubPricing(100, "USD")
	store := app.NewMemoryIdempotencyStore()
	uc := app.NewFinishTripUseCase(pricing, trip).WithIdempotency(store)

	in := app.FinishTripInput{TripID: "t1", VehicleType: "car", DistanceKM: 5.0, DurationMin: 15.0}

	// First finish succeeds.
	if _, err := uc.Execute(context.Background(), in); err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	// Second finish with same tripID must return AlreadyExists.
	if _, err := uc.Execute(context.Background(), in); err == nil {
		t.Fatal("second call: expected AlreadyExists error, got nil")
	}
}

// ─── PayRide ─────────────────────────────────────────────────────────────────

func TestPayRide_Success(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "payment_pending", FinalFareTotal: 325, FareCurrency: "USD"}

	uc := app.NewPayRideUseCase(trip)
	result, err := uc.Execute(context.Background(), app.PayRideInput{
		TripID:        "t1",
		PaymentMethod: "cash",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "settled" {
		t.Errorf("status = %q, want settled", result.Status)
	}
}

func TestPayRide_DefaultsMethodToCash(t *testing.T) {
	trip := newStubTrip()
	trip.trips["t1"] = &app.TripInfo{TripID: "t1", Status: "payment_pending"}

	uc := app.NewPayRideUseCase(trip)
	result, err := uc.Execute(context.Background(), app.PayRideInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "settled" {
		t.Errorf("status = %q, want settled", result.Status)
	}
}

func TestPayRide_EmptyTripID(t *testing.T) {
	uc := app.NewPayRideUseCase(newStubTrip())
	_, err := uc.Execute(context.Background(), app.PayRideInput{})
	if err == nil {
		t.Fatal("expected error for empty trip_id")
	}
}

func TestPayRide_TripNotFound(t *testing.T) {
	uc := app.NewPayRideUseCase(newStubTrip())
	_, err := uc.Execute(context.Background(), app.PayRideInput{TripID: "missing", PaymentMethod: "cash"})
	if err == nil {
		t.Fatal("expected error for missing trip")
	}
}

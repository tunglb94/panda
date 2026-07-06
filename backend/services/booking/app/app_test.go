package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fairride/booking/app"
)

// ─── stub clients ────────────────────────────────────────────────────────────

type stubTrip struct {
	trips      map[string]*app.TripInfo
	nextID     string
	createErr  error
	startErr   error
	cancelErr  error
	cancelled  []string // IDs cancelled via CancelTrip
}

func newStubTrip() *stubTrip {
	return &stubTrip{trips: make(map[string]*app.TripInfo), nextID: "trip-001"}
}

func (s *stubTrip) CreateTrip(_ context.Context, riderID, pickup, dropoff string) (string, error) {
	if s.createErr != nil {
		return "", s.createErr
	}
	id := s.nextID
	s.trips[id] = &app.TripInfo{
		TripID:         id,
		RiderID:        riderID,
		Status:         "pending",
		PickupAddress:  pickup,
		DropoffAddress: dropoff,
	}
	return id, nil
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

	uc := app.NewAcceptDispatchOfferUseCase(dispatch)
	if err := uc.Execute(context.Background(), "t1", "d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	di, _ := dispatch.GetDispatchStatus(context.Background(), "t1")
	if di.AssignedDriverID != "d1" {
		t.Errorf("AssignedDriverID = %q, want d1", di.AssignedDriverID)
	}
}

func TestAcceptDispatchOffer_DispatchError(t *testing.T) {
	dispatch := newStubDispatch()
	dispatch.acceptErr = errors.New("offer expired")

	uc := app.NewAcceptDispatchOfferUseCase(dispatch)
	if err := uc.Execute(context.Background(), "t1", "d1"); err == nil {
		t.Fatal("expected error, got nil")
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
	if result.Status != "completed" {
		t.Errorf("status = %q, want completed", result.Status)
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
	acceptUC := app.NewAcceptDispatchOfferUseCase(dispatchClient)
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

	// Step 5: Driver finishes the trip
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
	if result.Status != "completed" {
		t.Errorf("step 5: status = %q, want completed", result.Status)
	}
	if result.FinalFare != 325 {
		t.Errorf("step 5: FinalFare = %d, want 325", result.FinalFare)
	}

	// Step 6: Verify final booking details
	getUC := app.NewGetBookingDetailsUseCase(tripClient, dispatchClient)
	details, err := getUC.Execute(ctx, tripID)
	if err != nil {
		t.Fatalf("step 6 GetBookingDetails: %v", err)
	}
	if details.TripStatus != "completed" {
		t.Errorf("step 6: TripStatus = %q, want completed", details.TripStatus)
	}
	if details.FinalFare != 325 {
		t.Errorf("step 6: FinalFare = %d, want 325", details.FinalFare)
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
	uc := app.NewAcceptDispatchOfferUseCase(dispatch).WithIdempotency(store)

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

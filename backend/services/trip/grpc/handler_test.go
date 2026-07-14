package grpc_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fairride/shared/errors"
	"github.com/fairride/trip/app"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
	tripgrpc "github.com/fairride/trip/grpc"
	"github.com/fairride/trip/grpc/trippb"
	"github.com/fairride/trip/infrastructure/memory"
)

// ─── stub repo ───────────────────────────────────────────────────────────────

type stubRepo struct {
	trips map[string]*entity.Trip
}

var _ repository.TripRepository = (*stubRepo)(nil)

func newStub() *stubRepo { return &stubRepo{trips: make(map[string]*entity.Trip)} }

func (r *stubRepo) Save(_ context.Context, t *entity.Trip) error {
	r.trips[t.TripID] = t
	return nil
}
func (r *stubRepo) FindByID(_ context.Context, id string) (*entity.Trip, error) {
	t, ok := r.trips[id]
	if !ok {
		return nil, errors.NotFound("trip not found: " + id)
	}
	return t, nil
}
func (r *stubRepo) FindByRiderID(_ context.Context, riderID string) ([]*entity.Trip, error) {
	var out []*entity.Trip
	for _, t := range r.trips {
		if t.RiderID == riderID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *stubRepo) FindByDriverID(_ context.Context, driverID string) ([]*entity.Trip, error) {
	var out []*entity.Trip
	for _, t := range r.trips {
		if t.DriverID == driverID {
			out = append(out, t)
		}
	}
	return out, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

var testNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func newHandler(repo *stubRepo) *tripgrpc.Handler {
	return newHandlerWithDeliveryRepo(repo, memory.NewDeliveryRepository())
}

// newHandlerWithDeliveryRepo lets Delivery-lifecycle tests (Delivery V1
// Phase 4, docs/business/DELIVERY_V1_DESIGN.md) share the same
// DeliveryRepository instance the handler uses, so they can seed a
// Delivery aggregate before calling PickupParcel/StartDelivery/
// CompleteDelivery. newHandler above keeps its original signature for
// every pre-existing (Ride) test — zero call sites needed updating.
func newHandlerWithDeliveryRepo(repo *stubRepo, deliveryRepo *memory.DeliveryRepository) *tripgrpc.Handler {
	return tripgrpc.NewHandler(
		app.NewCreateTripUseCase(repo, deliveryRepo),
		app.NewCancelTripUseCase(repo),
		app.NewGetTripUseCase(repo),
		app.NewMarkDriverArrivedUseCase(repo),
		app.NewStartTripUseCase(repo),
		app.NewCompleteTripUseCase(repo),
		app.NewInitiatePaymentUseCase(repo),
		app.NewPayTripUseCase(repo),
		app.NewListTripsByRiderUseCase(repo),
		app.NewListTripsByDriverUseCase(repo),
		app.NewPickupParcelUseCase(repo, deliveryRepo),
		app.NewStartDeliveryUseCase(repo, deliveryRepo),
		app.NewCompleteDeliveryUseCase(repo, deliveryRepo),
		app.NewAcceptDeliveryUseCase(repo, deliveryRepo),
	)
}

func seedTrip(repo *stubRepo, tripID, riderID string, st entity.TripStatus) *entity.Trip {
	trip := entity.ReconstituteTrip(tripID, riderID, "", st, "pickup", "dropoff", "", 0, "", "", testNow, testNow, entity.CompleteFinancials{}, nil, nil, 0, entity.TripSummary{})
	_ = repo.Save(context.Background(), trip)
	return trip
}

// ─── CreateTrip ──────────────────────────────────────────────────────────────

func TestCreateTrip_OK(t *testing.T) {
	h := newHandler(newStub())
	resp, err := h.CreateTrip(context.Background(), &trippb.CreateTripRequest{
		RiderId:        "r1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Trip.TripId == "" {
		t.Error("expected a non-empty trip_id")
	}
	if resp.Trip.Status != "pending" {
		t.Errorf("status = %q, want pending", resp.Trip.Status)
	}
}

func TestCreateTrip_MissingRiderID(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.CreateTrip(context.Background(), &trippb.CreateTripRequest{
		PickupAddress:  "pickup",
		DropoffAddress: "dropoff",
	})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

func TestCreateTrip_MissingPickup(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.CreateTrip(context.Background(), &trippb.CreateTripRequest{
		RiderId:        "r1",
		DropoffAddress: "dropoff",
	})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── CancelTrip ──────────────────────────────────────────────────────────────

func TestCancelTrip_OK(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusPending)

	h := newHandler(repo)
	resp, err := h.CancelTrip(context.Background(), &trippb.CancelTripRequest{
		TripId: "t1",
		Reason: "changed mind",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Trip.Status != "cancelled" {
		t.Errorf("status = %q, want cancelled", resp.Trip.Status)
	}
	if resp.Trip.CancellationReason != "changed mind" {
		t.Errorf("reason = %q", resp.Trip.CancellationReason)
	}
}

func TestCancelTrip_NotFound(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.CancelTrip(context.Background(), &trippb.CancelTripRequest{TripId: "x"})
	s, _ := status.FromError(err)
	if s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", s.Code())
	}
}

func TestCancelTrip_InProgress(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusInProgress)

	h := newHandler(repo)
	_, err := h.CancelTrip(context.Background(), &trippb.CancelTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", s.Code())
	}
}

func TestCancelTrip_MissingTripID(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.CancelTrip(context.Background(), &trippb.CancelTripRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── GetTrip ─────────────────────────────────────────────────────────────────

func TestGetTrip_OK(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusSearching)

	h := newHandler(repo)
	resp, err := h.GetTrip(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Trip.TripId != "t1" {
		t.Errorf("trip_id = %q, want t1", resp.Trip.TripId)
	}
}

func TestGetTrip_NotFound(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.GetTrip(context.Background(), &trippb.GetTripRequest{TripId: "missing"})
	s, _ := status.FromError(err)
	if s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", s.Code())
	}
}

func TestGetTrip_MissingTripID(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.GetTrip(context.Background(), &trippb.GetTripRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── StartTrip ───────────────────────────────────────────────────────────────

func TestStartTrip_OK(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusDriverAssigned)

	h := newHandler(repo)
	resp, err := h.StartTrip(context.Background(), &trippb.StartTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Trip.Status != "in_progress" {
		t.Errorf("status = %q, want in_progress", resp.Trip.Status)
	}
}

func TestStartTrip_FromDriverArrived(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusDriverArrived)

	h := newHandler(repo)
	resp, err := h.StartTrip(context.Background(), &trippb.StartTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Trip.Status != "in_progress" {
		t.Errorf("status = %q, want in_progress", resp.Trip.Status)
	}
}

func TestStartTrip_NotFound(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.StartTrip(context.Background(), &trippb.StartTripRequest{TripId: "missing"})
	s, _ := status.FromError(err)
	if s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", s.Code())
	}
}

func TestStartTrip_WrongStatus(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusPending)

	h := newHandler(repo)
	_, err := h.StartTrip(context.Background(), &trippb.StartTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", s.Code())
	}
}

func TestStartTrip_MissingTripID(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.StartTrip(context.Background(), &trippb.StartTripRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── CompleteTrip ─────────────────────────────────────────────────────────────

func TestCompleteTrip_OK(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusInProgress)

	h := newHandler(repo)
	resp, err := h.CompleteTrip(context.Background(), &trippb.CompleteTripRequest{
		TripId:         "t1",
		FinalFareTotal: 325,
		FareCurrency:   "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Trip.Status != "completed" {
		t.Errorf("status = %q, want completed", resp.Trip.Status)
	}
	if resp.Trip.FinalFareTotal != 325 {
		t.Errorf("final_fare_total = %d, want 325", resp.Trip.FinalFareTotal)
	}
	if resp.Trip.FareCurrency != "USD" {
		t.Errorf("fare_currency = %q, want USD", resp.Trip.FareCurrency)
	}
}

func TestCompleteTrip_NotFound(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.CompleteTrip(context.Background(), &trippb.CompleteTripRequest{
		TripId:         "missing",
		FinalFareTotal: 100,
		FareCurrency:   "USD",
	})
	s, _ := status.FromError(err)
	if s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", s.Code())
	}
}

func TestCompleteTrip_WrongStatus(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusPending)

	h := newHandler(repo)
	_, err := h.CompleteTrip(context.Background(), &trippb.CompleteTripRequest{
		TripId:         "t1",
		FinalFareTotal: 100,
		FareCurrency:   "USD",
	})
	s, _ := status.FromError(err)
	if s.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", s.Code())
	}
}

func TestCompleteTrip_MissingTripID(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.CompleteTrip(context.Background(), &trippb.CompleteTripRequest{FareCurrency: "USD"})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

func TestCompleteTrip_MissingCurrency(t *testing.T) {
	h := newHandler(newStub())
	_, err := h.CompleteTrip(context.Background(), &trippb.CompleteTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── Delivery lifecycle gRPC: PickupParcel / StartDelivery / CompleteDelivery ──
// (Delivery V1 Phase 4, docs/business/DELIVERY_V1_DESIGN.md)

// seedDeliveryTrip seeds a Trip (TripType=delivery, Status=driver_arrived)
// and its linked Delivery aggregate at the given status, into repo/deliveryRepo.
func seedDeliveryTrip(repo *stubRepo, deliveryRepo *memory.DeliveryRepository, tripID, deliveryID string, deliveryStatus entity.DeliveryStatus) {
	trip := entity.ReconstituteTrip(tripID, "r1", "d1", entity.StatusDriverArrived, "pickup", "dropoff", "", 0, "", "", testNow, testNow, entity.CompleteFinancials{}, nil, nil, 0, entity.TripSummary{})
	trip.TripType = entity.TripTypeDelivery
	trip.DeliveryID = deliveryID
	_ = repo.Save(context.Background(), trip)

	delivery := entity.ReconstituteDelivery(
		deliveryID, "Nguyen Van A", "0912345678", "Tran Thi B", "0987654321",
		"", "", entity.PackageTypeSmall, 1.5, false, false, 500000,
		deliveryStatus, testNow, testNow,
	)
	_ = deliveryRepo.Save(context.Background(), delivery)
}

func TestPickupParcel_OK(t *testing.T) {
	repo := newStub()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(repo, deliveryRepo, "t1", "d1", entity.DeliveryStatusAccepted)

	h := newHandlerWithDeliveryRepo(repo, deliveryRepo)
	resp, err := h.PickupParcel(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetTrip().GetDeliveryStatus() != "PARCEL_PICKED_UP" {
		t.Errorf("delivery_status = %q, want PARCEL_PICKED_UP", resp.GetTrip().GetDeliveryStatus())
	}
	if resp.GetTrip().GetStatus() != "in_progress" {
		t.Errorf("trip status = %q, want in_progress (reused Trip.Start())", resp.GetTrip().GetStatus())
	}
}

func TestPickupParcel_WrongStatus(t *testing.T) {
	repo := newStub()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(repo, deliveryRepo, "t1", "d1", entity.DeliveryStatusCreated) // Accepted skipped

	h := newHandlerWithDeliveryRepo(repo, deliveryRepo)
	_, err := h.PickupParcel(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", s.Code())
	}
}

func TestPickupParcel_RideTripRejected(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusDriverArrived) // plain Ride trip
	h := newHandlerWithDeliveryRepo(repo, memory.NewDeliveryRepository())

	_, err := h.PickupParcel(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument (Ride trip)", s.Code())
	}
}

func TestPickupParcel_MissingTripID(t *testing.T) {
	h := newHandlerWithDeliveryRepo(newStub(), memory.NewDeliveryRepository())
	_, err := h.PickupParcel(context.Background(), &trippb.GetTripRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

func TestStartDelivery_OK(t *testing.T) {
	repo := newStub()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(repo, deliveryRepo, "t1", "d1", entity.DeliveryStatusParcelPickedUp)

	h := newHandlerWithDeliveryRepo(repo, deliveryRepo)
	resp, err := h.StartDelivery(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetTrip().GetDeliveryStatus() != "IN_DELIVERY" {
		t.Errorf("delivery_status = %q, want IN_DELIVERY", resp.GetTrip().GetDeliveryStatus())
	}
	// Trip.Status must be unaffected (seeded as driver_arrived).
	if resp.GetTrip().GetStatus() != "driver_arrived" {
		t.Errorf("trip status = %q, want unchanged (driver_arrived)", resp.GetTrip().GetStatus())
	}
}

func TestStartDelivery_WrongStatus(t *testing.T) {
	repo := newStub()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(repo, deliveryRepo, "t1", "d1", entity.DeliveryStatusAccepted) // ParcelPickedUp skipped

	h := newHandlerWithDeliveryRepo(repo, deliveryRepo)
	_, err := h.StartDelivery(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", s.Code())
	}
}

func TestStartDelivery_RideTripRejected(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusInProgress)
	h := newHandlerWithDeliveryRepo(repo, memory.NewDeliveryRepository())

	_, err := h.StartDelivery(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument (Ride trip)", s.Code())
	}
}

func TestCompleteDelivery_OK(t *testing.T) {
	repo := newStub()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(repo, deliveryRepo, "t1", "d1", entity.DeliveryStatusInDelivery)

	h := newHandlerWithDeliveryRepo(repo, deliveryRepo)
	resp, err := h.CompleteDelivery(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetTrip().GetDeliveryStatus() != "COMPLETED" {
		t.Errorf("delivery_status = %q, want COMPLETED", resp.GetTrip().GetDeliveryStatus())
	}
}

func TestCompleteDelivery_WrongStatus(t *testing.T) {
	repo := newStub()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(repo, deliveryRepo, "t1", "d1", entity.DeliveryStatusParcelPickedUp) // InDelivery skipped

	h := newHandlerWithDeliveryRepo(repo, deliveryRepo)
	_, err := h.CompleteDelivery(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", s.Code())
	}
}

func TestCompleteDelivery_RideTripRejected(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusInProgress)
	h := newHandlerWithDeliveryRepo(repo, memory.NewDeliveryRepository())

	_, err := h.CompleteDelivery(context.Background(), &trippb.GetTripRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument (Ride trip)", s.Code())
	}
}

// TestDeliveryLifecycle_FullHappyPath_gRPC exercises PickupParcel ->
// StartDelivery -> CompleteDelivery through the gRPC handler, exactly as a
// driver app would call them in sequence.
func TestDeliveryLifecycle_FullHappyPath_gRPC(t *testing.T) {
	repo := newStub()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(repo, deliveryRepo, "t1", "d1", entity.DeliveryStatusAccepted)
	h := newHandlerWithDeliveryRepo(repo, deliveryRepo)
	ctx := context.Background()

	if _, err := h.PickupParcel(ctx, &trippb.GetTripRequest{TripId: "t1"}); err != nil {
		t.Fatalf("PickupParcel: %v", err)
	}
	if _, err := h.StartDelivery(ctx, &trippb.GetTripRequest{TripId: "t1"}); err != nil {
		t.Fatalf("StartDelivery: %v", err)
	}
	resp, err := h.CompleteDelivery(ctx, &trippb.GetTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("CompleteDelivery: %v", err)
	}
	if resp.GetTrip().GetDeliveryStatus() != "COMPLETED" {
		t.Errorf("final delivery_status = %q, want COMPLETED", resp.GetTrip().GetDeliveryStatus())
	}
}

// ─── Ride regression: existing RPCs unaffected by Delivery V1 Phase 4 ──────

// TestRideRPCs_DeliveryStatusStaysEmpty confirms every existing (Ride) RPC
// response still has an empty delivery_status — the new field defaults to
// its proto3 zero value and no existing handler path was changed to
// populate it.
func TestRideRPCs_DeliveryStatusStaysEmpty(t *testing.T) {
	repo := newStub()
	seedTrip(repo, "t1", "r1", entity.StatusDriverAssigned)
	h := newHandler(repo)

	resp, err := h.StartTrip(context.Background(), &trippb.StartTripRequest{TripId: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetTrip().GetDeliveryStatus() != "" {
		t.Errorf("delivery_status = %q, want empty for a Ride trip", resp.GetTrip().GetDeliveryStatus())
	}
	if resp.GetTrip().GetTripType() != "" {
		t.Errorf("trip_type = %q, want empty for a Ride trip created via seedTrip", resp.GetTrip().GetTripType())
	}
}

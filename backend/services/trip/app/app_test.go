package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/shared/errors"
	"github.com/fairride/trip/app"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
	"github.com/fairride/trip/infrastructure/memory"
)

// ─── stub repository ─────────────────────────────────────────────────────────

type stubRepo struct {
	trips map[string]*entity.Trip
}

var _ repository.TripRepository = (*stubRepo)(nil)

func newStubRepo() *stubRepo {
	return &stubRepo{trips: make(map[string]*entity.Trip)}
}

func (r *stubRepo) Save(_ context.Context, trip *entity.Trip) error {
	r.trips[trip.TripID] = trip
	return nil
}

func (r *stubRepo) FindByID(_ context.Context, tripID string) (*entity.Trip, error) {
	t, ok := r.trips[tripID]
	if !ok {
		return nil, errors.NotFound("trip not found: " + tripID)
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

// ─── CreateTrip ──────────────────────────────────────────────────────────────

func TestCreateTrip_Valid(t *testing.T) {
	repo := newStubRepo()
	uc := app.NewCreateTripUseCase(repo, memory.NewDeliveryRepository())

	trip, err := uc.Execute(context.Background(), app.CreateTripInput{
		RiderID:        "r1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.TripID == "" {
		t.Error("expected a generated trip ID")
	}
	if trip.RiderID != "r1" {
		t.Errorf("RiderID = %q, want r1", trip.RiderID)
	}
	if trip.Status != entity.StatusPending {
		t.Errorf("Status = %q, want pending", trip.Status)
	}
	// Verify it was persisted in the stub
	if _, err := repo.FindByID(context.Background(), trip.TripID); err != nil {
		t.Error("trip was not saved to repo")
	}
}

func TestCreateTrip_EmptyRiderID(t *testing.T) {
	uc := app.NewCreateTripUseCase(newStubRepo(), memory.NewDeliveryRepository())
	_, err := uc.Execute(context.Background(), app.CreateTripInput{
		RiderID:        "",
		PickupAddress:  "pickup",
		DropoffAddress: "dropoff",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestCreateTrip_EmptyPickup(t *testing.T) {
	uc := app.NewCreateTripUseCase(newStubRepo(), memory.NewDeliveryRepository())
	_, err := uc.Execute(context.Background(), app.CreateTripInput{
		RiderID:        "r1",
		PickupAddress:  "",
		DropoffAddress: "dropoff",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

// ─── CreateTrip — Delivery (Delivery V1 Phase 2, docs/business/DELIVERY_V1_DESIGN.md) ──

func validDeliveryCreateTripInput() app.CreateTripInput {
	return app.CreateTripInput{
		RiderID:            "r1",
		PickupAddress:      "123 Main St",
		DropoffAddress:     "456 Elm Ave",
		TripType:           entity.TripTypeDelivery,
		PickupContactName:  "Nguyen Van A",
		PickupContactPhone: "0912345678",
		ReceiverName:       "Tran Thi B",
		ReceiverPhone:      "0987654321",
		PackageNote:        "handle with care",
		PackageValue:       500000,
		PackageWeightKg:    1.5,
	}
}

func TestCreateTrip_RideBookingPass(t *testing.T) {
	// TripType left at its zero value ("") — must behave exactly like a
	// plain Ride booking (backward compatibility check).
	repo := newStubRepo()
	uc := app.NewCreateTripUseCase(repo, memory.NewDeliveryRepository())

	trip, err := uc.Execute(context.Background(), app.CreateTripInput{
		RiderID:        "r1",
		PickupAddress:  "123 Main St",
		DropoffAddress: "456 Elm Ave",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.TripType != entity.TripTypeRide {
		t.Errorf("TripType = %q, want ride", trip.TripType)
	}
	if trip.DeliveryID != "" {
		t.Errorf("DeliveryID = %q, want empty for a Ride booking", trip.DeliveryID)
	}
}

func TestCreateTrip_DeliveryBookingPass(t *testing.T) {
	repo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	uc := app.NewCreateTripUseCase(repo, deliveryRepo)

	trip, err := uc.Execute(context.Background(), validDeliveryCreateTripInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.TripType != entity.TripTypeDelivery {
		t.Errorf("TripType = %q, want delivery", trip.TripType)
	}
	if trip.DeliveryID == "" {
		t.Fatal("expected a generated DeliveryID")
	}
	// The Delivery aggregate must actually have been persisted.
	delivery, err := deliveryRepo.FindByID(context.Background(), trip.DeliveryID)
	if err != nil {
		t.Fatalf("delivery was not saved: %v", err)
	}
	if delivery.SenderName != "Nguyen Van A" {
		t.Errorf("SenderName (from pickup_contact_name) = %q", delivery.SenderName)
	}
	if delivery.ReceiverName != "Tran Thi B" {
		t.Errorf("ReceiverName = %q", delivery.ReceiverName)
	}
	if delivery.DeclaredValue != 500000 {
		t.Errorf("DeclaredValue = %d, want 500000", delivery.DeclaredValue)
	}
	if delivery.CashOnDelivery {
		t.Error("CashOnDelivery must be false — COD is out of scope")
	}
}

func TestCreateTrip_Delivery_MissingReceiverFails(t *testing.T) {
	uc := app.NewCreateTripUseCase(newStubRepo(), memory.NewDeliveryRepository())
	in := validDeliveryCreateTripInput()
	in.ReceiverName = ""
	_, err := uc.Execute(context.Background(), in)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument for missing receiver, got %v", err)
	}
}

func TestCreateTrip_Delivery_MissingPickupContactFails(t *testing.T) {
	uc := app.NewCreateTripUseCase(newStubRepo(), memory.NewDeliveryRepository())
	in := validDeliveryCreateTripInput()
	in.PickupContactName = ""
	_, err := uc.Execute(context.Background(), in)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument for missing pickup contact, got %v", err)
	}
}

func TestCreateTrip_Delivery_InvalidPhoneFails(t *testing.T) {
	for _, tc := range []struct {
		name string
		mut  func(in *app.CreateTripInput)
	}{
		{"pickup contact phone", func(in *app.CreateTripInput) { in.PickupContactPhone = "not-a-phone" }},
		{"receiver phone", func(in *app.CreateTripInput) { in.ReceiverPhone = "123" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := app.NewCreateTripUseCase(newStubRepo(), memory.NewDeliveryRepository())
			in := validDeliveryCreateTripInput()
			tc.mut(&in)
			_, err := uc.Execute(context.Background(), in)
			if !errors.IsCode(err, errors.CodeInvalidArgument) {
				t.Errorf("expected InvalidArgument for invalid %s, got %v", tc.name, err)
			}
		})
	}
}

func TestCreateTrip_Delivery_NegativeWeightFails(t *testing.T) {
	uc := app.NewCreateTripUseCase(newStubRepo(), memory.NewDeliveryRepository())
	in := validDeliveryCreateTripInput()
	in.PackageWeightKg = -1
	_, err := uc.Execute(context.Background(), in)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument for negative weight, got %v", err)
	}
}

func TestCreateTrip_Delivery_NegativeValueFails(t *testing.T) {
	uc := app.NewCreateTripUseCase(newStubRepo(), memory.NewDeliveryRepository())
	in := validDeliveryCreateTripInput()
	in.PackageValue = -500
	_, err := uc.Execute(context.Background(), in)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument for negative package value, got %v", err)
	}
}

func TestCreateTrip_Delivery_DoesNotPersistTripOnDeliveryValidationFailure(t *testing.T) {
	// If Delivery validation fails, no orphaned Trip should be persisted.
	repo := newStubRepo()
	uc := app.NewCreateTripUseCase(repo, memory.NewDeliveryRepository())
	in := validDeliveryCreateTripInput()
	in.ReceiverPhone = "bad"
	_, err := uc.Execute(context.Background(), in)
	if err == nil {
		t.Fatal("expected an error")
	}
	if len(repo.trips) != 0 {
		t.Errorf("expected no trip to be persisted, found %d", len(repo.trips))
	}
}

// ─── Delivery lifecycle: PickupParcel / StartDelivery / CompleteDelivery ───
// (Delivery V1 Phase 4, docs/business/DELIVERY_V1_DESIGN.md)

// seedDeliveryTrip creates a Trip (TripType=Delivery, Status=driver_arrived
// — i.e. "ArrivedPickup" already happened via the existing, unchanged
// MarkDriverArrived flow) and its linked Delivery aggregate at the given
// DeliveryStatus, and saves both.
func seedDeliveryTrip(t *testing.T, tripRepo *stubRepo, deliveryRepo *memory.DeliveryRepository, tripID, deliveryID string, deliveryStatus entity.DeliveryStatus) {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	trip := entity.ReconstituteTrip(tripID, "rider1", "driver1", entity.StatusDriverArrived, "pickup", "dropoff", "", 0, "", "", now, now, entity.CompleteFinancials{})
	trip.TripType = entity.TripTypeDelivery
	trip.DeliveryID = deliveryID
	if err := tripRepo.Save(context.Background(), trip); err != nil {
		t.Fatalf("seed trip save: %v", err)
	}

	delivery := entity.ReconstituteDelivery(
		deliveryID, "Nguyen Van A", "0912345678", "Tran Thi B", "0987654321",
		"", "", entity.PackageTypeSmall, 1.5, false, false, 500000,
		deliveryStatus, now, now,
	)
	if err := deliveryRepo.Save(context.Background(), delivery); err != nil {
		t.Fatalf("seed delivery save: %v", err)
	}
}

func TestPickupParcel_LifecycleCorrect(t *testing.T) {
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusAccepted)

	uc := app.NewPickupParcelUseCase(tripRepo, deliveryRepo)
	delivery, err := uc.Execute(context.Background(), app.PickupParcelInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delivery.Status != entity.DeliveryStatusParcelPickedUp {
		t.Errorf("Delivery.Status = %q, want PARCEL_PICKED_UP", delivery.Status)
	}
	// Reuses Trip's existing Start() — Trip must now be in_progress.
	trip, err := tripRepo.FindByID(context.Background(), "t1")
	if err != nil {
		t.Fatalf("FindByID trip: %v", err)
	}
	if trip.Status != entity.StatusInProgress {
		t.Errorf("Trip.Status = %q, want in_progress (reused Trip.Start())", trip.Status)
	}
	// Repository: the saved Delivery must be retrievable with the new status.
	saved, err := deliveryRepo.FindByID(context.Background(), "d1")
	if err != nil {
		t.Fatalf("FindByID delivery: %v", err)
	}
	if saved.Status != entity.DeliveryStatusParcelPickedUp {
		t.Errorf("persisted Delivery.Status = %q, want PARCEL_PICKED_UP", saved.Status)
	}
}

func TestPickupParcel_LifecycleWrong_FromCreatedFails(t *testing.T) {
	// Accepted was skipped — must be rejected (Delivery-side precondition).
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusCreated)

	uc := app.NewPickupParcelUseCase(tripRepo, deliveryRepo)
	_, err := uc.Execute(context.Background(), app.PickupParcelInput{TripID: "t1"})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestPickupParcel_RejectsRideTrip(t *testing.T) {
	// Ride regression guard: calling a Delivery-lifecycle action on a Ride
	// trip must be rejected, not silently do something to a Ride trip.
	tripRepo := newStubRepo()
	makeTrip(tripRepo, "t1", "rider1", entity.StatusDriverArrived) // plain Ride trip (TripType zero value)
	deliveryRepo := memory.NewDeliveryRepository()

	uc := app.NewPickupParcelUseCase(tripRepo, deliveryRepo)
	_, err := uc.Execute(context.Background(), app.PickupParcelInput{TripID: "t1"})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument for a Ride trip, got %v", err)
	}
}

func TestPickupParcel_TripNotFound(t *testing.T) {
	uc := app.NewPickupParcelUseCase(newStubRepo(), memory.NewDeliveryRepository())
	_, err := uc.Execute(context.Background(), app.PickupParcelInput{TripID: "missing"})
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestPickupParcel_EmptyTripID(t *testing.T) {
	uc := app.NewPickupParcelUseCase(newStubRepo(), memory.NewDeliveryRepository())
	_, err := uc.Execute(context.Background(), app.PickupParcelInput{})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestStartDelivery_LifecycleCorrect(t *testing.T) {
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusParcelPickedUp)

	uc := app.NewStartDeliveryUseCase(tripRepo, deliveryRepo)
	delivery, err := uc.Execute(context.Background(), app.StartDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delivery.Status != entity.DeliveryStatusInDelivery {
		t.Errorf("Delivery.Status = %q, want IN_DELIVERY", delivery.Status)
	}
	// Trip.Status must NOT change — purely a Delivery-internal sub-status.
	trip, _ := tripRepo.FindByID(context.Background(), "t1")
	if trip.Status != entity.StatusDriverArrived {
		t.Errorf("Trip.Status = %q, want unchanged (driver_arrived)", trip.Status)
	}
}

func TestStartDelivery_LifecycleWrong_FromAcceptedFails(t *testing.T) {
	// ParcelPickedUp was skipped — must be rejected.
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusAccepted)

	uc := app.NewStartDeliveryUseCase(tripRepo, deliveryRepo)
	_, err := uc.Execute(context.Background(), app.StartDeliveryInput{TripID: "t1"})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestStartDelivery_RejectsRideTrip(t *testing.T) {
	tripRepo := newStubRepo()
	makeTrip(tripRepo, "t1", "rider1", entity.StatusInProgress)
	deliveryRepo := memory.NewDeliveryRepository()

	uc := app.NewStartDeliveryUseCase(tripRepo, deliveryRepo)
	_, err := uc.Execute(context.Background(), app.StartDeliveryInput{TripID: "t1"})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument for a Ride trip, got %v", err)
	}
}

func TestCompleteDelivery_LifecycleCorrect(t *testing.T) {
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusInDelivery)

	uc := app.NewCompleteDeliveryUseCase(tripRepo, deliveryRepo)
	delivery, err := uc.Execute(context.Background(), app.CompleteDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delivery.Status != entity.DeliveryStatusCompleted {
		t.Errorf("Delivery.Status = %q, want COMPLETED (chained MarkDelivered -> CompleteDelivery)", delivery.Status)
	}
	saved, err := deliveryRepo.FindByID(context.Background(), "d1")
	if err != nil {
		t.Fatalf("FindByID delivery: %v", err)
	}
	if saved.Status != entity.DeliveryStatusCompleted {
		t.Errorf("persisted Delivery.Status = %q, want COMPLETED", saved.Status)
	}
	// Trip.Complete() is deliberately NOT called (no Pricing integration
	// this phase) — Trip.Status stays exactly as seeded.
	trip, _ := tripRepo.FindByID(context.Background(), "t1")
	if trip.Status != entity.StatusDriverArrived {
		t.Errorf("Trip.Status = %q, want unchanged — Trip.Complete() must not be called without a real fare", trip.Status)
	}
}

func TestCompleteDelivery_LifecycleWrong_FromParcelPickedUpFails(t *testing.T) {
	// InDelivery was skipped ("Pickup -> Completed") — must be rejected.
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusParcelPickedUp)

	uc := app.NewCompleteDeliveryUseCase(tripRepo, deliveryRepo)
	_, err := uc.Execute(context.Background(), app.CompleteDeliveryInput{TripID: "t1"})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
	// Neither chained transition may have been persisted.
	saved, _ := deliveryRepo.FindByID(context.Background(), "d1")
	if saved.Status != entity.DeliveryStatusParcelPickedUp {
		t.Errorf("Delivery.Status = %q, want unchanged (PARCEL_PICKED_UP) after a rejected transition", saved.Status)
	}
}

func TestCompleteDelivery_RejectsRideTrip(t *testing.T) {
	tripRepo := newStubRepo()
	makeTrip(tripRepo, "t1", "rider1", entity.StatusInProgress)
	deliveryRepo := memory.NewDeliveryRepository()

	uc := app.NewCompleteDeliveryUseCase(tripRepo, deliveryRepo)
	_, err := uc.Execute(context.Background(), app.CompleteDeliveryInput{TripID: "t1"})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument for a Ride trip, got %v", err)
	}
}

// TestDeliveryLifecycle_FullHappyPath exercises the whole chain end to end
// through the three use cases, exactly as a driver app would call them in
// sequence, proving no step is skippable and the final state is COMPLETED.
func TestDeliveryLifecycle_FullHappyPath(t *testing.T) {
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusAccepted)
	ctx := context.Background()

	pickup := app.NewPickupParcelUseCase(tripRepo, deliveryRepo)
	if _, err := pickup.Execute(ctx, app.PickupParcelInput{TripID: "t1"}); err != nil {
		t.Fatalf("PickupParcel: %v", err)
	}

	start := app.NewStartDeliveryUseCase(tripRepo, deliveryRepo)
	if _, err := start.Execute(ctx, app.StartDeliveryInput{TripID: "t1"}); err != nil {
		t.Fatalf("StartDelivery: %v", err)
	}

	complete := app.NewCompleteDeliveryUseCase(tripRepo, deliveryRepo)
	delivery, err := complete.Execute(ctx, app.CompleteDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("CompleteDelivery: %v", err)
	}
	if delivery.Status != entity.DeliveryStatusCompleted {
		t.Errorf("final Status = %q, want COMPLETED", delivery.Status)
	}
}

// ─── AcceptDelivery (production hardening P0-1) ────────────────────────────
//
// Closes the real bug: AcceptDispatchOfferUseCase (Booking) never called
// Delivery.AcceptByDriver, so PickupParcel always failed with
// PreconditionFailed (see TestPickupParcel_LifecycleWrong_FromCreatedFails
// above, which intentionally still asserts that failure for a delivery
// that skipped Accepted — proving this fix doesn't weaken PickupParcel's
// own precondition, it just makes the real Accept step actually happen).

func TestAcceptDelivery_TransitionsCreatedToAccepted(t *testing.T) {
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusCreated)

	uc := app.NewAcceptDeliveryUseCase(tripRepo, deliveryRepo)
	delivery, err := uc.Execute(context.Background(), app.AcceptDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delivery == nil || delivery.Status != entity.DeliveryStatusAccepted {
		t.Fatalf("Delivery.Status = %+v, want ACCEPTED", delivery)
	}
	saved, err := deliveryRepo.FindByID(context.Background(), "d1")
	if err != nil {
		t.Fatalf("FindByID delivery: %v", err)
	}
	if saved.Status != entity.DeliveryStatusAccepted {
		t.Errorf("persisted Delivery.Status = %q, want ACCEPTED", saved.Status)
	}
}

func TestAcceptDelivery_IdempotentOnDuplicateCall(t *testing.T) {
	// A retried/duplicate accept HTTP request must not error and must not
	// re-run AcceptByDriver's precondition check (which would fail on a
	// non-CREATED status if this weren't guarded).
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusCreated)

	uc := app.NewAcceptDeliveryUseCase(tripRepo, deliveryRepo)
	ctx := context.Background()
	first, err := uc.Execute(ctx, app.AcceptDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	second, err := uc.Execute(ctx, app.AcceptDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("second (duplicate) call: unexpected error: %v", err)
	}
	if first.Status != entity.DeliveryStatusAccepted || second.Status != entity.DeliveryStatusAccepted {
		t.Errorf("both calls must report ACCEPTED, got first=%q second=%q", first.Status, second.Status)
	}
}

func TestAcceptDelivery_NoOpPastCreated(t *testing.T) {
	// Idempotent for any later status too, not just a literal duplicate —
	// a late/out-of-order retry after the driver already picked up the
	// parcel must not error or regress the status.
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusParcelPickedUp)

	uc := app.NewAcceptDeliveryUseCase(tripRepo, deliveryRepo)
	delivery, err := uc.Execute(context.Background(), app.AcceptDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delivery.Status != entity.DeliveryStatusParcelPickedUp {
		t.Errorf("Status = %q, want unchanged (PARCEL_PICKED_UP), not regressed to ACCEPTED", delivery.Status)
	}
}

func TestAcceptDelivery_NoOpForRideTrip(t *testing.T) {
	// Unlike PickupParcel/StartDelivery/CompleteDelivery, a Ride trip is
	// NOT an error here — Booking calls this unconditionally after every
	// accept (Ride or Delivery), so it must be a silent, harmless no-op.
	tripRepo := newStubRepo()
	makeTrip(tripRepo, "t1", "rider1", entity.StatusDriverAssigned)
	deliveryRepo := memory.NewDeliveryRepository()

	uc := app.NewAcceptDeliveryUseCase(tripRepo, deliveryRepo)
	delivery, err := uc.Execute(context.Background(), app.AcceptDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("expected no-op (nil, nil) for a Ride trip, got error: %v", err)
	}
	if delivery != nil {
		t.Errorf("expected nil Delivery for a Ride trip, got %+v", delivery)
	}
}

func TestAcceptDelivery_TripNotFound(t *testing.T) {
	uc := app.NewAcceptDeliveryUseCase(newStubRepo(), memory.NewDeliveryRepository())
	_, err := uc.Execute(context.Background(), app.AcceptDeliveryInput{TripID: "missing"})
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestAcceptDelivery_EmptyTripID(t *testing.T) {
	uc := app.NewAcceptDeliveryUseCase(newStubRepo(), memory.NewDeliveryRepository())
	_, err := uc.Execute(context.Background(), app.AcceptDeliveryInput{})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

// TestDeliveryLifecycle_FromRealAcceptThroughCompletion is the true
// end-to-end regression test for P0-1: AcceptDelivery -> PickupParcel ->
// StartDelivery -> CompleteDelivery, starting from CREATED (what a real
// driver accept leaves the delivery at today), proving the exact bug
// report ("Accept -> Arrived Pickup -> Pickup Parcel -> FAIL") is fixed.
func TestDeliveryLifecycle_FromRealAcceptThroughCompletion(t *testing.T) {
	tripRepo := newStubRepo()
	deliveryRepo := memory.NewDeliveryRepository()
	seedDeliveryTrip(t, tripRepo, deliveryRepo, "t1", "d1", entity.DeliveryStatusCreated)
	ctx := context.Background()

	accept := app.NewAcceptDeliveryUseCase(tripRepo, deliveryRepo)
	if _, err := accept.Execute(ctx, app.AcceptDeliveryInput{TripID: "t1"}); err != nil {
		t.Fatalf("AcceptDelivery: %v", err)
	}

	pickup := app.NewPickupParcelUseCase(tripRepo, deliveryRepo)
	if _, err := pickup.Execute(ctx, app.PickupParcelInput{TripID: "t1"}); err != nil {
		t.Fatalf("PickupParcel: %v (this is the exact P0-1 bug if it fails)", err)
	}

	start := app.NewStartDeliveryUseCase(tripRepo, deliveryRepo)
	if _, err := start.Execute(ctx, app.StartDeliveryInput{TripID: "t1"}); err != nil {
		t.Fatalf("StartDelivery: %v", err)
	}

	complete := app.NewCompleteDeliveryUseCase(tripRepo, deliveryRepo)
	delivery, err := complete.Execute(ctx, app.CompleteDeliveryInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("CompleteDelivery: %v", err)
	}
	if delivery.Status != entity.DeliveryStatusCompleted {
		t.Errorf("final Status = %q, want COMPLETED", delivery.Status)
	}
}

// ─── CancelTrip ──────────────────────────────────────────────────────────────

func makeTrip(repo *stubRepo, tripID, riderID string, status entity.TripStatus) *entity.Trip {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	trip := entity.ReconstituteTrip(tripID, riderID, "", status, "pickup", "dropoff", "", 0, "", "", now, now, entity.CompleteFinancials{})
	_ = repo.Save(context.Background(), trip)
	return trip
}

func TestCancelTrip_FromPending(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusPending)

	uc := app.NewCancelTripUseCase(repo)
	trip, err := uc.Execute(context.Background(), app.CancelTripInput{
		TripID: "t1",
		Reason: "changed mind",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusCancelled {
		t.Errorf("Status = %q, want cancelled", trip.Status)
	}
	if trip.CancellationReason != "changed mind" {
		t.Errorf("CancellationReason = %q", trip.CancellationReason)
	}
}

func TestCancelTrip_NotFound(t *testing.T) {
	uc := app.NewCancelTripUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), app.CancelTripInput{TripID: "nonexistent"})
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestCancelTrip_InProgressFails(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusInProgress)

	uc := app.NewCancelTripUseCase(repo)
	_, err := uc.Execute(context.Background(), app.CancelTripInput{TripID: "t1"})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── GetTrip ─────────────────────────────────────────────────────────────────

func TestGetTrip_Found(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusSearching)

	uc := app.NewGetTripUseCase(repo)
	trip, err := uc.Execute(context.Background(), "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.TripID != "t1" {
		t.Errorf("TripID = %q, want t1", trip.TripID)
	}
}

func TestGetTrip_NotFound(t *testing.T) {
	uc := app.NewGetTripUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), "missing")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

// ─── StartTrip ───────────────────────────────────────────────────────────────

func TestStartTrip_FromDriverAssigned(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusDriverAssigned)

	uc := app.NewStartTripUseCase(repo)
	trip, err := uc.Execute(context.Background(), app.StartTripInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusInProgress {
		t.Errorf("Status = %q, want in_progress", trip.Status)
	}
}

func TestStartTrip_FromDriverArrived(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusDriverArrived)

	uc := app.NewStartTripUseCase(repo)
	trip, err := uc.Execute(context.Background(), app.StartTripInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusInProgress {
		t.Errorf("Status = %q, want in_progress", trip.Status)
	}
}

func TestStartTrip_FromPendingFails(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusPending)

	uc := app.NewStartTripUseCase(repo)
	_, err := uc.Execute(context.Background(), app.StartTripInput{TripID: "t1"})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestStartTrip_NotFound(t *testing.T) {
	uc := app.NewStartTripUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), app.StartTripInput{TripID: "missing"})
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

// ─── CompleteTrip ─────────────────────────────────────────────────────────────

func TestCompleteTrip_FromInProgress(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusInProgress)

	uc := app.NewCompleteTripUseCase(repo)
	trip, err := uc.Execute(context.Background(), app.CompleteTripInput{
		TripID:         "t1",
		FinalFareTotal: 325,
		FareCurrency:   "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusCompleted {
		t.Errorf("Status = %q, want completed", trip.Status)
	}
	if trip.FinalFareTotal != 325 {
		t.Errorf("FinalFareTotal = %d, want 325", trip.FinalFareTotal)
	}
	if trip.FareCurrency != "USD" {
		t.Errorf("FareCurrency = %q, want USD", trip.FareCurrency)
	}
}

func TestCompleteTrip_FromPendingFails(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusPending)

	uc := app.NewCompleteTripUseCase(repo)
	_, err := uc.Execute(context.Background(), app.CompleteTripInput{
		TripID:         "t1",
		FinalFareTotal: 100,
		FareCurrency:   "USD",
	})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCompleteTrip_NotFound(t *testing.T) {
	uc := app.NewCompleteTripUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), app.CompleteTripInput{
		TripID:         "missing",
		FinalFareTotal: 100,
		FareCurrency:   "USD",
	})
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

// ─── InitiatePayment ──────────────────────────────────────────────────────────

func TestInitiatePayment_FromCompleted(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusCompleted)

	uc := app.NewInitiatePaymentUseCase(repo)
	trip, err := uc.Execute(context.Background(), "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusPaymentPending {
		t.Errorf("Status = %q, want payment_pending", trip.Status)
	}
}

func TestInitiatePayment_FromInProgressFails(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusInProgress)

	uc := app.NewInitiatePaymentUseCase(repo)
	_, err := uc.Execute(context.Background(), "t1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestInitiatePayment_NotFound(t *testing.T) {
	uc := app.NewInitiatePaymentUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), "missing")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

// ─── PayTrip ──────────────────────────────────────────────────────────────────

func TestPayTrip_FromPaymentPending(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusPaymentPending)

	uc := app.NewPayTripUseCase(repo)
	trip, err := uc.Execute(context.Background(), app.PayTripInput{
		TripID:        "t1",
		PaymentMethod: "cash",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusSettled {
		t.Errorf("Status = %q, want settled", trip.Status)
	}
	if trip.PaymentMethod != "cash" {
		t.Errorf("PaymentMethod = %q, want cash", trip.PaymentMethod)
	}
}

func TestPayTrip_DefaultsToCache(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusPaymentPending)

	uc := app.NewPayTripUseCase(repo)
	trip, err := uc.Execute(context.Background(), app.PayTripInput{TripID: "t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.PaymentMethod != "cash" {
		t.Errorf("PaymentMethod = %q, want cash (default)", trip.PaymentMethod)
	}
}

func TestPayTrip_FromCompletedFails(t *testing.T) {
	repo := newStubRepo()
	makeTrip(repo, "t1", "r1", entity.StatusCompleted)

	uc := app.NewPayTripUseCase(repo)
	_, err := uc.Execute(context.Background(), app.PayTripInput{TripID: "t1", PaymentMethod: "cash"})
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestPayTrip_NotFound(t *testing.T) {
	uc := app.NewPayTripUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), app.PayTripInput{TripID: "missing", PaymentMethod: "cash"})
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

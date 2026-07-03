package postgres_test

import (
	"context"
	"testing"

	"github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/infrastructure/postgres"
)

func newTestRepo() *postgres.TripRepository {
	return postgres.NewTripRepository(testPool)
}

func makeTrip(tripID, riderID string, status entity.TripStatus) *entity.Trip {
	return entity.ReconstituteTrip(
		tripID, riderID, "",
		status,
		"123 Pickup St", "456 Dropoff Ave", "",
		0, "",
		testNow, testNow,
	)
}

// ─── Save + FindByID ─────────────────────────────────────────────────────────

func TestSaveAndFindByID(t *testing.T) {
	setupTest(t)
	repo := newTestRepo()
	ctx := context.Background()

	trip := makeTrip("t1", "r1", entity.StatusPending)
	if err := repo.Save(ctx, trip); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, "t1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.TripID != "t1" {
		t.Errorf("TripID = %q, want t1", got.TripID)
	}
	if got.RiderID != "r1" {
		t.Errorf("RiderID = %q, want r1", got.RiderID)
	}
	if got.Status != entity.StatusPending {
		t.Errorf("Status = %q, want pending", got.Status)
	}
	if got.PickupAddress != "123 Pickup St" {
		t.Errorf("PickupAddress = %q", got.PickupAddress)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	setupTest(t)
	repo := newTestRepo()

	_, err := repo.FindByID(context.Background(), "nonexistent")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

// ─── Upsert (Save idempotency) ────────────────────────────────────────────────

func TestSave_UpdatesMutableFields(t *testing.T) {
	setupTest(t)
	repo := newTestRepo()
	ctx := context.Background()

	trip := makeTrip("t1", "r1", entity.StatusPending)
	if err := repo.Save(ctx, trip); err != nil {
		t.Fatalf("Save initial: %v", err)
	}

	// Simulate cancellation
	trip.Status = entity.StatusCancelled
	trip.CancellationReason = "test reason"
	trip.UpdatedAt = testNow.Add(1)
	if err := repo.Save(ctx, trip); err != nil {
		t.Fatalf("Save update: %v", err)
	}

	got, err := repo.FindByID(ctx, "t1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Status != entity.StatusCancelled {
		t.Errorf("Status = %q, want cancelled", got.Status)
	}
	if got.CancellationReason != "test reason" {
		t.Errorf("CancellationReason = %q, want 'test reason'", got.CancellationReason)
	}
	if got.RiderID != "r1" {
		t.Errorf("RiderID changed to %q — must not change on update", got.RiderID)
	}
}

func TestSave_DoesNotChangePickupDropoff(t *testing.T) {
	setupTest(t)
	repo := newTestRepo()
	ctx := context.Background()

	trip := makeTrip("t1", "r1", entity.StatusPending)
	if err := repo.Save(ctx, trip); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Attempt to mutate addresses (should be ignored by the upsert)
	trip.PickupAddress = "changed"
	trip.DropoffAddress = "changed"
	if err := repo.Save(ctx, trip); err != nil {
		t.Fatalf("Save update: %v", err)
	}

	got, err := repo.FindByID(ctx, "t1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.PickupAddress != "123 Pickup St" {
		t.Errorf("PickupAddress changed to %q — must not change on update", got.PickupAddress)
	}
}

// ─── FindByRiderID ───────────────────────────────────────────────────────────

func TestFindByRiderID_ReturnsAll(t *testing.T) {
	setupTest(t)
	repo := newTestRepo()
	ctx := context.Background()

	t1 := makeTrip("t1", "r1", entity.StatusPending)
	t2 := makeTrip("t2", "r1", entity.StatusCompleted)
	t3 := makeTrip("t3", "r2", entity.StatusPending) // different rider
	for _, trip := range []*entity.Trip{t1, t2, t3} {
		if err := repo.Save(ctx, trip); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	trips, err := repo.FindByRiderID(ctx, "r1")
	if err != nil {
		t.Fatalf("FindByRiderID: %v", err)
	}
	if len(trips) != 2 {
		t.Errorf("len = %d, want 2", len(trips))
	}
}

func TestFindByRiderID_EmptySliceWhenNone(t *testing.T) {
	setupTest(t)
	repo := newTestRepo()

	trips, err := repo.FindByRiderID(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("FindByRiderID: %v", err)
	}
	if len(trips) != 0 {
		t.Errorf("expected empty slice, got %d", len(trips))
	}
}

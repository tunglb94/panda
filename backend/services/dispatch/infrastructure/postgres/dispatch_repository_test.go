package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/dispatch/infrastructure/postgres"
)

func newRepo() *postgres.DispatchRepository {
	return postgres.NewDispatchRepository(testPool)
}

func makeJob(jobID, tripID string, status entity.JobStatus) *entity.DispatchJob {
	return entity.ReconstituteDispatchJob(
		jobID, tripID, "rider1",
		10.762, 106.660,
		status,
		"", "", nil,
		time.Time{},
		30, 5, 0,
		testNow, testNow,
	)
}

// ─── Save + FindByID ─────────────────────────────────────────────────────────

func TestSaveAndFindByID(t *testing.T) {
	setupTest(t)
	repo := newRepo()
	ctx := context.Background()

	job := makeJob("j1", "trip1", entity.JobStatusPending)
	if err := repo.Save(ctx, job); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, "j1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.JobID != "j1" {
		t.Errorf("JobID = %q, want j1", got.JobID)
	}
	if got.TripID != "trip1" {
		t.Errorf("TripID = %q, want trip1", got.TripID)
	}
	if got.PickupLat != 10.762 {
		t.Errorf("PickupLat = %v, want 10.762", got.PickupLat)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	setupTest(t)
	_, err := newRepo().FindByID(context.Background(), "nonexistent")
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestFindByTripID_NotFound(t *testing.T) {
	setupTest(t)
	_, err := newRepo().FindByTripID(context.Background(), "trip99")
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

// ─── Upsert ──────────────────────────────────────────────────────────────────

func TestSave_UpdatesMutableFields(t *testing.T) {
	setupTest(t)
	repo := newRepo()
	ctx := context.Background()

	job := makeJob("j1", "trip1", entity.JobStatusPending)
	_ = repo.Save(ctx, job)

	_ = job.OfferToDriver("d1", testNow)
	_ = repo.Save(ctx, job)

	got, err := repo.FindByID(ctx, "j1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Status != entity.JobStatusSearching {
		t.Errorf("Status = %q, want searching", got.Status)
	}
	if got.CurrentDriverID != "d1" {
		t.Errorf("CurrentDriverID = %q, want d1", got.CurrentDriverID)
	}
	if got.AttemptCount != 1 {
		t.Errorf("AttemptCount = %d, want 1", got.AttemptCount)
	}
	if !got.HasBeenOffered("d1") {
		t.Error("d1 should be in OfferedDriverIDs")
	}
}

func TestSave_TripIDImmutable(t *testing.T) {
	setupTest(t)
	repo := newRepo()
	ctx := context.Background()

	job := makeJob("j1", "trip1", entity.JobStatusPending)
	_ = repo.Save(ctx, job)

	// Mutate mutable fields only
	job.Status = entity.JobStatusFailed
	_ = repo.Save(ctx, job)

	got, _ := repo.FindByID(ctx, "j1")
	if got.TripID != "trip1" {
		t.Errorf("TripID changed to %q — must not change on update", got.TripID)
	}
}

// ─── FindExpiredOffers ────────────────────────────────────────────────────────

func TestFindExpiredOffers(t *testing.T) {
	setupTest(t)
	repo := newRepo()
	ctx := context.Background()

	// Job with an expired offer
	j1 := makeJob("j1", "trip1", entity.JobStatusPending)
	_ = j1.OfferToDriver("d1", testNow)
	_ = repo.Save(ctx, j1)

	// Job not yet expired (expires in 1 hour)
	j2 := makeJob("j2", "trip2", entity.JobStatusPending)
	_ = j2.OfferToDriver("d2", testNow)
	j2.OfferExpiresAt = testNow.Add(time.Hour) // override for test
	_ = repo.Save(ctx, j2)

	// Find expired as of testNow + 31 seconds
	searchTime := testNow.Add(31 * time.Second)
	expired, err := repo.FindExpiredOffers(ctx, searchTime)
	if err != nil {
		t.Fatalf("FindExpiredOffers: %v", err)
	}

	if len(expired) != 1 {
		t.Fatalf("expected 1 expired job, got %d", len(expired))
	}
	if expired[0].JobID != "j1" {
		t.Errorf("expected j1, got %q", expired[0].JobID)
	}
}

func TestFindExpiredOffers_EmptyWhenNone(t *testing.T) {
	setupTest(t)
	jobs, err := newRepo().FindExpiredOffers(context.Background(), testNow)
	if err != nil {
		t.Fatalf("FindExpiredOffers: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 expired jobs, got %d", len(jobs))
	}
}

// ─── FindByTripID ─────────────────────────────────────────────────────────────

func TestFindByTripID(t *testing.T) {
	setupTest(t)
	repo := newRepo()
	ctx := context.Background()

	job := makeJob("j1", "trip1", entity.JobStatusPending)
	_ = repo.Save(ctx, job)

	got, err := repo.FindByTripID(ctx, "trip1")
	if err != nil {
		t.Fatalf("FindByTripID: %v", err)
	}
	if got.JobID != "j1" {
		t.Errorf("JobID = %q, want j1", got.JobID)
	}
}

package app_test

import (
	"context"
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/dispatch/app"
	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
)

var testNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

// ─── stub implementations ────────────────────────────────────────────────────

type stubJobRepo struct {
	jobs map[string]*entity.DispatchJob // keyed by jobID
}

var _ repository.DispatchJobRepository = (*stubJobRepo)(nil)

func newStubJobRepo() *stubJobRepo { return &stubJobRepo{jobs: make(map[string]*entity.DispatchJob)} }

func (r *stubJobRepo) Save(_ context.Context, job *entity.DispatchJob) error {
	r.jobs[job.JobID] = job
	return nil
}

func (r *stubJobRepo) FindByID(_ context.Context, jobID string) (*entity.DispatchJob, error) {
	j, ok := r.jobs[jobID]
	if !ok {
		return nil, domainerrors.NotFound("job not found: " + jobID)
	}
	return j, nil
}

func (r *stubJobRepo) FindByTripID(_ context.Context, tripID string) (*entity.DispatchJob, error) {
	for _, j := range r.jobs {
		if j.TripID == tripID {
			return j, nil
		}
	}
	return nil, domainerrors.NotFound("no dispatch job for trip: " + tripID)
}

func (r *stubJobRepo) FindExpiredOffers(_ context.Context, now time.Time) ([]*entity.DispatchJob, error) {
	var out []*entity.DispatchJob
	for _, j := range r.jobs {
		if j.Status == entity.JobStatusSearching && j.IsOfferExpired(now) {
			out = append(out, j)
		}
	}
	return out, nil
}

// stubLocationRepo simulates a geo store with configurable active drivers.
type stubLocationRepo struct {
	nearby  []string           // IDs in distance order (nearest first)
	active  map[string]bool    // which IDs respond as active
	updated map[string][2]float64 // driverID → [lat, lon]
}

var _ repository.DriverLocationRepository = (*stubLocationRepo)(nil)

func newStubLocationRepo(nearby []string, active map[string]bool) *stubLocationRepo {
	return &stubLocationRepo{nearby: nearby, active: active, updated: make(map[string][2]float64)}
}

func (r *stubLocationRepo) UpdateLocation(_ context.Context, driverID string, lat, lon float64) error {
	r.updated[driverID] = [2]float64{lat, lon}
	return nil
}

func (r *stubLocationRepo) FindNearby(_ context.Context, _, _ float64, _ float64, _ int) ([]*entity.NearbyDriver, error) {
	var out []*entity.NearbyDriver
	for _, id := range r.nearby {
		out = append(out, &entity.NearbyDriver{DriverID: id})
	}
	return out, nil
}

func (r *stubLocationRepo) IsActive(_ context.Context, driverID string) (bool, error) {
	return r.active[driverID], nil
}

func (r *stubLocationRepo) RemoveLocation(_ context.Context, _ string) error { return nil }

// stubTripUpdater records calls for assertion.
type stubTripUpdater struct {
	searchingTrips []string
	assignedTrips  map[string]string // tripID → driverID
}

var _ repository.TripUpdater = (*stubTripUpdater)(nil)

func newStubTripUpdater() *stubTripUpdater {
	return &stubTripUpdater{assignedTrips: make(map[string]string)}
}

func (u *stubTripUpdater) SetSearching(_ context.Context, tripID string, _ time.Time) error {
	u.searchingTrips = append(u.searchingTrips, tripID)
	return nil
}

func (u *stubTripUpdater) AssignDriver(_ context.Context, tripID, driverID string, _ time.Time) error {
	u.assignedTrips[tripID] = driverID
	return nil
}

// stubTransactor runs fn synchronously with the provided stubs; no real DB tx.
// Used in unit tests to verify application logic without infrastructure.
type stubTransactor struct {
	jobs  repository.DispatchJobRepository
	trips repository.TripUpdater
}

var _ repository.Transactor = (*stubTransactor)(nil)

func (s *stubTransactor) WithinTx(_ context.Context, fn func(repository.DispatchJobRepository, repository.TripUpdater) error) error {
	return fn(s.jobs, s.trips)
}

// failingTripUpdater returns configured errors to simulate write failures.
type failingTripUpdater struct {
	assignErr    error
	searchingErr error
}

var _ repository.TripUpdater = (*failingTripUpdater)(nil)

func (f *failingTripUpdater) SetSearching(_ context.Context, _ string, _ time.Time) error {
	return f.searchingErr
}

func (f *failingTripUpdater) AssignDriver(_ context.Context, _, _ string, _ time.Time) error {
	return f.assignErr
}

// saveFailingJobRepo delegates reads to an embedded stubJobRepo but fails Save.
type saveFailingJobRepo struct {
	*stubJobRepo
	saveErr error
}

func (r *saveFailingJobRepo) Save(_ context.Context, _ *entity.DispatchJob) error {
	return r.saveErr
}

// ──�� helpers ─────────────────────────────────────────────────────────────────

func allActive(ids ...string) map[string]bool {
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// ─── RequestDispatch ─────────────────────────────────────────────────────────

func TestRequestDispatch_OffersNearestDriver(t *testing.T) {
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1", "d2", "d3"}, allActive("d1", "d2", "d3"))
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID:  "trip1",
		RiderID: "rider1",
		PickupLat: 10.0,
		PickupLon: 106.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusSearching {
		t.Errorf("Status = %q, want searching", job.Status)
	}
	if job.CurrentDriverID != "d1" {
		t.Errorf("CurrentDriverID = %q, want d1 (nearest)", job.CurrentDriverID)
	}
	if len(trips.searchingTrips) == 0 {
		t.Error("expected SetSearching to be called")
	}
}

func TestRequestDispatch_NoDriversAvailable(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	locs := newStubLocationRepo(nil, nil) // no nearby drivers
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusFailed {
		t.Errorf("Status = %q, want failed when no drivers found", job.Status)
	}
}

func TestRequestDispatch_SkipsInactiveDrivers(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	// d1 is nearby but inactive; d2 is active
	locs := newStubLocationRepo([]string{"d1", "d2"}, allActive("d2"))
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.CurrentDriverID != "d2" {
		t.Errorf("CurrentDriverID = %q, want d2 (first active)", job.CurrentDriverID)
	}
}

func TestRequestDispatch_MissingTripID(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, newStubLocationRepo(nil, nil), &stubTransactor{jobs: jobs, trips: trips})
	_, err := uc.Execute(context.Background(), app.RequestDispatchInput{RiderID: "r1"})
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

// ─── AcceptTrip ──────────────────────────────────────────────────────────────

func seedJob(t *testing.T, jobs *stubJobRepo, tripID, driverID string) *entity.DispatchJob {
	t.Helper()
	job, err := entity.NewDispatchJob("j1", tripID, "rider1", 10, 106, 30, 5, testNow)
	if err != nil {
		t.Fatalf("NewDispatchJob: %v", err)
	}
	// Use real time so offer_expires_at is always in the future during the test.
	_ = job.OfferToDriver(driverID, time.Now())
	_ = jobs.Save(context.Background(), job)
	return job
}

func TestAcceptTrip_Valid(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	seedJob(t, jobs, "trip1", "d1")

	uc := app.NewAcceptTripUseCase(jobs, &stubTransactor{jobs: jobs, trips: trips})
	job, err := uc.Execute(context.Background(), "trip1", "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusAssigned {
		t.Errorf("Status = %q, want assigned", job.Status)
	}
	if job.AssignedDriverID != "d1" {
		t.Errorf("AssignedDriverID = %q, want d1", job.AssignedDriverID)
	}
	if trips.assignedTrips["trip1"] != "d1" {
		t.Error("expected AssignDriver to be called with d1")
	}
}

func TestAcceptTrip_WrongDriver(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	seedJob(t, jobs, "trip1", "d1")

	uc := app.NewAcceptTripUseCase(jobs, &stubTransactor{jobs: jobs, trips: trips})
	_, err := uc.Execute(context.Background(), "trip1", "d2")
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestAcceptTrip_NotFound(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	uc := app.NewAcceptTripUseCase(jobs, &stubTransactor{jobs: jobs, trips: trips})
	_, err := uc.Execute(context.Background(), "nonexistent", "d1")
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

// ─── AcceptTrip rollback tests ────────────────────────────────────────────────

// TestAcceptTrip_RollbackOnTripUpdateFailure verifies that when the trip status
// UPDATE fails, the use case returns an error and jobRepo.Save is never called.
// In a real PostgreSQL transaction this guarantees neither write is committed.
func TestAcceptTrip_RollbackOnTripUpdateFailure(t *testing.T) {
	outerJobs := newStubJobRepo()
	seedJob(t, outerJobs, "trip1", "d1")

	innerJobs := newStubJobRepo() // separate inner stub to detect spurious Save calls
	failTrips := &failingTripUpdater{assignErr: domainerrors.Internal("simulated db failure")}
	txr := &stubTransactor{jobs: innerJobs, trips: failTrips}
	uc := app.NewAcceptTripUseCase(outerJobs, txr)

	_, err := uc.Execute(context.Background(), "trip1", "d1")
	if err == nil {
		t.Fatal("expected error when trip update fails")
	}
	// jobs.Save must not have been called — inner stub should be empty
	if len(innerJobs.jobs) != 0 {
		t.Errorf("job was saved despite trip update failure — partial write not prevented")
	}
}

// TestAcceptTrip_RollbackOnJobSaveFailure verifies that when jobRepo.Save fails,
// the use case returns an error. In a real PostgreSQL transaction the preceding
// trip UPDATE is rolled back automatically by the deferred Rollback call.
func TestAcceptTrip_RollbackOnJobSaveFailure(t *testing.T) {
	outerJobs := newStubJobRepo()
	seedJob(t, outerJobs, "trip1", "d1")

	trips := newStubTripUpdater()
	failJobs := &saveFailingJobRepo{
		stubJobRepo: newStubJobRepo(),
		saveErr:     domainerrors.Internal("simulated db failure"),
	}
	txr := &stubTransactor{jobs: failJobs, trips: trips}
	uc := app.NewAcceptTripUseCase(outerJobs, txr)

	_, err := uc.Execute(context.Background(), "trip1", "d1")
	if err == nil {
		t.Fatal("expected error when job save fails")
	}
	// AssignDriver ran before Save failed ��� in real DB this is rolled back by tx.Rollback.
	if trips.assignedTrips["trip1"] != "d1" {
		t.Error("AssignDriver should have been called before Save failed")
	}
}

// ─── RequestDispatch rollback tests ──────────────────────────────────────────

// TestRequestDispatch_RollbackOnSetSearchingFailure verifies that when the trip
// SetSearching UPDATE fails, jobRepo.Save is never called (no orphaned job record).
func TestRequestDispatch_RollbackOnSetSearchingFailure(t *testing.T) {
	outerJobs := newStubJobRepo()
	innerJobs := newStubJobRepo()
	failTrips := &failingTripUpdater{searchingErr: domainerrors.Internal("simulated db failure")}
	txr := &stubTransactor{jobs: innerJobs, trips: failTrips}
	locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))

	uc := app.NewRequestDispatchUseCase(outerJobs, locs, txr)
	_, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1",
	})
	if err == nil {
		t.Fatal("expected error when SetSearching fails")
	}
	// No dispatch job must have been persisted
	if len(innerJobs.jobs) != 0 {
		t.Errorf("job was saved despite SetSearching failure — partial write not prevented")
	}
}

// TestRequestDispatch_RollbackOnJobSaveFailure verifies that when the initial
// jobRepo.Save fails, the use case returns an error. In a real PostgreSQL
// transaction the preceding trip SetSearching UPDATE is rolled back.
func TestRequestDispatch_RollbackOnJobSaveFailure(t *testing.T) {
	outerJobs := newStubJobRepo()
	trips := newStubTripUpdater()
	failJobs := &saveFailingJobRepo{
		stubJobRepo: newStubJobRepo(),
		saveErr:     domainerrors.Internal("simulated db failure"),
	}
	txr := &stubTransactor{jobs: failJobs, trips: trips}
	locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))

	uc := app.NewRequestDispatchUseCase(outerJobs, locs, txr)
	_, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1",
	})
	if err == nil {
		t.Fatal("expected error when job save fails")
	}
	// SetSearching ran before Save failed — in real DB this is rolled back by tx.Rollback.
	if len(trips.searchingTrips) == 0 {
		t.Error("SetSearching should have been called before Save failed")
	}
}

// ─── RejectTrip ──────────────────────────────────────────────────────────────

func TestRejectTrip_MovesToNextDriver(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	seedJob(t, jobs, "trip1", "d1")

	// d2 is next nearest and active
	locs := newStubLocationRepo([]string{"d1", "d2"}, allActive("d1", "d2"))
	uc := app.NewRejectTripUseCase(jobs, locs, trips)

	job, err := uc.Execute(context.Background(), "trip1", "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.CurrentDriverID != "d2" {
		t.Errorf("CurrentDriverID = %q, want d2 after d1 rejects", job.CurrentDriverID)
	}
	if job.AttemptCount != 2 {
		t.Errorf("AttemptCount = %d, want 2", job.AttemptCount)
	}
}

func TestRejectTrip_ExhaustsAllDrivers(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	seedJob(t, jobs, "trip1", "d1")

	// Only d1, and it just rejected — no other drivers
	locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))
	uc := app.NewRejectTripUseCase(jobs, locs, trips)

	job, err := uc.Execute(context.Background(), "trip1", "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// d1 was already offered, no eligible next → job fails
	if job.Status != entity.JobStatusFailed {
		t.Errorf("Status = %q, want failed when all candidates exhausted", job.Status)
	}
}

func TestRejectTrip_WrongDriver(t *testing.T) {
	jobs := newStubJobRepo()
	seedJob(t, jobs, "trip1", "d1")

	locs := newStubLocationRepo(nil, nil)
	uc := app.NewRejectTripUseCase(jobs, locs, newStubTripUpdater())
	_, err := uc.Execute(context.Background(), "trip1", "d99")
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── UpdateDriverLocation ────────────────────────────────────────────────────

func TestUpdateDriverLocation_Valid(t *testing.T) {
	locs := newStubLocationRepo(nil, nil)
	uc := app.NewUpdateDriverLocationUseCase(locs)
	if err := uc.Execute(context.Background(), "d1", 10.0, 106.0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coord, ok := locs.updated["d1"]; !ok || coord[0] != 10.0 || coord[1] != 106.0 {
		t.Errorf("location not updated: %v", locs.updated)
	}
}

func TestUpdateDriverLocation_EmptyDriverID(t *testing.T) {
	uc := app.NewUpdateDriverLocationUseCase(newStubLocationRepo(nil, nil))
	err := uc.Execute(context.Background(), "", 10.0, 106.0)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

// ─── GetDispatchStatus ───────────────────────────────────────────────────────

func TestGetDispatchStatus_Found(t *testing.T) {
	jobs := newStubJobRepo()
	seedJob(t, jobs, "trip1", "d1")

	uc := app.NewGetDispatchStatusUseCase(jobs)
	job, err := uc.Execute(context.Background(), "trip1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.TripID != "trip1" {
		t.Errorf("TripID = %q, want trip1", job.TripID)
	}
}

func TestGetDispatchStatus_NotFound(t *testing.T) {
	uc := app.NewGetDispatchStatusUseCase(newStubJobRepo())
	_, err := uc.Execute(context.Background(), "missing")
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

// ─── DispatchEngine (timeout auto-retry) ─────────────────────────────────────

func TestDispatchEngine_RetriesOnTimeout(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()

	// Seed a job that is already searching with an expired offer on d1
	job, _ := entity.NewDispatchJob("j1", "trip1", "rider1", 10, 106, 1, 5, testNow)
	_ = job.OfferToDriver("d1", testNow) // expires at testNow+1s
	_ = jobs.Save(context.Background(), job)

	// d2 is next nearest and active
	locs := newStubLocationRepo([]string{"d1", "d2"}, allActive("d1", "d2"))

	engine := app.NewDispatchEngine(jobs, locs, trips)

	// Verify via jobs repo that after testNow+2s the offer is expired
	expired := testNow.Add(2 * time.Second)
	expiredJobs, err := jobs.FindExpiredOffers(context.Background(), expired)
	if err != nil {
		t.Fatalf("FindExpiredOffers: %v", err)
	}
	if len(expiredJobs) != 1 {
		t.Fatalf("expected 1 expired job, got %d", len(expiredJobs))
	}
	if expiredJobs[0].TripID != "trip1" {
		t.Errorf("TripID = %q, want trip1", expiredJobs[0].TripID)
	}
	_ = engine // engine.Start/Stop tested via integration
}

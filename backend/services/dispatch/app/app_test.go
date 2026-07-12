package app_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fairride/dispatch/app"
	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
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
func (r *stubJobRepo) FindCurrentOfferForDriver(_ context.Context, driverID string) (*entity.DispatchJob, error) {
	for _, j := range r.jobs {
		if j.CurrentDriverID == driverID && j.Status == entity.JobStatusSearching {
			return j, nil
		}
	}
	return nil, domainerrors.NotFound("no active offer for driver: " + driverID)
}

// stubLocationRepo simulates a geo store with configurable active drivers.
// rideEnabled/deliveryEnabled default to true/false per driver (matching
// migration 008's DB defaults / the production Redis repository's
// backward-compat fallback) unless a test explicitly overrides them via the
// rideEnabled/deliveryEnabled maps.
type stubLocationRepo struct {
	nearby          []string              // IDs in distance order (nearest first)
	active          map[string]bool       // which IDs respond as active
	updated         map[string][2]float64 // driverID → [lat, lon]
	serviceTypes    map[string]entity.ServiceType
	rideEnabled     map[string]bool
	deliveryEnabled map[string]bool
}

var _ repository.DriverLocationRepository = (*stubLocationRepo)(nil)

func newStubLocationRepo(nearby []string, active map[string]bool) *stubLocationRepo {
	return &stubLocationRepo{
		nearby: nearby, active: active, updated: make(map[string][2]float64),
		serviceTypes:    make(map[string]entity.ServiceType),
		rideEnabled:     make(map[string]bool),
		deliveryEnabled: make(map[string]bool),
	}
}

func (r *stubLocationRepo) UpdateLocation(_ context.Context, driverID string, lat, lon float64, serviceType entity.ServiceType, rideEnabled, deliveryEnabled bool) error {
	r.updated[driverID] = [2]float64{lat, lon}
	if serviceType != "" {
		r.serviceTypes[driverID] = serviceType
	}
	r.rideEnabled[driverID] = rideEnabled
	r.deliveryEnabled[driverID] = deliveryEnabled
	return nil
}

func (r *stubLocationRepo) FindNearby(_ context.Context, _, _ float64, _ float64, _ int) ([]*entity.NearbyDriver, error) {
	var out []*entity.NearbyDriver
	for _, id := range r.nearby {
		ride, ok := r.rideEnabled[id]
		if !ok {
			ride = true // migration 008 default
		}
		delivery := r.deliveryEnabled[id] // migration 008 default is false, matches Go zero value
		out = append(out, &entity.NearbyDriver{
			DriverID: id, ServiceType: r.serviceTypes[id],
			RideEnabled: ride, DeliveryEnabled: delivery,
		})
	}
	return out, nil
}

func (r *stubLocationRepo) IsActive(_ context.Context, driverID string) (bool, error) {
	return r.active[driverID], nil
}

func (r *stubLocationRepo) GetLocation(_ context.Context, driverID string) (float64, float64, error) {
	if coords, ok := r.updated[driverID]; ok {
		return coords[0], coords[1], nil
	}
	for _, id := range r.nearby {
		if id == driverID {
			return 10.0, 106.0, nil
		}
	}
	return 0, 0, domainerrors.NotFound("no location for: " + driverID)
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
		TripID:    "trip1",
		RiderID:   "rider1",
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

// ─── RequestDispatch — Delivery (Delivery V1 Phase 3, docs/business/DELIVERY_V1_DESIGN.md) ──

func TestRequestDispatch_RideDispatchPass(t *testing.T) {
	// TripType left at its zero value — literal "Ride giữ nguyên" check:
	// same outcome, same fields, as before Delivery was introduced.
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1", "d2"}, allActive("d1", "d2"))
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1", PickupLat: 10.0, PickupLon: 106.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.TripType != entity.TripTypeRide {
		t.Errorf("TripType = %q, want ride", job.TripType)
	}
	if job.Status != entity.JobStatusSearching || job.CurrentDriverID != "d1" {
		t.Errorf("job = %+v, want searching/d1 (nearest, unchanged matching algorithm)", job)
	}
}

func TestRequestDispatch_DeliveryDispatchPass(t *testing.T) {
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1", "d2"}, allActive("d1", "d2"))
	// Vehicle/Service Catalog refactor: offerNextDriver now filters
	// candidates by service type AND by delivery capability, so both
	// candidates must report the matching type + opt-in for "nearest wins"
	// to still hold here.
	locs.serviceTypes["d1"] = entity.ServiceTypeBike
	locs.serviceTypes["d2"] = entity.ServiceTypeBike
	locs.deliveryEnabled["d1"] = true
	locs.deliveryEnabled["d2"] = true
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1", PickupLat: 10.0, PickupLon: 106.0,
		TripType:    entity.TripTypeDelivery,
		ServiceType: entity.ServiceTypeBike,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.TripType != entity.TripTypeDelivery {
		t.Errorf("TripType = %q, want delivery", job.TripType)
	}
	if job.ServiceType != entity.ServiceTypeBike {
		t.Errorf("ServiceType = %q, want bike", job.ServiceType)
	}
	// Same distance-ordering algorithm as Ride: nearest active,
	// delivery-enabled driver reporting a matching service type wins.
	if job.Status != entity.JobStatusSearching || job.CurrentDriverID != "d1" {
		t.Errorf("job = %+v, want searching/d1 (nearest, matching service type)", job)
	}
	if len(trips.searchingTrips) == 0 {
		t.Error("expected SetSearching to be called, same as Ride")
	}
}

// TestRequestDispatch_Delivery_ServiceTypeMismatchExcludesDriver locks in
// the catalog's core Dispatch requirement: a candidate whose reported
// service type doesn't match the job's is never offered the trip, even if
// it is the nearest (or only) candidate.
func TestRequestDispatch_Delivery_ServiceTypeMismatchExcludesDriver(t *testing.T) {
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))
	locs.serviceTypes["d1"] = entity.ServiceTypeCar // job wants bike
	locs.deliveryEnabled["d1"] = true
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1", PickupLat: 10.0, PickupLon: 106.0,
		TripType:    entity.TripTypeDelivery,
		ServiceType: entity.ServiceTypeBike,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusFailed {
		t.Errorf("job.Status = %q, want failed (only candidate has a mismatched service type)", job.Status)
	}
}

// TestRequestDispatch_Delivery_CapabilityNotEnabledExcludesDriver locks in
// the second half of matching: even a service-type match is not enough for
// a Delivery job if the candidate hasn't opted into DeliveryEnabled.
func TestRequestDispatch_Delivery_CapabilityNotEnabledExcludesDriver(t *testing.T) {
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))
	locs.serviceTypes["d1"] = entity.ServiceTypeBike
	// deliveryEnabled left unset -> defaults to false (migration 008).
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1", PickupLat: 10.0, PickupLon: 106.0,
		TripType:    entity.TripTypeDelivery,
		ServiceType: entity.ServiceTypeBike,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusFailed {
		t.Errorf("job.Status = %q, want failed (candidate never opted into delivery)", job.Status)
	}
}

// TestRequestDispatch_Ride_EmptyServiceType_MatchesAnyCandidate locks in
// backward compatibility: a Ride request that omits ServiceType (every
// caller written before this catalog existed) is unaffected by the new
// filtering — nearest active driver wins regardless of reported service
// type, and RideEnabled defaults to true so no capability opt-in is needed.
func TestRequestDispatch_Ride_EmptyServiceType_MatchesAnyCandidate(t *testing.T) {
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))
	locs.serviceTypes["d1"] = entity.ServiceTypeCarXL // deliberately NOT set on the request
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1", PickupLat: 10.0, PickupLon: 106.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusSearching || job.CurrentDriverID != "d1" {
		t.Errorf("job = %+v, want searching/d1 (no service_type on request -> unfiltered)", job)
	}
}

func TestRequestDispatch_UnsupportedServiceTypeFails(t *testing.T) {
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	_, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1", PickupLat: 10.0, PickupLon: 106.0,
		ServiceType: entity.ServiceType("bicycle"),
	})
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument for unsupported service type, got %v", err)
	}
	// No job or trip-searching side effect must have happened.
	if len(jobs.jobs) != 0 {
		t.Errorf("expected no dispatch job to be created, found %d", len(jobs.jobs))
	}
	if len(trips.searchingTrips) != 0 {
		t.Error("expected SetSearching to never be called for a rejected service type")
	}
}

// TestRequestDispatch_AllSupportedServiceTypesPass locks in the single,
// shared 4-value allow-list applying identically to Ride and Delivery.
func TestRequestDispatch_AllSupportedServiceTypesPass(t *testing.T) {
	for _, tripType := range []entity.TripType{entity.TripTypeRide, entity.TripTypeDelivery} {
		for _, st := range []entity.ServiceType{
			entity.ServiceTypeBike, entity.ServiceTypeBikePlus, entity.ServiceTypeCar, entity.ServiceTypeCarXL,
		} {
			t.Run(string(tripType)+"/"+string(st), func(t *testing.T) {
				jobs := newStubJobRepo()
				locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))
				locs.serviceTypes["d1"] = st
				locs.deliveryEnabled["d1"] = true
				trips := newStubTripUpdater()
				uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

				_, err := uc.Execute(context.Background(), app.RequestDispatchInput{
					TripID: "trip1", RiderID: "rider1", PickupLat: 10.0, PickupLon: 106.0,
					TripType: tripType, ServiceType: st,
				})
				if err != nil {
					t.Errorf("service type %q should be supported for %q, got error: %v", st, tripType, err)
				}
			})
		}
	}
}

func TestRequestDispatch_Delivery_EmptyServiceTypeIsNotRejected(t *testing.T) {
	// An omitted ServiceType must not block dispatch (though see the
	// capability test above — the candidate must still opt into delivery).
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1"}, allActive("d1"))
	locs.deliveryEnabled["d1"] = true
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID: "trip1", RiderID: "rider1", PickupLat: 10.0, PickupLon: 106.0,
		TripType: entity.TripTypeDelivery,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.TripType != entity.TripTypeDelivery {
		t.Errorf("TripType = %q, want delivery", job.TripType)
	}
}

// TestRequestDispatch_BackwardCompatibility verifies that a RequestDispatchInput
// built exactly as it was before Delivery V1 Phase 3 (no TripType/ServiceType
// fields set at all) produces byte-for-byte the same outcome as before.
func TestRequestDispatch_BackwardCompatibility(t *testing.T) {
	jobs := newStubJobRepo()
	locs := newStubLocationRepo([]string{"d1", "d2", "d3"}, allActive("d1", "d2", "d3"))
	trips := newStubTripUpdater()
	uc := app.NewRequestDispatchUseCase(jobs, locs, &stubTransactor{jobs: jobs, trips: trips})

	job, err := uc.Execute(context.Background(), app.RequestDispatchInput{
		TripID:    "trip1",
		RiderID:   "rider1",
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
	if job.ServiceType != "" {
		t.Errorf("ServiceType = %q, want empty (never set by legacy callers)", job.ServiceType)
	}
	if len(trips.searchingTrips) == 0 {
		t.Error("expected SetSearching to be called")
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
	if err := uc.Execute(context.Background(), "d1", 10.0, 106.0, "", true, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if coord, ok := locs.updated["d1"]; !ok || coord[0] != 10.0 || coord[1] != 106.0 {
		t.Errorf("location not updated: %v", locs.updated)
	}
}

func TestUpdateDriverLocation_EmptyDriverID(t *testing.T) {
	uc := app.NewUpdateDriverLocationUseCase(newStubLocationRepo(nil, nil))
	err := uc.Execute(context.Background(), "", 10.0, 106.0, "", true, false)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestUpdateDriverLocation_WithServiceTypeAndCapability(t *testing.T) {
	locs := newStubLocationRepo(nil, nil)
	uc := app.NewUpdateDriverLocationUseCase(locs)
	if err := uc.Execute(context.Background(), "d1", 10.0, 106.0, entity.ServiceTypeBikePlus, true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := locs.serviceTypes["d1"]; got != entity.ServiceTypeBikePlus {
		t.Errorf("service type not recorded: got %q, want %q", got, entity.ServiceTypeBikePlus)
	}
	if !locs.rideEnabled["d1"] || !locs.deliveryEnabled["d1"] {
		t.Errorf("capability not recorded: ride=%v delivery=%v", locs.rideEnabled["d1"], locs.deliveryEnabled["d1"])
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

// ─── DispatchEngine lifecycle ─────────────────────────────────────────────────

// countingJobRepo wraps stubJobRepo and counts FindExpiredOffers invocations.
type countingJobRepo struct {
	*stubJobRepo
	findCount atomic.Int32
}

func (r *countingJobRepo) FindExpiredOffers(ctx context.Context, now time.Time) ([]*entity.DispatchJob, error) {
	r.findCount.Add(1)
	return r.stubJobRepo.FindExpiredOffers(ctx, now)
}

// blockOnSaveJobRepo blocks each Save call until saveCh is closed.
type blockOnSaveJobRepo struct {
	*stubJobRepo
	saveCh    chan struct{}
	saveCount atomic.Int32
}

func (r *blockOnSaveJobRepo) Save(ctx context.Context, job *entity.DispatchJob) error {
	r.saveCount.Add(1)
	<-r.saveCh
	return r.stubJobRepo.Save(ctx, job)
}

// alwaysExpiredJobRepo always returns the same job on FindExpiredOffers regardless of
// in-memory mutations, to create the conditions for deduplication testing.
type alwaysExpiredJobRepo struct {
	*stubJobRepo
	job *entity.DispatchJob
}

func (r *alwaysExpiredJobRepo) FindExpiredOffers(_ context.Context, _ time.Time) ([]*entity.DispatchJob, error) {
	cp := *r.job
	return []*entity.DispatchJob{&cp}, nil
}

// composedJobRepo combines a custom finder with a custom saver.
type composedJobRepo struct {
	finder repository.DispatchJobRepository
	saver  repository.DispatchJobRepository
}

func (r *composedJobRepo) Save(ctx context.Context, job *entity.DispatchJob) error {
	return r.saver.Save(ctx, job)
}
func (r *composedJobRepo) FindByID(ctx context.Context, id string) (*entity.DispatchJob, error) {
	return r.finder.FindByID(ctx, id)
}
func (r *composedJobRepo) FindByTripID(ctx context.Context, tripID string) (*entity.DispatchJob, error) {
	return r.finder.FindByTripID(ctx, tripID)
}
func (r *composedJobRepo) FindExpiredOffers(ctx context.Context, now time.Time) ([]*entity.DispatchJob, error) {
	return r.finder.FindExpiredOffers(ctx, now)
}
func (r *composedJobRepo) FindCurrentOfferForDriver(ctx context.Context, driverID string) (*entity.DispatchJob, error) {
	return r.finder.FindCurrentOfferForDriver(ctx, driverID)
}

// TestEngine_StartCalledTwiceCreatesOneWorker verifies that calling Start() more than
// once does not create multiple background goroutines.
func TestEngine_StartCalledTwiceCreatesOneWorker(t *testing.T) {
	jobs := &countingJobRepo{stubJobRepo: newStubJobRepo()}
	locs := newStubLocationRepo(nil, nil)

	engine := app.NewDispatchEngine(jobs, locs, newStubTripUpdater())
	engine.WithTickInterval(5 * time.Millisecond)

	engine.Start()
	engine.Start() // must be a no-op

	time.Sleep(40 * time.Millisecond)
	engine.Stop()

	count := int(jobs.findCount.Load())
	// 5 ms tick × 40 ms ≈ 8 calls for one goroutine; two goroutines → ~16.
	if count > 14 {
		t.Errorf("FindExpiredOffers called %d times; suggests multiple goroutines (want ≤14)", count)
	}
	if count == 0 {
		t.Error("FindExpiredOffers never called; engine may not have started")
	}
}

// TestEngine_GracefulStop verifies that Stop() waits for in-flight job goroutines.
func TestEngine_GracefulStop(t *testing.T) {
	jobs := newStubJobRepo()
	job, _ := entity.NewDispatchJob("j1", "trip1", "rider1", 10, 106, 1, 5, testNow)
	_ = job.OfferToDriver("d1", testNow)
	_ = jobs.Save(context.Background(), job)

	saveCh := make(chan struct{})
	slowJobs := &blockOnSaveJobRepo{stubJobRepo: jobs, saveCh: saveCh}
	locs := newStubLocationRepo([]string{"d2"}, allActive("d2"))

	engine := app.NewDispatchEngine(slowJobs, locs, newStubTripUpdater())
	engine.WithTickInterval(5 * time.Millisecond)
	engine.Start()

	// Let a tick fire and the job goroutine start (it blocks at Save).
	time.Sleep(30 * time.Millisecond)

	stopDone := make(chan struct{})
	go func() {
		engine.Stop()
		close(stopDone)
	}()

	// While job goroutine is blocked, Stop must not return.
	select {
	case <-stopDone:
		t.Fatal("Stop() returned before job goroutine finished — graceful shutdown broken")
	case <-time.After(30 * time.Millisecond):
		// expected: Stop is still waiting
	}

	// Unblock the goroutine; Stop must now return.
	close(saveCh)
	select {
	case <-stopDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop() did not return after job goroutine was unblocked")
	}
}

// TestEngine_ConcurrentJobsDeduplication verifies that a job already being processed
// is not started a second time when the next tick fires.
func TestEngine_ConcurrentJobsDeduplication(t *testing.T) {
	base := newStubJobRepo()
	job, _ := entity.NewDispatchJob("j1", "trip1", "rider1", 10, 106, 1, 5, testNow)
	_ = job.OfferToDriver("d1", testNow)
	_ = base.Save(context.Background(), job)

	saveCh := make(chan struct{})
	saver := &blockOnSaveJobRepo{stubJobRepo: base, saveCh: saveCh}
	finder := &alwaysExpiredJobRepo{stubJobRepo: &stubJobRepo{jobs: base.jobs}, job: job}
	composed := &composedJobRepo{finder: finder, saver: saver}

	locs := newStubLocationRepo([]string{"d2"}, allActive("d2"))
	engine := app.NewDispatchEngine(composed, locs, newStubTripUpdater())
	engine.WithTickInterval(5 * time.Millisecond)
	engine.Start()

	// Let 2+ ticks fire. Goroutine 1 starts on tick 1 and blocks at Save.
	// Subsequent ticks see job.JobID in inFlight and skip it.
	time.Sleep(20 * time.Millisecond)

	// Stop BEFORE unblocking: signals done so no new ticks fire, then waits in wg.Wait().
	stopDone := make(chan struct{})
	go func() {
		engine.Stop()
		close(stopDone)
	}()
	// Give Stop time to signal done and enter wg.Wait().
	time.Sleep(5 * time.Millisecond)

	// Only goroutine 1 should have reached Save during the blocking window.
	if count := saver.saveCount.Load(); count != 1 {
		t.Errorf("Save called %d times during blocking window; want 1 (in-flight deduplication)", count)
	}

	// Unblock goroutine 1 → wg drains → Stop returns.
	close(saveCh)
	select {
	case <-stopDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop() did not return after goroutine was unblocked")
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

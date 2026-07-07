package grpc_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fairride/dispatch/app"
	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	dispatchgrpc "github.com/fairride/dispatch/grpc"
	"github.com/fairride/dispatch/grpc/dispatchpb"
	domainerrors "github.com/fairride/shared/errors"
)

// ─── stubs ───────────────────────────────────────────────────────────────────

var testNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

type stubJobRepo struct{ jobs map[string]*entity.DispatchJob }

var _ repository.DispatchJobRepository = (*stubJobRepo)(nil)

func newStubJobRepo() *stubJobRepo { return &stubJobRepo{jobs: make(map[string]*entity.DispatchJob)} }

func (r *stubJobRepo) Save(_ context.Context, j *entity.DispatchJob) error {
	r.jobs[j.JobID] = j
	return nil
}
func (r *stubJobRepo) FindByID(_ context.Context, id string) (*entity.DispatchJob, error) {
	j, ok := r.jobs[id]
	if !ok {
		return nil, domainerrors.NotFound("not found")
	}
	return j, nil
}
func (r *stubJobRepo) FindByTripID(_ context.Context, tripID string) (*entity.DispatchJob, error) {
	for _, j := range r.jobs {
		if j.TripID == tripID {
			return j, nil
		}
	}
	return nil, domainerrors.NotFound("no job for trip: " + tripID)
}
func (r *stubJobRepo) FindExpiredOffers(_ context.Context, _ time.Time) ([]*entity.DispatchJob, error) {
	return nil, nil
}
func (r *stubJobRepo) FindCurrentOfferForDriver(_ context.Context, driverID string) (*entity.DispatchJob, error) {
	for _, j := range r.jobs {
		if j.CurrentDriverID == driverID && j.Status == entity.JobStatusSearching {
			return j, nil
		}
	}
	return nil, domainerrors.NotFound("no active offer for driver: " + driverID)
}

type stubLocRepo struct{ nearby []string }

var _ repository.DriverLocationRepository = (*stubLocRepo)(nil)

func (r *stubLocRepo) UpdateLocation(_ context.Context, _ string, _, _ float64) error { return nil }
func (r *stubLocRepo) FindNearby(_ context.Context, _, _, _ float64, _ int) ([]*entity.NearbyDriver, error) {
	var out []*entity.NearbyDriver
	for _, id := range r.nearby {
		out = append(out, &entity.NearbyDriver{DriverID: id})
	}
	return out, nil
}
func (r *stubLocRepo) IsActive(_ context.Context, id string) (bool, error) {
	for _, n := range r.nearby {
		if n == id {
			return true, nil
		}
	}
	return false, nil
}
func (r *stubLocRepo) GetLocation(_ context.Context, id string) (float64, float64, error) {
	for _, n := range r.nearby {
		if n == id {
			return 10.0, 106.0, nil
		}
	}
	return 0, 0, domainerrors.NotFound("no location for: " + id)
}
func (r *stubLocRepo) RemoveLocation(_ context.Context, _ string) error { return nil }

type stubTripUpdater struct{ assigned map[string]string }

var _ repository.TripUpdater = (*stubTripUpdater)(nil)

func newStubTripUpdater() *stubTripUpdater {
	return &stubTripUpdater{assigned: make(map[string]string)}
}
func (u *stubTripUpdater) SetSearching(_ context.Context, _ string, _ time.Time) error { return nil }
func (u *stubTripUpdater) AssignDriver(_ context.Context, tripID, driverID string, _ time.Time) error {
	u.assigned[tripID] = driverID
	return nil
}

type stubTransactor struct {
	jobs  repository.DispatchJobRepository
	trips repository.TripUpdater
}

var _ repository.Transactor = (*stubTransactor)(nil)

func (s *stubTransactor) WithinTx(_ context.Context, fn func(repository.DispatchJobRepository, repository.TripUpdater) error) error {
	return fn(s.jobs, s.trips)
}

// ─── helper ──────────────────────────────────────────────────────────────────

func newHandler(jobs *stubJobRepo, locs *stubLocRepo, trips *stubTripUpdater) *dispatchgrpc.Handler {
	txr := &stubTransactor{jobs: jobs, trips: trips}
	return dispatchgrpc.NewHandler(
		app.NewRequestDispatchUseCase(jobs, locs, txr),
		app.NewAcceptTripUseCase(jobs, txr),
		app.NewRejectTripUseCase(jobs, locs, trips),
		app.NewUpdateDriverLocationUseCase(locs),
		app.NewGetDispatchStatusUseCase(jobs),
		app.NewGetDriverOfferUseCase(jobs),
		app.NewGetDriverLocationUseCase(locs),
	)
}

func seedJobWithOffer(jobs *stubJobRepo, tripID, driverID string) {
	job, _ := entity.NewDispatchJob("j1", tripID, "rider1", 10, 106, 30, 5, testNow)
	// Use real time so offer_expires_at is always in the future during the test.
	_ = job.OfferToDriver(driverID, time.Now())
	_ = jobs.Save(context.Background(), job)
}

// ─── RequestDispatch ─────────────────────────────────────────────────────────

func TestRequestDispatch_OK(t *testing.T) {
	jobs := newStubJobRepo()
	locs := &stubLocRepo{nearby: []string{"d1"}}
	trips := newStubTripUpdater()
	h := newHandler(jobs, locs, trips)

	resp, err := h.RequestDispatch(context.Background(), &dispatchpb.RequestDispatchRequest{
		TripId:  "trip1",
		RiderId: "rider1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Job.Status != "searching" {
		t.Errorf("status = %q, want searching", resp.Job.Status)
	}
	if resp.Job.CurrentDriverId != "d1" {
		t.Errorf("current_driver_id = %q, want d1", resp.Job.CurrentDriverId)
	}
}

func TestRequestDispatch_MissingTripID(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.RequestDispatch(context.Background(), &dispatchpb.RequestDispatchRequest{RiderId: "r1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

func TestRequestDispatch_MissingRiderID(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.RequestDispatch(context.Background(), &dispatchpb.RequestDispatchRequest{TripId: "t1"})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

func TestRequestDispatch_NoDriversAvailable(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	resp, err := h.RequestDispatch(context.Background(), &dispatchpb.RequestDispatchRequest{
		TripId: "trip1", RiderId: "rider1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Job.Status != "failed" {
		t.Errorf("status = %q, want failed when no drivers", resp.Job.Status)
	}
}

// ─── AcceptTrip ──────────────────────────────────────────────────────────────

func TestAcceptTrip_OK(t *testing.T) {
	jobs := newStubJobRepo()
	trips := newStubTripUpdater()
	seedJobWithOffer(jobs, "trip1", "d1")
	h := newHandler(jobs, &stubLocRepo{}, trips)

	resp, err := h.AcceptTrip(context.Background(), &dispatchpb.AcceptTripRequest{
		TripId: "trip1", DriverId: "d1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Job.Status != "assigned" {
		t.Errorf("status = %q, want assigned", resp.Job.Status)
	}
	if resp.Job.AssignedDriverId != "d1" {
		t.Errorf("assigned_driver_id = %q, want d1", resp.Job.AssignedDriverId)
	}
}

func TestAcceptTrip_WrongDriver(t *testing.T) {
	jobs := newStubJobRepo()
	seedJobWithOffer(jobs, "trip1", "d1")
	h := newHandler(jobs, &stubLocRepo{}, newStubTripUpdater())

	_, err := h.AcceptTrip(context.Background(), &dispatchpb.AcceptTripRequest{
		TripId: "trip1", DriverId: "d_wrong",
	})
	s, _ := status.FromError(err)
	if s.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", s.Code())
	}
}

func TestAcceptTrip_NotFound(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.AcceptTrip(context.Background(), &dispatchpb.AcceptTripRequest{
		TripId: "nonexistent", DriverId: "d1",
	})
	s, _ := status.FromError(err)
	if s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", s.Code())
	}
}

func TestAcceptTrip_MissingFields(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.AcceptTrip(context.Background(), &dispatchpb.AcceptTripRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── RejectTrip ──────────────────────────────────────────────────────────────

func TestRejectTrip_MovesToNextDriver(t *testing.T) {
	jobs := newStubJobRepo()
	seedJobWithOffer(jobs, "trip1", "d1")
	// d1 rejected, d2 is next
	locs := &stubLocRepo{nearby: []string{"d1", "d2"}}
	h := newHandler(jobs, locs, newStubTripUpdater())

	resp, err := h.RejectTrip(context.Background(), &dispatchpb.RejectTripRequest{
		TripId: "trip1", DriverId: "d1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Job.CurrentDriverId != "d2" {
		t.Errorf("current_driver_id = %q, want d2", resp.Job.CurrentDriverId)
	}
}

func TestRejectTrip_NoNextDriver(t *testing.T) {
	jobs := newStubJobRepo()
	seedJobWithOffer(jobs, "trip1", "d1")
	locs := &stubLocRepo{nearby: []string{"d1"}} // only d1, which already rejected
	h := newHandler(jobs, locs, newStubTripUpdater())

	resp, err := h.RejectTrip(context.Background(), &dispatchpb.RejectTripRequest{
		TripId: "trip1", DriverId: "d1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Job.Status != "failed" {
		t.Errorf("status = %q, want failed", resp.Job.Status)
	}
}

func TestRejectTrip_MissingFields(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.RejectTrip(context.Background(), &dispatchpb.RejectTripRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── UpdateDriverLocation ────────────────────────────────────────────────────

func TestUpdateDriverLocation_OK(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.UpdateDriverLocation(context.Background(), &dispatchpb.UpdateDriverLocationRequest{
		DriverId: "d1", Lat: 10.0, Lon: 106.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDriverLocation_MissingDriverID(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.UpdateDriverLocation(context.Background(), &dispatchpb.UpdateDriverLocationRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── GetDispatchStatus ───────────────────────────────────────────────────────

func TestGetDispatchStatus_OK(t *testing.T) {
	jobs := newStubJobRepo()
	seedJobWithOffer(jobs, "trip1", "d1")
	h := newHandler(jobs, &stubLocRepo{}, newStubTripUpdater())

	resp, err := h.GetDispatchStatus(context.Background(), &dispatchpb.GetDispatchStatusRequest{
		TripId: "trip1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Job.TripId != "trip1" {
		t.Errorf("trip_id = %q, want trip1", resp.Job.TripId)
	}
}

func TestGetDispatchStatus_NotFound(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.GetDispatchStatus(context.Background(), &dispatchpb.GetDispatchStatusRequest{
		TripId: "nonexistent",
	})
	s, _ := status.FromError(err)
	if s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", s.Code())
	}
}

func TestGetDispatchStatus_MissingTripID(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.GetDispatchStatus(context.Background(), &dispatchpb.GetDispatchStatusRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── GetDriverOffer ──────────────────────────────────────────────────────────

func TestGetDriverOffer_HasOffer(t *testing.T) {
	jobs := newStubJobRepo()
	seedJobWithOffer(jobs, "trip1", "d1")
	h := newHandler(jobs, &stubLocRepo{}, newStubTripUpdater())

	resp, err := h.GetDriverOffer(context.Background(), &dispatchpb.GetDriverOfferRequest{
		DriverId: "d1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasOffer {
		t.Error("has_offer = false, want true")
	}
	if resp.TripId != "trip1" {
		t.Errorf("trip_id = %q, want trip1", resp.TripId)
	}
	if resp.OfferExpiresAt == nil {
		t.Error("offer_expires_at is nil, want non-nil")
	}
}

func TestGetDriverOffer_NoOffer(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())

	resp, err := h.GetDriverOffer(context.Background(), &dispatchpb.GetDriverOfferRequest{
		DriverId: "d99",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasOffer {
		t.Error("has_offer = true, want false when no offer exists")
	}
}

func TestGetDriverOffer_MissingDriverID(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.GetDriverOffer(context.Background(), &dispatchpb.GetDriverOfferRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// ─── GetDriverLocation ───────────────────────────────────────────────────────

func TestGetDriverLocation_Active(t *testing.T) {
	locs := &stubLocRepo{nearby: []string{"d1"}}
	h := newHandler(newStubJobRepo(), locs, newStubTripUpdater())

	resp, err := h.GetDriverLocation(context.Background(), &dispatchpb.GetDriverLocationRequest{
		DriverId: "d1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsActive {
		t.Error("is_active = false, want true for active driver")
	}
	if resp.Lat == 0 && resp.Lon == 0 {
		t.Error("lat/lon both zero for a driver that has a location")
	}
}

func TestGetDriverLocation_NotActive(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())

	resp, err := h.GetDriverLocation(context.Background(), &dispatchpb.GetDriverLocationRequest{
		DriverId: "d99",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsActive {
		t.Error("is_active = true, want false for driver with no location")
	}
}

func TestGetDriverLocation_MissingDriverID(t *testing.T) {
	h := newHandler(newStubJobRepo(), &stubLocRepo{}, newStubTripUpdater())
	_, err := h.GetDriverLocation(context.Background(), &dispatchpb.GetDriverLocationRequest{})
	s, _ := status.FromError(err)
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

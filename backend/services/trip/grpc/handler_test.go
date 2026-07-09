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
	tripgrpc "github.com/fairride/trip/grpc"
	"github.com/fairride/trip/grpc/trippb"
	"github.com/fairride/trip/domain/repository"
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
	return tripgrpc.NewHandler(
		app.NewCreateTripUseCase(repo),
		app.NewCancelTripUseCase(repo),
		app.NewGetTripUseCase(repo),
		app.NewMarkDriverArrivedUseCase(repo),
		app.NewStartTripUseCase(repo),
		app.NewCompleteTripUseCase(repo),
		app.NewInitiatePaymentUseCase(repo),
		app.NewPayTripUseCase(repo),
		app.NewListTripsByRiderUseCase(repo),
		app.NewListTripsByDriverUseCase(repo),
	)
}

func seedTrip(repo *stubRepo, tripID, riderID string, st entity.TripStatus) *entity.Trip {
	trip := entity.ReconstituteTrip(tripID, riderID, "", st, "pickup", "dropoff", "", 0, "", "", testNow, testNow)
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

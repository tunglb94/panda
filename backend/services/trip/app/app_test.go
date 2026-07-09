package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/shared/errors"
	"github.com/fairride/trip/app"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/domain/repository"
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
	uc := app.NewCreateTripUseCase(repo)

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
	uc := app.NewCreateTripUseCase(newStubRepo())
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
	uc := app.NewCreateTripUseCase(newStubRepo())
	_, err := uc.Execute(context.Background(), app.CreateTripInput{
		RiderID:        "r1",
		PickupAddress:  "",
		DropoffAddress: "dropoff",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

// ─── CancelTrip ──────────────────────────────────────────────────────────────

func makeTrip(repo *stubRepo, tripID, riderID string, status entity.TripStatus) *entity.Trip {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	trip := entity.ReconstituteTrip(tripID, riderID, "", status, "pickup", "dropoff", "", 0, "", "", now, now)
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

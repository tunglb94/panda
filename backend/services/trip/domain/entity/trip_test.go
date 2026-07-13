package entity_test

import (
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
)

var testNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// ─── NewTrip ─────────────────────────────────────────────────────────────────

func TestNewTrip_Valid(t *testing.T) {
	trip, err := entity.NewTrip("t1", "r1", "123 Main St", "456 Elm Ave", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.TripID != "t1" {
		t.Errorf("TripID = %q, want %q", trip.TripID, "t1")
	}
	if trip.RiderID != "r1" {
		t.Errorf("RiderID = %q, want %q", trip.RiderID, "r1")
	}
	if trip.Status != entity.StatusPending {
		t.Errorf("Status = %q, want %q", trip.Status, entity.StatusPending)
	}
	if trip.DriverID != "" {
		t.Errorf("DriverID should be empty, got %q", trip.DriverID)
	}
	if !trip.CreatedAt.Equal(testNow) {
		t.Errorf("CreatedAt = %v, want %v", trip.CreatedAt, testNow)
	}
}

func TestNewTrip_EmptyTripID(t *testing.T) {
	_, err := entity.NewTrip("", "r1", "pickup", "dropoff", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewTrip_EmptyRiderID(t *testing.T) {
	_, err := entity.NewTrip("t1", "", "pickup", "dropoff", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewTrip_EmptyPickup(t *testing.T) {
	_, err := entity.NewTrip("t1", "r1", "   ", "dropoff", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewTrip_EmptyDropoff(t *testing.T) {
	_, err := entity.NewTrip("t1", "r1", "pickup", "", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

// ─── Cancel ──────────────────────────────────────────────────────────────────

func tripAt(status entity.TripStatus) *entity.Trip {
	return entity.ReconstituteTrip("t1", "r1", "", status, "pickup", "dropoff", "", 0, "", "", testNow, testNow, entity.CompleteFinancials{})
}

func TestCancel_FromPending(t *testing.T) {
	trip := tripAt(entity.StatusPending)
	if err := trip.Cancel("rider changed mind", testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusCancelled {
		t.Errorf("Status = %q, want Cancelled", trip.Status)
	}
	if trip.CancellationReason != "rider changed mind" {
		t.Errorf("CancellationReason = %q", trip.CancellationReason)
	}
}

func TestCancel_FromSearching(t *testing.T) {
	trip := tripAt(entity.StatusSearching)
	if err := trip.Cancel("timeout", testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusCancelled {
		t.Errorf("Status = %q, want Cancelled", trip.Status)
	}
}

func TestCancel_FromDriverAssigned(t *testing.T) {
	trip := tripAt(entity.StatusDriverAssigned)
	if err := trip.Cancel("", testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusCancelled {
		t.Errorf("Status = %q, want Cancelled", trip.Status)
	}
}

func TestCancel_FromDriverArrived(t *testing.T) {
	trip := tripAt(entity.StatusDriverArrived)
	if err := trip.Cancel("", testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusCancelled {
		t.Errorf("Status = %q, want Cancelled", trip.Status)
	}
}

func TestCancel_FromInProgress(t *testing.T) {
	trip := tripAt(entity.StatusInProgress)
	err := trip.Cancel("", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCancel_FromCompleted(t *testing.T) {
	trip := tripAt(entity.StatusCompleted)
	err := trip.Cancel("", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCancel_AlreadyCancelled(t *testing.T) {
	trip := tripAt(entity.StatusCancelled)
	err := trip.Cancel("", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── Start ───────────────────────────────────────────────────────────────────

func TestStart_FromDriverAssigned(t *testing.T) {
	trip := tripAt(entity.StatusDriverAssigned)
	if err := trip.Start(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusInProgress {
		t.Errorf("Status = %q, want in_progress", trip.Status)
	}
}

func TestStart_FromDriverArrived(t *testing.T) {
	trip := tripAt(entity.StatusDriverArrived)
	if err := trip.Start(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Status != entity.StatusInProgress {
		t.Errorf("Status = %q, want in_progress", trip.Status)
	}
}

func TestStart_FromPendingFails(t *testing.T) {
	trip := tripAt(entity.StatusPending)
	err := trip.Start(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestStart_FromInProgressFails(t *testing.T) {
	trip := tripAt(entity.StatusInProgress)
	err := trip.Start(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── Complete ─────────────────────────────────────────────────────────────────

func TestComplete_FromInProgress(t *testing.T) {
	trip := tripAt(entity.StatusInProgress)
	if err := trip.Complete(325, "USD", entity.CompleteFinancials{}, testNow); err != nil {
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

func TestComplete_FromPendingFails(t *testing.T) {
	trip := tripAt(entity.StatusPending)
	err := trip.Complete(100, "USD", entity.CompleteFinancials{}, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestComplete_FromCompletedFails(t *testing.T) {
	trip := tripAt(entity.StatusCompleted)
	err := trip.Complete(100, "USD", entity.CompleteFinancials{}, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── Reconstitute ────────────────────────────────────────────────────────────

func TestReconstituteTrip_NoValidation(t *testing.T) {
	// Should not panic or error even with empty fields
	trip := entity.ReconstituteTrip("", "", "", entity.StatusCompleted, "", "", "some reason", 0, "", "", testNow, testNow, entity.CompleteFinancials{})
	if trip == nil {
		t.Fatal("expected non-nil trip")
	}
	if trip.Status != entity.StatusCompleted {
		t.Errorf("Status = %q, want Completed", trip.Status)
	}
}

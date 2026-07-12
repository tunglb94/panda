package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

var testNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func newJob(t *testing.T) *entity.DispatchJob {
	t.Helper()
	job, err := entity.NewDispatchJob("j1", "trip1", "rider1", 10.0, 106.0, 30, 5, testNow)
	if err != nil {
		t.Fatalf("NewDispatchJob: %v", err)
	}
	return job
}

// ─── NewDispatchJob ───────────────────────────────────────────────────────────

func TestNewDispatchJob_Valid(t *testing.T) {
	job := newJob(t)
	if job.Status != entity.JobStatusPending {
		t.Errorf("Status = %q, want pending", job.Status)
	}
	if job.OfferTimeoutSec != 30 {
		t.Errorf("OfferTimeoutSec = %d, want 30", job.OfferTimeoutSec)
	}
	if job.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", job.MaxAttempts)
	}
	if job.AttemptCount != 0 {
		t.Errorf("AttemptCount = %d, want 0", job.AttemptCount)
	}
}

// ─── TripType / ServiceType (Vehicle/Service Catalog refactor) ──

func TestNewDispatchJob_DefaultsTripTypeRide(t *testing.T) {
	job := newJob(t)
	if job.TripType != entity.TripTypeRide {
		t.Errorf("TripType = %q, want ride", job.TripType)
	}
	if job.ServiceType != "" {
		t.Errorf("ServiceType = %q, want empty", job.ServiceType)
	}
}

// TestServiceType_IsSupported locks in the single, shared allow-list used
// for both Ride and Delivery dispatch requests — no separate delivery-only
// list, since ServiceType no longer encodes TripType.
func TestServiceType_IsSupported(t *testing.T) {
	supported := []entity.ServiceType{entity.ServiceTypeBike, entity.ServiceTypeBikePlus, entity.ServiceTypeCar, entity.ServiceTypeCarXL}
	for _, st := range supported {
		if !st.IsSupported() {
			t.Errorf("%q should be supported", st)
		}
	}
	unsupported := []entity.ServiceType{"bicycle", "van", "truck", "delivery_bike", "delivery_car", ""}
	for _, st := range unsupported {
		if st.IsSupported() {
			t.Errorf("%q should not be supported", st)
		}
	}
}

func TestNewDispatchJob_DefaultsApplied(t *testing.T) {
	job, err := entity.NewDispatchJob("j1", "trip1", "rider1", 0, 0, 0, 0, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.OfferTimeoutSec != entity.DefaultOfferTimeoutSec {
		t.Errorf("OfferTimeoutSec = %d, want %d", job.OfferTimeoutSec, entity.DefaultOfferTimeoutSec)
	}
	if job.MaxAttempts != entity.DefaultMaxAttempts {
		t.Errorf("MaxAttempts = %d, want %d", job.MaxAttempts, entity.DefaultMaxAttempts)
	}
}

func TestNewDispatchJob_EmptyJobID(t *testing.T) {
	_, err := entity.NewDispatchJob("", "trip1", "rider1", 0, 0, 0, 0, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewDispatchJob_EmptyTripID(t *testing.T) {
	_, err := entity.NewDispatchJob("j1", "", "rider1", 0, 0, 0, 0, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestNewDispatchJob_EmptyRiderID(t *testing.T) {
	_, err := entity.NewDispatchJob("j1", "trip1", "", 0, 0, 0, 0, testNow)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

// ─── OfferToDriver ────────────────────────────────────────────────────────────

func TestOfferToDriver_FromPending(t *testing.T) {
	job := newJob(t)
	if err := job.OfferToDriver("d1", testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusSearching {
		t.Errorf("Status = %q, want searching", job.Status)
	}
	if job.CurrentDriverID != "d1" {
		t.Errorf("CurrentDriverID = %q, want d1", job.CurrentDriverID)
	}
	if job.AttemptCount != 1 {
		t.Errorf("AttemptCount = %d, want 1", job.AttemptCount)
	}
	if !job.HasBeenOffered("d1") {
		t.Error("d1 should be in OfferedDriverIDs")
	}
	if job.OfferExpiresAt.IsZero() {
		t.Error("OfferExpiresAt should not be zero")
	}
}

func TestOfferToDriver_ActiveOfferBlocks(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	err := job.OfferToDriver("d2", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestOfferToDriver_AfterClear(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	_ = job.Reject("d1", testNow) // clears offer
	if err := job.OfferToDriver("d2", testNow); err != nil {
		t.Fatalf("unexpected error after reject: %v", err)
	}
	if job.AttemptCount != 2 {
		t.Errorf("AttemptCount = %d, want 2", job.AttemptCount)
	}
}

func TestOfferToDriver_FromTerminalState(t *testing.T) {
	job := newJob(t)
	_ = job.MarkFailed(testNow)
	err := job.OfferToDriver("d1", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── Accept ──────────────────────────────────────────────────────────────────

func TestAccept_Valid(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	if err := job.Accept("d1", testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusAssigned {
		t.Errorf("Status = %q, want assigned", job.Status)
	}
	if job.AssignedDriverID != "d1" {
		t.Errorf("AssignedDriverID = %q, want d1", job.AssignedDriverID)
	}
	if job.CurrentDriverID != "" {
		t.Errorf("CurrentDriverID should be empty after accept, got %q", job.CurrentDriverID)
	}
}

func TestAccept_WrongDriver(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	err := job.Accept("d2", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestAccept_ExpiredOffer(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	expired := testNow.Add(31 * time.Second)
	err := job.Accept("d1", expired)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed for expired offer, got %v", err)
	}
}

func TestAccept_NotSearching(t *testing.T) {
	job := newJob(t)
	err := job.Accept("d1", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── Reject ──────────────────────────────────────────────────────────────────

func TestReject_Valid(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	if err := job.Reject("d1", testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusSearching {
		t.Errorf("Status = %q, want searching (stays searching after reject)", job.Status)
	}
	if job.CurrentDriverID != "" {
		t.Errorf("CurrentDriverID should be empty after reject")
	}
	if job.HasBeenOffered("d1") == false {
		t.Error("d1 should still be in OfferedDriverIDs after reject")
	}
}

func TestReject_WrongDriver(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	err := job.Reject("d2", testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── TimeoutOffer ─────────────────────────────────────────────────────────────

func TestTimeoutOffer_Valid(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	after := testNow.Add(31 * time.Second)
	if err := job.TimeoutOffer(after); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.CurrentDriverID != "" {
		t.Error("CurrentDriverID should be empty after timeout")
	}
}

func TestTimeoutOffer_NotExpired(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	err := job.TimeoutOffer(testNow.Add(5 * time.Second))
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed for non-expired offer, got %v", err)
	}
}

// ─── MarkFailed / Cancel ─────────────────────────────────────────────────────

func TestMarkFailed(t *testing.T) {
	job := newJob(t)
	_ = job.MarkFailed(testNow)
	if job.Status != entity.JobStatusFailed {
		t.Errorf("Status = %q, want failed", job.Status)
	}
}

func TestMarkFailed_FromAssigned(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	_ = job.Accept("d1", testNow)
	err := job.MarkFailed(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

func TestCancel_FromPending(t *testing.T) {
	job := newJob(t)
	if err := job.Cancel(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != entity.JobStatusCancelled {
		t.Errorf("Status = %q, want cancelled", job.Status)
	}
}

func TestCancel_FromAssigned(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	_ = job.Accept("d1", testNow)
	err := job.Cancel(testNow)
	if !domainerrors.IsCode(err, domainerrors.CodePreconditionFailed) {
		t.Errorf("expected PreconditionFailed, got %v", err)
	}
}

// ─── HasBeenOffered / IsOfferExpired ─────────────────────────────────────────

func TestHasBeenOffered(t *testing.T) {
	job := newJob(t)
	if job.HasBeenOffered("d1") {
		t.Error("should not be offered before any offers")
	}
	_ = job.OfferToDriver("d1", testNow)
	_ = job.Reject("d1", testNow)
	if !job.HasBeenOffered("d1") {
		t.Error("d1 should be in offered list after offer+reject")
	}
}

func TestIsOfferExpired_NoOffer(t *testing.T) {
	job := newJob(t)
	if job.IsOfferExpired(testNow) {
		t.Error("should not be expired when no offer is active")
	}
}

// ─── OfferedDriverIDsCSV ─────────────────────────────────────────────────────

func TestOfferedDriverIDsCSV(t *testing.T) {
	job := newJob(t)
	_ = job.OfferToDriver("d1", testNow)
	_ = job.Reject("d1", testNow)
	_ = job.OfferToDriver("d2", testNow)
	_ = job.Reject("d2", testNow)

	csv := job.OfferedDriverIDsCSV()
	if csv != "d1,d2" {
		t.Errorf("CSV = %q, want 'd1,d2'", csv)
	}
}

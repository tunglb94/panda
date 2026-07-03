package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
)

var testNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// ─── NewDriverProfile ─────────────────────────────────────────────────────────

func TestNewDriverProfile_Valid(t *testing.T) {
	d, err := entity.NewDriverProfile("d1", "u1", "LIC-001", entity.VehicleTypeCar, "Toyota", "Camry", "White", "ABC-123", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.DriverID != "d1" {
		t.Errorf("DriverID want d1 got %s", d.DriverID)
	}
	if d.OnlineStatus != entity.OnlineStatusOffline {
		t.Errorf("new driver should be offline")
	}
	if d.VerificationStatus != entity.VerificationStatusPending {
		t.Errorf("new driver should be pending")
	}
	if !d.CreatedAt.Equal(testNow) || !d.UpdatedAt.Equal(testNow) {
		t.Errorf("timestamps not set correctly")
	}
}

func TestNewDriverProfile_EmptyDriverID(t *testing.T) {
	_, err := entity.NewDriverProfile("", "u1", "LIC-001", entity.VehicleTypeCar, "", "", "", "ABC-123", testNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewDriverProfile_EmptyUserID(t *testing.T) {
	_, err := entity.NewDriverProfile("d1", "", "LIC-001", entity.VehicleTypeCar, "", "", "", "ABC-123", testNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewDriverProfile_EmptyLicense(t *testing.T) {
	_, err := entity.NewDriverProfile("d1", "u1", "   ", entity.VehicleTypeCar, "", "", "", "ABC-123", testNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewDriverProfile_InvalidVehicleType(t *testing.T) {
	_, err := entity.NewDriverProfile("d1", "u1", "LIC-001", entity.VehicleType("bicycle"), "", "", "", "ABC-123", testNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewDriverProfile_EmptyPlate(t *testing.T) {
	_, err := entity.NewDriverProfile("d1", "u1", "LIC-001", entity.VehicleTypeCar, "", "", "", "  ", testNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewDriverProfile_OptionalFieldsEmpty(t *testing.T) {
	// brand, model, color are optional
	d, err := entity.NewDriverProfile("d1", "u1", "LIC-001", entity.VehicleTypeMotorcycle, "", "", "", "XYZ-999", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.VehicleBrand != "" || d.VehicleModel != "" || d.VehicleColor != "" {
		t.Errorf("optional fields should be empty")
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestUpdate_Valid(t *testing.T) {
	d := newPendingDriver(t)
	later := testNow.Add(time.Hour)
	err := d.Update("LIC-NEW", entity.VehicleTypeVan, "Ford", "Transit", "Blue", "NEW-999", later)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.LicenseNumber != "LIC-NEW" {
		t.Errorf("license not updated")
	}
	if d.VehicleType != entity.VehicleTypeVan {
		t.Errorf("vehicle type not updated")
	}
	if !d.UpdatedAt.Equal(later) {
		t.Errorf("UpdatedAt not set to later")
	}
}

func TestUpdate_EmptyLicense(t *testing.T) {
	d := newPendingDriver(t)
	err := d.Update("", entity.VehicleTypeCar, "", "", "", "ABC", testNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestUpdate_EmptyPlate(t *testing.T) {
	d := newPendingDriver(t)
	err := d.Update("LIC", entity.VehicleTypeCar, "", "", "", "", testNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestUpdate_InvalidVehicleType(t *testing.T) {
	d := newPendingDriver(t)
	err := d.Update("LIC", entity.VehicleType("helicopter"), "", "", "", "PLATE", testNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

// ─── Verify ───────────────────────────────────────────────────────────────────

func TestVerify_FromPending(t *testing.T) {
	d := newPendingDriver(t)
	if err := d.Verify(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.VerificationStatus != entity.VerificationStatusVerified {
		t.Errorf("expected verified")
	}
}

func TestVerify_FromVerified(t *testing.T) {
	d := newVerifiedDriver(t)
	err := d.Verify(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

func TestVerify_FromRejected(t *testing.T) {
	d := newPendingDriver(t)
	_ = d.Reject(testNow)
	err := d.Verify(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

// ─── Reject ───────────────────────────────────────────────────────────────────

func TestReject_FromPending(t *testing.T) {
	d := newPendingDriver(t)
	if err := d.Reject(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.VerificationStatus != entity.VerificationStatusRejected {
		t.Errorf("expected rejected")
	}
}

func TestReject_FromVerified(t *testing.T) {
	d := newVerifiedDriver(t)
	err := d.Reject(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

func TestReject_FromSuspended(t *testing.T) {
	d := newSuspendedDriver(t)
	err := d.Reject(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

// ─── Suspend ──────────────────────────────────────────────────────────────────

func TestSuspend_FromVerified(t *testing.T) {
	d := newVerifiedDriver(t)
	_ = d.GoOnline(testNow) // go online first to verify Suspend forces offline
	if err := d.Suspend(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.VerificationStatus != entity.VerificationStatusSuspended {
		t.Errorf("expected suspended")
	}
	if d.OnlineStatus != entity.OnlineStatusOffline {
		t.Errorf("Suspend should force offline")
	}
}

func TestSuspend_FromPending(t *testing.T) {
	d := newPendingDriver(t)
	err := d.Suspend(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

func TestSuspend_FromSuspended(t *testing.T) {
	d := newSuspendedDriver(t)
	err := d.Suspend(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

// ─── Reinstate ────────────────────────────────────────────────────────────────

func TestReinstate_FromSuspended(t *testing.T) {
	d := newSuspendedDriver(t)
	if err := d.Reinstate(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.VerificationStatus != entity.VerificationStatusVerified {
		t.Errorf("expected verified after reinstate")
	}
}

func TestReinstate_FromVerified(t *testing.T) {
	d := newVerifiedDriver(t)
	err := d.Reinstate(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

func TestReinstate_FromPending(t *testing.T) {
	d := newPendingDriver(t)
	err := d.Reinstate(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

// ─── GoOnline / GoOffline ─────────────────────────────────────────────────────

func TestGoOnline_WhenVerified(t *testing.T) {
	d := newVerifiedDriver(t)
	if err := d.GoOnline(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.OnlineStatus != entity.OnlineStatusOnline {
		t.Errorf("expected online")
	}
}

func TestGoOnline_WhenPending(t *testing.T) {
	d := newPendingDriver(t)
	err := d.GoOnline(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

func TestGoOnline_WhenSuspended(t *testing.T) {
	d := newSuspendedDriver(t)
	err := d.GoOnline(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

func TestGoOnline_AlreadyOnline(t *testing.T) {
	d := newVerifiedDriver(t)
	_ = d.GoOnline(testNow)
	err := d.GoOnline(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

func TestGoOffline_WhenOnline(t *testing.T) {
	d := newVerifiedDriver(t)
	_ = d.GoOnline(testNow)
	if err := d.GoOffline(testNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.OnlineStatus != entity.OnlineStatusOffline {
		t.Errorf("expected offline")
	}
}

func TestGoOffline_AlreadyOffline(t *testing.T) {
	d := newVerifiedDriver(t)
	err := d.GoOffline(testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

// ─── ReconstituteDriverProfile ────────────────────────────────────────────────

func TestReconstituteDriverProfile(t *testing.T) {
	d := entity.ReconstituteDriverProfile(
		"d1", "u1", "LIC-001", entity.VehicleTypeCar,
		"Toyota", "Camry", "White", "ABC-123",
		entity.OnlineStatusOnline, entity.VerificationStatusVerified,
		testNow, testNow,
	)
	if d.DriverID != "d1" || d.OnlineStatus != entity.OnlineStatusOnline {
		t.Errorf("reconstitution failed: %+v", d)
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newPendingDriver(t *testing.T) *entity.DriverProfile {
	t.Helper()
	d, err := entity.NewDriverProfile("d1", "u1", "LIC-001", entity.VehicleTypeCar, "Toyota", "Camry", "White", "ABC-123", testNow)
	if err != nil {
		t.Fatalf("newPendingDriver: %v", err)
	}
	return d
}

func newVerifiedDriver(t *testing.T) *entity.DriverProfile {
	t.Helper()
	d := newPendingDriver(t)
	if err := d.Verify(testNow); err != nil {
		t.Fatalf("newVerifiedDriver: %v", err)
	}
	return d
}

func newSuspendedDriver(t *testing.T) *entity.DriverProfile {
	t.Helper()
	d := newVerifiedDriver(t)
	if err := d.Suspend(testNow); err != nil {
		t.Fatalf("newSuspendedDriver: %v", err)
	}
	return d
}

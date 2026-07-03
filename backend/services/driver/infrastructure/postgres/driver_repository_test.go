package postgres_test

import (
	"context"
	"testing"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/infrastructure/postgres"
	"github.com/fairride/shared/errors"
)

func TestSave_And_FindByID(t *testing.T) {
	setupTest(t)
	repo := postgres.NewDriverRepository(testPool)

	d, err := entity.NewDriverProfile("d1", "u1", "LIC-001", entity.VehicleTypeCar, "Toyota", "Camry", "White", "ABC-123", testNow)
	if err != nil {
		t.Fatalf("NewDriverProfile: %v", err)
	}

	if err := repo.Save(context.Background(), d); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(context.Background(), "d1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	assertDriverEqual(t, d, got)
}

func TestFindByUserID(t *testing.T) {
	setupTest(t)
	repo := postgres.NewDriverRepository(testPool)

	d, err := entity.NewDriverProfile("d2", "u2", "LIC-002", entity.VehicleTypeMotorcycle, "Honda", "CB500", "Red", "MOT-456", testNow)
	if err != nil {
		t.Fatalf("NewDriverProfile: %v", err)
	}
	if err := repo.Save(context.Background(), d); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByUserID(context.Background(), "u2")
	if err != nil {
		t.Fatalf("FindByUserID: %v", err)
	}
	if got.DriverID != "d2" {
		t.Errorf("want driver d2 got %s", got.DriverID)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	setupTest(t)
	repo := postgres.NewDriverRepository(testPool)

	_, err := repo.FindByID(context.Background(), "nonexistent")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound got %v", err)
	}
}

func TestFindByUserID_NotFound(t *testing.T) {
	setupTest(t)
	repo := postgres.NewDriverRepository(testPool)

	_, err := repo.FindByUserID(context.Background(), "no-user")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound got %v", err)
	}
}

func TestSave_Upsert_UpdatesFields(t *testing.T) {
	setupTest(t)
	repo := postgres.NewDriverRepository(testPool)

	d, _ := entity.NewDriverProfile("d3", "u3", "LIC-003", entity.VehicleTypeCar, "", "", "", "OLD-000", testNow)
	_ = repo.Save(context.Background(), d)

	later := testNow.Add(3600e9) // +1h
	_ = d.Update("LIC-NEW", entity.VehicleTypeVan, "Ford", "Transit", "Blue", "NEW-999", later)
	if err := repo.Save(context.Background(), d); err != nil {
		t.Fatalf("upsert Save: %v", err)
	}

	got, err := repo.FindByID(context.Background(), "d3")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.LicenseNumber != "LIC-NEW" {
		t.Errorf("license want LIC-NEW got %s", got.LicenseNumber)
	}
	if got.VehicleType != entity.VehicleTypeVan {
		t.Errorf("vehicle type not updated")
	}
	// user_id must be unchanged
	if got.UserID != "u3" {
		t.Errorf("user_id must not change on upsert, got %s", got.UserID)
	}
}

func TestSave_StatusFields_Persisted(t *testing.T) {
	setupTest(t)
	repo := postgres.NewDriverRepository(testPool)

	d, _ := entity.NewDriverProfile("d4", "u4", "LIC-004", entity.VehicleTypeCar, "", "", "", "PLT-004", testNow)
	_ = d.Verify(testNow)
	_ = d.GoOnline(testNow)
	_ = repo.Save(context.Background(), d)

	got, err := repo.FindByID(context.Background(), "d4")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.VerificationStatus != entity.VerificationStatusVerified {
		t.Errorf("verification_status not persisted")
	}
	if got.OnlineStatus != entity.OnlineStatusOnline {
		t.Errorf("online_status not persisted")
	}
}

func TestSave_UniqueUserID_Conflict(t *testing.T) {
	setupTest(t)
	repo := postgres.NewDriverRepository(testPool)

	d1, _ := entity.NewDriverProfile("dA", "userX", "LIC-A", entity.VehicleTypeCar, "", "", "", "PLT-A", testNow)
	_ = repo.Save(context.Background(), d1)

	// different driver_id but same user_id — should violate UNIQUE constraint on user_id
	d2, _ := entity.NewDriverProfile("dB", "userX", "LIC-B", entity.VehicleTypeCar, "", "", "", "PLT-B", testNow)
	err := repo.Save(context.Background(), d2)
	if !errors.IsCode(err, errors.CodeAlreadyExists) {
		t.Errorf("want CodeAlreadyExists got %v", err)
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func assertDriverEqual(t *testing.T, want, got *entity.DriverProfile) {
	t.Helper()
	if got.DriverID != want.DriverID {
		t.Errorf("DriverID want %s got %s", want.DriverID, got.DriverID)
	}
	if got.UserID != want.UserID {
		t.Errorf("UserID want %s got %s", want.UserID, got.UserID)
	}
	if got.LicenseNumber != want.LicenseNumber {
		t.Errorf("LicenseNumber want %s got %s", want.LicenseNumber, got.LicenseNumber)
	}
	if got.VehicleType != want.VehicleType {
		t.Errorf("VehicleType want %s got %s", want.VehicleType, got.VehicleType)
	}
	if got.VehicleBrand != want.VehicleBrand {
		t.Errorf("VehicleBrand want %s got %s", want.VehicleBrand, got.VehicleBrand)
	}
	if got.OnlineStatus != want.OnlineStatus {
		t.Errorf("OnlineStatus want %s got %s", want.OnlineStatus, got.OnlineStatus)
	}
	if got.VerificationStatus != want.VerificationStatus {
		t.Errorf("VerificationStatus want %s got %s", want.VerificationStatus, got.VerificationStatus)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("CreatedAt want %v got %v", want.CreatedAt, got.CreatedAt)
	}
}

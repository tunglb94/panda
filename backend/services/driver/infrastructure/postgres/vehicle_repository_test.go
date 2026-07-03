package postgres_test

import (
	"context"
	"testing"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/infrastructure/postgres"
	"github.com/fairride/shared/errors"
)

func newVehicleRepo() *postgres.VehicleRepository {
	return postgres.NewVehicleRepository(testPool)
}

func seedVehicleRow(t *testing.T, repo *postgres.VehicleRepository, vehicleID, driverID string) *entity.Vehicle {
	t.Helper()
	v := entity.ReconstituteVehicle(vehicleID, driverID, entity.VehicleTypeCar, "Toyota", "Camry", "White", "ABC-"+vehicleID, 2022, testNow, testNow)
	if err := repo.Save(context.Background(), v); err != nil {
		t.Fatalf("seedVehicleRow: %v", err)
	}
	return v
}

// ─── Save + FindByID ─────────────────────────────────────────────────────────

func TestVehicleSave_And_FindByID(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()

	v := entity.ReconstituteVehicle("v1", "d1", entity.VehicleTypeCar, "Toyota", "Camry", "White", "ABC-123", 2022, testNow, testNow)
	if err := repo.Save(context.Background(), v); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(context.Background(), "v1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	if got.VehicleID != "v1" {
		t.Errorf("VehicleID want v1 got %s", got.VehicleID)
	}
	if got.DriverID != "d1" {
		t.Errorf("DriverID want d1 got %s", got.DriverID)
	}
	if got.Type != entity.VehicleTypeCar {
		t.Errorf("Type want car got %s", got.Type)
	}
	if got.Brand != "Toyota" {
		t.Errorf("Brand want Toyota got %s", got.Brand)
	}
	if got.Year != 2022 {
		t.Errorf("Year want 2022 got %d", got.Year)
	}
	if !got.CreatedAt.Equal(testNow.UTC()) {
		t.Errorf("CreatedAt not persisted correctly")
	}
}

func TestVehicleFindByID_NotFound(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()

	_, err := repo.FindByID(context.Background(), "nonexistent")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound got %v", err)
	}
}

// ─── FindByDriverID ───────────────────────────────────────────────────────────

func TestVehicleFindByDriverID_ReturnsAll(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()

	seedVehicleRow(t, repo, "v1", "d1")
	seedVehicleRow(t, repo, "v2", "d1")
	seedVehicleRow(t, repo, "v3", "d2") // different driver

	list, err := repo.FindByDriverID(context.Background(), "d1")
	if err != nil {
		t.Fatalf("FindByDriverID: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("want 2 vehicles for d1 got %d", len(list))
	}
}

func TestVehicleFindByDriverID_EmptyForUnknownDriver(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()

	list, err := repo.FindByDriverID(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("want 0 vehicles got %d", len(list))
	}
}

// ─── Save upsert ─────────────────────────────────────────────────────────────

func TestVehicleSave_Upsert_UpdatesFields(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()

	v := entity.ReconstituteVehicle("v1", "d1", entity.VehicleTypeCar, "Old", "Old", "Old", "OLD-000", 2019, testNow, testNow)
	_ = repo.Save(context.Background(), v)

	later := testNow.Add(3600e9)
	_ = v.Update(entity.VehicleTypeVan, "Ford", "Transit", "Blue", "NEW-999", 2024, later)
	if err := repo.Save(context.Background(), v); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := repo.FindByID(context.Background(), "v1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Type != entity.VehicleTypeVan {
		t.Errorf("Type not updated")
	}
	if got.Brand != "Ford" {
		t.Errorf("Brand not updated")
	}
	if got.PlateNumber != "NEW-999" {
		t.Errorf("PlateNumber not updated")
	}
	if got.Year != 2024 {
		t.Errorf("Year not updated")
	}
	// driver_id must not change
	if got.DriverID != "d1" {
		t.Errorf("DriverID must not change on upsert")
	}
	// created_at must not change
	if !got.CreatedAt.Equal(testNow.UTC()) {
		t.Errorf("CreatedAt must not change on upsert")
	}
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestVehicleDelete_OK(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()
	seedVehicleRow(t, repo, "v1", "d1")

	if err := repo.Delete(context.Background(), "v1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(context.Background(), "v1")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound after delete, got %v", err)
	}
}

func TestVehicleDelete_NotFound(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()

	err := repo.Delete(context.Background(), "ghost")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound got %v", err)
	}
}

func TestVehicleDelete_OnlyDeletesTarget(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()
	seedVehicleRow(t, repo, "v1", "d1")
	seedVehicleRow(t, repo, "v2", "d1")

	if err := repo.Delete(context.Background(), "v1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	list, _ := repo.FindByDriverID(context.Background(), "d1")
	if len(list) != 1 || list[0].VehicleID != "v2" {
		t.Errorf("only v1 should be deleted, got %v", list)
	}
}

// ─── YearZero stored and retrieved correctly ──────────────────────────────────

func TestVehicleSave_YearZero(t *testing.T) {
	setupTest(t)
	repo := newVehicleRepo()

	v := entity.ReconstituteVehicle("v1", "d1", entity.VehicleTypeCar, "", "", "", "PLATE", 0, testNow, testNow)
	_ = repo.Save(context.Background(), v)

	got, _ := repo.FindByID(context.Background(), "v1")
	if got.Year != 0 {
		t.Errorf("Year want 0 got %d", got.Year)
	}
}

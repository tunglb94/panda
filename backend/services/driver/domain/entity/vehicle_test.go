package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
)

var vTestNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func validVehicle(t *testing.T) *entity.Vehicle {
	t.Helper()
	v, err := entity.NewVehicle("v1", "d1", entity.VehicleTypeCar, "Toyota", "Camry", "White", "ABC-123", 2022, vTestNow)
	if err != nil {
		t.Fatalf("validVehicle: %v", err)
	}
	return v
}

// ─── NewVehicle ───────────────────────────────────────────────────────────────

func TestNewVehicle_Valid(t *testing.T) {
	v := validVehicle(t)
	if v.VehicleID != "v1" {
		t.Errorf("VehicleID want v1 got %s", v.VehicleID)
	}
	if v.DriverID != "d1" {
		t.Errorf("DriverID want d1 got %s", v.DriverID)
	}
	if v.Type != entity.VehicleTypeCar {
		t.Errorf("Type want car got %s", v.Type)
	}
	if v.Year != 2022 {
		t.Errorf("Year want 2022 got %d", v.Year)
	}
	if !v.CreatedAt.Equal(vTestNow) || !v.UpdatedAt.Equal(vTestNow) {
		t.Errorf("timestamps not set to now")
	}
}

func TestNewVehicle_EmptyVehicleID(t *testing.T) {
	_, err := entity.NewVehicle("", "d1", entity.VehicleTypeCar, "", "", "", "PLATE", 0, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewVehicle_EmptyDriverID(t *testing.T) {
	_, err := entity.NewVehicle("v1", "", entity.VehicleTypeCar, "", "", "", "PLATE", 0, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewVehicle_InvalidVehicleType(t *testing.T) {
	_, err := entity.NewVehicle("v1", "d1", entity.VehicleType("tuk-tuk"), "", "", "", "PLATE", 0, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewVehicle_EmptyPlate(t *testing.T) {
	_, err := entity.NewVehicle("v1", "d1", entity.VehicleTypeCar, "", "", "", "  ", 0, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewVehicle_YearZero_NotProvided(t *testing.T) {
	v, err := entity.NewVehicle("v1", "d1", entity.VehicleTypeMotorcycle, "", "", "", "XYZ", 0, vTestNow)
	if err != nil {
		t.Fatalf("year=0 should be accepted: %v", err)
	}
	if v.Year != 0 {
		t.Errorf("Year want 0 got %d", v.Year)
	}
}

func TestNewVehicle_YearTooOld(t *testing.T) {
	_, err := entity.NewVehicle("v1", "d1", entity.VehicleTypeCar, "", "", "", "PLATE", 1899, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestNewVehicle_YearInFuture(t *testing.T) {
	futureYear := vTestNow.Year() + 2
	_, err := entity.NewVehicle("v1", "d1", entity.VehicleTypeCar, "", "", "", "PLATE", futureYear, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for year %d got %v", futureYear, err)
	}
}

func TestNewVehicle_YearCurrentPlusOne_Allowed(t *testing.T) {
	// next-model-year vehicles are allowed
	nextYear := vTestNow.Year() + 1
	_, err := entity.NewVehicle("v1", "d1", entity.VehicleTypeCar, "", "", "", "PLATE", nextYear, vTestNow)
	if err != nil {
		t.Errorf("year+1 should be allowed: %v", err)
	}
}

func TestNewVehicle_AllVehicleTypes(t *testing.T) {
	types := []entity.VehicleType{entity.VehicleTypeCar, entity.VehicleTypeMotorcycle, entity.VehicleTypeVan}
	for _, vt := range types {
		_, err := entity.NewVehicle("v1", "d1", vt, "", "", "", "P", 0, vTestNow)
		if err != nil {
			t.Errorf("type %s should be valid: %v", vt, err)
		}
	}
}

func TestNewVehicle_OptionalFieldsEmpty(t *testing.T) {
	v, err := entity.NewVehicle("v1", "d1", entity.VehicleTypeVan, "", "", "", "PLATE", 0, vTestNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Brand != "" || v.Model != "" || v.Color != "" {
		t.Errorf("optional fields should be empty")
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestVehicleUpdate_Valid(t *testing.T) {
	v := validVehicle(t)
	later := vTestNow.Add(time.Hour)
	err := v.Update(entity.VehicleTypeVan, "Ford", "Transit", "Blue", "NEW-000", 2024, later)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Type != entity.VehicleTypeVan {
		t.Errorf("Type not updated")
	}
	if v.Brand != "Ford" {
		t.Errorf("Brand not updated")
	}
	if v.PlateNumber != "NEW-000" {
		t.Errorf("PlateNumber not updated")
	}
	if v.Year != 2024 {
		t.Errorf("Year not updated")
	}
	if !v.UpdatedAt.Equal(later) {
		t.Errorf("UpdatedAt not set to later")
	}
}

func TestVehicleUpdate_EmptyPlate(t *testing.T) {
	v := validVehicle(t)
	err := v.Update(entity.VehicleTypeCar, "", "", "", "", 0, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestVehicleUpdate_InvalidVehicleType(t *testing.T) {
	v := validVehicle(t)
	err := v.Update(entity.VehicleType("boat"), "", "", "", "PLATE", 0, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestVehicleUpdate_BadYear(t *testing.T) {
	v := validVehicle(t)
	err := v.Update(entity.VehicleTypeCar, "", "", "", "PLATE", 1800, vTestNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestVehicleUpdate_DoesNotChangeVehicleIDOrDriverIDOrCreatedAt(t *testing.T) {
	v := validVehicle(t)
	origCreated := v.CreatedAt
	_ = v.Update(entity.VehicleTypeCar, "B", "M", "C", "P2", 0, vTestNow.Add(time.Hour))
	if v.VehicleID != "v1" {
		t.Errorf("VehicleID must not change")
	}
	if v.DriverID != "d1" {
		t.Errorf("DriverID must not change")
	}
	if !v.CreatedAt.Equal(origCreated) {
		t.Errorf("CreatedAt must not change")
	}
}

// ─── ReconstituteVehicle ─────────────────────────────────────────────────────

func TestReconstituteVehicle(t *testing.T) {
	v := entity.ReconstituteVehicle("v99", "d99", entity.VehicleTypeCar, "BMW", "3-Series", "Black", "BMV-999", 2020, vTestNow, vTestNow)
	if v.VehicleID != "v99" || v.Year != 2020 || v.Brand != "BMW" {
		t.Errorf("reconstitution failed: %+v", v)
	}
}

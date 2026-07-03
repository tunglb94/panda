package app_test

import (
	"context"
	"testing"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
)

// ─── in-memory vehicle stub ──────────────────────────────────────────────────

type stubVehicleRepo struct {
	byID       map[string]*entity.Vehicle
	byDriverID map[string][]*entity.Vehicle
}

func newStubVehicleRepo() *stubVehicleRepo {
	return &stubVehicleRepo{
		byID:       make(map[string]*entity.Vehicle),
		byDriverID: make(map[string][]*entity.Vehicle),
	}
}

func (r *stubVehicleRepo) FindByID(_ context.Context, id string) (*entity.Vehicle, error) {
	v, ok := r.byID[id]
	if !ok {
		return nil, errors.NotFound("vehicle not found: " + id)
	}
	return v, nil
}

func (r *stubVehicleRepo) FindByDriverID(_ context.Context, driverID string) ([]*entity.Vehicle, error) {
	return r.byDriverID[driverID], nil
}

func (r *stubVehicleRepo) Save(_ context.Context, v *entity.Vehicle) error {
	r.byID[v.VehicleID] = v
	// rebuild the slice for this driver
	existing := r.byDriverID[v.DriverID]
	updated := make([]*entity.Vehicle, 0, len(existing))
	for _, e := range existing {
		if e.VehicleID != v.VehicleID {
			updated = append(updated, e)
		}
	}
	r.byDriverID[v.DriverID] = append(updated, v)
	return nil
}

func (r *stubVehicleRepo) Delete(_ context.Context, vehicleID string) error {
	v, ok := r.byID[vehicleID]
	if !ok {
		return errors.NotFound("vehicle not found: " + vehicleID)
	}
	delete(r.byID, vehicleID)
	existing := r.byDriverID[v.DriverID]
	updated := make([]*entity.Vehicle, 0, len(existing))
	for _, e := range existing {
		if e.VehicleID != vehicleID {
			updated = append(updated, e)
		}
	}
	r.byDriverID[v.DriverID] = updated
	return nil
}

// ─── CreateVehicleUseCase ────────────────────────────────────────────────────

func TestCreateVehicle_Valid(t *testing.T) {
	repo := newStubVehicleRepo()
	uc := app.NewCreateVehicleUseCase(repo)

	in := app.CreateVehicleInput{
		DriverID:    "d1",
		VehicleType: entity.VehicleTypeCar,
		Brand:       "Toyota",
		Model:       "Camry",
		Color:       "White",
		PlateNumber: "ABC-123",
		Year:        2022,
	}
	v, err := uc.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.DriverID != "d1" {
		t.Errorf("DriverID want d1 got %s", v.DriverID)
	}
	if v.VehicleID == "" {
		t.Errorf("VehicleID must be generated")
	}
	if v.PlateNumber != "ABC-123" {
		t.Errorf("PlateNumber not set")
	}
}

func TestCreateVehicle_EmptyDriverID(t *testing.T) {
	repo := newStubVehicleRepo()
	uc := app.NewCreateVehicleUseCase(repo)
	_, err := uc.Execute(context.Background(), app.CreateVehicleInput{})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestCreateVehicle_InvalidType(t *testing.T) {
	repo := newStubVehicleRepo()
	uc := app.NewCreateVehicleUseCase(repo)
	_, err := uc.Execute(context.Background(), app.CreateVehicleInput{
		DriverID:    "d1",
		VehicleType: entity.VehicleType("rocket"),
		PlateNumber: "PLATE",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestCreateVehicle_EmptyPlate(t *testing.T) {
	repo := newStubVehicleRepo()
	uc := app.NewCreateVehicleUseCase(repo)
	_, err := uc.Execute(context.Background(), app.CreateVehicleInput{
		DriverID:    "d1",
		VehicleType: entity.VehicleTypeCar,
		PlateNumber: "",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestCreateVehicle_IDIsUnique(t *testing.T) {
	repo := newStubVehicleRepo()
	uc := app.NewCreateVehicleUseCase(repo)
	in := app.CreateVehicleInput{DriverID: "d1", VehicleType: entity.VehicleTypeCar, PlateNumber: "P"}

	v1, _ := uc.Execute(context.Background(), in)
	v2, _ := uc.Execute(context.Background(), in)
	if v1.VehicleID == v2.VehicleID {
		t.Errorf("generated IDs must be unique")
	}
}

// ─── UpdateVehicleUseCase ────────────────────────────────────────────────────

func seedVehicle(t *testing.T, repo *stubVehicleRepo) *entity.Vehicle {
	t.Helper()
	v := entity.ReconstituteVehicle("v1", "d1", entity.VehicleTypeCar, "Toyota", "Camry", "White", "OLD-000", 2020, testNow, testNow)
	_ = repo.Save(context.Background(), v)
	return v
}

func TestUpdateVehicle_Valid(t *testing.T) {
	repo := newStubVehicleRepo()
	seedVehicle(t, repo)
	uc := app.NewUpdateVehicleUseCase(repo)

	in := app.UpdateVehicleInput{
		VehicleID:   "v1",
		VehicleType: entity.VehicleTypeVan,
		Brand:       "Ford",
		Model:       "Transit",
		Color:       "Blue",
		PlateNumber: "NEW-999",
		Year:        2024,
	}
	v, err := uc.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Type != entity.VehicleTypeVan {
		t.Errorf("Type not updated")
	}
	if v.PlateNumber != "NEW-999" {
		t.Errorf("PlateNumber not updated")
	}
}

func TestUpdateVehicle_EmptyVehicleID(t *testing.T) {
	repo := newStubVehicleRepo()
	uc := app.NewUpdateVehicleUseCase(repo)
	_, err := uc.Execute(context.Background(), app.UpdateVehicleInput{})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestUpdateVehicle_NotFound(t *testing.T) {
	repo := newStubVehicleRepo()
	uc := app.NewUpdateVehicleUseCase(repo)
	_, err := uc.Execute(context.Background(), app.UpdateVehicleInput{
		VehicleID:   "ghost",
		VehicleType: entity.VehicleTypeCar,
		PlateNumber: "PLATE",
	})
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound got %v", err)
	}
}

// ─── DeleteVehicleUseCase ────────────────────────────────────────────────────

func TestDeleteVehicle_Valid(t *testing.T) {
	repo := newStubVehicleRepo()
	seedVehicle(t, repo)
	uc := app.NewDeleteVehicleUseCase(repo)

	if err := uc.Execute(context.Background(), "v1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// confirm gone
	_, err := repo.FindByID(context.Background(), "v1")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("vehicle should be deleted, got %v", err)
	}
}

func TestDeleteVehicle_EmptyID(t *testing.T) {
	uc := app.NewDeleteVehicleUseCase(newStubVehicleRepo())
	err := uc.Execute(context.Background(), "")
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestDeleteVehicle_NotFound(t *testing.T) {
	uc := app.NewDeleteVehicleUseCase(newStubVehicleRepo())
	err := uc.Execute(context.Background(), "ghost")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound got %v", err)
	}
}

// ─── ListVehiclesUseCase ──────────────────────────────────────────────────────

func TestListVehicles_ReturnsAll(t *testing.T) {
	repo := newStubVehicleRepo()
	v1 := entity.ReconstituteVehicle("v1", "d1", entity.VehicleTypeCar, "", "", "", "P1", 0, testNow, testNow)
	v2 := entity.ReconstituteVehicle("v2", "d1", entity.VehicleTypeMotorcycle, "", "", "", "P2", 0, testNow, testNow)
	_ = repo.Save(context.Background(), v1)
	_ = repo.Save(context.Background(), v2)

	uc := app.NewListVehiclesUseCase(repo)
	list, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("want 2 vehicles got %d", len(list))
	}
}

func TestListVehicles_EmptyDriverID(t *testing.T) {
	uc := app.NewListVehiclesUseCase(newStubVehicleRepo())
	_, err := uc.Execute(context.Background(), "")
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestListVehicles_EmptyList(t *testing.T) {
	uc := app.NewListVehiclesUseCase(newStubVehicleRepo())
	list, err := uc.Execute(context.Background(), "d-no-vehicles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("want empty list got %d", len(list))
	}
}

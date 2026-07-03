package grpc_test

import (
	"context"
	"testing"

	drivergrpc "github.com/fairride/driver/grpc"
	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/grpc/driverpb"
	"github.com/fairride/shared/errors"
	"google.golang.org/grpc/codes"
)

// ─── in-memory vehicle stub ──────────────────────────────────────────────────

type stubVehicleRepoH struct {
	byID       map[string]*entity.Vehicle
	byDriverID map[string][]*entity.Vehicle
}

func newStubVehicleRepoH() *stubVehicleRepoH {
	return &stubVehicleRepoH{
		byID:       make(map[string]*entity.Vehicle),
		byDriverID: make(map[string][]*entity.Vehicle),
	}
}

func (r *stubVehicleRepoH) FindByID(_ context.Context, id string) (*entity.Vehicle, error) {
	v, ok := r.byID[id]
	if !ok {
		return nil, errors.NotFound("vehicle not found: " + id)
	}
	return v, nil
}

func (r *stubVehicleRepoH) FindByDriverID(_ context.Context, driverID string) ([]*entity.Vehicle, error) {
	return r.byDriverID[driverID], nil
}

func (r *stubVehicleRepoH) Save(_ context.Context, v *entity.Vehicle) error {
	r.byID[v.VehicleID] = v
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

func (r *stubVehicleRepoH) Delete(_ context.Context, vehicleID string) error {
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

// ─── builder ─────────────────────────────────────────────────────────────────

func newVehicleHandler(repo *stubVehicleRepoH) *drivergrpc.VehicleHandler {
	return drivergrpc.NewVehicleHandler(
		app.NewCreateVehicleUseCase(repo),
		app.NewUpdateVehicleUseCase(repo),
		app.NewDeleteVehicleUseCase(repo),
		app.NewListVehiclesUseCase(repo),
	)
}

func seedVehicleH(t *testing.T, repo *stubVehicleRepoH, vehicleID, driverID string) *entity.Vehicle {
	t.Helper()
	v := entity.ReconstituteVehicle(vehicleID, driverID, entity.VehicleTypeCar, "Toyota", "Camry", "White", "PLATE-"+vehicleID, 2022, testNow, testNow)
	_ = repo.Save(context.Background(), v)
	return v
}

// ─── CreateVehicle ────────────────────────────────────────────────────────────

func TestVehicleCreate_OK(t *testing.T) {
	repo := newStubVehicleRepoH()
	h := newVehicleHandler(repo)

	resp, err := h.CreateVehicle(context.Background(), &driverpb.CreateVehicleRequest{
		DriverId:    "d1",
		Type:        "car",
		Brand:       "Toyota",
		Model:       "Camry",
		Color:       "White",
		PlateNumber: "ABC-123",
		Year:        2022,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Vehicle.DriverId != "d1" {
		t.Errorf("DriverId want d1 got %s", resp.Vehicle.DriverId)
	}
	if resp.Vehicle.VehicleId == "" {
		t.Errorf("VehicleId must be generated")
	}
	if resp.Vehicle.Type != "car" {
		t.Errorf("Type want car got %s", resp.Vehicle.Type)
	}
}

func TestVehicleCreate_EmptyDriverID(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	_, err := h.CreateVehicle(context.Background(), &driverpb.CreateVehicleRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestVehicleCreate_InvalidType(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	_, err := h.CreateVehicle(context.Background(), &driverpb.CreateVehicleRequest{
		DriverId:    "d1",
		Type:        "spaceship",
		PlateNumber: "PLATE",
	})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestVehicleCreate_EmptyPlate(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	_, err := h.CreateVehicle(context.Background(), &driverpb.CreateVehicleRequest{
		DriverId: "d1",
		Type:     "car",
	})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── UpdateVehicle ────────────────────────────────────────────────────────────

func TestVehicleUpdate_OK(t *testing.T) {
	repo := newStubVehicleRepoH()
	seedVehicleH(t, repo, "v1", "d1")
	h := newVehicleHandler(repo)

	resp, err := h.UpdateVehicle(context.Background(), &driverpb.UpdateVehicleRequest{
		VehicleId:   "v1",
		Type:        "van",
		Brand:       "Ford",
		Model:       "Transit",
		Color:       "Blue",
		PlateNumber: "NEW-999",
		Year:        2024,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Vehicle.Type != "van" {
		t.Errorf("Type not updated in response")
	}
	if resp.Vehicle.PlateNumber != "NEW-999" {
		t.Errorf("PlateNumber not updated in response")
	}
}

func TestVehicleUpdate_EmptyVehicleID(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	_, err := h.UpdateVehicle(context.Background(), &driverpb.UpdateVehicleRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestVehicleUpdate_NotFound(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	_, err := h.UpdateVehicle(context.Background(), &driverpb.UpdateVehicleRequest{
		VehicleId:   "ghost",
		Type:        "car",
		PlateNumber: "PLATE",
	})
	assertGRPCCode(t, err, codes.NotFound)
}

// ─── DeleteVehicle ────────────────────────────────────────────────────────────

func TestVehicleDelete_OK(t *testing.T) {
	repo := newStubVehicleRepoH()
	seedVehicleH(t, repo, "v1", "d1")
	h := newVehicleHandler(repo)

	_, err := h.DeleteVehicle(context.Background(), &driverpb.DeleteVehicleRequest{VehicleId: "v1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// confirm gone from stub
	_, findErr := repo.FindByID(context.Background(), "v1")
	if !errors.IsCode(findErr, errors.CodeNotFound) {
		t.Errorf("vehicle should be deleted")
	}
}

func TestVehicleDelete_EmptyID(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	_, err := h.DeleteVehicle(context.Background(), &driverpb.DeleteVehicleRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestVehicleDelete_NotFound(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	_, err := h.DeleteVehicle(context.Background(), &driverpb.DeleteVehicleRequest{VehicleId: "ghost"})
	assertGRPCCode(t, err, codes.NotFound)
}

// ─── ListVehicles ─────────────────────────────────────────────────────────────

func TestVehicleList_OK(t *testing.T) {
	repo := newStubVehicleRepoH()
	seedVehicleH(t, repo, "v1", "d1")
	seedVehicleH(t, repo, "v2", "d1")
	seedVehicleH(t, repo, "v3", "d2") // different driver
	h := newVehicleHandler(repo)

	resp, err := h.ListVehicles(context.Background(), &driverpb.ListVehiclesRequest{DriverId: "d1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Vehicles) != 2 {
		t.Errorf("want 2 vehicles for d1 got %d", len(resp.Vehicles))
	}
}

func TestVehicleList_EmptyDriverID(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	_, err := h.ListVehicles(context.Background(), &driverpb.ListVehiclesRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestVehicleList_EmptyResultForUnknownDriver(t *testing.T) {
	h := newVehicleHandler(newStubVehicleRepoH())
	resp, err := h.ListVehicles(context.Background(), &driverpb.ListVehiclesRequest{DriverId: "nobody"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Vehicles) != 0 {
		t.Errorf("want 0 vehicles got %d", len(resp.Vehicles))
	}
}

// ─── proto field coverage ─────────────────────────────────────────────────────

func TestVehicleCreate_ProtoFields(t *testing.T) {
	repo := newStubVehicleRepoH()
	h := newVehicleHandler(repo)

	resp, _ := h.CreateVehicle(context.Background(), &driverpb.CreateVehicleRequest{
		DriverId:    "d1",
		Type:        "motorcycle",
		Brand:       "Honda",
		Model:       "CB500",
		Color:       "Red",
		PlateNumber: "MOT-001",
		Year:        2021,
	})
	p := resp.Vehicle
	if p.Brand != "Honda" || p.Model != "CB500" || p.Color != "Red" {
		t.Errorf("optional fields not in proto: %+v", p)
	}
	if p.Year != 2021 {
		t.Errorf("Year want 2021 got %d", p.Year)
	}
	if p.CreatedAt == nil || p.UpdatedAt == nil {
		t.Errorf("timestamps must not be nil")
	}
}

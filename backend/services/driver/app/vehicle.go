package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	"github.com/fairride/shared/errors"
)

// ─── CreateVehicle ────────────────────────────────────────────────────────────

// CreateVehicleInput carries the data required to register a new vehicle.
type CreateVehicleInput struct {
	DriverID    string
	VehicleType entity.VehicleType
	Brand       string
	Model       string
	Color       string
	PlateNumber string
	Year        int
}

// CreateVehicleUseCase registers a new vehicle for a driver.
type CreateVehicleUseCase struct {
	repo repository.VehicleRepository
}

func NewCreateVehicleUseCase(repo repository.VehicleRepository) *CreateVehicleUseCase {
	return &CreateVehicleUseCase{repo: repo}
}

func (uc *CreateVehicleUseCase) Execute(ctx context.Context, in CreateVehicleInput) (*entity.Vehicle, error) {
	if in.DriverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	id, err := generateVehicleID()
	if err != nil {
		return nil, errors.Internal("failed to generate vehicle id").WithMeta("error", err.Error())
	}
	v, err := entity.NewVehicle(id, in.DriverID, in.VehicleType, in.Brand, in.Model, in.Color, in.PlateNumber, in.Year, time.Now())
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

// ─── UpdateVehicle ────────────────────────────────────────────────────────────

// UpdateVehicleInput carries the fields the caller wants to change.
type UpdateVehicleInput struct {
	VehicleID   string
	VehicleType entity.VehicleType
	Brand       string
	Model       string
	Color       string
	PlateNumber string
	Year        int
}

// UpdateVehicleUseCase replaces mutable fields on an existing vehicle.
type UpdateVehicleUseCase struct {
	repo repository.VehicleRepository
}

func NewUpdateVehicleUseCase(repo repository.VehicleRepository) *UpdateVehicleUseCase {
	return &UpdateVehicleUseCase{repo: repo}
}

func (uc *UpdateVehicleUseCase) Execute(ctx context.Context, in UpdateVehicleInput) (*entity.Vehicle, error) {
	if in.VehicleID == "" {
		return nil, errors.InvalidArgument("vehicle id must not be empty")
	}
	v, err := uc.repo.FindByID(ctx, in.VehicleID)
	if err != nil {
		return nil, err
	}
	if err := v.Update(in.VehicleType, in.Brand, in.Model, in.Color, in.PlateNumber, in.Year, time.Now()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

// ─── DeleteVehicle ────────────────────────────────────────────────────────────

// DeleteVehicleUseCase permanently removes a vehicle.
type DeleteVehicleUseCase struct {
	repo repository.VehicleRepository
}

func NewDeleteVehicleUseCase(repo repository.VehicleRepository) *DeleteVehicleUseCase {
	return &DeleteVehicleUseCase{repo: repo}
}

func (uc *DeleteVehicleUseCase) Execute(ctx context.Context, vehicleID string) error {
	if vehicleID == "" {
		return errors.InvalidArgument("vehicle id must not be empty")
	}
	return uc.repo.Delete(ctx, vehicleID)
}

// ─── ListVehicles ─────────────────────────────────────────────────────────────

// ListVehiclesUseCase returns all vehicles for a driver.
type ListVehiclesUseCase struct {
	repo repository.VehicleRepository
}

func NewListVehiclesUseCase(repo repository.VehicleRepository) *ListVehiclesUseCase {
	return &ListVehiclesUseCase{repo: repo}
}

func (uc *ListVehiclesUseCase) Execute(ctx context.Context, driverID string) ([]*entity.Vehicle, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	return uc.repo.FindByDriverID(ctx, driverID)
}

// ─── private helpers ──────────────────────────────────────────────────────────

// generateVehicleID produces a cryptographically random 32-char hex vehicle ID.
func generateVehicleID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

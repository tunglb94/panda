package repository

import (
	"context"

	"github.com/fairride/driver/domain/entity"
)

// VehicleRepository defines persistence operations for Vehicle.
// All methods return *errors.DomainError on failure.
type VehicleRepository interface {
	// FindByID returns a vehicle by its primary key.
	FindByID(ctx context.Context, vehicleID string) (*entity.Vehicle, error)

	// FindByDriverID returns all vehicles belonging to a driver.
	FindByDriverID(ctx context.Context, driverID string) ([]*entity.Vehicle, error)

	// Save creates or updates a vehicle (upsert by vehicle_id).
	Save(ctx context.Context, v *entity.Vehicle) error

	// Delete permanently removes a vehicle by its primary key.
	// Returns CodeNotFound if no vehicle with that ID exists.
	Delete(ctx context.Context, vehicleID string) error
}

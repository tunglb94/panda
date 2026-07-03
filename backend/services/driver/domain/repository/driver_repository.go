package repository

import (
	"context"

	"github.com/fairride/driver/domain/entity"
)

// DriverRepository defines persistence operations for DriverProfile.
// All methods return *errors.DomainError on failure.
type DriverRepository interface {
	// FindByID returns a driver profile by its primary key.
	FindByID(ctx context.Context, driverID string) (*entity.DriverProfile, error)

	// FindByUserID returns the driver profile associated with the given user identity.
	FindByUserID(ctx context.Context, userID string) (*entity.DriverProfile, error)

	// Save creates or updates a driver profile (upsert by driver_id).
	Save(ctx context.Context, d *entity.DriverProfile) error
}

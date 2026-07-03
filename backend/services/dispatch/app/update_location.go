package app

import (
	"context"

	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// UpdateDriverLocationUseCase updates a driver's GPS coordinates in the dispatch
// location store. Drivers should call this periodically while online.
type UpdateDriverLocationUseCase struct {
	locationRepo repository.DriverLocationRepository
}

func NewUpdateDriverLocationUseCase(locationRepo repository.DriverLocationRepository) *UpdateDriverLocationUseCase {
	return &UpdateDriverLocationUseCase{locationRepo: locationRepo}
}

func (uc *UpdateDriverLocationUseCase) Execute(ctx context.Context, driverID string, lat, lon float64) error {
	if driverID == "" {
		return domainerrors.InvalidArgument("driver_id is required")
	}
	return uc.locationRepo.UpdateLocation(ctx, driverID, lat, lon)
}

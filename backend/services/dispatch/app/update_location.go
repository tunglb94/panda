package app

import (
	"context"

	"github.com/fairride/dispatch/domain/entity"
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

// serviceType is optional (empty string = not reported) — older clients
// that don't send one keep working exactly as before this catalog existed.
// rideEnabled/deliveryEnabled should already reflect the caller's intended
// default for "not reported" (see repository.DriverLocationRepository's
// doc comment) — this use case passes them through unchanged.
func (uc *UpdateDriverLocationUseCase) Execute(ctx context.Context, driverID string, lat, lon float64, serviceType entity.ServiceType, rideEnabled, deliveryEnabled bool) error {
	if driverID == "" {
		return domainerrors.InvalidArgument("driver_id is required")
	}
	return uc.locationRepo.UpdateLocation(ctx, driverID, lat, lon, serviceType, rideEnabled, deliveryEnabled)
}

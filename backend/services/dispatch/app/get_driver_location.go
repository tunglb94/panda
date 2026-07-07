package app

import (
	"context"

	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// GetDriverLocationUseCase returns a driver's last known GPS coordinates and
// whether they are still within the active TTL window.
type GetDriverLocationUseCase struct {
	locationRepo repository.DriverLocationRepository
}

func NewGetDriverLocationUseCase(locationRepo repository.DriverLocationRepository) *GetDriverLocationUseCase {
	return &GetDriverLocationUseCase{locationRepo: locationRepo}
}

type DriverLocationResult struct {
	Lat      float64
	Lon      float64
	IsActive bool
}

func (uc *GetDriverLocationUseCase) Execute(ctx context.Context, driverID string) (*DriverLocationResult, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	lat, lon, err := uc.locationRepo.GetLocation(ctx, driverID)
	if err != nil {
		code := domainerrors.GetCode(err)
		if code == domainerrors.CodeNotFound {
			return &DriverLocationResult{IsActive: false}, nil
		}
		return nil, err
	}
	active, err := uc.locationRepo.IsActive(ctx, driverID)
	if err != nil {
		return nil, err
	}
	return &DriverLocationResult{Lat: lat, Lon: lon, IsActive: active}, nil
}

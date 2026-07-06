package app

import (
	"context"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// DriverOfferResult holds the active trip offer directed at a specific driver.
type DriverOfferResult struct {
	Job            *entity.DispatchJob
	OfferExpiresAt time.Time
}

// GetDriverOfferUseCase returns the current pending offer for a driver.
// Returns (nil, CodeNotFound) when the driver has no active offer.
type GetDriverOfferUseCase struct {
	jobRepo repository.DispatchJobRepository
}

func NewGetDriverOfferUseCase(jobRepo repository.DispatchJobRepository) *GetDriverOfferUseCase {
	return &GetDriverOfferUseCase{jobRepo: jobRepo}
}

func (uc *GetDriverOfferUseCase) Execute(ctx context.Context, driverID string) (*entity.DispatchJob, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	return uc.jobRepo.FindCurrentOfferForDriver(ctx, driverID)
}

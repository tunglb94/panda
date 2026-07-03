package app

import (
	"context"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// RejectTripUseCase records a driver rejection and offers the trip to the next
// nearest available driver.
type RejectTripUseCase struct {
	jobRepo      repository.DispatchJobRepository
	locationRepo repository.DriverLocationRepository
	tripUpdater  repository.TripUpdater
	radiusKM     float64
	searchLimit  int
}

func NewRejectTripUseCase(
	jobRepo repository.DispatchJobRepository,
	locationRepo repository.DriverLocationRepository,
	tripUpdater repository.TripUpdater,
) *RejectTripUseCase {
	return &RejectTripUseCase{
		jobRepo:      jobRepo,
		locationRepo: locationRepo,
		tripUpdater:  tripUpdater,
		radiusKM:     DefaultSearchRadiusKM,
		searchLimit:  DefaultSearchLimit,
	}
}

func (uc *RejectTripUseCase) Execute(ctx context.Context, tripID, driverID string) (*entity.DispatchJob, error) {
	if tripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id is required")
	}
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}

	job, err := uc.jobRepo.FindByTripID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := job.Reject(driverID, now); err != nil {
		return nil, err
	}

	// Offer to next nearest driver (or fail the job if exhausted).
	if err := offerNextDriver(ctx, job, uc.locationRepo, uc.tripUpdater, uc.jobRepo, uc.radiusKM, uc.searchLimit); err != nil {
		return nil, err
	}

	return job, nil
}

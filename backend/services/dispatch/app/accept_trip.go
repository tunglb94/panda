package app

import (
	"context"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// AcceptTripUseCase records driver acceptance and updates the trip to DriverAssigned.
type AcceptTripUseCase struct {
	jobRepo     repository.DispatchJobRepository
	tripUpdater repository.TripUpdater
}

func NewAcceptTripUseCase(
	jobRepo repository.DispatchJobRepository,
	tripUpdater repository.TripUpdater,
) *AcceptTripUseCase {
	return &AcceptTripUseCase{jobRepo: jobRepo, tripUpdater: tripUpdater}
}

func (uc *AcceptTripUseCase) Execute(ctx context.Context, tripID, driverID string) (*entity.DispatchJob, error) {
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
	if err := job.Accept(driverID, now); err != nil {
		return nil, err
	}

	if err := uc.tripUpdater.AssignDriver(ctx, tripID, driverID, now); err != nil {
		return nil, err
	}

	if err := uc.jobRepo.Save(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

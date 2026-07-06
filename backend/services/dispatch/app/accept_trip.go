package app

import (
	"context"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// AcceptTripUseCase records driver acceptance and atomically updates both the
// trip status and the dispatch job in a single PostgreSQL transaction.
type AcceptTripUseCase struct {
	jobRepo    repository.DispatchJobRepository
	transactor repository.Transactor
}

func NewAcceptTripUseCase(
	jobRepo repository.DispatchJobRepository,
	transactor repository.Transactor,
) *AcceptTripUseCase {
	return &AcceptTripUseCase{jobRepo: jobRepo, transactor: transactor}
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

	// Atomic: assign driver on trip + save dispatch job as assigned.
	// If either write fails the transaction is rolled back and no partial
	// state is committed.
	if err := uc.transactor.WithinTx(ctx, func(jobs repository.DispatchJobRepository, trips repository.TripUpdater) error {
		if err := trips.AssignDriver(ctx, tripID, driverID, now); err != nil {
			return err
		}
		return jobs.Save(ctx, job)
	}); err != nil {
		return nil, err
	}

	return job, nil
}

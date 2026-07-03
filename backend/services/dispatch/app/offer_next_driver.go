package app

import (
	"context"
	"time"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
)

const (
	// DefaultSearchRadiusKM is the radius used when searching for nearby drivers.
	DefaultSearchRadiusKM = 10.0
	// DefaultSearchLimit caps how many candidates the geo query returns.
	DefaultSearchLimit = 20
)

// offerNextDriver finds the closest eligible driver and issues a new offer.
// If max attempts are exhausted or no eligible driver exists, the job is marked
// failed and the trip stays in the Searching status for manual retry.
//
// Callers must have already cleared the current offer (via Reject or TimeoutOffer)
// before calling this function.
func offerNextDriver(
	ctx context.Context,
	job *entity.DispatchJob,
	locationRepo repository.DriverLocationRepository,
	tripUpdater repository.TripUpdater,
	jobRepo repository.DispatchJobRepository,
	radiusKM float64,
	limit int,
) error {
	now := time.Now().UTC()

	if job.AttemptCount >= job.MaxAttempts {
		_ = job.MarkFailed(now)
		return jobRepo.Save(ctx, job)
	}

	candidates, err := locationRepo.FindNearby(ctx, job.PickupLat, job.PickupLon, radiusKM, limit)
	if err != nil {
		return err
	}

	for _, candidate := range candidates {
		if job.HasBeenOffered(candidate.DriverID) {
			continue
		}
		active, err := locationRepo.IsActive(ctx, candidate.DriverID)
		if err != nil || !active {
			continue
		}
		if err := job.OfferToDriver(candidate.DriverID, now); err != nil {
			return err
		}
		return jobRepo.Save(ctx, job)
	}

	// No eligible driver found — exhaust the job.
	_ = job.MarkFailed(now)
	return jobRepo.Save(ctx, job)
}

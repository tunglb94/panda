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
// Matching (Vehicle/Service Catalog refactor) checks two independent
// things, both against the job's TripType/ServiceType:
//  1. Service-type match: when job.ServiceType is set, a candidate is only
//     eligible if its own reported ServiceType is an exact match (Bike only
//     offered to Bike, Car XL only to Car XL — never cross-tier). Empty
//     job.ServiceType (older caller, or a request that genuinely didn't
//     specify one) applies no service-type filtering at all — identical to
//     this function's behaviour before this catalog existed. A candidate
//     that hasn't reported a service type (candidate.ServiceType == "") is
//     excluded from any service-type-constrained search, since "unknown"
//     can't be proven to match.
//  2. Trip-type capability: a candidate is only eligible for a
//     TripTypeDelivery job if candidate.DeliveryEnabled, and only eligible
//     for a TripTypeRide job if candidate.RideEnabled. Backward
//     compatibility: DriverLocationRepository.FindNearby defaults an
//     unreported capability to RideEnabled=true/DeliveryEnabled=false
//     (mirroring migration 008's DB column defaults) rather than Go's
//     zero-value false/false — so drivers/clients that predate this field
//     keep matching Ride jobs exactly as before, and are simply not yet
//     eligible for Delivery until they explicitly opt in.
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
		if job.ServiceType != "" && candidate.ServiceType != job.ServiceType {
			continue
		}
		if job.TripType == entity.TripTypeDelivery && !candidate.DeliveryEnabled {
			continue
		}
		if job.TripType != entity.TripTypeDelivery && !candidate.RideEnabled {
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

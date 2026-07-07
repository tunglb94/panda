package repository

import (
	"context"
	"time"

	"github.com/fairride/dispatch/domain/entity"
)

// DispatchJobRepository persists DispatchJob aggregates.
type DispatchJobRepository interface {
	// Save upserts a job; trip_id is immutable after the first insert.
	Save(ctx context.Context, job *entity.DispatchJob) error

	// FindByID returns CodeNotFound if no job has the given job ID.
	FindByID(ctx context.Context, jobID string) (*entity.DispatchJob, error)

	// FindByTripID returns CodeNotFound if no job exists for the trip.
	FindByTripID(ctx context.Context, tripID string) (*entity.DispatchJob, error)

	// FindExpiredOffers returns searching jobs whose offer_expires_at is before now.
	// Used by the background engine to trigger automatic retries.
	FindExpiredOffers(ctx context.Context, now time.Time) ([]*entity.DispatchJob, error)

	// FindCurrentOfferForDriver returns the searching job that is currently
	// offering a trip to the given driver, or CodeNotFound if none exists.
	FindCurrentOfferForDriver(ctx context.Context, driverID string) (*entity.DispatchJob, error)
}

// DriverLocationRepository manages real-time driver coordinates for dispatch matching.
// Backed by Redis GEO + TTL keys; entirely independent of Phase 6 availability state.
type DriverLocationRepository interface {
	// UpdateLocation stores the driver's current coordinates and refreshes the active TTL.
	UpdateLocation(ctx context.Context, driverID string, lat, lon float64) error

	// FindNearby returns drivers within radiusKM of the given coordinates, nearest first.
	// At most limit drivers are returned.
	FindNearby(ctx context.Context, lat, lon float64, radiusKM float64, limit int) ([]*entity.NearbyDriver, error)

	// IsActive returns true if the driver has sent a location update within the TTL window.
	IsActive(ctx context.Context, driverID string) (bool, error)

	// GetLocation returns the driver's last known coordinates from the geo set.
	// Returns CodeNotFound if the driver has never reported a location.
	GetLocation(ctx context.Context, driverID string) (lat, lon float64, err error)

	// RemoveLocation removes a driver from the geo set (e.g., when they go offline).
	RemoveLocation(ctx context.Context, driverID string) error
}

// TripUpdater allows the dispatch service to update trip status directly.
// Both services share the same PostgreSQL database in the MVP.
type TripUpdater interface {
	// SetSearching transitions the trip to the Searching status.
	SetSearching(ctx context.Context, tripID string, now time.Time) error

	// AssignDriver transitions the trip to DriverAssigned status and records the driver.
	AssignDriver(ctx context.Context, tripID, driverID string, now time.Time) error
}

// Transactor executes fn inside a single PostgreSQL transaction.
// Both DispatchJobRepository and TripUpdater passed to fn are tx-scoped.
// If fn returns an error the transaction is rolled back and the error is
// returned unchanged. The caller must not retain the tx-scoped instances
// after fn returns.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(jobs DispatchJobRepository, trips TripUpdater) error) error
}

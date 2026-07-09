package repository

import (
	"context"

	"github.com/fairride/trip/domain/entity"
)

// TripRepository persists and retrieves Trip aggregates.
// All methods return *DomainError on failure.
type TripRepository interface {
	// Save upserts a trip. On insert rider_id is set; on update it is never changed.
	Save(ctx context.Context, trip *entity.Trip) error

	// FindByID returns CodeNotFound if no trip has the given ID.
	FindByID(ctx context.Context, tripID string) (*entity.Trip, error)

	// FindByRiderID returns all trips for the rider, ordered by created_at DESC.
	// Returns an empty slice (not an error) when none exist.
	FindByRiderID(ctx context.Context, riderID string) ([]*entity.Trip, error)

	// FindByDriverID returns all trips for the driver, ordered by created_at DESC.
	// Returns an empty slice (not an error) when none exist.
	FindByDriverID(ctx context.Context, driverID string) ([]*entity.Trip, error)
}

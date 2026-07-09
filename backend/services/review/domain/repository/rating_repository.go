package repository

import (
	"context"

	"github.com/fairride/review/domain/entity"
)

// RatingRepository persists and retrieves Rating aggregates.
// Ratings are immutable — Save is append-only (returns AlreadyExists if duplicate).
type RatingRepository interface {
	// Save stores a new rating. Returns CodeAlreadyExists if a rating for
	// the same (trip_id, role) already exists.
	Save(ctx context.Context, rating *entity.Rating) error

	// FindByTripAndRole returns the rating for a trip/role combination.
	// Returns CodeNotFound if no rating exists.
	FindByTripAndRole(ctx context.Context, tripID string, role entity.Role) (*entity.Rating, error)
}

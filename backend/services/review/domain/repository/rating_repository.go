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

	// FindAverageByRatee aggregates every rating rateeID has received from
	// the given rater role (e.g. role=RoleRider to average the ratings a
	// driver has received from riders). Returns count=0, avg=0 if the ratee
	// has no ratings yet — not an error, since "no ratings yet" is a normal
	// state for a new driver/rider, not a failure.
	FindAverageByRatee(ctx context.Context, rateeID string, role entity.Role) (avg float64, count int32, err error)
}

package entity

import (
	"time"

	domainerrors "github.com/fairride/shared/errors"
)

// Role identifies which party is giving the rating.
type Role string

const (
	RoleRider  Role = "rider"  // rider rates the driver
	RoleDriver Role = "driver" // driver rates the rider
)

// Rating is the aggregate root for a post-trip rating.
// Ratings are immutable once submitted.
type Rating struct {
	RatingID  string
	TripID    string
	RaterID   string // user ID of who rated
	RateeID   string // user ID of who was rated
	Role      Role
	Stars     int32 // 1-5
	Comment   string
	CreatedAt time.Time
}

// NewRating creates a validated Rating.
func NewRating(ratingID, tripID, raterID, rateeID string, role Role, stars int32, comment string, now time.Time) (*Rating, error) {
	if ratingID == "" {
		return nil, domainerrors.InvalidArgument("rating_id must not be empty")
	}
	if tripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id must not be empty")
	}
	if raterID == "" {
		return nil, domainerrors.InvalidArgument("rater_id must not be empty")
	}
	if rateeID == "" {
		return nil, domainerrors.InvalidArgument("ratee_id must not be empty")
	}
	if role != RoleRider && role != RoleDriver {
		return nil, domainerrors.InvalidArgument("role must be 'rider' or 'driver'")
	}
	if stars < 1 || stars > 5 {
		return nil, domainerrors.InvalidArgument("stars must be between 1 and 5")
	}
	return &Rating{
		RatingID:  ratingID,
		TripID:    tripID,
		RaterID:   raterID,
		RateeID:   rateeID,
		Role:      role,
		Stars:     stars,
		Comment:   comment,
		CreatedAt: now,
	}, nil
}

// ReconstituteRating rebuilds a Rating from persistence. No validation.
func ReconstituteRating(ratingID, tripID, raterID, rateeID string, role Role, stars int32, comment string, createdAt time.Time) *Rating {
	return &Rating{
		RatingID:  ratingID,
		TripID:    tripID,
		RaterID:   raterID,
		RateeID:   rateeID,
		Role:      role,
		Stars:     stars,
		Comment:   comment,
		CreatedAt: createdAt,
	}
}

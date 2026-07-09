package app

import (
	"context"
	"time"

	"github.com/fairride/review/domain/entity"
	"github.com/fairride/review/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/google/uuid"
)

// SubmitRatingUseCase records a post-trip rating.
type SubmitRatingUseCase struct {
	repo repository.RatingRepository
}

func NewSubmitRatingUseCase(repo repository.RatingRepository) *SubmitRatingUseCase {
	return &SubmitRatingUseCase{repo: repo}
}

// SubmitRatingInput carries all fields needed to create a rating.
type SubmitRatingInput struct {
	TripID  string
	RaterID string
	RateeID string
	Role    string // "rider" | "driver"
	Stars   int32
	Comment string
}

func (uc *SubmitRatingUseCase) Execute(ctx context.Context, in SubmitRatingInput) (*entity.Rating, error) {
	if in.TripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id is required")
	}
	if in.RaterID == "" {
		return nil, domainerrors.InvalidArgument("rater_id is required")
	}
	if in.RateeID == "" {
		return nil, domainerrors.InvalidArgument("ratee_id is required")
	}
	role := entity.Role(in.Role)
	rating, err := entity.NewRating(
		uuid.NewString(),
		in.TripID,
		in.RaterID,
		in.RateeID,
		role,
		in.Stars,
		in.Comment,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, rating); err != nil {
		return nil, err
	}
	return rating, nil
}

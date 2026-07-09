package app

import (
	"context"

	"github.com/fairride/review/domain/entity"
	"github.com/fairride/review/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// GetTripRatingUseCase retrieves a rating for a trip/role combination.
type GetTripRatingUseCase struct {
	repo repository.RatingRepository
}

func NewGetTripRatingUseCase(repo repository.RatingRepository) *GetTripRatingUseCase {
	return &GetTripRatingUseCase{repo: repo}
}

func (uc *GetTripRatingUseCase) Execute(ctx context.Context, tripID, role string) (*entity.Rating, error) {
	if tripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id is required")
	}
	r := entity.Role(role)
	if r != entity.RoleRider && r != entity.RoleDriver {
		return nil, domainerrors.InvalidArgument("role must be 'rider' or 'driver'")
	}
	return uc.repo.FindByTripAndRole(ctx, tripID, r)
}

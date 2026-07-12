package app

import (
	"context"

	"github.com/fairride/review/domain/entity"
	"github.com/fairride/review/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// GetAverageRatingUseCase is a read-only aggregate query, additive
// alongside SubmitRatingUseCase/GetTripRatingUseCase — it does not touch
// rating submission or Driver Economy in any way. Added for the
// Communication Module's Contact Card (Part 4), which needs a driver's (or
// rider's) overall rating, not just one trip's.
type GetAverageRatingUseCase struct {
	repo repository.RatingRepository
}

func NewGetAverageRatingUseCase(repo repository.RatingRepository) *GetAverageRatingUseCase {
	return &GetAverageRatingUseCase{repo: repo}
}

// Execute returns the average stars (0 if none yet) and rating count that
// rateeID has received from raters of the given role.
func (uc *GetAverageRatingUseCase) Execute(ctx context.Context, rateeID string, raterRole entity.Role) (float64, int32, error) {
	if rateeID == "" {
		return 0, 0, domainerrors.InvalidArgument("ratee_id is required")
	}
	if raterRole != entity.RoleRider && raterRole != entity.RoleDriver {
		return 0, 0, domainerrors.InvalidArgument("role must be 'rider' or 'driver'")
	}
	return uc.repo.FindAverageByRatee(ctx, rateeID, raterRole)
}

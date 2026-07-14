package app

import (
	"context"

	"github.com/fairride/shared/errors"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/domain/repository"
)

// GetRiderVerificationUseCase fetches a rider's KYC record by user ID.
type GetRiderVerificationUseCase struct {
	repo repository.RiderVerificationRepository
}

func NewGetRiderVerificationUseCase(repo repository.RiderVerificationRepository) *GetRiderVerificationUseCase {
	return &GetRiderVerificationUseCase{repo: repo}
}

// Execute returns CodeNotFound if the rider has never uploaded anything yet.
func (uc *GetRiderVerificationUseCase) Execute(ctx context.Context, userID string) (*entity.RiderVerification, error) {
	if userID == "" {
		return nil, errors.InvalidArgument("user id must not be empty")
	}
	return uc.repo.FindByUserID(ctx, userID)
}

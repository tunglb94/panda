package app

import (
	"context"
	"time"

	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/domain/repository"
)

// ListPendingRiderVerificationsUseCase backs the admin-only review API
// (GET /api/v1/admin/verifications/riders) — there is no dashboard UI in
// this phase, so this is how an admin sees what needs a decision.
type ListPendingRiderVerificationsUseCase struct {
	repo repository.RiderVerificationRepository
}

func NewListPendingRiderVerificationsUseCase(repo repository.RiderVerificationRepository) *ListPendingRiderVerificationsUseCase {
	return &ListPendingRiderVerificationsUseCase{repo: repo}
}

func (uc *ListPendingRiderVerificationsUseCase) Execute(ctx context.Context) ([]*entity.RiderVerification, error) {
	return uc.repo.ListByStatus(ctx, entity.RiderKYCPending)
}

// ReviewRiderVerificationUseCase applies an admin's approve/reject decision.
type ReviewRiderVerificationUseCase struct {
	repo repository.RiderVerificationRepository
}

func NewReviewRiderVerificationUseCase(repo repository.RiderVerificationRepository) *ReviewRiderVerificationUseCase {
	return &ReviewRiderVerificationUseCase{repo: repo}
}

func (uc *ReviewRiderVerificationUseCase) Approve(ctx context.Context, userID, reviewer string) (*entity.RiderVerification, error) {
	v, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := v.Approve(reviewer, time.Now()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (uc *ReviewRiderVerificationUseCase) Reject(ctx context.Context, userID, reviewer, reason string) (*entity.RiderVerification, error) {
	v, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := v.Reject(reviewer, reason, time.Now()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

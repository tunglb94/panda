package app

import (
	"context"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
)

// KYCSummary is the Admin Review Dashboard's 4 summary cards (Phần 10) —
// driver-verification counts by status. Counts, not full rows, so this
// stays cheap regardless of how many drivers exist.
type KYCSummary struct {
	Pending  int
	Approved int
	Rejected int
	Expired  int
}

type GetKYCSummaryUseCase struct {
	repo repository.DriverVerificationRepository
}

func NewGetKYCSummaryUseCase(repo repository.DriverVerificationRepository) *GetKYCSummaryUseCase {
	return &GetKYCSummaryUseCase{repo: repo}
}

func (uc *GetKYCSummaryUseCase) Execute(ctx context.Context) (*KYCSummary, error) {
	pending, err := uc.repo.CountByStatus(ctx, entity.KYCPending)
	if err != nil {
		return nil, err
	}
	approved, err := uc.repo.CountByStatus(ctx, entity.KYCApproved)
	if err != nil {
		return nil, err
	}
	rejected, err := uc.repo.CountByStatus(ctx, entity.KYCRejected)
	if err != nil {
		return nil, err
	}
	expired, err := uc.repo.CountByStatus(ctx, entity.KYCExpired)
	if err != nil {
		return nil, err
	}
	return &KYCSummary{Pending: pending, Approved: approved, Rejected: rejected, Expired: expired}, nil
}

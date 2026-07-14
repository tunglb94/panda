package app

import (
	"context"
	"time"

	"github.com/fairride/promotion/domain/repository"
)

// VoucherStats is the Admin app's per-voucher wallet/redemption summary
// (Voucher & Promotion Hardening Phase 4 — "Issued/Redeemed/Remaining/Expired
// cho từng voucher"). Remaining is read straight off the voucher's own
// RemainingBudget (already tracked), not recomputed here.
type VoucherStats struct {
	Issued   int64
	Redeemed int64
	Expired  int64
}

type VoucherStatsUseCase struct {
	repo repository.PromotionRepository
}

func NewVoucherStatsUseCase(repo repository.PromotionRepository) *VoucherStatsUseCase {
	return &VoucherStatsUseCase{repo: repo}
}

func (uc *VoucherStatsUseCase) Execute(ctx context.Context, voucherID string) (*VoucherStats, error) {
	issued, err := uc.repo.CountIssued(ctx, voucherID)
	if err != nil {
		return nil, err
	}
	redeemed, err := uc.repo.CountRedeemed(ctx, voucherID)
	if err != nil {
		return nil, err
	}
	expired, err := uc.repo.CountExpiredIssuances(ctx, voucherID, time.Now())
	if err != nil {
		return nil, err
	}
	return &VoucherStats{Issued: issued, Redeemed: redeemed, Expired: expired}, nil
}

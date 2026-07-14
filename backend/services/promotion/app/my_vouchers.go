package app

import (
	"context"
	"time"

	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/promotion/domain/repository"
)

// MyVouchersResult is the Rider app's voucher wallet: Available (issued to
// this rider, not yet used, campaign still active), Used (this rider has
// redeemed — status="redeemed"), Expired (issued to this rider but its
// window/budget lapsed before use). Only vouchers individually issued to
// this rider via PromotionRepository.IssueToRider appear here — never the
// full campaign catalog (see VoucherIssuance's doc comment).
type MyVouchersResult struct {
	Available []*entity.Voucher
	Used      []*entity.RedemptionRecord
	Expired   []*entity.Voucher
}

type MyVouchersUseCase struct {
	repo repository.PromotionRepository
}

func NewMyVouchersUseCase(repo repository.PromotionRepository) *MyVouchersUseCase {
	return &MyVouchersUseCase{repo: repo}
}

func (uc *MyVouchersUseCase) Execute(ctx context.Context, riderID string) (*MyVouchersResult, error) {
	issuances, err := uc.repo.ListIssuancesForRider(ctx, riderID)
	if err != nil {
		return nil, err
	}
	redemptions, err := uc.repo.ListRedemptionsByRider(ctx, riderID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := &MyVouchersResult{}
	for _, iss := range issuances {
		if iss.Status == entity.IssuanceStatusUsed || iss.Voucher == nil {
			continue
		}
		if iss.Voucher.EffectiveState(now) == "active" {
			result.Available = append(result.Available, iss.Voucher)
		} else {
			result.Expired = append(result.Expired, iss.Voucher)
		}
	}
	for _, r := range redemptions {
		if r.Status == "redeemed" {
			result.Used = append(result.Used, r)
		}
	}
	return result, nil
}

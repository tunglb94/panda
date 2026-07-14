package app

import (
	"context"
	"time"

	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/promotion/domain/repository"
)

// ReviewVoucherUseCase applies the Admin app's Enable/Disable/Delete
// actions — thin wrappers around Voucher's existing Activate/Pause/Cancel
// domain methods (reused as-is, not reimplemented) plus a Save. "Delete" in
// the Admin CRUD sense is Cancel (a soft, permanent stop — see Voucher.Cancel's
// doc comment): campaign/redemption history is never hard-deleted from the
// database.
type ReviewVoucherUseCase struct {
	repo repository.PromotionRepository
}

func NewReviewVoucherUseCase(repo repository.PromotionRepository) *ReviewVoucherUseCase {
	return &ReviewVoucherUseCase{repo: repo}
}

func (uc *ReviewVoucherUseCase) Enable(ctx context.Context, id string) (*entity.Voucher, error) {
	return uc.transition(ctx, id, func(v *entity.Voucher, now time.Time) error { return v.Activate(now) })
}

func (uc *ReviewVoucherUseCase) Disable(ctx context.Context, id string) (*entity.Voucher, error) {
	return uc.transition(ctx, id, func(v *entity.Voucher, now time.Time) error { return v.Pause(now) })
}

func (uc *ReviewVoucherUseCase) Delete(ctx context.Context, id string) (*entity.Voucher, error) {
	return uc.transition(ctx, id, func(v *entity.Voucher, now time.Time) error { return v.Cancel(now) })
}

func (uc *ReviewVoucherUseCase) transition(ctx context.Context, id string, apply func(*entity.Voucher, time.Time) error) (*entity.Voucher, error) {
	v, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apply(v, time.Now()); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

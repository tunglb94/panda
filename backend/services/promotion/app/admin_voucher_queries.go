package app

import (
	"context"

	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/promotion/domain/repository"
)

// ListVouchersUseCase backs the Admin app's CRUD list — every campaign, no
// filtering (see repository.PromotionRepository.FindAll's doc comment).
type ListVouchersUseCase struct {
	repo repository.PromotionRepository
}

func NewListVouchersUseCase(repo repository.PromotionRepository) *ListVouchersUseCase {
	return &ListVouchersUseCase{repo: repo}
}

func (uc *ListVouchersUseCase) Execute(ctx context.Context) ([]*entity.Voucher, error) {
	return uc.repo.FindAll(ctx)
}

// GetVoucherUseCase fetches one voucher by ID — the Admin app's edit-form load.
type GetVoucherUseCase struct {
	repo repository.PromotionRepository
}

func NewGetVoucherUseCase(repo repository.PromotionRepository) *GetVoucherUseCase {
	return &GetVoucherUseCase{repo: repo}
}

func (uc *GetVoucherUseCase) Execute(ctx context.Context, id string) (*entity.Voucher, error) {
	return uc.repo.FindByID(ctx, id)
}

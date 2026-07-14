package app

import (
	"context"
	"strings"
	"time"

	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/promotion/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// UpdateVoucherInput mirrors CreateVoucherInput minus Enabled (status is
// changed via ReviewVoucherUseCase's Enable/Disable/Cancel, not here — a
// single "update" call silently flipping status too would make the Admin
// app's audit trail ambiguous about which action actually happened).
type UpdateVoucherInput struct {
	ID              string
	Code            string
	Name            string
	Description     string
	DiscountType    entity.DiscountType
	DiscountValue   int64
	MaxDiscount     int64
	MinOrder        int64
	StartTime       time.Time
	EndTime         time.Time
	MaxUsage        int64
	MaxUsagePerUser int64
	ServiceTypes    []string
	TripTypes       []string
	Campaign        string
}

// UpdateVoucherUseCase edits a voucher's non-status, non-budget fields.
// Budget/RemainingBudget/UsageCount/Status are intentionally not editable
// here — budget changes and status changes are distinct admin actions with
// their own audit implications, not a side effect of "editing the form."
type UpdateVoucherUseCase struct {
	repo repository.PromotionRepository
}

func NewUpdateVoucherUseCase(repo repository.PromotionRepository) *UpdateVoucherUseCase {
	return &UpdateVoucherUseCase{repo: repo}
}

func (uc *UpdateVoucherUseCase) Execute(ctx context.Context, in UpdateVoucherInput) (*entity.Voucher, error) {
	v, err := uc.repo.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, domainerrors.InvalidArgument("voucher name is required")
	}
	if !in.DiscountType.Valid() {
		return nil, domainerrors.InvalidArgument("invalid discount_type: " + string(in.DiscountType))
	}
	if in.DiscountValue < 0 || in.MaxDiscount < 0 || in.MinOrder < 0 {
		return nil, domainerrors.InvalidArgument("discount_value/max_discount/min_order cannot be negative")
	}
	if in.DiscountType == entity.DiscountTypePercentage && in.DiscountValue > 100 {
		return nil, domainerrors.InvalidArgument("percentage discount_value cannot exceed 100")
	}
	if in.MaxUsage < 0 || in.MaxUsagePerUser < 0 {
		return nil, domainerrors.InvalidArgument("max_usage and max_usage_per_user cannot be negative")
	}
	if in.EndTime.IsZero() || !in.EndTime.After(in.StartTime) {
		return nil, domainerrors.InvalidArgument("end_time must be after start_time")
	}

	v.Code = strings.TrimSpace(in.Code)
	v.Name = name
	v.Description = in.Description
	v.DiscountType = in.DiscountType
	v.DiscountValue = in.DiscountValue
	v.MaxDiscount = in.MaxDiscount
	v.MinOrder = in.MinOrder
	v.StartTime = in.StartTime
	v.EndTime = in.EndTime
	v.MaxUsage = in.MaxUsage
	v.MaxUsagePerUser = in.MaxUsagePerUser
	v.ServiceTypes = in.ServiceTypes
	v.TripTypes = in.TripTypes
	v.Campaign = strings.TrimSpace(in.Campaign)
	v.UpdatedAt = time.Now()

	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

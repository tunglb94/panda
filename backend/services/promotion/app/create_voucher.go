package app

import (
	"context"
	"strings"
	"time"

	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/promotion/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// CreateVoucherInput carries every field the Admin app's voucher CRUD form
// collects. Fields not exposed by that simple form (Priority, Cities,
// Membership, NewUserOnly, Combinable, Stackable) get sensible defaults —
// see Execute. Every voucher created this way is Type=ManualCoupon (BRB
// §3.2.9); the richer BRB campaign types (First Ride, Birthday, ...) are
// seeded separately, not through this CRUD.
type CreateVoucherInput struct {
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
	Budget          int64
	ServiceTypes    []string
	TripTypes       []string
	Campaign        string
	Enabled         bool
}

type CreateVoucherUseCase struct {
	repo repository.PromotionRepository
}

func NewCreateVoucherUseCase(repo repository.PromotionRepository) *CreateVoucherUseCase {
	return &CreateVoucherUseCase{repo: repo}
}

func (uc *CreateVoucherUseCase) Execute(ctx context.Context, in CreateVoucherInput) (*entity.Voucher, error) {
	id, err := newID()
	if err != nil {
		return nil, domainerrors.Internal("generate voucher id").WithMeta("error", err.Error())
	}
	now := time.Now()
	v, err := entity.NewVoucher(
		id, strings.TrimSpace(in.Code), in.Name, in.Description,
		0, // Priority — not exposed by this simple CRUD form
		in.StartTime, in.EndTime,
		in.MaxUsage, in.MaxUsagePerUser, in.Budget,
		in.DiscountType, in.DiscountValue, in.MaxDiscount, in.MinOrder,
		nil, nil, nil, // VehicleTypes/Cities/Membership — not exposed by this form
		false, false, false, // NewUserOnly/Combinable/Stackable — not exposed by this form
		entity.PromotionTypeManualCoupon,
		now,
	)
	if err != nil {
		return nil, err
	}
	v.ServiceTypes = in.ServiceTypes
	v.TripTypes = in.TripTypes
	v.Campaign = strings.TrimSpace(in.Campaign)

	if in.Enabled {
		if err := v.Activate(now); err != nil {
			return nil, err
		}
	}

	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

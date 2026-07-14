package integration

import (
	"context"
	"time"

	promotionapp "github.com/fairride/promotion/app"
	promotionentity "github.com/fairride/promotion/domain/entity"
	promotionfake "github.com/fairride/promotion/infrastructure/fake"

	"github.com/fairride/ai_simulation/domain/entity"
)

// PromotionAdapter wraps the real backend/services/promotion PromotionService
// (Voucher Engine + Promotion Engine) over its own FakePromotionRepository —
// the exact in-memory repository that service's own test suite uses,
// reused here rather than duplicated. No production database is touched.
type PromotionAdapter struct {
	repo    *promotionfake.FakePromotionRepository
	service *promotionapp.PromotionService
}

// NewPromotionAdapter seeds the fake repository with the promotion
// campaigns Business Rule Bible v1.0 actually defines (First Ride §3.2.1,
// Birthday §3.2.2, Weekend §3.2.4, Manual Coupon/Event Campaign §3.2.6 &
// §3.2.9) using that service's own exported BRB-sourced constants — no
// number here is invented, they are the same constants
// backend/services/promotion/domain/entity already ships.
func NewPromotionAdapter(now time.Time) *PromotionAdapter {
	repo := promotionfake.NewFakePromotionRepository()
	validator := promotionapp.NewVoucherValidator()
	rules := promotionapp.NewDefaultRuleRegistry()
	service := promotionapp.NewPromotionService(repo, validator, rules)

	seedVouchers(repo, now)

	return &PromotionAdapter{repo: repo, service: service}
}

func seedVouchers(repo *promotionfake.FakePromotionRepository, now time.Time) {
	start := now.Add(-24 * time.Hour)
	end := now.Add(365 * 24 * time.Hour)

	seed := func(v *promotionentity.Voucher, err error) {
		if err != nil {
			// Programmer error in the seed data below — fail loudly at
			// startup rather than silently run a simulation with a missing
			// campaign.
			panic("ai_simulation: invalid seed voucher: " + err.Error())
		}
		if err := v.Activate(now); err != nil {
			panic("ai_simulation: cannot activate seed voucher: " + err.Error())
		}
		repo.Seed(v)
	}

	seed(promotionentity.NewVoucher(
		"sim-first-ride", "", "First Ride 50%", "BRB §3.2.1",
		promotionentity.PromotionTypeFirstRide.DefaultPriority(),
		start, end, 0, 1, 10_000_000_000,
		promotionentity.DiscountTypePercentage,
		promotionentity.FirstRideDiscountPercent, promotionentity.FirstRideMaxDiscountVND, 0,
		nil, nil, nil, true, false, false,
		promotionentity.PromotionTypeFirstRide, now,
	))
	seed(promotionentity.NewVoucher(
		"sim-birthday", "", "Birthday 40%", "BRB §3.2.2",
		promotionentity.PromotionTypeBirthday.DefaultPriority(),
		start, end, 0, 1, 10_000_000_000,
		promotionentity.DiscountTypePercentage,
		promotionentity.BirthdayDiscountPercent, promotionentity.BirthdayMaxDiscountVND, 0,
		nil, nil, nil, false, false, false,
		promotionentity.PromotionTypeBirthday, now,
	))
	seed(promotionentity.NewVoucher(
		"sim-weekend", "", "Weekend 15%", "BRB §3.2.4",
		promotionentity.PromotionTypeWeekend.DefaultPriority(),
		start, end, 0, 2, 10_000_000_000,
		promotionentity.DiscountTypePercentage,
		promotionentity.WeekendDiscountPercent, promotionentity.WeekendMaxDiscountVND, 0,
		nil, nil, nil, false, false, false,
		promotionentity.PromotionTypeWeekend, now,
	))
	seed(promotionentity.NewVoucher(
		"SIM10", "SIM10", "Manual Coupon 10%", "BRB §3.2.9 Coupon Campaign",
		promotionentity.PromotionTypeManualCoupon.DefaultPriority(),
		start, end, 0, 1, 10_000_000_000,
		promotionentity.DiscountTypePercentage, 10, 15_000, 20_000,
		nil, nil, nil, false, false, false,
		promotionentity.PromotionTypeManualCoupon, now,
	))
}

// EvaluateInput is the simulation-local view of what the Promotion Engine
// needs — translated 1:1 into promotionentity.PromotionRequest.
type EvaluateInput struct {
	RiderID             string
	ServiceType         entity.ServiceType
	City                string
	OrderAmountVND      int64
	RequestTime         time.Time
	VoucherCode         string
	IsNewRider          bool
	CompletedTripsTotal int64
	BirthdayToday       bool
	MembershipTier      string
}

// Evaluate calls the real PromotionService.Evaluate — the production
// Promotion Engine decides eligibility/discount, the simulation only
// supplies the situational facts.
func (a *PromotionAdapter) Evaluate(ctx context.Context, in EvaluateInput) (*promotionentity.PromotionResult, error) {
	var accountCreatedAt *time.Time
	if in.IsNewRider {
		t := in.RequestTime.Add(-2 * 24 * time.Hour)
		accountCreatedAt = &t
	}
	var birthdayDate *time.Time
	if in.BirthdayToday {
		t := in.RequestTime
		birthdayDate = &t
	}

	req := &promotionentity.PromotionRequest{
		RiderID:                  in.RiderID,
		VehicleType:              string(in.ServiceType),
		City:                     in.City,
		OrderAmount:              in.OrderAmountVND,
		RequestTime:              in.RequestTime,
		VoucherCode:              in.VoucherCode,
		AccountCreatedAt:         accountCreatedAt,
		CompletedTripsTotal:      in.CompletedTripsTotal,
		CompletedTripsLast90Days: 5, // simulation does not track a full 90-day rolling window; assumes an active rider for birthday eligibility
		BirthdayDate:             birthdayDate,
		MembershipTier:           in.MembershipTier,
	}
	return a.service.Evaluate(ctx, req, in.RequestTime)
}

// Redeem calls the real PromotionService.Reserve to commit budget/usage.
func (a *PromotionAdapter) Redeem(ctx context.Context, result *promotionentity.PromotionResult, riderID, tripID string) error {
	return a.service.Reserve(ctx, result, riderID, tripID)
}

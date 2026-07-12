package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fairride/promotion/app"
	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/promotion/domain/repository"
	"github.com/fairride/promotion/infrastructure/fake"
	sharederrors "github.com/fairride/shared/errors"
)

func newService(repo repository.PromotionRepository) *app.PromotionService {
	return app.NewPromotionService(repo, app.NewVoucherValidator(), app.NewDefaultRuleRegistry())
}

func seedFirstRideVoucher(t *testing.T, repo *fake.FakePromotionRepository, id string, priority int) *entity.Voucher {
	t.Helper()
	v, err := entity.NewVoucher(
		id, "", "First Ride 50%", "BRB §3.2.1", priority,
		now.Add(-time.Hour), now.Add(24*time.Hour),
		0, 1, 1_000_000,
		entity.DiscountTypePercentage, entity.FirstRideDiscountPercent, entity.FirstRideMaxDiscountVND, 0,
		nil, nil, nil,
		true, false, false,
		entity.PromotionTypeFirstRide, now,
	)
	if err != nil {
		t.Fatalf("seedFirstRideVoucher: %v", err)
	}
	if err := v.Activate(now); err != nil {
		t.Fatalf("activate: %v", err)
	}
	repo.Seed(v)
	return v
}

func seedWeekendVoucher(t *testing.T, repo *fake.FakePromotionRepository, id string, priority int) *entity.Voucher {
	t.Helper()
	v, err := entity.NewVoucher(
		id, "", "Weekend 15%", "BRB §3.2.4", priority,
		now.Add(-time.Hour), now.Add(24*time.Hour),
		0, 2, 1_000_000,
		entity.DiscountTypePercentage, entity.WeekendDiscountPercent, entity.WeekendMaxDiscountVND, 0,
		nil, nil, nil,
		false, false, false,
		entity.PromotionTypeWeekend, now,
	)
	if err != nil {
		t.Fatalf("seedWeekendVoucher: %v", err)
	}
	if err := v.Activate(now); err != nil {
		t.Fatalf("activate: %v", err)
	}
	repo.Seed(v)
	return v
}

func firstRideRequest() *entity.PromotionRequest {
	created := now.Add(-time.Hour)
	return &entity.PromotionRequest{
		RiderID:             "rider-1",
		VehicleType:         "car",
		City:                "Ho Chi Minh City",
		OrderAmount:         100_000,
		RequestTime:         now,
		CompletedTripsTotal: 0,
		AccountCreatedAt:    &created,
	}
}

func TestPromotionService_Evaluate_FirstRideApplied(t *testing.T) {
	repo := fake.NewFakePromotionRepository()
	seedFirstRideVoucher(t, repo, "v-first-ride", entity.PromotionTypeFirstRide.DefaultPriority())
	svc := newService(repo)

	result, err := svc.Evaluate(context.Background(), firstRideRequest(), now)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !result.Applied {
		t.Fatalf("expected a discount to apply, warnings: %v", result.Warnings)
	}
	// 50% of 100,000 = 50,000, capped by BRB §3.2.1 max 30,000
	if result.DiscountAmount != 30_000 {
		t.Fatalf("expected discount 30000 (capped), got %d", result.DiscountAmount)
	}
	if result.FinalOrderAmount != 70_000 {
		t.Fatalf("expected final amount 70000, got %d", result.FinalOrderAmount)
	}
}

func TestPromotionService_Evaluate_ManualCouponByCode(t *testing.T) {
	repo := fake.NewFakePromotionRepository()
	v, err := entity.NewVoucher(
		"v-coupon", "SUMMER50", "Summer Sale", "manual coupon", entity.PromotionTypeManualCoupon.DefaultPriority(),
		now.Add(-time.Hour), now.Add(24*time.Hour),
		0, 1, 1_000_000,
		entity.DiscountTypeFlat, 20_000, 0, 0,
		nil, nil, nil,
		false, false, false,
		entity.PromotionTypeManualCoupon, now,
	)
	if err != nil {
		t.Fatalf("build voucher: %v", err)
	}
	if err := v.Activate(now); err != nil {
		t.Fatalf("activate: %v", err)
	}
	repo.Seed(v)

	req := firstRideRequest()
	req.VoucherCode = "summer50" // case-insensitive
	svc := newService(repo)

	result, err := svc.Evaluate(context.Background(), req, now)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !result.Applied || result.DiscountAmount != 20_000 {
		t.Fatalf("expected flat 20000 discount applied, got %+v", result)
	}
}

func TestPromotionService_Evaluate_InvalidCodeReturnsNoDiscount(t *testing.T) {
	repo := fake.NewFakePromotionRepository()
	svc := newService(repo)

	req := firstRideRequest()
	req.VoucherCode = "DOES-NOT-EXIST"

	result, err := svc.Evaluate(context.Background(), req, now)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result.Applied {
		t.Fatal("expected no discount for an unknown voucher code")
	}
}

func TestPromotionService_Evaluate_PriorityOrdering(t *testing.T) {
	// Saturday + eligible-for-both First Ride (priority 30) and Weekend (priority 50).
	// BRB §3.4: First Ride outranks Weekend.
	repo := fake.NewFakePromotionRepository()
	seedFirstRideVoucher(t, repo, "v-first-ride", entity.PromotionTypeFirstRide.DefaultPriority())
	seedWeekendVoucher(t, repo, "v-weekend", entity.PromotionTypeWeekend.DefaultPriority())
	svc := newService(repo)

	saturday := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	created := saturday.Add(-time.Hour)
	req := &entity.PromotionRequest{
		RiderID:             "rider-1",
		VehicleType:         "car",
		City:                "Ho Chi Minh City",
		OrderAmount:         100_000,
		RequestTime:         saturday,
		CompletedTripsTotal: 0,
		AccountCreatedAt:    &created,
	}

	result, err := svc.Evaluate(context.Background(), req, saturday)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if !result.Applied || result.Type != entity.PromotionTypeFirstRide {
		t.Fatalf("expected First Ride to win over Weekend by priority, got %+v", result)
	}
	if len(result.Warnings) == 0 || !strings.Contains(strings.Join(result.Warnings, ";"), "Weekend") {
		t.Fatalf("expected a warning noting Weekend was eligible but not applied (BRB §4.7 one voucher per trip), got %v", result.Warnings)
	}
}

func TestPromotionService_Evaluate_WrongCityRejected(t *testing.T) {
	repo := fake.NewFakePromotionRepository()
	v := seedFirstRideVoucher(t, repo, "v-first-ride", entity.PromotionTypeFirstRide.DefaultPriority())
	v.Cities = []string{"Hanoi"}
	repo.Seed(v)
	svc := newService(repo)

	req := firstRideRequest() // City = Ho Chi Minh City
	result, err := svc.Evaluate(context.Background(), req, now)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result.Applied {
		t.Fatal("expected no discount: voucher restricted to Hanoi")
	}
}

func TestPromotionService_RedeemThenPerUserLimitBlocksReuse(t *testing.T) {
	repo := fake.NewFakePromotionRepository()
	seedFirstRideVoucher(t, repo, "v-first-ride", entity.PromotionTypeFirstRide.DefaultPriority())
	svc := newService(repo)
	ctx := context.Background()

	req := firstRideRequest()
	result, err := svc.Evaluate(ctx, req, now)
	if err != nil || !result.Applied {
		t.Fatalf("expected first evaluate to apply, err=%v result=%+v", err, result)
	}

	if err := svc.Redeem(ctx, result, req.RiderID, "trip-1"); err != nil {
		t.Fatalf("Redeem: %v", err)
	}

	// Same rider evaluates again (e.g. trying to reuse First Ride) — BRB §4.6
	// per-rider usage limit of 1 must now block it, even though the rider still
	// structurally looks like a "first ride" candidate.
	result2, err := svc.Evaluate(ctx, req, now)
	if err != nil {
		t.Fatalf("Evaluate (second): %v", err)
	}
	if result2.Applied {
		t.Fatal("expected second evaluation for the same rider to be blocked by per-user usage limit")
	}
}

func TestPromotionService_ReleaseRedemption_ReinstatesBudget(t *testing.T) {
	repo := fake.NewFakePromotionRepository()
	seedFirstRideVoucher(t, repo, "v-first-ride", entity.PromotionTypeFirstRide.DefaultPriority())
	svc := newService(repo)
	ctx := context.Background()

	req := firstRideRequest()
	result, err := svc.Evaluate(ctx, req, now)
	if err != nil || !result.Applied {
		t.Fatalf("expected evaluate to apply, err=%v result=%+v", err, result)
	}
	if err := svc.Redeem(ctx, result, req.RiderID, "trip-1"); err != nil {
		t.Fatalf("Redeem: %v", err)
	}

	v, err := repo.FindByID(ctx, "v-first-ride")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if v.RemainingBudget != 1_000_000-30_000 {
		t.Fatalf("expected budget decremented after redeem, got %d", v.RemainingBudget)
	}

	if err := svc.ReleaseRedemption(ctx, result, req.RiderID, "trip-1"); err != nil {
		t.Fatalf("ReleaseRedemption: %v", err)
	}

	v, err = repo.FindByID(ctx, "v-first-ride")
	if err != nil {
		t.Fatalf("FindByID after release: %v", err)
	}
	if v.RemainingBudget != 1_000_000 {
		t.Fatalf("expected budget reinstated to 1000000 after release, got %d", v.RemainingBudget)
	}

	// After release, the rider should be eligible again.
	result2, err := svc.Evaluate(ctx, req, now)
	if err != nil {
		t.Fatalf("Evaluate (after release): %v", err)
	}
	if !result2.Applied {
		t.Fatal("expected rider to be eligible again after release")
	}
}

// ─── hand-written mock: verifies PromotionService propagates repository errors ──

type mockErrRepo struct {
	repository.PromotionRepository // embed nil; only the methods below are exercised
	findErr                        error
}

func (m *mockErrRepo) FindByCode(context.Context, string) (*entity.Voucher, error) {
	return nil, m.findErr
}

func (m *mockErrRepo) FindAutoApplyCandidates(context.Context, string, string, []entity.PromotionType) ([]*entity.Voucher, error) {
	return nil, m.findErr
}

func TestPromotionService_Evaluate_PropagatesRepositoryInternalError(t *testing.T) {
	mock := &mockErrRepo{findErr: sharederrors.Internal("db is down")}
	svc := newService(mock)

	_, err := svc.Evaluate(context.Background(), firstRideRequest(), now)
	if !sharederrors.IsCode(err, sharederrors.CodeInternalError) {
		t.Fatalf("expected internal error to propagate, got %v", err)
	}
}

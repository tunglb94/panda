package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/promotion/domain/entity"
	"github.com/fairride/promotion/infrastructure/postgres"
	sharederrors "github.com/fairride/shared/errors"
)

func newTestVoucher(t *testing.T, id, code string, promoType entity.PromotionType, budget int64) *entity.Voucher {
	t.Helper()
	v, err := entity.NewVoucher(
		id, code, "Test Voucher "+id, "integration test fixture",
		promoType.DefaultPriority(),
		testNow.Add(-24*time.Hour), testNow.Add(24*time.Hour),
		0, 1, budget,
		entity.DiscountTypePercentage, 50, 30_000, 0,
		nil, nil, nil,
		true, false, false,
		promoType,
		testNow,
	)
	if err != nil {
		t.Fatalf("newTestVoucher: %v", err)
	}
	if err := v.Activate(testNow); err != nil {
		t.Fatalf("activate: %v", err)
	}
	return v
}

func TestVoucherRepository_SaveAndFindByID(t *testing.T) {
	if testPool == nil {
		t.Skip("DATABASE_URL not set")
	}
	setupTest(t)
	ctx := context.Background()
	repo := postgres.NewVoucherRepository(testPool)

	v := newTestVoucher(t, "v-1", "", entity.PromotionTypeFirstRide, 1_000_000)
	if err := repo.Save(ctx, v); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.FindByID(ctx, "v-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Name != v.Name || got.Budget != v.Budget || got.Type != v.Type {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, v)
	}
}

func TestVoucherRepository_FindByID_NotFound(t *testing.T) {
	if testPool == nil {
		t.Skip("DATABASE_URL not set")
	}
	setupTest(t)
	ctx := context.Background()
	repo := postgres.NewVoucherRepository(testPool)

	_, err := repo.FindByID(ctx, "does-not-exist")
	if !sharederrors.IsCode(err, sharederrors.CodeNotFound) {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestVoucherRepository_FindByCode(t *testing.T) {
	if testPool == nil {
		t.Skip("DATABASE_URL not set")
	}
	setupTest(t)
	ctx := context.Background()
	repo := postgres.NewVoucherRepository(testPool)

	v := newTestVoucher(t, "v-coupon", "SUMMER50", entity.PromotionTypeManualCoupon, 1_000_000)
	if err := repo.Save(ctx, v); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.FindByCode(ctx, "summer50")
	if err != nil {
		t.Fatalf("FindByCode (case-insensitive): %v", err)
	}
	if got.ID != "v-coupon" {
		t.Fatalf("expected v-coupon, got %s", got.ID)
	}
}

func TestVoucherRepository_FindAutoApplyCandidates(t *testing.T) {
	if testPool == nil {
		t.Skip("DATABASE_URL not set")
	}
	setupTest(t)
	ctx := context.Background()
	repo := postgres.NewVoucherRepository(testPool)

	autoApply := newTestVoucher(t, "v-auto", "", entity.PromotionTypeFirstRide, 1_000_000)
	coupon := newTestVoucher(t, "v-manual", "CODE123", entity.PromotionTypeManualCoupon, 1_000_000)
	if err := repo.Save(ctx, autoApply); err != nil {
		t.Fatalf("save autoApply: %v", err)
	}
	if err := repo.Save(ctx, coupon); err != nil {
		t.Fatalf("save coupon: %v", err)
	}

	candidates, err := repo.FindAutoApplyCandidates(ctx, "Hanoi", "car", entity.AllPromotionTypes())
	if err != nil {
		t.Fatalf("FindAutoApplyCandidates: %v", err)
	}
	if len(candidates) != 1 || candidates[0].ID != "v-auto" {
		t.Fatalf("expected only v-auto (code-less), got %+v", candidates)
	}
}

func TestVoucherRepository_RecordAndReleaseRedemption(t *testing.T) {
	if testPool == nil {
		t.Skip("DATABASE_URL not set")
	}
	setupTest(t)
	ctx := context.Background()
	repo := postgres.NewVoucherRepository(testPool)

	v := newTestVoucher(t, "v-redeem", "", entity.PromotionTypeFirstRide, 100_000)
	if err := repo.Save(ctx, v); err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := repo.RecordRedemption(ctx, "v-redeem", "rider-1", "trip-1", 30_000); err != nil {
		t.Fatalf("RecordRedemption: %v", err)
	}

	count, err := repo.UsageCountForRider(ctx, "v-redeem", "rider-1")
	if err != nil {
		t.Fatalf("UsageCountForRider: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected usage count 1, got %d", count)
	}

	got, err := repo.FindByID(ctx, "v-redeem")
	if err != nil {
		t.Fatalf("FindByID after redeem: %v", err)
	}
	if got.RemainingBudget != 70_000 {
		t.Fatalf("expected remaining budget 70000, got %d", got.RemainingBudget)
	}

	if err := repo.ReleaseRedemption(ctx, "v-redeem", "rider-1", "trip-1", 30_000); err != nil {
		t.Fatalf("ReleaseRedemption: %v", err)
	}

	count, err = repo.UsageCountForRider(ctx, "v-redeem", "rider-1")
	if err != nil {
		t.Fatalf("UsageCountForRider after release: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected usage count 0 after release, got %d", count)
	}

	got, err = repo.FindByID(ctx, "v-redeem")
	if err != nil {
		t.Fatalf("FindByID after release: %v", err)
	}
	if got.RemainingBudget != 100_000 {
		t.Fatalf("expected remaining budget reinstated to 100000, got %d", got.RemainingBudget)
	}
}

func TestVoucherRepository_RecordRedemption_BudgetExhausted(t *testing.T) {
	if testPool == nil {
		t.Skip("DATABASE_URL not set")
	}
	setupTest(t)
	ctx := context.Background()
	repo := postgres.NewVoucherRepository(testPool)

	v := newTestVoucher(t, "v-tight", "", entity.PromotionTypeFirstRide, 20_000)
	if err := repo.Save(ctx, v); err != nil {
		t.Fatalf("save: %v", err)
	}

	err := repo.RecordRedemption(ctx, "v-tight", "rider-1", "trip-1", 30_000)
	if !sharederrors.IsCode(err, sharederrors.CodeResourceExhausted) {
		t.Fatalf("expected ResourceExhausted, got %v", err)
	}
}

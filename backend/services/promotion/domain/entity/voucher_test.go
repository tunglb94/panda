package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/promotion/domain/entity"
	sharederrors "github.com/fairride/shared/errors"
)

var now = time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)

func validVoucherArgs() (id, code, name, description string) {
	return "v-1", "", "First Ride 50%", "test fixture"
}

func TestNewVoucher_RejectsZeroBudget(t *testing.T) {
	id, code, name, desc := validVoucherArgs()
	_, err := entity.NewVoucher(
		id, code, name, desc, 1,
		now, now.Add(24*time.Hour),
		0, 1, 0, // budget = 0
		entity.DiscountTypePercentage, 50, 30_000, 0,
		nil, nil, nil,
		true, false, false,
		entity.PromotionTypeFirstRide, now,
	)
	if !sharederrors.IsCode(err, sharederrors.CodeInvalidArgument) {
		t.Fatalf("expected InvalidArgument for zero budget (BRB §3.3 Rule 1), got %v", err)
	}
}

func TestNewVoucher_RejectsOpenEndedCampaign(t *testing.T) {
	id, code, name, desc := validVoucherArgs()
	_, err := entity.NewVoucher(
		id, code, name, desc, 1,
		now, time.Time{}, // no end time
		0, 1, 100_000,
		entity.DiscountTypePercentage, 50, 30_000, 0,
		nil, nil, nil,
		true, false, false,
		entity.PromotionTypeFirstRide, now,
	)
	if !sharederrors.IsCode(err, sharederrors.CodeInvalidArgument) {
		t.Fatalf("expected InvalidArgument for open-ended campaign (BRB §3.6), got %v", err)
	}
}

func TestNewVoucher_RejectsPercentageOver100(t *testing.T) {
	id, code, name, desc := validVoucherArgs()
	_, err := entity.NewVoucher(
		id, code, name, desc, 1,
		now, now.Add(24*time.Hour),
		0, 1, 100_000,
		entity.DiscountTypePercentage, 150, 30_000, 0,
		nil, nil, nil,
		true, false, false,
		entity.PromotionTypeFirstRide, now,
	)
	if !sharederrors.IsCode(err, sharederrors.CodeInvalidArgument) {
		t.Fatalf("expected InvalidArgument for >100%% discount, got %v", err)
	}
}

func TestNewVoucher_Valid(t *testing.T) {
	id, code, name, desc := validVoucherArgs()
	v, err := entity.NewVoucher(
		id, code, name, desc, 30,
		now, now.Add(24*time.Hour),
		0, 1, 100_000,
		entity.DiscountTypePercentage, 50, 30_000, 0,
		nil, nil, nil,
		true, false, false,
		entity.PromotionTypeFirstRide, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.VoucherStatusDraft {
		t.Fatalf("expected new voucher to start Draft, got %s", v.Status)
	}
	if v.RemainingBudget != v.Budget {
		t.Fatalf("expected remaining budget to equal budget on creation")
	}
}

func TestVoucher_ReserveAndRelease(t *testing.T) {
	v := mustVoucher(t, 100_000)
	if err := v.Activate(now); err != nil {
		t.Fatalf("activate: %v", err)
	}

	if err := v.Reserve(30_000, now); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if v.RemainingBudget != 70_000 {
		t.Fatalf("expected remaining budget 70000, got %d", v.RemainingBudget)
	}
	if v.UsageCount != 1 {
		t.Fatalf("expected usage count 1, got %d", v.UsageCount)
	}

	if err := v.Release(30_000, now); err != nil {
		t.Fatalf("release: %v", err)
	}
	if v.RemainingBudget != 100_000 {
		t.Fatalf("expected remaining budget reinstated to 100000, got %d", v.RemainingBudget)
	}
	if v.UsageCount != 0 {
		t.Fatalf("expected usage count back to 0, got %d", v.UsageCount)
	}
}

func TestVoucher_ReserveExhaustsAtZeroBudget(t *testing.T) {
	v := mustVoucher(t, 30_000)
	if err := v.Activate(now); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := v.Reserve(30_000, now); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if v.Status != entity.VoucherStatusExhausted {
		t.Fatalf("expected status Exhausted after budget hits 0 (BRB §3.3 Rule 3), got %s", v.Status)
	}
}

func TestVoucher_ReserveRejectsWhenBudgetInsufficient(t *testing.T) {
	v := mustVoucher(t, 10_000)
	if err := v.Activate(now); err != nil {
		t.Fatalf("activate: %v", err)
	}
	err := v.Reserve(30_000, now)
	if !sharederrors.IsCode(err, sharederrors.CodeResourceExhausted) {
		t.Fatalf("expected ResourceExhausted, got %v", err)
	}
}

func TestVoucher_PerUserLimitDefaultsToOne(t *testing.T) {
	v := mustVoucherWithPerUserLimit(t, 0)
	if v.PerUserLimit() != 1 {
		t.Fatalf("expected default per-user limit 1 (BRB §4.6), got %d", v.PerUserLimit())
	}
}

func mustVoucher(t *testing.T, budget int64) *entity.Voucher {
	t.Helper()
	id, code, name, desc := validVoucherArgs()
	v, err := entity.NewVoucher(
		id, code, name, desc, 30,
		now, now.Add(24*time.Hour),
		0, 1, budget,
		entity.DiscountTypePercentage, 50, 30_000, 0,
		nil, nil, nil,
		true, false, false,
		entity.PromotionTypeFirstRide, now,
	)
	if err != nil {
		t.Fatalf("mustVoucher: %v", err)
	}
	return v
}

func mustVoucherWithPerUserLimit(t *testing.T, perUserLimit int64) *entity.Voucher {
	t.Helper()
	id, code, name, desc := validVoucherArgs()
	v, err := entity.NewVoucher(
		id, code, name, desc, 30,
		now, now.Add(24*time.Hour),
		0, perUserLimit, 100_000,
		entity.DiscountTypePercentage, 50, 30_000, 0,
		nil, nil, nil,
		true, false, false,
		entity.PromotionTypeFirstRide, now,
	)
	if err != nil {
		t.Fatalf("mustVoucherWithPerUserLimit: %v", err)
	}
	return v
}

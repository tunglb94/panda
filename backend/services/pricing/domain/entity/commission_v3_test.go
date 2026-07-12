package entity_test

import (
	"testing"

	"github.com/fairride/pricing/domain/entity"
)

func testCommissionConfig() entity.CommissionConfigV3 {
	return entity.CommissionConfigV3{RateByTier: map[entity.CommissionTier]float64{
		entity.CommissionTierBronze:   0.16,
		entity.CommissionTierSilver:   0.15,
		entity.CommissionTierGold:     0.14,
		entity.CommissionTierPlatinum: 0.13,
		entity.CommissionTierDiamond:  0.12,
	}}
}

func TestCommissionConfigV3_Rate_KnownTiers(t *testing.T) {
	cfg := testCommissionConfig()
	cases := map[entity.CommissionTier]float64{
		entity.CommissionTierBronze:   0.16,
		entity.CommissionTierSilver:   0.15,
		entity.CommissionTierGold:     0.14,
		entity.CommissionTierPlatinum: 0.13,
		entity.CommissionTierDiamond:  0.12,
	}
	for tier, want := range cases {
		if got := cfg.Rate(tier); got != want {
			t.Errorf("%s: got %v, want %v", tier, got, want)
		}
	}
}

func TestCommissionConfigV3_Rate_UnknownFailsClosedToBronze(t *testing.T) {
	cfg := testCommissionConfig()
	if got := cfg.Rate(entity.CommissionTier("unknown")); got != 0.16 {
		t.Errorf("unknown tier: got %v, want Bronze rate 0.16", got)
	}
	if got := cfg.Rate(""); got != 0.16 {
		t.Errorf("empty tier: got %v, want Bronze rate 0.16", got)
	}
}

func TestCommissionConfigV3_Rate_MonotonicByTier(t *testing.T) {
	cfg := testCommissionConfig()
	tiers := []entity.CommissionTier{
		entity.CommissionTierBronze, entity.CommissionTierSilver, entity.CommissionTierGold,
		entity.CommissionTierPlatinum, entity.CommissionTierDiamond,
	}
	prev := 1.0
	for _, tier := range tiers {
		rate := cfg.Rate(tier)
		if rate > prev {
			t.Fatalf("commission rate must decrease Bronze->Diamond, but %s (%v) > previous (%v)", tier, rate, prev)
		}
		prev = rate
	}
}

package entity_test

import (
	"testing"

	"github.com/fairride/pricing/domain/entity"
)

func carTiers() []entity.DistanceTier {
	return []entity.DistanceTier{
		{FromKM: 0, ToKM: 2, RatePerKM: 9500},
		{FromKM: 2, ToKM: 5, RatePerKM: 8600},
		{FromKM: 5, ToKM: 10, RatePerKM: 7800},
		{FromKM: 10, ToKM: 20, RatePerKM: 7000},
		{FromKM: 20, ToKM: 40, RatePerKM: 6200},
		{FromKM: 40, ToKM: 60, RatePerKM: 5400},
		{FromKM: 60, ToKM: 0, RatePerKM: 4600}, // open-ended
	}
}

func TestDistanceFareForTiers_Zero(t *testing.T) {
	if got := entity.DistanceFareForTiers(carTiers(), 0); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestDistanceFareForTiers_Negative(t *testing.T) {
	if got := entity.DistanceFareForTiers(carTiers(), -5); got != 0 {
		t.Errorf("got %d, want 0 for negative distance", got)
	}
}

func TestDistanceFareForTiers_WithinFirstTier(t *testing.T) {
	// 1km entirely within tier 1 (0-2km @ 9500/km).
	if got := entity.DistanceFareForTiers(carTiers(), 1); got != 9500 {
		t.Errorf("got %d, want 9500", got)
	}
}

func TestDistanceFareForTiers_ExactTierBoundary(t *testing.T) {
	// Exactly 2km: all of tier 1 (0-2 @ 9500), none of tier 2.
	want := int64(2 * 9500)
	if got := entity.DistanceFareForTiers(carTiers(), 2); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestDistanceFareForTiers_SpansMultipleTiers(t *testing.T) {
	// 25km: 2@9500 + 3@8600 + 5@7800 + 10@7000 + 5@6200
	want := int64(2*9500 + 3*8600 + 5*7800 + 10*7000 + 5*6200)
	if got := entity.DistanceFareForTiers(carTiers(), 25); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestDistanceFareForTiers_OpenEndedLastTier(t *testing.T) {
	// 100km: 2@9500 + 3@8600 + 5@7800 + 10@7000 + 20@6200 + 20@5400 + 40@4600
	want := int64(2*9500 + 3*8600 + 5*7800 + 10*7000 + 20*6200 + 20*5400 + 40*4600)
	if got := entity.DistanceFareForTiers(carTiers(), 100); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestDistanceFareForTiers_MonotonicallyIncreasing(t *testing.T) {
	// A longer trip must never cost less distance-fare than a shorter one —
	// safety invariant regardless of tier shape.
	tiers := carTiers()
	prev := int64(-1)
	for km := 0.0; km <= 120; km += 0.5 {
		got := entity.DistanceFareForTiers(tiers, km)
		if got < prev {
			t.Fatalf("distance fare decreased at %.1fkm: got %d, previous %d", km, got, prev)
		}
		prev = got
	}
}

func TestDistanceFareForTiers_EmptyTiers(t *testing.T) {
	if got := entity.DistanceFareForTiers(nil, 10); got != 0 {
		t.Errorf("got %d, want 0 for nil tiers", got)
	}
}

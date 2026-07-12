package simulation_test

import (
	"testing"

	"github.com/fairride/pricing/simulation"
)

// ─── BƯỚC 7 — safety invariants must hold across every scenario ────────────

func TestAllScenarios_NoSafetyViolations(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	scenarios := simulation.AllScenarios()
	if len(scenarios) < 100 {
		t.Fatalf("expected at least 100 scenarios (BƯỚC 3), got %d", len(scenarios))
	}
	for _, sc := range scenarios {
		fb, err := sim.Simulate(sc.Input)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", sc.Name, err)
			continue
		}
		if issues := simulation.Validate(fb); len(issues) > 0 {
			t.Errorf("%s: safety invariant violated: %v", sc.Name, issues)
		}
	}
}

func TestAllScenarios_UniqueNames(t *testing.T) {
	seen := map[string]bool{}
	for _, sc := range simulation.AllScenarios() {
		if seen[sc.Name] {
			t.Errorf("duplicate scenario name: %s", sc.Name)
		}
		seen[sc.Name] = true
	}
}

// ─── Specific safety-clamp behaviours ───────────────────────────────────────

func TestVoucherExceedingTripValue_IsClampedNotNegative(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	in := simulation.TripInput{
		VehicleType: simulation.VehicleCar,
		DistanceKM:  2,
		DurationMin: 6,
		DriverTier:  simulation.TierBronze,
		Promotion:   &simulation.PromotionInput{Label: "HugeVoucher", DiscountVND: 999_999},
	}
	fb, err := sim.Simulate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.CustomerTotal < 0 {
		t.Errorf("CustomerTotal went negative: %d", fb.CustomerTotal)
	}
	if fb.PromotionApplied > fb.PromotionRequested {
		t.Errorf("PromotionApplied (%d) exceeded PromotionRequested (%d)", fb.PromotionApplied, fb.PromotionRequested)
	}
}

func TestMinimumDriverEarningGuarantee_AlwaysMet(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	in := simulation.TripInput{
		VehicleType: simulation.VehicleMotorcycle,
		DistanceKM:  0.2,
		DurationMin: 1,
		DriverTier:  simulation.TierBronze,
	}
	fb, err := sim.Simulate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.NetDriver < simulation.MinimumDriverEarningVND {
		t.Errorf("NetDriver (%d) below guarantee (%d)", fb.NetDriver, simulation.MinimumDriverEarningVND)
	}
}

func TestSurgeNeverExceedsCap(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	in := simulation.TripInput{
		VehicleType: simulation.VehicleCar,
		DistanceKM:  10,
		DurationMin: 20,
		DriverTier:  simulation.TierBronze,
		DSR:         999, // absurd input on purpose
	}
	fb, err := sim.Simulate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.SurgeMultiplier > simulation.MaxSurgeMultiplier {
		t.Errorf("SurgeMultiplier %v exceeded cap %v", fb.SurgeMultiplier, simulation.MaxSurgeMultiplier)
	}
}

func TestPriceCapNeverExceeded(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	in := simulation.TripInput{
		VehicleType: simulation.VehicleVan,
		DistanceKM:  200, // pathological long trip
		DurationMin: 500,
		DriverTier:  simulation.TierBronze,
		DSR:         5.0,
		IsHoliday:   true,
		Weather:     simulation.WeatherRain,
	}
	fb, err := sim.Simulate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare > simulation.PriceCapVND {
		t.Errorf("RideFare %d exceeded PriceCapVND %d", fb.RideFare, simulation.PriceCapVND)
	}
	if !fb.PriceCapApplied {
		t.Error("expected PriceCapApplied to be true for a pathologically long trip")
	}
}

func TestPeakNeverStacksWithSurge(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	for _, sc := range simulation.AllScenarios() {
		if sc.Name != "SurgeSuppressesPeak_SurgeActive" {
			continue
		}
		fb, err := sim.Simulate(sc.Input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fb.PeakApplied {
			t.Error("PeakApplied was true while surge was active — BRB §2.2.12 violated")
		}
		if fb.SurgeMultiplier <= 1.0 {
			t.Error("expected surge to be active in this scenario")
		}
	}
}

func TestWaitingGraceApplied(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	in := simulation.TripInput{
		VehicleType: simulation.VehicleCar,
		DistanceKM:  5,
		DurationMin: 11,
		WaitingMin:  3, // exactly the grace period — should bill 0
		DriverTier:  simulation.TierBronze,
	}
	fb, err := sim.Simulate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.WaitingFee != 0 {
		t.Errorf("WaitingFee should be 0 within the 3-minute grace period, got %d", fb.WaitingFee)
	}
}

func TestPassengerLevelHasNoPricingEffect(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	base := simulation.TripInput{
		VehicleType: simulation.VehicleCar,
		DistanceKM:  6,
		DurationMin: 13,
		DriverTier:  simulation.TierBronze,
	}
	withoutLevel, err := sim.Simulate(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	withLevel := base
	withLevel.PassengerLevel = "Gold"
	withLevelResult, err := sim.Simulate(withLevel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if withoutLevel.CustomerTotal != withLevelResult.CustomerTotal {
		t.Errorf("PassengerLevel changed CustomerTotal (%d vs %d) — BRB §10.5 has no rider-tier pricing rule yet",
			withoutLevel.CustomerTotal, withLevelResult.CustomerTotal)
	}
}

func TestUnsupportedVehicleType_ReturnsError(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	_, err := sim.Simulate(simulation.TripInput{VehicleType: "helicopter", DistanceKM: 5, DurationMin: 10})
	if err == nil {
		t.Fatal("expected error for unsupported vehicle type")
	}
}

func TestNegativeDistance_ReturnsError(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	_, err := sim.Simulate(simulation.TripInput{VehicleType: simulation.VehicleCar, DistanceKM: -1, DurationMin: 10})
	if err == nil {
		t.Fatal("expected error for negative distance")
	}
}

// ─── BƯỚC 4/5/6 smoke tests — these mainly guard against panics / gross
// structural regressions; the numeric report lives in
// docs/business/PRICING_SIMULATION_REPORT.md, generated by
// cmd/pricing-simulate. ─────────────────────────────────────────────────────

func TestCompareToMarket_ReturnsAllEightCompetitors(t *testing.T) {
	sim := simulation.NewDefaultSimulator()
	fb, err := sim.Simulate(simulation.TripInput{VehicleType: simulation.VehicleCar, DistanceKM: 8, DurationMin: 18, DriverTier: simulation.TierBronze})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rows := simulation.CompareToMarket(fb)
	if len(rows) != 8 {
		t.Errorf("expected 8 competitor rows, got %d", len(rows))
	}
}

func TestOptimize_ReturnsCandidatesSortedByPriority(t *testing.T) {
	scenarios := simulation.AllScenarios()
	scores := simulation.Optimize(scenarios, simulation.DefaultCandidateGrid())
	if len(scores) == 0 {
		t.Fatal("expected at least one candidate score")
	}
	for i := 1; i < len(scores); i++ {
		if scores[i].DriverEarnsMoreCount > scores[i-1].DriverEarnsMoreCount {
			t.Errorf("candidates not sorted by DriverEarnsMoreCount at index %d", i)
		}
	}
}

func TestSensitivity_DoNotPanic(t *testing.T) {
	scenarios := simulation.AllScenarios()
	_ = simulation.RunFuelShock(scenarios, 0.20)
	_ = simulation.RunPromotionShock(scenarios, 1.5)
	_ = simulation.RunCommissionShock(scenarios, 0.16)
	_ = simulation.RunDriverSupplyDoubled(scenarios)
	_ = simulation.RunRiderDemand5x(scenarios)
}

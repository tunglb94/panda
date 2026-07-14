package app_test

import (
	"testing"
	"time"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
)

// ─── Backward compatibility: existing Estimate/CalculateFinal are unchanged ──

func TestFareCalculator_EstimateIsBackwardCompatible(t *testing.T) {
	calc := app.NewFareCalculator(entity.DefaultFareConfig())

	// Same case as TestEstimate_Car_BasicTrip in fare_calculator_test.go —
	// must produce byte-identical output after the Dynamic Pricing Engine
	// refactor, regardless of when the test happens to run.
	fb, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 71500 || fb.Total != 73500 {
		t.Fatalf("Dynamic Pricing Engine refactor changed existing output: RideFare=%d (want 71500), Total=%d (want 73500)", fb.RideFare, fb.Total)
	}
}

func TestFareCalculator_EstimateNeverSurgesEvenAtNight(t *testing.T) {
	// Regression guard: Estimate must stay neutral no matter what wall-clock
	// time it's called at, since every rule ships disabled by default.
	calc := app.NewFareCalculator(entity.DefaultFareConfig())
	fb, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.Total != 73500 {
		t.Fatalf("expected Total 73500 regardless of current wall-clock time, got %d", fb.Total)
	}
}

// ─── EstimateWithContext: the engine actually surges when rules are enabled ──

func TestFareCalculator_EstimateWithContext_NightSurgeApplied(t *testing.T) {
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameNight] = app.RuleConfig{Enabled: true, Priority: 40, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.NightSurchargeMultiplier}
	calc := app.NewFareCalculatorWithPipeline(entity.DefaultFareConfig(), app.NewDefaultPricingPipeline(), configs)

	night := time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC)
	ctx := entity.PricingContext{VehicleType: entity.VehicleTypeCar, RequestTime: night}

	fb, err := calc.EstimateWithContext(entity.VehicleTypeCar, 5.0, 15.0, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Base ride fare (pre-surge) = 71500 (from TestEstimate_Car_BasicTrip).
	// Surged: round(71500 * 1.20) = 85800. Total = 85800 + 2000 booking fee = 87800.
	if fb.RideFare != 85800 {
		t.Fatalf("RideFare: got %d, want 85800 (71500 x 1.20 night surcharge)", fb.RideFare)
	}
	if fb.Total != 87800 {
		t.Fatalf("Total: got %d, want 87800", fb.Total)
	}
}

func TestFareCalculator_EstimateWithContext_AirportFeeAddedBeforeSurge(t *testing.T) {
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameAirport] = app.RuleConfig{Enabled: true, Priority: 5, Weight: 1.0, MinSurge: 0, MaxSurge: float64(entity.AirportFeeVND)}
	configs[app.RuleNameNight] = app.RuleConfig{Enabled: true, Priority: 40, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.NightSurchargeMultiplier}
	calc := app.NewFareCalculatorWithPipeline(entity.DefaultFareConfig(), app.NewDefaultPricingPipeline(), configs)

	night := time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC)
	ctx := entity.PricingContext{VehicleType: entity.VehicleTypeCar, RequestTime: night, IsAirportZone: true}

	fb, err := calc.EstimateWithContext(entity.VehicleTypeCar, 5.0, 15.0, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// BRB §2.13.4: surge applies to (base+distance+time+airport fee).
	// (71,500 + 10,000) x 1.20 = 97,800. Total = 97,800 + 2,000 = 99,800.
	if fb.RideFare != 97_800 {
		t.Fatalf("RideFare: got %d, want 97800", fb.RideFare)
	}
	if fb.Total != 99_800 {
		t.Fatalf("Total: got %d, want 99800", fb.Total)
	}
}

func TestFareCalculator_EstimateWithContext_MinimumFareStillEnforcedAfterSurge(t *testing.T) {
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameNight] = app.RuleConfig{Enabled: true, Priority: 40, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.NightSurchargeMultiplier}
	calc := app.NewFareCalculatorWithPipeline(entity.DefaultFareConfig(), app.NewDefaultPricingPipeline(), configs)

	night := time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC)
	ctx := entity.PricingContext{VehicleType: entity.VehicleTypeCar, RequestTime: night}

	// 0km/0min: base=11250, surged 11250*1.20=13500, still below the 25000 minimum fare.
	fb, err := calc.EstimateWithContext(entity.VehicleTypeCar, 0, 0, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 25000 {
		t.Fatalf("RideFare: got %d, want 25000 (minimum fare still enforced after surge)", fb.RideFare)
	}
}

func TestFareCalculator_CalculateFinalIsBackwardCompatible(t *testing.T) {
	calc := app.NewFareCalculator(entity.DefaultFareConfig())
	fb, err := calc.CalculateFinal(entity.VehicleTypeVan, 10.0, 20.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.RideFare != 78000 || fb.Total != 80000 {
		t.Fatalf("Dynamic Pricing Engine refactor changed existing output: RideFare=%d (want 78000), Total=%d (want 80000)", fb.RideFare, fb.Total)
	}
	if !fb.IsFinal {
		t.Fatal("IsFinal should be true for CalculateFinal")
	}
}

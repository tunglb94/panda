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
	if fb.RideFare != 36000 || fb.Total != 38000 {
		t.Fatalf("Dynamic Pricing Engine refactor changed existing output: RideFare=%d (want 36000), Total=%d (want 38000)", fb.RideFare, fb.Total)
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
	if fb.Total != 38000 {
		t.Fatalf("expected Total 38000 regardless of current wall-clock time, got %d", fb.Total)
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
	// Base ride fare (pre-surge) = 36000 (from TestEstimate_Car_BasicTrip).
	// Surged: round(36000 * 1.20) = 43200. Total = 43200 + 2000 booking fee = 45200.
	if fb.RideFare != 43200 {
		t.Fatalf("RideFare: got %d, want 43200 (36000 x 1.20 night surcharge)", fb.RideFare)
	}
	if fb.Total != 45200 {
		t.Fatalf("Total: got %d, want 45200", fb.Total)
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
	// (36,000 + 10,000) x 1.20 = 55,200. Total = 55,200 + 2,000 = 57,200.
	if fb.RideFare != 55_200 {
		t.Fatalf("RideFare: got %d, want 55200", fb.RideFare)
	}
	if fb.Total != 57_200 {
		t.Fatalf("Total: got %d, want 57200", fb.Total)
	}
}

func TestFareCalculator_EstimateWithContext_MinimumFareStillEnforcedAfterSurge(t *testing.T) {
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameNight] = app.RuleConfig{Enabled: true, Priority: 40, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.NightSurchargeMultiplier}
	calc := app.NewFareCalculatorWithPipeline(entity.DefaultFareConfig(), app.NewDefaultPricingPipeline(), configs)

	night := time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC)
	ctx := entity.PricingContext{VehicleType: entity.VehicleTypeCar, RequestTime: night}

	// 0km/0min: base=10000, surged 10000*1.20=12000, still below the 25000 minimum fare.
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

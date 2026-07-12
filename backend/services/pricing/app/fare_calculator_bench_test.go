package app_test

import (
	"testing"
	"time"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
)

// BenchmarkEstimate_AllRulesDisabled measures the production hot path
// (Estimate, as called by the gRPC handler) — every rule disabled, the
// pipeline should cost almost nothing since PricingEvaluator short-circuits
// before calling any rule's Evaluate.
func BenchmarkEstimate_AllRulesDisabled(b *testing.B) {
	calc := app.NewFareCalculator(entity.DefaultFareConfig())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := calc.Estimate(entity.VehicleTypeCar, 5.0, 15.0); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEstimateWithContext_AllRulesEnabled measures the worst case: every
// rule enabled and evaluated (the Dynamic Pricing Engine actually running
// its full rule set), to quantify the refactor's overhead when the engine is
// fully active.
func BenchmarkEstimateWithContext_AllRulesEnabled(b *testing.B) {
	configs := app.DefaultRuleConfigs()
	for name, cfg := range configs {
		cfg.Enabled = true
		configs[name] = cfg
	}
	calc := app.NewFareCalculatorWithPipeline(entity.DefaultFareConfig(), app.NewDefaultPricingPipeline(), configs)

	ctx := entity.PricingContext{
		VehicleType:      entity.VehicleTypeCar,
		RequestTime:      time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC),
		ActiveRequests:   18,
		AvailableDrivers: 10,
		IsRainActive:     true,
		IsHoliday:        true,
		IsAirportZone:    true,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := calc.EstimateWithContext(entity.VehicleTypeCar, 5.0, 15.0, ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPricingPipeline_Evaluate isolates the pipeline's own cost from
// fare arithmetic.
func BenchmarkPricingPipeline_Evaluate(b *testing.B) {
	pipeline := app.NewDefaultPricingPipeline()
	configs := app.DefaultRuleConfigs()
	for name, cfg := range configs {
		cfg.Enabled = true
		configs[name] = cfg
	}
	ctx := entity.PricingContext{
		RequestTime:      time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC),
		ActiveRequests:   18,
		AvailableDrivers: 10,
		IsRainActive:     true,
		IsHoliday:        true,
		IsAirportZone:    true,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Evaluate(ctx, configs)
	}
}

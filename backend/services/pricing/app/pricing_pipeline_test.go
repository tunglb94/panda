package app_test

import (
	"testing"
	"time"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
)

func enabledCfg(priority int, weight, minSurge, maxSurge float64) app.RuleConfig {
	return app.RuleConfig{Enabled: true, Priority: priority, Weight: weight, MinSurge: minSurge, MaxSurge: maxSurge}
}

func TestPricingPipeline_AllDisabled_IsNeutral(t *testing.T) {
	pipeline := app.NewDefaultPricingPipeline()
	ctx := entity.PricingContext{
		RequestTime: nightTime(), // would trigger Night if enabled
		ActiveRequests: 100, AvailableDrivers: 1, // would trigger Demand Surge if enabled
		IsRainActive: true, IsHoliday: true, IsAirportZone: true,
	}

	result := pipeline.Evaluate(ctx, app.DefaultRuleConfigs())
	if !result.Neutral() {
		t.Fatalf("expected neutral result with every rule disabled, got multiplier=%v flat=%v", result.FinalMultiplier, result.FlatSurcharge)
	}
	if len(result.AppliedRules) != 0 {
		t.Fatalf("expected zero applied rules, got %+v", result.AppliedRules)
	}
}

func TestPricingPipeline_NightOnly(t *testing.T) {
	pipeline := app.NewDefaultPricingPipeline()
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameNight] = enabledCfg(40, 1.0, 1.0, entity.NightSurchargeMultiplier)

	result := pipeline.Evaluate(entity.PricingContext{RequestTime: nightTime()}, configs)
	if result.FinalMultiplier != entity.NightSurchargeMultiplier {
		t.Fatalf("expected multiplier %v, got %v", entity.NightSurchargeMultiplier, result.FinalMultiplier)
	}
}

func TestPricingPipeline_NightHolidayStack_UnderCap(t *testing.T) {
	pipeline := app.NewDefaultPricingPipeline()
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameNight] = enabledCfg(40, 1.0, 1.0, entity.NightSurchargeMultiplier)
	configs[app.RuleNameHoliday] = enabledCfg(50, 1.0, 1.0, entity.HolidaySurchargeMultiplier)

	result := pipeline.Evaluate(entity.PricingContext{RequestTime: nightTime(), IsHoliday: true}, configs)
	want := entity.NightSurchargeMultiplier * entity.HolidaySurchargeMultiplier // 1.20 * 1.15 = 1.38, under the 1.50 cap
	if result.FinalMultiplier != want {
		t.Fatalf("expected combined multiplier %v (under BRB §2.2.11 cap %v), got %v", want, entity.MaxCombinedNightHolidayMultiplier, result.FinalMultiplier)
	}
	if result.FinalMultiplier > entity.MaxCombinedNightHolidayMultiplier {
		t.Fatalf("BRB §2.2.11 combined cap violated")
	}
}

func TestPricingPipeline_NightHolidayCap_Enforced(t *testing.T) {
	// Deliberately misconfigured weights (2x) to push the raw product above
	// BRB's x1.50 Night+Holiday cap, proving the pipeline's cap is a real
	// safety net and not just decorative — this is not a scenario BRB's own
	// default multipliers alone can reach (1.20 x 1.15 = 1.38).
	pipeline := app.NewDefaultPricingPipeline()
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameNight] = enabledCfg(40, 2.0, 1.0, 3.0)
	configs[app.RuleNameHoliday] = enabledCfg(50, 2.0, 1.0, 3.0)

	result := pipeline.Evaluate(entity.PricingContext{RequestTime: nightTime(), IsHoliday: true}, configs)
	if result.FinalMultiplier != entity.MaxCombinedNightHolidayMultiplier {
		t.Fatalf("expected pipeline to cap combined Night+Holiday at %v, got %v", entity.MaxCombinedNightHolidayMultiplier, result.FinalMultiplier)
	}
}

func TestPricingPipeline_NightHolidayRainCap_Enforced(t *testing.T) {
	pipeline := app.NewDefaultPricingPipeline()
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameNight] = enabledCfg(40, 2.0, 1.0, 3.0)
	configs[app.RuleNameHoliday] = enabledCfg(50, 2.0, 1.0, 3.0)
	configs[app.RuleNameRain] = enabledCfg(60, 2.0, 1.0, 3.0)

	result := pipeline.Evaluate(entity.PricingContext{RequestTime: nightTime(), IsHoliday: true, IsRainActive: true}, configs)
	if result.FinalMultiplier != entity.MaxCombinedNightHolidayRainMultiplier {
		t.Fatalf("expected pipeline to cap combined Night+Holiday+Rain at %v (BRB §2.2.13), got %v", entity.MaxCombinedNightHolidayRainMultiplier, result.FinalMultiplier)
	}
}

func TestPricingPipeline_DemandSurgeSupersedesPeakHour(t *testing.T) {
	pipeline := app.NewDefaultPricingPipeline()
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameDemandSurge] = enabledCfg(10, 1.0, 1.0, entity.MaxDemandSurgeMultiplier)
	configs[app.RuleNamePeakHour] = enabledCfg(30, 1.0, 1.0, entity.PeakHourSurchargeMultiplier)

	// Monday 08:00 is within the default morning peak window AND demand is
	// high enough to trigger Demand Surge (DSR = 20/10 = 2.0 -> x1.6).
	monday0800 := time.Date(2026, 7, 6, 8, 0, 0, 0, time.UTC)
	ctx := entity.PricingContext{RequestTime: monday0800, ActiveRequests: 20, AvailableDrivers: 10}

	result := pipeline.Evaluate(ctx, configs)
	if result.FinalMultiplier != 1.6 {
		t.Fatalf("expected demand surge multiplier 1.6 alone (peak superseded per BRB §2.2.12), got %v", result.FinalMultiplier)
	}
	for _, r := range result.AppliedRules {
		if r.RuleName == app.RuleNamePeakHour {
			t.Fatalf("Peak Hour must not appear in AppliedRules when Dynamic Surge is active")
		}
	}
	foundSkippedPeak := false
	for _, r := range result.SkippedRules {
		if r.RuleName == app.RuleNamePeakHour {
			foundSkippedPeak = true
		}
	}
	if !foundSkippedPeak {
		t.Fatalf("expected Peak Hour to appear in SkippedRules with a superseded reason")
	}
}

func TestPricingPipeline_AirportFlatFee(t *testing.T) {
	pipeline := app.NewDefaultPricingPipeline()
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameAirport] = app.RuleConfig{Enabled: true, Priority: 5, Weight: 1.0, MinSurge: 0, MaxSurge: float64(entity.AirportFeeVND)}

	result := pipeline.Evaluate(entity.PricingContext{IsAirportZone: true}, configs)
	if result.FlatSurcharge != entity.AirportFeeVND {
		t.Fatalf("expected flat surcharge %d, got %d", entity.AirportFeeVND, result.FlatSurcharge)
	}
	if result.FinalMultiplier != 1.0 {
		t.Fatalf("airport fee must not affect the multiplier, got %v", result.FinalMultiplier)
	}
}

func TestPricingPipeline_TODORulesNeverAffectResult(t *testing.T) {
	pipeline := app.NewDefaultPricingPipeline()
	configs := app.DefaultRuleConfigs()
	configs[app.RuleNameSupplySurge] = enabledCfg(20, 1.0, 1.0, entity.MaxDemandSurgeMultiplier)
	configs[app.RuleNameTraffic] = enabledCfg(70, 1.0, 1.0, 2.0)
	configs[app.RuleNameSpecialEvent] = enabledCfg(80, 1.0, 1.0, 2.0)

	ctx := entity.PricingContext{
		ActiveRequests: 100, AvailableDrivers: 1,
		TrafficLevel: entity.TrafficLevelHeavy, IsSpecialEvent: true,
	}
	result := pipeline.Evaluate(ctx, configs)
	if !result.Neutral() {
		t.Fatalf("enabling TODO-stub rules must not change the result even when enabled, got multiplier=%v flat=%v", result.FinalMultiplier, result.FlatSurcharge)
	}
}

package app_test

import (
	"testing"
	"time"

	"github.com/fairride/pricing/app"
	"github.com/fairride/pricing/domain/entity"
)

func nightTime() time.Time {
	return time.Date(2026, 7, 6, 23, 0, 0, 0, time.UTC)
}

func TestPricingEvaluator_DisabledRuleNeverRuns(t *testing.T) {
	eval := app.NewPricingEvaluator()
	rule := app.NewRainSurchargeRule()
	cfg := app.RuleConfig{Enabled: false}

	out := eval.Evaluate(rule, cfg, entity.PricingContext{IsRainActive: true})
	if out.Applied {
		t.Fatalf("disabled rule must never apply regardless of context, got %+v", out)
	}
}

func TestPricingEvaluator_WeightScalesMultiplier(t *testing.T) {
	eval := app.NewPricingEvaluator()
	rule := app.NewNightSurchargeRule(entity.NightWindowStartHour, entity.NightWindowEndHour)
	// Night's raw multiplier is 1.20. Weight 0.5 should halve the *excess*
	// over 1.0: 1 + (1.20-1)*0.5 = 1.10.
	cfg := app.RuleConfig{Enabled: true, Weight: 0.5, MinSurge: 1.0, MaxSurge: 2.0}

	out := eval.Evaluate(rule, cfg, entity.PricingContext{RequestTime: nightTime()})
	if !out.Applied {
		t.Fatalf("expected rule to apply, got %+v", out)
	}
	if out.Multiplier != 1.10 {
		t.Fatalf("expected weighted multiplier 1.10, got %v", out.Multiplier)
	}
}

func TestPricingEvaluator_MaxSurgeClamps(t *testing.T) {
	eval := app.NewPricingEvaluator()
	rule := app.NewNightSurchargeRule(entity.NightWindowStartHour, entity.NightWindowEndHour)
	// Operator caps this rule at 1.05, tighter than BRB's 1.20 default.
	cfg := app.RuleConfig{Enabled: true, Weight: 1.0, MinSurge: 1.0, MaxSurge: 1.05}

	out := eval.Evaluate(rule, cfg, entity.PricingContext{RequestTime: nightTime()})
	if out.Multiplier != 1.05 {
		t.Fatalf("expected clamped multiplier 1.05, got %v", out.Multiplier)
	}
}

func TestPricingEvaluator_ZeroWeightNeutralizesRule(t *testing.T) {
	eval := app.NewPricingEvaluator()
	rule := app.NewNightSurchargeRule(entity.NightWindowStartHour, entity.NightWindowEndHour)
	cfg := app.RuleConfig{Enabled: true, Weight: 0, MinSurge: 1.0, MaxSurge: 2.0}

	// Weight left at its Go zero-value should be treated as 1.0 (full
	// effect), matching DefaultRuleConfigs' convention — a caller who forgot
	// to set Weight should not accidentally disable every rule they enabled.
	out := eval.Evaluate(rule, cfg, entity.PricingContext{RequestTime: nightTime()})
	if out.Multiplier != entity.NightSurchargeMultiplier {
		t.Fatalf("expected zero-value Weight to default to 1.0 (full effect %v), got %v", entity.NightSurchargeMultiplier, out.Multiplier)
	}
}

func TestPricingEvaluator_FlatFeeWeightAndClamp(t *testing.T) {
	eval := app.NewPricingEvaluator()
	rule := app.NewAirportFeeRule()
	cfg := app.RuleConfig{Enabled: true, Weight: 0.5, MinSurge: 0, MaxSurge: float64(entity.AirportFeeVND)}

	out := eval.Evaluate(rule, cfg, entity.PricingContext{IsAirportZone: true})
	if out.FlatAmount != entity.AirportFeeVND/2 {
		t.Fatalf("expected half airport fee %d, got %d", entity.AirportFeeVND/2, out.FlatAmount)
	}
}

package app

import "github.com/fairride/pricing/domain/entity"

// PricingEvaluator runs a single PricingRule against a PricingContext and
// applies its RuleConfig (Enabled/Weight/MinSurge/MaxSurge) uniformly — the
// one place operator-facing configuration turns a rule's raw BRB-sourced
// output into the value actually used. Rules themselves never see RuleConfig
// (see PricingRule.Evaluate's signature), which keeps rule formulas and
// operator tuning fully decoupled.
type PricingEvaluator struct{}

func NewPricingEvaluator() *PricingEvaluator {
	return &PricingEvaluator{}
}

// Evaluate returns rule's contribution for ctx, after applying cfg. If
// !cfg.Enabled, the rule's Evaluate method is not even called — a disabled
// rule can never affect PricingResult, by construction.
func (e *PricingEvaluator) Evaluate(rule PricingRule, cfg RuleConfig, ctx entity.PricingContext) entity.RuleOutcome {
	if !cfg.Enabled {
		return entity.RuleOutcome{
			RuleName: rule.Name(), Category: rule.Category(), Applied: false,
			Reason: "rule disabled by configuration",
		}
	}

	raw := rule.Evaluate(ctx)
	if !raw.Applied {
		return raw
	}

	weight := cfg.Weight
	if weight == 0 {
		weight = 1.0
	}

	switch rule.Category() {
	case entity.CategoryFlatFee:
		adjusted := raw
		adjusted.FlatAmount = clampInt64(int64(float64(raw.FlatAmount)*weight), int64(cfg.MinSurge), int64(cfg.MaxSurge))
		return adjusted

	case entity.CategoryStaticSurcharge, entity.CategoryDynamicSurge:
		weighted := 1 + (raw.Multiplier-1)*weight
		adjusted := raw
		adjusted.Multiplier = clampFloat(weighted, cfg.MinSurge, cfg.MaxSurge)
		if adjusted.Multiplier <= 1.0 {
			adjusted.Applied = false
			adjusted.Reason = raw.Reason + " (weighted/clamped to no effect)"
		}
		return adjusted

	default: // entity.CategoryNotDefined
		return raw
	}
}

func clampFloat(v, lo, hi float64) float64 {
	if hi > 0 && v > hi {
		v = hi
	}
	if lo > 0 && v < lo {
		v = lo
	}
	return v
}

func clampInt64(v, lo, hi int64) int64 {
	if hi > 0 && v > hi {
		v = hi
	}
	if v < lo {
		v = lo
	}
	return v
}

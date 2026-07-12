package app

import (
	"sort"

	"github.com/fairride/pricing/domain/entity"
)

// PricingPipeline orchestrates every registered PricingRule into one
// entity.PricingResult. It is the ONLY place that encodes how BRB's surge
// rules interact (stacking, mutual exclusion, combined caps) — individual
// rules never know about each other. This keeps FareCalculator, and every
// rule, free of the "if else dài" the sprint brief prohibits: each rule is a
// pure function of (context) -> (raw contribution), and this pipeline is a
// small, fixed set of table lookups and BRB-cited combination formulas, not
// a growing branch tree.
//
// BRB interaction rules encoded here:
//   - §2.2.12: Dynamic Surge supersedes Peak Hour Surcharge (mutually exclusive)
//   - §2.2.11: Night x Holiday combined capped at x1.50
//   - §2.2.13: Night x Holiday x Rain combined (when Rain participates) capped at x1.60
//   - §2.13.3: Dynamic Surge alone hard-capped at x2.0, no exception
//   - §2.13.4: surge multiplier applies to (Base+Distance+Time+Airport Fee),
//     never to Booking Fee/Toll/Waiting Fee — enforced by the caller
//     (fare_calculator.go), not here; this pipeline only produces the
//     multiplier and flat surcharge.
type PricingPipeline struct {
	rules     []PricingRule
	evaluator *PricingEvaluator
}

// NewPricingPipeline builds a pipeline from an explicit rule set — used by
// tests and by callers who want a subset of rules.
func NewPricingPipeline(rules []PricingRule) *PricingPipeline {
	return &PricingPipeline{rules: rules, evaluator: NewPricingEvaluator()}
}

// NewDefaultPricingPipeline wires all 9 requested surge types, using BRB's
// stated default structural configuration (DSR tiers, peak-hour windows,
// night window). Operators may build a different pipeline (e.g. custom
// per-city peak windows) via NewPricingPipeline.
func NewDefaultPricingPipeline() *PricingPipeline {
	return NewPricingPipeline([]PricingRule{
		NewDemandSurgeRule(entity.DefaultDSRTiers()),
		NewSupplySurgeRule(),
		NewPeakHourRule(entity.DefaultPeakHourWindows()),
		NewNightSurchargeRule(entity.NightWindowStartHour, entity.NightWindowEndHour),
		NewHolidaySurchargeRule(),
		NewRainSurchargeRule(),
		NewAirportFeeRule(),
		NewTrafficSurgeRule(),
		NewSpecialEventRule(),
	})
}

// Evaluate runs every rule against ctx (governed by configs) and combines
// the outcomes into one PricingResult.
func (p *PricingPipeline) Evaluate(ctx entity.PricingContext, configs RuleConfigMap) entity.PricingResult {
	type ordered struct {
		outcome  entity.RuleOutcome
		priority int
	}

	results := make([]ordered, 0, len(p.rules))
	for _, rule := range p.rules {
		cfg := configs.Get(rule.Name())
		results = append(results, ordered{outcome: p.evaluator.Evaluate(rule, cfg, ctx), priority: cfg.Priority})
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].priority < results[j].priority })

	byName := make(map[string]entity.RuleOutcome, len(results))
	for _, r := range results {
		byName[r.outcome.RuleName] = r.outcome
	}

	// ─── Dynamic Surge: take the strongest applied dynamic-surge rule ───────
	dynamicApplied := false
	dynamicMultiplier := 1.0
	dynamicLabel := ""
	for _, r := range results {
		if r.outcome.Category == entity.CategoryDynamicSurge && r.outcome.Applied {
			dynamicApplied = true
			if r.outcome.Multiplier > dynamicMultiplier {
				dynamicMultiplier = r.outcome.Multiplier
				dynamicLabel = r.outcome.Label
			}
		}
	}
	if dynamicMultiplier > entity.MaxDemandSurgeMultiplier {
		dynamicMultiplier = entity.MaxDemandSurgeMultiplier // BRB §2.13.3, no exception
	}

	// ─── Static surcharges: Night x Holiday x Rain, BRB-capped, x Peak Hour ──
	nightOut, hasNight := byName[RuleNameNight]
	holidayOut, hasHoliday := byName[RuleNameHoliday]
	rainOut, hasRain := byName[RuleNameRain]
	peakOut, hasPeak := byName[RuleNamePeakHour]

	nightActive := hasNight && nightOut.Applied
	holidayActive := hasHoliday && holidayOut.Applied
	rainActive := hasRain && rainOut.Applied
	// BRB §2.2.12: Dynamic Surge supersedes Peak Hour — Peak never
	// contributes while Dynamic Surge is active.
	peakActive := hasPeak && peakOut.Applied && !dynamicApplied

	nightHolidayRain := 1.0
	if nightActive {
		nightHolidayRain *= nightOut.Multiplier
	}
	if holidayActive {
		nightHolidayRain *= holidayOut.Multiplier
	}
	if rainActive {
		nightHolidayRain *= rainOut.Multiplier
	}

	var combinedCap float64
	switch {
	case rainActive && (nightActive || holidayActive):
		combinedCap = entity.MaxCombinedNightHolidayRainMultiplier // BRB §2.2.13
	case nightActive && holidayActive:
		combinedCap = entity.MaxCombinedNightHolidayMultiplier // BRB §2.2.11
	}
	if combinedCap > 0 && nightHolidayRain > combinedCap {
		nightHolidayRain = combinedCap
	}

	staticCombined := nightHolidayRain
	if peakActive {
		staticCombined *= peakOut.Multiplier
	}

	// ─── Combine static x dynamic (BRB §2.13.4 base) ─────────────────────────
	finalMultiplier := staticCombined
	label := ""
	if dynamicApplied {
		finalMultiplier *= dynamicMultiplier
		label = dynamicLabel
	} else if peakActive {
		label = peakOut.Label
	}

	// ─── Flat fees (Airport) ──────────────────────────────────────────────
	var flatSurcharge int64
	for _, r := range results {
		if r.outcome.Category == entity.CategoryFlatFee && r.outcome.Applied {
			flatSurcharge += r.outcome.FlatAmount
		}
	}

	// ─── Build transparency lists, reflecting the Peak-Hour exclusion ───────
	applied := make([]entity.RuleOutcome, 0, len(results))
	skipped := make([]entity.RuleOutcome, 0, len(results))
	for _, r := range results {
		out := r.outcome
		if out.RuleName == RuleNamePeakHour && out.Applied && dynamicApplied {
			out.Applied = false
			out.Reason = "superseded by dynamic surge (BRB §2.2.12: Dynamic Surge supersedes Peak Hour Surcharge)"
		}
		if out.Applied {
			applied = append(applied, out)
		} else {
			skipped = append(skipped, out)
		}
	}

	return entity.PricingResult{
		FinalMultiplier: finalMultiplier,
		FlatSurcharge:   flatSurcharge,
		Label:           label,
		AppliedRules:    applied,
		SkippedRules:    skipped,
	}
}

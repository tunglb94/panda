package app

import (
	"sort"
	"time"

	"github.com/fairride/pricing/domain/entity"
)

// This file implements the 6 surge types with a real, BRB-sourced formula:
// Demand Surge, Peak Hour, Night, Holiday, Rain, Airport.
//
// Each rule's STRUCTURAL configuration (which time windows, which DSR tiers)
// is injected at construction time, defaulting to BRB's stated defaults via
// entity.Default*() — never an inline literal inside Evaluate. Each rule's
// OPERATIONAL configuration (enabled/priority/weight/min/max) lives in
// RuleConfig and is applied uniformly by PricingEvaluator, not by the rule
// itself.

// ─── Demand Surge — BRB §2.13.2 ────────────────────────────────────────────

type DemandSurgeRule struct {
	tiers []entity.DSRTier // sorted ascending by MinDSR
}

// NewDemandSurgeRule builds a rule from a DSR-tier table. Pass
// entity.DefaultDSRTiers() for BRB's exact §2.13.2 table.
func NewDemandSurgeRule(tiers []entity.DSRTier) *DemandSurgeRule {
	sorted := append([]entity.DSRTier(nil), tiers...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].MinDSR < sorted[j].MinDSR })
	return &DemandSurgeRule{tiers: sorted}
}

func (r *DemandSurgeRule) Name() string                  { return RuleNameDemandSurge }
func (r *DemandSurgeRule) Category() entity.RuleCategory { return entity.CategoryDynamicSurge }

func (r *DemandSurgeRule) Evaluate(ctx entity.PricingContext) entity.RuleOutcome {
	dsr := demandSupplyRatio(ctx.ActiveRequests, ctx.AvailableDrivers)
	tier := matchDSRTier(r.tiers, dsr)

	multiplier := tier.Multiplier
	// BRB §2.13.3: "Surge multiplier cannot exceed x2.0. No exception." This
	// hard ceiling is enforced here regardless of the tier table or any
	// operator-configured RuleConfig.MaxSurge, which may only tighten it.
	if multiplier > entity.MaxDemandSurgeMultiplier {
		multiplier = entity.MaxDemandSurgeMultiplier
	}

	if multiplier <= 1.0 {
		return entity.RuleOutcome{
			RuleName: r.Name(), Category: r.Category(), Applied: false,
			Multiplier: 1.0, Reason: "demand-supply ratio below surge threshold (DSR < 1.2)",
		}
	}
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: true,
		Multiplier: multiplier, Label: tier.Label,
		Reason: "BRB §2.13.2 demand-supply ratio surge",
	}
}

func demandSupplyRatio(activeRequests, availableDrivers int) float64 {
	if activeRequests <= 0 {
		return 0
	}
	if availableDrivers <= 0 {
		// No drivers available with active demand: treat as the worst-case
		// ratio so the tier lookup lands on the maximum tier, rather than
		// dividing by zero.
		return 1e9
	}
	return float64(activeRequests) / float64(availableDrivers)
}

// matchDSRTier returns the highest tier whose MinDSR <= dsr. tiers must be
// sorted ascending by MinDSR (NewDemandSurgeRule guarantees this). A lookup
// table walk, not an if/else chain.
func matchDSRTier(tiers []entity.DSRTier, dsr float64) entity.DSRTier {
	best := entity.DSRTier{MinDSR: 0, Multiplier: 1.0, Label: "Normal pricing"}
	for _, t := range tiers {
		if dsr >= t.MinDSR {
			best = t
		}
	}
	return best
}

// ─── Peak Hour — BRB §2.2.12 ────────────────────────────────────────────────

type PeakHourRule struct {
	windows []entity.TimeWindow
}

// NewPeakHourRule builds a rule from a set of time windows. Pass
// entity.DefaultPeakHourWindows() for BRB's stated default (Mon-Fri
// 07:00-09:00 and 17:00-20:00); operators may pass a per-city override.
func NewPeakHourRule(windows []entity.TimeWindow) *PeakHourRule {
	return &PeakHourRule{windows: windows}
}

func (r *PeakHourRule) Name() string                  { return RuleNamePeakHour }
func (r *PeakHourRule) Category() entity.RuleCategory { return entity.CategoryStaticSurcharge }

func (r *PeakHourRule) Evaluate(ctx entity.PricingContext) entity.RuleOutcome {
	if !inAnyWindow(r.windows, ctx.RequestTime) {
		return entity.RuleOutcome{
			RuleName: r.Name(), Category: r.Category(), Applied: false,
			Multiplier: 1.0, Reason: "not within a configured peak-hour window",
		}
	}
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: true,
		Multiplier: entity.PeakHourSurchargeMultiplier, Label: "Peak hour",
		Reason: "BRB §2.2.12 peak-hour surcharge",
	}
}

func inAnyWindow(windows []entity.TimeWindow, t time.Time) bool {
	weekday := int(t.Weekday())
	hour := t.Hour()
	for _, w := range windows {
		if !containsWeekday(w.Weekdays, weekday) {
			continue
		}
		if hour >= w.StartHour && hour < w.EndHour {
			return true
		}
	}
	return false
}

func containsWeekday(weekdays []int, day int) bool {
	for _, d := range weekdays {
		if d == day {
			return true
		}
	}
	return false
}

// ─── Night — BRB §2.2.10 ────────────────────────────────────────────────────

type NightSurchargeRule struct {
	startHour int // inclusive, e.g. 22
	endHour   int // exclusive, e.g. 5 (wraps past midnight)
}

// NewNightSurchargeRule builds a rule from a night window. Pass
// entity.NightWindowStartHour, entity.NightWindowEndHour for BRB's stated
// default (22:00-05:00).
func NewNightSurchargeRule(startHour, endHour int) *NightSurchargeRule {
	return &NightSurchargeRule{startHour: startHour, endHour: endHour}
}

func (r *NightSurchargeRule) Name() string                  { return RuleNameNight }
func (r *NightSurchargeRule) Category() entity.RuleCategory { return entity.CategoryStaticSurcharge }

func (r *NightSurchargeRule) Evaluate(ctx entity.PricingContext) entity.RuleOutcome {
	hour := ctx.RequestTime.Hour()
	isNight := hour >= r.startHour || hour < r.endHour
	if !isNight {
		return entity.RuleOutcome{
			RuleName: r.Name(), Category: r.Category(), Applied: false,
			Multiplier: 1.0, Reason: "request time is outside the night window",
		}
	}
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: true,
		Multiplier: entity.NightSurchargeMultiplier, Label: "Night",
		Reason: "BRB §2.2.10 night surcharge (22:00-05:00)",
	}
}

// ─── Holiday — BRB §2.2.11 ──────────────────────────────────────────────────

type HolidaySurchargeRule struct{}

func NewHolidaySurchargeRule() *HolidaySurchargeRule { return &HolidaySurchargeRule{} }

func (r *HolidaySurchargeRule) Name() string                  { return RuleNameHoliday }
func (r *HolidaySurchargeRule) Category() entity.RuleCategory { return entity.CategoryStaticSurcharge }

func (r *HolidaySurchargeRule) Evaluate(ctx entity.PricingContext) entity.RuleOutcome {
	if !ctx.IsHoliday {
		return entity.RuleOutcome{
			RuleName: r.Name(), Category: r.Category(), Applied: false,
			Multiplier: 1.0, Reason: "today is not a configured holiday",
		}
	}
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: true,
		Multiplier: entity.HolidaySurchargeMultiplier, Label: "Holiday",
		Reason: "BRB §2.2.11 holiday surcharge",
	}
}

// ─── Rain — BRB §2.2.13 ─────────────────────────────────────────────────────

type RainSurchargeRule struct{}

func NewRainSurchargeRule() *RainSurchargeRule { return &RainSurchargeRule{} }

func (r *RainSurchargeRule) Name() string                  { return RuleNameRain }
func (r *RainSurchargeRule) Category() entity.RuleCategory { return entity.CategoryStaticSurcharge }

func (r *RainSurchargeRule) Evaluate(ctx entity.PricingContext) entity.RuleOutcome {
	if !ctx.IsRainActive {
		return entity.RuleOutcome{
			RuleName: r.Name(), Category: r.Category(), Applied: false,
			Multiplier: 1.0, Reason: "rain surcharge not currently active",
		}
	}
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: true,
		Multiplier: entity.RainSurchargeMultiplier, Label: "Rain demand",
		Reason: "BRB §2.2.13 rain surcharge",
	}
}

// ─── Airport — BRB §2.2.7 ───────────────────────────────────────────────────

type AirportFeeRule struct{}

func NewAirportFeeRule() *AirportFeeRule { return &AirportFeeRule{} }

func (r *AirportFeeRule) Name() string                  { return RuleNameAirport }
func (r *AirportFeeRule) Category() entity.RuleCategory { return entity.CategoryFlatFee }

func (r *AirportFeeRule) Evaluate(ctx entity.PricingContext) entity.RuleOutcome {
	if !ctx.IsAirportZone {
		return entity.RuleOutcome{
			RuleName: r.Name(), Category: r.Category(), Applied: false,
			Reason: "pickup/dropoff is not an airport zone",
		}
	}
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: true,
		FlatAmount: entity.AirportFeeVND,
		Reason:     "BRB §2.2.7 airport fee (charged once per trip)",
	}
}

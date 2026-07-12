package entity

import "time"

// TrafficLevel is a coarse signal for road congestion. No BRB surge rule
// consumes this today (BRB captures traffic through Time Fare, §2.2.3 — a
// slow trip already produces more chargeable minutes). Kept on PricingContext
// only so a future, BRB-approved Traffic rule has somewhere to read from.
type TrafficLevel string

const (
	TrafficLevelNormal TrafficLevel = "normal"
	TrafficLevelHeavy  TrafficLevel = "heavy"
)

// PricingContext carries every signal a PricingRule might need to decide
// whether, and how much, it should contribute to the surge multiplier for
// one fare calculation. Fields default to their neutral zero value (no
// surge signal present) — see NeutralContext.
//
// Fields fall into two groups:
//   - Derivable from RequestTime alone (Night, Peak Hour): the rule computes
//     activation itself from RequestTime + its own configurable time windows.
//     No boolean field exists for these — see NightSurchargeRule, PeakHourRule.
//   - External signals this service does not own (Rain, Holiday, live demand/
//     supply counts): callers populate these from Ops/weather/dispatch data.
//     This engine never invents these values.
type PricingContext struct {
	VehicleType VehicleType
	RequestTime time.Time

	// Demand Surge (BRB §2.13.2). DSR = ActiveRequests / AvailableDrivers.
	ActiveRequests   int
	AvailableDrivers int

	// Supply Surge — no BRB formula exists separate from the unified DSR
	// above (see app/rules_todo.go SupplySurgeRule). Field kept for a future,
	// BRB-approved supply-specific signal.
	AvailableDriversTrend float64

	// Airport (BRB §2.2.7).
	IsAirportZone bool

	// AirportLeg — Pricing V3 only (PRICING_V3_DESIGN.md Phần 7). Which side
	// of the trip is at the airport, so AirportFeeRuleV3 can charge a
	// leg-specific fee instead of V2's single flat AirportFeeVND. Zero value
	// (AirportLegNone) is fully backward compatible: the V2 AirportFeeRule
	// never reads this field, only IsAirportZone.
	AirportLeg AirportLeg

	// Rain (BRB §2.2.13) — Ops-activated or weather-API-triggered externally.
	IsRainActive bool

	// Holiday (BRB §2.2.11) — resolved externally against the admin-portal
	// holiday calendar; this engine does not maintain holiday dates.
	IsHoliday bool

	// Traffic — no BRB surge rule; see TrafficLevel doc comment.
	TrafficLevel TrafficLevel

	// Special Event — no BRB rule exists (Festival Promotion, BRB §3.2.6, is a
	// promotion/discount, not a fare surge, and lives in the Promotion Engine).
	IsSpecialEvent bool
}

// NeutralContext builds a PricingContext with every surge signal at its
// inactive default. Every PricingRule evaluated against a NeutralContext
// with an all-disabled RuleConfig (see DefaultRuleConfigs in the app
// package) produces a 1.0 multiplier / 0 flat surcharge, guaranteeing
// FareCalculator's existing output is unchanged.
func NeutralContext(vehicleType VehicleType, requestTime time.Time) PricingContext {
	return PricingContext{
		VehicleType:      vehicleType,
		RequestTime:      requestTime,
		ActiveRequests:   0,
		AvailableDrivers: 0,
		TrafficLevel:     TrafficLevelNormal,
	}
}

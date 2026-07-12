package entity

// Constants and tables sourced verbatim from Business Rule Bible v1.0 (BRB)
// Part 2. These are DEFAULT values a RuleConfig may reference; no rule
// implementation hardcodes a number that isn't declared here, and every
// constant below cites its BRB section. Where BRB does not define a number
// (Supply Surge, Traffic, Special Event), no constant exists — see
// app/rules_todo.go.
const (
	// BRB §2.2.10 Night Surcharge — 22:00-05:00 local time.
	NightSurchargeMultiplier float64 = 1.20
	NightWindowStartHour     int     = 22
	NightWindowEndHour       int     = 5

	// BRB §2.2.11 Holiday Surcharge. Whether "today" is a holiday is NOT
	// computed here — BRB: "Holiday list. Maintained in the admin portal per
	// city/country." Callers populate PricingContext.IsHoliday from that
	// external calendar; this engine only applies the multiplier.
	HolidaySurchargeMultiplier float64 = 1.15

	// BRB §2.2.12 Peak Hour Surcharge — default windows, "adjustable per city."
	PeakHourSurchargeMultiplier float64 = 1.10

	// BRB §2.2.13 Rain Surcharge. Whether rain is active is NOT computed here
	// — BRB: triggered manually by Ops or by a verified weather API at
	// >=2mm/hour. Callers populate PricingContext.IsRainActive.
	RainSurchargeMultiplier float64 = 1.15

	// BRB §2.2.11: "Combined surcharge multipliers from Night + Holiday cannot
	// exceed x1.50."
	MaxCombinedNightHolidayMultiplier float64 = 1.50

	// BRB §2.2.13: "Night + Holiday + Rain multipliers applied simultaneously
	// cannot exceed x1.60."
	MaxCombinedNightHolidayRainMultiplier float64 = 1.60

	// BRB §2.13.3 Maximum Surge: "Surge multiplier cannot exceed x2.0. No
	// exception." This is a hard platform safety invariant, not an
	// operator-configurable value — RuleConfig.MaxSurge can only tighten this
	// further, never loosen it. See DemandSurgeRule.
	MaxDemandSurgeMultiplier float64 = 2.0

	// BRB §2.2.7 Airport Fee — fixed surcharge, not a multiplier. Applied once
	// per trip even if both origin and destination are airport zones.
	AirportFeeVND int64 = 10_000
)

// DSRTier is one row of the BRB §2.13.2 Demand-Supply Ratio -> Surge
// Multiplier table. MinDSR is inclusive; a DSR matches the highest tier whose
// MinDSR it meets or exceeds.
type DSRTier struct {
	MinDSR     float64
	Multiplier float64
	Label      string
}

// DefaultDSRTiers is the exact BRB §2.13.2 table:
//
//	DSR Range     Multiplier  Label
//	< 1.2         x1.0        Normal pricing
//	1.2 - 1.5     x1.2        Busy
//	1.5 - 2.0     x1.4        High demand
//	2.0 - 2.5     x1.6        Very high demand
//	2.5 - 3.0     x1.8        Peak demand
//	> 3.0         x2.0        Maximum surge
//
// This is a data table, not an if/else chain — DemandSurgeRule looks up the
// matching tier instead of branching on DSR value.
func DefaultDSRTiers() []DSRTier {
	return []DSRTier{
		{MinDSR: 0.0, Multiplier: 1.0, Label: "Normal pricing"},
		{MinDSR: 1.2, Multiplier: 1.2, Label: "Busy"},
		{MinDSR: 1.5, Multiplier: 1.4, Label: "High demand"},
		{MinDSR: 2.0, Multiplier: 1.6, Label: "Very high demand"},
		{MinDSR: 2.5, Multiplier: 1.8, Label: "Peak demand"},
		{MinDSR: 3.0, Multiplier: 2.0, Label: "Maximum surge"},
	}
}

// TimeWindow describes a recurring weekly time-of-day window in local time,
// used to configure Peak Hour windows without hardcoding them into rule
// logic ("adjustable per city" per BRB §2.2.12).
type TimeWindow struct {
	Weekdays  []int // time.Weekday values (0=Sunday .. 6=Saturday)
	StartHour int   // inclusive, 0-23
	EndHour   int   // exclusive, 0-23
}

// DefaultPeakHourWindows returns BRB §2.2.12's stated default windows:
// Monday-Friday 07:00-09:00 and 17:00-20:00. Operators may configure
// different windows per city via RuleConfig without changing rule code.
func DefaultPeakHourWindows() []TimeWindow {
	weekdays := []int{1, 2, 3, 4, 5} // Monday-Friday
	return []TimeWindow{
		{Weekdays: weekdays, StartHour: 7, EndHour: 9},
		{Weekdays: weekdays, StartHour: 17, EndHour: 20},
	}
}

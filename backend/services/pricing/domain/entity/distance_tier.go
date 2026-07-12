package entity

import "math"

// Package-level note: Pricing V3 (docs/business/PRICING_V3_DESIGN.md Phần 4,
// reviewed in docs/business/PRICING_V3_REVIEW.md Phần 5/13 mục 1-2) replaces
// the single VehicleRates.PerKmRate (one flat rate for the whole trip) with a
// degressive distance-tier table: the first kilometres cost more per km than
// the last, mirroring every competitor rate card researched in
// docs/business/MARKET_PRICING_RESEARCH.md §2. No tier boundary or rate is a
// literal in this file — every number a caller uses comes from
// backend/services/pricing/config (YAML), never a Go constant here.

// DistanceTier is one band of a degressive per-km rate table.
//
// FromKM is inclusive, ToKM is exclusive. The last tier in a table must have
// ToKM <= 0, which this package treats as "open-ended" (no upper bound) —
// PRICING_V3_REVIEW.md Phần 13 mục 2 flagged the ORIGINAL design's open-ended
// last tier as a real risk (fare keeps falling forever on very long trips
// with no floor); DistanceFareForTiers below enforces that the open-ended
// tier's rate is never used at a rate lower than itself by construction (it
// is the last tier evaluated), and config validation (config.Validate) can
// additionally require operators to set a floor if desired — this package
// does not silently invent one.
type DistanceTier struct {
	FromKM    float64
	ToKM      float64 // <= 0 means "open-ended, no upper bound"
	RatePerKM int64   // VND per km within [FromKM, ToKM)
}

// openEnded reports whether t has no upper bound.
func (t DistanceTier) openEnded() bool {
	return t.ToKM <= 0
}

// DistanceFareForTiers computes the total distance fare for a trip of
// distanceKM by walking a degressive tier table — a table lookup + running
// sum, not an if/else chain, matching the same "rule engine, not branching"
// discipline already used by PricingPipeline/PricingRule for surge.
//
// tiers must be sorted ascending by FromKM with no gaps (config.Validate
// enforces this at load time so this function can stay a pure, unchecked
// loop — the same division of responsibility PricingRule uses: rules trust
// their injected config, RuleConfig/config validation is where invariants
// are enforced once).
func DistanceFareForTiers(tiers []DistanceTier, distanceKM float64) int64 {
	if distanceKM <= 0 || len(tiers) == 0 {
		return 0
	}

	var total float64
	for _, tier := range tiers {
		if distanceKM <= tier.FromKM {
			break
		}
		upper := distanceKM
		if !tier.openEnded() && tier.ToKM < upper {
			upper = tier.ToKM
		}
		segmentKM := upper - tier.FromKM
		if segmentKM <= 0 {
			continue
		}
		total += segmentKM * float64(tier.RatePerKM)
	}
	return int64(math.Round(total))
}

// VehicleRatesV3 holds the per-vehicle-class Pricing V3 fare parameters.
// All monetary fields are in whole VND (BRB §2.16 — no decimal subunit).
type VehicleRatesV3 struct {
	// BaseFare is charged every trip regardless of distance or time —
	// unchanged role from V2 (BRB §2.2.1).
	BaseFare int64

	// DistanceTiers replaces V2's single PerKmRate — see PRICING_V3_DESIGN.md
	// Phần 4. Must be sorted ascending by FromKM (config.Load validates this).
	DistanceTiers []DistanceTier

	// TrafficTimePerMinute is what V2 called "PerMinuteRate" / "Time Fare" —
	// renamed per PRICING_V3_DESIGN.md Phần 5.1 to make explicit that it only
	// bills minutes spent below the slow-traffic speed threshold, never
	// "every minute of the trip" (BRB §2.2.3's own rule, just a clearer name).
	TrafficTimePerMinute int64

	// WaitingFeePerMinute / WaitingGraceMinutes — BRB §2.2.9, billed from the
	// moment the driver marks "Arrived" (pre-trip), a different concept from
	// TrafficTimePerMinute (in-trip). See PRICING_V3_DESIGN.md Phần 5.1.
	WaitingFeePerMinute int64
	WaitingGraceMinutes int

	// MinimumFare is the floor applied to (BaseFare + DistanceFare +
	// TrafficTimeFare), before BookingFee — unchanged role from V2 (BRB §2.2.4).
	MinimumFare int64

	// BookingFee is the flat, non-surged, 100%-platform-revenue fee added
	// after minimum-fare enforcement — unchanged role from V2 (BRB §2.2.5).
	BookingFee int64
}

// FareConfigV3 holds the complete Pricing V3 configuration for one market
// (city). Loaded exclusively via backend/services/pricing/config — no
// literal VehicleRatesV3 value is constructed inline anywhere else in this
// service, so every number in a running Pricing V3 calculation traces back
// to a config file, never a Go source literal (per the sprint's "không
// hardcode bất kỳ con số nào" requirement).
type FareConfigV3 struct {
	CurrencyCode string
	Rates        map[VehicleType]VehicleRatesV3
}

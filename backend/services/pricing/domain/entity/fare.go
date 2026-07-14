// Package entity defines the pricing domain model for FAIRRIDE.
// All monetary amounts are int64 in the smallest unit of the configured currency
// (e.g. cents for USD, satang for THB — but whole VND, which has no subunit
// per Business Rule Bible v1.0 §2.2.9). This avoids floating-point rounding
// errors in final stored values while keeping intermediate math in float64.
package entity

// VehicleType matches the enum in the driver service; duplicated here so the
// pricing service has no cross-service import dependency.
type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeVan        VehicleType = "van"
	VehicleTypeBikePlus   VehicleType = "bike_plus"
	VehicleTypeCarXL      VehicleType = "car_xl"
)

// VehicleRates holds the per-vehicle-type fare parameters.
// All monetary fields are in the smallest unit of the configured currency.
type VehicleRates struct {
	// BaseFare is charged every trip regardless of distance or time.
	BaseFare int64
	// PerKmRate is multiplied by distance in km and rounded to the nearest unit.
	PerKmRate int64
	// PerMinuteRate is multiplied by duration in minutes and rounded to the nearest unit.
	PerMinuteRate int64
	// MinimumFare is the floor applied to (BaseFare + DistanceFare + TimeFare).
	// BookingFee is added on top of this floor, not included in it.
	MinimumFare int64
	// BookingFee is the fixed platform fee added after minimum fare enforcement.
	BookingFee int64
}

// FareConfig holds the complete pricing configuration for one market.
// Rates are keyed by VehicleType so a single service instance supports all types.
type FareConfig struct {
	// CurrencyCode is the ISO 4217 code for the configured currency (e.g. "USD", "VND", "THB").
	CurrencyCode string
	// Rates maps each supported vehicle type to its fare parameters.
	Rates map[VehicleType]VehicleRates
}

// DefaultFareConfig returns the launch-market rates. VND has no decimal
// subunit (BRB v1.0 §2.2.9), so amounts here are whole VND, not cents.
// This is Pricing V2's config — the DEFAULT engine (PricingVersionV2, see
// app/feature_flag.go); Pricing V3's tiered config is a separate, currently
// opt-in system (config/pricing_v3.default.yaml, PRICING_VERSION=v3) not
// touched by this calibration.
//
// Car/Motorcycle calibrated 2026-07-14 to sit within ±5% of the real
// Grab/Be/Xanh SM average (docs/business/MARKET_PRICING_RESEARCH.md Phần 2,
// sourced 2026-07-11) across the full 2-60km range, replacing an earlier
// deliberate "10-20% cheaper" positioning that had drifted to -42.5% to
// -52.8% below market once actually checked against real fares — coefficients
// were fit by linear regression against the market curve (BaseFare+BookingFee
// as intercept, PerKmRate+2.2×PerMinuteRate as slope, matching this formula's
// own shape — DurationMin ≈ 2.2×DistanceKM per the research doc's convention),
// NOT copied from any single competitor's price:
//   - Motorcycle: fits the market curve almost exactly (max |err| ~0.1%,
//     13 distances 2-60km) — Grab/Be/Xanh SM bike pricing has no long-haul
//     distance tiering in the public rate cards found, so a flat rate tracks
//     it closely.
//   - Car: max |err| ~2.9% across the same 13 distances — comfortably inside
//     the ±5% target despite the market curve being mildly degressive
//     (Be/Xanh SM taper their /km on long trips; Grab's public rate card has
//     no visible taper). A flat V2 rate cannot fit a genuinely tiered curve
//     at both very-short and very-long distances simultaneously — see
//     Pricing V3 (config/pricing_v3.default.yaml) for the distance-tiered
//     engine that exists for exactly this reason.
//
// CarXL/BikePlus have no independent multi-platform market data (BRB/PS/the
// research doc all note Vietnam platforms rarely publish standalone XL rate
// cards) — kept at the same ratios over Car/Motorcycle the previous
// calibration already used (CarXL ≈1.3x Car per the research doc's XL
// assumption, BikePlus ≈1.2x Motorcycle), recomputed off the new base
// numbers rather than independently fit. Xanh SM's published Premium/Luxury
// rate (21,000/km flat) is a different, higher product tier and is not a
// valid ±5% target for a standard XL/7-seat class.
//
// VehicleTypeVan is unchanged — explicitly legacy/not part of the
// rider-facing catalog (CarXL replaced it there, see the pre-existing note
// below), so it was left out of this pass.
func DefaultFareConfig() FareConfig {
	return FareConfig{
		CurrencyCode: "VND",
		Rates: map[VehicleType]VehicleRates{
			VehicleTypeMotorcycle: {
				BaseFare:      2_700,
				PerKmRate:     4_140,
				PerMinuteRate: 200,
				MinimumFare:   9_000,
				BookingFee:    2_000,
			},
			VehicleTypeBikePlus: {
				BaseFare:      3_200,
				PerKmRate:     5_000,
				PerMinuteRate: 240,
				MinimumFare:   11_000,
				BookingFee:    2_000,
			},
			VehicleTypeCar: {
				BaseFare:      11_250,
				PerKmRate:     10_700,
				PerMinuteRate: 450,
				MinimumFare:   25_000,
				BookingFee:    2_000,
			},
			VehicleTypeCarXL: {
				BaseFare:      14_600,
				PerKmRate:     13_900,
				PerMinuteRate: 580,
				MinimumFare:   32_000,
				BookingFee:    2_000,
			},
			VehicleTypeVan: {
				BaseFare:      18_000, // BRB §2.2.1 XL — legacy, unchanged (see doc comment above)
				PerKmRate:     5_000,  // BRB §2.2.2 XL
				PerMinuteRate: 500,    // BRB §2.2.3 XL
				MinimumFare:   40_000, // BRB §2.2.4 XL
				BookingFee:    2_000,  // BRB §2.2.5 (flat, all classes)
			},
		},
	}
}

// FareBreakdown is the fully itemised fare returned to callers.
// Monetary fields are in the smallest currency unit defined by CurrencyCode.
//
// Calculation rules (enforced by FareCalculator):
//
//	ride_fare = max(BaseFare + DistanceFare + TimeFare, MinimumFare)
//	Total     = RideFare + BookingFee
type FareBreakdown struct {
	VehicleType  VehicleType
	DistanceKM   float64
	DurationMin  float64
	BaseFare     int64
	DistanceFare int64
	TimeFare     int64
	BookingFee   int64
	RideFare     int64 // base + distance + time, after minimum enforcement
	Total        int64 // RideFare + BookingFee
	CurrencyCode string
	IsFinal      bool // false = estimate, true = post-trip final
}

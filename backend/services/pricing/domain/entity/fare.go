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
)

// NOTE (Vehicle/Service Catalog refactor): Pricing is explicitly out of
// scope for this refactor ("Không thay đổi Pricing") — it keeps rating by
// VehicleType only, exactly as before. This means Bike and Bike Plus (both
// VehicleType=motorcycle) currently price identically, and Car XL
// (VehicleType=van, see driver's ServiceType.RequiredVehicleType) prices
// the same as a plain Van/XL booking — there is no Bike-Plus-specific or
// Car-XL-specific rate card. See the refactor's report, "Known Gap".

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

// DefaultFareConfig returns the launch-market rates from the Business Rule
// Bible v1.0 (docs/business/business-rule-bible-v1.0.md §2.2.1-§2.2.5). VND
// has no decimal subunit (§2.2.9), so amounts here are whole VND, not cents.
//
// VehicleTypeCar uses the BRB "Standard (4-seat)" row and VehicleTypeVan uses
// "XL (7-seat)" — the closest match for each of this service's three
// VehicleType values. BRB v1.0 defines no motorcycle-specific rates; the
// VehicleTypeMotorcycle figures below are an interim estimate (~40% of the
// Standard car rate, roughly matching observed market bike/car ratios) and
// are NOT sourced from the BRB — replace once product defines an official
// motorcycle rate.
func DefaultFareConfig() FareConfig {
	return FareConfig{
		CurrencyCode: "VND",
		Rates: map[VehicleType]VehicleRates{
			VehicleTypeCar: {
				BaseFare:      10_000, // BRB §2.2.1 Standard
				PerKmRate:     4_000,  // BRB §2.2.2 Standard
				PerMinuteRate: 400,    // BRB §2.2.3 Standard
				MinimumFare:   25_000, // BRB §2.2.4 Standard
				BookingFee:    2_000,  // BRB §2.2.5 (flat, all classes)
			},
			VehicleTypeMotorcycle: {
				BaseFare:      5_000,  // interim estimate — not in BRB v1.0
				PerKmRate:     1_600,  // interim estimate — not in BRB v1.0
				PerMinuteRate: 200,    // interim estimate — not in BRB v1.0
				MinimumFare:   12_000, // interim estimate — not in BRB v1.0
				BookingFee:    2_000,  // BRB §2.2.5 (flat, all classes)
			},
			VehicleTypeVan: {
				BaseFare:      18_000, // BRB §2.2.1 XL
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

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
//
// Car/Motorcycle/BikePlus/CarXL are calibrated to undercut Be/Xanh SM by
// ~10-20% on a real ~11.2km reference route (R7 Phú Mỹ Hưng → Ga T3 Tân
// Sơn Nhất) rather than taken from BRB §2.2.1-§2.2.5 verbatim — the BRB
// Standard/XL rows priced noticeably above that competitor band once
// checked against real fares. BikePlus/CarXL are new rate cards (~1.2x
// Bike, ~1.35x Car) so they no longer alias Motorcycle/Van pricing.
// VehicleTypeVan keeps its original BRB XL row — it's legacy, not part of
// the rider-facing catalog (CarXL replaced it there).
func DefaultFareConfig() FareConfig {
	return FareConfig{
		CurrencyCode: "VND",
		Rates: map[VehicleType]VehicleRates{
			VehicleTypeMotorcycle: {
				BaseFare:      8_000,
				PerKmRate:     2_800,
				PerMinuteRate: 200,
				MinimumFare:   15_000,
				BookingFee:    2_000,
			},
			VehicleTypeBikePlus: {
				BaseFare:      10_000,
				PerKmRate:     3_400,
				PerMinuteRate: 250,
				MinimumFare:   18_000,
				BookingFee:    2_000,
			},
			VehicleTypeCar: {
				BaseFare:      15_000,
				PerKmRate:     6_500,
				PerMinuteRate: 400,
				MinimumFare:   30_000,
				BookingFee:    2_000,
			},
			VehicleTypeCarXL: {
				BaseFare:      22_000,
				PerKmRate:     8_500,
				PerMinuteRate: 550,
				MinimumFare:   45_000,
				BookingFee:    2_000,
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

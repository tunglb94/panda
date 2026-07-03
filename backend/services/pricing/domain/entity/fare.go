// Package entity defines the pricing domain model for FAIRRIDE.
// All monetary amounts are int64 in the smallest unit of the configured currency
// (e.g. cents for USD, satang for THB). This avoids floating-point rounding errors
// in final stored values while keeping intermediate math in float64.
package entity

// VehicleType matches the enum in the driver service; duplicated here so the
// pricing service has no cross-service import dependency.
type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeVan        VehicleType = "van"
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

// DefaultFareConfig returns conservative USD-equivalent defaults suitable for
// testing and early development. Operators MUST override these for production.
// At default rates a 5 km / 15 min car trip totals $2.75 + $0.50 booking fee = $3.25.
func DefaultFareConfig() FareConfig {
	return FareConfig{
		CurrencyCode: "USD",
		Rates: map[VehicleType]VehicleRates{
			VehicleTypeCar: {
				BaseFare:      50,  // $0.50
				PerKmRate:     30,  // $0.30/km
				PerMinuteRate: 5,   // $0.05/min
				MinimumFare:   200, // $2.00
				BookingFee:    50,  // $0.50
			},
			VehicleTypeMotorcycle: {
				BaseFare:      30,  // $0.30
				PerKmRate:     20,  // $0.20/km
				PerMinuteRate: 3,   // $0.03/min
				MinimumFare:   150, // $1.50
				BookingFee:    30,  // $0.30
			},
			VehicleTypeVan: {
				BaseFare:      100, // $1.00
				PerKmRate:     50,  // $0.50/km
				PerMinuteRate: 8,   // $0.08/min
				MinimumFare:   300, // $3.00
				BookingFee:    75,  // $0.75
			},
		},
	}
}

// FareBreakdown is the fully itemised fare returned to callers.
// Monetary fields are in the smallest currency unit defined by CurrencyCode.
//
// Calculation rules (enforced by FareCalculator):
//   ride_fare = max(BaseFare + DistanceFare + TimeFare, MinimumFare)
//   Total     = RideFare + BookingFee
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

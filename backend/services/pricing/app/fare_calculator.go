// Package app contains the fare calculation business logic for FAIRRIDE.
// No surge, promotions, coupons, dynamic pricing, or peak-hour pricing.
package app

import (
	"math"

	"github.com/fairride/pricing/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

// FareCalculator calculates ride fares from a FareConfig.
// It is stateless and safe for concurrent use.
type FareCalculator struct {
	config entity.FareConfig
}

// NewFareCalculator creates a FareCalculator with the given config.
// Panics if config.Rates is nil (misconfiguration).
func NewFareCalculator(config entity.FareConfig) *FareCalculator {
	if config.Rates == nil {
		panic("pricing: FareConfig.Rates must not be nil")
	}
	return &FareCalculator{config: config}
}

// Estimate calculates an upfront fare before the trip starts.
// The same formula is used for both estimates and final fares (upfront pricing guarantee).
//
// Returns CodeInvalidArgument for unknown vehicle types or negative inputs.
func (c *FareCalculator) Estimate(vehicleType entity.VehicleType, distanceKM, durationMin float64) (*entity.FareBreakdown, error) {
	return c.calculate(vehicleType, distanceKM, durationMin, false)
}

// CalculateFinal calculates the fare after trip completion using actual distance and time.
// Uses the same formula as Estimate — no deviation penalty in MVP.
//
// Returns CodeInvalidArgument for unknown vehicle types or negative inputs.
func (c *FareCalculator) CalculateFinal(vehicleType entity.VehicleType, distanceKM, durationMin float64) (*entity.FareBreakdown, error) {
	return c.calculate(vehicleType, distanceKM, durationMin, true)
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (c *FareCalculator) calculate(vehicleType entity.VehicleType, distanceKM, durationMin float64, isFinal bool) (*entity.FareBreakdown, error) {
	if distanceKM < 0 {
		return nil, domainerrors.InvalidArgument("distance_km must be non-negative")
	}
	if durationMin < 0 {
		return nil, domainerrors.InvalidArgument("duration_minutes must be non-negative")
	}

	rates, ok := c.config.Rates[vehicleType]
	if !ok {
		return nil, domainerrors.InvalidArgument("unsupported vehicle type: " + string(vehicleType))
	}

	baseFare := rates.BaseFare
	distanceFare := roundToUnit(float64(rates.PerKmRate) * distanceKM)
	timeFare := roundToUnit(float64(rates.PerMinuteRate) * durationMin)

	rideFare := baseFare + distanceFare + timeFare
	if rideFare < rates.MinimumFare {
		rideFare = rates.MinimumFare
	}

	total := rideFare + rates.BookingFee

	return &entity.FareBreakdown{
		VehicleType:  vehicleType,
		DistanceKM:   distanceKM,
		DurationMin:  durationMin,
		BaseFare:     baseFare,
		DistanceFare: distanceFare,
		TimeFare:     timeFare,
		BookingFee:   rates.BookingFee,
		RideFare:     rideFare,
		Total:        total,
		CurrencyCode: c.config.CurrencyCode,
		IsFinal:      isFinal,
	}, nil
}

// roundToUnit converts a float64 fare component to the nearest integer currency unit.
// Ties round away from zero (standard rounding behaviour).
func roundToUnit(v float64) int64 {
	return int64(math.Round(v))
}

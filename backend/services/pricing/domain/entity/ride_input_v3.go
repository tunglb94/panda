package entity

import "time"

// RideInputV3 is every signal FareCalculatorV3.Estimate/CalculateFinal needs
// — the Pricing V3 analogue of PricingContext, extended with the fields V2
// never modelled (SlowTrafficMin/WaitingMin) and the new production
// responsibilities (CommissionTier, a pre-resolved Voucher amount). Mirrors
// the shape of backend/services/pricing/simulation's TripInput deliberately
// (same field names/roles where they overlap) so a caller familiar with one
// is immediately familiar with the other — but this type is NOT imported by
// or from the simulation package, keeping production and simulation
// decoupled per fare_calculator.go's existing isolation discipline.
type RideInputV3 struct {
	VehicleType VehicleType
	DistanceKM  float64
	DurationMin float64

	// SlowTrafficMin is the portion of DurationMin spent below the slow-
	// traffic speed threshold (BRB §2.2.3 "greater of time or distance" rule
	// — mutually exclusive with DistanceFare for the same portion of the
	// trip). 0 means "not modelled / assume no slow traffic", the same
	// default simulation.TripInput uses.
	SlowTrafficMin float64

	// WaitingMin is total minutes the driver waited after marking "Arrived",
	// before VehicleRatesV3.WaitingGraceMinutes is subtracted. New to
	// production in V3 — V2's FareCalculator never accepted this input.
	WaitingMin float64

	RequestTime time.Time // drives Night/Peak detection (existing PricingRule behaviour, unchanged)

	IsHoliday    bool
	IsRainActive bool

	// ActiveRequests / AvailableDrivers drive Demand Surge (BRB §2.13.2),
	// unchanged from PricingContext.
	ActiveRequests   int
	AvailableDrivers int

	// AirportLeg drives AirportFeeRuleV3 (rules_airport_v3.go). AirportLegNone
	// means this trip has no airport pickup/dropoff component.
	AirportLeg AirportLeg

	// CommissionTier selects the driver's take-rate from CommissionConfigV3
	// (BRB §7.1). Empty/unrecognised values fail closed to Bronze — see
	// CommissionConfigV3.Rate.
	CommissionTier CommissionTier

	// VoucherLabel/VoucherDiscountVND are a PRE-RESOLVED promotion outcome —
	// Pricing V3 does not decide voucher eligibility (that remains the
	// Promotion Engine's job, BRB Part 3/4); it only applies whatever amount
	// the caller supplies, clamped so the rider is never charged negative
	// (BRB §4.9). Zero VoucherDiscountVND means no promotion applies.
	VoucherLabel       string
	VoucherDiscountVND int64
}

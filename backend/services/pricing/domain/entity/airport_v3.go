package entity

// AirportLeg identifies which side of an airport trip a fare calculation is
// for — PRICING_V3_DESIGN.md Phần 7 splits the old single flat Airport Fee
// (BRB §2.2.7) into leg-specific fees because MARKET_PRICING_RESEARCH.md
// Phần 1.3 found the flat fee, applied identically to every vehicle class,
// created a real bias: competitors do not charge a motorcycle the same
// airport surcharge as a car.
//
// Queue Compensation (paid to the driver while queuing at the airport lot)
// and Priority Dispatch (a dispatch-ordering perk for high-tier drivers) are
// PRICING_V3_DESIGN.md Phần 7 concepts too, but neither changes what the
// rider pays — Queue Compensation is a platform-funded driver payment
// (same treatment as Long Pickup Compensation) and Priority Dispatch is a
// dispatch-ordering concern. Neither belongs in a fare-facing PricingRule,
// so this package only models the two components that actually change the
// rider's fare: Pickup and Dropoff.
type AirportLeg string

const (
	AirportLegNone    AirportLeg = ""
	AirportLegPickup  AirportLeg = "pickup"
	AirportLegDropoff AirportLeg = "dropoff"
)

// AirportFeeConfigV3 holds the Pickup/Dropoff fee for each vehicle type,
// loaded from config (backend/services/pricing/config) — never hardcoded.
// A vehicle type absent from a map (e.g. motorcycle, per
// MARKET_PRICING_RESEARCH.md Phần 1.3) simply has no entry, which
// AirportFeeRuleV3 treats as a 0 VND fee rather than a missing-config error,
// since "this vehicle class pays nothing at the airport" is a valid,
// deliberate configuration, not a mistake.
type AirportFeeConfigV3 struct {
	PickupFee  map[VehicleType]int64
	DropoffFee map[VehicleType]int64
}

// FeeFor returns the configured fee for (vehicleType, leg). Returns 0 for
// AirportLegNone or an unconfigured vehicle type — never an error, since a
// missing entry is a legitimate "no airport surcharge for this class"
// configuration, not a fail-closed condition (unlike RuleConfigMap.Get,
// which fails closed for *disabled* rules — here the rule can be enabled
// while simply having no fee for a given vehicle class).
func (c AirportFeeConfigV3) FeeFor(vehicleType VehicleType, leg AirportLeg) int64 {
	switch leg {
	case AirportLegPickup:
		return c.PickupFee[vehicleType]
	case AirportLegDropoff:
		return c.DropoffFee[vehicleType]
	default:
		return 0
	}
}

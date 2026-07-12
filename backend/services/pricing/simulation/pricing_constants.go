// Package simulation is a standalone, read-only pricing SIMULATION engine for
// Panda. It is completely isolated from the production Pricing Service
// (package app, package grpc, cmd/server) — nothing in cmd/server/main.go or
// grpc/handler.go imports this package, and this package imports nothing from
// them. It exists to let product/finance test pricing rules (including rules
// proposed but not yet ratified into the Business Rule Bible) before any of
// them are ever wired into production.
//
// Source of truth: per the sprint brief, when this package and
// docs/business/PRICING_STRATEGY.md disagree, docs/business/
// business-rule-bible-v1.0.md (BRB) wins. Every constant below cites the BRB
// section it was taken from. Where BRB is silent and a value was necessary to
// produce a runnable engine, the constant is marked "ASSUMPTION" and must be
// resolved by the CPO/CFO (mirroring BRB's own "Unresolved Business Decision"
// pattern in Appendix B) before any of this is proposed for production use.
//
// Currency: VND, integer, no decimal subunit (BRB §2.16) — matches BRB's
// launch-market currency. This deliberately diverges from the production
// Pricing Service's current DefaultFareConfig, which is wired to USD test
// values (see the audit in docs/business/PRICING_SIMULATION_REPORT.md §1).
package simulation

// ─── Vehicle types ──────────────────────────────────────────────────────────
//
// GAP (documented in the audit): BRB §2.2.1 prices vehicle classes "Standard
// (4-seat) / Premium (4-seat) / XL (7-seat)". The production Pricing Service's
// entity.VehicleType enum (car / motorcycle / van — see domain/entity/fare.go)
// does not match those names, and BRB never prices a two-wheeler at all even
// though the shipped Rider app's vehicle selector offers car/motorcycle/van
// (see apps/rider/lib/features/booking/domain/models/vehicle_option.dart).
// This package keeps the production VehicleType names so simulation results
// are directly comparable to what would actually ship, and maps them onto BRB
// rates as: car → BRB "Standard", van → BRB "XL". Motorcycle has no BRB rate
// at all — see MotorcycleFareRatio below (ASSUMPTION).

type VehicleType string

const (
	VehicleCar        VehicleType = "car"
	VehicleMotorcycle VehicleType = "motorcycle"
	VehicleVan        VehicleType = "van"
)

// MotorcycleFareRatio — ASSUMPTION, not in BRB. BRB prices no two-wheeler
// category. A motorcycle trip is estimated at 60% of the "car" (BRB
// Standard) rate for every per-component field (base/distance/time/minimum),
// consistent with the typical motorcycle-vs-car fare ratio observed in other
// Southeast Asian ride-hailing markets (Grab/Be/Xanh SM all price bike well
// below their 4-seat car tier). This ratio MUST be replaced with a real
// CPO-approved rate before any production use — it is a placeholder that
// makes the simulator runnable for all three vehicle types the app actually
// offers, not a business decision.
const MotorcycleFareRatio = 0.60

// ─── Fare components (BRB §2.2) ─────────────────────────────────────────────

// VehicleRates holds the per-vehicle-class fare parameters, all in VND.
type VehicleRates struct {
	BaseFare      int64 // BRB §2.2.1 — flat, charged every trip, never surged
	PerKmRate     int64 // BRB §2.2.2 — per km, billed only while speed ≥ 10 km/h
	PerMinuteRate int64 // BRB §2.2.3 — per minute, billed only while speed < 10 km/h
	MinimumFare   int64 // BRB §2.2.4 — floor on (base+distance+time), before booking fee
}

// DefaultVehicleRates are the BRB §2.2.1–§2.2.4 "primary city, launch" VND
// rates, keyed by the production VehicleType names (see GAP note above).
func DefaultVehicleRates() map[VehicleType]VehicleRates {
	car := VehicleRates{
		BaseFare:      10_000, // BRB §2.2.1 Standard
		PerKmRate:     4_000,  // BRB §2.2.2 Standard
		PerMinuteRate: 400,    // BRB §2.2.3 Standard
		MinimumFare:   25_000, // BRB §2.2.4 Standard
	}
	van := VehicleRates{
		BaseFare:      18_000, // BRB §2.2.1 XL
		PerKmRate:     5_000,  // BRB §2.2.2 XL
		PerMinuteRate: 500,    // BRB §2.2.3 XL
		MinimumFare:   40_000, // BRB §2.2.4 XL
	}
	moto := VehicleRates{
		BaseFare:      scaleVND(car.BaseFare, MotorcycleFareRatio),
		PerKmRate:     scaleVND(car.PerKmRate, MotorcycleFareRatio),
		PerMinuteRate: scaleVND(car.PerMinuteRate, MotorcycleFareRatio),
		MinimumFare:   scaleVND(car.MinimumFare, MotorcycleFareRatio),
	}
	return map[VehicleType]VehicleRates{
		VehicleCar:        car,
		VehicleVan:        van,
		VehicleMotorcycle: moto,
	}
}

// TimeFareSpeedThresholdKPH is the speed below which movement is billed as
// time fare instead of distance fare — BRB §2.2.3. The two components are
// mutually exclusive within any given second of the trip (BRB: "greater of
// time or distance", same approach as Uber/Lyft).
const TimeFareSpeedThresholdKPH = 10.0

// BookingFee is the flat, non-surged, 100%-platform-revenue fee added after
// minimum-fare enforcement — BRB §2.2.5. Same for every vehicle type.
// Per docs/business/PRICING_STRATEGY.md §3.3, this also serves as the
// engine's "Service Fee" output — BRB has no separate Service Fee line, and
// inventing a second identical fee would double-count platform revenue.
const BookingFee int64 = 2_000

// ─── Commission / driver tiers (BRB §7.1) ───────────────────────────────────

type DriverTier string

const (
	TierBronze   DriverTier = "bronze"
	TierSilver   DriverTier = "silver"
	TierGold     DriverTier = "gold"
	TierPlatinum DriverTier = "platinum"
	TierDiamond  DriverTier = "diamond"
)

// CommissionRate returns the platform's take rate for a driver tier —
// BRB §7.1 (commission is applied to metered fare + applicable surcharges;
// never to BookingFee or to Toll/Bridge/Parking pass-through amounts).
func CommissionRate(tier DriverTier) float64 {
	switch tier {
	case TierSilver:
		return 0.18
	case TierGold:
		return 0.16
	case TierPlatinum:
		return 0.14
	case TierDiamond:
		return 0.12
	default: // TierBronze and unrecognised input both fall back to the entry rate.
		return 0.20
	}
}

// ─── Surcharges (BRB §2.2.7–§2.2.13) ────────────────────────────────────────

const (
	// AirportFee — BRB §2.2.7. Fixed, shared with the driver at the trip's
	// normal commission split, charged once even if both ends are airport zones.
	AirportFee int64 = 10_000

	// WaitingGraceMinutes / WaitingFeePerMinute — BRB §2.2.9. First 3 minutes
	// after the driver marks "Arrived" are free; billed from minute 4 onward.
	WaitingGraceMinutes  = 3
	WaitingFeePerMinute  int64 = 500

	// NightSurchargeMultiplier — BRB §2.2.10. 22:00–05:00, applied to
	// (base+distance+time), not to BookingFee/Toll.
	NightSurchargeMultiplier = 1.20
	NightStartHour           = 22 // inclusive
	NightEndHour             = 5  // exclusive

	// HolidaySurchargeMultiplier — BRB §2.2.11.
	HolidaySurchargeMultiplier = 1.15

	// PeakHourSurchargeMultiplier — BRB §2.2.12. Does NOT stack with Dynamic
	// Surge — when surge is active it supersedes peak-hour pricing entirely.
	PeakHourSurchargeMultiplier = 1.10

	// RainSurchargeMultiplier — BRB §2.2.13.
	RainSurchargeMultiplier = 1.15

	// StaticSurchargeCap — BRB §2.2.11 ("Night+Holiday cannot exceed ×1.50")
	// and §2.2.13 ("Night+Holiday+Rain cannot exceed ×1.60"). This engine
	// always evaluates all three together, so the wider ×1.60 cap applies;
	// the narrower ×1.50 case is just this cap never being reached because
	// Rain was inactive.
	StaticSurchargeCap = 1.60
)

// PeakWindow describes a static peak-hour time band — BRB §2.2.12 defaults.
type PeakWindow struct {
	StartHour, EndHour int  // [StartHour, EndHour)
	WeekdayOnly        bool
}

// DefaultPeakWindows — BRB §2.2.12 defaults (Mon–Fri 07:00–09:00, 17:00–20:00).
func DefaultPeakWindows() []PeakWindow {
	return []PeakWindow{
		{StartHour: 7, EndHour: 9, WeekdayOnly: true},
		{StartHour: 17, EndHour: 20, WeekdayOnly: true},
	}
}

// ─── Dynamic surge (BRB §2.13) ──────────────────────────────────────────────

// SurgeBand is one row of the BRB §2.13.2 demand/supply-ratio → multiplier
// table. DSR = active requests ÷ available drivers in a zone.
type SurgeBand struct {
	MaxDSR     float64 // upper bound of this band, exclusive; +Inf for the last band
	Multiplier float64
}

// DefaultSurgeBands — BRB §2.13.2 table, verbatim.
func DefaultSurgeBands() []SurgeBand {
	return []SurgeBand{
		{MaxDSR: 1.2, Multiplier: 1.0},
		{MaxDSR: 1.5, Multiplier: 1.2},
		{MaxDSR: 2.0, Multiplier: 1.4},
		{MaxDSR: 2.5, Multiplier: 1.6},
		{MaxDSR: 3.0, Multiplier: 1.8},
		{MaxDSR: -1, Multiplier: 2.0}, // MaxDSR<0 is the sentinel for "> 3.0" (last band)
	}
}

// MaxSurgeMultiplier — BRB §2.13.3. Absolute, no exception.
const MaxSurgeMultiplier = 2.0

// PriceCapVND — BRB §2.13.6. Absolute maximum fare (before promotions),
// regardless of surge/distance/time, for a Standard-class city trip.
const PriceCapVND int64 = 500_000

// ─── Driver protection (BRB §2.14) ──────────────────────────────────────────

// MinimumDriverEarningVND — BRB §2.14. A driver who completes a trip always
// earns at least this much net of commission; the platform tops up the gap.
// This top-up is a platform cost, not recovered from the rider.
const MinimumDriverEarningVND int64 = 20_000

// ─── Rounding (BRB §2.15) ────────────────────────────────────────────────────

const (
	// RiderRoundingUnit — total billed to rider rounds UP to the nearest
	// this-many VND.
	RiderRoundingUnit int64 = 500
	// DriverRoundingUnit — driver payout rounds UP to the nearest this-many VND.
	DriverRoundingUnit int64 = 100
)

// ─── VAT (ASSUMPTION — BRB is silent on ride-level VAT) ─────────────────────
//
// BRB §6.10 says platform tax is "calculated and remitted by the Finance
// team independently of this document" — it gives no per-trip VAT rate to
// plug into a fare formula. Vietnam's standard VAT rate for transport
// services is 10% (8% under the temporary reduction that has applied in
// various periods). This engine applies VAT to Platform Revenue only
// (commission + booking fee) — never to the driver's share, never to
// pass-through Toll/Bridge/Parking — as the least-surprising interpretation,
// but this is explicitly a simulation ASSUMPTION requiring a CFO decision,
// analogous to BRB Appendix B's "Unresolved Business Decision" items.
const AssumedVATRate = 0.10

// ─── Insurance (ASSUMPTION — BRB Appendix B, UBD-004: unresolved) ───────────
//
// BRB Appendix B UBD-004 states no insurance product has been procured yet.
// This engine therefore defaults per-trip insurance cost to zero rather than
// inventing a premium — the field exists in the output so the report can
// show "—" honestly instead of omitting the line item the sprint requested.
const AssumedInsuranceCostVND int64 = 0

// ─── Promotion clamp (BRB §4.9) ─────────────────────────────────────────────
//
// A discount can never take the rider's payment below zero; any excess is
// forfeited, never credited (unless the voucher is explicitly a "Wallet
// Cashback" type, which is out of scope for a per-trip fare simulation).

// ─── Proposed-but-unratified rules (docs/business/PRICING_STRATEGY.md) ─────
//
// The following are explicitly marked [MỚI — cần tu chính BRB] in the
// strategy document — i.e. proposed, NOT part of the Business Rule Bible.
// They are implemented here specifically so this simulation engine can
// numerically test them before anyone proposes ratifying them into BRB —
// that is the entire purpose of a simulation engine. They must never be
// treated as already-approved.

const (
	// LongPickupCompensationNearVND / FarVND — PRICING_STRATEGY §2.2. Funded
	// 100% by the platform, never charged to the rider — dispatch-matching a
	// far-away driver is a platform inefficiency, not the rider's fault.
	LongPickupNearKM             = 3.0 // 3–5 km
	LongPickupFarKM              = 5.0 // >5 km
	LongPickupCompensationNearVND int64 = 10_000
	LongPickupCompensationFarVND  int64 = 20_000

	// BridgeParking pass-through — PRICING_STRATEGY §2.2, same treatment as
	// BRB §2.2.8 Toll: paid at cost, zero commission, 100% to the amount
	// the rider reimburses the driver.
)

// scaleVND scales a VND amount by a ratio and rounds to the nearest VND.
func scaleVND(amount int64, ratio float64) int64 {
	return int64(float64(amount)*ratio + 0.5)
}

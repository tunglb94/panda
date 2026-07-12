package simulation

import (
	"time"

	domainerrors "github.com/fairride/shared/errors"
)

// WeatherCondition is a simulation input signal — see PRICING_STRATEGY.md §5.2.
type WeatherCondition string

const (
	WeatherClear WeatherCondition = "clear"
	WeatherRain  WeatherCondition = "rain"
)

// PromotionInput is a pre-resolved discount amount. This engine does not
// itself decide *whether* a rider qualifies for First Ride / Student /
// Membership / Voucher (that eligibility logic belongs to the real
// Promotion Engine, BRB Part 3) — the caller resolves eligibility and hands
// this engine the resulting VND amount to apply, tagged with a Label purely
// for reporting.
type PromotionInput struct {
	DiscountVND int64
	Label       string
}

// TripInput is every signal BƯỚC 2 of the sprint brief asks this engine to
// accept. Pickup/Destination are free-form labels for reporting only — this
// engine does not geocode; IsAirportZone/PickupDistanceKM are what actually
// drive pricing, exactly as the production Route/Geo engines would resolve
// them upstream of Pricing in the real system.
type TripInput struct {
	PickupLabel      string
	DestinationLabel string

	VehicleType VehicleType
	DistanceKM  float64
	DurationMin float64 // total trip duration, minutes

	// SlowTrafficMin is the portion of DurationMin spent below
	// TimeFareSpeedThresholdKPH (BRB §2.2.3's "greater of time or distance"
	// rule). 0 means "not modeled" — the whole trip is billed as distance
	// fare, which is the correct default for a normal-traffic trip.
	SlowTrafficMin float64

	// WaitingMin is total minutes the driver waited after marking "Arrived",
	// before the BRB §2.2.9 grace period is subtracted.
	WaitingMin float64

	// PickupDistanceKM is the driver's distance to the pickup point at
	// dispatch time — drives the proposed Long Pickup Compensation
	// (PRICING_STRATEGY §2.2, NOT YET in BRB).
	PickupDistanceKM float64

	RequestTime time.Time // drives Night/Peak/weekday detection
	IsHoliday   bool
	Weather     WeatherCondition

	// DSR is the demand/supply ratio a real Dispatch/Geo engine would
	// compute per zone (BRB §2.13.2). 0 or any value < 1.2 means no surge.
	DSR float64

	IsAirportZone bool

	// BridgeFeeVND / ParkingFeeVND are pass-through costs the driver
	// declares — PRICING_STRATEGY §2.2 (Bridge) and BRB §2.2.8 (same
	// treatment as Toll). 100% to the driver, 0% commission.
	BridgeFeeVND  int64
	ParkingFeeVND int64

	Promotion *PromotionInput

	DriverTier DriverTier

	// PassengerLevel is accepted but has NO pricing effect in this engine.
	// BRB §10.5 explicitly defers rider loyalty tiers to a future phase
	// ("Loyalty tiers are not implemented at launch") — applying a discount
	// or surcharge by passenger level here would be inventing a BRB rule
	// that does not exist. The field exists so callers can pass it through
	// for reporting/future-readiness without this engine silently ignoring
	// a caller's intent versus never having accepted it at all.
	PassengerLevel string
}

// FareBreakdown is the fully itemised simulation result — covers every
// field BƯỚC 2 asks for, plus the intermediate figures needed to audit how
// each was derived (BRB §1.2 Principle 2: "rules are public").
type FareBreakdown struct {
	VehicleType VehicleType
	DistanceKM  float64
	DurationMin float64

	// Core metered components — BRB §2.2.1–§2.2.4.
	BaseFare     int64
	DistanceFare int64
	TimeFare     int64

	MeteredSubtotal   int64 // BaseFare + DistanceFare + TimeFare
	MinimumFareForced bool  // true if MeteredSubtotal was below the vehicle's minimum fare

	// Multipliers actually applied, for transparency.
	NightApplied     bool
	HolidayApplied   bool
	RainApplied      bool
	PeakApplied      bool // only ever true when SurgeMultiplier == 1.0 (BRB §2.2.12)
	StaticMultiplier float64
	SurgeMultiplier  float64
	PriceCapApplied  bool

	RideFare int64 // metered subtotal after minimum-fare floor + surge + static multiplier + price cap

	// Flat / pass-through add-ons.
	AirportFeeApplied      int64
	WaitingFee             int64
	BridgeFee              int64
	ParkingFee             int64
	LongPickupCompensation int64 // PRICING_STRATEGY proposal, platform-funded, NOT in BRB yet

	// Booking fee doubles as "Service Fee" per PRICING_STRATEGY §3.3.
	ServiceFee int64

	PromotionLabel    string
	PromotionRequested int64
	PromotionApplied  int64 // clamped so it can never exceed the pre-discount total (BRB §4.9)

	Commission int64
	VAT        int64 // ASSUMPTION — see AssumedVATRate
	Insurance  int64 // ASSUMPTION — see AssumedInsuranceCostVND, currently always 0

	DriverIncomeGross   int64 // driver's share of metered+airport+waiting, plus 100% of bridge/parking
	MinimumEarningTopUp int64 // BRB §2.14, platform-funded
	NetDriver           int64 // final driver payout, rounded up to DriverRoundingUnit

	CustomerTotal   int64 // final rider payout, rounded up to RiderRoundingUnit
	PlatformRevenue int64 // Commission + ServiceFee − PromotionApplied − TopUp − LongPickupCompensation − Insurance (pre-VAT)
	Profit          int64 // PlatformRevenue − VAT
	MarginPct       float64 // Profit / CustomerTotal, as a fraction (0.12 = 12%)

	Warnings []string
}

// Simulator computes fares against a configurable rate table — never the
// production FareConfig, so changing simulation parameters can never affect
// production, and vice versa.
type Simulator struct {
	Rates       map[VehicleType]VehicleRates
	PeakWindows []PeakWindow
	SurgeBands  []SurgeBand
}

// NewDefaultSimulator builds a Simulator from every BRB-sourced default in
// pricing_constants.go.
func NewDefaultSimulator() *Simulator {
	return &Simulator{
		Rates:       DefaultVehicleRates(),
		PeakWindows: DefaultPeakWindows(),
		SurgeBands:  DefaultSurgeBands(),
	}
}

// Simulate computes a full fare breakdown for one trip. It never returns a
// breakdown that violates a BƯỚC-7 safety invariant — see safety.go; any
// clamp applied is recorded in Warnings rather than silently hidden.
func (s *Simulator) Simulate(in TripInput) (*FareBreakdown, error) {
	if in.DistanceKM < 0 {
		return nil, domainerrors.InvalidArgument("distance_km must be non-negative")
	}
	if in.DurationMin < 0 {
		return nil, domainerrors.InvalidArgument("duration_minutes must be non-negative")
	}
	rates, ok := s.Rates[in.VehicleType]
	if !ok {
		return nil, domainerrors.InvalidArgument("unsupported vehicle type: " + string(in.VehicleType))
	}

	fb := &FareBreakdown{VehicleType: in.VehicleType, DistanceKM: in.DistanceKM, DurationMin: in.DurationMin}

	// ─── 1. Metered components (BRB §2.2.1–§2.2.3) ─────────────────────────
	fastKM := in.DistanceKM
	slowMin := clampMin(in.SlowTrafficMin, 0, in.DurationMin)
	fb.BaseFare = rates.BaseFare
	fb.DistanceFare = roundVND(float64(rates.PerKmRate) * fastKM)
	fb.TimeFare = roundVND(float64(rates.PerMinuteRate) * slowMin)
	fb.MeteredSubtotal = fb.BaseFare + fb.DistanceFare + fb.TimeFare

	// ─── 2. Minimum fare floor (BRB §2.2.4) — applied BEFORE surge/surcharges,
	// matching the production calculator's order and BRB §2.17B Example A. ──
	rideFareBase := fb.MeteredSubtotal
	if rideFareBase < rates.MinimumFare {
		rideFareBase = rates.MinimumFare
		fb.MinimumFareForced = true
	}

	// ─── 3. Dynamic surge (BRB §2.13) ───────────────────────────────────────
	fb.SurgeMultiplier = SurgeMultiplierForDSR(s.SurgeBands, in.DSR)

	// ─── 4. Static surcharges (BRB §2.2.10–§2.2.13) ─────────────────────────
	fb.NightApplied = isNightHour(in.RequestTime)
	fb.HolidayApplied = in.IsHoliday
	fb.RainApplied = in.Weather == WeatherRain
	// Peak never stacks with Surge (BRB §2.2.12) — only evaluated when surge is inactive.
	fb.PeakApplied = fb.SurgeMultiplier == 1.0 && isPeakHour(s.PeakWindows, in.RequestTime)

	static := 1.0
	if fb.NightApplied {
		static *= NightSurchargeMultiplier
	}
	if fb.HolidayApplied {
		static *= HolidaySurchargeMultiplier
	}
	if fb.RainApplied {
		static *= RainSurchargeMultiplier
	}
	if fb.PeakApplied {
		static *= PeakHourSurchargeMultiplier
	}
	if static > StaticSurchargeCap {
		static = StaticSurchargeCap
	}
	fb.StaticMultiplier = static

	rideFare := roundVND(float64(rideFareBase) * fb.SurgeMultiplier * fb.StaticMultiplier)

	// ─── 5. Airport / waiting (BRB §2.2.7, §2.2.9) ──────────────────────────
	if in.IsAirportZone {
		fb.AirportFeeApplied = AirportFee
	}
	chargeableWaitMin := clampMin(in.WaitingMin-WaitingGraceMinutes, 0, in.WaitingMin)
	fb.WaitingFee = roundVND(float64(WaitingFeePerMinute) * chargeableWaitMin)

	// ─── 6. Price cap (BRB §2.13.6) — applied to the metered/surge component
	// only; flat add-ons (airport/waiting/booking/pass-through) are not
	// scaled by surge in the first place so they are excluded from the cap
	// pool by construction. ───────────────────────────────────────────────
	if rideFare > PriceCapVND {
		rideFare = PriceCapVND
		fb.PriceCapApplied = true
	}
	fb.RideFare = rideFare

	// ─── 7. Pass-through add-ons (BRB §2.2.8; PRICING_STRATEGY Bridge/Parking/Long-Pickup) ──
	fb.BridgeFee = maxInt64(in.BridgeFeeVND, 0)
	fb.ParkingFee = maxInt64(in.ParkingFeeVND, 0)
	fb.LongPickupCompensation = longPickupCompensation(in.PickupDistanceKM)

	fb.ServiceFee = BookingFee

	// ─── 8. Promotion (BRB §4.9 / §6.5 Case A: platform-funded, driver
	// unaffected, discount clamped so the rider never pays < 0). ────────────
	preDiscountTotal := fb.RideFare + fb.AirportFeeApplied + fb.WaitingFee + fb.BridgeFee + fb.ParkingFee + fb.ServiceFee
	if in.Promotion != nil {
		fb.PromotionLabel = in.Promotion.Label
		fb.PromotionRequested = in.Promotion.DiscountVND
		fb.PromotionApplied = clampInt64(in.Promotion.DiscountVND, 0, preDiscountTotal)
		if fb.PromotionRequested > fb.PromotionApplied {
			fb.Warnings = append(fb.Warnings, "promotion discount exceeded trip value — clamped per BRB §4.9")
		}
	}

	fb.CustomerTotal = roundUpToUnit(preDiscountTotal-fb.PromotionApplied, RiderRoundingUnit)

	// ─── 9. Driver income (BRB §6.4, §2.2.6, §2.2.7, §2.2.9) ────────────────
	commissionBase := fb.RideFare + fb.AirportFeeApplied + fb.WaitingFee
	rate := CommissionRate(in.DriverTier)
	fb.Commission = roundVND(float64(commissionBase) * rate)
	driverShareOfMetered := commissionBase - fb.Commission
	fb.DriverIncomeGross = driverShareOfMetered + fb.BridgeFee + fb.ParkingFee

	// ─── 10. Minimum earning guarantee (BRB §2.14) ──────────────────────────
	if fb.DriverIncomeGross < MinimumDriverEarningVND {
		fb.MinimumEarningTopUp = MinimumDriverEarningVND - fb.DriverIncomeGross
	}
	netDriverPreRound := fb.DriverIncomeGross + fb.MinimumEarningTopUp + fb.LongPickupCompensation
	fb.NetDriver = roundUpToUnit(netDriverPreRound, DriverRoundingUnit)

	// ─── 11. Platform revenue → VAT → Profit ────────────────────────────────
	platformRevenueGross := fb.Commission + fb.ServiceFee
	platformAfterPromo := platformRevenueGross - fb.PromotionApplied
	platformAfterDriverProtection := platformAfterPromo - fb.MinimumEarningTopUp - fb.LongPickupCompensation - AssumedInsuranceCostVND
	fb.Insurance = AssumedInsuranceCostVND

	vatBase := platformAfterDriverProtection
	if vatBase < 0 {
		vatBase = 0 // VAT never applies to a negative base
	}
	fb.VAT = roundVND(float64(vatBase) * AssumedVATRate)
	fb.PlatformRevenue = platformAfterDriverProtection
	fb.Profit = platformAfterDriverProtection - fb.VAT
	if fb.CustomerTotal > 0 {
		fb.MarginPct = float64(fb.Profit) / float64(fb.CustomerTotal)
	}

	applySafetyClamps(fb)

	return fb, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────

// SurgeMultiplierForDSR looks up the BRB §2.13.2 band for a demand/supply
// ratio, capped at MaxSurgeMultiplier (defensive — the default table
// already tops out at exactly that cap).
func SurgeMultiplierForDSR(bands []SurgeBand, dsr float64) float64 {
	for _, b := range bands {
		if b.MaxDSR < 0 || dsr < b.MaxDSR {
			if b.Multiplier > MaxSurgeMultiplier {
				return MaxSurgeMultiplier
			}
			return b.Multiplier
		}
	}
	return MaxSurgeMultiplier
}

func isNightHour(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	h := t.Hour()
	return h >= NightStartHour || h < NightEndHour
}

func isPeakHour(windows []PeakWindow, t time.Time) bool {
	if t.IsZero() {
		return false
	}
	weekday := t.Weekday() >= time.Monday && t.Weekday() <= time.Friday
	h := t.Hour()
	for _, w := range windows {
		if w.WeekdayOnly && !weekday {
			continue
		}
		if h >= w.StartHour && h < w.EndHour {
			return true
		}
	}
	return false
}

// longPickupCompensation implements the PRICING_STRATEGY §2.2 proposal
// (NOT YET in BRB — see pricing_constants.go doc comment).
func longPickupCompensation(pickupDistanceKM float64) int64 {
	switch {
	case pickupDistanceKM > LongPickupFarKM:
		return LongPickupCompensationFarVND
	case pickupDistanceKM > LongPickupNearKM:
		return LongPickupCompensationNearVND
	default:
		return 0
	}
}

func roundVND(v float64) int64 {
	if v < 0 {
		return -roundVND(-v)
	}
	return int64(v + 0.5)
}

// roundUpToUnit rounds amount UP to the nearest multiple of unit —
// BRB §2.15 (rider totals round up to 500 VND, driver payouts round up to
// 100 VND, "so the driver never loses money due to rounding").
func roundUpToUnit(amount, unit int64) int64 {
	if unit <= 0 {
		return amount
	}
	if amount <= 0 {
		return 0
	}
	rem := amount % unit
	if rem == 0 {
		return amount
	}
	return amount + (unit - rem)
}

// clampInt64 clamps an int64 amount into [lo, hi] — used for money values
// (VND has no fractional subunit, BRB §2.16), as distinct from clampMin
// which operates on the float64 distance/duration/minute inputs.
func clampInt64(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if hi >= lo && v > hi {
		return hi
	}
	return v
}

func clampMin(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if hi >= lo && v > hi {
		return hi
	}
	return v
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

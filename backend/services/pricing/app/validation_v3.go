package app

import (
	"math"

	"github.com/fairride/pricing/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

// validateRideInput rejects malformed input BEFORE any arithmetic runs —
// sprint brief PHẦN 11: distance âm, NaN, Infinity are all checked here
// rather than left to silently propagate into the fare formula and produce
// a nonsensical (but not obviously wrong-looking) result.
func validateRideInput(input entity.RideInputV3) error {
	if err := requireFiniteNonNegative("distance_km", input.DistanceKM); err != nil {
		return err
	}
	if err := requireFiniteNonNegative("duration_minutes", input.DurationMin); err != nil {
		return err
	}
	if err := requireFinite("slow_traffic_min", input.SlowTrafficMin); err != nil {
		return err
	}
	if input.SlowTrafficMin < 0 {
		return domainerrors.InvalidArgument("slow_traffic_min must be non-negative")
	}
	if err := requireFiniteNonNegative("waiting_min", input.WaitingMin); err != nil {
		return err
	}
	if input.VoucherDiscountVND < 0 {
		return domainerrors.InvalidArgument("voucher_discount_vnd must be non-negative")
	}
	return nil
}

func requireFinite(field string, v float64) error {
	if math.IsNaN(v) {
		return domainerrors.InvalidArgument(field + " must not be NaN")
	}
	if math.IsInf(v, 0) {
		return domainerrors.InvalidArgument(field + " must not be Infinity")
	}
	return nil
}

func requireFiniteNonNegative(field string, v float64) error {
	if err := requireFinite(field, v); err != nil {
		return err
	}
	if v < 0 {
		return domainerrors.InvalidArgument(field + " must be non-negative")
	}
	return nil
}

// ValidateFullBreakdown is the sprint's PHẦN 11 safety net, run on every
// FullFareBreakdownV3 FareCalculatorV3 produces before returning it to a
// caller — a defensive re-check on the OUTPUT, independent of
// validateRideInput's check on the INPUT, so a bug introduced anywhere in
// the calculation (not just malformed input) is still caught here rather
// than silently reaching a caller.
func ValidateFullBreakdown(fb *entity.FullFareBreakdownV3) error {
	if fb.FinalFare < 0 {
		return domainerrors.Internal("computed final fare is negative — this is a calculator bug, not a caller input error")
	}
	if fb.RideFare < 0 {
		return domainerrors.Internal("computed ride fare is negative")
	}
	if fb.Commission < 0 {
		return domainerrors.Internal("computed commission is negative")
	}
	if fb.CommissionRate < 0 || fb.CommissionRate > 1.0 {
		return domainerrors.Internal("commission rate must be within [0, 1.0] (i.e. 0-100%)")
	}
	if fb.VoucherDiscount > fb.RideFare+fb.WaitingFee+fb.PlatformFee {
		return domainerrors.Internal("voucher discount exceeds the pre-discount total — clamp logic did not apply")
	}
	if fb.VAT < 0 {
		return domainerrors.Internal("computed VAT is negative")
	}
	if fb.DriverIncome < 0 {
		return domainerrors.Internal("computed driver income is negative")
	}
	// Overflow guard: every monetary field here is derived from
	// DistanceKM/DurationMin x configured rates. A trip long enough to
	// approach math.MaxInt64 VND (>10^18 VND, i.e. tens of billions of km at
	// any realistic rate) indicates upstream bad data, not a real trip.
	const overflowGuard = 1_000_000_000_000_000 // 1 quadrillion VND — no real Panda trip approaches this
	if fb.FinalFare > overflowGuard || fb.RideFare > overflowGuard {
		return domainerrors.InvalidArgument("computed fare exceeds a sane upper bound — check distance/duration input for a data error")
	}
	return nil
}

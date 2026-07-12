// Pricing V3 fare calculator — docs/business/PRICING_V3_DESIGN.md,
// docs/business/PRICING_V3_REVIEW.md. Reuses the existing Dynamic Pricing
// Engine (PricingPipeline / PricingEvaluator / PricingRule / RuleConfig /
// PricingContext / PricingResult, all unmodified by this file — see
// pricing_pipeline_v3.go for the one additive pipeline constructor) for
// every surge/surcharge decision, exactly as required by the sprint's PHẦN 1
// ("vẫn giữ nguyên kiến trúc. Không quay về if-else."). What V3 adds on top
// is: degressive Distance Tier (distance_tier.go), a Traffic/Waiting time
// split, leg-specific Airport Fee (rules_airport_v3.go), and — new to
// production — Commission/VAT/Driver-Income/Platform-Revenue.
package app

import (
	"time"

	"github.com/fairride/pricing/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

// FareCalculatorV3 is the Pricing V3 counterpart to FareCalculator (V2,
// fare_calculator.go — completely unmodified by this file). Stateless and
// safe for concurrent use, matching V2's contract.
type FareCalculatorV3 struct {
	config        entity.FareConfigV3
	airportConfig entity.AirportFeeConfigV3
	commission    entity.CommissionConfigV3
	vatRate       float64
	ruleConfigs   RuleConfigMap
}

// NewFareCalculatorV3 builds a V3 calculator from config loaded via
// backend/services/pricing/config (or, in tests, a hand-built config) — no
// argument here is ever a Go literal in production wiring (cmd/server/main.go).
// ruleConfigs governs the reused surge engine (Demand/Peak/Night/Holiday/
// Rain/Airport) exactly as it does for V2; pass DefaultRuleConfigsV3() for
// BRB's stated structural defaults.
func NewFareCalculatorV3(
	config entity.FareConfigV3,
	airportConfig entity.AirportFeeConfigV3,
	commission entity.CommissionConfigV3,
	vatRate float64,
	ruleConfigs RuleConfigMap,
) *FareCalculatorV3 {
	if config.Rates == nil {
		panic("pricing: FareConfigV3.Rates must not be nil")
	}
	if ruleConfigs == nil {
		ruleConfigs = DefaultRuleConfigs()
	}
	return &FareCalculatorV3{
		config:        config,
		airportConfig: airportConfig,
		commission:    commission,
		vatRate:       vatRate,
		ruleConfigs:   ruleConfigs,
	}
}

// EstimateV3 calculates an upfront fare before the trip starts.
func (c *FareCalculatorV3) EstimateV3(input entity.RideInputV3) (*entity.FullFareBreakdownV3, error) {
	return c.calculate(input, false)
}

// CalculateFinalV3 calculates the fare after trip completion using actual
// distance/time/waiting. Same formula as EstimateV3 — no deviation penalty,
// matching V2's upfront-pricing guarantee.
func (c *FareCalculatorV3) CalculateFinalV3(input entity.RideInputV3) (*entity.FullFareBreakdownV3, error) {
	return c.calculate(input, true)
}

func (c *FareCalculatorV3) calculate(input entity.RideInputV3, isFinal bool) (*entity.FullFareBreakdownV3, error) {
	if err := validateRideInput(input); err != nil {
		return nil, err
	}

	rates, ok := c.config.Rates[input.VehicleType]
	if !ok {
		return nil, domainerrors.InvalidArgument("unsupported vehicle type: " + string(input.VehicleType))
	}

	requestTime := input.RequestTime
	if requestTime.IsZero() {
		requestTime = time.Now()
	}

	// ─── 1. Metered components (PRICING_V3_DESIGN.md Phần 4/5) ────────────
	baseFare := rates.BaseFare
	distanceFare := entity.DistanceFareForTiers(rates.DistanceTiers, input.DistanceKM)
	slowMin := clampFloatToRange(input.SlowTrafficMin, 0, input.DurationMin)
	trafficTimeFare := roundToUnit(float64(rates.TrafficTimePerMinute) * slowMin)
	const movingTimeFare = 0 // always — see FullFareBreakdownV3.MovingTimeFare doc comment

	metered := baseFare + distanceFare + movingTimeFare + trafficTimeFare

	// ─── 2. Surge + Airport, via the EXISTING Rule Engine ──────────────────
	pipeline := NewDefaultPricingPipelineV3(c.airportConfig, input.VehicleType)
	ctx := entity.PricingContext{
		VehicleType:      input.VehicleType,
		RequestTime:      requestTime,
		ActiveRequests:   input.ActiveRequests,
		AvailableDrivers: input.AvailableDrivers,
		IsAirportZone:    input.AirportLeg != entity.AirportLegNone,
		AirportLeg:       input.AirportLeg,
		IsRainActive:     input.IsRainActive,
		IsHoliday:        input.IsHoliday,
	}
	surge := pipeline.Evaluate(ctx, c.ruleConfigs)

	surchargeableBase := metered + surge.FlatSurcharge
	rideFareRaw := roundToUnit(float64(surchargeableBase) * surge.FinalMultiplier)

	minimumFareForced := rideFareRaw < rates.MinimumFare
	rideFare := rideFareRaw
	if minimumFareForced {
		rideFare = rates.MinimumFare
	}

	// ─── 3. Waiting Fee — never surged, never counted toward Minimum Fare
	// (PRICING_V3_DESIGN.md Phần 5.1: a pre-trip cost, different in kind from
	// in-trip Traffic Time) ─────────────────────────────────────────────────
	graceMin := float64(rates.WaitingGraceMinutes)
	chargeableWaitMin := clampFloatToRange(input.WaitingMin-graceMin, 0, input.WaitingMin)
	waitingFee := roundToUnit(float64(rates.WaitingFeePerMinute) * chargeableWaitMin)

	platformFee := rates.BookingFee
	preDiscountTotal := rideFare + waitingFee + platformFee

	// ─── 4. Voucher — pre-resolved by the Promotion Engine, applied here
	// (BRB §4.9: never below 0, and never above the pre-discount total —
	// note this is a plain [0, preDiscountTotal] clamp, deliberately NOT the
	// pricing_evaluator.go clampInt64 helper, whose "hi<=0 means uncapped"
	// convention is specific to RuleConfig.MaxSurge and would be wrong here:
	// a preDiscountTotal of exactly 0 must still cap the voucher at 0) ───────
	voucherDiscount := clampToRange(input.VoucherDiscountVND, 0, preDiscountTotal)
	finalFare := preDiscountTotal - voucherDiscount

	// ─── 5. Commission / Driver Income / VAT / Platform Revenue (new
	// production responsibility — see commission_v3.go) ─────────────────────
	commissionRate := c.commission.Rate(input.CommissionTier)
	commissionBase := rideFare + waitingFee // BRB §2.2.6/§2.2.9: airport fee (already inside rideFare) and waiting fee are commission-bearing; booking fee/voucher are not
	commission := roundToUnit(float64(commissionBase) * commissionRate)
	driverIncome := commissionBase - commission

	platformRevenueGross := commission + platformFee
	platformAfterVoucher := platformRevenueGross - voucherDiscount
	vatBase := platformAfterVoucher
	if vatBase < 0 {
		vatBase = 0
	}
	vat := roundToUnit(float64(vatBase) * c.vatRate)
	platformRevenue := platformAfterVoucher - vat

	fb := &entity.FullFareBreakdownV3{
		VehicleType:  input.VehicleType,
		DistanceKM:   input.DistanceKM,
		DurationMin:  input.DurationMin,
		CurrencyCode: c.config.CurrencyCode,
		IsFinal:      isFinal,

		BaseFare:        baseFare,
		DistanceFare:    distanceFare,
		MovingTimeFare:  movingTimeFare,
		TrafficTimeFare: trafficTimeFare,
		WaitingFee:      waitingFee,

		AirportFee: surge.FlatSurcharge,
		AirportLeg: input.AirportLeg,

		RideFare:          rideFare,
		MinimumFareForced: minimumFareForced,

		SurgeMultiplier: surge.FinalMultiplier,
		SurgeLabel:      surge.Label,

		VoucherLabel:     input.VoucherLabel,
		VoucherRequested: input.VoucherDiscountVND,
		VoucherDiscount:  voucherDiscount,

		CommissionTier: input.CommissionTier,
		CommissionRate: commissionRate,
		Commission:     commission,
		PlatformFee:    platformFee,
		VATRate:        c.vatRate,
		VAT:            vat,

		DriverIncome:    driverIncome,
		PlatformRevenue: platformRevenue,
		FinalFare:       finalFare,
	}

	if err := ValidateFullBreakdown(fb); err != nil {
		return nil, err
	}
	return fb, nil
}

// clampToRange restricts v to [lo, hi] unconditionally (unlike
// pricing_evaluator.go's clampInt64, whose hi<=0 sentinel means "no cap" —
// a convention specific to RuleConfig.MaxSurge, not appropriate for clamping
// a voucher discount against an actual fare total that can legitimately be 0).
func clampToRange(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clampFloatToRange is clampToRange's float64 counterpart — same
// unconditional-bounds semantics, distinct from pricing_evaluator.go's
// clampFloat (whose lo<=0/hi<=0 sentinels mean "unbounded", correct for
// RuleConfig.MinSurge/MaxSurge but wrong for a distance/duration value where
// 0 is a real, meaningful lower bound).
func clampFloatToRange(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

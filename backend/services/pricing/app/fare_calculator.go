// Package app contains the fare calculation business logic for FAIRRIDE.
//
// FareCalculator is now internally powered by a Dynamic Pricing Engine (see
// pricing_rule.go / pricing_evaluator.go / pricing_pipeline.go /
// rules_defined.go / rules_todo.go): a Rule Engine supporting Demand Surge,
// Supply Surge, Peak Hour, Airport, Rain, Holiday, Traffic, Night, and
// Special Event as independently enable/disable/priority/weight/min-max
// configurable rules, combined per Business Rule Bible v1.0 Part 2 (§2.2.7,
// §2.2.10-§2.2.13, §2.13).
//
// Estimate and CalculateFinal — the two methods the gRPC handler calls —
// remain 100% backward compatible: every rule ships DISABLED by default
// (see DefaultRuleConfigs), so the pipeline always returns a neutral result
// (multiplier 1.0, no flat surcharge) and these two methods produce byte-
// for-byte identical output to before this refactor. EstimateWithContext and
// CalculateFinalWithContext are new, additive methods for callers that want
// to actually exercise the dynamic pricing engine; they are not yet wired to
// the gRPC API (no protobuf change was made in this refactor).
package app

import (
	"math"
	"time"

	"github.com/fairride/pricing/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

// FareCalculator calculates ride fares from a FareConfig, optionally
// surged by a PricingPipeline. It is stateless and safe for concurrent use.
type FareCalculator struct {
	config      entity.FareConfig
	pipeline    *PricingPipeline
	ruleConfigs RuleConfigMap
}

// NewFareCalculator creates a FareCalculator with the given config, wired to
// the default Dynamic Pricing Engine with every surge rule DISABLED — this
// is the exact configuration that guarantees Estimate/CalculateFinal are
// unchanged from the pre-refactor calculator.
// Panics if config.Rates is nil (misconfiguration).
func NewFareCalculator(config entity.FareConfig) *FareCalculator {
	if config.Rates == nil {
		panic("pricing: FareConfig.Rates must not be nil")
	}
	return &FareCalculator{
		config:      config,
		pipeline:    NewDefaultPricingPipeline(),
		ruleConfigs: DefaultRuleConfigs(),
	}
}

// NewFareCalculatorWithPipeline creates a FareCalculator with a caller-
// supplied pipeline and rule configuration — e.g. to enable specific surge
// rules for a city, or to inject a test pipeline with a subset of rules.
// A nil pipeline or ruleConfigs falls back to the same disabled-by-default
// engine NewFareCalculator uses.
// Panics if config.Rates is nil (misconfiguration).
func NewFareCalculatorWithPipeline(config entity.FareConfig, pipeline *PricingPipeline, ruleConfigs RuleConfigMap) *FareCalculator {
	if config.Rates == nil {
		panic("pricing: FareConfig.Rates must not be nil")
	}
	if pipeline == nil {
		pipeline = NewDefaultPricingPipeline()
	}
	if ruleConfigs == nil {
		ruleConfigs = DefaultRuleConfigs()
	}
	return &FareCalculator{config: config, pipeline: pipeline, ruleConfigs: ruleConfigs}
}

// Estimate calculates an upfront fare before the trip starts.
// The same formula is used for both estimates and final fares (upfront pricing guarantee).
//
// Returns CodeInvalidArgument for unknown vehicle types or negative inputs.
func (c *FareCalculator) Estimate(vehicleType entity.VehicleType, distanceKM, durationMin float64) (*entity.FareBreakdown, error) {
	return c.calculate(vehicleType, distanceKM, durationMin, false, entity.NeutralContext(vehicleType, time.Now()))
}

// CalculateFinal calculates the fare after trip completion using actual distance and time.
// Uses the same formula as Estimate — no deviation penalty in MVP.
//
// Returns CodeInvalidArgument for unknown vehicle types or negative inputs.
func (c *FareCalculator) CalculateFinal(vehicleType entity.VehicleType, distanceKM, durationMin float64) (*entity.FareBreakdown, error) {
	return c.calculate(vehicleType, distanceKM, durationMin, true, entity.NeutralContext(vehicleType, time.Now()))
}

// EstimateWithContext is Estimate, but driven by the Dynamic Pricing Engine
// using a caller-supplied PricingContext (real-time demand/supply, weather,
// holiday-calendar signals, etc). Additive method — does not affect Estimate.
func (c *FareCalculator) EstimateWithContext(vehicleType entity.VehicleType, distanceKM, durationMin float64, ctx entity.PricingContext) (*entity.FareBreakdown, error) {
	return c.calculate(vehicleType, distanceKM, durationMin, false, ctx)
}

// CalculateFinalWithContext is CalculateFinal, but driven by the Dynamic
// Pricing Engine using a caller-supplied PricingContext. Additive method —
// does not affect CalculateFinal.
func (c *FareCalculator) CalculateFinalWithContext(vehicleType entity.VehicleType, distanceKM, durationMin float64, ctx entity.PricingContext) (*entity.FareBreakdown, error) {
	return c.calculate(vehicleType, distanceKM, durationMin, true, ctx)
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (c *FareCalculator) calculate(vehicleType entity.VehicleType, distanceKM, durationMin float64, isFinal bool, ctx entity.PricingContext) (*entity.FareBreakdown, error) {
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

	// BRB §2.13.4: surge multiplier applies to (Base + Distance + Time +
	// Airport Fee); Booking Fee is added afterward, never surged.
	surge := c.pipeline.Evaluate(ctx, c.ruleConfigs)
	surchargeableBase := baseFare + distanceFare + timeFare + surge.FlatSurcharge
	rideFare := roundToUnit(float64(surchargeableBase) * surge.FinalMultiplier)

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

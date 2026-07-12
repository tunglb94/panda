package app

import "github.com/fairride/pricing/domain/entity"

// AirportFeeRuleV3 is the Pricing V3 replacement for AirportFeeRule
// (rules_defined.go) — same PricingRule interface, same CategoryFlatFee, same
// RuleConfigMap/PricingEvaluator plumbing (PHẦN 1 of the sprint: "vẫn giữ
// nguyên kiến trúc, không quay về if-else"). The only difference is WHERE
// the fee amount comes from: V2's AirportFeeRule always returns the single
// package-constant entity.AirportFeeVND; V3 looks up a config-driven,
// leg-specific, vehicle-specific amount (PRICING_V3_DESIGN.md Phần 7),
// fixing the bias docs/business/MARKET_PRICING_RESEARCH.md Phần 1.3 found
// (a flat fee charged identically to a motorcycle and a car).
//
// This is an ADDITIVE new rule, not a modification of AirportFeeRule — V2
// pipelines (NewDefaultPricingPipeline) are completely untouched and keep
// using AirportFeeRule exactly as before.
type AirportFeeRuleV3 struct {
	config      entity.AirportFeeConfigV3
	vehicleType entity.VehicleType
}

// NewAirportFeeRuleV3 builds a rule scoped to one vehicle type (a
// FareCalculatorV3 constructs one per Estimate/CalculateFinal call, since
// the fee amount depends on which vehicle is being priced — RuleConfig has
// no notion of "current vehicle type", so this is threaded through the
// rule's constructor instead, matching how NewDemandSurgeRule/NewPeakHourRule
// thread their own structural config through their constructors).
func NewAirportFeeRuleV3(config entity.AirportFeeConfigV3, vehicleType entity.VehicleType) *AirportFeeRuleV3 {
	return &AirportFeeRuleV3{config: config, vehicleType: vehicleType}
}

func (r *AirportFeeRuleV3) Name() string                  { return RuleNameAirport }
func (r *AirportFeeRuleV3) Category() entity.RuleCategory { return entity.CategoryFlatFee }

func (r *AirportFeeRuleV3) Evaluate(ctx entity.PricingContext) entity.RuleOutcome {
	if ctx.AirportLeg == entity.AirportLegNone {
		return entity.RuleOutcome{
			RuleName: r.Name(), Category: r.Category(), Applied: false,
			Reason: "trip has no airport pickup/dropoff leg",
		}
	}
	fee := r.config.FeeFor(r.vehicleType, ctx.AirportLeg)
	if fee <= 0 {
		return entity.RuleOutcome{
			RuleName: r.Name(), Category: r.Category(), Applied: false,
			Reason: "no airport fee configured for this vehicle type/leg (PRICING_V3_DESIGN.md Phần 7 — e.g. motorcycle is deliberately 0)",
		}
	}
	label := "Airport"
	if ctx.AirportLeg == entity.AirportLegPickup {
		label = "Airport Pickup"
	} else if ctx.AirportLeg == entity.AirportLegDropoff {
		label = "Airport Dropoff"
	}
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: true,
		FlatAmount: fee, Label: label,
		Reason: "PRICING_V3_DESIGN.md Phần 7 airport pickup/dropoff fee",
	}
}

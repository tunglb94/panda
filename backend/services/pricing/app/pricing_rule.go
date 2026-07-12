package app

import "github.com/fairride/pricing/domain/entity"

// PricingRule is one pluggable surge factor. Each of the 9 requested surge
// types (Demand Surge, Supply Surge, Peak Hour, Airport, Rain, Holiday,
// Traffic, Night, Special Event) is exactly one PricingRule implementation —
// see rules_defined.go for the 6 with a real BRB-sourced formula and
// rules_todo.go for the 3 with no approved formula yet.
//
// A rule never decides on its own whether it is "on" — that is RuleConfig's
// job (Enabled/Priority/Weight/MinSurge/MaxSurge, all caller-supplied, no
// hardcoded toggles). A rule only answers: given this context and this
// config, what do I contribute?
type PricingRule interface {
	// Name identifies the rule (used as the RuleConfig map key and in
	// PricingResult.AppliedRules/SkippedRules for transparency).
	Name() string

	// Category governs how PricingPipeline combines this rule's output with
	// its siblings (see entity.RuleCategory).
	Category() entity.RuleCategory

	// Evaluate returns this rule's raw contribution for ctx, before
	// RuleConfig.Weight/MinSurge/MaxSurge are applied by PricingEvaluator.
	// Implementations must be pure and side-effect free.
	Evaluate(ctx entity.PricingContext) entity.RuleOutcome
}

// RuleConfig is the operator-facing configuration surface for one
// PricingRule — exactly the knobs requested: enable, disable, priority,
// weight, max surge, min surge. No rule's activation logic is hardcoded
// outside of this struct's fields.
type RuleConfig struct {
	Enabled bool

	// Priority orders evaluation (lower = evaluated first). Rules within the
	// same Category are commutative (multiplication / addition), so Priority
	// does not change the numeric result — it only controls the order rules
	// appear in AppliedRules/SkippedRules and, for CategoryDynamicSurge vs
	// Peak Hour exclusion, which one is evaluated as "the" dynamic signal
	// when more than one dynamic-surge rule is enabled simultaneously.
	Priority int

	// Weight scales how much of the rule's raw effect is applied:
	//   multiplier categories: appliedMultiplier = 1 + (rawMultiplier-1)*Weight
	//   flat_fee category:     appliedFlatAmount = rawFlatAmount * Weight
	// Weight=1.0 reproduces the BRB-sourced value exactly; Weight=0.5 halves
	// the rule's effect without touching its formula. Weight is an operator
	// dial, not a business-approved number in itself.
	Weight float64

	// MinSurge/MaxSurge clamp the rule's contribution after Weight is
	// applied: multiplier categories clamp the multiplier; flat_fee clamps
	// the flat amount. For CategoryDynamicSurge, MaxSurge can only tighten
	// BRB's hard x2.0 ceiling (§2.13.3), never loosen it — see DemandSurgeRule.
	MinSurge float64
	MaxSurge float64
}

// RuleConfigMap maps each rule's Name() to its configuration.
type RuleConfigMap map[string]RuleConfig

// DefaultRuleConfigs returns every rule disabled (Enabled: false), Weight 1.0,
// and Min/Max wide open. This is the configuration FareCalculator uses for
// its existing Estimate/CalculateFinal methods — with every rule disabled,
// PricingPipeline always returns a neutral PricingResult (multiplier 1.0,
// flat surcharge 0), which is what guarantees 100% backward-compatible
// output. Operators (or a future caller using EstimateWithContext) enable
// specific rules by copying this map and flipping Enabled per rule.
func DefaultRuleConfigs() RuleConfigMap {
	cfg := RuleConfig{Enabled: false, Priority: 0, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.MaxDemandSurgeMultiplier}
	flatCfg := RuleConfig{Enabled: false, Priority: 0, Weight: 1.0, MinSurge: 0, MaxSurge: float64(entity.AirportFeeVND)}

	return RuleConfigMap{
		RuleNameDemandSurge:  {Enabled: false, Priority: 10, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.MaxDemandSurgeMultiplier},
		RuleNameSupplySurge:  {Enabled: false, Priority: 20, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.MaxDemandSurgeMultiplier},
		RuleNamePeakHour:     {Enabled: false, Priority: 30, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.PeakHourSurchargeMultiplier},
		RuleNameNight:        {Enabled: false, Priority: 40, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.NightSurchargeMultiplier},
		RuleNameHoliday:      {Enabled: false, Priority: 50, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.HolidaySurchargeMultiplier},
		RuleNameRain:         {Enabled: false, Priority: 60, Weight: 1.0, MinSurge: 1.0, MaxSurge: entity.RainSurchargeMultiplier},
		RuleNameAirport:      flatCfg,
		RuleNameTraffic:      cfg,
		RuleNameSpecialEvent: cfg,
	}
}

// Get returns the config for ruleName, or a disabled zero-value config if
// unset (fail closed: an unconfigured rule never contributes).
func (m RuleConfigMap) Get(ruleName string) RuleConfig {
	if cfg, ok := m[ruleName]; ok {
		return cfg
	}
	return RuleConfig{Enabled: false, Weight: 1.0}
}

// Rule name constants — the single source of truth for RuleConfigMap keys
// and PricingRule.Name() return values, avoiding stringly-typed drift.
const (
	RuleNameDemandSurge  = "demand_surge"
	RuleNameSupplySurge  = "supply_surge"
	RuleNamePeakHour     = "peak_hour"
	RuleNameNight        = "night_surcharge"
	RuleNameHoliday      = "holiday_surcharge"
	RuleNameRain         = "rain_surcharge"
	RuleNameAirport      = "airport_fee"
	RuleNameTraffic      = "traffic_surge"
	RuleNameSpecialEvent = "special_event_surge"
)

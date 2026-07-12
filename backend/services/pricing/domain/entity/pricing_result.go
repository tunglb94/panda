package entity

// RuleCategory governs how a rule's contribution combines with its siblings
// in the PricingPipeline. Categories exist so the pipeline can apply BRB's
// combination rules (stacking, mutual exclusion, caps) generically instead
// of a long if/else keyed on rule name.
type RuleCategory string

const (
	// CategoryStaticSurcharge: Night, Holiday, Rain, Peak Hour. Combine by
	// multiplication, subject to BRB's combined caps (§2.2.11, §2.2.13).
	CategoryStaticSurcharge RuleCategory = "static_surcharge"

	// CategoryDynamicSurge: Demand Surge (and, once BRB defines one, Supply
	// Surge). Mutually exclusive with Peak Hour (BRB §2.2.12: "Peak Hour
	// Surcharge does not stack with Dynamic Surge... Dynamic Surge
	// supersedes"). Hard-capped alone at x2.0 (BRB §2.13.3).
	CategoryDynamicSurge RuleCategory = "dynamic_surge"

	// CategoryFlatFee: Airport. Additive VND amount, not a multiplier.
	CategoryFlatFee RuleCategory = "flat_fee"

	// CategoryNotDefined: Traffic, Special Event, Supply Surge today. Always
	// contributes nothing (Applied=false) — see app/rules_todo.go.
	CategoryNotDefined RuleCategory = "not_defined"
)

// RuleOutcome is what a single PricingRule produces evaluating one
// PricingContext.
type RuleOutcome struct {
	RuleName   string
	Category   RuleCategory
	Applied    bool    // false = condition not met, rule disabled, or TODO stub
	Multiplier float64 // multiplicative contribution; meaningful only if Applied && Category uses a multiplier
	FlatAmount int64   // additive contribution (smallest currency unit); meaningful only if Applied && Category == CategoryFlatFee
	Label      string  // BRB §2.13.5 rider-facing label, e.g. "Busy", "High demand"
	Reason     string  // why applied / not applied / TODO explanation
}

// PricingResult is PricingPipeline's final, combined output for one
// PricingContext.
type PricingResult struct {
	// FinalMultiplier is the single combined multiplier to apply to
	// (BaseFare + DistanceFare + TimeFare + FlatSurcharge), per BRB §2.13.4.
	// Always 1.0 when every rule is disabled or inapplicable.
	FinalMultiplier float64

	// FlatSurcharge is the combined additive amount (e.g. Airport Fee),
	// added to the surge-multiplied base per BRB §2.13.4. Always 0 when
	// Airport is disabled or not applicable.
	FlatSurcharge int64

	// Label is the dominant rider-facing surge label (from Dynamic Surge if
	// active, else Peak Hour if active, else empty). BRB §2.13.5 transparency.
	Label string

	AppliedRules []RuleOutcome // rules that actually contributed
	SkippedRules []RuleOutcome // rules that were disabled, or evaluated but did not apply
}

// Neutral reports whether this result has no effect on the base fare —
// used by tests/callers to assert backward compatibility.
func (r PricingResult) Neutral() bool {
	return r.FinalMultiplier == 1.0 && r.FlatSurcharge == 0
}

package app

import "github.com/fairride/pricing/domain/entity"

// This file implements the 3 requested surge types with NO approved BRB
// formula: Supply Surge, Traffic, Special Event. Each is a real,
// PricingRule-implementing Go type — fully wired into the pipeline and
// configurable like every other rule — but Evaluate always returns
// Applied=false with a reason explaining why, so the pipeline can never
// silently invent a discount or surcharge for these. This mirrors the same
// "reuse existing rule or write TODO, never invent a new one" discipline
// used for the Promotion Engine (see backend/services/promotion).
//
// TODO(business): once BRB defines a formula for one of these, replace its
// entry in NewDefaultRuleRegistry with a real implementation modeled on
// rules_defined.go, and add its constants to
// domain/entity/surge_constants.go.

// ─── Supply Surge ───────────────────────────────────────────────────────────

// SupplySurgeRule exists because the sprint brief lists "Supply Surge" as a
// distinct required type, but BRB §2.13.2 defines only ONE unified
// Demand-Supply Ratio (DSR = active requests / available drivers) — there is
// no BRB formula that reacts to driver supply alone, separate from that
// ratio. Low driver availability already raises the DSR (and therefore
// DemandSurgeRule's output) through the denominator. Implementing a
// second, supply-only multiplier would double-count the same signal and
// is not something BRB defines — so this stays a TODO stub.
type SupplySurgeRule struct{}

func NewSupplySurgeRule() *SupplySurgeRule { return &SupplySurgeRule{} }

func (r *SupplySurgeRule) Name() string                  { return RuleNameSupplySurge }
func (r *SupplySurgeRule) Category() entity.RuleCategory { return entity.CategoryNotDefined }

func (r *SupplySurgeRule) Evaluate(_ entity.PricingContext) entity.RuleOutcome {
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: false,
		Reason: "TODO: BRB §2.13.2 defines a single unified Demand-Supply Ratio, not a separate supply-only surge formula. Low driver supply already raises DemandSurgeRule's output via the DSR denominator. A distinct supply-only rule requires a new BRB-approved formula before it can contribute.",
	}
}

// ─── Traffic ─────────────────────────────────────────────────────────────

// TrafficSurgeRule exists because the sprint brief lists "Traffic" as a
// required type, but BRB has no traffic-based surge multiplier. BRB §2.2.3
// explicitly states traffic congestion is already priced through Time Fare
// ("In heavy traffic, a ride that covers 2 km may take 20 minutes... Time
// fare compensates for slow-moving conditions"). Adding a traffic multiplier
// on top would double-charge the same condition and contradicts BRB's
// stated rationale — so this stays a TODO stub.
type TrafficSurgeRule struct{}

func NewTrafficSurgeRule() *TrafficSurgeRule { return &TrafficSurgeRule{} }

func (r *TrafficSurgeRule) Name() string                  { return RuleNameTraffic }
func (r *TrafficSurgeRule) Category() entity.RuleCategory { return entity.CategoryNotDefined }

func (r *TrafficSurgeRule) Evaluate(_ entity.PricingContext) entity.RuleOutcome {
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: false,
		Reason: "TODO: BRB §2.2.3 already prices traffic congestion through Time Fare (more chargeable minutes for a slow trip). BRB defines no separate traffic surge multiplier; adding one would double-charge the same condition and requires a new BRB-approved rule first.",
	}
}

// ─── Special Event ───────────────────────────────────────────────────────

// SpecialEventRule exists because the sprint brief lists "Special Event" as
// a required type, but no BRB fare surge rule exists for it. BRB §3.2.6
// Festival Promotion is a discount campaign in the Promotion Engine, not a
// fare surcharge, and cannot be reused here without inventing a new,
// unapproved formula.
type SpecialEventRule struct{}

func NewSpecialEventRule() *SpecialEventRule { return &SpecialEventRule{} }

func (r *SpecialEventRule) Name() string                  { return RuleNameSpecialEvent }
func (r *SpecialEventRule) Category() entity.RuleCategory { return entity.CategoryNotDefined }

func (r *SpecialEventRule) Evaluate(_ entity.PricingContext) entity.RuleOutcome {
	return entity.RuleOutcome{
		RuleName: r.Name(), Category: r.Category(), Applied: false,
		Reason: "TODO: no BRB fare-surge rule exists for special events. BRB §3.2.6 Festival Promotion is a Promotion Engine discount campaign, not a fare surcharge, and is not a valid source for a surge formula here. Requires a new BRB-approved rule.",
	}
}

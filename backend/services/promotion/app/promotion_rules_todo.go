package app

import (
	"context"
	"time"

	"github.com/fairride/promotion/domain/entity"
)

// TODORule backs every PromotionType that the sprint brief requires the
// engine to "support" but that has no approved rule in Business Rule Bible
// v1.0, PRICING_STRATEGY.md, or ECONOMY_ENGINE.md. Per the explicit
// instruction "Nếu thiếu => chỉ ghi TODO. KHÔNG tự nghĩ rule mới", this rule
// NEVER grants a discount — IsEligible always returns false with a reason
// that names the missing rule source, so the engine fails closed instead of
// inventing a number.
//
// TODO(business): once one of these types gets an approved BRB rule, replace
// its registry entry in promotion_rule.go (NewDefaultRuleRegistry) with a
// real implementation modeled on promotion_rules_defined.go, and add its
// constants to domain/entity/promotion_constants.go.
type TODORule struct {
	promoType entity.PromotionType
	reason    string
}

func NewTODORule(t entity.PromotionType) *TODORule {
	return &TODORule{promoType: t, reason: todoReason(t)}
}

func (r *TODORule) Type() entity.PromotionType { return r.promoType }

func (r *TODORule) IsEligible(_ context.Context, _ *entity.Voucher, _ *entity.PromotionRequest, _ time.Time) (bool, string) {
	return false, r.reason
}

func todoReason(t entity.PromotionType) string {
	switch t {
	case entity.PromotionTypeComeback:
		return "TODO: Comeback promotion has no BRB-approved rule. PRICING_STRATEGY.md §7.2.4 proposes 30% off, max 25,000 VND, for riders inactive 30+ days with 3+ prior trips, but this is marked [MỚI] — not yet in Business Rule Bible v1.0. Requires formal BRB amendment before this can grant a discount."
	case entity.PromotionTypeStudent:
		return "TODO: Student promotion has no BRB-approved rule. PRICING_STRATEGY.md §7.2.2 proposes 10% off, max 5 trips/week, but this is marked [MỚI] — not yet in Business Rule Bible v1.0. Also requires a student-verification data source this engine does not have."
	case entity.PromotionTypeAirport:
		return "TODO: Airport promotion has no BRB-approved rule. Business Rule Bible v1.0 only defines an Airport Fee (surcharge, Part 2), not an Airport discount. ECONOMY_ENGINE.md marks this as [MỚI]."
	case entity.PromotionTypeNightRide:
		return "TODO: Night Ride promotion has no BRB-approved rule. ECONOMY_ENGINE.md marks this as [MỚI]. Business Rule Bible v1.0 only defines a Night surcharge (Part 2), not a Night discount."
	case entity.PromotionTypeNewCity:
		return "TODO: New City promotion has no rule defined anywhere in Business Rule Bible v1.0, PRICING_STRATEGY.md, or ECONOMY_ENGINE.md."
	case entity.PromotionTypeFlashSale:
		return "TODO: Flash Sale promotion has no rule defined anywhere in Business Rule Bible v1.0, PRICING_STRATEGY.md, or ECONOMY_ENGINE.md."
	default:
		return "TODO: no approved business rule exists for promotion type " + string(t)
	}
}

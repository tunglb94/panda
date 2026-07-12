package app

import (
	"context"
	"time"

	"github.com/fairride/promotion/domain/entity"
)

// PromotionRule decides TYPE-SPECIFIC eligibility — facts a generic
// VoucherValidator cannot know, such as "has this rider ever completed a
// trip" (First Ride) or "is today within 1 day of the rider's birthday"
// (Birthday). It does NOT re-check the generic fields (status/timing/budget/
// usage/city/vehicle/membership/min-order) — VoucherValidator already covers
// those. A PromotionRule also never computes a discount amount; discount
// math is generic (entity.ComputeDiscount) driven by the Voucher's own
// DiscountType/DiscountValue/MaxDiscount fields.
type PromotionRule interface {
	Type() entity.PromotionType

	// IsEligible returns (true, "") if req satisfies this promotion type's
	// eligibility criteria at time now, or (false, reason) otherwise. reason
	// is a short, human-readable explanation surfaced in PromotionResult.
	IsEligible(ctx context.Context, v *entity.Voucher, req *entity.PromotionRequest, now time.Time) (bool, string)
}

// RuleRegistry maps each PromotionType to its rule implementation.
type RuleRegistry map[entity.PromotionType]PromotionRule

// NewDefaultRuleRegistry wires every PromotionType the engine recognizes:
// BRB-defined types get their real rule (app/promotion_rules_defined.go),
// everything else gets the safe not-eligible TODO stub
// (app/promotion_rules_todo.go). This ensures Evaluate() never panics on an
// unregistered type and never silently grants a discount for a type BRB
// hasn't approved.
func NewDefaultRuleRegistry() RuleRegistry {
	reg := RuleRegistry{}
	for _, rule := range []PromotionRule{
		NewFirstRideRule(),
		NewBirthdayRule(),
		NewGoldenHourRule(),
		NewWeekendRule(),
		NewRainRule(),
		NewEventCampaignRule(),
		NewReferralRule(),
		NewManualCouponRule(),
		NewMembershipRule(),
	} {
		reg[rule.Type()] = rule
	}
	for _, t := range []entity.PromotionType{
		entity.PromotionTypeComeback,
		entity.PromotionTypeStudent,
		entity.PromotionTypeAirport,
		entity.PromotionTypeNightRide,
		entity.PromotionTypeNewCity,
		entity.PromotionTypeFlashSale,
	} {
		reg[t] = NewTODORule(t)
	}
	return reg
}

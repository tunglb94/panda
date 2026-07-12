package entity

// PromotionType identifies which promotion mechanic a Voucher instance implements.
//
// Rule source discipline (per project instruction): a type is only backed by
// real calculation logic if Business Rule Bible v1.0 (BRB) defines it. Types
// BRB does not define are wired into the engine (so the platform can create
// campaigns of that type) but their PromotionRule implementation never grants
// a discount — it returns TODO-not-eligible until BRB is amended. See
// app/promotion_rules_todo.go.
type PromotionType string

const (
	// Defined in BRB Part 3 — real eligibility + discount logic implemented.
	PromotionTypeFirstRide     PromotionType = "first_ride"     // BRB §3.2.1
	PromotionTypeBirthday      PromotionType = "birthday"       // BRB §3.2.2
	PromotionTypeGoldenHour    PromotionType = "golden_hour"    // BRB §3.2.3
	PromotionTypeWeekend       PromotionType = "weekend"        // BRB §3.2.4
	PromotionTypeRain          PromotionType = "rain"           // BRB §3.2.5
	PromotionTypeEventCampaign PromotionType = "event_campaign" // BRB §3.2.6 (Festival Promotion)
	PromotionTypeReferral      PromotionType = "referral"       // BRB §3.2.7
	PromotionTypeManualCoupon  PromotionType = "manual_coupon"  // BRB §3.2.9 (Coupon Campaign)

	// Requested by this sprint but NOT defined anywhere in BRB / PRICING_STRATEGY /
	// ECONOMY_ENGINE with an approved number. Wired as a type; PromotionRule
	// returns not-eligible with a TODO reason. See app/promotion_rules_todo.go.
	PromotionTypeComeback  PromotionType = "comeback"   // PRICING_STRATEGY §7.2.4 proposes 30%/25,000 VND — NOT BRB-approved
	PromotionTypeStudent   PromotionType = "student"    // PRICING_STRATEGY §7.2.2 proposes 10% — NOT BRB-approved
	PromotionTypeAirport   PromotionType = "airport"    // ECONOMY_ENGINE marks [MỚI] — BRB only defines an Airport Fee (surcharge), not an Airport discount
	PromotionTypeNightRide PromotionType = "night_ride" // ECONOMY_ENGINE marks [MỚI] — no BRB discount rule
	PromotionTypeNewCity   PromotionType = "new_city"   // not mentioned in BRB/PRICING_STRATEGY/ECONOMY_ENGINE
	PromotionTypeFlashSale PromotionType = "flash_sale" // not mentioned in BRB/PRICING_STRATEGY/ECONOMY_ENGINE

	// Membership is explicitly NOT a discount rule. ECONOMY_ENGINE §8.1: "Membership
	// không bao giờ thay đổi công thức giá cước — chỉ thay đổi dịch vụ đi kèm."
	// This type exists only so a campaign can be gated to a membership tier via
	// Voucher.Membership (an eligibility filter), never as an independent discount
	// generator. See MembershipRule in app/promotion_rules_defined.go.
	PromotionTypeMembership PromotionType = "membership"
)

// definedTypes lists promotion types with a real, BRB-sourced discount rule.
var definedTypes = map[PromotionType]bool{
	PromotionTypeFirstRide:     true,
	PromotionTypeBirthday:      true,
	PromotionTypeGoldenHour:    true,
	PromotionTypeWeekend:       true,
	PromotionTypeRain:          true,
	PromotionTypeEventCampaign: true,
	PromotionTypeReferral:      true,
	PromotionTypeManualCoupon:  true,
}

// IsDefinedInBRB reports whether t has an approved, BRB-sourced rule.
// Types that return false are architecturally supported but functionally
// TODO — they must never silently grant a discount.
func (t PromotionType) IsDefinedInBRB() bool {
	return definedTypes[t]
}

// AllPromotionTypes returns every promotion type the engine recognizes,
// in BRB Campaign Priority order (§3.4) followed by the non-BRB / TODO types.
func AllPromotionTypes() []PromotionType {
	return []PromotionType{
		PromotionTypeManualCoupon,  // vouchers/coupons — highest priority (§3.4 #1)
		PromotionTypeReferral,      // §3.4 #2
		PromotionTypeFirstRide,     // §3.4 #3
		PromotionTypeEventCampaign, // §3.4 #4 (Festival)
		PromotionTypeGoldenHour,    // §3.4 #5 (Golden Hour / Rain / Weekend — highest value wins)
		PromotionTypeRain,          // §3.4 #5
		PromotionTypeWeekend,       // §3.4 #5
		PromotionTypeBirthday,      // §3.4 #6
		// Cashback (§3.4 #7) is post-trip and out of scope for this engine — see
		// PromotionService doc comment.
		PromotionTypeMembership,
		PromotionTypeComeback,
		PromotionTypeStudent,
		PromotionTypeAirport,
		PromotionTypeNightRide,
		PromotionTypeNewCity,
		PromotionTypeFlashSale,
	}
}

// DefaultPriority returns the BRB §3.4 default campaign priority for well-known
// types (lower number = evaluated first / wins ties). Voucher.Priority is the
// actual field the engine uses; this is only a seed value for creating BRB-
// aligned default campaigns and for tests.
func (t PromotionType) DefaultPriority() int {
	switch t {
	case PromotionTypeManualCoupon:
		return 10
	case PromotionTypeReferral:
		return 20
	case PromotionTypeFirstRide:
		return 30
	case PromotionTypeEventCampaign:
		return 40
	case PromotionTypeGoldenHour, PromotionTypeRain, PromotionTypeWeekend:
		return 50
	case PromotionTypeBirthday:
		return 60
	default:
		// Non-BRB / TODO types: lowest priority by default (never meant to win
		// over a real rule while still unapproved).
		return 1000
	}
}

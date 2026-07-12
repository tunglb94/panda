package app

import (
	"context"
	"strings"
	"time"

	"github.com/fairride/promotion/domain/entity"
)

// ─── First Ride — BRB §3.2.1 ───────────────────────────────────────────────

type FirstRideRule struct{}

func NewFirstRideRule() *FirstRideRule { return &FirstRideRule{} }

func (r *FirstRideRule) Type() entity.PromotionType { return entity.PromotionTypeFirstRide }

func (r *FirstRideRule) IsEligible(_ context.Context, _ *entity.Voucher, req *entity.PromotionRequest, now time.Time) (bool, string) {
	if req.CompletedTripsTotal != 0 {
		return false, "rider has already completed at least one trip"
	}
	if req.AccountCreatedAt == nil {
		return false, "account creation date unknown"
	}
	ageDays := int64(now.Sub(*req.AccountCreatedAt).Hours() / 24)
	if ageDays > entity.FirstRideAccountAgeDays {
		return false, "account is older than 30 days"
	}
	// BRB §3.2.1: "One promotion per phone number, per device ID." The device/
	// phone-uniqueness check belongs to the identity/fraud layer (this request
	// does not carry device/phone identifiers) — PromotionService relies on
	// VoucherValidator's per-rider usage check as the enforced backstop here.
	return true, ""
}

// ─── Birthday — BRB §3.2.2 ─────────────────────────────────────────────────

type BirthdayRule struct{}

func NewBirthdayRule() *BirthdayRule { return &BirthdayRule{} }

func (r *BirthdayRule) Type() entity.PromotionType { return entity.PromotionTypeBirthday }

func (r *BirthdayRule) IsEligible(_ context.Context, _ *entity.Voucher, req *entity.PromotionRequest, now time.Time) (bool, string) {
	if req.CompletedTripsLast90Days < entity.BirthdayMinTrips90Days {
		return false, "fewer than 3 completed trips in the past 90 days"
	}
	if req.BirthdayDate == nil {
		return false, "birthday not verified on account profile"
	}
	if !withinDayWindow(*req.BirthdayDate, now, int(entity.BirthdayWindowDays)) {
		return false, "not within the birthday window (+/- 1 day)"
	}
	return true, ""
}

// withinDayWindow reports whether now falls within +/- windowDays of the
// month-and-day of birthday (year-independent, handles Dec/Jan wraparound).
func withinDayWindow(birthday, now time.Time, windowDays int) bool {
	thisYear := time.Date(now.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, now.Location())
	candidates := []time.Time{
		thisYear,
		thisYear.AddDate(-1, 0, 0),
		thisYear.AddDate(1, 0, 0),
	}
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for _, c := range candidates {
		diff := nowDate.Sub(c).Hours() / 24
		if diff < 0 {
			diff = -diff
		}
		if int(diff) <= windowDays {
			return true
		}
	}
	return false
}

// ─── Golden Hour — BRB §3.2.3 ──────────────────────────────────────────────

type GoldenHourRule struct{}

func NewGoldenHourRule() *GoldenHourRule { return &GoldenHourRule{} }

func (r *GoldenHourRule) Type() entity.PromotionType { return entity.PromotionTypeGoldenHour }

func (r *GoldenHourRule) IsEligible(_ context.Context, _ *entity.Voucher, req *entity.PromotionRequest, _ time.Time) (bool, string) {
	if !req.IsGoldenHourWindowActive {
		return false, "not within an active Golden Hour window"
	}
	// TODO(promotion-engine): BRB §3.2.3 caps usage at "once per Golden Hour
	// window per day." This engine only enforces a total per-rider cap via
	// Voucher.MaxUsagePerUser (no daily reset) because a daily-usage counter is
	// not part of the requested repository contract. Do not treat the total
	// cap as equivalent to BRB's daily cap.
	return true, ""
}

// ─── Weekend — BRB §3.2.4 ──────────────────────────────────────────────────

type WeekendRule struct{}

func NewWeekendRule() *WeekendRule { return &WeekendRule{} }

func (r *WeekendRule) Type() entity.PromotionType { return entity.PromotionTypeWeekend }

func (r *WeekendRule) IsEligible(_ context.Context, _ *entity.Voucher, req *entity.PromotionRequest, _ time.Time) (bool, string) {
	weekday := req.RequestTime.Weekday()
	if weekday != time.Saturday && weekday != time.Sunday {
		return false, "not a weekend"
	}
	// TODO(promotion-engine): BRB §3.2.4 caps usage at "two rides per rider per
	// weekend." Same limitation as GoldenHourRule — only a total per-rider cap
	// is enforced (Voucher.MaxUsagePerUser), not a per-weekend reset.
	return true, ""
}

// ─── Rain Campaign — BRB §3.2.5 ────────────────────────────────────────────

type RainRule struct{}

func NewRainRule() *RainRule { return &RainRule{} }

func (r *RainRule) Type() entity.PromotionType { return entity.PromotionTypeRain }

func (r *RainRule) IsEligible(_ context.Context, _ *entity.Voucher, req *entity.PromotionRequest, _ time.Time) (bool, string) {
	if !req.IsRainSurchargeActive {
		return false, "rain surcharge is not currently active for this zone"
	}
	if req.RiderActiveSinceDays < entity.RainMinActiveDays {
		return false, "rider has been active for fewer than 7 days"
	}
	return true, ""
}

// ─── Event Campaign / Festival — BRB §3.2.6 ────────────────────────────────

type EventCampaignRule struct{}

func NewEventCampaignRule() *EventCampaignRule { return &EventCampaignRule{} }

func (r *EventCampaignRule) Type() entity.PromotionType { return entity.PromotionTypeEventCampaign }

func (r *EventCampaignRule) IsEligible(_ context.Context, _ *entity.Voucher, _ *entity.PromotionRequest, _ time.Time) (bool, string) {
	// BRB §3.2.6: festival campaigns are "custom per festival" — the campaign's
	// own start/end/discount fields (already checked by VoucherValidator) ARE
	// the eligibility definition. No additional per-rider criteria exist in BRB.
	return true, ""
}

// ─── Referral — BRB §3.2.7 ─────────────────────────────────────────────────

type ReferralRule struct{}

func NewReferralRule() *ReferralRule { return &ReferralRule{} }

func (r *ReferralRule) Type() entity.PromotionType { return entity.PromotionTypeReferral }

func (r *ReferralRule) IsEligible(_ context.Context, _ *entity.Voucher, req *entity.PromotionRequest, _ time.Time) (bool, string) {
	if req.ReferralCode == "" {
		return false, "no referral code supplied"
	}
	// BRB §3.4 #2: Referral only applies to "the rider's first trip."
	if !req.IsReferredFirstTrip {
		return false, "referral discount only applies to the referred rider's first trip"
	}
	// BRB §3.2.7 fraud rules (different phone/device/payment method, WiFi IP
	// soft-flag, 7-day hold) require identity/device data this engine does not
	// receive. TODO(promotion-engine): enforce via a Fraud/Risk service call
	// before this discount is committed at Redeem time.
	return true, ""
}

// ─── Manual Coupon / Coupon Campaign — BRB §3.2.9 ──────────────────────────

type ManualCouponRule struct{}

func NewManualCouponRule() *ManualCouponRule { return &ManualCouponRule{} }

func (r *ManualCouponRule) Type() entity.PromotionType { return entity.PromotionTypeManualCoupon }

func (r *ManualCouponRule) IsEligible(_ context.Context, v *entity.Voucher, req *entity.PromotionRequest, _ time.Time) (bool, string) {
	if req.VoucherCode == "" {
		return false, "no coupon code entered"
	}
	if !strings.EqualFold(req.VoucherCode, v.Code) {
		return false, "coupon code does not match"
	}
	return true, ""
}

// ─── Membership — ECONOMY_ENGINE §8.1 ──────────────────────────────────────

type MembershipRule struct{}

func NewMembershipRule() *MembershipRule { return &MembershipRule{} }

func (r *MembershipRule) Type() entity.PromotionType { return entity.PromotionTypeMembership }

// IsEligible always returns true: membership eligibility is already enforced
// generically by VoucherValidator.checkMembership (Voucher.Membership gate).
// This rule intentionally never adds extra criteria and never scales a
// discount by tier — ECONOMY_ENGINE §8.1: "Membership không bao giờ thay đổi
// công thức giá cước — chỉ thay đổi dịch vụ đi kèm." A membership-gated
// campaign is just a Voucher with Membership set and Type=membership; the
// discount amount comes from the same generic DiscountType/DiscountValue
// fields every other campaign uses.
func (r *MembershipRule) IsEligible(_ context.Context, _ *entity.Voucher, _ *entity.PromotionRequest, _ time.Time) (bool, string) {
	return true, ""
}

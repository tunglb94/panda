package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fairride/promotion/app"
	"github.com/fairride/promotion/domain/entity"
)

var ctx = context.Background()

func TestFirstRideRule(t *testing.T) {
	rule := app.NewFirstRideRule()
	created := now.Add(-10 * 24 * time.Hour)

	eligible, _ := rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		CompletedTripsTotal: 0,
		AccountCreatedAt:    &created,
	}, now)
	if !eligible {
		t.Fatal("expected eligible: new account, zero trips, within 30 days")
	}

	eligible, reason := rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		CompletedTripsTotal: 1,
		AccountCreatedAt:    &created,
	}, now)
	if eligible {
		t.Fatalf("expected NOT eligible: rider already has a completed trip, got reason=%q", reason)
	}

	oldAccount := now.Add(-40 * 24 * time.Hour)
	eligible, _ = rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		CompletedTripsTotal: 0,
		AccountCreatedAt:    &oldAccount,
	}, now)
	if eligible {
		t.Fatal("expected NOT eligible: account older than 30 days")
	}
}

func TestBirthdayRule(t *testing.T) {
	rule := app.NewBirthdayRule()

	todayBirthday := time.Date(1995, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	eligible, _ := rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		CompletedTripsLast90Days: 3,
		BirthdayDate:             &todayBirthday,
	}, now)
	if !eligible {
		t.Fatal("expected eligible: today is the birthday")
	}

	farBirthday := time.Date(1995, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 10)
	eligible, _ = rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		CompletedTripsLast90Days: 3,
		BirthdayDate:             &farBirthday,
	}, now)
	if eligible {
		t.Fatal("expected NOT eligible: birthday 10 days away, outside +/- 1 day window")
	}

	eligible, reason := rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		CompletedTripsLast90Days: 2, // BRB §3.2.2 requires >= 3
		BirthdayDate:             &todayBirthday,
	}, now)
	if eligible {
		t.Fatalf("expected NOT eligible: fewer than 3 trips in 90 days, got reason=%q", reason)
	}
}

func TestWeekendRule(t *testing.T) {
	rule := app.NewWeekendRule()
	saturday := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC) // 2026-07-11 is a Saturday
	wednesday := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)

	eligible, _ := rule.IsEligible(ctx, nil, &entity.PromotionRequest{RequestTime: saturday}, saturday)
	if !eligible {
		t.Fatal("expected eligible on Saturday")
	}
	eligible, _ = rule.IsEligible(ctx, nil, &entity.PromotionRequest{RequestTime: wednesday}, wednesday)
	if eligible {
		t.Fatal("expected NOT eligible on Wednesday")
	}
}

func TestRainRule(t *testing.T) {
	rule := app.NewRainRule()

	eligible, _ := rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		IsRainSurchargeActive: true,
		RiderActiveSinceDays:  10,
	}, now)
	if !eligible {
		t.Fatal("expected eligible: rain active, rider active 10 days")
	}

	eligible, _ = rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		IsRainSurchargeActive: false,
		RiderActiveSinceDays:  10,
	}, now)
	if eligible {
		t.Fatal("expected NOT eligible: rain not active")
	}

	eligible, reason := rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		IsRainSurchargeActive: true,
		RiderActiveSinceDays:  2,
	}, now)
	if eligible {
		t.Fatalf("expected NOT eligible: rider active < 7 days, got reason=%q", reason)
	}
}

func TestGoldenHourRule(t *testing.T) {
	rule := app.NewGoldenHourRule()
	eligible, _ := rule.IsEligible(ctx, nil, &entity.PromotionRequest{IsGoldenHourWindowActive: true}, now)
	if !eligible {
		t.Fatal("expected eligible: golden hour window active")
	}
	eligible, _ = rule.IsEligible(ctx, nil, &entity.PromotionRequest{IsGoldenHourWindowActive: false}, now)
	if eligible {
		t.Fatal("expected NOT eligible: golden hour window not active")
	}
}

func TestReferralRule(t *testing.T) {
	rule := app.NewReferralRule()

	eligible, _ := rule.IsEligible(ctx, nil, &entity.PromotionRequest{}, now)
	if eligible {
		t.Fatal("expected NOT eligible: no referral code")
	}

	eligible, _ = rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		ReferralCode:        "ABC123",
		IsReferredFirstTrip: false,
	}, now)
	if eligible {
		t.Fatal("expected NOT eligible: not the referred rider's first trip (BRB §3.4 #2)")
	}

	eligible, _ = rule.IsEligible(ctx, nil, &entity.PromotionRequest{
		ReferralCode:        "ABC123",
		IsReferredFirstTrip: true,
	}, now)
	if !eligible {
		t.Fatal("expected eligible: referral code present, is first trip")
	}
}

func TestManualCouponRule(t *testing.T) {
	rule := app.NewManualCouponRule()
	v := &entity.Voucher{Code: "SUMMER50"}

	eligible, _ := rule.IsEligible(ctx, v, &entity.PromotionRequest{VoucherCode: "SUMMER50"}, now)
	if !eligible {
		t.Fatal("expected eligible: exact code match")
	}
	eligible, _ = rule.IsEligible(ctx, v, &entity.PromotionRequest{VoucherCode: "summer50"}, now)
	if !eligible {
		t.Fatal("expected eligible: case-insensitive code match")
	}
	eligible, _ = rule.IsEligible(ctx, v, &entity.PromotionRequest{VoucherCode: "WRONG"}, now)
	if eligible {
		t.Fatal("expected NOT eligible: code mismatch")
	}
}

func TestEventCampaignRule_AlwaysEligible(t *testing.T) {
	eligible, _ := app.NewEventCampaignRule().IsEligible(ctx, nil, &entity.PromotionRequest{}, now)
	if !eligible {
		t.Fatal("event campaign eligibility is fully driven by generic voucher fields; should be true")
	}
}

func TestMembershipRule_AlwaysEligible(t *testing.T) {
	eligible, _ := app.NewMembershipRule().IsEligible(ctx, nil, &entity.PromotionRequest{}, now)
	if !eligible {
		t.Fatal("membership never adds extra eligibility criteria (ECONOMY_ENGINE §8.1); should be true")
	}
}

func TestTODORules_NeverGrantDiscount(t *testing.T) {
	todoTypes := []entity.PromotionType{
		entity.PromotionTypeComeback,
		entity.PromotionTypeStudent,
		entity.PromotionTypeAirport,
		entity.PromotionTypeNightRide,
		entity.PromotionTypeNewCity,
		entity.PromotionTypeFlashSale,
	}
	for _, pt := range todoTypes {
		rule := app.NewTODORule(pt)
		if rule.Type() != pt {
			t.Fatalf("Type() mismatch for %s", pt)
		}
		eligible, reason := rule.IsEligible(ctx, nil, &entity.PromotionRequest{}, now)
		if eligible {
			t.Fatalf("%s: TODO rule must never grant eligibility (no BRB-approved rule)", pt)
		}
		if !strings.Contains(reason, "TODO") {
			t.Fatalf("%s: expected reason to disclose TODO status, got %q", pt, reason)
		}
	}
}

func TestDefaultRuleRegistry_CoversEveryPromotionType(t *testing.T) {
	registry := app.NewDefaultRuleRegistry()
	for _, pt := range entity.AllPromotionTypes() {
		if _, ok := registry[pt]; !ok {
			t.Fatalf("promotion type %s has no registered rule", pt)
		}
	}
}

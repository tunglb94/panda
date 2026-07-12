package ruleengine

import "testing"

func TestVoucherUseDecision_BelowMinimumOrderForcesKeep(t *testing.T) {
	out := VoucherUseDecision(VoucherUseInput{
		DiscountPercent: 50, OrderAmountVND: 10_000, MinOrderVND: 50_000,
	})
	if out.Decision != DecisionKeepVoucher || out.NeedsAI || out.Confidence != 1 {
		t.Errorf("order below voucher minimum must confidently keep, got %+v", out)
	}
}

func TestVoucherUseDecision_ConfidentUseAtHighDiscount(t *testing.T) {
	out := VoucherUseDecision(VoucherUseInput{DiscountPercent: 30, PriceSensitivity: 0.1, TripCount: 200})
	if out.Decision != DecisionUseVoucher || out.NeedsAI {
		t.Errorf("expected confident use at >=30%% discount, got %+v", out)
	}
}

func TestVoucherUseDecision_ConfidentKeepAtTrivialDiscount(t *testing.T) {
	out := VoucherUseDecision(VoucherUseInput{DiscountPercent: 5, PriceSensitivity: 0.9, TripCount: 0})
	if out.Decision != DecisionKeepVoucher || out.NeedsAI {
		t.Errorf("expected confident keep at <=5%% discount, got %+v", out)
	}
}

func TestVoucherUseDecision_AmbiguousMiddleDefersToAI(t *testing.T) {
	out := VoucherUseDecision(VoucherUseInput{DiscountPercent: 15, PriceSensitivity: 0.5, TripCount: 25})
	if !out.NeedsAI {
		t.Fatalf("expected the 6-29%% band to be ambiguous, got %+v", out)
	}
	if out.Decision != DecisionUseVoucher {
		t.Errorf("expected the safe fallback to favor using it before expiry, got %q", out.Decision)
	}
}

func TestVoucherUseDecision_HighPriceSensitivityLeansUse(t *testing.T) {
	sensitive := VoucherUseDecision(VoucherUseInput{DiscountPercent: 15, PriceSensitivity: 0.95, TripCount: 0})
	insensitive := VoucherUseDecision(VoucherUseInput{DiscountPercent: 15, PriceSensitivity: 0.05, TripCount: 100})
	if sensitive.Decision != DecisionUseVoucher || sensitive.NeedsAI {
		t.Errorf("a very price-sensitive new rider should confidently use, got %+v", sensitive)
	}
	if insensitive.Decision != DecisionKeepVoucher || insensitive.NeedsAI {
		t.Errorf("a loyal, insensitive rider should confidently keep, got %+v", insensitive)
	}
}

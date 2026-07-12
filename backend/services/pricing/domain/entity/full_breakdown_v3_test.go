package entity_test

import (
	"strings"
	"testing"

	"github.com/fairride/pricing/domain/entity"
)

func TestFullFareBreakdownV3_Explanation_SkipsInapplicableLines(t *testing.T) {
	fb := entity.FullFareBreakdownV3{
		BaseFare:     13000,
		DistanceFare: 20000,
		FinalFare:    36000,
		PlatformFee:  3000,
		// TrafficTimeFare, WaitingFee, AirportFee, SurgeMultiplier, VoucherDiscount all zero/1.0.
		SurgeMultiplier: 1.0,
	}
	lines := fb.Explanation()
	for _, l := range lines {
		if l.Label == "Traffic" || l.Label == "Waiting" || strings.Contains(l.Label, "Airport") ||
			l.Label == "Surge" || l.Label == "Voucher" {
			t.Errorf("inapplicable line %q should not appear when its amount is 0", l.Label)
		}
	}
	// Base Fare, Distance, Booking Fee, Final must always appear.
	labels := map[string]bool{}
	for _, l := range lines {
		labels[l.Label] = true
	}
	for _, want := range []string{"Base Fare", "Distance", "Booking Fee", "Final"} {
		if !labels[want] {
			t.Errorf("expected line %q to be present", want)
		}
	}
}

func TestFullFareBreakdownV3_Explanation_IncludesAppliedLines(t *testing.T) {
	fb := entity.FullFareBreakdownV3{
		BaseFare:        13000,
		DistanceFare:    20000,
		TrafficTimeFare: 2000,
		WaitingFee:      1500,
		AirportFee:      15000,
		AirportLeg:      entity.AirportLegPickup,
		RideFare:        51000,
		SurgeMultiplier: 1.2,
		SurgeLabel:      "Busy",
		PlatformFee:     3000,
		VoucherDiscount: 5000,
		VoucherLabel:    "First Ride",
		FinalFare:       50500,
	}
	lines := fb.Explanation()
	labels := map[string]int64{}
	for _, l := range lines {
		labels[l.Label] = l.Amount
	}
	if _, ok := labels["Traffic"]; !ok {
		t.Error("expected Traffic line when TrafficTimeFare > 0")
	}
	if _, ok := labels["Waiting"]; !ok {
		t.Error("expected Waiting line when WaitingFee > 0")
	}
	if _, ok := labels["Airport Pickup"]; !ok {
		t.Error("expected 'Airport Pickup' line for AirportLegPickup")
	}
	if _, ok := labels["Busy"]; !ok {
		t.Error("expected surge line labelled with SurgeLabel when SurgeMultiplier > 1.0")
	}
	if amt, ok := labels["First Ride"]; !ok || amt != -5000 {
		t.Errorf("expected voucher line 'First Ride' = -5000, got %v (present=%v)", amt, ok)
	}
	if labels["Final"] != 50500 {
		t.Errorf("Final line = %d, want 50500", labels["Final"])
	}
}

func TestFullFareBreakdownV3_ExplanationString_Format(t *testing.T) {
	fb := entity.FullFareBreakdownV3{
		BaseFare:     25000,
		DistanceFare: 38000,
		PlatformFee:  0,
		FinalFare:    63000,
	}
	s := fb.ExplanationString()
	if !strings.Contains(s, "Base Fare") || !strings.Contains(s, "25.000") {
		t.Errorf("expected Base Fare line with thousands-separated amount, got:\n%s", s)
	}
	if !strings.Contains(s, "Final") || !strings.Contains(s, "63.000") {
		t.Errorf("expected Final line with 63.000, got:\n%s", s)
	}
	if !strings.HasSuffix(s, "\n") {
		t.Error("expected explanation string to end with a newline after the last line")
	}
}

func TestFullFareBreakdownV3_ExplanationString_NegativeVoucherFormatting(t *testing.T) {
	fb := entity.FullFareBreakdownV3{
		BaseFare:        25000,
		DistanceFare:    38000,
		VoucherDiscount: 15000,
		VoucherLabel:    "Voucher",
		FinalFare:       48000,
	}
	s := fb.ExplanationString()
	if !strings.Contains(s, "-15.000") {
		t.Errorf("expected a negative voucher amount '-15.000' in explanation, got:\n%s", s)
	}
}

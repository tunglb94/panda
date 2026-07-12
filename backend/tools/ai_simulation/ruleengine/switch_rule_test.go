package ruleengine

import "testing"

func TestSwitchAppDecision_NoCompetitorPrice(t *testing.T) {
	out := SwitchAppDecision(SwitchAppInput{PandaFareVND: 91_000, CompetitorFareVND: 0})
	if out.Decision != DecisionStay || out.NeedsAI {
		t.Errorf("expected confident Stay with no competitor price, got %+v", out)
	}
}

func TestSwitchAppDecision_ConfidentStayWhenMuchCheaper(t *testing.T) {
	out := SwitchAppDecision(SwitchAppInput{
		PandaFareVND: 70_000, CompetitorFareVND: 100_000,
		PriceSensitivity: 0.9, Patience: 0.1, Membership: "free",
	})
	if out.Decision != DecisionStay || out.NeedsAI {
		t.Errorf("expected confident Stay when Panda is >10%% cheaper, got %+v", out)
	}
}

func TestSwitchAppDecision_ConfidentSwitchWhenMuchPricierAndDisloyal(t *testing.T) {
	out := SwitchAppDecision(SwitchAppInput{
		PandaFareVND: 130_000, CompetitorFareVND: 100_000, // +30%
		PriceSensitivity: 0.9, Patience: 0.1, Membership: "free",
	})
	if out.Decision != DecisionSwitch || out.NeedsAI {
		t.Errorf("expected confident Switch for a much pricier, low-loyalty rider, got %+v", out)
	}
}

func TestSwitchAppDecision_LoyaltyOverridesPureEconomics(t *testing.T) {
	out := SwitchAppDecision(SwitchAppInput{
		PandaFareVND: 130_000, CompetitorFareVND: 100_000, // +30%, same gap as above
		PriceSensitivity: 0.9, Patience: 0.95, Membership: "diamond", // high loyalty
	})
	if out.Decision == DecisionSwitch && !out.NeedsAI {
		t.Errorf("a loyal diamond member should not confidently switch on price alone, got %+v", out)
	}
}

func TestSwitchAppDecision_WorkedExampleIsAmbiguous(t *testing.T) {
	// The sprint brief's own worked example: Grab 96k vs Panda 91k.
	out := SwitchAppDecision(SwitchAppInput{
		PandaFareVND: 91_000, CompetitorFareVND: 96_000,
		PriceSensitivity: 0.6, Patience: 0.5, Membership: "free",
	})
	if !out.NeedsAI {
		t.Fatalf("expected the brief's worked example to be ambiguous, got %+v", out)
	}
	if out.Decision != DecisionStay {
		t.Errorf("expected the safe fallback to be Stay (Panda already cheaper), got %q", out.Decision)
	}
}

func TestMembershipLoyalty_Ordering(t *testing.T) {
	tiers := []string{"free", "silver", "gold", "diamond"}
	prev := -1.0
	for _, tier := range tiers {
		v := membershipLoyalty(tier)
		if v <= prev {
			t.Errorf("expected loyalty to strictly increase by tier, %q gave %v (prev %v)", tier, v, prev)
		}
		prev = v
	}
	if membershipLoyalty("unknown") != membershipLoyalty("free") {
		t.Errorf("unknown tier should fall back to the same loyalty as free")
	}
}

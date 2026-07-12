package ruleengine

import "testing"

func TestFatigueDecision_HardFloors(t *testing.T) {
	cases := []struct {
		name string
		in   FatigueInput
	}{
		{"critical battery", FatigueInput{PhoneBattery: 0.04, Fuel: 0.5, Fatigue: 0.1, HoursOnlineToday: 1}},
		{"critical fuel", FatigueInput{PhoneBattery: 0.5, Fuel: 0.03, Fatigue: 0.1, HoursOnlineToday: 1}},
		{"hours ceiling", FatigueInput{PhoneBattery: 0.9, Fuel: 0.9, Fatigue: 0.1, HoursOnlineToday: 12}},
		{"fatigue ceiling", FatigueInput{PhoneBattery: 0.9, Fuel: 0.9, Fatigue: 0.95, HoursOnlineToday: 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := FatigueDecision(c.in)
			if out.Decision != DecisionStop {
				t.Errorf("expected Stop, got %s", out.Decision)
			}
			if out.NeedsAI {
				t.Errorf("hard floor must never defer to AI")
			}
			if out.Confidence != 1 {
				t.Errorf("expected full confidence on a hard floor, got %v", out.Confidence)
			}
		})
	}
}

func TestFatigueDecision_ConfidentContinue(t *testing.T) {
	out := FatigueDecision(FatigueInput{PhoneBattery: 0.9, Fuel: 0.9, Fatigue: 0.2, HoursOnlineToday: 1})
	if out.Decision != DecisionContinue || out.NeedsAI {
		t.Errorf("expected confident Continue, got %+v", out)
	}
}

func TestFatigueDecision_AmbiguousBandDefersToAI(t *testing.T) {
	// Moderate fatigue and hours, no income pressure — lands in the
	// documented ambiguous band (0.45-0.75).
	out := FatigueDecision(FatigueInput{PhoneBattery: 0.9, Fuel: 0.9, Fatigue: 0.65, HoursOnlineToday: 7})
	if !out.NeedsAI {
		t.Fatalf("expected NeedsAI=true in the ambiguous band, got %+v", out)
	}
	// Even when ambiguous, a safe fallback decision must still be present
	// (Ollama-down mode must keep working).
	if out.Decision != DecisionContinue && out.Decision != DecisionStop {
		t.Errorf("expected a valid fallback decision, got %q", out.Decision)
	}
}

func TestFatigueDecision_IncomeShortfallPullsTowardContinue(t *testing.T) {
	// Fatigue=0.6, hours=6 -> score 0.555, inside the ambiguous band
	// (0.45-0.75) with no income pressure.
	base := FatigueInput{PhoneBattery: 0.9, Fuel: 0.9, Fatigue: 0.6, HoursOnlineToday: 6}
	withoutTarget := FatigueDecision(base)
	if !withoutTarget.NeedsAI {
		t.Fatalf("expected the baseline case to be ambiguous, got %+v", withoutTarget)
	}

	// A large shortfall against the driver's daily target pulls the score
	// down by up to 0.2, enough to drop below the ambiguity floor into a
	// confident Continue — income pressure keeps drivers going.
	withShortfall := base
	withShortfall.DailyTargetVND = 1_000_000
	withShortfall.IncomeTodayVND = 100_000
	withTarget := FatigueDecision(withShortfall)

	if withTarget.NeedsAI {
		t.Errorf("large income shortfall should resolve confidently, got %+v", withTarget)
	}
	if withTarget.Decision != DecisionContinue {
		t.Errorf("income shortfall should pull the driver toward Continue, got %q", withTarget.Decision)
	}
}

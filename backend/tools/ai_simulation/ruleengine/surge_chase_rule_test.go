package ruleengine

import "testing"

func TestSurgeChaseDecision_LowFuelForcesStay(t *testing.T) {
	out := SurgeChaseDecision(SurgeChaseInput{SurgeMultiplier: 3.0, DistanceKM: 1, Fuel: 0.1})
	if out.Decision != DecisionStayPut || out.NeedsAI {
		t.Errorf("low fuel must confidently stay regardless of surge size, got %+v", out)
	}
}

func TestSurgeChaseDecision_ConfidentChaseBigSurgeShortHop(t *testing.T) {
	out := SurgeChaseDecision(SurgeChaseInput{SurgeMultiplier: 2.0, DistanceKM: 3, Fuel: 0.9, Fatigue: 0.1})
	if out.Decision != DecisionChaseSurge || out.NeedsAI {
		t.Errorf("expected confident chase for a big, close surge, got %+v", out)
	}
}

func TestSurgeChaseDecision_ConfidentStaySmallSurgeOrFarZone(t *testing.T) {
	small := SurgeChaseDecision(SurgeChaseInput{SurgeMultiplier: 1.1, DistanceKM: 3, Fuel: 0.9})
	if small.Decision != DecisionStayPut || small.NeedsAI {
		t.Errorf("expected confident stay for a trivial surge, got %+v", small)
	}
	far := SurgeChaseDecision(SurgeChaseInput{SurgeMultiplier: 2.0, DistanceKM: 20, Fuel: 0.9})
	if far.Decision != DecisionStayPut || far.NeedsAI {
		t.Errorf("expected confident stay for a too-far zone, got %+v", far)
	}
}

func TestSurgeChaseDecision_AmbiguousMiddleDefersToAI(t *testing.T) {
	out := SurgeChaseDecision(SurgeChaseInput{SurgeMultiplier: 1.7, DistanceKM: 6, Fuel: 0.9, Fatigue: 0.1})
	if !out.NeedsAI {
		t.Fatalf("expected a moderate surge at moderate distance to be ambiguous, got %+v", out)
	}
	if out.Decision != DecisionStayPut {
		t.Errorf("expected the safe fallback to be StayPut (no speculative relocation), got %q", out.Decision)
	}
}

func TestSurgeChaseDecision_IncomeShortfallRaisesChaseScore(t *testing.T) {
	// Same inputs as the ambiguous-band case above (score ~0.51).
	base := SurgeChaseInput{SurgeMultiplier: 1.7, DistanceKM: 6, Fuel: 0.9, Fatigue: 0.1}
	without := SurgeChaseDecision(base)
	if !without.NeedsAI {
		t.Fatalf("expected the baseline case to be ambiguous, got %+v", without)
	}

	// A full shortfall against the daily target adds up to +0.15, enough to
	// push the score above the ambiguity ceiling into a confident chase.
	withShortfall := base
	withShortfall.DailyTargetVND = 1_000_000
	withShortfall.IncomeTodayVND = 0
	with := SurgeChaseDecision(withShortfall)

	if with.NeedsAI {
		t.Errorf("full income shortfall should resolve confidently, got %+v", with)
	}
	if with.Decision != DecisionChaseSurge {
		t.Errorf("income shortfall should push the driver toward chasing the surge, got %q", with.Decision)
	}
}

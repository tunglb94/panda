package ruleengine

const (
	DecisionChaseSurge Decision = "chase_surge"
	DecisionStayPut    Decision = "stay_put"
)

// SurgeChaseInput carries what SurgeChaseDecision needs.
type SurgeChaseInput struct {
	SurgeMultiplier float64 // in the target zone
	DistanceKM      float64 // from driver's current zone to the surge zone
	Fatigue         float64 // 0-1
	Fuel            float64 // 0-1
	IncomeTodayVND  int64
	DailyTargetVND  int64
}

// SurgeChaseDecision implements the sprint brief's worked example verbatim:
// "Driver thấy surge → phi4 quyết định Có chạy sang khu vực đó hay không."
func SurgeChaseDecision(in SurgeChaseInput) Outcome {
	if in.Fuel <= 0.15 {
		return Outcome{Decision: DecisionStayPut, Confidence: 0.95, Reason: "insufficient fuel/charge to relocate"}
	}

	// Confident chase: big surge, short hop.
	if in.SurgeMultiplier >= 1.8 && in.DistanceKM <= 5 {
		return Outcome{Decision: DecisionChaseSurge, Confidence: 0.9, Reason: "high surge, short relocation"}
	}
	// Confident stay: surge too small to matter, or the zone is too far to
	// be worth the deadhead drive.
	if in.SurgeMultiplier < 1.2 || in.DistanceKM > 15 {
		return Outcome{Decision: DecisionStayPut, Confidence: 0.85, Reason: "surge too small or zone too far"}
	}

	// Ambiguous middle: moderate surge at moderate distance. Fatigue and
	// whether the driver has already hit their daily target pull the score.
	chaseScore := clamp01(0.5*clamp01((in.SurgeMultiplier-1.0)/1.0) +
		0.3*(1-clamp01(in.DistanceKM/15)) -
		0.2*in.Fatigue)
	if in.DailyTargetVND > 0 && in.IncomeTodayVND < in.DailyTargetVND {
		shortfall := 1 - clamp01(float64(in.IncomeTodayVND)/float64(in.DailyTargetVND))
		chaseScore += 0.15 * shortfall
	}
	chaseScore = clamp01(chaseScore)

	const ambiguityLow, ambiguityHigh = 0.35, 0.65
	switch {
	case chaseScore < ambiguityLow:
		return Outcome{Decision: DecisionStayPut, Confidence: 1 - chaseScore, Reason: "chase score below threshold"}
	case chaseScore > ambiguityHigh:
		return Outcome{Decision: DecisionChaseSurge, Confidence: chaseScore, Reason: "chase score above threshold"}
	default:
		return Outcome{
			Decision:   DecisionStayPut, // safe fallback: don't relocate speculatively without AI confirmation
			Confidence: chaseScore,
			NeedsAI:    true,
			Reason:     "chase score in ambiguous band",
		}
	}
}

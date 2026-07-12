package ruleengine

const (
	DecisionContinue Decision = "continue"
	DecisionStop     Decision = "stop"
)

// FatigueInput carries exactly what FatigueDecision needs — kept as a small
// struct (rather than passing the full DriverAgent) so this file has no
// dependency on domain/entity and stays trivially unit-testable.
type FatigueInput struct {
	Fatigue          float64 // 0-1
	HoursOnlineToday float64
	PhoneBattery     float64 // 0-1
	Fuel             float64 // 0-1
	IncomeTodayVND   int64
	DailyTargetVND   int64 // driver's rough personal daily income goal
}

// FatigueDecision implements the sprint brief's worked example verbatim:
// "Driver đã chạy 8 tiếng → Rule Engine đánh dấu Fatigue=High → phi4 quyết
// định 'Tắt app' hay 'Chạy thêm'."
//
// Hard safety floors are decided by the rule engine alone, never deferred to
// AI: critically low battery/fuel, or fatigue/hours past any reasonable
// limit, always means Stop — an LLM should never be in the loop for a
// safety-critical floor. The genuinely ambiguous middle (driver is tired but
// hasn't hit bottom yet, and hasn't reached their income target) is where
// AI judgment adds real value over a hard threshold.
func FatigueDecision(in FatigueInput) Outcome {
	// Hard floors — 100% rule engine, no exception.
	if in.PhoneBattery <= 0.05 {
		return Outcome{Decision: DecisionStop, Confidence: 1, Reason: "phone battery critical"}
	}
	if in.Fuel <= 0.05 {
		return Outcome{Decision: DecisionStop, Confidence: 1, Reason: "fuel/charge critical"}
	}
	if in.HoursOnlineToday >= 12 || in.Fatigue >= 0.92 {
		return Outcome{Decision: DecisionStop, Confidence: 1, Reason: "hard fatigue/hours ceiling"}
	}

	// Confident continue — fresh driver, no reason to stop.
	if in.Fatigue <= 0.35 && in.HoursOnlineToday < 4 {
		return Outcome{Decision: DecisionContinue, Confidence: 0.95, Reason: "low fatigue, early shift"}
	}

	// Score rises with fatigue and hours, falls when the driver is still
	// short of their own daily income target (income pressure keeps
	// drivers going past comfortable fatigue — a real, documented behavior
	// this simulation intentionally models, not a BRB rule).
	fatigueScore := 0.55*in.Fatigue + 0.45*clamp01(in.HoursOnlineToday/12)
	if in.DailyTargetVND > 0 && in.IncomeTodayVND < in.DailyTargetVND {
		shortfall := 1 - clamp01(float64(in.IncomeTodayVND)/float64(in.DailyTargetVND))
		fatigueScore -= 0.2 * shortfall // income pressure pulls the score back toward "continue"
	}
	fatigueScore = clamp01(fatigueScore)

	const ambiguityLow, ambiguityHigh = 0.45, 0.75
	switch {
	case fatigueScore < ambiguityLow:
		return Outcome{Decision: DecisionContinue, Confidence: 1 - fatigueScore, Reason: "fatigue score below threshold"}
	case fatigueScore > ambiguityHigh:
		return Outcome{Decision: DecisionStop, Confidence: fatigueScore, Reason: "fatigue score above threshold"}
	default:
		// Ambiguous — rule engine still returns a safe fallback (favor
		// stopping, the conservative/safety-first default) in case AI is
		// unavailable, but flags this for AI review.
		return Outcome{
			Decision:   DecisionStop,
			Confidence: fatigueScore,
			NeedsAI:    true,
			Reason:     "fatigue score in ambiguous band",
		}
	}
}

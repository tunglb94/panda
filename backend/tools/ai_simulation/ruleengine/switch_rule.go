package ruleengine

const (
	DecisionStay   Decision = "stay"
	DecisionSwitch Decision = "switch"
)

// SwitchAppInput carries what SwitchAppDecision needs.
type SwitchAppInput struct {
	PandaFareVND      int64
	CompetitorFareVND int64 // e.g. Grab's quoted price for the same trip
	PriceSensitivity  float64 // 0-1
	Patience          float64 // 0-1, higher = more tolerant/loyal
	Membership        string  // "free"|"silver"|"gold"|"diamond" — higher tiers lean stickier
}

// SwitchAppDecision implements the sprint brief's worked example verbatim:
// "Khách thấy Grab 96k, Panda 91k → phi4 quyết định Đổi app hay tiếp tục."
// Note the example itself has Panda CHEAPER — the point is price is not the
// only factor even when Panda wins: small gaps still leave room for habit,
// app familiarity, or driver availability perception to tip a rider toward
// the competitor, which is exactly the kind of judgment call this hands to
// AI rather than a hard "cheaper always wins" rule.
func SwitchAppDecision(in SwitchAppInput) Outcome {
	if in.CompetitorFareVND <= 0 {
		return Outcome{Decision: DecisionStay, Confidence: 1, Reason: "no competitor price available"}
	}

	// gapRatio > 0 means Panda is more expensive than the competitor;
	// gapRatio < 0 means Panda is cheaper.
	gapRatio := float64(in.PandaFareVND-in.CompetitorFareVND) / float64(in.CompetitorFareVND)

	// Confident stay: Panda meaningfully cheaper (>10%).
	if gapRatio <= -0.10 {
		return Outcome{Decision: DecisionStay, Confidence: 0.9, Reason: "Panda meaningfully cheaper"}
	}
	// Confident switch: Panda meaningfully more expensive (>25%), scaled
	// down by patience/membership loyalty which still shifts pure economics.
	if gapRatio >= 0.25 {
		loyalty := clamp01(0.5*in.Patience + 0.5*membershipLoyalty(in.Membership))
		if loyalty < 0.3 {
			return Outcome{Decision: DecisionSwitch, Confidence: 0.9, Reason: "Panda much pricier, low loyalty"}
		}
	}

	// Ambiguous band: Panda slightly cheaper up to meaningfully pricier —
	// exactly the sprint brief's example (Panda 91k vs Grab 96k, gapRatio
	// ≈ -0.052). Score blends price sensitivity against loyalty signals.
	switchScore := clamp01(0.5*clamp01((gapRatio+0.10)/0.35)*in.PriceSensitivity +
		0.5*(1-clamp01(0.5*in.Patience+0.5*membershipLoyalty(in.Membership))))

	const ambiguityLow, ambiguityHigh = 0.3, 0.7
	switch {
	case switchScore < ambiguityLow:
		return Outcome{Decision: DecisionStay, Confidence: 1 - switchScore, Reason: "switch score below threshold"}
	case switchScore > ambiguityHigh:
		return Outcome{Decision: DecisionSwitch, Confidence: switchScore, Reason: "switch score above threshold"}
	default:
		return Outcome{
			Decision:   DecisionStay, // safe fallback if AI unavailable: assume retained
			Confidence: switchScore,
			NeedsAI:    true,
			Reason:     "switch score in ambiguous band",
		}
	}
}

func membershipLoyalty(tier string) float64 {
	switch tier {
	case "diamond":
		return 0.9
	case "gold":
		return 0.7
	case "silver":
		return 0.5
	default: // "free" or unknown
		return 0.2
	}
}

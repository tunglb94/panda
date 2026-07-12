package ruleengine

const (
	DecisionUseVoucher  Decision = "use_voucher"
	DecisionKeepVoucher Decision = "keep_voucher"
)

// VoucherUseInput carries what VoucherUseDecision needs.
type VoucherUseInput struct {
	DiscountPercent  int
	PriceSensitivity float64 // 0-1
	TripCount        int     // loyalty/experience proxy
	OrderAmountVND   int64
	MinOrderVND      int64 // voucher's own minimum order (BRB §4.8) — 0 if none
}

// VoucherUseDecision implements the sprint brief's worked example verbatim:
// "Khách được voucher → phi4 quyết định Có sử dụng hay giữ lại." A voucher a
// rider chooses NOT to use this trip stays in VoucherCatalog for a future
// trip (see simulation/ride_flow.go) — "keep" is a real deferral, not a loss.
func VoucherUseDecision(in VoucherUseInput) Outcome {
	if in.MinOrderVND > 0 && in.OrderAmountVND < in.MinOrderVND {
		return Outcome{Decision: DecisionKeepVoucher, Confidence: 1, Reason: "order below voucher minimum"}
	}

	// Confident use: large discount, obviously worth it regardless of profile.
	if in.DiscountPercent >= 30 {
		return Outcome{Decision: DecisionUseVoucher, Confidence: 0.92, Reason: "high discount value"}
	}
	// Confident keep: trivial discount, not worth spending a limited-use
	// voucher on a low-value moment.
	if in.DiscountPercent <= 5 {
		return Outcome{Decision: DecisionKeepVoucher, Confidence: 0.85, Reason: "negligible discount value"}
	}

	// Ambiguous middle (6-29%): the sprint brief's example band. A more
	// price-sensitive or newer (lower trip count) rider leans toward using
	// it now; a loyal, less price-sensitive rider leans toward saving it
	// for a bigger trip.
	useScore := clamp01(0.6*in.PriceSensitivity + 0.4*(1-clamp01(float64(in.TripCount)/50)))

	const ambiguityLow, ambiguityHigh = 0.3, 0.7
	switch {
	case useScore < ambiguityLow:
		return Outcome{Decision: DecisionKeepVoucher, Confidence: 1 - useScore, Reason: "use score below threshold"}
	case useScore > ambiguityHigh:
		return Outcome{Decision: DecisionUseVoucher, Confidence: useScore, Reason: "use score above threshold"}
	default:
		return Outcome{
			Decision:   DecisionUseVoucher, // safe fallback: use it now rather than risk it expiring unused
			Confidence: useScore,
			NeedsAI:    true,
			Reason:     "use score in ambiguous band",
		}
	}
}

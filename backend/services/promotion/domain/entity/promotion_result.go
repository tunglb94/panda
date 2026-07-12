package entity

// PromotionResult is PromotionService.Evaluate's output: how much discount
// (if any) applies to the trip described by the PromotionRequest.
type PromotionResult struct {
	Applied bool

	VoucherID   string
	VoucherCode string
	VoucherName string
	Type        PromotionType

	DiscountType   DiscountType
	DiscountAmount int64 // VND, final amount after BRB §4.9 clamp (never > OriginalOrderAmount)

	OriginalOrderAmount int64
	FinalOrderAmount    int64 // OriginalOrderAmount - DiscountAmount, never negative (BRB §4.9)

	// Reason explains the outcome for transparency (BRB §1.2 philosophy: riders
	// and support staff should always be able to see why a promotion did or
	// did not apply).
	Reason string

	// Warnings surfaces non-fatal notes, e.g. other eligible-but-lower-priority
	// candidates that were not applied because BRB §4.7 allows at most one
	// voucher per trip, or a TODO-type campaign that was skipped because BRB
	// has no approved rule for it yet.
	Warnings []string
}

// NoDiscount builds a PromotionResult representing "nothing applied."
func NoDiscount(orderAmount int64, reason string, warnings []string) *PromotionResult {
	return &PromotionResult{
		Applied:             false,
		OriginalOrderAmount: orderAmount,
		FinalOrderAmount:    orderAmount,
		Reason:              reason,
		Warnings:            warnings,
	}
}

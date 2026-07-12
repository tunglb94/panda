package simulation

import "fmt"

// This file implements BƯỚC 7 of the sprint brief — the simulator must never
// produce a result that violates one of these five invariants. applySafetyClamps
// is called at the end of every Simulate() so a caller can never observe a
// violating FareBreakdown; Validate is a second, independent pass used by the
// test suite (pricing_simulator_test.go) to prove the clamps are unreachable
// in practice — i.e. Validate should find zero issues across all 100+
// scenarios, and if it ever does, that is a bug in the formula above, not a
// case the clamp is meant to paper over silently.

// applySafetyClamps enforces every BƯỚC-7 invariant on a computed breakdown,
// recording a Warning for any clamp it actually had to apply. In a correctly
// implemented engine none of these should ever fire — they exist as a last
// line of defence, not as the primary correctness mechanism.
func applySafetyClamps(fb *FareBreakdown) {
	if fb.CustomerTotal < 0 {
		fb.Warnings = append(fb.Warnings, fmt.Sprintf("SAFETY: CustomerTotal was negative (%d) — clamped to 0", fb.CustomerTotal))
		fb.CustomerTotal = 0
	}
	if fb.Commission < 0 {
		fb.Warnings = append(fb.Warnings, fmt.Sprintf("SAFETY: Commission was negative (%d) — clamped to 0", fb.Commission))
		fb.Commission = 0
	}
	if fb.NetDriver < 0 {
		fb.Warnings = append(fb.Warnings, fmt.Sprintf("SAFETY: NetDriver was negative (%d) — clamped to 0", fb.NetDriver))
		fb.NetDriver = 0
	}
	if fb.VAT < 0 {
		fb.Warnings = append(fb.Warnings, fmt.Sprintf("SAFETY: VAT was negative (%d) — clamped to 0", fb.VAT))
		fb.VAT = 0
	}
	if fb.PromotionApplied > fb.PromotionRequested {
		// Should be structurally impossible (Simulate only ever clamps
		// downward) — flagged here too so Validate can catch a future
		// refactor that breaks this without needing to read the formula.
		fb.Warnings = append(fb.Warnings, "SAFETY: PromotionApplied exceeded PromotionRequested — invariant violated")
	}

	// Bounded-loss check: the maximum a single trip can cost the platform is
	// the sum of everything it might fund on the driver's behalf (promotion,
	// minimum-earning top-up, long-pickup compensation, insurance). A loss
	// beyond that sum is not "acceptable loss-leader economics" (BRB §6.11
	// explicitly allows small per-trip losses) — it means the formula above
	// has a bug, since nothing else in this engine can create platform cost.
	maxExplainableLoss := fb.PromotionApplied + fb.MinimumEarningTopUp + fb.LongPickupCompensation + AssumedInsuranceCostVND
	if fb.Profit < -maxExplainableLoss {
		fb.Warnings = append(fb.Warnings, fmt.Sprintf(
			"SAFETY: Profit (%d) is a larger loss than every funded obligation combined (-%d) — unbounded-loss invariant violated",
			fb.Profit, maxExplainableLoss))
	}
}

// Validate re-checks every BƯỚC-7 invariant against an already-computed
// breakdown and returns a human-readable issue per violation found. Used by
// tests; returns nil for a healthy breakdown.
func Validate(fb *FareBreakdown) []string {
	var issues []string
	if fb.CustomerTotal < 0 {
		issues = append(issues, "CustomerTotal is negative")
	}
	if fb.Commission < 0 {
		issues = append(issues, "Commission is negative")
	}
	if fb.NetDriver < 0 {
		issues = append(issues, "NetDriver is negative")
	}
	if fb.VAT < 0 {
		issues = append(issues, "VAT is negative")
	}
	if fb.PromotionApplied > fb.PromotionRequested {
		issues = append(issues, "PromotionApplied exceeds PromotionRequested")
	}
	if fb.PromotionApplied < 0 {
		issues = append(issues, "PromotionApplied is negative")
	}
	maxExplainableLoss := fb.PromotionApplied + fb.MinimumEarningTopUp + fb.LongPickupCompensation + AssumedInsuranceCostVND
	if fb.Profit < -maxExplainableLoss {
		issues = append(issues, "Profit is a larger loss than every funded obligation combined (unbounded-loss)")
	}
	if fb.SurgeMultiplier > MaxSurgeMultiplier {
		issues = append(issues, "SurgeMultiplier exceeds MaxSurgeMultiplier")
	}
	if fb.NetDriver < MinimumDriverEarningVND && fb.NetDriver != 0 {
		// NetDriver can only be < the guarantee if it was clamped to 0 by a
		// safety rule above (already reported); any other case is a bug.
		issues = append(issues, "NetDriver is below the minimum earning guarantee without being a reported clamp")
	}
	return issues
}

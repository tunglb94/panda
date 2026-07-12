package aiengine

import "fmt"

// Prompt templates for the 4 decision families the sprint brief names
// explicitly. Each is deliberately short — well under the 300-token budget
// (these run more like 40-80 tokens) — a compact categorical description of
// the situation plus a strict single-word answer format, since the model is
// acting as a classifier, not a conversational agent. Keeping the exact
// wording centralized here (rather than inlined at call sites) is also what
// makes the cache in cache.go actually work: identical situations must
// produce byte-identical prompts to hit the same cache key.

// FatiguePrompt mirrors ruleengine.FatigueDecision's ambiguous band.
func FatiguePrompt(fatigueBucket string, hoursOnline int, incomeBucket string, metTarget bool) string {
	return fmt.Sprintf(
		"You are a ride-hailing driver assistant. Driver fatigue=%s, hours online today=%d, income today=%s of daily target, target met=%v. "+
			"Should the driver continue working or stop? Reply with exactly one word: CONTINUE or STOP.",
		fatigueBucket, hoursOnline, incomeBucket, metTarget,
	)
}

// SwitchAppPrompt mirrors ruleengine.SwitchAppDecision's ambiguous band.
func SwitchAppPrompt(pandaFare, competitorFare int64, sensitivityBucket, membership string) string {
	return fmt.Sprintf(
		"A ride-hailing rider sees Panda price %d VND and a competitor app price %d VND for the same trip. "+
			"Rider price sensitivity=%s, membership tier=%s. Will the rider switch to the competitor app or stay on Panda? "+
			"Reply with exactly one word: SWITCH or STAY.",
		pandaFare, competitorFare, sensitivityBucket, membership,
	)
}

// VoucherUsePrompt mirrors ruleengine.VoucherUseDecision's ambiguous band.
func VoucherUsePrompt(discountPercent int, sensitivityBucket string, loyaltyBucket string) string {
	return fmt.Sprintf(
		"A ride-hailing rider has an active voucher worth %d%% off this trip. Rider price sensitivity=%s, loyalty=%s. "+
			"Should the rider use the voucher now or keep it for a future trip? Reply with exactly one word: USE or KEEP.",
		discountPercent, sensitivityBucket, loyaltyBucket,
	)
}

// SurgeChasePrompt mirrors ruleengine.SurgeChaseDecision's ambiguous band.
func SurgeChasePrompt(surgeMultiplier float64, distanceKM float64, fatigueBucket string, metTarget bool) string {
	return fmt.Sprintf(
		"A ride-hailing driver sees surge pricing x%.1f in a zone %.0fkm away. Driver fatigue=%s, daily income target met=%v. "+
			"Should the driver relocate to chase the surge or stay put? Reply with exactly one word: CHASE or STAY.",
		surgeMultiplier, distanceKM, fatigueBucket, metTarget,
	)
}

// Bucket* helpers discretize continuous values into a small number of
// categories before they enter a prompt. This is deliberate, not just
// tidiness: it keeps the space of distinct prompts small, which is what
// makes the cache hit rate high in a simulation with thousands of agents —
// many agents in "fatigue=high, 8 hours online" are the exact same cache
// entry instead of thousands of near-duplicate floating-point prompts.

func BucketLevel3(v float64) string { // low/medium/high
	switch {
	case v < 0.34:
		return "low"
	case v < 0.67:
		return "medium"
	default:
		return "high"
	}
}

func BucketIncomeProgress(current, target int64) string {
	if target <= 0 {
		return "unknown"
	}
	ratio := float64(current) / float64(target)
	switch {
	case ratio < 0.34:
		return "low"
	case ratio < 0.67:
		return "medium"
	default:
		return "high"
	}
}

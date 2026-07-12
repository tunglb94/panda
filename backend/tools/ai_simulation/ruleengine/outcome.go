// Package ruleengine implements the deterministic 95% of every agent
// decision in the simulation. Each rule function is a pure, fast,
// thoroughly-unit-testable calculation over agent/world state. When a
// situation is genuinely ambiguous — the clear-cut thresholds don't apply —
// the rule marks NeedsAI=true and hands off to aiengine for the nuanced 5%.
// The rule engine ALWAYS also computes a safe fallback Decision even when
// NeedsAI is true, so the simulation keeps running correctly if Ollama is
// unavailable (per the sprint brief: "Nếu Ollama chết → Simulation vẫn
// chạy 100% Rule Engine").
package ruleengine

// Decision is a generic yes/no-style outcome; each decision family defines
// its own two string constants (see fatigue_rule.go, switch_rule.go, etc.)
// so call sites read naturally ("Continue"/"Stop" rather than a bare bool).
type Decision string

// Outcome is what every rule function returns.
type Outcome struct {
	Decision Decision

	// Confidence is how sure the deterministic rule is, 0-1. High
	// confidence (>= ambiguity band) means the rule decides outright.
	Confidence float64

	// NeedsAI is true when Confidence falls inside the rule's ambiguity
	// band — the case is a genuine judgment call, not a hard threshold.
	NeedsAI bool

	// Reason is a short, human-readable explanation (used in logs/benchmark,
	// not shown to any end user — this is a simulation, not production UI).
	Reason string
}

// clamp01 keeps a computed score within the valid [0,1] probability range.
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

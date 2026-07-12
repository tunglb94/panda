package aiengine

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/ai_simulation/ruleengine"
)

func TestParseBinary(t *testing.T) {
	parse := ParseBinary("CONTINUE", ruleengine.DecisionContinue, "STOP", ruleengine.DecisionStop)

	cases := []struct {
		raw      string
		wantDec  ruleengine.Decision
		wantOK   bool
	}{
		{"CONTINUE", ruleengine.DecisionContinue, true},
		{"I think the driver should continue driving.", ruleengine.DecisionContinue, true},
		{"stop", ruleengine.DecisionStop, true},
		{"Stop now.", ruleengine.DecisionStop, true},
		{"maybe, unclear", "", false},
		{"CONTINUE but also STOP", "", false}, // both present -> unparseable
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := parse(c.raw)
		if ok != c.wantOK || (ok && got != c.wantDec) {
			t.Errorf("parse(%q) = (%q, %v), want (%q, %v)", c.raw, got, ok, c.wantDec, c.wantOK)
		}
	}
}

// newOfflineEngine builds a DecisionEngine as if Ollama was unreachable at
// startup, without paying the real 3-second probe timeout — exercising
// exactly the "Nếu Ollama chết → 100% Rule Engine" code path.
func newOfflineEngine() *DecisionEngine {
	return &DecisionEngine{
		cache:   NewResponseCache(),
		breaker: NewCircuitBreaker(5, 30*time.Second),
		timeout: time.Second,
		enabled: false,
	}
}

func TestDecisionEngine_FallsBackToRuleWhenDisabled(t *testing.T) {
	e := newOfflineEngine()
	result := e.Decide(context.Background(), DecisionRequest{
		Prompt:   "should the driver continue?",
		Parse:    ParseBinary("CONTINUE", ruleengine.DecisionContinue, "STOP", ruleengine.DecisionStop),
		Fallback: ruleengine.DecisionStop,
	})
	if result.Source != "rule_fallback" || result.Decision != ruleengine.DecisionStop {
		t.Errorf("expected a rule_fallback result when AI is disabled, got %+v", result)
	}

	snap := e.StatsSnapshot()
	if snap.RuleFallbackUsed != 1 || snap.AICalls != 0 {
		t.Errorf("expected 1 rule fallback and 0 AI calls, got %+v", snap)
	}
}

func TestDecisionEngine_CacheHitAvoidsRuleFallbackAndAICall(t *testing.T) {
	e := newOfflineEngine()
	// Pre-seed the cache as if a prior (real) AI call had already answered
	// this exact prompt.
	e.cache.Put("should the driver continue?", "CONTINUE")

	result := e.Decide(context.Background(), DecisionRequest{
		Prompt:   "should the driver continue?",
		Parse:    ParseBinary("CONTINUE", ruleengine.DecisionContinue, "STOP", ruleengine.DecisionStop),
		Fallback: ruleengine.DecisionStop,
	})
	if result.Source != "cache" || result.Decision != ruleengine.DecisionContinue {
		t.Errorf("expected a cache hit to short-circuit both the rule fallback and any AI call, got %+v", result)
	}

	snap := e.StatsSnapshot()
	if snap.CacheHits != 1 || snap.AICalls != 0 || snap.RuleFallbackUsed != 0 {
		t.Errorf("expected exactly 1 cache hit and no AI calls/fallbacks, got %+v", snap)
	}
}

func TestDecisionEngine_IdenticalPromptsCachedAcrossCalls(t *testing.T) {
	e := newOfflineEngine()
	e.cache.Put("same prompt", "STOP")

	parse := ParseBinary("CONTINUE", ruleengine.DecisionContinue, "STOP", ruleengine.DecisionStop)
	for i := 0; i < 5; i++ {
		result := e.Decide(context.Background(), DecisionRequest{
			Prompt: "same prompt", Parse: parse, Fallback: ruleengine.DecisionContinue,
		})
		if result.Source != "cache" {
			t.Errorf("call %d: expected repeated identical prompts to stay cache hits, got source=%q", i, result.Source)
		}
	}
	snap := e.StatsSnapshot()
	if snap.CacheHits != 5 {
		t.Errorf("expected 5 cache hits total, got %d", snap.CacheHits)
	}
	if snap.CacheSize != 1 {
		t.Errorf("expected the cache to still hold exactly 1 distinct prompt, got %d", snap.CacheSize)
	}
}

func TestDecisionEngine_OpenCircuitFallsBackWithoutCallingClient(t *testing.T) {
	e := newOfflineEngine()
	e.enabled = true // pretend Ollama was reachable at startup...
	e.breaker.RecordFailure()
	e.breaker.RecordFailure()
	e.breaker.RecordFailure()
	e.breaker.RecordFailure()
	e.breaker.RecordFailure() // ...but 5 consecutive failures have since opened the breaker

	result := e.Decide(context.Background(), DecisionRequest{
		Prompt:   "unseen prompt",
		Parse:    ParseBinary("USE", ruleengine.DecisionUseVoucher, "KEEP", ruleengine.DecisionKeepVoucher),
		Fallback: ruleengine.DecisionKeepVoucher,
	})
	if result.Source != "rule_fallback" || result.Decision != ruleengine.DecisionKeepVoucher {
		t.Errorf("expected an open circuit to fall back to the rule decision without a client call, got %+v", result)
	}
}

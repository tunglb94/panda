package aiengine

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fairride/ai_simulation/ruleengine"
)

// DecisionEngine is the single entry point the simulation calls for the 5%
// of decisions the rule engine marked ambiguous. It is safe to construct
// with a nil/unreachable Ollama (Enabled will be false) — every call then
// falls straight through to the rule's own fallback decision, satisfying
// "Nếu Ollama chết → Simulation vẫn chạy 100% Rule Engine".
type DecisionEngine struct {
	client  *OllamaClient
	cache   *ResponseCache
	breaker *CircuitBreaker
	timeout time.Duration
	enabled bool

	aiCalls      int64
	cacheHits    int64
	ruleFallback int64
	timeouts     int64

	latencyMu    sync.Mutex
	totalLatency time.Duration
}

// NewDecisionEngine probes Ollama once at construction time (short timeout)
// so the simulation knows upfront whether AI is available at all — a dead
// Ollama should not cost a multi-second HTTP timeout on every single one of
// potentially thousands of ambiguous decisions during the run.
func NewDecisionEngine(baseURL, model string, callTimeout time.Duration) *DecisionEngine {
	client := NewOllamaClient(baseURL, model, callTimeout)
	e := &DecisionEngine{
		client:  client,
		cache:   NewResponseCache(),
		breaker: NewCircuitBreaker(5, 30*time.Second),
		timeout: callTimeout,
	}
	probeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	e.enabled = client.Ping(probeCtx) == nil
	return e
}

// Enabled reports whether Ollama was reachable at startup. Even when true,
// individual calls can still fall back (timeout, circuit open, unparseable
// response) — this only reflects the initial probe.
func (e *DecisionEngine) Enabled() bool { return e.enabled }

// ParseFunc turns the model's raw text into one of exactly two valid
// decisions, or reports false if the response didn't clearly match either.
type ParseFunc func(raw string) (ruleengine.Decision, bool)

// DecisionRequest bundles what Decide needs: the exact prompt (used as the
// cache key), a parser for the two valid answers, and the rule engine's own
// fallback decision to use when AI is unavailable/unparseable/times out.
type DecisionRequest struct {
	Prompt   string
	Parse    ParseFunc
	Fallback ruleengine.Decision
}

// DecisionResult is what Decide returns.
type DecisionResult struct {
	Decision ruleengine.Decision
	Source   string // "ai" | "cache" | "rule_fallback"
}

// Decide resolves one ambiguous decision: cache -> circuit-gated Ollama call
// -> rule fallback, in that order. Never returns an error — a fully-offline
// simulation run is a valid, expected mode of operation, not a failure.
func (e *DecisionEngine) Decide(ctx context.Context, req DecisionRequest) DecisionResult {
	if cached, ok := e.cache.Get(req.Prompt); ok {
		e.cache.RecordHit()
		atomic.AddInt64(&e.cacheHits, 1)
		if d, ok := req.Parse(cached); ok {
			return DecisionResult{Decision: d, Source: "cache"}
		}
		// Cached-but-unparseable (shouldn't normally happen) — fall through
		// to rule fallback rather than re-calling AI for a value already
		// known to be bad.
		atomic.AddInt64(&e.ruleFallback, 1)
		return DecisionResult{Decision: req.Fallback, Source: "rule_fallback"}
	}
	e.cache.RecordMiss()

	if !e.enabled || !e.breaker.Allow() {
		atomic.AddInt64(&e.ruleFallback, 1)
		return DecisionResult{Decision: req.Fallback, Source: "rule_fallback"}
	}

	callCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	start := time.Now()
	raw, err := e.client.Generate(callCtx, req.Prompt)
	elapsed := time.Since(start)
	e.latencyMu.Lock()
	e.totalLatency += elapsed
	e.latencyMu.Unlock()

	if err != nil {
		e.breaker.RecordFailure()
		if callCtx.Err() != nil {
			atomic.AddInt64(&e.timeouts, 1)
		}
		atomic.AddInt64(&e.ruleFallback, 1)
		return DecisionResult{Decision: req.Fallback, Source: "rule_fallback"}
	}
	e.breaker.RecordSuccess()
	atomic.AddInt64(&e.aiCalls, 1)

	e.cache.Put(req.Prompt, raw)

	if d, ok := req.Parse(raw); ok {
		return DecisionResult{Decision: d, Source: "ai"}
	}
	atomic.AddInt64(&e.ruleFallback, 1)
	return DecisionResult{Decision: req.Fallback, Source: "rule_fallback"}
}

// GenerateReport makes one free-form report-writing call (used by the
// insights package to write simulation_summary.md/business_recommendation.md)
// — reuses this engine's enabled/circuit-breaker state so a already-known-dead
// Ollama doesn't also eat this call's timeout, but is NOT cached (each report
// call's prompt embeds the full, run-specific facts JSON, so identical
// prompts across runs are not expected the way the 4 decision types'
// bucketed prompts are). ok=false on any failure — the caller must fall back
// to a deterministic renderer, per this whole simulation's "Nếu Ollama chết
// -> vẫn chạy" discipline applying equally to report generation.
func (e *DecisionEngine) GenerateReport(ctx context.Context, prompt string) (text string, ok bool) {
	if !e.enabled || !e.breaker.Allow() {
		return "", false
	}
	callCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	raw, err := e.client.GenerateReport(callCtx, prompt)
	if err != nil {
		e.breaker.RecordFailure()
		return "", false
	}
	e.breaker.RecordSuccess()
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	return raw, true
}

// Stats snapshots the engine's counters for the benchmark report.
type Stats struct {
	AICalls          int64
	CacheHits        int64
	RuleFallbackUsed int64
	Timeouts         int64
	CacheSize        int
	AvgAILatencyMS   float64
	CircuitOpen      bool
}

func (e *DecisionEngine) StatsSnapshot() Stats {
	e.latencyMu.Lock()
	total := e.totalLatency
	e.latencyMu.Unlock()

	calls := atomic.LoadInt64(&e.aiCalls)
	var avg float64
	if calls > 0 {
		avg = float64(total.Milliseconds()) / float64(calls)
	}
	return Stats{
		AICalls:          calls,
		CacheHits:        atomic.LoadInt64(&e.cacheHits),
		RuleFallbackUsed: atomic.LoadInt64(&e.ruleFallback),
		Timeouts:         atomic.LoadInt64(&e.timeouts),
		CacheSize:        e.cache.Size(),
		AvgAILatencyMS:   avg,
		CircuitOpen:      e.breaker.IsOpen(),
	}
}

// ParseBinary is the shared parser for every decision family in this
// simulation: exactly two valid single-word answers, matched
// case-insensitively anywhere in the model's response (models sometimes
// wrap the answer in a short sentence despite the prompt's instruction).
func ParseBinary(trueWord string, trueDecision ruleengine.Decision, falseWord string, falseDecision ruleengine.Decision) ParseFunc {
	return func(raw string) (ruleengine.Decision, bool) {
		upper := strings.ToUpper(raw)
		hasTrue := strings.Contains(upper, strings.ToUpper(trueWord))
		hasFalse := strings.Contains(upper, strings.ToUpper(falseWord))
		switch {
		case hasTrue && !hasFalse:
			return trueDecision, true
		case hasFalse && !hasTrue:
			return falseDecision, true
		default:
			return "", false
		}
	}
}

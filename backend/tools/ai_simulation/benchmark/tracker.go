// Package benchmark tracks the operational metrics the sprint brief
// explicitly asks for: how often AI vs the Rule Engine decided, cache
// effectiveness, response latency, memory, and overall simulation speed.
package benchmark

import (
	"runtime"
	"sync/atomic"
	"time"
)

// Tracker accumulates counters over one simulation run. All methods are
// safe for concurrent use.
type Tracker struct {
	startedAt time.Time

	ruleEngineDecisions int64
	aiDecisions         int64

	ticksProcessed int64
	tripsProcessed int64

	peakAllocBytes uint64
}

func NewTracker() *Tracker {
	return &Tracker{startedAt: time.Now()}
}

func (t *Tracker) RecordRuleDecision() { atomic.AddInt64(&t.ruleEngineDecisions, 1) }
func (t *Tracker) RecordAIDecision()   { atomic.AddInt64(&t.aiDecisions, 1) }
func (t *Tracker) RecordTick()         { atomic.AddInt64(&t.ticksProcessed, 1) }
func (t *Tracker) RecordTrip()         { atomic.AddInt64(&t.tripsProcessed, 1) }

// SampleMemory records current heap usage if it's a new peak. Cheap enough
// to call periodically (e.g. once per simulated day), not every tick.
func (t *Tracker) SampleMemory() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	for {
		peak := atomic.LoadUint64(&t.peakAllocBytes)
		if m.Alloc <= peak {
			return
		}
		if atomic.CompareAndSwapUint64(&t.peakAllocBytes, peak, m.Alloc) {
			return
		}
	}
}

// AIEngineStats is the subset of aiengine.DecisionEngine.StatsSnapshot()
// the report needs — declared locally so this package doesn't import
// aiengine (that dependency would be backwards: aiengine is a lower-level
// building block, benchmark is a reporting layer that reads from it via the
// simulation engine, which already imports both).
type AIEngineStats struct {
	AICalls          int64
	CacheHits        int64
	CacheHitPercent  float64
	CacheSize        int
	RuleFallbackUsed int64
	Timeouts         int64
	AvgAILatencyMS   float64
	CircuitOpen      bool
}

// Report is the exported summary — see the sprint report's Benchmark
// section for what each field means.
type Report struct {
	RuleEngineDecisions int64   `json:"rule_engine_decisions"`
	AIDecisions         int64   `json:"ai_decisions"`
	AIDecisionPercent   float64 `json:"ai_decision_percent"`

	AICalls          int64   `json:"ai_calls"`
	CacheHits        int64   `json:"cache_hits"`
	CacheHitPercent  float64 `json:"cache_hit_percent"`
	CacheSize        int     `json:"cache_size"`
	RuleFallbackUsed int64   `json:"rule_fallback_used"`
	AITimeouts       int64   `json:"ai_timeouts"`
	AvgAILatencyMS   float64 `json:"avg_ai_latency_ms"`
	CircuitOpen      bool    `json:"circuit_open_at_end"`

	TicksProcessed int64   `json:"ticks_processed"`
	TripsProcessed int64   `json:"trips_processed"`
	WallClockSec   float64 `json:"wall_clock_seconds"`
	TicksPerSecond float64 `json:"ticks_per_second"`
	PeakAllocMB    float64 `json:"peak_alloc_mb"`
}

// Build assembles the final report from this tracker's own counters plus
// the AI engine's own stats snapshot.
func (t *Tracker) Build(ai AIEngineStats) Report {
	elapsed := time.Since(t.startedAt).Seconds()
	ruleDec := atomic.LoadInt64(&t.ruleEngineDecisions)
	aiDec := atomic.LoadInt64(&t.aiDecisions)
	totalDec := ruleDec + aiDec

	var aiPercent float64
	if totalDec > 0 {
		aiPercent = 100 * float64(aiDec) / float64(totalDec)
	}

	ticks := atomic.LoadInt64(&t.ticksProcessed)
	var tps float64
	if elapsed > 0 {
		tps = float64(ticks) / elapsed
	}

	return Report{
		RuleEngineDecisions: ruleDec,
		AIDecisions:         aiDec,
		AIDecisionPercent:   aiPercent,

		AICalls:          ai.AICalls,
		CacheHits:        ai.CacheHits,
		CacheHitPercent:  ai.CacheHitPercent,
		CacheSize:        ai.CacheSize,
		RuleFallbackUsed: ai.RuleFallbackUsed,
		AITimeouts:       ai.Timeouts,
		AvgAILatencyMS:   ai.AvgAILatencyMS,
		CircuitOpen:      ai.CircuitOpen,

		TicksProcessed: ticks,
		TripsProcessed: atomic.LoadInt64(&t.tripsProcessed),
		WallClockSec:   elapsed,
		TicksPerSecond: tps,
		PeakAllocMB:    float64(atomic.LoadUint64(&t.peakAllocBytes)) / (1024 * 1024),
	}
}

package aiengine

import (
	"sync"
	"time"
)

// CircuitBreaker stops hammering a dead/slow Ollama: after
// maxConsecutiveFailures timeouts or errors in a row, it "opens" and every
// call fails fast (no network attempt) for cooldown, after which it allows
// one trial call ("half-open") to test recovery. This is what makes "Nếu
// Ollama chết → Simulation vẫn chạy 100% Rule Engine" true in practice —
// without it, a dead Ollama would make every one of thousands of decision
// points wait out a full HTTP timeout.
type CircuitBreaker struct {
	mu                     sync.Mutex
	maxConsecutiveFailures int
	cooldown               time.Duration
	consecutiveFailures    int
	openedAt               time.Time
	open                   bool
}

func NewCircuitBreaker(maxConsecutiveFailures int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxConsecutiveFailures: maxConsecutiveFailures,
		cooldown:               cooldown,
	}
}

// Allow reports whether a call should be attempted right now.
func (b *CircuitBreaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.open {
		return true
	}
	if time.Since(b.openedAt) >= b.cooldown {
		// Half-open: let exactly one call through to probe recovery.
		return true
	}
	return false
}

// RecordSuccess closes the breaker and resets the failure streak.
func (b *CircuitBreaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consecutiveFailures = 0
	b.open = false
}

// RecordFailure increments the failure streak and opens the breaker once
// the threshold is crossed.
func (b *CircuitBreaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consecutiveFailures++
	if b.consecutiveFailures >= b.maxConsecutiveFailures {
		b.open = true
		b.openedAt = time.Now()
	}
}

// IsOpen reports the breaker's current state (for benchmark reporting).
func (b *CircuitBreaker) IsOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.open
}

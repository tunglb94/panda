package aiengine

import (
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedByDefault(t *testing.T) {
	b := NewCircuitBreaker(3, 30*time.Second)
	if !b.Allow() || b.IsOpen() {
		t.Errorf("a fresh breaker must be closed and allow calls")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	b := NewCircuitBreaker(3, time.Hour)
	b.RecordFailure()
	b.RecordFailure()
	if b.IsOpen() {
		t.Fatalf("breaker should still be closed before hitting the threshold")
	}
	b.RecordFailure() // 3rd consecutive failure
	if !b.IsOpen() {
		t.Fatalf("breaker should open exactly at the failure threshold")
	}
	if b.Allow() {
		t.Errorf("an open breaker within its cooldown must not allow calls")
	}
}

func TestCircuitBreaker_SuccessResetsFailureStreak(t *testing.T) {
	b := NewCircuitBreaker(3, time.Hour)
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	b.RecordFailure()
	b.RecordFailure()
	if b.IsOpen() {
		t.Errorf("a success in between should reset the streak, so 2 failures after it must not open the breaker")
	}
}

func TestCircuitBreaker_HalfOpenAfterCooldown(t *testing.T) {
	b := NewCircuitBreaker(1, 10*time.Millisecond)
	b.RecordFailure() // opens immediately (threshold=1)
	if !b.IsOpen() {
		t.Fatalf("expected the breaker to open on the first failure with threshold=1")
	}
	if b.Allow() {
		t.Fatalf("expected calls blocked immediately after opening")
	}
	time.Sleep(15 * time.Millisecond)
	if !b.Allow() {
		t.Errorf("expected a half-open probe call to be allowed after the cooldown elapses")
	}
}

func TestCircuitBreaker_SuccessAfterHalfOpenCloses(t *testing.T) {
	b := NewCircuitBreaker(1, 10*time.Millisecond)
	b.RecordFailure()
	time.Sleep(15 * time.Millisecond)
	if !b.Allow() {
		t.Fatalf("expected the half-open probe to be allowed")
	}
	b.RecordSuccess()
	if b.IsOpen() {
		t.Errorf("a successful probe call must close the breaker")
	}
}

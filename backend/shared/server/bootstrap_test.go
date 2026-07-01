package server_test

import (
	"testing"

	"github.com/fairride/shared/server"
)

func TestReadinessTracker_EmptyIsReady(t *testing.T) {
	r := server.NewReadinessTracker()
	if !r.IsReady() {
		t.Error("empty tracker should be ready (no checks = all pass)")
	}
}

func TestReadinessTracker_AllPass(t *testing.T) {
	r := server.NewReadinessTracker()
	r.Set("db", true)
	r.Set("redis", true)
	if !r.IsReady() {
		t.Error("tracker with all passing checks should be ready")
	}
}

func TestReadinessTracker_OneFails(t *testing.T) {
	r := server.NewReadinessTracker()
	r.Set("db", true)
	r.Set("redis", false)
	if r.IsReady() {
		t.Error("tracker with one failing check must not be ready")
	}
}

func TestReadinessTracker_RecoverAfterFail(t *testing.T) {
	r := server.NewReadinessTracker()
	r.Set("db", false)
	if r.IsReady() {
		t.Error("tracker with failing check must not be ready")
	}
	r.Set("db", true)
	if !r.IsReady() {
		t.Error("tracker should be ready after check recovers")
	}
}

func TestReadinessTracker_ConcurrentSafe(t *testing.T) {
	r := server.NewReadinessTracker()
	done := make(chan struct{})

	go func() {
		for range 1000 {
			r.Set("check", true)
		}
		close(done)
	}()

	for range 1000 {
		_ = r.IsReady()
	}
	<-done
}

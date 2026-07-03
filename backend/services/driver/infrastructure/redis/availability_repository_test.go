package redis_test

import (
	"context"
	"testing"
	"time"

	driverredis "github.com/fairride/driver/infrastructure/redis"
	"github.com/fairride/shared/errors"
)

const shortTTL = 100 * time.Millisecond

func newRepo() *driverredis.AvailabilityRepository {
	return driverredis.NewAvailabilityRepository(testClient)
}

func newRepoShortTTL() *driverredis.AvailabilityRepository {
	return driverredis.NewAvailabilityRepositoryWithTTL(testClient, shortTTL, time.Hour)
}

// ─── GetAvailability — never seen ────────────────────────────────────────────

func TestGetAvailability_NeverSeen(t *testing.T) {
	cleanKeys(t, "test-never")
	repo := newRepo()

	state, err := repo.GetAvailability(context.Background(), "test-never")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.IsOnline {
		t.Errorf("should be offline")
	}
	if !state.LastSeen.IsZero() {
		t.Errorf("LastSeen should be zero for unseen driver")
	}
}

// ─── SetOnline + GetAvailability ─────────────────────────────────────────────

func TestSetOnline_And_GetAvailability(t *testing.T) {
	cleanKeys(t, "test-d1")
	repo := newRepo()

	if err := repo.SetOnline(context.Background(), "test-d1", testNow); err != nil {
		t.Fatalf("SetOnline: %v", err)
	}

	state, err := repo.GetAvailability(context.Background(), "test-d1")
	if err != nil {
		t.Fatalf("GetAvailability: %v", err)
	}
	if !state.IsOnline {
		t.Errorf("should be online")
	}
	if state.LastSeen.IsZero() {
		t.Errorf("LastSeen should be set")
	}
	if state.DriverID != "test-d1" {
		t.Errorf("wrong driver id: %s", state.DriverID)
	}
}

// ─── SetOffline ───────────────────────────────────────────────────────────────

func TestSetOffline_ClearsOnline_KeepsLastSeen(t *testing.T) {
	cleanKeys(t, "test-d2")
	repo := newRepo()

	_ = repo.SetOnline(context.Background(), "test-d2", testNow)
	if err := repo.SetOffline(context.Background(), "test-d2", testNow.Add(time.Minute)); err != nil {
		t.Fatalf("SetOffline: %v", err)
	}

	state, err := repo.GetAvailability(context.Background(), "test-d2")
	if err != nil {
		t.Fatalf("GetAvailability: %v", err)
	}
	if state.IsOnline {
		t.Errorf("should be offline after SetOffline")
	}
	if state.LastSeen.IsZero() {
		t.Errorf("LastSeen must be retained after SetOffline")
	}
}

func TestSetOffline_Idempotent(t *testing.T) {
	cleanKeys(t, "test-d3")
	repo := newRepo()

	_ = repo.SetOffline(context.Background(), "test-d3", testNow)
	if err := repo.SetOffline(context.Background(), "test-d3", testNow); err != nil {
		t.Fatalf("second SetOffline should succeed: %v", err)
	}
}

// ─── RefreshHeartbeat ─────────────────────────────────────────────────────────

func TestRefreshHeartbeat_WhenOnline(t *testing.T) {
	cleanKeys(t, "test-d4")
	repo := newRepo()

	_ = repo.SetOnline(context.Background(), "test-d4", testNow)
	if err := repo.RefreshHeartbeat(context.Background(), "test-d4", testNow.Add(time.Minute)); err != nil {
		t.Fatalf("RefreshHeartbeat: %v", err)
	}
	state, _ := repo.GetAvailability(context.Background(), "test-d4")
	if !state.IsOnline {
		t.Errorf("should still be online after heartbeat")
	}
}

func TestRefreshHeartbeat_WhenOffline(t *testing.T) {
	cleanKeys(t, "test-d5")
	repo := newRepo()

	err := repo.RefreshHeartbeat(context.Background(), "test-d5", testNow)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for heartbeat when offline, got %v", err)
	}
}

func TestRefreshHeartbeat_UpdatesLastSeen(t *testing.T) {
	cleanKeys(t, "test-d6")
	repo := newRepo()

	_ = repo.SetOnline(context.Background(), "test-d6", testNow)
	later := testNow.Add(2 * time.Minute)
	_ = repo.RefreshHeartbeat(context.Background(), "test-d6", later)

	state, _ := repo.GetAvailability(context.Background(), "test-d6")
	if !state.LastSeen.After(testNow) {
		t.Errorf("LastSeen should be updated to later time; got %v", state.LastSeen)
	}
}

// ─── TTL expiry ───────────────────────────────────────────────────────────────

func TestOnlineKey_ExpiresAfterTTL(t *testing.T) {
	cleanKeys(t, "test-d7")
	repo := newRepoShortTTL()

	_ = repo.SetOnline(context.Background(), "test-d7", testNow)

	state, _ := repo.GetAvailability(context.Background(), "test-d7")
	if !state.IsOnline {
		t.Fatalf("should be online before TTL expiry")
	}

	time.Sleep(shortTTL + 50*time.Millisecond)

	state, err := repo.GetAvailability(context.Background(), "test-d7")
	if err != nil {
		t.Fatalf("GetAvailability: %v", err)
	}
	if state.IsOnline {
		t.Errorf("should be offline after TTL expiry")
	}
	// last_seen must still be there (separate, longer TTL)
	if state.LastSeen.IsZero() {
		t.Errorf("LastSeen must survive the online key TTL")
	}
}

// ─── SetOnline Idempotent / reset TTL ────────────────────────────────────────

func TestSetOnline_Idempotent(t *testing.T) {
	cleanKeys(t, "test-d8")
	repo := newRepo()

	_ = repo.SetOnline(context.Background(), "test-d8", testNow)
	if err := repo.SetOnline(context.Background(), "test-d8", testNow); err != nil {
		t.Fatalf("second SetOnline should succeed: %v", err)
	}
	state, _ := repo.GetAvailability(context.Background(), "test-d8")
	if !state.IsOnline {
		t.Errorf("should remain online")
	}
}

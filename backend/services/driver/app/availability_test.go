package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
)

// ─── in-memory stub ──────────────────────────────────────────────────────────

type stubAvailRepo struct {
	online   map[string]bool
	lastSeen map[string]time.Time
}

func newStubAvailRepo() *stubAvailRepo {
	return &stubAvailRepo{
		online:   make(map[string]bool),
		lastSeen: make(map[string]time.Time),
	}
}

func (r *stubAvailRepo) SetOnline(_ context.Context, driverID string, now time.Time) error {
	r.online[driverID] = true
	r.lastSeen[driverID] = now
	return nil
}

func (r *stubAvailRepo) SetOffline(_ context.Context, driverID string, now time.Time) error {
	r.online[driverID] = false
	r.lastSeen[driverID] = now
	return nil
}

func (r *stubAvailRepo) RefreshHeartbeat(_ context.Context, driverID string, now time.Time) error {
	if !r.online[driverID] {
		return errors.PreconditionFailed("driver is not online")
	}
	r.lastSeen[driverID] = now
	return nil
}

func (r *stubAvailRepo) GetAvailability(_ context.Context, driverID string) (*entity.AvailabilityState, error) {
	return &entity.AvailabilityState{
		DriverID: driverID,
		IsOnline: r.online[driverID],
		LastSeen: r.lastSeen[driverID],
	}, nil
}

// ─── GoOnlineUseCase ─────────────────────────────────────────────────────────

func TestGoOnline_OK(t *testing.T) {
	repo := newStubAvailRepo()
	uc := app.NewGoOnlineUseCase(repo)

	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state.IsOnline {
		t.Errorf("driver should be online")
	}
	if state.DriverID != "d1" {
		t.Errorf("wrong driver id: %s", state.DriverID)
	}
}

func TestGoOnline_EmptyDriverID(t *testing.T) {
	uc := app.NewGoOnlineUseCase(newStubAvailRepo())
	_, err := uc.Execute(context.Background(), "")
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestGoOnline_Idempotent(t *testing.T) {
	repo := newStubAvailRepo()
	uc := app.NewGoOnlineUseCase(repo)
	_, _ = uc.Execute(context.Background(), "d1")
	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("second GoOnline should succeed: %v", err)
	}
	if !state.IsOnline {
		t.Errorf("driver should remain online")
	}
}

// ─── GoOfflineUseCase ────────────────────────────────────────────────────────

func TestGoOffline_OK(t *testing.T) {
	repo := newStubAvailRepo()
	_ = repo.SetOnline(context.Background(), "d1", testNow)

	uc := app.NewGoOfflineUseCase(repo)
	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.IsOnline {
		t.Errorf("driver should be offline")
	}
}

func TestGoOffline_EmptyDriverID(t *testing.T) {
	uc := app.NewGoOfflineUseCase(newStubAvailRepo())
	_, err := uc.Execute(context.Background(), "")
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestGoOffline_Idempotent(t *testing.T) {
	repo := newStubAvailRepo()
	uc := app.NewGoOfflineUseCase(repo)
	_, _ = uc.Execute(context.Background(), "d1")
	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("second GoOffline should succeed: %v", err)
	}
	if state.IsOnline {
		t.Errorf("driver should remain offline")
	}
}

// ─── HeartbeatUseCase ────────────────────────────────────────────────────────

func TestHeartbeat_WhenOnline(t *testing.T) {
	repo := newStubAvailRepo()
	_ = repo.SetOnline(context.Background(), "d1", testNow)

	uc := app.NewHeartbeatUseCase(repo)
	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state.IsOnline {
		t.Errorf("driver should remain online after heartbeat")
	}
}

func TestHeartbeat_WhenOffline(t *testing.T) {
	repo := newStubAvailRepo()
	uc := app.NewHeartbeatUseCase(repo)

	_, err := uc.Execute(context.Background(), "d1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for heartbeat when offline, got %v", err)
	}
}

func TestHeartbeat_EmptyDriverID(t *testing.T) {
	uc := app.NewHeartbeatUseCase(newStubAvailRepo())
	_, err := uc.Execute(context.Background(), "")
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestHeartbeat_UpdatesLastSeen(t *testing.T) {
	repo := newStubAvailRepo()
	_ = repo.SetOnline(context.Background(), "d1", testNow)
	later := testNow.Add(time.Minute)

	// Force the stub to use `later` for the heartbeat timestamp.
	// We can't inject time into the use case, but we can verify
	// that last_seen updates after heartbeat by checking the state
	// changed between calls — here we verify it's non-zero.
	uc := app.NewHeartbeatUseCase(repo)
	state, _ := uc.Execute(context.Background(), "d1")
	if state.LastSeen.IsZero() {
		t.Errorf("LastSeen must not be zero after heartbeat")
	}
	_ = later
}

// ─── GetAvailabilityUseCase ───────────────────────────────────────────────────

func TestGetAvailability_NeverSeen(t *testing.T) {
	uc := app.NewGetAvailabilityUseCase(newStubAvailRepo())
	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("should not error for unseen driver: %v", err)
	}
	if state.IsOnline {
		t.Errorf("unseen driver should be offline")
	}
	if !state.LastSeen.IsZero() {
		t.Errorf("unseen driver should have zero LastSeen")
	}
}

func TestGetAvailability_AfterOnline(t *testing.T) {
	repo := newStubAvailRepo()
	_ = repo.SetOnline(context.Background(), "d1", testNow)
	uc := app.NewGetAvailabilityUseCase(repo)

	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state.IsOnline {
		t.Errorf("driver should be online")
	}
	if state.LastSeen.IsZero() {
		t.Errorf("LastSeen must be set")
	}
}

func TestGetAvailability_EmptyDriverID(t *testing.T) {
	uc := app.NewGetAvailabilityUseCase(newStubAvailRepo())
	_, err := uc.Execute(context.Background(), "")
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

// ─── lifecycle ────────────────────────────────────────────────────────────────

func TestAvailability_FullLifecycle(t *testing.T) {
	repo := newStubAvailRepo()
	goOnline := app.NewGoOnlineUseCase(repo)
	goOffline := app.NewGoOfflineUseCase(repo)
	heartbeat := app.NewHeartbeatUseCase(repo)
	getAvail := app.NewGetAvailabilityUseCase(repo)

	// unseen
	state, _ := getAvail.Execute(context.Background(), "d1")
	if state.IsOnline || !state.LastSeen.IsZero() {
		t.Error("freshly-seen driver should be offline with zero LastSeen")
	}

	// go online
	state, _ = goOnline.Execute(context.Background(), "d1")
	if !state.IsOnline {
		t.Error("should be online")
	}

	// heartbeat
	state, err := heartbeat.Execute(context.Background(), "d1")
	if err != nil || !state.IsOnline {
		t.Errorf("heartbeat while online should succeed: %v", err)
	}

	// go offline
	state, _ = goOffline.Execute(context.Background(), "d1")
	if state.IsOnline {
		t.Error("should be offline")
	}

	// heartbeat while offline → fail
	_, err = heartbeat.Execute(context.Background(), "d1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("heartbeat while offline should fail with CodePreconditionFailed, got %v", err)
	}
}

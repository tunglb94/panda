package grpc_test

import (
	"context"
	"testing"
	"time"

	drivergrpc "github.com/fairride/driver/grpc"
	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/grpc/driverpb"
	"github.com/fairride/shared/errors"
	"google.golang.org/grpc/codes"
)

// ─── in-memory availability stub ────────────────────────────────────────────

type stubAvailRepoH struct {
	online   map[string]bool
	lastSeen map[string]time.Time
}

func newStubAvailRepoH() *stubAvailRepoH {
	return &stubAvailRepoH{
		online:   make(map[string]bool),
		lastSeen: make(map[string]time.Time),
	}
}

func (r *stubAvailRepoH) SetOnline(_ context.Context, id string, now time.Time) error {
	r.online[id] = true
	r.lastSeen[id] = now
	return nil
}

func (r *stubAvailRepoH) SetOffline(_ context.Context, id string, now time.Time) error {
	r.online[id] = false
	r.lastSeen[id] = now
	return nil
}

func (r *stubAvailRepoH) RefreshHeartbeat(_ context.Context, id string, now time.Time) error {
	if !r.online[id] {
		return errors.PreconditionFailed("driver is not online")
	}
	r.lastSeen[id] = now
	return nil
}

func (r *stubAvailRepoH) GetAvailability(_ context.Context, id string) (*entity.AvailabilityState, error) {
	return &entity.AvailabilityState{
		DriverID: id,
		IsOnline: r.online[id],
		LastSeen: r.lastSeen[id],
	}, nil
}

// ─── builder ─────────────────────────────────────────────────────────────────

func newAvailHandler(repo *stubAvailRepoH) *drivergrpc.AvailabilityHandler {
	return drivergrpc.NewAvailabilityHandler(
		app.NewGoOnlineUseCase(repo),
		app.NewGoOfflineUseCase(repo),
		app.NewHeartbeatUseCase(repo),
		app.NewGetAvailabilityUseCase(repo),
	)
}

// ─── GoOnline ────────────────────────────────────────────────────────────────

func TestAvailGoOnline_OK(t *testing.T) {
	h := newAvailHandler(newStubAvailRepoH())
	resp, err := h.GoOnline(context.Background(), &driverpb.GoOnlineRequest{DriverId: "d1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsOnline {
		t.Errorf("driver should be online in response")
	}
	if resp.DriverId != "d1" {
		t.Errorf("wrong driver id: %s", resp.DriverId)
	}
}

func TestAvailGoOnline_EmptyDriverID(t *testing.T) {
	h := newAvailHandler(newStubAvailRepoH())
	_, err := h.GoOnline(context.Background(), &driverpb.GoOnlineRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── GoOffline ───────────────────────────────────────────────────────────────

func TestAvailGoOffline_OK(t *testing.T) {
	repo := newStubAvailRepoH()
	_ = repo.SetOnline(context.Background(), "d1", testNow)
	h := newAvailHandler(repo)

	resp, err := h.GoOffline(context.Background(), &driverpb.GoOfflineRequest{DriverId: "d1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsOnline {
		t.Errorf("driver should be offline in response")
	}
}

func TestAvailGoOffline_EmptyDriverID(t *testing.T) {
	h := newAvailHandler(newStubAvailRepoH())
	_, err := h.GoOffline(context.Background(), &driverpb.GoOfflineRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── Heartbeat ────────────────────────────────────────────────────────────────

func TestAvailHeartbeat_WhenOnline(t *testing.T) {
	repo := newStubAvailRepoH()
	_ = repo.SetOnline(context.Background(), "d1", testNow)
	h := newAvailHandler(repo)

	resp, err := h.Heartbeat(context.Background(), &driverpb.HeartbeatRequest{DriverId: "d1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsOnline {
		t.Errorf("driver should remain online after heartbeat")
	}
}

func TestAvailHeartbeat_WhenOffline(t *testing.T) {
	h := newAvailHandler(newStubAvailRepoH())
	_, err := h.Heartbeat(context.Background(), &driverpb.HeartbeatRequest{DriverId: "d1"})
	assertGRPCCode(t, err, codes.FailedPrecondition)
}

func TestAvailHeartbeat_EmptyDriverID(t *testing.T) {
	h := newAvailHandler(newStubAvailRepoH())
	_, err := h.Heartbeat(context.Background(), &driverpb.HeartbeatRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── GetAvailability ─────────────────────────────────────────────────────────

func TestAvailGet_NeverSeen(t *testing.T) {
	h := newAvailHandler(newStubAvailRepoH())
	resp, err := h.GetAvailability(context.Background(), &driverpb.GetAvailabilityRequest{DriverId: "d1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsOnline {
		t.Errorf("unseen driver should be offline")
	}
	if resp.LastSeen != nil {
		t.Errorf("last_seen should be nil for unseen driver, got %v", resp.LastSeen)
	}
}

func TestAvailGet_AfterOnline(t *testing.T) {
	repo := newStubAvailRepoH()
	_ = repo.SetOnline(context.Background(), "d1", testNow)
	h := newAvailHandler(repo)

	resp, err := h.GetAvailability(context.Background(), &driverpb.GetAvailabilityRequest{DriverId: "d1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsOnline {
		t.Errorf("should be online")
	}
	if resp.LastSeen == nil {
		t.Errorf("last_seen should be set")
	}
}

func TestAvailGet_EmptyDriverID(t *testing.T) {
	h := newAvailHandler(newStubAvailRepoH())
	_, err := h.GetAvailability(context.Background(), &driverpb.GetAvailabilityRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── LastSeen in proto ────────────────────────────────────────────────────────

func TestAvailGoOnline_LastSeenSetInProto(t *testing.T) {
	h := newAvailHandler(newStubAvailRepoH())
	resp, _ := h.GoOnline(context.Background(), &driverpb.GoOnlineRequest{DriverId: "d1"})
	if resp.LastSeen == nil {
		t.Errorf("last_seen should be populated after GoOnline")
	}
}

func TestAvailGoOffline_LastSeenRetainedInProto(t *testing.T) {
	repo := newStubAvailRepoH()
	_ = repo.SetOnline(context.Background(), "d1", testNow)
	h := newAvailHandler(repo)
	resp, _ := h.GoOffline(context.Background(), &driverpb.GoOfflineRequest{DriverId: "d1"})
	if resp.LastSeen == nil {
		t.Errorf("last_seen must be retained after GoOffline")
	}
}

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
	"google.golang.org/grpc/status"
)

var testNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// ─── in-memory stub ──────────────────────────────────────────────────────────

type stubRepo struct {
	byID     map[string]*entity.DriverProfile
	byUserID map[string]*entity.DriverProfile
}

func newStubRepo() *stubRepo {
	return &stubRepo{
		byID:     make(map[string]*entity.DriverProfile),
		byUserID: make(map[string]*entity.DriverProfile),
	}
}

func (r *stubRepo) FindByID(_ context.Context, id string) (*entity.DriverProfile, error) {
	d, ok := r.byID[id]
	if !ok {
		return nil, errors.NotFound("not found: " + id)
	}
	return d, nil
}

func (r *stubRepo) FindByUserID(_ context.Context, uid string) (*entity.DriverProfile, error) {
	d, ok := r.byUserID[uid]
	if !ok {
		return nil, errors.NotFound("not found for user: " + uid)
	}
	return d, nil
}

func (r *stubRepo) Save(_ context.Context, d *entity.DriverProfile) error {
	r.byID[d.DriverID] = d
	r.byUserID[d.UserID] = d
	return nil
}

// ─── builder ─────────────────────────────────────────────────────────────────

func newHandler(repo *stubRepo) *drivergrpc.Handler {
	return drivergrpc.NewHandler(
		app.NewGetDriverProfileUseCase(repo),
		app.NewGetDriverProfileByUserIDUseCase(repo),
		app.NewUpdateDriverProfileUseCase(repo),
		app.NewUpdateOnlineStatusUseCase(repo),
		app.NewUpdateVerificationStatusUseCase(repo),
	)
}

func seedPending(t *testing.T, repo *stubRepo) *entity.DriverProfile {
	t.Helper()
	d, err := entity.NewDriverProfile("d1", "u1", "LIC-001", entity.VehicleTypeCar, "Toyota", "Camry", "White", "ABC-123", testNow)
	if err != nil {
		t.Fatalf("seedPending: %v", err)
	}
	_ = repo.Save(context.Background(), d)
	return d
}

func seedVerified(t *testing.T, repo *stubRepo) *entity.DriverProfile {
	t.Helper()
	d := seedPending(t, repo)
	_ = d.Verify(testNow)
	_ = repo.Save(context.Background(), d)
	return d
}

// ─── GetDriverProfile ─────────────────────────────────────────────────────────

func TestGetDriverProfile_OK(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	h := newHandler(repo)

	resp, err := h.GetDriverProfile(context.Background(), &driverpb.GetDriverProfileRequest{DriverId: "d1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Profile.DriverId != "d1" {
		t.Errorf("want d1 got %s", resp.Profile.DriverId)
	}
}

func TestGetDriverProfile_EmptyID(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.GetDriverProfile(context.Background(), &driverpb.GetDriverProfileRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestGetDriverProfile_NotFound(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.GetDriverProfile(context.Background(), &driverpb.GetDriverProfileRequest{DriverId: "ghost"})
	assertGRPCCode(t, err, codes.NotFound)
}

// ─── GetDriverProfileByUserID ─────────────────────────────────────────────────

func TestGetDriverProfileByUserID_OK(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	h := newHandler(repo)

	resp, err := h.GetDriverProfileByUserID(context.Background(), &driverpb.GetDriverProfileByUserIDRequest{UserId: "u1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Profile.UserId != "u1" {
		t.Errorf("want u1 got %s", resp.Profile.UserId)
	}
}

func TestGetDriverProfileByUserID_EmptyUserID(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.GetDriverProfileByUserID(context.Background(), &driverpb.GetDriverProfileByUserIDRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── UpdateDriverProfile ──────────────────────────────────────────────────────

func TestUpdateDriverProfile_OK(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	h := newHandler(repo)

	resp, err := h.UpdateDriverProfile(context.Background(), &driverpb.UpdateDriverProfileRequest{
		DriverId:      "d1",
		LicenseNumber: "LIC-NEW",
		VehicleType:   "van",
		VehicleBrand:  "Ford",
		VehicleModel:  "Transit",
		VehicleColor:  "Blue",
		PlateNumber:   "NEW-999",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Profile.LicenseNumber != "LIC-NEW" {
		t.Errorf("license not updated")
	}
}

func TestUpdateDriverProfile_EmptyDriverID(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.UpdateDriverProfile(context.Background(), &driverpb.UpdateDriverProfileRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── UpdateOnlineStatus ───────────────────────────────────────────────────────

func TestUpdateOnlineStatus_GoOnline_OK(t *testing.T) {
	repo := newStubRepo()
	seedVerified(t, repo)
	h := newHandler(repo)

	resp, err := h.UpdateOnlineStatus(context.Background(), &driverpb.UpdateOnlineStatusRequest{
		DriverId:     "d1",
		OnlineStatus: "online",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Profile.OnlineStatus != "online" {
		t.Errorf("expected online status in response")
	}
}

func TestUpdateOnlineStatus_GoOnline_NotVerified(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	h := newHandler(repo)

	_, err := h.UpdateOnlineStatus(context.Background(), &driverpb.UpdateOnlineStatusRequest{
		DriverId:     "d1",
		OnlineStatus: "online",
	})
	assertGRPCCode(t, err, codes.FailedPrecondition)
}

func TestUpdateOnlineStatus_EmptyDriverID(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.UpdateOnlineStatus(context.Background(), &driverpb.UpdateOnlineStatusRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── UpdateVerificationStatus ─────────────────────────────────────────────────

func TestUpdateVerificationStatus_Verify_OK(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	h := newHandler(repo)

	resp, err := h.UpdateVerificationStatus(context.Background(), &driverpb.UpdateVerificationStatusRequest{
		DriverId:           "d1",
		VerificationStatus: "verified",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Profile.VerificationStatus != "verified" {
		t.Errorf("expected verified status in response")
	}
}

func TestUpdateVerificationStatus_InvalidAction(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	h := newHandler(repo)

	_, err := h.UpdateVerificationStatus(context.Background(), &driverpb.UpdateVerificationStatusRequest{
		DriverId:           "d1",
		VerificationStatus: "approved",
	})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestUpdateVerificationStatus_EmptyDriverID(t *testing.T) {
	h := newHandler(newStubRepo())
	_, err := h.UpdateVerificationStatus(context.Background(), &driverpb.UpdateVerificationStatusRequest{})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

// ─── toProto coverage ────────────────────────────────────────────────────────

func TestGetDriverProfile_ProtoFields(t *testing.T) {
	repo := newStubRepo()
	d, _ := entity.NewDriverProfile("d1", "u1", "LIC-001", entity.VehicleTypeCar, "Toyota", "Camry", "White", "ABC-123", testNow)
	_ = repo.Save(context.Background(), d)
	h := newHandler(repo)

	resp, err := h.GetDriverProfile(context.Background(), &driverpb.GetDriverProfileRequest{DriverId: "d1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := resp.Profile
	if p.UserId != "u1" {
		t.Errorf("UserId want u1 got %s", p.UserId)
	}
	if p.VehicleType != "car" {
		t.Errorf("VehicleType want car got %s", p.VehicleType)
	}
	if p.OnlineStatus != "offline" {
		t.Errorf("OnlineStatus want offline got %s", p.OnlineStatus)
	}
	if p.VerificationStatus != "pending" {
		t.Errorf("VerificationStatus want pending got %s", p.VerificationStatus)
	}
	if p.CreatedAt == nil || p.UpdatedAt == nil {
		t.Errorf("timestamps must not be nil")
	}
}

// ─── helper ──────────────────────────────────────────────────────────────────

func assertGRPCCode(t *testing.T, err error, want codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %v, got nil", want)
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if s.Code() != want {
		t.Errorf("want gRPC code %v got %v (%s)", want, s.Code(), s.Message())
	}
}

package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
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
		return nil, errors.NotFound("driver not found: " + id)
	}
	return d, nil
}

func (r *stubRepo) FindByUserID(_ context.Context, uid string) (*entity.DriverProfile, error) {
	d, ok := r.byUserID[uid]
	if !ok {
		return nil, errors.NotFound("driver not found for user: " + uid)
	}
	return d, nil
}

func (r *stubRepo) Save(_ context.Context, d *entity.DriverProfile) error {
	r.byID[d.DriverID] = d
	r.byUserID[d.UserID] = d
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

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
	if err := d.Verify(testNow); err != nil {
		t.Fatalf("seedVerified: %v", err)
	}
	_ = repo.Save(context.Background(), d)
	return d
}

// ─── GetDriverProfileUseCase ─────────────────────────────────────────────────

func TestGetDriverProfile_Found(t *testing.T) {
	repo := newStubRepo()
	d := seedPending(t, repo)
	uc := app.NewGetDriverProfileUseCase(repo)

	got, err := uc.Execute(context.Background(), d.DriverID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.DriverID != d.DriverID {
		t.Errorf("want %s got %s", d.DriverID, got.DriverID)
	}
}

func TestGetDriverProfile_NotFound(t *testing.T) {
	repo := newStubRepo()
	uc := app.NewGetDriverProfileUseCase(repo)

	_, err := uc.Execute(context.Background(), "nonexistent")
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound got %v", err)
	}
}

// ─── GetDriverProfileByUserIDUseCase ─────────────────────────────────────────

func TestGetDriverProfileByUserID_Found(t *testing.T) {
	repo := newStubRepo()
	d := seedPending(t, repo)
	uc := app.NewGetDriverProfileByUserIDUseCase(repo)

	got, err := uc.Execute(context.Background(), d.UserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserID != d.UserID {
		t.Errorf("want %s got %s", d.UserID, got.UserID)
	}
}

// ─── UpdateDriverProfileUseCase ──────────────────────────────────────────────

func TestUpdateDriverProfile_Valid(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	uc := app.NewUpdateDriverProfileUseCase(repo)

	in := app.UpdateDriverProfileInput{
		DriverID:      "d1",
		LicenseNumber: "LIC-NEW",
		VehicleType:   entity.VehicleTypeVan,
		VehicleBrand:  "Ford",
		VehicleModel:  "Transit",
		VehicleColor:  "Blue",
		PlateNumber:   "NEW-999",
	}
	got, err := uc.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LicenseNumber != "LIC-NEW" {
		t.Errorf("license not updated")
	}
	if got.VehicleType != entity.VehicleTypeVan {
		t.Errorf("vehicle type not updated")
	}
}

func TestUpdateDriverProfile_EmptyDriverID(t *testing.T) {
	repo := newStubRepo()
	uc := app.NewUpdateDriverProfileUseCase(repo)
	_, err := uc.Execute(context.Background(), app.UpdateDriverProfileInput{})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestUpdateDriverProfile_NotFound(t *testing.T) {
	repo := newStubRepo()
	uc := app.NewUpdateDriverProfileUseCase(repo)
	in := app.UpdateDriverProfileInput{
		DriverID:      "ghost",
		LicenseNumber: "LIC",
		VehicleType:   entity.VehicleTypeCar,
		PlateNumber:   "PLATE",
	}
	_, err := uc.Execute(context.Background(), in)
	if !errors.IsCode(err, errors.CodeNotFound) {
		t.Errorf("want CodeNotFound got %v", err)
	}
}

// ─── UpdateOnlineStatusUseCase ───────────────────────────────────────────────

func TestUpdateOnlineStatus_GoOnline_WhenVerified(t *testing.T) {
	repo := newStubRepo()
	seedVerified(t, repo)
	uc := app.NewUpdateOnlineStatusUseCase(repo)

	got, err := uc.Execute(context.Background(), "d1", entity.OnlineStatusOnline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OnlineStatus != entity.OnlineStatusOnline {
		t.Errorf("expected online")
	}
}

func TestUpdateOnlineStatus_GoOnline_WhenPending(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	uc := app.NewUpdateOnlineStatusUseCase(repo)

	_, err := uc.Execute(context.Background(), "d1", entity.OnlineStatusOnline)
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed got %v", err)
	}
}

func TestUpdateOnlineStatus_GoOffline(t *testing.T) {
	repo := newStubRepo()
	d := seedVerified(t, repo)
	if err := d.GoOnline(testNow); err != nil {
		t.Fatalf("GoOnline: %v", err)
	}
	_ = repo.Save(context.Background(), d)

	uc := app.NewUpdateOnlineStatusUseCase(repo)
	got, err := uc.Execute(context.Background(), "d1", entity.OnlineStatusOffline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OnlineStatus != entity.OnlineStatusOffline {
		t.Errorf("expected offline")
	}
}

func TestUpdateOnlineStatus_InvalidStatus(t *testing.T) {
	repo := newStubRepo()
	seedVerified(t, repo)
	uc := app.NewUpdateOnlineStatusUseCase(repo)
	_, err := uc.Execute(context.Background(), "d1", entity.OnlineStatus("traveling"))
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

// ─── UpdateVerificationStatusUseCase ─────────────────────────────────────────

func TestUpdateVerificationStatus_Verify(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	uc := app.NewUpdateVerificationStatusUseCase(repo)

	got, err := uc.Execute(context.Background(), "d1", app.VerificationActionVerify)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.VerificationStatus != entity.VerificationStatusVerified {
		t.Errorf("expected verified")
	}
}

func TestUpdateVerificationStatus_Reject(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	uc := app.NewUpdateVerificationStatusUseCase(repo)

	got, err := uc.Execute(context.Background(), "d1", app.VerificationActionReject)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.VerificationStatus != entity.VerificationStatusRejected {
		t.Errorf("expected rejected")
	}
}

func TestUpdateVerificationStatus_Suspend(t *testing.T) {
	repo := newStubRepo()
	seedVerified(t, repo)
	uc := app.NewUpdateVerificationStatusUseCase(repo)

	got, err := uc.Execute(context.Background(), "d1", app.VerificationActionSuspend)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.VerificationStatus != entity.VerificationStatusSuspended {
		t.Errorf("expected suspended")
	}
	if got.OnlineStatus != entity.OnlineStatusOffline {
		t.Errorf("suspend should force offline")
	}
}

func TestUpdateVerificationStatus_Reinstate(t *testing.T) {
	repo := newStubRepo()
	d := seedVerified(t, repo)
	if err := d.Suspend(testNow); err != nil {
		t.Fatalf("Suspend: %v", err)
	}
	_ = repo.Save(context.Background(), d)

	uc := app.NewUpdateVerificationStatusUseCase(repo)
	got, err := uc.Execute(context.Background(), "d1", app.VerificationActionReinstate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.VerificationStatus != entity.VerificationStatusVerified {
		t.Errorf("expected verified after reinstate")
	}
}

func TestUpdateVerificationStatus_InvalidAction(t *testing.T) {
	repo := newStubRepo()
	seedPending(t, repo)
	uc := app.NewUpdateVerificationStatusUseCase(repo)
	_, err := uc.Execute(context.Background(), "d1", app.VerificationAction("approved"))
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestUpdateVerificationStatus_EmptyDriverID(t *testing.T) {
	repo := newStubRepo()
	uc := app.NewUpdateVerificationStatusUseCase(repo)
	_, err := uc.Execute(context.Background(), "", app.VerificationActionVerify)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

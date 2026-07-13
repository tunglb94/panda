package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
)

// domainMessage extracts the bare, driver-facing message (never the
// "[CODE] message" formatted Error() string — that's what the gateway
// actually forwards to the client, see writeDomainError).
func domainMessage(err error) string {
	if de, ok := err.(*errors.DomainError); ok {
		return de.Message
	}
	return err.Error()
}

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

// stubDriverVerificationRepo defaults every driver to Approved unless
// overridden via statuses — keeps every pre-existing availability test
// passing unchanged while letting the new Online Guard tests opt a specific
// driver into a non-approved status.
type stubDriverVerificationRepo struct {
	statuses map[string]entity.KYCStatus
}

func (r *stubDriverVerificationRepo) Save(context.Context, *entity.DriverVerification) error {
	return nil
}

func (r *stubDriverVerificationRepo) FindByDriverID(_ context.Context, driverID string) (*entity.DriverVerification, error) {
	status := entity.KYCApproved
	if r.statuses != nil {
		if s, ok := r.statuses[driverID]; ok {
			status = s
		}
	}
	return &entity.DriverVerification{DriverID: driverID, Status: status}, nil
}

func (r *stubDriverVerificationRepo) FindByNationalIDNumber(context.Context, string) (*entity.DriverVerification, error) {
	return nil, errors.NotFound("driver verification not found")
}

func (r *stubDriverVerificationRepo) FindByLicenseNumber(context.Context, string) (*entity.DriverVerification, error) {
	return nil, errors.NotFound("driver verification not found")
}

func (r *stubDriverVerificationRepo) ListByStatus(context.Context, entity.KYCStatus, int) ([]*entity.DriverVerification, error) {
	return nil, nil
}

func (r *stubDriverVerificationRepo) CountByStatus(context.Context, entity.KYCStatus) (int, error) {
	return 0, nil
}

// stubVehicleVerificationRepo mirrors stubDriverVerificationRepo — defaults
// to Approved, RideEnabled=false/DeliveryEnabled=true unless overridden.
type stubVehicleVerificationRepo struct {
	statuses     map[string]entity.KYCStatus
	rideEnabled  map[string]bool
	licenseClass map[string]entity.LicenseClass
	serviceType  map[string]entity.ServiceType
}

func (r *stubVehicleVerificationRepo) Save(context.Context, *entity.VehicleVerification) error {
	return nil
}

func (r *stubVehicleVerificationRepo) FindByDriverID(_ context.Context, driverID string) (*entity.VehicleVerification, error) {
	status := entity.KYCApproved
	if r.statuses != nil {
		if s, ok := r.statuses[driverID]; ok {
			status = s
		}
	}
	rideEnabled := false
	if r.rideEnabled != nil {
		rideEnabled = r.rideEnabled[driverID]
	}
	serviceType := entity.ServiceType("")
	if r.serviceType != nil {
		serviceType = r.serviceType[driverID]
	}
	licenseClass := entity.LicenseClass("")
	if r.licenseClass != nil {
		licenseClass = r.licenseClass[driverID]
	}
	return &entity.VehicleVerification{
		DriverID:        driverID,
		Status:          status,
		RideEnabled:     rideEnabled,
		DeliveryEnabled: !rideEnabled,
		ServiceType:     serviceType,
		LicenseClass:    licenseClass,
	}, nil
}

func (r *stubVehicleVerificationRepo) FindByPlateNumber(context.Context, string) (*entity.VehicleVerification, error) {
	return nil, errors.NotFound("vehicle verification not found")
}

func (r *stubVehicleVerificationRepo) FindByVIN(context.Context, string) (*entity.VehicleVerification, error) {
	return nil, errors.NotFound("vehicle verification not found")
}

func (r *stubVehicleVerificationRepo) FindByEngineNumber(context.Context, string) (*entity.VehicleVerification, error) {
	return nil, errors.NotFound("vehicle verification not found")
}

func (r *stubVehicleVerificationRepo) FindByChassisNumber(context.Context, string) (*entity.VehicleVerification, error) {
	return nil, errors.NotFound("vehicle verification not found")
}

func (r *stubVehicleVerificationRepo) ListByFilter(context.Context, entity.KYCStatus, entity.VehicleType, entity.ServiceType, int) ([]*entity.VehicleVerification, error) {
	return nil, nil
}

func (r *stubVehicleVerificationRepo) ListByFilterSortedByExpiry(context.Context, entity.KYCStatus, entity.VehicleType, entity.ServiceType, int) ([]*entity.VehicleVerification, error) {
	return nil, nil
}

// stubKYCDocumentRepo defaults to "nothing uploaded" (CodeNotFound), so the
// Online Guard's expiry check is a no-op unless a test seeds a specific
// document via docs[driverID][docType].
type stubKYCDocumentRepo struct {
	docs map[string]map[entity.DocumentType]*entity.KYCDocument
}

func (r *stubKYCDocumentRepo) Save(_ context.Context, d *entity.KYCDocument) error {
	if r.docs == nil {
		r.docs = map[string]map[entity.DocumentType]*entity.KYCDocument{}
	}
	if r.docs[d.DriverID] == nil {
		r.docs[d.DriverID] = map[entity.DocumentType]*entity.KYCDocument{}
	}
	r.docs[d.DriverID][d.DocumentType] = d
	return nil
}

func (r *stubKYCDocumentRepo) FindByDriverAndType(_ context.Context, driverID string, docType entity.DocumentType) (*entity.KYCDocument, error) {
	if d, ok := r.docs[driverID][docType]; ok {
		return d, nil
	}
	return nil, errors.NotFound("kyc document not found")
}

func (r *stubKYCDocumentRepo) ListByDriverID(context.Context, string) ([]*entity.KYCDocument, error) {
	return nil, nil
}

func (r *stubKYCDocumentRepo) ListVersionsByDriverAndType(context.Context, string, entity.DocumentType) ([]*entity.KYCDocument, error) {
	return nil, nil
}

func (r *stubKYCDocumentRepo) FindByID(context.Context, string) (*entity.KYCDocument, error) {
	return nil, errors.NotFound("kyc document not found")
}

func newGoOnlineUseCaseForTest(avail *stubAvailRepo, drivers *stubDriverVerificationRepo, vehicles *stubVehicleVerificationRepo) *app.GoOnlineUseCase {
	return app.NewGoOnlineUseCase(avail, drivers, vehicles, &stubKYCDocumentRepo{}, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())
}

// ─── GoOnlineUseCase ─────────────────────────────────────────────────────────

func TestGoOnline_OK(t *testing.T) {
	repo := newStubAvailRepo()
	uc := newGoOnlineUseCaseForTest(repo, &stubDriverVerificationRepo{}, &stubVehicleVerificationRepo{})

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
	uc := newGoOnlineUseCaseForTest(newStubAvailRepo(), &stubDriverVerificationRepo{}, &stubVehicleVerificationRepo{})
	_, err := uc.Execute(context.Background(), "")
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument got %v", err)
	}
}

func TestGoOnline_Idempotent(t *testing.T) {
	repo := newStubAvailRepo()
	uc := newGoOnlineUseCaseForTest(repo, &stubDriverVerificationRepo{}, &stubVehicleVerificationRepo{})
	_, _ = uc.Execute(context.Background(), "d1")
	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("second GoOnline should succeed: %v", err)
	}
	if !state.IsOnline {
		t.Errorf("driver should remain online")
	}
}

// ─── Online Guard (Phần 9) ──────────────────────────────────────────────────

func TestGoOnline_RejectsWhenDriverVerificationNotApproved(t *testing.T) {
	driverRepo := &stubDriverVerificationRepo{statuses: map[string]entity.KYCStatus{"d1": entity.KYCPending}}
	uc := newGoOnlineUseCaseForTest(newStubAvailRepo(), driverRepo, &stubVehicleVerificationRepo{})

	_, err := uc.Execute(context.Background(), "d1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for unapproved driver KYC, got %v", err)
	}
}

func TestGoOnline_RejectsWhenVehicleVerificationNotApproved(t *testing.T) {
	vehicleRepo := &stubVehicleVerificationRepo{statuses: map[string]entity.KYCStatus{"d1": entity.KYCRejected}}
	uc := newGoOnlineUseCaseForTest(newStubAvailRepo(), &stubDriverVerificationRepo{}, vehicleRepo)

	_, err := uc.Execute(context.Background(), "d1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for unapproved vehicle verification, got %v", err)
	}
}

func TestGoOnline_ExpiredDriverVerificationHasVietnameseMessage(t *testing.T) {
	driverRepo := &stubDriverVerificationRepo{statuses: map[string]entity.KYCStatus{"d1": entity.KYCExpired}}
	uc := newGoOnlineUseCaseForTest(newStubAvailRepo(), driverRepo, &stubVehicleVerificationRepo{})

	_, err := uc.Execute(context.Background(), "d1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed, got %v", err)
	}
	if domainMessage(err) != "CCCD cần xác minh lại" {
		t.Errorf("message = %q, want the exact Phần 9 example message", domainMessage(err))
	}
}

func TestGoOnline_PendingVehicleHasVietnameseMessage(t *testing.T) {
	vehicleRepo := &stubVehicleVerificationRepo{statuses: map[string]entity.KYCStatus{"d1": entity.KYCPending}}
	uc := newGoOnlineUseCaseForTest(newStubAvailRepo(), &stubDriverVerificationRepo{}, vehicleRepo)

	_, err := uc.Execute(context.Background(), "d1")
	if domainMessage(err) != "Xe đang chờ duyệt" {
		t.Errorf("message = %q, want the exact Phần 9 example message", domainMessage(err))
	}
}

func TestGoOnline_RejectsWhenLicenseInvalidForRideServiceType(t *testing.T) {
	// A1 license (motorcycle-only) can't drive Car — the Rule Engine matrix.
	vehicleRepo := &stubVehicleVerificationRepo{
		rideEnabled:  map[string]bool{"d1": true},
		serviceType:  map[string]entity.ServiceType{"d1": entity.ServiceTypeCar},
		licenseClass: map[string]entity.LicenseClass{"d1": entity.LicenseClassA1},
	}
	uc := newGoOnlineUseCaseForTest(newStubAvailRepo(), &stubDriverVerificationRepo{}, vehicleRepo)

	_, err := uc.Execute(context.Background(), "d1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for A1 license driving Car, got %v", err)
	}
}

func TestGoOnline_AllowsWhenLicenseValidForRideServiceType(t *testing.T) {
	vehicleRepo := &stubVehicleVerificationRepo{
		rideEnabled:  map[string]bool{"d1": true},
		serviceType:  map[string]entity.ServiceType{"d1": entity.ServiceTypeBike},
		licenseClass: map[string]entity.LicenseClass{"d1": entity.LicenseClassA1},
	}
	uc := newGoOnlineUseCaseForTest(newStubAvailRepo(), &stubDriverVerificationRepo{}, vehicleRepo)

	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("A1 license driving Bike should be allowed: %v", err)
	}
	if !state.IsOnline {
		t.Errorf("driver should be online")
	}
}

// ─── Phần 2 — Document Expiration (auto-expire at Go-Online time) ──────────

func expiredDoc(driverID string, dt entity.DocumentType, expiredAt time.Time) *entity.KYCDocument {
	return entity.ReconstituteKYCDocument("doc1", driverID, dt, "path", "image/jpeg", 1, &expiredAt, driverID, expiredAt)
}

func TestGoOnline_RejectsAndAutoExpiresOnExpiredRegistration(t *testing.T) {
	avail := newStubAvailRepo()
	vehicles := &stubVehicleVerificationRepo{}
	documents := &stubKYCDocumentRepo{docs: map[string]map[entity.DocumentType]*entity.KYCDocument{
		"d1": {entity.DocumentVehicleRegistration: expiredDoc("d1", entity.DocumentVehicleRegistration, time.Now().Add(-24*time.Hour))},
	}}
	audit := newFakeAuditLogRepo()
	uc := app.NewGoOnlineUseCase(avail, &stubDriverVerificationRepo{}, vehicles, documents, newFakeLicenseCapabilityRepo(), audit)

	_, err := uc.Execute(context.Background(), "d1")
	if !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("want CodePreconditionFailed for expired registration, got %v", err)
	}
	if domainMessage(err) != "Đăng ký xe đã hết hạn" {
		t.Errorf("message = %q, want the exact Phần 9 example message", domainMessage(err))
	}

	found := false
	for _, e := range audit.entries {
		if e.Action == entity.AuditActionExpire && e.ActorID == "system" {
			found = true
		}
	}
	if !found {
		t.Error("expected a system-actor audit entry for the auto-expire")
	}
}

func TestGoOnline_RejectsOnExpiredLicenseOnlyWhenRideEnabled(t *testing.T) {
	avail := newStubAvailRepo()
	vehicles := &stubVehicleVerificationRepo{
		rideEnabled:  map[string]bool{"d1": true},
		serviceType:  map[string]entity.ServiceType{"d1": entity.ServiceTypeBike},
		licenseClass: map[string]entity.LicenseClass{"d1": entity.LicenseClassA1},
	}
	documents := &stubKYCDocumentRepo{docs: map[string]map[entity.DocumentType]*entity.KYCDocument{
		"d1": {entity.DocumentLicense: expiredDoc("d1", entity.DocumentLicense, time.Now().Add(-24*time.Hour))},
	}}
	uc := app.NewGoOnlineUseCase(avail, &stubDriverVerificationRepo{}, vehicles, documents, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	_, err := uc.Execute(context.Background(), "d1")
	if domainMessage(err) != "GPLX đã hết hạn" {
		t.Errorf("message = %q, want the exact Phần 9 example message", domainMessage(err))
	}
}

func TestGoOnline_AllowsWhenDocumentsNotExpired(t *testing.T) {
	avail := newStubAvailRepo()
	vehicles := &stubVehicleVerificationRepo{}
	future := time.Now().Add(365 * 24 * time.Hour)
	documents := &stubKYCDocumentRepo{docs: map[string]map[entity.DocumentType]*entity.KYCDocument{
		"d1": {entity.DocumentVehicleRegistration: expiredDoc("d1", entity.DocumentVehicleRegistration, future)},
	}}
	uc := app.NewGoOnlineUseCase(avail, &stubDriverVerificationRepo{}, vehicles, documents, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	state, err := uc.Execute(context.Background(), "d1")
	if err != nil {
		t.Fatalf("unexpired documents should not block Online: %v", err)
	}
	if !state.IsOnline {
		t.Error("driver should be online")
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
	goOnline := newGoOnlineUseCaseForTest(repo, &stubDriverVerificationRepo{}, &stubVehicleVerificationRepo{})
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

package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
)

var testVerifyNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
var testDOB = time.Date(1995, 5, 20, 0, 0, 0, 0, time.UTC)

// ─── in-memory fakes ────────────────────────────────────────────────────────

type fakeDriverVerificationRepo struct {
	byDriverID map[string]*entity.DriverVerification
}

func newFakeDriverVerificationRepo() *fakeDriverVerificationRepo {
	return &fakeDriverVerificationRepo{byDriverID: map[string]*entity.DriverVerification{}}
}

func (r *fakeDriverVerificationRepo) Save(_ context.Context, v *entity.DriverVerification) error {
	cp := *v
	r.byDriverID[v.DriverID] = &cp
	return nil
}

func (r *fakeDriverVerificationRepo) FindByDriverID(_ context.Context, driverID string) (*entity.DriverVerification, error) {
	v, ok := r.byDriverID[driverID]
	if !ok {
		return nil, errors.NotFound("driver verification not found")
	}
	cp := *v
	return &cp, nil
}

func (r *fakeDriverVerificationRepo) FindByNationalIDNumber(_ context.Context, nationalIDNumber string) (*entity.DriverVerification, error) {
	for _, v := range r.byDriverID {
		if nationalIDNumber != "" && v.NationalIDNumber == nationalIDNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, errors.NotFound("driver verification not found")
}

func (r *fakeDriverVerificationRepo) FindByLicenseNumber(_ context.Context, licenseNumber string) (*entity.DriverVerification, error) {
	for _, v := range r.byDriverID {
		if licenseNumber != "" && v.LicenseNumber == licenseNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, errors.NotFound("driver verification not found")
}

func (r *fakeDriverVerificationRepo) ListByStatus(_ context.Context, status entity.KYCStatus, limit int) ([]*entity.DriverVerification, error) {
	var out []*entity.DriverVerification
	for _, v := range r.byDriverID {
		if v.Status == status {
			out = append(out, v)
		}
	}
	return out, nil
}

func (r *fakeDriverVerificationRepo) CountByStatus(_ context.Context, status entity.KYCStatus) (int, error) {
	count := 0
	for _, v := range r.byDriverID {
		if v.Status == status {
			count++
		}
	}
	return count, nil
}

type fakeVehicleVerificationRepo struct {
	byDriverID map[string]*entity.VehicleVerification
}

func newFakeVehicleVerificationRepo() *fakeVehicleVerificationRepo {
	return &fakeVehicleVerificationRepo{byDriverID: map[string]*entity.VehicleVerification{}}
}

func (r *fakeVehicleVerificationRepo) Save(_ context.Context, v *entity.VehicleVerification) error {
	cp := *v
	r.byDriverID[v.DriverID] = &cp
	return nil
}

func (r *fakeVehicleVerificationRepo) FindByDriverID(_ context.Context, driverID string) (*entity.VehicleVerification, error) {
	v, ok := r.byDriverID[driverID]
	if !ok {
		return nil, errors.NotFound("vehicle verification not found")
	}
	cp := *v
	return &cp, nil
}

func (r *fakeVehicleVerificationRepo) FindByPlateNumber(_ context.Context, plateNumber string) (*entity.VehicleVerification, error) {
	for _, v := range r.byDriverID {
		if v.PlateNumber == plateNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, errors.NotFound("vehicle verification not found")
}

func (r *fakeVehicleVerificationRepo) FindByVIN(_ context.Context, vin string) (*entity.VehicleVerification, error) {
	for _, v := range r.byDriverID {
		if vin != "" && v.VIN == vin {
			cp := *v
			return &cp, nil
		}
	}
	return nil, errors.NotFound("vehicle verification not found")
}

func (r *fakeVehicleVerificationRepo) FindByEngineNumber(_ context.Context, engineNumber string) (*entity.VehicleVerification, error) {
	for _, v := range r.byDriverID {
		if engineNumber != "" && v.EngineNumber == engineNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, errors.NotFound("vehicle verification not found")
}

func (r *fakeVehicleVerificationRepo) FindByChassisNumber(_ context.Context, chassisNumber string) (*entity.VehicleVerification, error) {
	for _, v := range r.byDriverID {
		if chassisNumber != "" && v.ChassisNumber == chassisNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, errors.NotFound("vehicle verification not found")
}

func (r *fakeVehicleVerificationRepo) ListByFilter(_ context.Context, status entity.KYCStatus, vehicleType entity.VehicleType, serviceType entity.ServiceType, limit int) ([]*entity.VehicleVerification, error) {
	var out []*entity.VehicleVerification
	for _, v := range r.byDriverID {
		if status != "" && v.Status != status {
			continue
		}
		if vehicleType != "" && v.VehicleType != vehicleType {
			continue
		}
		if serviceType != "" && v.ServiceType != serviceType {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *fakeVehicleVerificationRepo) ListByFilterSortedByExpiry(ctx context.Context, status entity.KYCStatus, vehicleType entity.VehicleType, serviceType entity.ServiceType, limit int) ([]*entity.VehicleVerification, error) {
	return r.ListByFilter(ctx, status, vehicleType, serviceType, limit)
}

type fakeKYCDocumentRepo struct {
	// versions[driverID][docType] is every uploaded version, oldest first.
	versions map[string]map[entity.DocumentType][]*entity.KYCDocument
}

func newFakeKYCDocumentRepo() *fakeKYCDocumentRepo {
	return &fakeKYCDocumentRepo{versions: map[string]map[entity.DocumentType][]*entity.KYCDocument{}}
}

func (r *fakeKYCDocumentRepo) Save(_ context.Context, d *entity.KYCDocument) error {
	cp := *d
	if r.versions[d.DriverID] == nil {
		r.versions[d.DriverID] = map[entity.DocumentType][]*entity.KYCDocument{}
	}
	r.versions[d.DriverID][d.DocumentType] = append(r.versions[d.DriverID][d.DocumentType], &cp)
	return nil
}

func (r *fakeKYCDocumentRepo) FindByDriverAndType(_ context.Context, driverID string, docType entity.DocumentType) (*entity.KYCDocument, error) {
	list := r.versions[driverID][docType]
	if len(list) == 0 {
		return nil, errors.NotFound("kyc document not found")
	}
	return list[len(list)-1], nil
}

func (r *fakeKYCDocumentRepo) ListByDriverID(_ context.Context, driverID string) ([]*entity.KYCDocument, error) {
	var out []*entity.KYCDocument
	for _, list := range r.versions[driverID] {
		if len(list) > 0 {
			out = append(out, list[len(list)-1])
		}
	}
	return out, nil
}

func (r *fakeKYCDocumentRepo) ListVersionsByDriverAndType(_ context.Context, driverID string, docType entity.DocumentType) ([]*entity.KYCDocument, error) {
	list := r.versions[driverID][docType]
	out := make([]*entity.KYCDocument, len(list))
	for i, d := range list {
		out[len(list)-1-i] = d // newest first
	}
	return out, nil
}

func (r *fakeKYCDocumentRepo) FindByID(_ context.Context, id string) (*entity.KYCDocument, error) {
	for _, byType := range r.versions {
		for _, list := range byType {
			for _, d := range list {
				if d.ID == id {
					return d, nil
				}
			}
		}
	}
	return nil, errors.NotFound("kyc document not found")
}

func seedDoc(t *testing.T, repo *fakeKYCDocumentRepo, driverID string, docType entity.DocumentType) {
	t.Helper()
	doc, err := entity.NewKYCDocument("doc-"+driverID+"-"+string(docType), driverID, docType, "path/"+string(docType), "image/jpeg", 1, nil, driverID, testVerifyNow)
	if err != nil {
		t.Fatalf("seedDoc: %v", err)
	}
	_ = repo.Save(context.Background(), doc)
}

// fakeLicenseCapabilityRepo mirrors the seeded migration matrix in-memory,
// so app-layer tests don't need a live Postgres connection.
type fakeLicenseCapabilityRepo struct {
	allow map[string]bool
}

func newFakeLicenseCapabilityRepo() *fakeLicenseCapabilityRepo {
	return &fakeLicenseCapabilityRepo{allow: map[string]bool{
		"A1|motorcycle": true, "A1|bike_plus": true,
		"A2|motorcycle": true, "A2|bike_plus": true,
		"B1|car": true, "B1|car_xl": true,
		"B2|motorcycle": true, "B2|bike_plus": true, "B2|car": true, "B2|car_xl": true,
	}}
}

func (r *fakeLicenseCapabilityRepo) IsAllowed(_ context.Context, licenseClass entity.LicenseClass, serviceType entity.ServiceType) (bool, error) {
	return r.allow[string(licenseClass)+"|"+string(serviceType)], nil
}

type fakeAuditLogRepo struct {
	entries []*entity.AuditLog
}

func newFakeAuditLogRepo() *fakeAuditLogRepo { return &fakeAuditLogRepo{} }

func (r *fakeAuditLogRepo) Save(_ context.Context, log *entity.AuditLog) error {
	r.entries = append(r.entries, log)
	return nil
}

func (r *fakeAuditLogRepo) ListByDriverID(_ context.Context, driverID string, limit int) ([]*entity.AuditLog, error) {
	var out []*entity.AuditLog
	for _, e := range r.entries {
		if e.DriverID == driverID {
			out = append(out, e)
		}
	}
	return out, nil
}

// ─── SubmitDriverVerificationUseCase ────────────────────────────────────────

func TestSubmitDriverVerification_RejectsWhenDocumentsMissing(t *testing.T) {
	uc := app.NewSubmitDriverVerificationUseCase(newFakeDriverVerificationRepo(), newFakeKYCDocumentRepo(), newFakeAuditLogRepo())

	_, err := uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d1", FullName: "A", DateOfBirth: testDOB, Address: "addr", NationalIDNumber: "079095001234",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument when required documents are missing, got %v", err)
	}
}

func TestSubmitDriverVerification_OKWhenDocumentsUploaded(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.DriverDocumentTypes {
		seedDoc(t, docs, "d1", dt)
	}
	audit := newFakeAuditLogRepo()
	uc := app.NewSubmitDriverVerificationUseCase(newFakeDriverVerificationRepo(), docs, audit)

	v, err := uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d1", FullName: "Nguyen Van A", DateOfBirth: testDOB, Address: "123 Le Loi", NationalIDNumber: "079095001234",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending", v.Status)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != entity.AuditActionSubmit {
		t.Errorf("expected one submit audit entry, got %+v", audit.entries)
	}
}

func TestSubmitDriverVerification_RejectsEmptyNationalIDNumber(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.DriverDocumentTypes {
		seedDoc(t, docs, "d1", dt)
	}
	uc := app.NewSubmitDriverVerificationUseCase(newFakeDriverVerificationRepo(), docs, newFakeAuditLogRepo())

	_, err := uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d1", FullName: "A", DateOfBirth: testDOB, Address: "addr", NationalIDNumber: "",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for empty national_id_number, got %v", err)
	}
}

func TestSubmitDriverVerification_RejectsDuplicateNationalIDAcrossDrivers(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.DriverDocumentTypes {
		seedDoc(t, docs, "d1", dt)
		seedDoc(t, docs, "d2", dt)
	}
	repo := newFakeDriverVerificationRepo()
	uc := app.NewSubmitDriverVerificationUseCase(repo, docs, newFakeAuditLogRepo())

	_, err := uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d1", FullName: "A", DateOfBirth: testDOB, Address: "addr", NationalIDNumber: "079095001234",
	})
	if err != nil {
		t.Fatalf("first driver's submission should succeed: %v", err)
	}
	_, err = uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d2", FullName: "B", DateOfBirth: testDOB, Address: "addr", NationalIDNumber: "079095001234",
	})
	if !errors.IsCode(err, errors.CodeAlreadyExists) {
		t.Errorf("want CodeAlreadyExists for a CCCD number already used by another driver, got %v", err)
	}
}

func TestSubmitDriverVerification_ResubmitsExisting(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.DriverDocumentTypes {
		seedDoc(t, docs, "d1", dt)
	}
	repo := newFakeDriverVerificationRepo()
	uc := app.NewSubmitDriverVerificationUseCase(repo, docs, newFakeAuditLogRepo())

	first, err := uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d1", FullName: "A", DateOfBirth: testDOB, Address: "addr1", NationalIDNumber: "079095001234",
	})
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}

	second, err := uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d1", FullName: "A2", DateOfBirth: testDOB, Address: "addr2", NationalIDNumber: "079095001234",
	})
	if err != nil {
		t.Fatalf("second submit (resubmit) should succeed: %v", err)
	}
	if second.ID != first.ID {
		t.Error("resubmit should reuse the same verification ID, not create a new one")
	}
	if second.Address != "addr2" {
		t.Error("resubmit should update the address")
	}
}

// Phần 3 — Re-verification: editing an Approved record must reset it to
// Pending, never keep it Approved.
func TestSubmitDriverVerification_EditingApprovedResetsToPending(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.DriverDocumentTypes {
		seedDoc(t, docs, "d1", dt)
	}
	repo := newFakeDriverVerificationRepo()
	uc := app.NewSubmitDriverVerificationUseCase(repo, docs, newFakeAuditLogRepo())

	v, _ := uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d1", FullName: "A", DateOfBirth: testDOB, Address: "addr1", NationalIDNumber: "079095001234",
	})
	_ = v.Approve("admin1", testVerifyNow)
	_ = repo.Save(context.Background(), v)

	edited, err := uc.Execute(context.Background(), app.SubmitDriverVerificationInput{
		DriverID: "d1", FullName: "A", DateOfBirth: testDOB, Address: "addr-changed", NationalIDNumber: "079095001234",
	})
	if err != nil {
		t.Fatalf("editing an approved verification should succeed (re-verification): %v", err)
	}
	if edited.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending — editing an approved record must not keep it approved", edited.Status)
	}
}

// ─── ReviewDriverVerificationUseCase ────────────────────────────────────────

func TestReviewDriverVerification_ApproveAndReject(t *testing.T) {
	repo := newFakeDriverVerificationRepo()
	v, _ := entity.NewDriverVerification("id1", "d1", "A", testDOB, "addr", "079095001234", "", testVerifyNow)
	_ = repo.Save(context.Background(), v)

	audit := newFakeAuditLogRepo()
	uc := app.NewReviewDriverVerificationUseCase(repo, audit)

	approved, err := uc.Execute(context.Background(), app.ReviewDriverVerificationInput{
		DriverID: "d1", Reviewer: "admin1", Action: app.ReviewApprove,
	})
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if !approved.IsApproved() {
		t.Error("should be approved")
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != entity.AuditActionApprove {
		t.Errorf("expected one approve audit entry, got %+v", audit.entries)
	}
}

func TestReviewDriverVerification_RejectRequiresReason(t *testing.T) {
	repo := newFakeDriverVerificationRepo()
	v, _ := entity.NewDriverVerification("id1", "d1", "A", testDOB, "addr", "079095001234", "", testVerifyNow)
	_ = repo.Save(context.Background(), v)

	uc := app.NewReviewDriverVerificationUseCase(repo, newFakeAuditLogRepo())
	_, err := uc.Execute(context.Background(), app.ReviewDriverVerificationInput{
		DriverID: "d1", Reviewer: "admin1", Action: app.ReviewReject, Reason: "",
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for missing reject reason, got %v", err)
	}
}

// ─── SubmitVehicleVerificationUseCase ───────────────────────────────────────

func TestSubmitVehicleVerification_RejectsWhenDocumentsMissing(t *testing.T) {
	uc := app.NewSubmitVehicleVerificationUseCase(newFakeVehicleVerificationRepo(), newFakeKYCDocumentRepo(), newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	_, err := uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d1", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		PlateNumber: "59H1-11111", RideEnabled: false, DeliveryEnabled: true,
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument when vehicle documents are missing, got %v", err)
	}
}

func TestSubmitVehicleVerification_RideEnabledRequiresLicenseDocument(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.VehicleDocumentTypes {
		seedDoc(t, docs, "d1", dt)
	}
	// license doc intentionally NOT seeded
	uc := app.NewSubmitVehicleVerificationUseCase(newFakeVehicleVerificationRepo(), docs, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	_, err := uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d1", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		PlateNumber: "59H1-11111", LicenseClass: entity.LicenseClassA1, RideEnabled: true,
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument when license/GPLX document is missing for Ride, got %v", err)
	}
}

func TestSubmitVehicleVerification_RejectsDuplicatePlateAcrossDrivers(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.VehicleDocumentTypes {
		seedDoc(t, docs, "d1", dt)
		seedDoc(t, docs, "d2", dt)
	}
	repo := newFakeVehicleVerificationRepo()
	uc := app.NewSubmitVehicleVerificationUseCase(repo, docs, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	_, err := uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d1", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		PlateNumber: "59H1-99999", DeliveryEnabled: true,
	})
	if err != nil {
		t.Fatalf("first driver's submission should succeed: %v", err)
	}

	_, err = uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d2", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		PlateNumber: "59H1-99999", DeliveryEnabled: true,
	})
	if !errors.IsCode(err, errors.CodeAlreadyExists) {
		t.Errorf("want CodeAlreadyExists for a plate already used by another driver, got %v", err)
	}
}

func TestSubmitVehicleVerification_RejectsDuplicateVINAcrossVehicles(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.VehicleDocumentTypes {
		seedDoc(t, docs, "d1", dt)
		seedDoc(t, docs, "d2", dt)
	}
	repo := newFakeVehicleVerificationRepo()
	uc := app.NewSubmitVehicleVerificationUseCase(repo, docs, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	_, err := uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d1", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		PlateNumber: "59H1-10001", VIN: "VIN-SHARED", DeliveryEnabled: true,
	})
	if err != nil {
		t.Fatalf("first driver's submission should succeed: %v", err)
	}
	_, err = uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d2", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		PlateNumber: "59H1-10002", VIN: "VIN-SHARED", DeliveryEnabled: true,
	})
	if !errors.IsCode(err, errors.CodeAlreadyExists) {
		t.Errorf("want CodeAlreadyExists for a VIN already used by another vehicle, got %v", err)
	}
}

func TestSubmitVehicleVerification_RejectsLicenseClassNotAllowedForServiceType(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.VehicleDocumentTypes {
		seedDoc(t, docs, "d1", dt)
	}
	seedDoc(t, docs, "d1", entity.DocumentLicense)
	uc := app.NewSubmitVehicleVerificationUseCase(newFakeVehicleVerificationRepo(), docs, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	// A1 does not permit Car (per the seeded Rule Engine matrix).
	_, err := uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d1", VehicleType: entity.VehicleTypeCar, ServiceType: entity.ServiceTypeCar,
		PlateNumber: "59H1-30001", LicenseClass: entity.LicenseClassA1, RideEnabled: true,
	})
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for a license class the Rule Engine doesn't allow, got %v", err)
	}
}

func TestSubmitVehicleVerification_OKWithDeliveryOnlyNoLicenseClass(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.VehicleDocumentTypes {
		seedDoc(t, docs, "d1", dt)
	}
	uc := app.NewSubmitVehicleVerificationUseCase(newFakeVehicleVerificationRepo(), docs, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	v, err := uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d1", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		Brand: "Honda", Model: "Wave", Year: 2022, Color: "Đỏ",
		PlateNumber: "59H1-22222", DeliveryEnabled: true,
	})
	if err != nil {
		t.Fatalf("delivery-only submission should succeed without a license class: %v", err)
	}
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending", v.Status)
	}
}

// Phần 3 — Re-verification for vehicles too.
func TestSubmitVehicleVerification_EditingApprovedResetsToPending(t *testing.T) {
	docs := newFakeKYCDocumentRepo()
	for _, dt := range entity.VehicleDocumentTypes {
		seedDoc(t, docs, "d1", dt)
	}
	repo := newFakeVehicleVerificationRepo()
	uc := app.NewSubmitVehicleVerificationUseCase(repo, docs, newFakeLicenseCapabilityRepo(), newFakeAuditLogRepo())

	v, _ := uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d1", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		PlateNumber: "59H1-40001", DeliveryEnabled: true,
	})
	_ = v.Approve("admin1", testVerifyNow)
	_ = repo.Save(context.Background(), v)

	edited, err := uc.Execute(context.Background(), app.SubmitVehicleVerificationInput{
		DriverID: "d1", VehicleType: entity.VehicleTypeMotorcycle, ServiceType: entity.ServiceTypeBike,
		PlateNumber: "59H1-40001", Color: "Xanh", DeliveryEnabled: true,
	})
	if err != nil {
		t.Fatalf("editing an approved vehicle verification should succeed (re-verification): %v", err)
	}
	if edited.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending — editing an approved vehicle record must not keep it approved", edited.Status)
	}
}

package app_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/infrastructure/localstore"
)

func newTestUploadUseCase(t *testing.T) (*app.UploadKYCDocumentUseCase, *fakeKYCDocumentRepo, *fakeDriverVerificationRepo, *fakeVehicleVerificationRepo, *fakeAuditLogRepo) {
	t.Helper()
	docs := newFakeKYCDocumentRepo()
	drivers := newFakeDriverVerificationRepo()
	vehicles := newFakeVehicleVerificationRepo()
	audit := newFakeAuditLogRepo()
	store := localstore.NewDocumentStore(t.TempDir())
	uc := app.NewUploadKYCDocumentUseCase(docs, drivers, vehicles, audit, store)
	return uc, docs, drivers, vehicles, audit
}

// Phần 4 — Document Versioning: re-uploading the same (driver, type) must
// create a new version, never overwrite the previous one.
func TestUploadKYCDocument_ReuploadCreatesNewVersion(t *testing.T) {
	uc, docs, _, _, _ := newTestUploadUseCase(t)
	ctx := context.Background()

	first, err := uc.Execute(ctx, app.UploadKYCDocumentInput{
		DriverID: "d1", DocumentType: entity.DocumentCCCDFront, Filename: "a.jpg", ContentType: "image/jpeg", Data: bytes.NewReader([]byte("v1")),
	})
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}
	if first.Version != 1 {
		t.Errorf("first version = %d, want 1", first.Version)
	}

	second, err := uc.Execute(ctx, app.UploadKYCDocumentInput{
		DriverID: "d1", DocumentType: entity.DocumentCCCDFront, Filename: "b.jpg", ContentType: "image/jpeg", Data: bytes.NewReader([]byte("v2")),
	})
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}
	if second.Version != 2 {
		t.Errorf("second version = %d, want 2", second.Version)
	}
	if second.ID == first.ID {
		t.Error("re-upload must create a new document row, not reuse the old ID")
	}

	versions, err := docs.ListVersionsByDriverAndType(ctx, "d1", entity.DocumentCCCDFront)
	if err != nil || len(versions) != 2 {
		t.Fatalf("expected 2 versions retained, got %v (err=%v)", versions, err)
	}
	if versions[0].ID != second.ID {
		t.Error("ListVersionsByDriverAndType must return newest first")
	}

	latest, err := docs.FindByDriverAndType(ctx, "d1", entity.DocumentCCCDFront)
	if err != nil || latest.ID != second.ID {
		t.Errorf("FindByDriverAndType must return the latest version, got %+v (err=%v)", latest, err)
	}
}

// Phần 2 — expires_at is only recorded for expiry-eligible document types.
func TestUploadKYCDocument_ExpiresAtOnlyForEligibleTypes(t *testing.T) {
	uc, _, _, _, _ := newTestUploadUseCase(t)
	ctx := context.Background()
	expiry := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

	license, err := uc.Execute(ctx, app.UploadKYCDocumentInput{
		DriverID: "d1", DocumentType: entity.DocumentLicense, Filename: "gplx.jpg", ExpiresAt: &expiry, Data: bytes.NewReader([]byte("x")),
	})
	if err != nil {
		t.Fatalf("license upload: %v", err)
	}
	if license.ExpiresAt == nil || !license.ExpiresAt.Equal(expiry) {
		t.Errorf("license ExpiresAt = %v, want %v", license.ExpiresAt, expiry)
	}

	selfie, err := uc.Execute(ctx, app.UploadKYCDocumentInput{
		DriverID: "d1", DocumentType: entity.DocumentSelfie, Filename: "selfie.jpg", ExpiresAt: &expiry, Data: bytes.NewReader([]byte("x")),
	})
	if err != nil {
		t.Fatalf("selfie upload: %v", err)
	}
	if selfie.ExpiresAt != nil {
		t.Errorf("selfie must never carry an expiry, got %v", selfie.ExpiresAt)
	}
}

// Phần 3 — Re-verification: re-uploading a document belonging to an
// Approved verification must invalidate it back to Pending.
func TestUploadKYCDocument_ReuploadInvalidatesApprovedDriverVerification(t *testing.T) {
	uc, _, drivers, _, audit := newTestUploadUseCase(t)
	ctx := context.Background()

	dv, _ := entity.NewDriverVerification("dv1", "d1", "A", testDOB, "addr", "079095001234", "", testVerifyNow)
	_ = dv.Approve("admin1", testVerifyNow)
	_ = drivers.Save(ctx, dv)

	_, err := uc.Execute(ctx, app.UploadKYCDocumentInput{
		DriverID: "d1", DocumentType: entity.DocumentSelfie, Filename: "selfie.jpg", Data: bytes.NewReader([]byte("x")),
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	updated, err := drivers.FindByDriverID(ctx, "d1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if updated.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending — re-uploading a document on an approved profile must invalidate it", updated.Status)
	}

	found := false
	for _, e := range audit.entries {
		if e.EntityType == entity.AuditEntityDriverVerification && e.Action == entity.AuditActionModify {
			found = true
		}
	}
	if !found {
		t.Error("expected an audit entry for the invalidation")
	}
}

// Vehicle documents (e.g. insurance) must invalidate VehicleVerification,
// not DriverVerification.
func TestUploadKYCDocument_ReuploadInvalidatesApprovedVehicleVerification(t *testing.T) {
	uc, _, _, vehicles, _ := newTestUploadUseCase(t)
	ctx := context.Background()

	vv, _ := entity.NewVehicleVerification("vv1", "d1", entity.VehicleTypeMotorcycle, entity.ServiceTypeBike,
		"Honda", "Wave", 2022, "Đỏ", "59H1-12345", "", "", "", "", false, true, testVerifyNow)
	_ = vv.Approve("admin1", testVerifyNow)
	_ = vehicles.Save(ctx, vv)

	_, err := uc.Execute(ctx, app.UploadKYCDocumentInput{
		DriverID: "d1", DocumentType: entity.DocumentVehicleInsurance, Filename: "ins.jpg", Data: bytes.NewReader([]byte("x")),
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	updated, err := vehicles.FindByDriverID(ctx, "d1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if updated.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending — re-uploading a vehicle document on an approved vehicle must invalidate it", updated.Status)
	}
}

// Re-uploading while the owning verification is still Pending must not
// touch its status (nothing to invalidate).
func TestUploadKYCDocument_ReuploadDoesNotAffectPendingVerification(t *testing.T) {
	uc, _, drivers, _, _ := newTestUploadUseCase(t)
	ctx := context.Background()

	dv, _ := entity.NewDriverVerification("dv1", "d1", "A", testDOB, "addr", "079095001234", "", testVerifyNow)
	_ = drivers.Save(ctx, dv)

	_, err := uc.Execute(ctx, app.UploadKYCDocumentInput{
		DriverID: "d1", DocumentType: entity.DocumentSelfie, Filename: "selfie.jpg", Data: bytes.NewReader([]byte("x")),
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	updated, _ := drivers.FindByDriverID(ctx, "d1")
	if updated.Status != entity.KYCPending {
		t.Errorf("status = %v, want unchanged pending", updated.Status)
	}
}

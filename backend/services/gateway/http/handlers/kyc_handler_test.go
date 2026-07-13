package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	driverapp "github.com/fairride/driver/app"
	driverentity "github.com/fairride/driver/domain/entity"
	driverlocalstore "github.com/fairride/driver/infrastructure/localstore"
	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/identity/infrastructure/jwt"
	domainerrors "github.com/fairride/shared/errors"
)

// ─── in-memory fakes (mirrors driver/app's own test fakes) ─────────────────

type fakeDVRepo struct {
	byDriverID map[string]*driverentity.DriverVerification
}

func newFakeDVRepo() *fakeDVRepo {
	return &fakeDVRepo{byDriverID: map[string]*driverentity.DriverVerification{}}
}

func (r *fakeDVRepo) Save(_ context.Context, v *driverentity.DriverVerification) error {
	cp := *v
	r.byDriverID[v.DriverID] = &cp
	return nil
}

func (r *fakeDVRepo) FindByDriverID(_ context.Context, driverID string) (*driverentity.DriverVerification, error) {
	v, ok := r.byDriverID[driverID]
	if !ok {
		return nil, domainerrors.NotFound("driver verification not found")
	}
	cp := *v
	return &cp, nil
}

func (r *fakeDVRepo) FindByNationalIDNumber(_ context.Context, nationalIDNumber string) (*driverentity.DriverVerification, error) {
	for _, v := range r.byDriverID {
		if nationalIDNumber != "" && v.NationalIDNumber == nationalIDNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, domainerrors.NotFound("driver verification not found")
}

func (r *fakeDVRepo) FindByLicenseNumber(_ context.Context, licenseNumber string) (*driverentity.DriverVerification, error) {
	for _, v := range r.byDriverID {
		if licenseNumber != "" && v.LicenseNumber == licenseNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, domainerrors.NotFound("driver verification not found")
}

func (r *fakeDVRepo) ListByStatus(_ context.Context, status driverentity.KYCStatus, _ int) ([]*driverentity.DriverVerification, error) {
	var out []*driverentity.DriverVerification
	for _, v := range r.byDriverID {
		if v.Status == status {
			out = append(out, v)
		}
	}
	return out, nil
}

func (r *fakeDVRepo) CountByStatus(_ context.Context, status driverentity.KYCStatus) (int, error) {
	count := 0
	for _, v := range r.byDriverID {
		if v.Status == status {
			count++
		}
	}
	return count, nil
}

type fakeVVRepo struct {
	byDriverID map[string]*driverentity.VehicleVerification
}

func newFakeVVRepo() *fakeVVRepo {
	return &fakeVVRepo{byDriverID: map[string]*driverentity.VehicleVerification{}}
}

func (r *fakeVVRepo) Save(_ context.Context, v *driverentity.VehicleVerification) error {
	cp := *v
	r.byDriverID[v.DriverID] = &cp
	return nil
}

func (r *fakeVVRepo) FindByDriverID(_ context.Context, driverID string) (*driverentity.VehicleVerification, error) {
	v, ok := r.byDriverID[driverID]
	if !ok {
		return nil, domainerrors.NotFound("vehicle verification not found")
	}
	cp := *v
	return &cp, nil
}

func (r *fakeVVRepo) FindByPlateNumber(_ context.Context, plate string) (*driverentity.VehicleVerification, error) {
	for _, v := range r.byDriverID {
		if v.PlateNumber == plate {
			cp := *v
			return &cp, nil
		}
	}
	return nil, domainerrors.NotFound("vehicle verification not found")
}

func (r *fakeVVRepo) FindByVIN(_ context.Context, vin string) (*driverentity.VehicleVerification, error) {
	for _, v := range r.byDriverID {
		if vin != "" && v.VIN == vin {
			cp := *v
			return &cp, nil
		}
	}
	return nil, domainerrors.NotFound("vehicle verification not found")
}

func (r *fakeVVRepo) FindByEngineNumber(_ context.Context, engineNumber string) (*driverentity.VehicleVerification, error) {
	for _, v := range r.byDriverID {
		if engineNumber != "" && v.EngineNumber == engineNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, domainerrors.NotFound("vehicle verification not found")
}

func (r *fakeVVRepo) FindByChassisNumber(_ context.Context, chassisNumber string) (*driverentity.VehicleVerification, error) {
	for _, v := range r.byDriverID {
		if chassisNumber != "" && v.ChassisNumber == chassisNumber {
			cp := *v
			return &cp, nil
		}
	}
	return nil, domainerrors.NotFound("vehicle verification not found")
}

func (r *fakeVVRepo) ListByFilter(_ context.Context, status driverentity.KYCStatus, vt driverentity.VehicleType, st driverentity.ServiceType, _ int) ([]*driverentity.VehicleVerification, error) {
	var out []*driverentity.VehicleVerification
	for _, v := range r.byDriverID {
		if status != "" && v.Status != status {
			continue
		}
		if vt != "" && v.VehicleType != vt {
			continue
		}
		if st != "" && v.ServiceType != st {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *fakeVVRepo) ListByFilterSortedByExpiry(ctx context.Context, status driverentity.KYCStatus, vt driverentity.VehicleType, st driverentity.ServiceType, limit int) ([]*driverentity.VehicleVerification, error) {
	return r.ListByFilter(ctx, status, vt, st, limit)
}

type fakeDocRepo struct {
	// versions[driverID][docType] holds every uploaded version, oldest first.
	versions map[string]map[driverentity.DocumentType][]*driverentity.KYCDocument
}

func newFakeDocRepo() *fakeDocRepo {
	return &fakeDocRepo{versions: map[string]map[driverentity.DocumentType][]*driverentity.KYCDocument{}}
}

func (r *fakeDocRepo) Save(_ context.Context, d *driverentity.KYCDocument) error {
	cp := *d
	if r.versions[d.DriverID] == nil {
		r.versions[d.DriverID] = map[driverentity.DocumentType][]*driverentity.KYCDocument{}
	}
	r.versions[d.DriverID][d.DocumentType] = append(r.versions[d.DriverID][d.DocumentType], &cp)
	return nil
}

func (r *fakeDocRepo) FindByDriverAndType(_ context.Context, driverID string, dt driverentity.DocumentType) (*driverentity.KYCDocument, error) {
	list := r.versions[driverID][dt]
	if len(list) == 0 {
		return nil, domainerrors.NotFound("kyc document not found")
	}
	return list[len(list)-1], nil
}

func (r *fakeDocRepo) ListByDriverID(_ context.Context, driverID string) ([]*driverentity.KYCDocument, error) {
	var out []*driverentity.KYCDocument
	for _, list := range r.versions[driverID] {
		if len(list) > 0 {
			out = append(out, list[len(list)-1])
		}
	}
	return out, nil
}

func (r *fakeDocRepo) ListVersionsByDriverAndType(_ context.Context, driverID string, dt driverentity.DocumentType) ([]*driverentity.KYCDocument, error) {
	list := r.versions[driverID][dt]
	out := make([]*driverentity.KYCDocument, len(list))
	for i, d := range list {
		out[len(list)-1-i] = d
	}
	return out, nil
}

func (r *fakeDocRepo) FindByID(_ context.Context, id string) (*driverentity.KYCDocument, error) {
	for _, byType := range r.versions {
		for _, list := range byType {
			for _, d := range list {
				if d.ID == id {
					return d, nil
				}
			}
		}
	}
	return nil, domainerrors.NotFound("kyc document not found")
}

// fakeLicenseRules mirrors the seeded migration matrix.
type fakeLicenseRules struct{}

func (fakeLicenseRules) IsAllowed(_ context.Context, licenseClass driverentity.LicenseClass, serviceType driverentity.ServiceType) (bool, error) {
	allow := map[string]bool{
		"A1|motorcycle": true, "A1|bike_plus": true,
		"A2|motorcycle": true, "A2|bike_plus": true,
		"B1|car": true, "B1|car_xl": true,
		"B2|motorcycle": true, "B2|bike_plus": true, "B2|car": true, "B2|car_xl": true,
	}
	return allow[string(licenseClass)+"|"+string(serviceType)], nil
}

type fakeAuditRepo struct {
	entries []*driverentity.AuditLog
}

func (r *fakeAuditRepo) Save(_ context.Context, log *driverentity.AuditLog) error {
	r.entries = append(r.entries, log)
	return nil
}

func (r *fakeAuditRepo) ListByDriverID(_ context.Context, driverID string, _ int) ([]*driverentity.AuditLog, error) {
	var out []*driverentity.AuditLog
	for _, e := range r.entries {
		if e.DriverID == driverID {
			out = append(out, e)
		}
	}
	return out, nil
}

// ─── builder ─────────────────────────────────────────────────────────────────

func buildKYCHandler(dv *fakeDVRepo, vv *fakeVVRepo, docs *fakeDocRepo) *handlers.KYCHandler {
	audit := &fakeAuditRepo{}
	return handlers.NewKYCHandler(
		driverapp.NewSubmitDriverVerificationUseCase(dv, docs, audit),
		driverapp.NewUpdateDriverVerificationUseCase(dv, audit),
		driverapp.NewGetDriverVerificationUseCase(dv),
		driverapp.NewSubmitVehicleVerificationUseCase(vv, docs, fakeLicenseRules{}, audit),
		driverapp.NewUpdateVehicleVerificationUseCase(vv, fakeLicenseRules{}, audit),
		driverapp.NewGetVehicleVerificationUseCase(vv),
		nil, // upload use case not exercised by these JSON-body tests
		driverapp.NewListKYCDocumentsUseCase(docs),
		driverapp.NewListKYCDocumentVersionsUseCase(docs),
	)
}

func kycRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	return injectClaims(r, &jwt.AccessClaims{UserID: "d1", UserType: "driver"})
}

func seedKYCDoc(t *testing.T, docs *fakeDocRepo, driverID string, dt driverentity.DocumentType) {
	t.Helper()
	doc, err := driverentity.NewKYCDocument("doc-"+driverID+"-"+string(dt), driverID, dt, "path/"+string(dt), "image/jpeg", 1, nil, driverID, time.Now().UTC())
	if err != nil {
		t.Fatalf("seedKYCDoc: %v", err)
	}
	_ = docs.Save(context.Background(), doc)
}

// ─── SubmitDriverVerification ───────────────────────────────────────────────

func TestKYCHandler_ServiceUnavailableWhenNotConfigured(t *testing.T) {
	h := handlers.NewKYCHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	h.SubmitDriverVerification(w, kycRequest(t, http.MethodPost, "/api/v1/driver/verification", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestKYCHandler_SubmitDriverVerification_RejectsMissingDocuments(t *testing.T) {
	h := buildKYCHandler(newFakeDVRepo(), newFakeVVRepo(), newFakeDocRepo())
	w := httptest.NewRecorder()
	h.SubmitDriverVerification(w, kycRequest(t, http.MethodPost, "/api/v1/driver/verification", map[string]any{
		"full_name": "Nguyen Van A", "date_of_birth": "1995-05-20", "address": "123 Le Loi", "national_id_number": "079095001234",
	}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 when required documents are missing, body=%s", w.Code, w.Body.String())
	}
}

func TestKYCHandler_SubmitDriverVerification_OK(t *testing.T) {
	docs := newFakeDocRepo()
	for _, dt := range driverentity.DriverDocumentTypes {
		seedKYCDoc(t, docs, "d1", dt)
	}
	h := buildKYCHandler(newFakeDVRepo(), newFakeVVRepo(), docs)

	w := httptest.NewRecorder()
	h.SubmitDriverVerification(w, kycRequest(t, http.MethodPost, "/api/v1/driver/verification", map[string]any{
		"full_name": "Nguyen Van A", "date_of_birth": "1995-05-20", "address": "123 Le Loi", "national_id_number": "079095001234",
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "pending" {
		t.Errorf("status field = %v, want pending", body["status"])
	}
	if _, hasPath := body["storage_path"]; hasPath {
		t.Error("response must never contain a storage_path field")
	}
}

func TestKYCHandler_SubmitDriverVerification_RejectsEmptyNationalIDNumber(t *testing.T) {
	docs := newFakeDocRepo()
	for _, dt := range driverentity.DriverDocumentTypes {
		seedKYCDoc(t, docs, "d1", dt)
	}
	h := buildKYCHandler(newFakeDVRepo(), newFakeVVRepo(), docs)

	w := httptest.NewRecorder()
	h.SubmitDriverVerification(w, kycRequest(t, http.MethodPost, "/api/v1/driver/verification", map[string]any{
		"full_name": "Nguyen Van A", "date_of_birth": "1995-05-20", "address": "123 Le Loi", "national_id_number": "",
	}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for empty national_id_number, body=%s", w.Code, w.Body.String())
	}
}

func TestKYCHandler_GetDriverVerification_NotFound(t *testing.T) {
	h := buildKYCHandler(newFakeDVRepo(), newFakeVVRepo(), newFakeDocRepo())
	w := httptest.NewRecorder()
	h.GetDriverVerification(w, kycRequest(t, http.MethodGet, "/api/v1/driver/verification", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404, body=%s", w.Code, w.Body.String())
	}
}

// ─── SubmitVehicleVerification ──────────────────────────────────────────────

func TestKYCHandler_SubmitVehicleVerification_RejectsMissingDocuments(t *testing.T) {
	h := buildKYCHandler(newFakeDVRepo(), newFakeVVRepo(), newFakeDocRepo())
	w := httptest.NewRecorder()
	h.SubmitVehicleVerification(w, kycRequest(t, http.MethodPost, "/api/v1/vehicle/verification", map[string]any{
		"vehicle_type": "motorcycle", "service_type": "motorcycle", "plate_number": "59H1-11111", "delivery_enabled": true,
	}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 when vehicle documents are missing, body=%s", w.Code, w.Body.String())
	}
}

func TestKYCHandler_SubmitVehicleVerification_OK(t *testing.T) {
	docs := newFakeDocRepo()
	for _, dt := range driverentity.VehicleDocumentTypes {
		seedKYCDoc(t, docs, "d1", dt)
	}
	h := buildKYCHandler(newFakeDVRepo(), newFakeVVRepo(), docs)

	w := httptest.NewRecorder()
	h.SubmitVehicleVerification(w, kycRequest(t, http.MethodPost, "/api/v1/vehicle/verification", map[string]any{
		"vehicle_type": "motorcycle", "service_type": "motorcycle",
		"brand": "Honda", "model": "Wave", "year": 2022, "color": "Đỏ",
		"plate_number": "59H1-22222", "delivery_enabled": true,
	}))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if _, hasVIN := body["vin"]; !hasVIN {
		t.Error("response should include a vin field (even if empty)")
	}
	if _, hasPerms := body["permissions"]; !hasPerms {
		t.Error("response should include the derived permissions field (Phần 8)")
	}
}

func TestKYCHandler_SubmitVehicleVerification_RejectsLicenseClassNotAllowed(t *testing.T) {
	docs := newFakeDocRepo()
	for _, dt := range driverentity.VehicleDocumentTypes {
		seedKYCDoc(t, docs, "d1", dt)
	}
	seedKYCDoc(t, docs, "d1", driverentity.DocumentLicense)
	h := buildKYCHandler(newFakeDVRepo(), newFakeVVRepo(), docs)

	w := httptest.NewRecorder()
	h.SubmitVehicleVerification(w, kycRequest(t, http.MethodPost, "/api/v1/vehicle/verification", map[string]any{
		"vehicle_type": "car", "service_type": "car", "plate_number": "59H1-30001",
		"license_class": "A1", "ride_enabled": true,
	}))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for a license class the Rule Engine doesn't allow, body=%s", w.Code, w.Body.String())
	}
}

// ─── Documents checklist ────────────────────────────────────────────────────

func TestKYCHandler_ListDocuments_ReportsUploadedTrueFalse(t *testing.T) {
	docs := newFakeDocRepo()
	seedKYCDoc(t, docs, "d1", driverentity.DocumentCCCDFront)
	h := buildKYCHandler(newFakeDVRepo(), newFakeVVRepo(), docs)

	w := httptest.NewRecorder()
	h.ListDocuments(w, kycRequest(t, http.MethodGet, "/api/v1/driver/verification/documents", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	items, _ := body["documents"].([]any)
	if len(items) == 0 {
		t.Fatal("expected the full document-type checklist, got none")
	}
	found := false
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if item["document_type"] == "cccd_front" {
			found = true
			if item["uploaded"] != true {
				t.Error("cccd_front should be reported as uploaded")
			}
			if item["version"] != float64(1) {
				t.Errorf("version = %v, want 1", item["version"])
			}
		}
		if item["document_type"] == "selfie" && item["uploaded"] != false {
			t.Error("selfie should be reported as not uploaded")
		}
	}
	if !found {
		t.Error("expected cccd_front in the checklist")
	}
}

// ─── multipart upload (exercised via a real ApiClient-shaped request) ──────

func buildKYCHandlerWithUpload(t *testing.T, dv *fakeDVRepo, vv *fakeVVRepo, docs *fakeDocRepo) *handlers.KYCHandler {
	t.Helper()
	audit := &fakeAuditRepo{}
	store := driverlocalstore.NewDocumentStore(t.TempDir())
	uploadUC := driverapp.NewUploadKYCDocumentUseCase(docs, dv, vv, audit, store)
	return handlers.NewKYCHandler(
		driverapp.NewSubmitDriverVerificationUseCase(dv, docs, audit),
		driverapp.NewUpdateDriverVerificationUseCase(dv, audit),
		driverapp.NewGetDriverVerificationUseCase(dv),
		driverapp.NewSubmitVehicleVerificationUseCase(vv, docs, fakeLicenseRules{}, audit),
		driverapp.NewUpdateVehicleVerificationUseCase(vv, fakeLicenseRules{}, audit),
		driverapp.NewGetVehicleVerificationUseCase(vv),
		uploadUC,
		driverapp.NewListKYCDocumentsUseCase(docs),
		driverapp.NewListKYCDocumentVersionsUseCase(docs),
	)
}

func TestKYCHandler_UploadDocument_OK(t *testing.T) {
	dv, vv, docs := newFakeDVRepo(), newFakeVVRepo(), newFakeDocRepo()
	h := buildKYCHandlerWithUpload(t, dv, vv, docs)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("document_type", "selfie")
	fw, _ := mw.CreateFormFile("file", "selfie.jpg")
	_, _ = io.Copy(fw, bytes.NewBufferString("fake-image-bytes"))
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/api/v1/driver/verification/documents", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r = injectClaims(r, &jwt.AccessClaims{UserID: "d1", UserType: "driver"})

	w := httptest.NewRecorder()
	h.UploadDocument(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["uploaded"] != true {
		t.Error("expected uploaded=true")
	}
	if body["version"] != float64(1) {
		t.Errorf("version = %v, want 1", body["version"])
	}
	if _, hasPath := body["storage_path"]; hasPath {
		t.Error("response must never contain a storage_path field")
	}
}

func TestKYCHandler_UploadDocument_WithExpiryDate(t *testing.T) {
	dv, vv, docs := newFakeDVRepo(), newFakeVVRepo(), newFakeDocRepo()
	h := buildKYCHandlerWithUpload(t, dv, vv, docs)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("document_type", "vehicle_insurance")
	_ = mw.WriteField("expires_at", "2030-01-01")
	fw, _ := mw.CreateFormFile("file", "insurance.jpg")
	_, _ = io.Copy(fw, bytes.NewBufferString("fake-image-bytes"))
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/api/v1/driver/verification/documents", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r = injectClaims(r, &jwt.AccessClaims{UserID: "d1", UserType: "driver"})

	w := httptest.NewRecorder()
	h.UploadDocument(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["expires_at"] != "2030-01-01" {
		t.Errorf("expires_at = %v, want 2030-01-01", body["expires_at"])
	}
	if body["expired"] != false {
		t.Error("a 2030 expiry should not be reported as expired")
	}
}

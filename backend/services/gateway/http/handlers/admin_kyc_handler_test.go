package handlers_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	driverapp "github.com/fairride/driver/app"
	driverentity "github.com/fairride/driver/domain/entity"
	driverlocalstore "github.com/fairride/driver/infrastructure/localstore"
	"github.com/fairride/gateway/http/handlers"
	identityentity "github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/infrastructure/jwt"
	domainerrors "github.com/fairride/shared/errors"
)

// ─── phone-enrichment fakes ─────────────────────────────────────────────────

type fakeDriverFinderByID struct {
	byID map[string]*driverentity.DriverProfile
}

func (f *fakeDriverFinderByID) FindByID(_ context.Context, driverID string) (*driverentity.DriverProfile, error) {
	if p, ok := f.byID[driverID]; ok {
		return p, nil
	}
	return nil, domainerrors.NotFound("driver not found")
}

type fakeUserFinderByID struct {
	byID map[string]*identityentity.User
}

func (f *fakeUserFinderByID) FindByID(_ context.Context, userID string) (*identityentity.User, error) {
	if u, ok := f.byID[userID]; ok {
		return u, nil
	}
	return nil, domainerrors.NotFound("user not found")
}

func buildAdminKYCHandler(dv *fakeDVRepo, vv *fakeVVRepo, docs *fakeDocRepo, store string) *handlers.AdminKYCHandler {
	audit := &fakeAuditRepo{}
	return handlers.NewAdminKYCHandler(
		driverapp.NewListDriverVerificationsUseCase(dv),
		driverapp.NewReviewDriverVerificationUseCase(dv, audit),
		driverapp.NewGetDriverVerificationUseCase(dv),
		driverapp.NewListVehicleVerificationsUseCase(vv),
		driverapp.NewReviewVehicleVerificationUseCase(vv, audit),
		driverapp.NewGetVehicleVerificationUseCase(vv),
		driverapp.NewListKYCDocumentsUseCase(docs),
		driverapp.NewGetKYCDocumentUseCase(docs),
		driverlocalstore.NewDocumentStore(store),
		driverapp.NewListAuditLogsUseCase(audit),
		driverapp.NewGetKYCSummaryUseCase(dv),
		nil, nil,
	)
}

func adminRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	return injectClaims(r, &jwt.AccessClaims{UserID: "admin1", UserType: "admin"})
}

func TestAdminKYCHandler_ServiceUnavailableWhenNotConfigured(t *testing.T) {
	h := handlers.NewAdminKYCHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	h.ListDriverVerifications(w, adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/drivers", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestAdminKYCHandler_ListDriverVerifications_DefaultsToPending(t *testing.T) {
	dv := newFakeDVRepo()
	v, _ := driverentity.NewDriverVerification("id1", "d1", "A", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095001234", "", time.Now().UTC())
	_ = dv.Save(context.Background(), v)
	h := buildAdminKYCHandler(dv, newFakeVVRepo(), newFakeDocRepo(), t.TempDir())

	w := httptest.NewRecorder()
	h.ListDriverVerifications(w, adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/drivers", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	items, _ := body["verifications"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 pending verification, got %d", len(items))
	}
}

func TestAdminKYCHandler_ApproveDriverVerification(t *testing.T) {
	dv := newFakeDVRepo()
	v, _ := driverentity.NewDriverVerification("id1", "d1", "A", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095001234", "", time.Now().UTC())
	_ = dv.Save(context.Background(), v)
	h := buildAdminKYCHandler(dv, newFakeVVRepo(), newFakeDocRepo(), t.TempDir())

	r := adminRequest(t, http.MethodPost, "/api/v1/admin/verifications/drivers/d1/approve", nil)
	r.SetPathValue("driverID", "d1")
	w := httptest.NewRecorder()
	h.ApproveDriverVerification(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "approved" {
		t.Errorf("status = %v, want approved", body["status"])
	}
}

func TestAdminKYCHandler_RejectDriverVerification_RequiresReason(t *testing.T) {
	dv := newFakeDVRepo()
	v, _ := driverentity.NewDriverVerification("id1", "d1", "A", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095001234", "", time.Now().UTC())
	_ = dv.Save(context.Background(), v)
	h := buildAdminKYCHandler(dv, newFakeVVRepo(), newFakeDocRepo(), t.TempDir())

	r := adminRequest(t, http.MethodPost, "/api/v1/admin/verifications/drivers/d1/reject", map[string]any{"reason": ""})
	r.SetPathValue("driverID", "d1")
	w := httptest.NewRecorder()
	h.RejectDriverVerification(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for empty reject reason, body=%s", w.Code, w.Body.String())
	}

	r2 := adminRequest(t, http.MethodPost, "/api/v1/admin/verifications/drivers/d1/reject", map[string]any{"reason": "CCCD mờ"})
	r2.SetPathValue("driverID", "d1")
	w2 := httptest.NewRecorder()
	h.RejectDriverVerification(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 with a real reason, body=%s", w2.Code, w2.Body.String())
	}
}

func TestAdminKYCHandler_ListDriverDocuments_IncludesDocumentID(t *testing.T) {
	docs := newFakeDocRepo()
	seedKYCDoc(t, docs, "d1", driverentity.DocumentCCCDFront)
	h := buildAdminKYCHandler(newFakeDVRepo(), newFakeVVRepo(), docs, t.TempDir())

	r := adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/drivers/d1/documents", nil)
	r.SetPathValue("driverID", "d1")
	w := httptest.NewRecorder()
	h.ListDriverDocuments(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	items, _ := body["documents"].([]any)
	found := false
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if item["document_type"] == "cccd_front" {
			found = true
			if item["document_id"] == nil || item["document_id"] == "" {
				t.Error("uploaded document should carry a document_id for the admin viewer to fetch")
			}
		}
	}
	if !found {
		t.Fatal("expected cccd_front in the admin checklist")
	}
}

// ─── Phone enrichment + search/sort ─────────────────────────────────────────

func buildAdminKYCHandlerWithPhone(dv *fakeDVRepo, vv *fakeVVRepo, docs *fakeDocRepo, store string, drivers *fakeDriverFinderByID, users *fakeUserFinderByID) *handlers.AdminKYCHandler {
	audit := &fakeAuditRepo{}
	return handlers.NewAdminKYCHandler(
		driverapp.NewListDriverVerificationsUseCase(dv),
		driverapp.NewReviewDriverVerificationUseCase(dv, audit),
		driverapp.NewGetDriverVerificationUseCase(dv),
		driverapp.NewListVehicleVerificationsUseCase(vv),
		driverapp.NewReviewVehicleVerificationUseCase(vv, audit),
		driverapp.NewGetVehicleVerificationUseCase(vv),
		driverapp.NewListKYCDocumentsUseCase(docs),
		driverapp.NewGetKYCDocumentUseCase(docs),
		driverlocalstore.NewDocumentStore(store),
		driverapp.NewListAuditLogsUseCase(audit),
		driverapp.NewGetKYCSummaryUseCase(dv),
		drivers, users,
	)
}

func TestAdminKYCHandler_ListDriverVerifications_EnrichesPhoneAndSearches(t *testing.T) {
	dv := newFakeDVRepo()
	v1, _ := driverentity.NewDriverVerification("id1", "d1", "Nguyen Van A", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095001234", "", time.Now().UTC())
	v2, _ := driverentity.NewDriverVerification("id2", "d2", "Tran Thi B", time.Date(1996, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095009999", "", time.Now().UTC().Add(time.Minute))
	_ = dv.Save(context.Background(), v1)
	_ = dv.Save(context.Background(), v2)

	drivers := &fakeDriverFinderByID{byID: map[string]*driverentity.DriverProfile{
		"d1": driverentity.ReconstituteDriverProfile("d1", "user-1", "LIC1", driverentity.VehicleTypeCar, "", "", "", "P1", driverentity.OnlineStatusOffline, driverentity.VerificationStatusVerified, time.Now(), time.Now(), driverentity.ServiceTypeCar, true, false),
		"d2": driverentity.ReconstituteDriverProfile("d2", "user-2", "LIC2", driverentity.VehicleTypeCar, "", "", "", "P2", driverentity.OnlineStatusOffline, driverentity.VerificationStatusVerified, time.Now(), time.Now(), driverentity.ServiceTypeCar, true, false),
	}}
	users := &fakeUserFinderByID{byID: map[string]*identityentity.User{
		"user-1": identityentity.ReconstituteUser("user-1", "+84900001111", "A", "", "", identityentity.TypeDriver, identityentity.StatusActive, "role-1", false, time.Now(), time.Now()),
		"user-2": identityentity.ReconstituteUser("user-2", "+84900002222", "B", "", "", identityentity.TypeDriver, identityentity.StatusActive, "role-1", false, time.Now(), time.Now()),
	}}
	h := buildAdminKYCHandlerWithPhone(dv, newFakeVVRepo(), newFakeDocRepo(), t.TempDir(), drivers, users)

	w := httptest.NewRecorder()
	h.ListDriverVerifications(w, adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/drivers", nil))
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	items, _ := body["verifications"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 verifications, got %d", len(items))
	}
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if item["driver_id"] == "d1" && item["phone"] != "+84900001111" {
			t.Errorf("d1 phone = %v, want +84900001111", item["phone"])
		}
	}

	// Search by phone substring should return only d2.
	w2 := httptest.NewRecorder()
	h.ListDriverVerifications(w2, adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/drivers?q=2222", nil))
	var body2 map[string]any
	_ = json.NewDecoder(w2.Body).Decode(&body2)
	items2, _ := body2["verifications"].([]any)
	if len(items2) != 1 {
		t.Fatalf("expected 1 verification matching phone search, got %d", len(items2))
	}
	first, _ := items2[0].(map[string]any)
	if first["driver_id"] != "d2" {
		t.Errorf("driver_id = %v, want d2", first["driver_id"])
	}
}

func TestAdminKYCHandler_ListDriverVerifications_SortAscOrdersOldestFirst(t *testing.T) {
	dv := newFakeDVRepo()
	older, _ := driverentity.NewDriverVerification("id1", "d1", "A", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095001111", "", time.Now().UTC().Add(-time.Hour))
	newer, _ := driverentity.NewDriverVerification("id2", "d2", "B", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095002222", "", time.Now().UTC())
	_ = dv.Save(context.Background(), older)
	_ = dv.Save(context.Background(), newer)
	h := buildAdminKYCHandler(dv, newFakeVVRepo(), newFakeDocRepo(), t.TempDir())

	w := httptest.NewRecorder()
	h.ListDriverVerifications(w, adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/drivers?sort=asc", nil))
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	items, _ := body["verifications"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 verifications, got %d", len(items))
	}
	first, _ := items[0].(map[string]any)
	if first["driver_id"] != "d1" {
		t.Errorf("sort=asc: first driver_id = %v, want d1 (oldest submission)", first["driver_id"])
	}
}

// ─── Combined detail ─────────────────────────────────────────────────────────

func TestAdminKYCHandler_GetDriverKYCDetail_CombinesEverything(t *testing.T) {
	dv := newFakeDVRepo()
	v, _ := driverentity.NewDriverVerification("id1", "d1", "A", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095001234", "", time.Now().UTC())
	_ = dv.Save(context.Background(), v)
	docs := newFakeDocRepo()
	seedKYCDoc(t, docs, "d1", driverentity.DocumentCCCDFront)
	h := buildAdminKYCHandler(dv, newFakeVVRepo(), docs, t.TempDir())

	r := adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/drivers/d1/detail", nil)
	r.SetPathValue("driverID", "d1")
	w := httptest.NewRecorder()
	h.GetDriverKYCDetail(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["driver_verification"] == nil {
		t.Error("expected driver_verification in detail response")
	}
	if body["documents"] == nil {
		t.Error("expected documents checklist in detail response")
	}
	if body["vehicle_verification"] != nil {
		t.Error("expected nil vehicle_verification — none was submitted")
	}
}

// ─── ZIP download ────────────────────────────────────────────────────────────

func TestAdminKYCHandler_DownloadDriverDocumentsZip_ContainsUploadedFile(t *testing.T) {
	docs := newFakeDocRepo()
	storeDir := t.TempDir()
	store := driverlocalstore.NewDocumentStore(storeDir)
	relPath, err := store.Save(context.Background(), "d1", "cccd_front", ".jpg", strings.NewReader("fake-jpeg-bytes"))
	if err != nil {
		t.Fatalf("store.Save: %v", err)
	}
	doc, _ := driverentity.NewKYCDocument("doc1", "d1", driverentity.DocumentCCCDFront, relPath, "image/jpeg", 1, nil, "d1", time.Now().UTC())
	_ = docs.Save(context.Background(), doc)

	h := buildAdminKYCHandler(newFakeDVRepo(), newFakeVVRepo(), docs, storeDir)
	r := adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/drivers/d1/documents.zip", nil)
	r.SetPathValue("driverID", "d1")
	w := httptest.NewRecorder()
	h.DownloadDriverDocumentsZip(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	zr, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil {
		t.Fatalf("response is not a valid zip: %v", err)
	}
	if len(zr.File) != 1 {
		t.Fatalf("expected 1 file in zip, got %d", len(zr.File))
	}
	rc, err := zr.File[0].Open()
	if err != nil {
		t.Fatalf("open zip entry: %v", err)
	}
	defer rc.Close()
	content, _ := io.ReadAll(rc)
	if string(content) != "fake-jpeg-bytes" {
		t.Errorf("zip entry content = %q, want %q", content, "fake-jpeg-bytes")
	}
}

// ─── Dashboard summary ───────────────────────────────────────────────────────

func TestAdminKYCHandler_GetKYCSummary_ReturnsCounts(t *testing.T) {
	dv := newFakeDVRepo()
	pending, _ := driverentity.NewDriverVerification("id1", "d1", "A", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095001111", "", time.Now().UTC())
	approved, _ := driverentity.NewDriverVerification("id2", "d2", "B", time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), "addr", "079095002222", "", time.Now().UTC())
	_ = approved.Approve("admin1", time.Now().UTC())
	_ = dv.Save(context.Background(), pending)
	_ = dv.Save(context.Background(), approved)
	h := buildAdminKYCHandler(dv, newFakeVVRepo(), newFakeDocRepo(), t.TempDir())

	w := httptest.NewRecorder()
	h.GetKYCSummary(w, adminRequest(t, http.MethodGet, "/api/v1/admin/verifications/summary", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["pending"] != float64(1) {
		t.Errorf("pending = %v, want 1", body["pending"])
	}
	if body["approved"] != float64(1) {
		t.Errorf("approved = %v, want 1", body["approved"])
	}
}

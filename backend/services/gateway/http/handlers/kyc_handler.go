package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	driverapp "github.com/fairride/driver/app"
	driverentity "github.com/fairride/driver/domain/entity"
	"github.com/fairride/gateway/http/middleware"
)

const dobLayout = "2006-01-02"
const expiryLayout = "2006-01-02"

// documentsExpiringSoonWindow is Phần 11's "gần hết hạn 30 ngày -> banner vàng".
const documentsExpiringSoonWindow = 30 * 24 * time.Hour

// KYCHandler exposes Driver KYC + Vehicle Verification submission/reading
// over HTTP, scoped to the currently-authenticated driver only (Phần 9/13 —
// a driver only ever sees/edits their own record; claims.UserID is always
// used as the driver_id, never a client-supplied one). Runs entirely
// in-process against the driver service's own Go packages — no new gRPC
// surface, same pattern as the Communication Module's notification/review
// wiring (no protoc/buf toolchain in this environment).
type KYCHandler struct {
	submitDriver *driverapp.SubmitDriverVerificationUseCase
	updateDriver *driverapp.UpdateDriverVerificationUseCase
	getDriver    *driverapp.GetDriverVerificationUseCase

	submitVehicle *driverapp.SubmitVehicleVerificationUseCase
	updateVehicle *driverapp.UpdateVehicleVerificationUseCase
	getVehicle    *driverapp.GetVehicleVerificationUseCase

	uploadDocument       *driverapp.UploadKYCDocumentUseCase
	listDocuments        *driverapp.ListKYCDocumentsUseCase
	listDocumentVersions *driverapp.ListKYCDocumentVersionsUseCase
}

func NewKYCHandler(
	submitDriver *driverapp.SubmitDriverVerificationUseCase,
	updateDriver *driverapp.UpdateDriverVerificationUseCase,
	getDriver *driverapp.GetDriverVerificationUseCase,
	submitVehicle *driverapp.SubmitVehicleVerificationUseCase,
	updateVehicle *driverapp.UpdateVehicleVerificationUseCase,
	getVehicle *driverapp.GetVehicleVerificationUseCase,
	uploadDocument *driverapp.UploadKYCDocumentUseCase,
	listDocuments *driverapp.ListKYCDocumentsUseCase,
	listDocumentVersions *driverapp.ListKYCDocumentVersionsUseCase,
) *KYCHandler {
	return &KYCHandler{
		submitDriver: submitDriver, updateDriver: updateDriver, getDriver: getDriver,
		submitVehicle: submitVehicle, updateVehicle: updateVehicle, getVehicle: getVehicle,
		uploadDocument: uploadDocument, listDocuments: listDocuments, listDocumentVersions: listDocumentVersions,
	}
}

func (h *KYCHandler) configured() bool { return h != nil && h.submitDriver != nil }

func (h *KYCHandler) unavailable(w http.ResponseWriter) bool {
	if h.configured() {
		return false
	}
	writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "KYC service not configured"})
	return true
}

// ─── Driver verification (personal identity KYC) ──────────────────────────────

type driverVerificationRequest struct {
	FullName         string `json:"full_name"`
	DateOfBirth      string `json:"date_of_birth"` // "YYYY-MM-DD"
	Address          string `json:"address"`
	NationalIDNumber string `json:"national_id_number"`
	LicenseNumber    string `json:"license_number"`
}

func (h *KYCHandler) parseDriverVerificationRequest(w http.ResponseWriter, r *http.Request) (driverapp.SubmitDriverVerificationInput, string, bool) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return driverapp.SubmitDriverVerificationInput{}, "", false
	}
	var req driverVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return driverapp.SubmitDriverVerificationInput{}, "", false
	}
	dob, err := time.Parse(dobLayout, req.DateOfBirth)
	if err != nil {
		writeBadRequest(w, "date_of_birth must be in YYYY-MM-DD format")
		return driverapp.SubmitDriverVerificationInput{}, "", false
	}
	return driverapp.SubmitDriverVerificationInput{
		DriverID:         claims.UserID,
		FullName:         req.FullName,
		DateOfBirth:      dob,
		Address:          req.Address,
		NationalIDNumber: req.NationalIDNumber,
		LicenseNumber:    req.LicenseNumber,
	}, claims.UserID, true
}

// SubmitDriverVerification handles POST /api/v1/driver/verification.
func (h *KYCHandler) SubmitDriverVerification(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	in, _, ok := h.parseDriverVerificationRequest(w, r)
	if !ok {
		return
	}
	v, err := h.submitDriver.Execute(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, driverVerificationJSON(v))
}

// UpdateDriverVerification handles PUT /api/v1/driver/verification.
func (h *KYCHandler) UpdateDriverVerification(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	in, _, ok := h.parseDriverVerificationRequest(w, r)
	if !ok {
		return
	}
	v, err := h.updateDriver.Execute(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, driverVerificationJSON(v))
}

// GetDriverVerification handles GET /api/v1/driver/verification.
func (h *KYCHandler) GetDriverVerification(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	v, err := h.getDriver.Execute(r.Context(), claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, driverVerificationJSON(v))
}

func driverVerificationJSON(v *driverentity.DriverVerification) map[string]any {
	body := map[string]any{
		"id":                 v.ID,
		"driver_id":          v.DriverID,
		"full_name":          v.FullName,
		"date_of_birth":      v.DateOfBirth.Format(dobLayout),
		"address":            v.Address,
		"national_id_number": v.NationalIDNumber,
		"license_number":     v.LicenseNumber,
		"status":             string(v.Status),
		"submitted_at":       v.SubmittedAt.UTC().Format(time.RFC3339),
		"reject_reason":      v.RejectReason,
	}
	if v.ApprovedAt != nil {
		body["approved_at"] = v.ApprovedAt.UTC().Format(time.RFC3339)
	}
	if v.RejectedAt != nil {
		body["rejected_at"] = v.RejectedAt.UTC().Format(time.RFC3339)
	}
	if v.ExpiredAt != nil {
		body["expired_at"] = v.ExpiredAt.UTC().Format(time.RFC3339)
	}
	return body
}

// ─── Vehicle verification ──────────────────────────────────────────────────────

type vehicleVerificationRequest struct {
	VehicleType     string `json:"vehicle_type"`
	ServiceType     string `json:"service_type"`
	Brand           string `json:"brand"`
	Model           string `json:"model"`
	Year            int    `json:"year"`
	Color           string `json:"color"`
	PlateNumber     string `json:"plate_number"`
	VIN             string `json:"vin"`
	EngineNumber    string `json:"engine_number"`
	ChassisNumber   string `json:"chassis_number"`
	LicenseClass    string `json:"license_class"`
	RideEnabled     bool   `json:"ride_enabled"`
	DeliveryEnabled bool   `json:"delivery_enabled"`
}

func (h *KYCHandler) parseVehicleVerificationRequest(w http.ResponseWriter, r *http.Request) (driverapp.SubmitVehicleVerificationInput, bool) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return driverapp.SubmitVehicleVerificationInput{}, false
	}
	var req vehicleVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return driverapp.SubmitVehicleVerificationInput{}, false
	}
	return driverapp.SubmitVehicleVerificationInput{
		DriverID:        claims.UserID,
		VehicleType:     driverentity.VehicleType(req.VehicleType),
		ServiceType:     driverentity.ServiceType(req.ServiceType),
		Brand:           req.Brand,
		Model:           req.Model,
		Year:            req.Year,
		Color:           req.Color,
		PlateNumber:     req.PlateNumber,
		VIN:             req.VIN,
		EngineNumber:    req.EngineNumber,
		ChassisNumber:   req.ChassisNumber,
		LicenseClass:    driverentity.LicenseClass(req.LicenseClass),
		RideEnabled:     req.RideEnabled,
		DeliveryEnabled: req.DeliveryEnabled,
	}, true
}

// SubmitVehicleVerification handles POST /api/v1/vehicle/verification.
func (h *KYCHandler) SubmitVehicleVerification(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	in, ok := h.parseVehicleVerificationRequest(w, r)
	if !ok {
		return
	}
	v, err := h.submitVehicle.Execute(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, vehicleVerificationJSON(v))
}

// UpdateVehicleVerification handles PUT /api/v1/vehicle/verification.
func (h *KYCHandler) UpdateVehicleVerification(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	in, ok := h.parseVehicleVerificationRequest(w, r)
	if !ok {
		return
	}
	v, err := h.updateVehicle.Execute(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vehicleVerificationJSON(v))
}

// GetVehicleVerification handles GET /api/v1/vehicle/verification.
func (h *KYCHandler) GetVehicleVerification(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	v, err := h.getVehicle.Execute(r.Context(), claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vehicleVerificationJSON(v))
}

func vehicleVerificationJSON(v *driverentity.VehicleVerification) map[string]any {
	permissions := v.Permissions()
	permStrs := make([]string, len(permissions))
	for i, p := range permissions {
		permStrs[i] = string(p)
	}
	body := map[string]any{
		"id":               v.ID,
		"driver_id":        v.DriverID,
		"vehicle_type":     string(v.VehicleType),
		"service_type":     string(v.ServiceType),
		"brand":            v.Brand,
		"model":            v.Model,
		"year":             v.Year,
		"color":            v.Color,
		"plate_number":     v.PlateNumber,
		"vin":              v.VIN,
		"engine_number":    v.EngineNumber,
		"chassis_number":   v.ChassisNumber,
		"license_class":    string(v.LicenseClass),
		"ride_enabled":     v.RideEnabled,
		"delivery_enabled": v.DeliveryEnabled,
		"permissions":      permStrs,
		"status":           string(v.Status),
		"submitted_at":     v.SubmittedAt.UTC().Format(time.RFC3339),
		"reject_reason":    v.RejectReason,
	}
	if v.ApprovedAt != nil {
		body["approved_at"] = v.ApprovedAt.UTC().Format(time.RFC3339)
	}
	if v.RejectedAt != nil {
		body["rejected_at"] = v.RejectedAt.UTC().Format(time.RFC3339)
	}
	if v.ExpiredAt != nil {
		body["expired_at"] = v.ExpiredAt.UTC().Format(time.RFC3339)
	}
	return body
}

// ─── Documents ──────────────────────────────────────────────────────────────

// allDocumentTypes drives the "uploaded: true/false" checklist response —
// keep in sync with driverentity's DriverDocumentTypes/RideLicenseDocumentTypes/VehicleDocumentTypes.
var allDocumentTypes = []driverentity.DocumentType{
	driverentity.DocumentCCCDFront,
	driverentity.DocumentCCCDBack,
	driverentity.DocumentSelfie,
	driverentity.DocumentLicense,
	driverentity.DocumentVehicleRegistration,
	driverentity.DocumentVehicleInsurance,
	driverentity.DocumentVehicleInspection,
}

// UploadDocument handles POST /api/v1/driver/verification/documents
// (multipart/form-data: field "document_type", file field "file", optional
// field "expires_at" as YYYY-MM-DD — only meaningful for
// entity.ExpiringDocumentTypes). Never returns the on-disk storage path
// (Phần 9/13) — only confirms what was uploaded, its version, and when.
func (h *KYCHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	// 10 MiB is generous for a phone-camera photo of a document; multipart
	// bodies larger than this are rejected outright rather than silently
	// buffered to disk.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeBadRequest(w, "invalid multipart form")
		return
	}
	docType := driverentity.DocumentType(r.FormValue("document_type"))
	if err := driverentity.ValidateDocumentType(docType); err != nil {
		writeDomainError(w, err)
		return
	}
	var expiresAt *time.Time
	if raw := r.FormValue("expires_at"); raw != "" {
		parsed, err := time.Parse(expiryLayout, raw)
		if err != nil {
			writeBadRequest(w, "expires_at must be in YYYY-MM-DD format")
			return
		}
		expiresAt = &parsed
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeBadRequest(w, "file is required")
		return
	}
	defer file.Close()

	doc, err := h.uploadDocument.Execute(r.Context(), driverapp.UploadKYCDocumentInput{
		DriverID:     claims.UserID,
		DocumentType: docType,
		Filename:     header.Filename,
		ContentType:  header.Header.Get("Content-Type"),
		ExpiresAt:    expiresAt,
		Data:         file,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, documentChecklistItemJSON(doc, true))
}

// ListDocuments handles GET /api/v1/driver/verification/documents — a
// checklist of every known document type, its version/expiry, and whether
// it's been uploaded, for the wizard's Step progress UI and Phần 11's
// expiry banners.
func (h *KYCHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	uploaded, err := h.listDocuments.Execute(r.Context(), claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"documents": documentChecklistJSON(uploaded)})
}

// ListDocumentVersions handles
// GET /api/v1/driver/verification/documents/{documentType}/versions — the
// Phần 4/11 upload history for one document type, newest first.
func (h *KYCHandler) ListDocumentVersions(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	docType := driverentity.DocumentType(r.PathValue("documentType"))
	versions, err := h.listDocumentVersions.Execute(r.Context(), claims.UserID, docType)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(versions))
	for i, d := range versions {
		items[i] = documentChecklistItemJSON(d, true)
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": items})
}

func documentChecklistJSON(uploaded []*driverentity.KYCDocument) []map[string]any {
	byType := make(map[driverentity.DocumentType]*driverentity.KYCDocument, len(uploaded))
	for _, d := range uploaded {
		byType[d.DocumentType] = d
	}
	items := make([]map[string]any, 0, len(allDocumentTypes))
	for _, dt := range allDocumentTypes {
		if d, ok := byType[dt]; ok {
			items = append(items, documentChecklistItemJSON(d, true))
		} else {
			items = append(items, map[string]any{"document_type": string(dt), "uploaded": false})
		}
	}
	return items
}

// documentChecklistItemJSON never includes storage_path (Phần 9/13).
func documentChecklistItemJSON(d *driverentity.KYCDocument, uploaded bool) map[string]any {
	item := map[string]any{
		"document_type": string(d.DocumentType),
		"uploaded":      uploaded,
		"uploaded_at":   d.UploadedAt.UTC().Format(time.RFC3339),
		"version":       d.Version,
	}
	if d.ExpiresAt != nil {
		now := time.Now().UTC()
		item["expires_at"] = d.ExpiresAt.UTC().Format(dobLayout)
		item["expired"] = d.IsExpired(now)
		item["expiring_soon"] = d.ExpiresWithin(documentsExpiringSoonWindow, now)
	}
	return item
}

// documentChecklistItemJSONForAdmin adds the opaque document id (Phần 12 —
// the admin UI needs it to fetch bytes via GetDocument); never adds
// storage_path.
func documentChecklistItemJSONForAdmin(d *driverentity.KYCDocument) map[string]any {
	item := documentChecklistItemJSON(d, true)
	item["document_id"] = d.ID
	return item
}

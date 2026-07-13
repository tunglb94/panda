package handlers

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	driverapp "github.com/fairride/driver/app"
	driverentity "github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/infrastructure/localstore"
	"github.com/fairride/gateway/http/middleware"
)

// AdminKYCHandler exposes the KYC review dashboard (Phần 12) — every route
// this handler serves must be wrapped in both `auth` and
// `middleware.RequireAdmin` by the router; this handler itself does not
// re-check the role, it trusts the middleware chain (consistent with every
// other handler in this package trusting `middleware.Auth` upstream).
type AdminKYCHandler struct {
	listDriverVerifications  *driverapp.ListDriverVerificationsUseCase
	reviewDriverVerification *driverapp.ReviewDriverVerificationUseCase
	getDriverVerification    *driverapp.GetDriverVerificationUseCase

	listVehicleVerifications  *driverapp.ListVehicleVerificationsUseCase
	reviewVehicleVerification *driverapp.ReviewVehicleVerificationUseCase
	getVehicleVerification    *driverapp.GetVehicleVerificationUseCase

	listDocuments *driverapp.ListKYCDocumentsUseCase
	getDocument   *driverapp.GetKYCDocumentUseCase
	documentStore *localstore.DocumentStore

	listAuditLogs *driverapp.ListAuditLogsUseCase
	getSummary    *driverapp.GetKYCSummaryUseCase

	// Phone enrichment (Phần 1/2 — SĐT column + search). Optional: nil
	// simply omits the phone field/skips it in search instead of failing
	// the whole request — the review flow itself never depends on it.
	drivers driverByIDFinder
	users   userByIDFinder
}

func NewAdminKYCHandler(
	listDriverVerifications *driverapp.ListDriverVerificationsUseCase,
	reviewDriverVerification *driverapp.ReviewDriverVerificationUseCase,
	getDriverVerification *driverapp.GetDriverVerificationUseCase,
	listVehicleVerifications *driverapp.ListVehicleVerificationsUseCase,
	reviewVehicleVerification *driverapp.ReviewVehicleVerificationUseCase,
	getVehicleVerification *driverapp.GetVehicleVerificationUseCase,
	listDocuments *driverapp.ListKYCDocumentsUseCase,
	getDocument *driverapp.GetKYCDocumentUseCase,
	documentStore *localstore.DocumentStore,
	listAuditLogs *driverapp.ListAuditLogsUseCase,
	getSummary *driverapp.GetKYCSummaryUseCase,
	drivers driverByIDFinder,
	users userByIDFinder,
) *AdminKYCHandler {
	return &AdminKYCHandler{
		listDriverVerifications: listDriverVerifications, reviewDriverVerification: reviewDriverVerification,
		getDriverVerification:    getDriverVerification,
		listVehicleVerifications: listVehicleVerifications, reviewVehicleVerification: reviewVehicleVerification,
		getVehicleVerification: getVehicleVerification,
		listDocuments:          listDocuments, getDocument: getDocument, documentStore: documentStore,
		listAuditLogs: listAuditLogs, getSummary: getSummary,
		drivers: drivers, users: users,
	}
}

func (h *AdminKYCHandler) configured() bool { return h != nil && h.listDriverVerifications != nil }

func (h *AdminKYCHandler) unavailable(w http.ResponseWriter) bool {
	if h.configured() {
		return false
	}
	writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "admin KYC service not configured"})
	return true
}

// phoneForDriver resolves driverID -> phone number via driver profile ->
// identity user, best-effort ("" on any failure or missing dependency —
// never fails the caller's request over a cosmetic enrichment field).
func (h *AdminKYCHandler) phoneForDriver(ctx context.Context, driverID string) string {
	if h.drivers == nil || h.users == nil {
		return ""
	}
	profile, err := h.drivers.FindByID(ctx, driverID)
	if err != nil || profile == nil {
		return ""
	}
	user, err := h.users.FindByID(ctx, profile.UserID)
	if err != nil || user == nil {
		return ""
	}
	return user.PhoneNumber
}

// ─── Driver verification review ────────────────────────────────────────────────

// ListDriverVerifications handles
// GET /api/v1/admin/verifications/drivers?status=&limit=&q=&sort=
//   - status: pending/under_review/approved/rejected/expired (Phần 12's filter), defaults to pending.
//   - q: case-insensitive substring match against full name, phone, or CCCD (Phần 1's search).
//   - sort: "asc" (oldest first) or "desc" (newest first, the default).
func (h *AdminKYCHandler) ListDriverVerifications(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	status := driverentity.KYCStatus(r.URL.Query().Get("status"))
	if status == "" {
		status = driverentity.KYCPending
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 200 // admin dashboard wants "enough to search/sort across", not a small page
	}
	list, err := h.listDriverVerifications.Execute(r.Context(), status, limit)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	type row struct {
		v     *driverentity.DriverVerification
		phone string
	}
	rows := make([]row, len(list))
	for i, v := range list {
		rows[i] = row{v: v, phone: h.phoneForDriver(r.Context(), v.DriverID)}
	}

	if q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q"))); q != "" {
		filtered := rows[:0]
		for _, rw := range rows {
			if strings.Contains(strings.ToLower(rw.v.FullName), q) ||
				strings.Contains(strings.ToLower(rw.phone), q) ||
				strings.Contains(strings.ToLower(rw.v.NationalIDNumber), q) {
				filtered = append(filtered, rw)
			}
		}
		rows = filtered
	}

	// ListByStatus already returns newest-first; only re-sort for "asc".
	if r.URL.Query().Get("sort") == "asc" {
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].v.SubmittedAt.Before(rows[j].v.SubmittedAt) })
	}

	items := make([]map[string]any, len(rows))
	for i, rw := range rows {
		item := driverVerificationJSON(rw.v)
		item["phone"] = rw.phone
		if h.getVehicleVerification != nil {
			if vv, err := h.getVehicleVerification.Execute(r.Context(), rw.v.DriverID); err == nil && vv != nil {
				item["vehicle_type"] = string(vv.VehicleType)
				item["service_type"] = string(vv.ServiceType)
			}
		}
		items[i] = item
	}
	writeJSON(w, http.StatusOK, map[string]any{"verifications": items})
}

type reviewRequest struct {
	Reason string `json:"reason"`
}

// ApproveDriverVerification handles POST /api/v1/admin/verifications/drivers/{driverID}/approve.
func (h *AdminKYCHandler) ApproveDriverVerification(w http.ResponseWriter, r *http.Request) {
	h.reviewDriver(w, r, driverapp.ReviewApprove)
}

// RejectDriverVerification handles POST /api/v1/admin/verifications/drivers/{driverID}/reject.
func (h *AdminKYCHandler) RejectDriverVerification(w http.ResponseWriter, r *http.Request) {
	h.reviewDriver(w, r, driverapp.ReviewReject)
}

func (h *AdminKYCHandler) reviewDriver(w http.ResponseWriter, r *http.Request, action driverapp.ReviewAction) {
	if h.unavailable(w) {
		return
	}
	driverID := r.PathValue("driverID")
	if driverID == "" {
		writeBadRequest(w, "driverID is required")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req reviewRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // reason only required for reject; entity validates

	v, err := h.reviewDriverVerification.Execute(r.Context(), driverapp.ReviewDriverVerificationInput{
		DriverID: driverID, Reviewer: claims.UserID, Action: action, Reason: req.Reason,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, driverVerificationJSON(v))
}

// ─── Vehicle verification review ───────────────────────────────────────────────

// ListVehicleVerifications handles
// GET /api/v1/admin/verifications/vehicles?status=&vehicle_type=&service_type=&limit=&sort=
// — sort=expiry orders by nearest upcoming document expiry (Phần 12).
func (h *AdminKYCHandler) ListVehicleVerifications(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	status := driverentity.KYCStatus(r.URL.Query().Get("status"))
	if status == "" {
		status = driverentity.KYCPending
	}
	vehicleType := driverentity.VehicleType(r.URL.Query().Get("vehicle_type"))
	serviceType := driverentity.ServiceType(r.URL.Query().Get("service_type"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	sortByExpiry := r.URL.Query().Get("sort") == "expiry"

	list, err := h.listVehicleVerifications.Execute(r.Context(), status, vehicleType, serviceType, limit, sortByExpiry)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(list))
	for i, v := range list {
		items[i] = vehicleVerificationJSON(v)
	}
	writeJSON(w, http.StatusOK, map[string]any{"verifications": items})
}

// ApproveVehicleVerification handles POST /api/v1/admin/verifications/vehicles/{driverID}/approve.
func (h *AdminKYCHandler) ApproveVehicleVerification(w http.ResponseWriter, r *http.Request) {
	h.reviewVehicle(w, r, driverapp.ReviewApprove)
}

// RejectVehicleVerification handles POST /api/v1/admin/verifications/vehicles/{driverID}/reject.
func (h *AdminKYCHandler) RejectVehicleVerification(w http.ResponseWriter, r *http.Request) {
	h.reviewVehicle(w, r, driverapp.ReviewReject)
}

func (h *AdminKYCHandler) reviewVehicle(w http.ResponseWriter, r *http.Request, action driverapp.ReviewAction) {
	if h.unavailable(w) {
		return
	}
	driverID := r.PathValue("driverID")
	if driverID == "" {
		writeBadRequest(w, "driverID is required")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req reviewRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	v, err := h.reviewVehicleVerification.Execute(r.Context(), driverapp.ReviewVehicleVerificationInput{
		DriverID: driverID, Reviewer: claims.UserID, Action: action, Reason: req.Reason,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vehicleVerificationJSON(v))
}

// ─── Combined detail (Phần 2) ───────────────────────────────────────────────────

// GetDriverKYCDetail handles GET /api/v1/admin/verifications/drivers/{driverID}/detail
// — everything the review Drawer/Dialog needs in one call: driver info,
// vehicle info, the document checklist (with document_id per uploaded item,
// for fetching bytes via GetDocument/fullscreen viewer), and the full audit
// timeline (Phần 9). Any one piece being not-yet-submitted is not an error
// for the other pieces — each is simply omitted (null) from the response.
func (h *AdminKYCHandler) GetDriverKYCDetail(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID := r.PathValue("driverID")
	if driverID == "" {
		writeBadRequest(w, "driverID is required")
		return
	}

	resp := map[string]any{"driver_id": driverID, "phone": h.phoneForDriver(r.Context(), driverID)}

	if dv, err := h.getDriverVerification.Execute(r.Context(), driverID); err == nil && dv != nil {
		resp["driver_verification"] = driverVerificationJSON(dv)
	}
	if vv, err := h.getVehicleVerification.Execute(r.Context(), driverID); err == nil && vv != nil {
		resp["vehicle_verification"] = vehicleVerificationJSON(vv)
	}
	if uploaded, err := h.listDocuments.Execute(r.Context(), driverID); err == nil {
		byType := make(map[driverentity.DocumentType]*driverentity.KYCDocument, len(uploaded))
		for _, d := range uploaded {
			byType[d.DocumentType] = d
		}
		items := make([]map[string]any, 0, len(allDocumentTypes))
		for _, dt := range allDocumentTypes {
			if d, ok := byType[dt]; ok {
				items = append(items, documentChecklistItemJSONForAdmin(d))
			} else {
				items = append(items, map[string]any{"document_type": string(dt), "uploaded": false})
			}
		}
		resp["documents"] = items
	}
	if h.listAuditLogs != nil {
		if logs, err := h.listAuditLogs.Execute(r.Context(), driverID, 0); err == nil {
			items := make([]map[string]any, len(logs))
			for i, l := range logs {
				items[i] = auditLogJSON(l)
			}
			resp["audit_log"] = items
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func auditLogJSON(l *driverentity.AuditLog) map[string]any {
	return map[string]any{
		"id":          l.ID,
		"entity_type": string(l.EntityType),
		"action":      string(l.Action),
		"actor_id":    l.ActorID,
		"reason":      l.Reason,
		"created_at":  l.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// ─── Dashboard summary (Phần 10) ────────────────────────────────────────────────

// GetKYCSummary handles GET /api/v1/admin/verifications/summary — the 4
// dashboard cards (Pending/Approved/Rejected/Expired), driver-verification
// counts only (vehicle verification tracks the same lifecycle per driver,
// so the driver-verification count is the meaningful "how many drivers need
// attention" figure for an MVP dashboard).
func (h *AdminKYCHandler) GetKYCSummary(w http.ResponseWriter, r *http.Request) {
	if h.getSummary == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "admin KYC service not configured"})
		return
	}
	summary, err := h.getSummary.Execute(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"pending":  summary.Pending,
		"approved": summary.Approved,
		"rejected": summary.Rejected,
		"expired":  summary.Expired,
	})
}

// ─── Document review ────────────────────────────────────────────────────────

// ListDriverDocuments handles GET /api/v1/admin/verifications/drivers/{driverID}/documents
// — same checklist shape as KYCHandler.ListDocuments but for any driverID
// (admin only), plus each uploaded entry's opaque document id so the admin
// UI can fetch its bytes via GetDocument.
func (h *AdminKYCHandler) ListDriverDocuments(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID := r.PathValue("driverID")
	if driverID == "" {
		writeBadRequest(w, "driverID is required")
		return
	}
	uploaded, err := h.listDocuments.Execute(r.Context(), driverID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	byType := make(map[driverentity.DocumentType]*driverentity.KYCDocument, len(uploaded))
	for _, d := range uploaded {
		byType[d.DocumentType] = d
	}
	items := make([]map[string]any, 0, len(allDocumentTypes))
	for _, dt := range allDocumentTypes {
		if d, ok := byType[dt]; ok {
			items = append(items, documentChecklistItemJSONForAdmin(d))
		} else {
			items = append(items, map[string]any{"document_type": string(dt), "uploaded": false})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"documents": items})
}

// GetDocument handles GET /api/v1/admin/verifications/documents/{documentID}
// — streams the raw file bytes. Admin-only (Phần 9/13). The storage path
// itself is resolved server-side from the document's metadata and never
// appears in any response — the client only ever supplies/sees the opaque
// document id.
func (h *AdminKYCHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	id := r.PathValue("documentID")
	if id == "" {
		writeBadRequest(w, "documentID is required")
		return
	}
	doc, err := h.getDocument.Execute(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	f, err := h.documentStore.Open(doc.StoragePath)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	defer f.Close()

	if doc.ContentType != "" {
		w.Header().Set("Content-Type", doc.ContentType)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}

// DownloadDriverDocumentsZip handles
// GET /api/v1/admin/verifications/drivers/{driverID}/documents.zip — Phần 3's
// "Download ZIP" action. Bundles the latest uploaded version of every
// document type this driver has (never a historical version — this is for
// a quick offline review copy, not a full audit export). Best-effort per
// file: one unreadable file is skipped rather than failing the whole
// archive, since the admin still wants everything else that IS readable.
func (h *AdminKYCHandler) DownloadDriverDocumentsZip(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	driverID := r.PathValue("driverID")
	if driverID == "" {
		writeBadRequest(w, "driverID is required")
		return
	}
	uploaded, err := h.listDocuments.Execute(r.Context(), driverID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+driverID+`-kyc-documents.zip"`)
	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, d := range uploaded {
		f, err := h.documentStore.Open(d.StoragePath)
		if err != nil {
			continue
		}
		ext := path.Ext(d.StoragePath)
		entry, err := zw.Create(string(d.DocumentType) + "_v" + strconv.Itoa(d.Version) + ext)
		if err == nil {
			_, _ = io.Copy(entry, f)
		}
		f.Close()
	}
}

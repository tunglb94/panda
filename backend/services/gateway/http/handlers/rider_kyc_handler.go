package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/fairride/gateway/http/middleware"
	userapp "github.com/fairride/user/app"
	userentity "github.com/fairride/user/domain/entity"
)

const riderKYCDobLayout = "2006-01-02"

// RiderKYCHandler implements the Rider KYC endpoints — a deliberately
// smaller counterpart to KYCHandler (driver side): one draft record, two
// photos, three text fields, no versioning/expiry.
type RiderKYCHandler struct {
	uploadDocument *userapp.UploadRiderDocumentUseCase
	submit         *userapp.SubmitRiderVerificationUseCase
	get            *userapp.GetRiderVerificationUseCase
}

func NewRiderKYCHandler(
	uploadDocument *userapp.UploadRiderDocumentUseCase,
	submit *userapp.SubmitRiderVerificationUseCase,
	get *userapp.GetRiderVerificationUseCase,
) *RiderKYCHandler {
	return &RiderKYCHandler{uploadDocument: uploadDocument, submit: submit, get: get}
}

func (h *RiderKYCHandler) unavailable(w http.ResponseWriter) bool {
	if h.uploadDocument == nil || h.submit == nil || h.get == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "rider kyc service not configured"})
		return true
	}
	return false
}

// UploadDocument handles POST /api/v1/rider/verification/documents
// (multipart/form-data: field "document_type" = cccd_front|cccd_back, file field "file").
func (h *RiderKYCHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeBadRequest(w, "invalid multipart form")
		return
	}
	docType := userapp.RiderDocumentType(r.FormValue("document_type"))
	file, header, err := r.FormFile("file")
	if err != nil {
		writeBadRequest(w, "file is required")
		return
	}
	defer file.Close()

	v, err := h.uploadDocument.Execute(r.Context(), userapp.UploadRiderDocumentInput{
		UserID:       claims.UserID,
		DocumentType: docType,
		Ext:          filepath.Ext(header.Filename),
		Data:         file,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, riderVerificationJSON(v))
}

type submitRiderVerificationRequest struct {
	FullName         string `json:"full_name"`
	DateOfBirth      string `json:"date_of_birth"`
	NationalIDNumber string `json:"national_id_number"`
}

// SubmitVerification handles POST /api/v1/rider/verification.
func (h *RiderKYCHandler) SubmitVerification(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req submitRiderVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	dob, err := time.Parse(riderKYCDobLayout, strings.TrimSpace(req.DateOfBirth))
	if err != nil {
		writeBadRequest(w, "date_of_birth must be in YYYY-MM-DD format")
		return
	}

	v, err := h.submit.Execute(r.Context(), userapp.SubmitRiderVerificationInput{
		UserID:           claims.UserID,
		FullName:         req.FullName,
		DateOfBirth:      dob,
		NationalIDNumber: req.NationalIDNumber,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, riderVerificationJSON(v))
}

// GetVerification handles GET /api/v1/rider/verification.
func (h *RiderKYCHandler) GetVerification(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	v, err := h.get.Execute(r.Context(), claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, riderVerificationJSON(v))
}

func riderVerificationJSON(v *userentity.RiderVerification) map[string]any {
	body := map[string]any{
		"id":                  v.ID,
		"user_id":             v.UserID,
		"full_name":           v.FullName,
		"national_id_number":  v.NationalIDNumber,
		"status":              string(v.Status),
		"cccd_front_uploaded": v.CCCDFrontPath != "",
		"cccd_back_uploaded":  v.CCCDBackPath != "",
		"reject_reason":       v.RejectReason,
		// review_mode is the AI pipeline's own recommendation (informational
		// — Status remains authoritative). Raw ocr_result/vision_result JSON
		// blobs are intentionally not exposed here; see riderVerificationAdminJSON.
		"review_mode": string(v.ReviewMode),
	}
	if !v.DateOfBirth.IsZero() {
		body["date_of_birth"] = v.DateOfBirth.Format(riderKYCDobLayout)
	}
	if v.SubmittedAt != nil {
		body["submitted_at"] = v.SubmittedAt.UTC().Format(time.RFC3339)
	}
	if v.ApprovedAt != nil {
		body["approved_at"] = v.ApprovedAt.UTC().Format(time.RFC3339)
	}
	if v.RejectedAt != nil {
		body["rejected_at"] = v.RejectedAt.UTC().Format(time.RFC3339)
	}
	return body
}

// riderVerificationAdminJSON extends riderVerificationJSON with the AI
// pipeline's full audit trail — only exposed to admins (ListPending), since
// the raw OCR/Vision provider output isn't meaningful to a rider.
func riderVerificationAdminJSON(v *userentity.RiderVerification) map[string]any {
	body := riderVerificationJSON(v)
	body["ai_confidence"] = v.AIConfidence
	body["ocr_result"] = v.OCRResult
	body["vision_result"] = v.VisionResult
	return body
}

// AdminRiderKYCHandler implements the RequireAdmin-gated review API — no
// dashboard UI in this phase (Admin app is out of scope), just the
// endpoints so a decision is operable (see plan's Known Gaps).
type AdminRiderKYCHandler struct {
	list   *userapp.ListPendingRiderVerificationsUseCase
	review *userapp.ReviewRiderVerificationUseCase
}

func NewAdminRiderKYCHandler(list *userapp.ListPendingRiderVerificationsUseCase, review *userapp.ReviewRiderVerificationUseCase) *AdminRiderKYCHandler {
	return &AdminRiderKYCHandler{list: list, review: review}
}

func (h *AdminRiderKYCHandler) unavailable(w http.ResponseWriter) bool {
	if h.list == nil || h.review == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "rider kyc admin service not configured"})
		return true
	}
	return false
}

// ListPending handles GET /api/v1/admin/verifications/riders.
func (h *AdminRiderKYCHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	items, err := h.list.Execute(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, v := range items {
		out = append(out, riderVerificationAdminJSON(v))
	}
	writeJSON(w, http.StatusOK, map[string]any{"riders": out})
}

type reviewRiderVerificationRequest struct {
	Reason string `json:"reason"`
}

// Approve handles POST /api/v1/admin/verifications/riders/{userID}/approve.
func (h *AdminRiderKYCHandler) Approve(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	userID := r.PathValue("userID")
	v, err := h.review.Approve(r.Context(), userID, claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, riderVerificationJSON(v))
}

// Reject handles POST /api/v1/admin/verifications/riders/{userID}/reject.
func (h *AdminRiderKYCHandler) Reject(w http.ResponseWriter, r *http.Request) {
	if h.unavailable(w) {
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req reviewRiderVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	userID := r.PathValue("userID")
	v, err := h.review.Reject(r.Context(), userID, claims.UserID, req.Reason)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, riderVerificationJSON(v))
}

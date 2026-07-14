// Package entity — RiderVerification is the Rider KYC record: a
// deliberately minimal counterpart to the driver service's
// DriverVerification, covering only the fields the spec calls for (Họ tên,
// Ngày sinh, CCCD, ảnh CCCD trước/sau) with no document versioning/expiry —
// riders re-upload in place, there is no regulatory expiry on a rider's ID
// photo the way there is on a driver's GPLX/vehicle documents.
package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// RiderKYCStatus is the review lifecycle for a RiderVerification. Approval
// is manual (RequireAdmin-gated API, no dashboard in this phase — see the
// plan's Known Gaps) so a submitted record stays Pending until an admin
// calls Approve/Reject.
type RiderKYCStatus string

const (
	RiderKYCPending  RiderKYCStatus = "pending"
	RiderKYCApproved RiderKYCStatus = "approved"
	RiderKYCRejected RiderKYCStatus = "rejected"
)

// ReviewMode is the AI pipeline's own recommendation, recorded as an audit
// trail — see ApplyPipelineDecision. Independent of Status: an admin's
// later Approve/Reject through the existing RequireAdmin API can still
// override whatever the pipeline recommended; ReviewMode just records what
// the pipeline said at Submit time.
type ReviewMode string

const (
	ReviewModeAutoApproved ReviewMode = "auto_approved"
	ReviewModeManualReview ReviewMode = "manual_review"
	ReviewModeRejected     ReviewMode = "rejected"
)

// pipelineReviewer is recorded as Reviewer when ApplyPipelineDecision
// itself transitions Status (AutoApproved/Rejected) — distinguishes an
// automated decision from a real admin's ID in audit trails/logs.
const pipelineReviewer = "system:ai-pipeline"

// RiderVerification tracks one rider's identity verification. A row is
// created as a draft (empty text fields, Pending) the moment the rider
// uploads their first CCCD photo — see NewDraftRiderVerification — and is
// filled in by Submit once both photos and all three text fields are present.
type RiderVerification struct {
	ID               string
	UserID           string
	FullName         string
	DateOfBirth      time.Time
	NationalIDNumber string
	CCCDFrontPath    string
	CCCDBackPath     string
	Status           RiderKYCStatus
	SubmittedAt      *time.Time
	ApprovedAt       *time.Time
	RejectedAt       *time.Time
	Reviewer         string
	RejectReason     string
	ReviewMode       ReviewMode
	AIConfidence     float64
	OCRResult        string // raw JSON, opaque to this entity
	VisionResult     string // raw JSON, opaque to this entity
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewDraftRiderVerification creates an empty Pending draft — used the first
// time a rider uploads a CCCD photo, before any text field is known.
func NewDraftRiderVerification(id, userID string, now time.Time) (*RiderVerification, error) {
	if id == "" {
		return nil, errors.InvalidArgument("id must not be empty")
	}
	if userID == "" {
		return nil, errors.InvalidArgument("user_id must not be empty")
	}
	return &RiderVerification{
		ID:        id,
		UserID:    userID,
		Status:    RiderKYCPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ReconstituteRiderVerification rebuilds a RiderVerification from a persistence record. No validation.
func ReconstituteRiderVerification(
	id, userID, fullName string,
	dateOfBirth time.Time,
	nationalIDNumber, cccdFrontPath, cccdBackPath string,
	status RiderKYCStatus,
	submittedAt, approvedAt, rejectedAt *time.Time,
	reviewer, rejectReason string,
	reviewMode ReviewMode,
	aiConfidence float64,
	ocrResult, visionResult string,
	createdAt, updatedAt time.Time,
) *RiderVerification {
	return &RiderVerification{
		ID:               id,
		UserID:           userID,
		FullName:         fullName,
		DateOfBirth:      dateOfBirth,
		NationalIDNumber: nationalIDNumber,
		CCCDFrontPath:    cccdFrontPath,
		CCCDBackPath:     cccdBackPath,
		Status:           status,
		SubmittedAt:      submittedAt,
		ApprovedAt:       approvedAt,
		RejectedAt:       rejectedAt,
		Reviewer:         reviewer,
		RejectReason:     rejectReason,
		ReviewMode:       reviewMode,
		AIConfidence:     aiConfidence,
		OCRResult:        ocrResult,
		VisionResult:     visionResult,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

// AttachFrontPhoto/AttachBackPhoto record an uploaded CCCD photo's storage
// path. Re-uploading (front or back) while Rejected moves the record back
// to Pending — the rider is editing in response to a rejection.
func (v *RiderVerification) AttachFrontPhoto(path string, now time.Time) {
	v.CCCDFrontPath = path
	v.reopenIfRejected(now)
}

func (v *RiderVerification) AttachBackPhoto(path string, now time.Time) {
	v.CCCDBackPath = path
	v.reopenIfRejected(now)
}

func (v *RiderVerification) reopenIfRejected(now time.Time) {
	if v.Status != RiderKYCRejected {
		v.UpdatedAt = now
		return
	}
	v.Status = RiderKYCPending
	v.RejectedAt = nil
	v.RejectReason = ""
	v.UpdatedAt = now
}

// Submit fills in the text fields and moves the record to Pending review.
// Requires both CCCD photos to already be attached (Phần "Không cho submit
// thiếu file", mirrored from the driver KYC flow).
func (v *RiderVerification) Submit(fullName string, dateOfBirth time.Time, nationalIDNumber string, now time.Time) error {
	if strings.TrimSpace(fullName) == "" {
		return errors.InvalidArgument("full_name must not be empty")
	}
	if dateOfBirth.IsZero() || !dateOfBirth.Before(now) {
		return errors.InvalidArgument("date_of_birth must be a valid past date")
	}
	if strings.TrimSpace(nationalIDNumber) == "" {
		return errors.InvalidArgument("national_id_number must not be empty")
	}
	if v.CCCDFrontPath == "" || v.CCCDBackPath == "" {
		return errors.InvalidArgument("please upload both CCCD photos before submitting")
	}
	v.FullName = strings.TrimSpace(fullName)
	v.DateOfBirth = dateOfBirth
	v.NationalIDNumber = strings.TrimSpace(nationalIDNumber)
	v.Status = RiderKYCPending
	t := now
	v.SubmittedAt = &t
	v.RejectedAt = nil
	v.RejectReason = ""
	v.UpdatedAt = now
	return nil
}

// Approve transitions Pending -> Approved.
func (v *RiderVerification) Approve(reviewer string, now time.Time) error {
	if v.Status != RiderKYCPending {
		return errors.PreconditionFailed("rider verification cannot be approved from status: " + string(v.Status))
	}
	v.Status = RiderKYCApproved
	t := now
	v.ApprovedAt = &t
	v.Reviewer = reviewer
	v.RejectReason = ""
	v.UpdatedAt = now
	return nil
}

// Reject transitions Pending -> Rejected. reason is required so the rider
// can see why and correct it.
func (v *RiderVerification) Reject(reviewer, reason string, now time.Time) error {
	if v.Status != RiderKYCPending {
		return errors.PreconditionFailed("rider verification cannot be rejected from status: " + string(v.Status))
	}
	if strings.TrimSpace(reason) == "" {
		return errors.InvalidArgument("reject_reason must not be empty")
	}
	v.Status = RiderKYCRejected
	t := now
	v.RejectedAt = &t
	v.Reviewer = reviewer
	v.RejectReason = strings.TrimSpace(reason)
	v.UpdatedAt = now
	return nil
}

func (v *RiderVerification) IsApproved() bool { return v.Status == RiderKYCApproved }

// ApplyPipelineDecision records the AI pipeline's Upload -> Rule Engine ->
// OCR -> Vision -> Decision outcome (see user/app/kyc_decision.go), called
// once right after Submit while Status is still Pending. mode drives what
// happens to Status:
//   - ReviewModeAutoApproved: transitions straight to Approved, no admin
//     needed — only reachable once real OCR/Vision confidence crosses the
//     auto-approve threshold; the mock providers wired up today never do.
//   - ReviewModeRejected: transitions straight to Rejected with reason.
//   - ReviewModeManualReview: Status is left as Pending — the existing
//     RequireAdmin Approve/Reject API is unchanged and is what resolves it.
//
// Only valid to call while Status == Pending (i.e. immediately after
// Submit) — returns CodePreconditionFailed otherwise, since re-running the
// pipeline against an already-decided record would silently override a
// human admin's prior call.
func (v *RiderVerification) ApplyPipelineDecision(mode ReviewMode, confidence float64, ocrResult, visionResult, reason string, now time.Time) error {
	if v.Status != RiderKYCPending {
		return errors.PreconditionFailed("pipeline decision can only be applied to a pending verification, current status: " + string(v.Status))
	}
	v.ReviewMode = mode
	v.AIConfidence = confidence
	v.OCRResult = ocrResult
	v.VisionResult = visionResult
	v.UpdatedAt = now

	switch mode {
	case ReviewModeAutoApproved:
		return v.Approve(pipelineReviewer, now)
	case ReviewModeRejected:
		return v.Reject(pipelineReviewer, reason, now)
	default:
		return nil // ReviewModeManualReview — stays Pending for a human admin.
	}
}

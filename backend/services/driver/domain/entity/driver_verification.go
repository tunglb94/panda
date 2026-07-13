package entity

import (
	"strings"
	"time"

	"github.com/fairride/shared/errors"
)

// DriverVerification is the KYC record for a driver's personal identity —
// CCCD front/back, selfie, and (when the driver wants Ride capability)
// GPLX/license number. One per driver (driver_id unique). Separate from
// VehicleVerification, which covers the vehicle itself — a driver's
// identity can be Approved while their vehicle is still Pending, and vice
// versa; the Online Guard requires both Approved (see
// driver/app/availability.go).
//
// NationalIDNumber (Phần 5 — Duplicate Detection) is the CCCD's printed ID
// number — a manually-entered text field, not OCR-extracted (OCR is out of
// scope for this phase); it's the only way to compare two CCCDs for
// duplicates since the uploaded images themselves are opaque files.
type DriverVerification struct {
	ID               string
	DriverID         string
	FullName         string
	DateOfBirth      time.Time
	Address          string
	NationalIDNumber string // CCCD number
	LicenseNumber    string // GPLX number; may be empty until Ride capability is requested
	Status           KYCStatus
	SubmittedAt      time.Time
	ApprovedAt       *time.Time
	RejectedAt       *time.Time
	ExpiredAt        *time.Time
	Reviewer         string
	RejectReason     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewDriverVerification creates a validated DriverVerification in
// KYCPending. Documents (CCCD front/back, selfie) are validated as already
// uploaded by the use case layer (SubmitDriverVerificationUseCase), not
// here — this entity only owns the text fields.
func NewDriverVerification(id, driverID, fullName string, dateOfBirth time.Time, address, nationalIDNumber, licenseNumber string, now time.Time) (*DriverVerification, error) {
	if id == "" {
		return nil, errors.InvalidArgument("id must not be empty")
	}
	if driverID == "" {
		return nil, errors.InvalidArgument("driver_id must not be empty")
	}
	if strings.TrimSpace(fullName) == "" {
		return nil, errors.InvalidArgument("full_name must not be empty")
	}
	if dateOfBirth.IsZero() {
		return nil, errors.InvalidArgument("date_of_birth is required")
	}
	if !dateOfBirth.Before(now) {
		return nil, errors.InvalidArgument("date_of_birth must be in the past")
	}
	if strings.TrimSpace(address) == "" {
		return nil, errors.InvalidArgument("address must not be empty")
	}
	if strings.TrimSpace(nationalIDNumber) == "" {
		return nil, errors.InvalidArgument("national_id_number must not be empty")
	}
	return &DriverVerification{
		ID:               id,
		DriverID:         driverID,
		FullName:         strings.TrimSpace(fullName),
		DateOfBirth:      dateOfBirth,
		Address:          strings.TrimSpace(address),
		NationalIDNumber: strings.TrimSpace(nationalIDNumber),
		LicenseNumber:    strings.TrimSpace(licenseNumber),
		Status:           KYCPending,
		SubmittedAt:      now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// ReconstituteDriverVerification rebuilds a DriverVerification from a persistence record. No validation.
func ReconstituteDriverVerification(
	id, driverID, fullName string,
	dateOfBirth time.Time,
	address, nationalIDNumber, licenseNumber string,
	status KYCStatus,
	submittedAt time.Time,
	approvedAt, rejectedAt, expiredAt *time.Time,
	reviewer, rejectReason string,
	createdAt, updatedAt time.Time,
) *DriverVerification {
	return &DriverVerification{
		ID:               id,
		DriverID:         driverID,
		FullName:         fullName,
		DateOfBirth:      dateOfBirth,
		Address:          address,
		NationalIDNumber: nationalIDNumber,
		LicenseNumber:    licenseNumber,
		Status:           status,
		SubmittedAt:      submittedAt,
		ApprovedAt:       approvedAt,
		RejectedAt:       rejectedAt,
		ExpiredAt:        expiredAt,
		Reviewer:         reviewer,
		RejectReason:     rejectReason,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

// Resubmit lets a driver edit and resubmit a Pending or Rejected
// verification — moves (back) to KYCPending and clears any prior rejection.
// Returns CodePreconditionFailed while UnderReview (an admin is already
// looking at it). To edit an Approved or Expired verification, callers must
// call Invalidate first (Phần 3 — Re-verification).
func (v *DriverVerification) Resubmit(fullName string, dateOfBirth time.Time, address, nationalIDNumber, licenseNumber string, now time.Time) error {
	if v.Status != KYCPending && v.Status != KYCRejected {
		return errors.PreconditionFailed("driver verification cannot be edited from status: " + string(v.Status))
	}
	if strings.TrimSpace(fullName) == "" {
		return errors.InvalidArgument("full_name must not be empty")
	}
	if dateOfBirth.IsZero() || !dateOfBirth.Before(now) {
		return errors.InvalidArgument("date_of_birth must be a valid past date")
	}
	if strings.TrimSpace(address) == "" {
		return errors.InvalidArgument("address must not be empty")
	}
	if strings.TrimSpace(nationalIDNumber) == "" {
		return errors.InvalidArgument("national_id_number must not be empty")
	}
	v.FullName = strings.TrimSpace(fullName)
	v.DateOfBirth = dateOfBirth
	v.Address = strings.TrimSpace(address)
	v.NationalIDNumber = strings.TrimSpace(nationalIDNumber)
	v.LicenseNumber = strings.TrimSpace(licenseNumber)
	v.Status = KYCPending
	v.SubmittedAt = now
	v.RejectedAt = nil
	v.RejectReason = ""
	v.UpdatedAt = now
	return nil
}

// Invalidate resets an Approved or Expired verification back to Pending —
// Phần 3 (Re-verification): "Không được giữ trạng thái Approved" whenever
// the driver changes personal info or re-uploads a document (CCCD/selfie).
// A no-op from any other status.
func (v *DriverVerification) Invalidate(now time.Time) {
	if v.Status != KYCApproved && v.Status != KYCExpired {
		return
	}
	v.Status = KYCPending
	v.SubmittedAt = now
	v.ApprovedAt = nil
	v.RejectedAt = nil
	v.ExpiredAt = nil
	v.RejectReason = ""
	v.UpdatedAt = now
}

// StartReview transitions Pending -> UnderReview.
func (v *DriverVerification) StartReview(reviewer string, now time.Time) error {
	if v.Status != KYCPending {
		return errors.PreconditionFailed("only pending verifications can start review")
	}
	v.Status = KYCUnderReview
	v.Reviewer = reviewer
	v.UpdatedAt = now
	return nil
}

// Approve transitions Pending/UnderReview -> Approved.
func (v *DriverVerification) Approve(reviewer string, now time.Time) error {
	if v.Status != KYCPending && v.Status != KYCUnderReview {
		return errors.PreconditionFailed("driver verification cannot be approved from status: " + string(v.Status))
	}
	v.Status = KYCApproved
	v.ApprovedAt = &now
	v.ExpiredAt = nil
	v.Reviewer = reviewer
	v.RejectReason = ""
	v.UpdatedAt = now
	return nil
}

// Reject transitions Pending/UnderReview -> Rejected. reason is required
// (Phần 6 — the driver must see why).
func (v *DriverVerification) Reject(reviewer, reason string, now time.Time) error {
	if v.Status != KYCPending && v.Status != KYCUnderReview {
		return errors.PreconditionFailed("driver verification cannot be rejected from status: " + string(v.Status))
	}
	if strings.TrimSpace(reason) == "" {
		return errors.InvalidArgument("reject_reason must not be empty")
	}
	v.Status = KYCRejected
	v.RejectedAt = &now
	v.Reviewer = reviewer
	v.RejectReason = strings.TrimSpace(reason)
	v.UpdatedAt = now
	return nil
}

// Expire transitions Approved -> Expired. Symmetric with
// VehicleVerification.Expire — DriverVerification's own documents (CCCD,
// selfie) currently carry no expiry date (Phần 2 only lists GPLX/vehicle
// registration/insurance/inspection), so nothing calls this today, but the
// entity supports it uniformly in case a future document type needs it.
func (v *DriverVerification) Expire(reason string, now time.Time) error {
	if v.Status != KYCApproved {
		return errors.PreconditionFailed("only an approved driver verification can expire, current status: " + string(v.Status))
	}
	v.Status = KYCExpired
	v.ExpiredAt = &now
	v.RejectReason = strings.TrimSpace(reason)
	v.UpdatedAt = now
	return nil
}

func (v *DriverVerification) IsApproved() bool { return v.Status == KYCApproved }

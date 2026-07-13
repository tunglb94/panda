package entity

import "github.com/fairride/shared/errors"

// KYCStatus is the review lifecycle shared by DriverVerification and
// VehicleVerification. Deliberately a distinct type from the existing
// VerificationStatus (pending/verified/rejected/suspended, on DriverProfile
// itself) — that older field is a coarse legacy toggle this refactor does
// not touch; KYCStatus is the richer, document-backed, admin-reviewed
// status that actually gates Online (see driver/app/availability.go).
type KYCStatus string

const (
	KYCPending     KYCStatus = "pending"
	KYCUnderReview KYCStatus = "under_review"
	KYCApproved    KYCStatus = "approved"
	KYCRejected    KYCStatus = "rejected"
	KYCExpired     KYCStatus = "expired"
)

func validateKYCStatus(s KYCStatus) error {
	switch s {
	case KYCPending, KYCUnderReview, KYCApproved, KYCRejected, KYCExpired:
		return nil
	default:
		return errors.InvalidArgument("unknown KYC status: " + string(s))
	}
}

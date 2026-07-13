package entity

import (
	"time"

	"github.com/fairride/shared/errors"
)

// DocumentType enumerates every KYC document Panda currently collects.
// Shared by both DriverVerification (identity docs) and VehicleVerification
// (vehicle docs) — a single kyc_documents table, not two.
type DocumentType string

const (
	DocumentCCCDFront           DocumentType = "cccd_front"
	DocumentCCCDBack            DocumentType = "cccd_back"
	DocumentSelfie              DocumentType = "selfie"
	DocumentLicense             DocumentType = "license" // GPLX
	DocumentVehicleRegistration DocumentType = "vehicle_registration"
	DocumentVehicleInsurance    DocumentType = "vehicle_insurance"
	// DocumentVehicleInspection is "Đăng kiểm" (Phần 2 of the Hardening
	// spec) — optional (not in VehicleDocumentTypes below, so submission
	// never requires it), but when uploaded, its expiry is tracked and
	// checked by the Online Guard the same as registration/insurance.
	DocumentVehicleInspection DocumentType = "vehicle_inspection"
)

var validDocumentTypes = map[DocumentType]bool{
	DocumentCCCDFront:           true,
	DocumentCCCDBack:            true,
	DocumentSelfie:              true,
	DocumentLicense:             true,
	DocumentVehicleRegistration: true,
	DocumentVehicleInsurance:    true,
	DocumentVehicleInspection:   true,
}

// ValidateDocumentType reports whether t is a known document type.
func ValidateDocumentType(t DocumentType) error {
	if !validDocumentTypes[t] {
		return errors.InvalidArgument("unknown document type: " + string(t))
	}
	return nil
}

// ExpiringDocumentTypes are the document types Phần 2 requires an expiry
// date for (GPLX, Đăng ký xe, Bảo hiểm, Đăng kiểm) — CCCD/Selfie never
// expire in this model. The Online Guard and the upload endpoint both
// consult this to decide whether expires_at is meaningful for a given type.
var ExpiringDocumentTypes = map[DocumentType]bool{
	DocumentLicense:             true,
	DocumentVehicleRegistration: true,
	DocumentVehicleInsurance:    true,
	DocumentVehicleInspection:   true,
}

// DriverDocumentTypes are the documents SubmitDriverVerificationUseCase
// requires to already be uploaded before it will accept a submission
// (Phần 10 — "Không cho submit thiếu file"). GPLX/license is required
// separately, only when RideEnabled (Phần 2 — Delivery doesn't need it),
// see RideLicenseDocumentTypes.
var DriverDocumentTypes = []DocumentType{DocumentCCCDFront, DocumentCCCDBack, DocumentSelfie}

// RideLicenseDocumentTypes must be uploaded before a vehicle verification
// requesting RideEnabled can be submitted.
var RideLicenseDocumentTypes = []DocumentType{DocumentLicense}

// VehicleDocumentTypes must be uploaded before SubmitVehicleVerificationUseCase
// will accept a submission. Vehicle inspection is deliberately excluded —
// it's optional (see DocumentVehicleInspection's doc comment).
var VehicleDocumentTypes = []DocumentType{DocumentVehicleRegistration, DocumentVehicleInsurance}

// KYCDocument is one uploaded file's metadata. StoragePath is an internal,
// local-disk-relative reference — never serialized in any API response
// (Phần 9/13 — "Không expose file path"); only DocumentType/UploadedAt/
// Version/ExpiresAt are surfaced to clients, and only an opaque ID is used
// to fetch bytes back (admin document-review endpoint only).
//
// Phần 4 — Document Versioning: each upload creates a NEW row (Version =
// previous max + 1) rather than overwriting the previous one — the
// repository never issues an UPDATE against an existing document row, and
// the underlying file on disk is never deleted or overwritten either (see
// localstore.DocumentStore, which already names every saved file with a
// fresh UUID). UploadedBy records who performed the upload (always the
// driver themselves today, since the only upload endpoint is driver-facing
// and always writes claims.UserID — recorded explicitly so a future
// admin-assisted upload path doesn't need a schema change).
type KYCDocument struct {
	ID           string
	DriverID     string
	DocumentType DocumentType
	StoragePath  string
	ContentType  string
	Version      int
	ExpiresAt    *time.Time
	UploadedBy   string
	UploadedAt   time.Time
}

func NewKYCDocument(id, driverID string, docType DocumentType, storagePath, contentType string, version int, expiresAt *time.Time, uploadedBy string, now time.Time) (*KYCDocument, error) {
	if id == "" || driverID == "" {
		return nil, errors.InvalidArgument("id and driver_id are required")
	}
	if err := ValidateDocumentType(docType); err != nil {
		return nil, err
	}
	if storagePath == "" {
		return nil, errors.InvalidArgument("storage_path must not be empty")
	}
	if version < 1 {
		version = 1
	}
	return &KYCDocument{
		ID:           id,
		DriverID:     driverID,
		DocumentType: docType,
		StoragePath:  storagePath,
		ContentType:  contentType,
		Version:      version,
		ExpiresAt:    expiresAt,
		UploadedBy:   uploadedBy,
		UploadedAt:   now,
	}, nil
}

// ReconstituteKYCDocument rebuilds a KYCDocument from a persistence record. No validation.
func ReconstituteKYCDocument(id, driverID string, docType DocumentType, storagePath, contentType string, version int, expiresAt *time.Time, uploadedBy string, uploadedAt time.Time) *KYCDocument {
	return &KYCDocument{
		ID:           id,
		DriverID:     driverID,
		DocumentType: docType,
		StoragePath:  storagePath,
		ContentType:  contentType,
		Version:      version,
		ExpiresAt:    expiresAt,
		UploadedBy:   uploadedBy,
		UploadedAt:   uploadedAt,
	}
}

// IsExpired reports whether this document's ExpiresAt has passed as of now.
// Documents with no ExpiresAt (nil — CCCD/Selfie, or an expiry-eligible
// type uploaded without one) never expire.
func (d *KYCDocument) IsExpired(now time.Time) bool {
	return d.ExpiresAt != nil && d.ExpiresAt.Before(now)
}

// ExpiresWithin reports whether this document expires within the given
// window from now (Phần 11 — "gần hết hạn 30 ngày -> banner vàng"), but has
// not already expired.
func (d *KYCDocument) ExpiresWithin(window time.Duration, now time.Time) bool {
	if d.ExpiresAt == nil {
		return false
	}
	return !d.IsExpired(now) && d.ExpiresAt.Before(now.Add(window))
}

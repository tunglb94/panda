package repository

import (
	"context"

	"github.com/fairride/driver/domain/entity"
)

// DriverVerificationRepository persists DriverVerification aggregates.
// All methods return *errors.DomainError on failure.
type DriverVerificationRepository interface {
	Save(ctx context.Context, v *entity.DriverVerification) error
	FindByDriverID(ctx context.Context, driverID string) (*entity.DriverVerification, error)
	// FindByNationalIDNumber returns CodeNotFound if no verification uses
	// this CCCD number — enforces CCCD uniqueness across drivers (Phần 5).
	FindByNationalIDNumber(ctx context.Context, nationalIDNumber string) (*entity.DriverVerification, error)
	// FindByLicenseNumber returns CodeNotFound if no verification uses this
	// GPLX number — enforces license number uniqueness across drivers (Phần 5).
	FindByLicenseNumber(ctx context.Context, licenseNumber string) (*entity.DriverVerification, error)
	// ListByStatus lists verifications in a given status, newest submission
	// first, capped at limit — used by the admin review dashboard.
	ListByStatus(ctx context.Context, status entity.KYCStatus, limit int) ([]*entity.DriverVerification, error)
	// CountByStatus returns how many verifications are in a given status —
	// the admin dashboard's 4 summary cards (Phần 10), without paging
	// through full rows just to count them.
	CountByStatus(ctx context.Context, status entity.KYCStatus) (int, error)
}

// VehicleVerificationRepository persists VehicleVerification aggregates.
type VehicleVerificationRepository interface {
	Save(ctx context.Context, v *entity.VehicleVerification) error
	FindByDriverID(ctx context.Context, driverID string) (*entity.VehicleVerification, error)
	// FindByPlateNumber returns CodeNotFound if no verification uses this
	// plate — used to enforce plate uniqueness (Phần 5) across drivers.
	FindByPlateNumber(ctx context.Context, plateNumber string) (*entity.VehicleVerification, error)
	// FindByVIN/FindByEngineNumber/FindByChassisNumber return CodeNotFound
	// if no verification uses that value — enforce vehicle-identity
	// uniqueness (Phần 6) whenever a driver supplies one (all three fields
	// are optional; empty values are never checked).
	FindByVIN(ctx context.Context, vin string) (*entity.VehicleVerification, error)
	FindByEngineNumber(ctx context.Context, engineNumber string) (*entity.VehicleVerification, error)
	FindByChassisNumber(ctx context.Context, chassisNumber string) (*entity.VehicleVerification, error)
	// ListByFilter lists verifications matching status/vehicleType/serviceType
	// (each optional — empty string/"" means "any"), newest submission
	// first, capped at limit — admin review dashboard filters.
	ListByFilter(ctx context.Context, status entity.KYCStatus, vehicleType entity.VehicleType, serviceType entity.ServiceType, limit int) ([]*entity.VehicleVerification, error)
	// ListByFilterSortedByExpiry is the same filter, but ordered by each
	// vehicle's nearest upcoming document expiry (registration/insurance/
	// inspection/license) ascending, NULLs (no expiry-tracked document
	// uploaded) last — Phần 12's admin dashboard "Sort: Expiry gần nhất".
	ListByFilterSortedByExpiry(ctx context.Context, status entity.KYCStatus, vehicleType entity.VehicleType, serviceType entity.ServiceType, limit int) ([]*entity.VehicleVerification, error)
}

// KYCDocumentRepository persists uploaded KYC document metadata (never the
// file bytes themselves — those live on local disk, see
// driver/infrastructure/localstore). Phần 4 — Save always INSERTs a new
// version; it never updates or deletes an existing row.
type KYCDocumentRepository interface {
	Save(ctx context.Context, d *entity.KYCDocument) error
	// FindByDriverAndType returns the LATEST version of driverID's docType
	// document, or CodeNotFound if none has ever been uploaded.
	FindByDriverAndType(ctx context.Context, driverID string, docType entity.DocumentType) (*entity.KYCDocument, error)
	// ListByDriverID returns the latest version of every document type
	// driverID has uploaded (one row per type, not full history).
	ListByDriverID(ctx context.Context, driverID string) ([]*entity.KYCDocument, error)
	// ListVersionsByDriverAndType returns every version ever uploaded for
	// (driverID, docType), newest first — Phần 4/11's version history.
	ListVersionsByDriverAndType(ctx context.Context, driverID string, docType entity.DocumentType) ([]*entity.KYCDocument, error)
	FindByID(ctx context.Context, id string) (*entity.KYCDocument, error)
}

// LicenseCapabilityRepository is the Rule Engine behind Phần 1 — which
// ServiceType a LicenseClass permits for Ride capability. Backed by the
// license_capabilities table so a change in Vietnamese law is a data
// UPDATE, never a Go code change or migration.
type LicenseCapabilityRepository interface {
	// IsAllowed reports whether licenseClass permits serviceType. An
	// unknown (licenseClass, serviceType) pair is treated as not allowed
	// (deny-by-default), matching the old hardcoded map's behavior.
	IsAllowed(ctx context.Context, licenseClass entity.LicenseClass, serviceType entity.ServiceType) (bool, error)
}

// AuditLogRepository persists AuditLog entries (Phần 7). Append-only — no
// Update/Delete method exists on this interface by design.
type AuditLogRepository interface {
	Save(ctx context.Context, log *entity.AuditLog) error
	// ListByDriverID returns every audit entry for driverID, newest first —
	// available for a future admin audit-trail view (Known Gap: no UI
	// consumes this yet, but the data is captured from day one).
	ListByDriverID(ctx context.Context, driverID string, limit int) ([]*entity.AuditLog, error)
}

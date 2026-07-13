package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// SubmitVehicleVerificationInput carries the vehicle-info fields collected
// in Steps 4/5 of the Become Driver wizard. VIN/EngineNumber/ChassisNumber
// (Phần 6) are optional — validated for duplicates only when non-empty.
type SubmitVehicleVerificationInput struct {
	DriverID        string
	VehicleType     entity.VehicleType
	ServiceType     entity.ServiceType
	Brand           string
	Model           string
	Year            int
	Color           string
	PlateNumber     string
	VIN             string
	EngineNumber    string
	ChassisNumber   string
	LicenseClass    entity.LicenseClass
	RideEnabled     bool
	DeliveryEnabled bool
}

// checkVehicleDedup enforces Phần 5 (plate) and Phần 6 (VIN/engine/chassis,
// only when supplied) uniqueness across drivers/vehicles, with clear
// Vietnamese error messages.
func checkVehicleDedup(ctx context.Context, repo repository.VehicleVerificationRepository, driverID string, in SubmitVehicleVerificationInput) error {
	if existing, err := repo.FindByPlateNumber(ctx, in.PlateNumber); err == nil {
		if existing.DriverID != driverID {
			return domainerrors.AlreadyExists("biển số đã được đăng ký bởi một tài xế khác")
		}
	} else if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
		return err
	}
	if in.VIN != "" {
		if existing, err := repo.FindByVIN(ctx, in.VIN); err == nil {
			if existing.DriverID != driverID {
				return domainerrors.AlreadyExists("số VIN đã được đăng ký cho một xe khác")
			}
		} else if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return err
		}
	}
	if in.EngineNumber != "" {
		if existing, err := repo.FindByEngineNumber(ctx, in.EngineNumber); err == nil {
			if existing.DriverID != driverID {
				return domainerrors.AlreadyExists("số máy đã được đăng ký cho một xe khác")
			}
		} else if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return err
		}
	}
	if in.ChassisNumber != "" {
		if existing, err := repo.FindByChassisNumber(ctx, in.ChassisNumber); err == nil {
			if existing.DriverID != driverID {
				return domainerrors.AlreadyExists("số khung đã được đăng ký cho một xe khác")
			}
		} else if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return err
		}
	}
	return nil
}

// checkLicenseRule enforces Phần 1's Rule Engine — only when RideEnabled
// (Phần 2: Delivery never checks GPLX).
func checkLicenseRule(ctx context.Context, rules repository.LicenseCapabilityRepository, in SubmitVehicleVerificationInput) error {
	if !in.RideEnabled {
		return nil
	}
	allowed, err := rules.IsAllowed(ctx, in.LicenseClass, in.ServiceType)
	if err != nil {
		return err
	}
	if !allowed {
		return domainerrors.InvalidArgument("hạng bằng lái " + string(in.LicenseClass) + " không phù hợp với loại xe đã đăng ký")
	}
	return nil
}

// SubmitVehicleVerificationUseCase creates a vehicle KYC record (or edits/
// resubmits an existing one — Phần 3: editing an Approved/Expired record
// invalidates it back to Pending first). Requires vehicle registration +
// insurance documents to already be uploaded, and (only when RideEnabled)
// the license/GPLX document too. Enforces plate/VIN/engine/chassis
// uniqueness (Phần 5/6) and the License Capability rule (Phần 1). Every
// call is audited (Phần 7).
type SubmitVehicleVerificationUseCase struct {
	repo      repository.VehicleVerificationRepository
	documents repository.KYCDocumentRepository
	rules     repository.LicenseCapabilityRepository
	audit     repository.AuditLogRepository
}

func NewSubmitVehicleVerificationUseCase(repo repository.VehicleVerificationRepository, documents repository.KYCDocumentRepository, rules repository.LicenseCapabilityRepository, audit repository.AuditLogRepository) *SubmitVehicleVerificationUseCase {
	return &SubmitVehicleVerificationUseCase{repo: repo, documents: documents, rules: rules, audit: audit}
}

func (uc *SubmitVehicleVerificationUseCase) Execute(ctx context.Context, in SubmitVehicleVerificationInput) (*entity.VehicleVerification, error) {
	if in.DriverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	if err := requireDocuments(ctx, uc.documents, in.DriverID, entity.VehicleDocumentTypes); err != nil {
		return nil, err
	}
	if in.RideEnabled {
		if err := requireDocuments(ctx, uc.documents, in.DriverID, entity.RideLicenseDocumentTypes); err != nil {
			return nil, err
		}
	}
	if err := checkVehicleDedup(ctx, uc.repo, in.DriverID, in); err != nil {
		return nil, err
	}
	if err := checkLicenseRule(ctx, uc.rules, in); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	existing, err := uc.repo.FindByDriverID(ctx, in.DriverID)
	if err == nil {
		oldValue := vehicleVerificationSnapshot(existing)
		existing.Invalidate(now)
		if resubmitErr := existing.Resubmit(in.VehicleType, in.ServiceType, in.Brand, in.Model, in.Year, in.Color, in.PlateNumber, in.VIN, in.EngineNumber, in.ChassisNumber, in.LicenseClass, in.RideEnabled, in.DeliveryEnabled, now); resubmitErr != nil {
			return nil, resubmitErr
		}
		if err := uc.repo.Save(ctx, existing); err != nil {
			return nil, err
		}
		if err := recordAudit(ctx, uc.audit, entity.AuditEntityVehicleVerification, existing.ID, in.DriverID, entity.AuditActionModify, in.DriverID, oldValue, vehicleVerificationSnapshot(existing), ""); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
		return nil, err
	}
	v, err := entity.NewVehicleVerification(uuid.NewString(), in.DriverID, in.VehicleType, in.ServiceType, in.Brand, in.Model, in.Year, in.Color, in.PlateNumber, in.VIN, in.EngineNumber, in.ChassisNumber, in.LicenseClass, in.RideEnabled, in.DeliveryEnabled, now)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	if err := recordAudit(ctx, uc.audit, entity.AuditEntityVehicleVerification, v.ID, in.DriverID, entity.AuditActionSubmit, in.DriverID, "", vehicleVerificationSnapshot(v), ""); err != nil {
		return nil, err
	}
	return v, nil
}

// UpdateVehicleVerificationUseCase edits an existing vehicle verification —
// PUT semantics. Same re-verification/dedup/license-rule enforcement as Submit.
type UpdateVehicleVerificationUseCase struct {
	repo  repository.VehicleVerificationRepository
	rules repository.LicenseCapabilityRepository
	audit repository.AuditLogRepository
}

func NewUpdateVehicleVerificationUseCase(repo repository.VehicleVerificationRepository, rules repository.LicenseCapabilityRepository, audit repository.AuditLogRepository) *UpdateVehicleVerificationUseCase {
	return &UpdateVehicleVerificationUseCase{repo: repo, rules: rules, audit: audit}
}

func (uc *UpdateVehicleVerificationUseCase) Execute(ctx context.Context, in SubmitVehicleVerificationInput) (*entity.VehicleVerification, error) {
	if in.DriverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	if err := checkVehicleDedup(ctx, uc.repo, in.DriverID, in); err != nil {
		return nil, err
	}
	if err := checkLicenseRule(ctx, uc.rules, in); err != nil {
		return nil, err
	}
	v, err := uc.repo.FindByDriverID(ctx, in.DriverID)
	if err != nil {
		return nil, err
	}
	oldValue := vehicleVerificationSnapshot(v)
	now := time.Now().UTC()
	v.Invalidate(now)
	if err := v.Resubmit(in.VehicleType, in.ServiceType, in.Brand, in.Model, in.Year, in.Color, in.PlateNumber, in.VIN, in.EngineNumber, in.ChassisNumber, in.LicenseClass, in.RideEnabled, in.DeliveryEnabled, now); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	if err := recordAudit(ctx, uc.audit, entity.AuditEntityVehicleVerification, v.ID, in.DriverID, entity.AuditActionModify, in.DriverID, oldValue, vehicleVerificationSnapshot(v), ""); err != nil {
		return nil, err
	}
	return v, nil
}

// GetVehicleVerificationUseCase returns a driver's own vehicle KYC record.
type GetVehicleVerificationUseCase struct {
	repo repository.VehicleVerificationRepository
}

func NewGetVehicleVerificationUseCase(repo repository.VehicleVerificationRepository) *GetVehicleVerificationUseCase {
	return &GetVehicleVerificationUseCase{repo: repo}
}

func (uc *GetVehicleVerificationUseCase) Execute(ctx context.Context, driverID string) (*entity.VehicleVerification, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	return uc.repo.FindByDriverID(ctx, driverID)
}

// ReviewVehicleVerificationInput carries an admin's approve/reject decision.
type ReviewVehicleVerificationInput struct {
	DriverID string
	Reviewer string
	Action   ReviewAction
	Reason   string
}

// ReviewVehicleVerificationUseCase applies an admin decision.
type ReviewVehicleVerificationUseCase struct {
	repo  repository.VehicleVerificationRepository
	audit repository.AuditLogRepository
}

func NewReviewVehicleVerificationUseCase(repo repository.VehicleVerificationRepository, audit repository.AuditLogRepository) *ReviewVehicleVerificationUseCase {
	return &ReviewVehicleVerificationUseCase{repo: repo, audit: audit}
}

func (uc *ReviewVehicleVerificationUseCase) Execute(ctx context.Context, in ReviewVehicleVerificationInput) (*entity.VehicleVerification, error) {
	if in.DriverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	v, err := uc.repo.FindByDriverID(ctx, in.DriverID)
	if err != nil {
		return nil, err
	}
	oldValue := vehicleVerificationSnapshot(v)
	now := time.Now().UTC()
	var auditAction entity.AuditAction
	switch in.Action {
	case ReviewApprove:
		if err := v.Approve(in.Reviewer, now); err != nil {
			return nil, err
		}
		auditAction = entity.AuditActionApprove
	case ReviewReject:
		if err := v.Reject(in.Reviewer, in.Reason, now); err != nil {
			return nil, err
		}
		auditAction = entity.AuditActionReject
	default:
		return nil, domainerrors.InvalidArgument("unknown review action: " + string(in.Action))
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	if err := recordAudit(ctx, uc.audit, entity.AuditEntityVehicleVerification, v.ID, in.DriverID, auditAction, in.Reviewer, oldValue, vehicleVerificationSnapshot(v), in.Reason); err != nil {
		return nil, err
	}
	return v, nil
}

// ListVehicleVerificationsUseCase powers the admin review dashboard —
// status/vehicleType/serviceType are each optional filters (empty string
// means "any"). sortByExpiry (Phần 12) orders by nearest upcoming
// expiry-tracked document instead of newest-submission-first.
type ListVehicleVerificationsUseCase struct {
	repo repository.VehicleVerificationRepository
}

func NewListVehicleVerificationsUseCase(repo repository.VehicleVerificationRepository) *ListVehicleVerificationsUseCase {
	return &ListVehicleVerificationsUseCase{repo: repo}
}

func (uc *ListVehicleVerificationsUseCase) Execute(ctx context.Context, status entity.KYCStatus, vehicleType entity.VehicleType, serviceType entity.ServiceType, limit int, sortByExpiry bool) ([]*entity.VehicleVerification, error) {
	if sortByExpiry {
		return uc.repo.ListByFilterSortedByExpiry(ctx, status, vehicleType, serviceType, limit)
	}
	return uc.repo.ListByFilter(ctx, status, vehicleType, serviceType, limit)
}

package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// ReviewAction names the two admin decisions available on a pending/
// under-review KYC record — shared by driver and vehicle review use cases.
type ReviewAction string

const (
	ReviewApprove ReviewAction = "approve"
	ReviewReject  ReviewAction = "reject"
)

// SubmitDriverVerificationInput carries the personal-info fields collected
// in Step 1 of the Become Driver wizard. NationalIDNumber (Phần 5) is the
// CCCD's printed number — required, checked for duplicates across drivers.
type SubmitDriverVerificationInput struct {
	DriverID         string
	FullName         string
	DateOfBirth      time.Time
	Address          string
	NationalIDNumber string
	LicenseNumber    string
}

func checkDriverDedup(ctx context.Context, repo repository.DriverVerificationRepository, driverID, nationalIDNumber, licenseNumber string) error {
	if nationalIDNumber != "" {
		existing, err := repo.FindByNationalIDNumber(ctx, nationalIDNumber)
		if err == nil && existing.DriverID != driverID {
			return domainerrors.AlreadyExists("số CCCD đã được đăng ký bởi một tài xế khác")
		}
		if err != nil && domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return err
		}
	}
	if licenseNumber != "" {
		existing, err := repo.FindByLicenseNumber(ctx, licenseNumber)
		if err == nil && existing.DriverID != driverID {
			return domainerrors.AlreadyExists("số GPLX đã được đăng ký bởi một tài xế khác")
		}
		if err != nil && domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return err
		}
	}
	return nil
}

// SubmitDriverVerificationUseCase creates a driver's KYC record (or, if one
// already exists, edits and resubmits it — Phần 3: editing an
// Approved/Expired record is first Invalidated back to Pending). Requires
// CCCD front/back + selfie to already be uploaded (Phần 10 — "Không cho
// submit thiếu file") and enforces CCCD/GPLX-number uniqueness (Phần 5).
// Every call is audited (Phần 7).
type SubmitDriverVerificationUseCase struct {
	repo      repository.DriverVerificationRepository
	documents repository.KYCDocumentRepository
	audit     repository.AuditLogRepository
}

func NewSubmitDriverVerificationUseCase(repo repository.DriverVerificationRepository, documents repository.KYCDocumentRepository, audit repository.AuditLogRepository) *SubmitDriverVerificationUseCase {
	return &SubmitDriverVerificationUseCase{repo: repo, documents: documents, audit: audit}
}

func (uc *SubmitDriverVerificationUseCase) Execute(ctx context.Context, in SubmitDriverVerificationInput) (*entity.DriverVerification, error) {
	if in.DriverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	if err := requireDocuments(ctx, uc.documents, in.DriverID, entity.DriverDocumentTypes); err != nil {
		return nil, err
	}
	if err := checkDriverDedup(ctx, uc.repo, in.DriverID, in.NationalIDNumber, in.LicenseNumber); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	existing, err := uc.repo.FindByDriverID(ctx, in.DriverID)
	if err == nil {
		oldValue := driverVerificationSnapshot(existing)
		existing.Invalidate(now)
		if resubmitErr := existing.Resubmit(in.FullName, in.DateOfBirth, in.Address, in.NationalIDNumber, in.LicenseNumber, now); resubmitErr != nil {
			return nil, resubmitErr
		}
		if err := uc.repo.Save(ctx, existing); err != nil {
			return nil, err
		}
		if err := recordAudit(ctx, uc.audit, entity.AuditEntityDriverVerification, existing.ID, in.DriverID, entity.AuditActionModify, in.DriverID, oldValue, driverVerificationSnapshot(existing), ""); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
		return nil, err
	}
	v, err := entity.NewDriverVerification(uuid.NewString(), in.DriverID, in.FullName, in.DateOfBirth, in.Address, in.NationalIDNumber, in.LicenseNumber, now)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	if err := recordAudit(ctx, uc.audit, entity.AuditEntityDriverVerification, v.ID, in.DriverID, entity.AuditActionSubmit, in.DriverID, "", driverVerificationSnapshot(v), ""); err != nil {
		return nil, err
	}
	return v, nil
}

// UpdateDriverVerificationUseCase edits an existing driver verification —
// PUT semantics, unlike Submit's create-or-resubmit. Same re-verification
// and dedup rules as Submit.
type UpdateDriverVerificationUseCase struct {
	repo  repository.DriverVerificationRepository
	audit repository.AuditLogRepository
}

func NewUpdateDriverVerificationUseCase(repo repository.DriverVerificationRepository, audit repository.AuditLogRepository) *UpdateDriverVerificationUseCase {
	return &UpdateDriverVerificationUseCase{repo: repo, audit: audit}
}

func (uc *UpdateDriverVerificationUseCase) Execute(ctx context.Context, in SubmitDriverVerificationInput) (*entity.DriverVerification, error) {
	if in.DriverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	if err := checkDriverDedup(ctx, uc.repo, in.DriverID, in.NationalIDNumber, in.LicenseNumber); err != nil {
		return nil, err
	}
	v, err := uc.repo.FindByDriverID(ctx, in.DriverID)
	if err != nil {
		return nil, err
	}
	oldValue := driverVerificationSnapshot(v)
	now := time.Now().UTC()
	v.Invalidate(now)
	if err := v.Resubmit(in.FullName, in.DateOfBirth, in.Address, in.NationalIDNumber, in.LicenseNumber, now); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	if err := recordAudit(ctx, uc.audit, entity.AuditEntityDriverVerification, v.ID, in.DriverID, entity.AuditActionModify, in.DriverID, oldValue, driverVerificationSnapshot(v), ""); err != nil {
		return nil, err
	}
	return v, nil
}

// GetDriverVerificationUseCase returns a driver's own KYC record.
type GetDriverVerificationUseCase struct {
	repo repository.DriverVerificationRepository
}

func NewGetDriverVerificationUseCase(repo repository.DriverVerificationRepository) *GetDriverVerificationUseCase {
	return &GetDriverVerificationUseCase{repo: repo}
}

func (uc *GetDriverVerificationUseCase) Execute(ctx context.Context, driverID string) (*entity.DriverVerification, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	return uc.repo.FindByDriverID(ctx, driverID)
}

// ReviewDriverVerificationInput carries an admin's approve/reject decision.
type ReviewDriverVerificationInput struct {
	DriverID string
	Reviewer string
	Action   ReviewAction
	Reason   string // required when Action == ReviewReject
}

// ReviewDriverVerificationUseCase applies an admin decision (Phần 9 —
// callers must independently verify the requester is an admin before
// calling this; this use case trusts its Reviewer input).
type ReviewDriverVerificationUseCase struct {
	repo  repository.DriverVerificationRepository
	audit repository.AuditLogRepository
}

func NewReviewDriverVerificationUseCase(repo repository.DriverVerificationRepository, audit repository.AuditLogRepository) *ReviewDriverVerificationUseCase {
	return &ReviewDriverVerificationUseCase{repo: repo, audit: audit}
}

func (uc *ReviewDriverVerificationUseCase) Execute(ctx context.Context, in ReviewDriverVerificationInput) (*entity.DriverVerification, error) {
	if in.DriverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	v, err := uc.repo.FindByDriverID(ctx, in.DriverID)
	if err != nil {
		return nil, err
	}
	oldValue := driverVerificationSnapshot(v)
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
	if err := recordAudit(ctx, uc.audit, entity.AuditEntityDriverVerification, v.ID, in.DriverID, auditAction, in.Reviewer, oldValue, driverVerificationSnapshot(v), in.Reason); err != nil {
		return nil, err
	}
	return v, nil
}

// ListDriverVerificationsUseCase powers the admin review dashboard.
type ListDriverVerificationsUseCase struct {
	repo repository.DriverVerificationRepository
}

func NewListDriverVerificationsUseCase(repo repository.DriverVerificationRepository) *ListDriverVerificationsUseCase {
	return &ListDriverVerificationsUseCase{repo: repo}
}

func (uc *ListDriverVerificationsUseCase) Execute(ctx context.Context, status entity.KYCStatus, limit int) ([]*entity.DriverVerification, error) {
	return uc.repo.ListByStatus(ctx, status, limit)
}

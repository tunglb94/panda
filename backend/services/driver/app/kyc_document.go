package app

import (
	"context"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	"github.com/fairride/driver/infrastructure/localstore"
	domainerrors "github.com/fairride/shared/errors"
)

// requireDocuments returns InvalidArgument naming the first of types that
// hasn't been uploaded yet for driverID — the enforcement behind Phần 10's
// "Không cho submit thiếu file", shared by both Submit use cases.
func requireDocuments(ctx context.Context, documents repository.KYCDocumentRepository, driverID string, types []entity.DocumentType) error {
	for _, dt := range types {
		if _, err := documents.FindByDriverAndType(ctx, driverID, dt); err != nil {
			if domainerrors.GetCode(err) == domainerrors.CodeNotFound {
				return domainerrors.InvalidArgument("missing required document: " + string(dt))
			}
			return err
		}
	}
	return nil
}

// documentOwner reports which verification a document type belongs to —
// CCCD/Selfie identify the driver; License/VehicleRegistration/
// VehicleInsurance/VehicleInspection identify the vehicle (License is
// checked as part of vehicle submission — see RideLicenseDocumentTypes —
// since its relevance is gated on the vehicle's RideEnabled flag, not the
// driver's identity). Used by UploadKYCDocumentUseCase to decide which
// verification to invalidate (Phần 3) when a document is re-uploaded.
func documentOwner(docType entity.DocumentType) entity.AuditEntityType {
	switch docType {
	case entity.DocumentCCCDFront, entity.DocumentCCCDBack, entity.DocumentSelfie:
		return entity.AuditEntityDriverVerification
	default:
		return entity.AuditEntityVehicleVerification
	}
}

// UploadKYCDocumentInput carries one raw file upload. ExpiresAt (Phần 2) is
// only meaningful for entity.ExpiringDocumentTypes — ignored otherwise.
type UploadKYCDocumentInput struct {
	DriverID     string
	DocumentType entity.DocumentType
	Filename     string // used only to derive a storage extension
	ContentType  string
	ExpiresAt    *time.Time
	Data         io.Reader
}

// UploadKYCDocumentUseCase saves a document to local disk (no cloud upload)
// and records its metadata as a NEW version (Phần 4 — never overwrites a
// previous upload). If the document's owning verification (Driver or
// Vehicle) is currently Approved or Expired, re-uploading invalidates it
// back to Pending (Phần 3 — Re-verification) so a driver can't keep Online
// eligibility with a swapped-out document that was never reviewed.
type UploadKYCDocumentUseCase struct {
	documents            repository.KYCDocumentRepository
	driverVerifications  repository.DriverVerificationRepository
	vehicleVerifications repository.VehicleVerificationRepository
	audit                repository.AuditLogRepository
	store                *localstore.DocumentStore
}

func NewUploadKYCDocumentUseCase(
	documents repository.KYCDocumentRepository,
	driverVerifications repository.DriverVerificationRepository,
	vehicleVerifications repository.VehicleVerificationRepository,
	audit repository.AuditLogRepository,
	store *localstore.DocumentStore,
) *UploadKYCDocumentUseCase {
	return &UploadKYCDocumentUseCase{
		documents: documents, driverVerifications: driverVerifications, vehicleVerifications: vehicleVerifications,
		audit: audit, store: store,
	}
}

func (uc *UploadKYCDocumentUseCase) Execute(ctx context.Context, in UploadKYCDocumentInput) (*entity.KYCDocument, error) {
	if in.DriverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	if err := entity.ValidateDocumentType(in.DocumentType); err != nil {
		return nil, err
	}
	if in.Data == nil {
		return nil, domainerrors.InvalidArgument("file data is required")
	}
	nextVersion := 1
	if prior, err := uc.documents.FindByDriverAndType(ctx, in.DriverID, in.DocumentType); err == nil {
		nextVersion = prior.Version + 1
	} else if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
		return nil, err
	}

	relPath, err := uc.store.Save(ctx, in.DriverID, string(in.DocumentType), filepath.Ext(in.Filename), in.Data)
	if err != nil {
		return nil, err
	}
	var expiresAt *time.Time
	if entity.ExpiringDocumentTypes[in.DocumentType] {
		expiresAt = in.ExpiresAt
	}
	now := time.Now().UTC()
	doc, err := entity.NewKYCDocument(uuid.NewString(), in.DriverID, in.DocumentType, relPath, in.ContentType, nextVersion, expiresAt, in.DriverID, now)
	if err != nil {
		return nil, err
	}
	if err := uc.documents.Save(ctx, doc); err != nil {
		return nil, err
	}
	if err := recordAudit(ctx, uc.audit, entity.AuditEntityKYCDocument, doc.ID, in.DriverID, entity.AuditActionSubmit, in.DriverID, "", string(doc.DocumentType)+" v"+strconv.Itoa(doc.Version), ""); err != nil {
		return nil, err
	}
	if err := uc.invalidateOwner(ctx, in.DriverID, in.DocumentType, now); err != nil {
		return nil, err
	}
	return doc, nil
}

// invalidateOwner is the Phần 3 hook: re-uploading a document that belongs
// to an Approved/Expired verification forces it back to Pending.
func (uc *UploadKYCDocumentUseCase) invalidateOwner(ctx context.Context, driverID string, docType entity.DocumentType, now time.Time) error {
	switch documentOwner(docType) {
	case entity.AuditEntityDriverVerification:
		dv, err := uc.driverVerifications.FindByDriverID(ctx, driverID)
		if err != nil {
			if domainerrors.GetCode(err) == domainerrors.CodeNotFound {
				return nil
			}
			return err
		}
		if dv.Status != entity.KYCApproved && dv.Status != entity.KYCExpired {
			return nil
		}
		oldValue := driverVerificationSnapshot(dv)
		dv.Invalidate(now)
		if err := uc.driverVerifications.Save(ctx, dv); err != nil {
			return err
		}
		return recordAudit(ctx, uc.audit, entity.AuditEntityDriverVerification, dv.ID, driverID, entity.AuditActionModify, driverID, oldValue, driverVerificationSnapshot(dv), "document re-uploaded: "+string(docType))
	default:
		vv, err := uc.vehicleVerifications.FindByDriverID(ctx, driverID)
		if err != nil {
			if domainerrors.GetCode(err) == domainerrors.CodeNotFound {
				return nil
			}
			return err
		}
		if vv.Status != entity.KYCApproved && vv.Status != entity.KYCExpired {
			return nil
		}
		oldValue := vehicleVerificationSnapshot(vv)
		vv.Invalidate(now)
		if err := uc.vehicleVerifications.Save(ctx, vv); err != nil {
			return err
		}
		return recordAudit(ctx, uc.audit, entity.AuditEntityVehicleVerification, vv.ID, driverID, entity.AuditActionModify, driverID, oldValue, vehicleVerificationSnapshot(vv), "document re-uploaded: "+string(docType))
	}
}

// ListKYCDocumentsUseCase returns the latest version of every document a
// driver has uploaded — used to render the "uploaded: true/false"
// checklist (Phần 9/13: never the storage path itself).
type ListKYCDocumentsUseCase struct {
	documents repository.KYCDocumentRepository
}

func NewListKYCDocumentsUseCase(documents repository.KYCDocumentRepository) *ListKYCDocumentsUseCase {
	return &ListKYCDocumentsUseCase{documents: documents}
}

func (uc *ListKYCDocumentsUseCase) Execute(ctx context.Context, driverID string) ([]*entity.KYCDocument, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	return uc.documents.ListByDriverID(ctx, driverID)
}

// ListKYCDocumentVersionsUseCase returns every version ever uploaded for
// one (driver, document type), newest first — Phần 4/11's version history.
type ListKYCDocumentVersionsUseCase struct {
	documents repository.KYCDocumentRepository
}

func NewListKYCDocumentVersionsUseCase(documents repository.KYCDocumentRepository) *ListKYCDocumentVersionsUseCase {
	return &ListKYCDocumentVersionsUseCase{documents: documents}
}

func (uc *ListKYCDocumentVersionsUseCase) Execute(ctx context.Context, driverID string, docType entity.DocumentType) ([]*entity.KYCDocument, error) {
	if driverID == "" {
		return nil, domainerrors.InvalidArgument("driver_id is required")
	}
	if err := entity.ValidateDocumentType(docType); err != nil {
		return nil, err
	}
	return uc.documents.ListVersionsByDriverAndType(ctx, driverID, docType)
}

// GetKYCDocumentUseCase returns one document's metadata (including its
// internal StoragePath) — the caller (an admin-only gateway handler) is
// responsible for authorizing the request and never exposing StoragePath
// itself; it is used only to open the file server-side.
type GetKYCDocumentUseCase struct {
	documents repository.KYCDocumentRepository
}

func NewGetKYCDocumentUseCase(documents repository.KYCDocumentRepository) *GetKYCDocumentUseCase {
	return &GetKYCDocumentUseCase{documents: documents}
}

func (uc *GetKYCDocumentUseCase) Execute(ctx context.Context, id string) (*entity.KYCDocument, error) {
	if id == "" {
		return nil, domainerrors.InvalidArgument("id is required")
	}
	return uc.documents.FindByID(ctx, id)
}

package app

import (
	"bytes"
	"context"
	"io"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/domain/repository"
)

// RiderDocumentType is one of the two CCCD photos the Rider KYC form collects.
type RiderDocumentType string

const (
	RiderDocumentCCCDFront RiderDocumentType = "cccd_front"
	RiderDocumentCCCDBack  RiderDocumentType = "cccd_back"
)

func (t RiderDocumentType) valid() bool {
	return t == RiderDocumentCCCDFront || t == RiderDocumentCCCDBack
}

// documentStore abstracts localstore.DocumentStore for unit-testability.
type documentStore interface {
	Save(ctx context.Context, userID string, docType string, ext string, data io.Reader) (string, error)
}

// UploadRiderDocumentInput carries one uploaded CCCD photo.
type UploadRiderDocumentInput struct {
	UserID       string
	DocumentType RiderDocumentType
	Ext          string
	Data         io.Reader
}

// UploadRiderDocumentUseCase saves one CCCD photo and attaches its path to
// the rider's (possibly just-created) draft verification record.
type UploadRiderDocumentUseCase struct {
	repo  repository.RiderVerificationRepository
	store documentStore
}

func NewUploadRiderDocumentUseCase(repo repository.RiderVerificationRepository, store documentStore) *UploadRiderDocumentUseCase {
	return &UploadRiderDocumentUseCase{repo: repo, store: store}
}

func (uc *UploadRiderDocumentUseCase) Execute(ctx context.Context, in UploadRiderDocumentInput) (*entity.RiderVerification, error) {
	if in.UserID == "" {
		return nil, domainerrors.InvalidArgument("user_id is required")
	}
	if !in.DocumentType.valid() {
		return nil, domainerrors.InvalidArgument("unknown document type: " + string(in.DocumentType))
	}

	// Read fully into memory — the Rule Engine (Upload -> Rule Engine ->
	// OCR -> Vision -> Decision) needs to inspect the whole image before
	// it's accepted; the handler already caps multipart bodies at 10 MiB,
	// so buffering here is bounded and safe.
	data, err := io.ReadAll(in.Data)
	if err != nil {
		return nil, domainerrors.Internal("read uploaded file").WithMeta("error", err.Error())
	}
	if rule := RunRuleEngine(data, string(in.DocumentType)); !rule.OK {
		return nil, domainerrors.InvalidArgument(rule.Reason)
	}

	now := time.Now()
	v, err := uc.repo.FindByUserID(ctx, in.UserID)
	if err != nil {
		if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return nil, err
		}
		id, idErr := newID()
		if idErr != nil {
			return nil, domainerrors.Internal("generate rider verification id").WithMeta("error", idErr.Error())
		}
		v, err = entity.NewDraftRiderVerification(id, in.UserID, now)
		if err != nil {
			return nil, err
		}
	}

	path, err := uc.store.Save(ctx, in.UserID, string(in.DocumentType), in.Ext, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	if in.DocumentType == RiderDocumentCCCDFront {
		v.AttachFrontPhoto(path, now)
	} else {
		v.AttachBackPhoto(path, now)
	}

	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

package app

import (
	"context"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/domain/repository"
	"github.com/fairride/user/infrastructure/ocr"
	"github.com/fairride/user/infrastructure/vision"
)

// SubmitRiderVerificationInput carries the three text fields collected on
// the Rider KYC screen — the two CCCD photos must already have been
// uploaded via UploadRiderDocumentUseCase.
type SubmitRiderVerificationInput struct {
	UserID           string
	FullName         string
	DateOfBirth      time.Time
	NationalIDNumber string
}

// SubmitRiderVerificationUseCase creates or resubmits a rider's KYC record,
// then immediately runs the OCR/Vision/Decision stages of the AI pipeline
// (the Rule Engine already ran per-photo at upload time — see
// UploadRiderDocumentUseCase) and records the result via
// RiderVerification.ApplyPipelineDecision.
type SubmitRiderVerificationUseCase struct {
	repo   repository.RiderVerificationRepository
	ocr    ocr.Provider
	vision vision.Provider
}

func NewSubmitRiderVerificationUseCase(repo repository.RiderVerificationRepository, ocrProvider ocr.Provider, visionProvider vision.Provider) *SubmitRiderVerificationUseCase {
	return &SubmitRiderVerificationUseCase{repo: repo, ocr: ocrProvider, vision: visionProvider}
}

func (uc *SubmitRiderVerificationUseCase) Execute(ctx context.Context, in SubmitRiderVerificationInput) (*entity.RiderVerification, error) {
	if in.UserID == "" {
		return nil, domainerrors.InvalidArgument("user_id is required")
	}

	now := time.Now()
	v, err := uc.repo.FindByUserID(ctx, in.UserID)
	if err != nil {
		if domainerrors.GetCode(err) == domainerrors.CodeNotFound {
			return nil, domainerrors.InvalidArgument("please upload both CCCD photos before submitting")
		}
		return nil, err
	}

	if err := v.Submit(in.FullName, in.DateOfBirth, in.NationalIDNumber, now); err != nil {
		return nil, err
	}

	// Both photos already passed the Rule Engine individually at upload
	// time (see UploadRiderDocumentUseCase) — rule.OK is true here by
	// construction; Decide still takes it so the pipeline's shape matches
	// Upload -> Rule Engine -> OCR -> Vision -> Decision exactly.
	ocrResult, err := uc.ocr.Extract(ctx, v.CCCDFrontPath, string(RiderDocumentCCCDFront))
	if err != nil {
		return nil, domainerrors.Internal("ocr extraction failed").WithMeta("error", err.Error())
	}
	visionResult, err := uc.vision.Analyze(ctx, v.CCCDFrontPath, string(RiderDocumentCCCDFront))
	if err != nil {
		return nil, domainerrors.Internal("vision analysis failed").WithMeta("error", err.Error())
	}
	mode, confidence, reason := Decide(RuleEngineResult{OK: true}, ocrResult, visionResult)
	if err := v.ApplyPipelineDecision(mode, confidence, ocrResult.Raw, visionResult.Raw, reason, time.Now()); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

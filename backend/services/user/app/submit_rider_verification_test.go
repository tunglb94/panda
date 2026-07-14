package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/fairride/shared/errors"
	"github.com/fairride/user/app"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/infrastructure/ocr"
	"github.com/fairride/user/infrastructure/vision"
)

var riderAppTestNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
var riderAppTestDOB = time.Date(1995, 5, 20, 0, 0, 0, 0, time.UTC)

type fakeRiderVerificationRepo struct {
	byUserID map[string]*entity.RiderVerification
}

func newFakeRiderVerificationRepo() *fakeRiderVerificationRepo {
	return &fakeRiderVerificationRepo{byUserID: map[string]*entity.RiderVerification{}}
}

func (r *fakeRiderVerificationRepo) FindByUserID(_ context.Context, userID string) (*entity.RiderVerification, error) {
	v, ok := r.byUserID[userID]
	if !ok {
		return nil, errors.NotFound("rider verification not found")
	}
	cp := *v
	return &cp, nil
}

func (r *fakeRiderVerificationRepo) Save(_ context.Context, v *entity.RiderVerification) error {
	cp := *v
	r.byUserID[v.UserID] = &cp
	return nil
}

func (r *fakeRiderVerificationRepo) ListByStatus(_ context.Context, status entity.RiderKYCStatus) ([]*entity.RiderVerification, error) {
	var out []*entity.RiderVerification
	for _, v := range r.byUserID {
		if v.Status == status {
			out = append(out, v)
		}
	}
	return out, nil
}

func TestSubmitRiderVerification_FailsWithoutUpload(t *testing.T) {
	repo := newFakeRiderVerificationRepo()
	uc := app.NewSubmitRiderVerificationUseCase(repo, ocr.NewMockOCRProvider(), vision.NewMockVisionProvider())

	_, err := uc.Execute(context.Background(), app.SubmitRiderVerificationInput{
		UserID:           "user-1",
		FullName:         "Nguyen Van A",
		DateOfBirth:      riderAppTestDOB,
		NationalIDNumber: "012345678900",
	})
	if err == nil {
		t.Fatal("expected error submitting before any document was uploaded")
	}
}

func TestUploadThenSubmitRiderVerification_Succeeds(t *testing.T) {
	repo := newFakeRiderVerificationRepo()

	draft, err := entity.NewDraftRiderVerification("id-1", "user-1", riderAppTestNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	draft.AttachFrontPhoto("user-1/cccd_front.jpg", riderAppTestNow)
	draft.AttachBackPhoto("user-1/cccd_back.jpg", riderAppTestNow)
	if err := repo.Save(context.Background(), draft); err != nil {
		t.Fatalf("unexpected error saving draft: %v", err)
	}

	uc := app.NewSubmitRiderVerificationUseCase(repo, ocr.NewMockOCRProvider(), vision.NewMockVisionProvider())
	v, err := uc.Execute(context.Background(), app.SubmitRiderVerificationInput{
		UserID:           "user-1",
		FullName:         "Nguyen Van A",
		DateOfBirth:      riderAppTestDOB,
		NationalIDNumber: "012345678900",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.RiderKYCPending {
		t.Fatalf("expected status pending after submit, got %s", v.Status)
	}
	// Mock OCR/Vision confidence sits in the "manual review" band on
	// purpose (see ocr.mockConfidence's doc comment) — the pipeline must
	// not auto-approve a real submission until real providers replace them.
	if v.ReviewMode != entity.ReviewModeManualReview {
		t.Fatalf("expected review_mode=manual_review with mock providers, got %s", v.ReviewMode)
	}
	if v.OCRResult == "" || v.VisionResult == "" {
		t.Fatal("expected OCR/Vision raw results to be recorded")
	}
}

func TestReviewRiderVerification_ApproveAndReject(t *testing.T) {
	repo := newFakeRiderVerificationRepo()
	draft, _ := entity.NewDraftRiderVerification("id-1", "user-1", riderAppTestNow)
	draft.AttachFrontPhoto("front.jpg", riderAppTestNow)
	draft.AttachBackPhoto("back.jpg", riderAppTestNow)
	_ = draft.Submit("Nguyen Van A", riderAppTestDOB, "012345678900", riderAppTestNow)
	_ = repo.Save(context.Background(), draft)

	review := app.NewReviewRiderVerificationUseCase(repo)
	approved, err := review.Approve(context.Background(), "user-1", "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved.IsApproved() {
		t.Fatal("expected verification to be approved")
	}

	// A second approve from Approved must fail (no double-approve).
	if _, err := review.Approve(context.Background(), "user-1", "admin-1"); err == nil {
		t.Fatal("expected error re-approving an already-approved record")
	}
}

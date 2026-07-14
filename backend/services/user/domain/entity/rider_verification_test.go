package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/user/domain/entity"
)

var riderTestNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
var riderTestDOB = time.Date(1995, 5, 20, 0, 0, 0, 0, time.UTC)

func submittedDraft(t *testing.T) *entity.RiderVerification {
	t.Helper()
	v, _ := entity.NewDraftRiderVerification("id-1", "user-1", riderTestNow)
	v.AttachFrontPhoto("front.jpg", riderTestNow)
	v.AttachBackPhoto("back.jpg", riderTestNow)
	if err := v.Submit("Nguyen Van A", riderTestDOB, "012345678900", riderTestNow); err != nil {
		t.Fatalf("submittedDraft: unexpected error: %v", err)
	}
	return v
}

func TestApplyPipelineDecision_AutoApproved_TransitionsStatus(t *testing.T) {
	v := submittedDraft(t)
	if err := v.ApplyPipelineDecision(entity.ReviewModeAutoApproved, 0.97, `{"ocr":1}`, `{"vision":1}`, "", riderTestNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.RiderKYCApproved {
		t.Fatalf("expected Status=Approved, got %s", v.Status)
	}
	if v.ReviewMode != entity.ReviewModeAutoApproved {
		t.Fatalf("expected ReviewMode=AutoApproved, got %s", v.ReviewMode)
	}
	if v.AIConfidence != 0.97 {
		t.Fatalf("expected AIConfidence=0.97, got %v", v.AIConfidence)
	}
}

func TestApplyPipelineDecision_Rejected_TransitionsStatus(t *testing.T) {
	v := submittedDraft(t)
	if err := v.ApplyPipelineDecision(entity.ReviewModeRejected, 0.1, "{}", "{}", "độ tin cậy thấp", riderTestNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.RiderKYCRejected {
		t.Fatalf("expected Status=Rejected, got %s", v.Status)
	}
	if v.RejectReason != "độ tin cậy thấp" {
		t.Fatalf("expected reject reason to propagate, got %q", v.RejectReason)
	}
}

func TestApplyPipelineDecision_ManualReview_StaysPending(t *testing.T) {
	v := submittedDraft(t)
	if err := v.ApplyPipelineDecision(entity.ReviewModeManualReview, 0.5, "{}", "{}", "", riderTestNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.RiderKYCPending {
		t.Fatalf("expected Status to stay Pending for ManualReview, got %s", v.Status)
	}
	if v.ReviewMode != entity.ReviewModeManualReview {
		t.Fatalf("expected ReviewMode=ManualReview, got %s", v.ReviewMode)
	}
}

func TestApplyPipelineDecision_OnlyValidFromPending(t *testing.T) {
	v := submittedDraft(t)
	if err := v.Approve("admin-1", riderTestNow); err != nil {
		t.Fatalf("unexpected error approving: %v", err)
	}
	if err := v.ApplyPipelineDecision(entity.ReviewModeAutoApproved, 0.99, "{}", "{}", "", riderTestNow); err == nil {
		t.Fatal("expected error applying a pipeline decision to an already-Approved record")
	}
}

func TestSubmit_RequiresBothPhotos(t *testing.T) {
	v, err := entity.NewDraftRiderVerification("id-1", "user-1", riderTestNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := v.Submit("Nguyen Van A", riderTestDOB, "012345678900", riderTestNow); err == nil {
		t.Fatal("expected error submitting without any photo uploaded")
	}
	v.AttachFrontPhoto("user-1/cccd_front_abc.jpg", riderTestNow)
	if err := v.Submit("Nguyen Van A", riderTestDOB, "012345678900", riderTestNow); err == nil {
		t.Fatal("expected error submitting with only the front photo uploaded")
	}
	v.AttachBackPhoto("user-1/cccd_back_abc.jpg", riderTestNow)
	if err := v.Submit("Nguyen Van A", riderTestDOB, "012345678900", riderTestNow); err != nil {
		t.Fatalf("expected success once both photos are uploaded, got %v", err)
	}
	if v.Status != entity.RiderKYCPending {
		t.Fatalf("expected status Pending after submit, got %s", v.Status)
	}
	if v.SubmittedAt == nil {
		t.Fatal("expected SubmittedAt to be set after submit")
	}
}

func TestSubmit_ValidatesFields(t *testing.T) {
	v, _ := entity.NewDraftRiderVerification("id-1", "user-1", riderTestNow)
	v.AttachFrontPhoto("front.jpg", riderTestNow)
	v.AttachBackPhoto("back.jpg", riderTestNow)

	if err := v.Submit("", riderTestDOB, "012345678900", riderTestNow); err == nil {
		t.Fatal("expected error for empty full name")
	}
	if err := v.Submit("Nguyen Van A", time.Time{}, "012345678900", riderTestNow); err == nil {
		t.Fatal("expected error for zero date of birth")
	}
	if err := v.Submit("Nguyen Van A", riderTestNow.Add(24*time.Hour), "012345678900", riderTestNow); err == nil {
		t.Fatal("expected error for a future date of birth")
	}
	if err := v.Submit("Nguyen Van A", riderTestDOB, "", riderTestNow); err == nil {
		t.Fatal("expected error for empty national id number")
	}
}

func TestApproveReject_OnlyFromPending(t *testing.T) {
	v, _ := entity.NewDraftRiderVerification("id-1", "user-1", riderTestNow)
	v.AttachFrontPhoto("front.jpg", riderTestNow)
	v.AttachBackPhoto("back.jpg", riderTestNow)
	if err := v.Submit("Nguyen Van A", riderTestDOB, "012345678900", riderTestNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := v.Approve("admin-1", riderTestNow); err != nil {
		t.Fatalf("expected approve to succeed from pending, got %v", err)
	}
	if !v.IsApproved() {
		t.Fatal("expected IsApproved() true after Approve")
	}
	if err := v.Approve("admin-1", riderTestNow); err == nil {
		t.Fatal("expected error approving an already-approved record")
	}
}

func TestReject_RequiresReason(t *testing.T) {
	v, _ := entity.NewDraftRiderVerification("id-1", "user-1", riderTestNow)
	v.AttachFrontPhoto("front.jpg", riderTestNow)
	v.AttachBackPhoto("back.jpg", riderTestNow)
	_ = v.Submit("Nguyen Van A", riderTestDOB, "012345678900", riderTestNow)

	if err := v.Reject("admin-1", "", riderTestNow); err == nil {
		t.Fatal("expected error rejecting without a reason")
	}
	if err := v.Reject("admin-1", "ảnh mờ", riderTestNow); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.RiderKYCRejected {
		t.Fatalf("expected status Rejected, got %s", v.Status)
	}
}

func TestAttachPhoto_ReopensRejected(t *testing.T) {
	v, _ := entity.NewDraftRiderVerification("id-1", "user-1", riderTestNow)
	v.AttachFrontPhoto("front.jpg", riderTestNow)
	v.AttachBackPhoto("back.jpg", riderTestNow)
	_ = v.Submit("Nguyen Van A", riderTestDOB, "012345678900", riderTestNow)
	_ = v.Reject("admin-1", "ảnh mờ", riderTestNow)

	v.AttachFrontPhoto("front-v2.jpg", riderTestNow)
	if v.Status != entity.RiderKYCPending {
		t.Fatalf("expected re-uploading a photo after Rejected to reopen to Pending, got %s", v.Status)
	}
	if v.RejectReason != "" {
		t.Fatal("expected reject reason to be cleared after reopening")
	}
}

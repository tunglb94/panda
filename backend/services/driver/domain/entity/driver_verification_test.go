package entity_test

import (
	"testing"
	"time"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/shared/errors"
)

var dvNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
var dvDOB = time.Date(1995, 5, 20, 0, 0, 0, 0, time.UTC)

func TestNewDriverVerification_OK(t *testing.T) {
	v, err := entity.NewDriverVerification("id1", "d1", "Nguyen Van A", dvDOB, "123 Le Loi", "079095001234", "GPLX123", dvNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending", v.Status)
	}
	if v.IsApproved() {
		t.Error("freshly created verification must not be approved")
	}
}

func TestNewDriverVerification_RejectsFutureDateOfBirth(t *testing.T) {
	future := dvNow.Add(24 * time.Hour)
	_, err := entity.NewDriverVerification("id1", "d1", "A", future, "addr", "079095001234", "", dvNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument, got %v", err)
	}
}

func TestNewDriverVerification_RejectsEmptyFullName(t *testing.T) {
	_, err := entity.NewDriverVerification("id1", "d1", "  ", dvDOB, "addr", "079095001234", "", dvNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument, got %v", err)
	}
}

func TestNewDriverVerification_RejectsEmptyNationalIDNumber(t *testing.T) {
	_, err := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "", "", dvNow)
	if !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for empty national_id_number, got %v", err)
	}
}

func TestDriverVerification_ApproveRejectLifecycle(t *testing.T) {
	v, _ := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "079095001234", "", dvNow)

	if err := v.Approve("admin1", dvNow); err != nil {
		t.Fatalf("approve from pending should succeed: %v", err)
	}
	if !v.IsApproved() {
		t.Error("should be approved")
	}
	if v.ApprovedAt == nil {
		t.Error("ApprovedAt must be set")
	}

	// Cannot approve again from Approved.
	if err := v.Approve("admin1", dvNow); !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("re-approving should fail with PreconditionFailed, got %v", err)
	}
}

func TestDriverVerification_RejectRequiresReason(t *testing.T) {
	v, _ := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "079095001234", "", dvNow)
	if err := v.Reject("admin1", "", dvNow); !errors.IsCode(err, errors.CodeInvalidArgument) {
		t.Errorf("want CodeInvalidArgument for empty reason, got %v", err)
	}
	if err := v.Reject("admin1", "CCCD mờ, không đọc được", dvNow); err != nil {
		t.Fatalf("reject with reason should succeed: %v", err)
	}
	if v.Status != entity.KYCRejected {
		t.Errorf("status = %v, want rejected", v.Status)
	}
	if v.RejectReason == "" {
		t.Error("reject reason must be persisted")
	}
}

func TestDriverVerification_ResubmitAfterRejectionClearsReason(t *testing.T) {
	v, _ := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "079095001234", "", dvNow)
	_ = v.Reject("admin1", "bad photo", dvNow)

	if err := v.Resubmit("A", dvDOB, "new addr", "079095001234", "GPLX1", dvNow.Add(time.Hour)); err != nil {
		t.Fatalf("resubmit after rejection should succeed: %v", err)
	}
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending after resubmit", v.Status)
	}
	if v.RejectReason != "" || v.RejectedAt != nil {
		t.Error("resubmit must clear prior rejection")
	}
}

func TestDriverVerification_CannotEditWhileUnderReview(t *testing.T) {
	v, _ := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "079095001234", "", dvNow)
	_ = v.StartReview("admin1", dvNow)

	if err := v.Resubmit("B", dvDOB, "addr2", "079095001234", "", dvNow); !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("editing an under-review verification should fail, got %v", err)
	}
}

func TestDriverVerification_CannotResubmitDirectlyFromApproved(t *testing.T) {
	v, _ := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "079095001234", "", dvNow)
	_ = v.Approve("admin1", dvNow)

	if err := v.Resubmit("B", dvDOB, "addr2", "079095001234", "", dvNow); !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("Resubmit must require Invalidate first from Approved, got %v", err)
	}
}

func TestDriverVerification_InvalidateResetsApprovedToPending(t *testing.T) {
	v, _ := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "079095001234", "", dvNow)
	_ = v.Approve("admin1", dvNow)

	v.Invalidate(dvNow.Add(time.Hour))
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want pending after Invalidate", v.Status)
	}
	if v.ApprovedAt != nil {
		t.Error("Invalidate must clear ApprovedAt")
	}
	// Now Resubmit succeeds.
	if err := v.Resubmit("B", dvDOB, "addr2", "079095001234", "", dvNow.Add(2*time.Hour)); err != nil {
		t.Fatalf("resubmit after Invalidate should succeed: %v", err)
	}
}

func TestDriverVerification_InvalidateIsNoOpFromPending(t *testing.T) {
	v, _ := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "079095001234", "", dvNow)
	v.Invalidate(dvNow.Add(time.Hour))
	if v.Status != entity.KYCPending {
		t.Errorf("status = %v, want unchanged pending", v.Status)
	}
	if v.SubmittedAt != dvNow {
		t.Error("Invalidate must be a no-op (including SubmittedAt) from a status with nothing to invalidate")
	}
}

func TestDriverVerification_ExpireOnlyFromApproved(t *testing.T) {
	v, _ := entity.NewDriverVerification("id1", "d1", "A", dvDOB, "addr", "079095001234", "", dvNow)
	if err := v.Expire("CCCD cần xác minh lại", dvNow); !errors.IsCode(err, errors.CodePreconditionFailed) {
		t.Errorf("expiring a non-approved verification should fail, got %v", err)
	}
	_ = v.Approve("admin1", dvNow)
	if err := v.Expire("CCCD cần xác minh lại", dvNow.Add(time.Hour)); err != nil {
		t.Fatalf("expire from approved should succeed: %v", err)
	}
	if v.Status != entity.KYCExpired {
		t.Errorf("status = %v, want expired", v.Status)
	}
	if v.ExpiredAt == nil {
		t.Error("ExpiredAt must be set")
	}
}

package app_test

import (
	"testing"

	"github.com/fairride/user/app"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/infrastructure/ocr"
	"github.com/fairride/user/infrastructure/vision"
)

func TestDecide_RuleEngineFailure_AlwaysRejected(t *testing.T) {
	mode, confidence, reason := app.Decide(
		app.RuleEngineResult{OK: false, Reason: "ảnh mờ"},
		ocr.Result{Confidence: 0.99},
		vision.Result{Confidence: 0.99},
	)
	if mode != entity.ReviewModeRejected {
		t.Fatalf("expected Rejected regardless of AI confidence when the rule engine fails, got %s", mode)
	}
	if reason != "ảnh mờ" {
		t.Fatalf("expected the rule engine's own reason to propagate, got %q", reason)
	}
	if confidence != 0 {
		t.Fatalf("expected confidence 0 on a rule-engine rejection, got %v", confidence)
	}
}

func TestDecide_MockProviders_AlwaysManualReview(t *testing.T) {
	// Both mock providers report 0.5 — must land squarely in ManualReview,
	// never AutoApproved or Rejected, until real providers replace them.
	mode, _, _ := app.Decide(
		app.RuleEngineResult{OK: true},
		ocr.Result{Confidence: 0.5},
		vision.Result{Confidence: 0.5},
	)
	if mode != entity.ReviewModeManualReview {
		t.Fatalf("expected ManualReview for mock-level confidence, got %s", mode)
	}
}

func TestDecide_HighConfidence_AutoApproved(t *testing.T) {
	mode, _, _ := app.Decide(
		app.RuleEngineResult{OK: true},
		ocr.Result{Confidence: 0.98},
		vision.Result{Confidence: 0.97},
	)
	if mode != entity.ReviewModeAutoApproved {
		t.Fatalf("expected AutoApproved for high confidence, got %s", mode)
	}
}

func TestDecide_LowConfidence_Rejected(t *testing.T) {
	mode, _, _ := app.Decide(
		app.RuleEngineResult{OK: true},
		ocr.Result{Confidence: 0.1},
		vision.Result{Confidence: 0.1},
	)
	if mode != entity.ReviewModeRejected {
		t.Fatalf("expected Rejected for low confidence, got %s", mode)
	}
}

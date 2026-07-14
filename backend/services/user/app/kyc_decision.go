package app

import (
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/infrastructure/ocr"
	"github.com/fairride/user/infrastructure/vision"
)

// Auto-approve/reject thresholds. Deliberately conservative — see
// ocr.mockConfidence/vision.mockConfidence's doc comments: the mock
// providers wired up today always return 0.5, comfortably inside
// [rejectConfidenceThreshold, autoApproveConfidenceThreshold), so every
// real submission that passes the Rule Engine lands in ManualReview until
// a real OCR/Vision provider replaces the mocks.
const (
	autoApproveConfidenceThreshold = 0.95
	rejectConfidenceThreshold      = 0.30
)

// Decide combines the Rule Engine result with OCR/Vision confidence into a
// single ReviewMode recommendation. This is the pipeline's actual business
// logic — swapping MockOCRProvider/MockVisionProvider for real providers
// later changes nothing here, only the Confidence values flowing in.
func Decide(rule RuleEngineResult, ocrResult ocr.Result, visionResult vision.Result) (mode entity.ReviewMode, confidence float64, reason string) {
	if !rule.OK {
		return entity.ReviewModeRejected, 0, rule.Reason
	}

	confidence = averageConfidence(ocrResult.Confidence, visionResult.Confidence)

	if confidence >= autoApproveConfidenceThreshold {
		return entity.ReviewModeAutoApproved, confidence, ""
	}
	if confidence < rejectConfidenceThreshold {
		return entity.ReviewModeRejected, confidence, "độ tin cậy xác minh quá thấp"
	}
	return entity.ReviewModeManualReview, confidence, ""
}

func averageConfidence(a, b float64) float64 {
	return (a + b) / 2
}

package ocr

import "context"

// mockConfidence is deliberately in the "manual review" band (see
// user/app.Decide's thresholds) — a mock result must never be trusted
// enough to auto-approve or auto-reject a real rider's KYC. Today, with
// only mock providers wired up, every submission that passes the Rule
// Engine lands in ManualReview, i.e. the existing admin-approval flow is
// completely unchanged (see plan's Rider AI KYC Foundation phase).
const mockConfidence = 0.5

// MockOCRProvider is the only Provider implemented today. It performs no
// real text extraction.
type MockOCRProvider struct{}

func NewMockOCRProvider() *MockOCRProvider { return &MockOCRProvider{} }

func (p *MockOCRProvider) Extract(_ context.Context, _ string, _ string) (Result, error) {
	return Result{
		Success:         true,
		Confidence:      mockConfidence,
		ExtractedFields: map[string]string{},
		Raw:             `{"provider":"mock"}`,
	}, nil
}

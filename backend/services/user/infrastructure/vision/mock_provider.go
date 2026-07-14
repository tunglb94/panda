package vision

import "context"

// mockConfidence mirrors ocr.mockConfidence's rationale exactly — stays in
// the "manual review" band so a mock result can never auto-approve or
// auto-reject a real submission.
const mockConfidence = 0.5

// MockVisionProvider is the only Provider implemented today. It performs
// no real image analysis.
type MockVisionProvider struct{}

func NewMockVisionProvider() *MockVisionProvider { return &MockVisionProvider{} }

func (p *MockVisionProvider) Analyze(_ context.Context, _ string, _ string) (Result, error) {
	return Result{
		Success:    true,
		Confidence: mockConfidence,
		Labels:     []string{},
		Raw:        `{"provider":"mock"}`,
	}, nil
}

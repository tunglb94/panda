// Package ocr defines the OCR abstraction for the Rider KYC AI pipeline
// (Upload -> Rule Engine -> OCR -> Vision -> Decision). The use case layer
// (user/app) only ever depends on the Provider interface — swapping
// MockOCRProvider for PaddleOCR/Qwen2.5-VL/Gemma Vision later means adding
// one new file here and changing a single line in
// gateway/cmd/server/main.go. No use case or Decision logic changes.
package ocr

import "context"

// Result is what one OCR pass over a document image reports.
type Result struct {
	Success bool
	// Confidence is 0.0-1.0 — how sure the OCR engine is about what it
	// extracted. user/app.Decide compares this against threshold constants
	// to choose auto-approve vs manual-review vs reject.
	Confidence float64
	// ExtractedFields is opaque key/value text the OCR engine read off the
	// document (e.g. "id_number", "full_name") — not cross-checked against
	// the rider's submitted form fields in this phase (see plan's Known Gaps).
	ExtractedFields map[string]string
	// Raw is the provider's full raw response, stored verbatim in
	// RiderVerification.OCRResult for audit/debugging.
	Raw string
}

// Provider extracts text from a document image. imagePath is the value
// stored in RiderVerification.CCCDFrontPath/CCCDBackPath (a storage-layer
// relative path, not raw bytes) — a real implementation resolves it via its
// own storage integration, matching the split MockOCRProvider already uses.
type Provider interface {
	Extract(ctx context.Context, imagePath string, documentType string) (Result, error)
}

// Package vision defines the Vision abstraction for the Rider KYC AI
// pipeline (Upload -> Rule Engine -> OCR -> Vision -> Decision). Mirrors
// package ocr's shape exactly — see its doc comment for the swap-later
// rationale (PaddleOCR/Qwen2.5-VL/Gemma Vision).
package vision

import "context"

// Result is what one Vision pass over a document image reports —
// document-authenticity/tampering/liveness-style signals, as opposed to
// ocr.Result's text extraction.
type Result struct {
	Success bool
	// Confidence is 0.0-1.0 — see ocr.Result.Confidence's doc comment; the
	// same threshold logic in user/app.Decide consumes both.
	Confidence float64
	// Labels is opaque classification output (e.g. "document_detected",
	// "face_detected") a real provider would return.
	Labels []string
	// Raw is the provider's full raw response, stored verbatim in
	// RiderVerification.VisionResult for audit/debugging.
	Raw string
}

// Provider analyzes a document image. See ocr.Provider's doc comment for
// why imagePath is a storage-layer path, not raw bytes.
type Provider interface {
	Analyze(ctx context.Context, imagePath string, documentType string) (Result, error)
}

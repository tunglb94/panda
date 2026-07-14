package app

import (
	"bytes"
	"image"
	_ "image/jpeg" // format registration for image.DecodeConfig/Decode
	_ "image/png"  // format registration for image.DecodeConfig/Decode
	"math"
)

// minImageWidth/minImageHeight are deliberately modest — this is a phone
// camera photo of a CCCD, not a scan; the goal is catching obviously
// unusable uploads (thumbnails, screenshots of a blank page), not enforcing
// a print-quality bar.
const (
	minImageWidth  = 400
	minImageHeight = 300

	// blankStdDevThreshold: a real CCCD photo has real photographic
	// variance; a solid-color or near-solid image (blank page, black
	// camera-failure frame) has almost none. Measured on a coarse
	// grayscale sample, not per-pixel — see isLikelyBlank.
	blankStdDevThreshold = 4.0
)

// RuleEngineResult is the first, deterministic stage of the KYC pipeline
// (Upload -> Rule Engine -> OCR -> Vision -> Decision) — no AI involved.
type RuleEngineResult struct {
	OK     bool
	Reason string // Vietnamese, shown directly to the rider — empty when OK
}

func ruleOK() RuleEngineResult { return RuleEngineResult{OK: true} }

func ruleFail(reason string) RuleEngineResult { return RuleEngineResult{OK: false, Reason: reason} }

// RunRuleEngine validates one uploaded document image before it's accepted
// onto disk: decodable, not corrupt, minimum resolution, not blank/solid
// color. documentType is accepted for interface symmetry with the rest of
// the pipeline (OCR/Vision take it too) — the actual cccd_front/cccd_back
// enum check already happens at the handler layer before this runs.
func RunRuleEngine(data []byte, documentType string) RuleEngineResult {
	if len(data) == 0 {
		return ruleFail("file trống")
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return ruleFail("không đọc được ảnh — file có thể bị hỏng hoặc không đúng định dạng")
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width < minImageWidth || height < minImageHeight {
		return ruleFail("ảnh có độ phân giải quá thấp, vui lòng chụp lại")
	}

	if isLikelyBlank(img) {
		return ruleFail("ảnh trắng hoặc đen, vui lòng chụp lại")
	}

	return ruleOK()
}

// isLikelyBlank samples a coarse grid of pixels (not every pixel — a full
// scan is unnecessary for this check and wasteful on a multi-megapixel
// photo) and reports true when their grayscale values have almost no
// spread, i.e. the image is essentially one solid color.
func isLikelyBlank(img image.Image) bool {
	bounds := img.Bounds()
	const gridSize = 16
	stepX := bounds.Dx() / gridSize
	stepY := bounds.Dy() / gridSize
	if stepX == 0 || stepY == 0 {
		return false // image too small to sample meaningfully — let the resolution check handle it
	}

	var sum, sumSquares float64
	var count int
	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			r, g, b, _ := img.At(x, y).RGBA()
			// Standard luma weighting, scaled down from RGBA's 16-bit range.
			gray := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 257
			sum += gray
			sumSquares += gray * gray
			count++
		}
	}
	if count == 0 {
		return false
	}
	mean := sum / float64(count)
	variance := sumSquares/float64(count) - mean*mean
	if variance < 0 {
		variance = 0
	}
	stdDev := math.Sqrt(variance)
	return stdDev < blankStdDevThreshold
}

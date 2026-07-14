package app_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/fairride/user/app"
)

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// validPhoto builds a 640x480 image with real pixel variance (a gradient),
// large enough and varied enough to pass every Rule Engine check.
func validPhoto(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}
	return encodePNG(t, img)
}

func TestRunRuleEngine_EmptyFile(t *testing.T) {
	result := app.RunRuleEngine([]byte{}, "cccd_front")
	if result.OK {
		t.Fatal("expected empty file to fail")
	}
}

func TestRunRuleEngine_CorruptFile(t *testing.T) {
	result := app.RunRuleEngine([]byte("this is not an image"), "cccd_front")
	if result.OK {
		t.Fatal("expected corrupt/non-image data to fail")
	}
}

func TestRunRuleEngine_TooSmallResolution(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 5), G: uint8(y * 5), B: 100, A: 255})
		}
	}
	result := app.RunRuleEngine(encodePNG(t, img), "cccd_front")
	if result.OK {
		t.Fatal("expected below-minimum resolution to fail")
	}
}

func TestRunRuleEngine_BlankImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, color.White)
		}
	}
	result := app.RunRuleEngine(encodePNG(t, img), "cccd_front")
	if result.OK {
		t.Fatal("expected an all-white (blank) image to fail")
	}
}

func TestRunRuleEngine_ValidPhoto_Passes(t *testing.T) {
	result := app.RunRuleEngine(validPhoto(t), "cccd_front")
	if !result.OK {
		t.Fatalf("expected a valid, varied, large-enough photo to pass, got reason: %s", result.Reason)
	}
}

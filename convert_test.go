package main

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestConvertOutputIsJPEG(t *testing.T) {
	src := readTestFile(t, "./testdata/oyaki.jpg")
	buf, err := convert(src, 90)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if !isValidJPEG(buf.Bytes()) {
		t.Errorf("output is not a valid JPEG (first bytes: %x)", buf.Bytes()[:2])
	}
}

func TestConvertOutputSmallerThanInput(t *testing.T) {
	src := readTestFile(t, "./testdata/oyaki.jpg")
	buf, err := convert(src, 90)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if buf.Len() >= len(src) {
		t.Errorf("output size %d is not smaller than input size %d", buf.Len(), len(src))
	}
}

func TestConvertQuality(t *testing.T) {
	src := readTestFile(t, "./testdata/oyaki.jpg")

	// convert() uses the package-level `quality` variable, not the `q` parameter.
	origQuality := quality
	defer func() { quality = origQuality }()

	quality = 90
	buf90, err := convert(src, 90)
	if err != nil {
		t.Fatalf("convert(90) failed: %v", err)
	}

	quality = 30
	buf30, err := convert(src, 30)
	if err != nil {
		t.Fatalf("convert(30) failed: %v", err)
	}

	if buf30.Len() >= buf90.Len() {
		t.Errorf("quality=30 output (%d bytes) should be smaller than quality=90 output (%d bytes)", buf30.Len(), buf90.Len())
	}
}

func TestConvertAutoRotate(t *testing.T) {
	src := readTestFile(t, "./testdata/oyaki.jpg")
	buf, err := convert(src, 90)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	out := buf.Bytes()
	img, err := vips.NewImageFromBuffer(out)
	if err != nil {
		t.Fatalf("vips.NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	if img.Width() == 0 || img.Height() == 0 {
		t.Errorf("image size is zero: %dx%d", img.Width(), img.Height())
	}

	// oyaki.jpg is already orientation=1 (upper-left), so it should remain 1 after AutoRotate.
	if orient := img.Orientation(); orient != 1 {
		t.Errorf("expected orientation 1, got %d", orient)
	}
}

// TestConvertAutoRotateExif は EXIF orientation=6 (90度時計回り) の画像が
// AutoRotate によって正しく補正されることを確認する。
// 入力: 600x800 + orientation=6 → 出力: 800x600 + orientation=1
func TestConvertAutoRotateExif(t *testing.T) {
	src := readTestFile(t, "./testdata/color_pattern_rot6.jpg")

	origImg, err := vips.NewImageFromBuffer(src)
	if err != nil {
		t.Fatalf("original image: %v", err)
	}
	defer origImg.Close()
	if origImg.Orientation() != 6 {
		t.Fatalf("precondition: expected orientation=6, got %d", origImg.Orientation())
	}

	buf, err := convert(src, 90)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}

	outImg, err := vips.NewImageFromBuffer(buf.Bytes())
	if err != nil {
		t.Fatalf("output image: %v", err)
	}
	defer outImg.Close()

	// AutoRotate 後は orientation が 1 (upper-left) になること
	if outImg.Orientation() != 1 {
		t.Errorf("orientation after AutoRotate = %d, want 1", outImg.Orientation())
	}
	// 600x800 + orientation=6 → AutoRotate で 800x600 になること
	if outImg.Width() != 800 || outImg.Height() != 600 {
		t.Errorf("size after AutoRotate = %dx%d, want 800x600", outImg.Width(), outImg.Height())
	}
}

// TestConvertColorPatternOutputIsJPEG は著作権フリーのカラーパターン画像でも
// JPEG 変換が正常動作することを確認する。
func TestConvertColorPatternOutputIsJPEG(t *testing.T) {
	src := readTestFile(t, "./testdata/color_pattern.jpg")
	buf, err := convert(src, 90)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !isValidJPEG(buf.Bytes()) {
		t.Errorf("output is not a valid JPEG")
	}
}

func BenchmarkConvert(b *testing.B) {
	src := readTestFile(b, "./testdata/oyaki.jpg")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := convert(src, 90); err != nil {
			b.Fatalf("convert failed: %v", err)
		}
	}
}

func BenchmarkConvertColorPattern(b *testing.B) {
	src := readTestFile(b, "./testdata/color_pattern.jpg")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := convert(src, 90); err != nil {
			b.Fatalf("convert failed: %v", err)
		}
	}
}

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestProxyWebP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))

	orgSrvURL = origin.URL
	url := ts.URL + "/oyaki.jpg.webp"

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := doWebp(req)
	if err != nil {
		t.Fatal(err)
	} else {
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
	// match with origin file info
	if resp.Header.Get("Content-Type") != "image/jpeg" {
		t.Error("wrong header Content-Type")
		t.Error(resp.Header)
	}
}

func TestConvJPG2WebP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))

	orgSrvURL = origin.URL
	url := ts.URL + "/oyaki.jpg.webp"

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := doWebp(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	srcBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_, err = convWebp(srcBytes, 90)
	if err != nil {
		t.Fatal(err)
	}

}

func TestConvWebpOutputIsWebP(t *testing.T) {
	src := readTestFile(t, "./testdata/oyaki.jpg")
	buf, err := convWebp(src, 90)
	if err != nil {
		t.Fatalf("convWebp failed: %v", err)
	}
	if !isValidWebP(buf.Bytes()) {
		b := buf.Bytes()
		if len(b) >= 12 {
			t.Errorf("output is not a valid WebP: first 12 bytes: %x", b[:12])
		} else {
			t.Errorf("output is not a valid WebP: data too short (%d bytes)", len(b))
		}
	}
}

func TestConvWebpStripMetadata(t *testing.T) {
	src := readTestFile(t, "./testdata/oyaki.jpg")
	buf, err := convWebp(src, 90)
	if err != nil {
		t.Fatalf("convWebp failed: %v", err)
	}

	img, err := vips.NewImageFromBuffer(buf.Bytes())
	if err != nil {
		t.Fatalf("vips.NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()

	// StripMetadata: true means EXIF orientation tag should be absent (0).
	if orient := img.Orientation(); orient != 0 {
		t.Errorf("expected EXIF orientation to be stripped (0), got %d", orient)
	}
}

// TestConvWebpColorPatternOutputIsWebP は color_pattern.jpg でも WebP 変換が動くことを確認する。
func TestConvWebpColorPatternOutputIsWebP(t *testing.T) {
	src := readTestFile(t, "./testdata/color_pattern.jpg")
	buf, err := convWebp(src, 90)
	if err != nil {
		t.Fatalf("convWebp: %v", err)
	}
	if !isValidWebP(buf.Bytes()) {
		t.Errorf("output is not a valid WebP")
	}
}

// TestConvWebpColorPatternRot6AutoRotate は orientation=6 の画像を convWebp に渡したとき
// NoAutoRotate: false によって自動補正され、800x600 の WebP になることを確認する。
func TestConvWebpColorPatternRot6AutoRotate(t *testing.T) {
	src := readTestFile(t, "./testdata/color_pattern_rot6.jpg")

	origImg, err := vips.NewImageFromBuffer(src)
	if err != nil {
		t.Fatalf("original image: %v", err)
	}
	defer origImg.Close()
	if origImg.Orientation() != 6 {
		t.Fatalf("precondition: expected orientation=6, got %d", origImg.Orientation())
	}

	buf, err := convWebp(src, 90)
	if err != nil {
		t.Fatalf("convWebp: %v", err)
	}
	if !isValidWebP(buf.Bytes()) {
		t.Errorf("output is not a valid WebP")
	}

	outImg, err := vips.NewImageFromBuffer(buf.Bytes())
	if err != nil {
		t.Fatalf("output image: %v", err)
	}
	defer outImg.Close()
	// 600x800 + orientation=6 → AutoRotate で 800x600 になること
	if outImg.Width() != 800 || outImg.Height() != 600 {
		t.Errorf("size after convWebp = %dx%d, want 800x600", outImg.Width(), outImg.Height())
	}
}

func BenchmarkConvWebP(b *testing.B) {
	src := readTestFile(b, "./testdata/oyaki.jpg")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := convWebp(src, 90); err != nil {
			b.Fatalf("convWebp failed: %v", err)
		}
	}
}

func BenchmarkConvJPG2WebP(b *testing.B) {
	f, err := os.Open("./testdata/oyaki.jpg")
	if err != nil {
		b.Fatal("failed to open testdata")
	}
	defer f.Close()

	// to re-use src bytes
	src, err := io.ReadAll(f)
	if err != nil {
		b.Fatal("failed to open testdata")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = convWebp(src, 90); err != nil {
			b.Fail()
		}
	}
}

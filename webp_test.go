package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/h2non/bimg"
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

	meta, err := bimg.NewImage(buf.Bytes()).Metadata()
	if err != nil {
		t.Fatalf("bimg.Metadata failed: %v", err)
	}

	// StripMetadata: true means EXIF should be empty/zeroed.
	exif := meta.EXIF
	empty := bimg.EXIF{}
	if exif != empty {
		t.Errorf("expected EXIF to be empty after StripMetadata, got %+v", exif)
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

	origMeta, err := bimg.NewImage(src).Metadata()
	if err != nil {
		t.Fatalf("original metadata: %v", err)
	}
	if origMeta.Orientation != 6 {
		t.Fatalf("precondition: expected orientation=6, got %d", origMeta.Orientation)
	}

	buf, err := convWebp(src, 90)
	if err != nil {
		t.Fatalf("convWebp: %v", err)
	}
	if !isValidWebP(buf.Bytes()) {
		t.Errorf("output is not a valid WebP")
	}

	outMeta, err := bimg.NewImage(buf.Bytes()).Metadata()
	if err != nil {
		t.Fatalf("output metadata: %v", err)
	}
	// 600x800 + orientation=6 → AutoRotate で 800x600 になること
	if outMeta.Size.Width != 800 || outMeta.Size.Height != 600 {
		t.Errorf("size after convWebp = %dx%d, want 800x600", outMeta.Size.Width, outMeta.Size.Height)
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

func BenchmarkConvJPG2WebP_bimg(b *testing.B) {
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

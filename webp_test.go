package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/disintegration/imaging"
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
	_, err = convWebp(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

}

func BenchmarkConvJPG2WebP_OldCwebpMethod(b *testing.B) {
	tmpF, err := os.CreateTemp("/tmp", "")
	if err != nil {
		b.Fatal("failed to create tmp file")
	}
	defer tmpF.Close()
	defer os.Remove(tmpF.Name())

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
		b.StopTimer()
		srcBuf := bytes.NewBuffer(src)
		b.StartTimer()
		img, err := imaging.Decode(srcBuf, imaging.AutoOrientation(true))
		if err != nil {
			b.Fail()
		}

		if err := imaging.Encode(tmpF, img, imaging.JPEG); err != nil {
			b.Fail()
		}

		params := []string{"-quiet", "-mt", "-jpeg_like", f.Name(), "-o", "-"}
		if _, err = exec.Command("cwebp", params...).Output(); err != nil {
			b.Fail()
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
		b.StopTimer()
		srcBuf := bytes.NewBuffer(src)
		b.StartTimer()
		if _, err = convWebp(srcBuf); err != nil {
			b.Fail()
		}
	}
}

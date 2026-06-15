package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestRoot(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	proxy(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got http %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequestHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	cxff := "127.0.0.1"
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")

		sxff := r.Header.Get("X-Forwarded-For")
		if sxff != cxff {
			t.Errorf("X-Forwarded-For header is %s, want %s", sxff, cxff)
		}
	}))

	orgSrvURL = origin.URL

	req, err := http.NewRequest("GET", ts.URL+"/oyaki.jpg", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Forwarded-For", cxff)
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
}

func TestProxyJPEG(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))

	orgSrvURL = origin.URL

	url := ts.URL + "/oyaki.jpg"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	orgRes, err := http.Get(orgSrvURL)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("HTTP status is %d, want %d", res.StatusCode, http.StatusOK)
	}

	if res.ContentLength < 0 {
		t.Errorf("Content-Length header does not exist")
	}

	if res.ContentLength >= orgRes.ContentLength {
		t.Errorf("value of Content-Length header %d is larger than origin's one %d", res.ContentLength, orgRes.ContentLength)
	}
}

func TestOriginNotModifiedJPEG(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", "2023-01-01T00:00:00")
		w.WriteHeader(http.StatusNotModified)
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))

	orgSrvURL = origin.URL

	url := ts.URL + "/oyaki.jpg"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get(orgSrvURL)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotModified {
		t.Errorf("HTTP status is %d, want %d", res.StatusCode, http.StatusNotModified)
	}

	if res.ContentLength < 0 {
		t.Errorf("Content-Length header does not exist")
	}
}

func TestProxyPNG(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/corn.png")
	}))

	orgSrvURL = origin.URL
	url := ts.URL + "/corn.png"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	orgRes, err := http.Get(orgSrvURL)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("HTTP status is %d, want %d", res.StatusCode, http.StatusOK)
	}

	if res.ContentLength < 0 {
		t.Errorf("Content-Length header does not exist")
	}

	if res.ContentLength != orgRes.ContentLength {
		t.Errorf("value of Content-Length header %d is not equal to origin's one, want %d", res.ContentLength, orgRes.ContentLength)
	}
}

func TestOriginNotExist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}))

	orgSrvURL = origin.URL

	url := ts.URL + "/none.jpg"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("HTTP status is %d, want %d", res.StatusCode, http.StatusNotFound)
	}
}

func TestOriginForbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	}))

	orgSrvURL = origin.URL

	url := ts.URL + "/forbidden.jpg"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusForbidden {
		t.Errorf("HTTP status is %d, want %d", res.StatusCode, http.StatusNotFound)
	}
}

func TestOriginBadGateWay(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
	}))

	orgSrvURL = origin.URL

	url := ts.URL + "/bad.jpg"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusBadGateway {
		t.Errorf("HTTP status is %d, want %d", res.StatusCode, http.StatusBadGateway)
	}
}

func TestOriginNotModifiedPNG(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", "2023-01-01T00:00:00")
		w.WriteHeader(http.StatusNotModified)
		http.ServeFile(w, r, "./testdata/corn.png")
	}))

	orgSrvURL = origin.URL
	url := ts.URL + "/corn.png"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get(orgSrvURL)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotModified {
		t.Errorf("HTTP status is %d, want %d", res.StatusCode, http.StatusNotModified)
	}

	if res.Header.Get("Last-Modified") == "" {
		t.Errorf("Not set header")
	}

	if res.ContentLength < 0 {
		t.Errorf("Content-Length header does not exist")
	}
}

func TestProxyJPEGContentType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	res, err := http.Get(ts.URL + "/oyaki.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	ct := res.Header.Get("Content-Type")
	if ct != "image/jpeg" {
		t.Errorf("Content-Type is %q, want %q", ct, "image/jpeg")
	}
}

func TestProxyPNGContentType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/corn.png")
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	res, err := http.Get(ts.URL + "/corn.png")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	ct := res.Header.Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("Content-Type is %q, want %q", ct, "image/png")
	}
}

func TestProxyWebPEndToEnd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	res, err := http.Get(ts.URL + "/oyaki.jpg.webp")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("HTTP status is %d, want %d", res.StatusCode, http.StatusOK)
	}

	ct := res.Header.Get("Content-Type")
	if ct != "image/webp" {
		t.Errorf("Content-Type is %q, want %q", ct, "image/webp")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !isValidWebP(body) {
		t.Errorf("response body is not a valid WebP image")
	}
}

func TestProxyContentLength(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	res, err := http.Get(ts.URL + "/oyaki.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if res.ContentLength < 0 {
		t.Errorf("Content-Length header is missing")
		return
	}
	if int(res.ContentLength) != len(body) {
		t.Errorf("Content-Length %d does not match actual body length %d", res.ContentLength, len(body))
	}
}

func TestProxyIfModifiedSince(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	ifModifiedSince := "Mon, 01 Jan 2024 00:00:00 GMT"
	receivedHeader := ""

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("If-Modified-Since")
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	req, err := http.NewRequest("GET", ts.URL+"/oyaki.jpg", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("If-Modified-Since", ifModifiedSince)

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	io.ReadAll(res.Body)

	if receivedHeader != ifModifiedSince {
		t.Errorf("origin received If-Modified-Since %q, want %q", receivedHeader, ifModifiedSince)
	}
}

func TestProxyLastModifiedFromOrigin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	lastModified := "Mon, 01 Jan 2024 00:00:00 GMT"

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("./testdata/oyaki.jpg")
		if err != nil {
			http.Error(w, "read failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Last-Modified", lastModified)
		w.Write(data)
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	res, err := http.Get(ts.URL + "/oyaki.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	io.ReadAll(res.Body)

	got := res.Header.Get("Last-Modified")
	if got != lastModified {
		t.Errorf("Last-Modified is %q, want %q", got, lastModified)
	}
}

func TestProxyLastModifiedAutoSet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("./testdata/oyaki.jpg")
		if err != nil {
			http.Error(w, "read failed", http.StatusInternalServerError)
			return
		}
		// No Last-Modified header set — proxy should auto-set it.
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(data)
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	before := time.Now().UTC().Add(-2 * time.Second)

	res, err := http.Get(ts.URL + "/oyaki.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	io.ReadAll(res.Body)

	lm := res.Header.Get("Last-Modified")
	if lm == "" {
		t.Errorf("Last-Modified header is missing; proxy should auto-set it")
		return
	}

	parsed, err := time.Parse(http.TimeFormat, lm)
	if err != nil {
		t.Errorf("Last-Modified %q cannot be parsed: %v", lm, err)
		return
	}
	if parsed.Before(before) {
		t.Errorf("auto-set Last-Modified %q is too far in the past", lm)
	}
}

func TestProxyOriginErrors(t *testing.T) {
	cases := []struct {
		name           string
		originStatus   int
		expectedStatus int
	}{
		{"404 Not Found", http.StatusNotFound, http.StatusNotFound},
		{"403 Forbidden", http.StatusForbidden, http.StatusForbidden},
		{"502 Bad Gateway", http.StatusBadGateway, http.StatusBadGateway},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(proxy))
			defer ts.Close()

			origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, tc.name, tc.originStatus)
			}))
			defer origin.Close()

			orgSrvURL = origin.URL

			res, err := http.Get(ts.URL + "/test.jpg")
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()

			if res.StatusCode != tc.expectedStatus {
				t.Errorf("HTTP status is %d, want %d", res.StatusCode, tc.expectedStatus)
			}
		})
	}
}

func BenchmarkProxyWebP(b *testing.B) {
	b.ResetTimer()
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))

	orgSrvURL = origin.URL

	url := ts.URL + "/oyaki.jpg.webp"

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		c := new(http.Client)
		resp, err := c.Do(req)
		if err != nil {
			b.Fatal(err)
		} else {
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
}

func BenchmarkProxyJpeg(b *testing.B) {
	b.ResetTimer()
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))

	orgSrvURL = origin.URL

	url := ts.URL + "/oyaki.jpg"

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		client := new(http.Client)
		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		} else {
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
}

func BenchmarkProxyPNG(b *testing.B) {
	b.ResetTimer()
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/corn.png")
	}))

	orgSrvURL = origin.URL
	url := ts.URL + "/corn.png"

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		client := new(http.Client)
		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		} else {
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
}

// TestProxyColorPatternWebPEndToEnd は color_pattern.jpg を .webp でリクエストして
// 有効な WebP が返ることを確認する。
func TestProxyColorPatternWebPEndToEnd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/color_pattern.jpg")
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	res, err := http.Get(ts.URL + "/color_pattern.jpg.webp")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); ct != "image/webp" {
		t.Errorf("Content-Type = %q, want %q", ct, "image/webp")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !isValidWebP(body) {
		t.Errorf("response body is not valid WebP")
	}
}

// TestProxyColorPatternRot6WebP は orientation=6 の JPEG を .webp でリクエストしたとき
// proxy が AutoRotate まで含めて正しく処理し、800x600 の WebP を返すことを確認する。
func TestProxyColorPatternRot6WebP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))
	defer ts.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/color_pattern_rot6.jpg")
	}))
	defer origin.Close()

	orgSrvURL = origin.URL

	res, err := http.Get(ts.URL + "/color_pattern_rot6.jpg.webp")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); ct != "image/webp" {
		t.Errorf("Content-Type = %q, want %q", ct, "image/webp")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !isValidWebP(body) {
		t.Errorf("response body is not valid WebP")
	}

	// orientation=6 (600x800) → AutoRotate → 800x600 になること
	img, err := vips.NewImageFromBuffer(body)
	if err != nil {
		t.Fatalf("vips.NewImageFromBuffer failed: %v", err)
	}
	defer img.Close()
	if img.Width() != 800 || img.Height() != 600 {
		t.Errorf("size = %dx%d, want 800x600", img.Width(), img.Height())
	}
}

func BenchmarkProxyWebPColorPattern(b *testing.B) {
	b.ResetTimer()
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/color_pattern.jpg")
	}))

	orgSrvURL = origin.URL
	url := ts.URL + "/color_pattern.jpg.webp"

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", url, nil)
		c := new(http.Client)
		resp, err := c.Do(req)
		if err != nil {
			b.Fatal(err)
		} else {
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
}

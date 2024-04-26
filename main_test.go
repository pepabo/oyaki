package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoot(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ph := &ProxyHandler{}
	ph.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got http %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequestHeader(t *testing.T) {
	cxff := "127.0.0.1"
	origin, ts := setupOriginAndOyaki(
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./testdata/oyaki.jpg")

			sxff := r.Header.Get("X-Forwarded-For")
			if sxff != cxff {
				t.Errorf("X-Forwarded-For header is %s, want %s", sxff, cxff)
			}
		})
	defer ts.Close()
	defer origin.Close()

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
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	})
	defer ts.Close()
	defer origin.Close()

	url := ts.URL + "/oyaki.jpg"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	orgRes, err := http.Get(origin.URL)
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
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", "2023-01-01T00:00:00")
		w.WriteHeader(http.StatusNotModified)
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	})
	defer ts.Close()
	defer origin.Close()

	url := ts.URL + "/oyaki.jpg"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get(origin.URL)
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
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/corn.png")
	})
	defer ts.Close()
	defer origin.Close()

	url := ts.URL + "/corn.png"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	orgRes, err := http.Get(origin.URL)
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
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	})
	defer ts.Close()
	defer origin.Close()

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
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	})
	defer ts.Close()
	defer origin.Close()

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
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
	})
	defer ts.Close()
	defer origin.Close()

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
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", "2023-01-01T00:00:00")
		w.WriteHeader(http.StatusNotModified)
		http.ServeFile(w, r, "./testdata/corn.png")
	})
	defer ts.Close()
	defer origin.Close()

	url := ts.URL + "/corn.png"

	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get(origin.URL)
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

func setupOriginAndOyaki(
	originHandler func(http.ResponseWriter, *http.Request),
) (*httptest.Server, *httptest.Server) {
	origin := httptest.NewServer(http.HandlerFunc(originHandler))
	originServerURL := origin.URL
	oyakiHandler := &ProxyHandler{
		originConfig: OriginConfig{
			ServerURL: originServerURL,
		},
		logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	oyaki := httptest.NewServer(oyakiHandler)
	return origin, oyaki
}

func BenchmarkProxyJpeg(b *testing.B) {
	b.ResetTimer()

	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	})
	defer ts.Close()
	defer origin.Close()

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

	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/corn.png")
	})
	defer ts.Close()
	defer origin.Close()

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

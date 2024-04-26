package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

func TestProxyWebP(t *testing.T) {
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	})
	defer ts.Close()
	defer origin.Close()

	url := ts.URL + "/oyaki.jpg.webp"

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := doWebp(slog.New(slog.NewTextHandler(io.Discard, nil)), req)
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
	origin, ts := setupOriginAndOyaki(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	})
	defer ts.Close()
	defer origin.Close()

	url := ts.URL + "/oyaki.jpg.webp"
	nopLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := doWebp(nopLogger, req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = convWebp(context.Background(), nopLogger, resp.Body, []string{})
	if err != nil {
		t.Fatal(err)
	}

}

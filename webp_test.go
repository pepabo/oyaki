package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyWebP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(proxy))

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testdata/oyaki.jpg")
	}))

	orgSrvURL = origin.URL
	url := ts.URL + "/oyaki.jpg.webp"

	req, _ := http.NewRequest("GET", url, nil)

	ctx := context.Background()
	resp, err := doWebp(ctx, req)
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
	ctx := context.Background()
	resp, err := doWebp(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = convWebp(ctx, resp.Body, []string{})
	if err != nil {
		t.Fatal(err)
	}

}

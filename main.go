package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"log/slog"

	"github.com/fujiwara/ridge"
)

type OriginConfig struct {
	ServerURL string
}

var client http.Client
var version = ""

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logLevel := slog.LevelInfo
	if v := os.Getenv("OYAKI_LOGLEVEL"); v != "" {
		level := strings.ToLower(v)
		switch level {
		case "debug":
			logLevel = slog.LevelDebug
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		}
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	var ver bool

	flag.BoolVar(&ver, "version", false, "show version")
	flag.Parse()

	if ver {
		fmt.Printf("oyaki %s\n", getVersion())
		return
	}

	ph := &ProxyHandler{
		logger: logger,
	}
	orgScheme := os.Getenv("OYAKI_ORIGIN_SCHEME")
	orgHost := os.Getenv("OYAKI_ORIGIN_HOST")
	if orgScheme == "" {
		orgScheme = "https"
	}
	ph.originConfig = OriginConfig{
		ServerURL: orgScheme + "://" + orgHost,
	}

	if q := os.Getenv("OYAKI_QUALITY"); q != "" {
		quality, err := strconv.Atoi(q)
		if err != nil {
			// defaulting
			quality = 90
		}
		ph.Quality = quality
	}

	logger.InfoContext(ctx, "starting oyaki", "version", getVersion())
	mux := http.NewServeMux()
	mux.Handle("/", ph)
	ridge.RunWithContext(ctx, ":8080", "/", mux)
}

type ProxyHandler struct {
	logger *slog.Logger
	// originConfigは環境変数で一度読み込まれたあとは書き換わらない
	originConfig OriginConfig
	Quality      int
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.RequestURI()
	if path == "/" {
		fmt.Fprintln(w, "Oyaki lives!")
		return
	}

	orgURL, err := url.Parse(ph.originConfig.ServerURL + path)
	if err != nil {
		http.Error(w, "Invalid origin URL", http.StatusBadRequest)
		ph.logger.ErrorContext(r.Context(), "Invalid origin URL", "error", err)
		return
	}

	req, err := http.NewRequest("GET", orgURL.String(), nil)
	if err != nil {
		http.Error(w, "Request Failed", http.StatusInternalServerError)
		ph.logger.ErrorContext(r.Context(), "Request Failed", "error", err)
		return
	}
	req.Header.Set("User-Agent", "oyaki")

	if r.Header.Get("If-Modified-Since") != "" {
		req.Header.Set("If-Modified-Since", r.Header.Get("If-Modified-Since"))
	}

	xff := r.Header.Get("X-Forwarded-For")
	if len(xff) > 1 {
		req.Header.Set("X-Forwarded-For", xff)
	}
	var orgRes *http.Response
	pathExt := filepath.Ext(req.URL.Path)
	if pathExt == ".webp" {
		orgRes, err = doWebp(ph.logger, req)
	} else {
		orgRes, err = client.Do(req)
	}

	if err != nil {
		http.Error(w, "Get origin failed", http.StatusForbidden)
		ph.logger.ErrorContext(r.Context(), "Get origin failed", "error", err)
		return
	}

	if orgRes.StatusCode == http.StatusNotFound || orgRes.StatusCode == http.StatusForbidden {
		http.Error(w, "Get origin failed", orgRes.StatusCode)
		ph.logger.ErrorContext(r.Context(), "Get origin failed", "error", err)
		return
	}

	defer orgRes.Body.Close()
	if orgRes.Header.Get("Last-Modified") != "" {
		w.Header().Set("Last-Modified", orgRes.Header.Get("Last-Modified"))
	} else {
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	}

	if orgRes.StatusCode == http.StatusNotModified {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if orgRes.StatusCode != http.StatusOK {
		http.Error(w, "Get origin failed", http.StatusBadGateway)
		ph.logger.ErrorContext(r.Context(), "Get origin failed", "status", orgRes.Status)
		return
	}

	ct := orgRes.Header.Get("Content-Type")
	cl := orgRes.Header.Get("Content-Length")

	if ct != "image/jpeg" {
		w.Header().Set("Content-Type", ct)
		if cl != "" {
			w.Header().Set("Content-Length", cl)
		}

		_, err := io.Copy(w, orgRes.Body)
		if err != nil {
			// ignore already close client.
			if !errors.Is(err, syscall.EPIPE) {
				http.Error(w, "Read origin body failed", http.StatusInternalServerError)
				ph.logger.ErrorContext(r.Context(), "Read origin body failed", "error", err)
			}
		}
		return
	}
	var buf *bytes.Buffer
	if pathExt == ".webp" {
		resBytes, err := io.ReadAll(orgRes.Body)
		if err != nil {
			http.Error(w, "Read origin body failed", http.StatusInternalServerError)
			ph.logger.ErrorContext(r.Context(), "Read origin body failed", "error", err)
			return
		}

		body := io.NopCloser(bytes.NewBuffer(resBytes))
		defer body.Close()
		buf, err = convWebp(r.Context(), ph.logger, body, []string{})
		if err == nil {
			defer buf.Reset()
			w.Header().Set("Content-Type", "image/webp")
		} else {
			// if err, normally convertion will be proceeded
			body = io.NopCloser(bytes.NewBuffer(resBytes))
			defer body.Close()
			buf, err = convert(body, ph.Quality)
			if err != nil {
				http.Error(w, "Image convert failed", http.StatusInternalServerError)
				ph.logger.ErrorContext(r.Context(), "Image convert failed", "error", err)
				return
			}
			defer buf.Reset()
			w.Header().Set("Content-Type", "image/jpeg")
		}
	} else {
		buf, err = convert(orgRes.Body, ph.Quality)
		if err != nil {
			http.Error(w, "Image convert failed", http.StatusInternalServerError)
			ph.logger.ErrorContext(r.Context(), "Image convert failed", "error", err)
			return
		}
		defer buf.Reset()
		w.Header().Set("Content-Type", "image/jpeg")
	}
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	if _, err := io.Copy(w, buf); err != nil {
		// ignore already close client.
		if !errors.Is(err, syscall.EPIPE) {
			http.Error(w, "Write responce failed", http.StatusInternalServerError)
			ph.logger.ErrorContext(r.Context(), "write response failed", "error", err)
		}
	}
}

func getVersion() string {
	if version != "" {
		return version
	}

	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}
	return i.Main.Version
}

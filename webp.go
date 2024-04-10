package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/disintegration/imaging"
)

func doWebp(logger *slog.Logger, req *http.Request) (*http.Response, error) {
	var orgRes *http.Response
	orgURL := req.URL
	newPath := orgURL.Path[:len(orgURL.Path)-len(".webp")]
	newOrgURL, err := url.Parse(fmt.Sprintf("%s://%s%s?%s", orgURL.Scheme, orgURL.Host, newPath, orgURL.RawQuery))
	if err != nil {
		logger.ErrorContext(req.Context(), "failed to parse URL", "error", err)
		return nil, err
	}
	newReq, err := http.NewRequest("GET", newOrgURL.String(), nil)
	if err != nil {
		logger.ErrorContext(req.Context(), "failed to create new request", "error", err)
		return nil, err
	}
	newReq.Header = req.Header

	orgRes, err = client.Do(newReq)
	if err != nil {
		logger.ErrorContext(req.Context(), "failed to send request to origin", "error", err)
		return nil, err
	}
	if orgRes.StatusCode != 200 && orgRes.StatusCode != 304 {
		logger.ErrorContext(req.Context(), "origin response status code must be 200 or 304", "status", orgRes.Status)
		return nil, fmt.Errorf("origin response is not 200 or 304")
	}

	if orgRes == nil {
		return nil, fmt.Errorf("origin response is not found")
	}
	return orgRes, nil
}

func convWebp(
	ctx context.Context,
	logger *slog.Logger,
	src io.Reader,
	params []string,
) (*bytes.Buffer, error) {
	f, err := os.CreateTemp("/tmp", "")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	img, err := imaging.Decode(src, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	if err := imaging.Encode(f, img, imaging.JPEG); err != nil {
		return nil, err
	}

	params = append(params, "-quiet", "-mt", "-jpeg_like", f.Name(), "-o", "-")
	out, err := exec.Command("cwebp", params...).Output()
	if err != nil {
		logger.ErrorContext(ctx, "failed to convert to webp with cwebp", "error", err)
		return nil, err
	}
	return bytes.NewBuffer(out), nil
}

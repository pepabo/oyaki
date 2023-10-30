package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"go.opentelemetry.io/otel/trace"
)

func doWebp(ctx context.Context, req *http.Request) (*http.Response, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "doWebp")
	defer span.End()

	var orgRes *http.Response
	orgURL := req.URL
	newPath := orgURL.Path[:len(orgURL.Path)-len(".webp")]
	newOrgURL, err := url.Parse(fmt.Sprintf("%s://%s%s?%s", orgURL.Scheme, orgURL.Host, newPath, orgURL.RawQuery))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	newReq, err := http.NewRequest("GET", newOrgURL.String(), nil)
	newReq.Header = req.Header
	if err != nil {
		log.Println(err)
		return nil, err
	}
	orgRes, err = client.Do(newReq)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if orgRes.StatusCode != 200 && orgRes.StatusCode != 304 {
		log.Println(orgRes.Status)
		return nil, err
	}
	return orgRes, nil
}

func convWebp(ctx context.Context, src io.Reader, params []string) (*bytes.Buffer, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "convWebp")
	defer span.End()

	f, err := os.CreateTemp("/tmp", "")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	_, err = io.Copy(f, src)
	if err != nil {
		return nil, err
	}
	params = append(params, "-quiet", "-mt", "-jpeg_like", f.Name(), "-o", "-")
	out, err := exec.Command("cwebp", params...).Output()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return bytes.NewBuffer(out), nil
}

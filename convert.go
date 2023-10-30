package main

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"

	"go.opentelemetry.io/otel/trace"
)

func convert(ctx context.Context, src io.Reader, q int) (*bytes.Buffer, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "convert")
	defer span.End()

	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}

	return buf, nil
}

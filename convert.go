package main

import (
	"bytes"
	"context"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
	"go.opentelemetry.io/otel/trace"
)

func convert(ctx context.Context, src io.Reader, q int) (*bytes.Buffer, error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "convert")
	defer span.End()

	img, err := imaging.Decode(src, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}

	return buf, nil
}

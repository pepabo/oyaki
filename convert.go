package main

import (
	"bytes"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
)

func convert(src io.Reader, q int) (*bytes.Buffer, error) {
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

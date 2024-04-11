package main

import (
	"bytes"
	"io"

	"github.com/h2non/bimg"
)

func convert(src io.Reader, q int) (*bytes.Buffer, error) {
	out, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	img, err := bimg.NewImage(out).AutoRotate()
	if err != nil {
		return nil, err
	}

	processed, err := bimg.NewImage(img).Process(bimg.Options{Quality: quality})
	if err != nil {
		return nil, err
	}
	jpegImg, err := bimg.NewImage(processed).Convert(bimg.JPEG)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(jpegImg), nil
}

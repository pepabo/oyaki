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

	opts := bimg.Options{
		Type:    bimg.JPEG,
		Quality: quality,
		// NoAutoRotateはデフォルトでfalseで、勝手にrotateしてくれる
	}
	jpegImg, err := bimg.NewImage(out).Process(opts)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(jpegImg), nil
}

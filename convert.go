package main

import (
	"bytes"

	"github.com/davidbyttow/govips/v2/vips"
)

func convert(src []byte, q int) (*bytes.Buffer, error) {
	img, err := vips.NewImageFromBuffer(src)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	// 動作検証の結果こちらは明示的にAutoRotateしないと動かなかった
	if err := img.AutoRotate(); err != nil {
		return nil, err
	}

	ep := vips.NewJpegExportParams()
	ep.Quality = q

	buf, _, err := img.ExportJpeg(ep)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(buf), nil
}

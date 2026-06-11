package main

import (
	"bytes"

	"github.com/h2non/bimg"
)

func convert(src []byte, q int) (*bytes.Buffer, error) {
	// 動作検証の結果こちらは明示的にAutoRotateしないと動かなかった
	img, err := bimg.NewImage(src).AutoRotate()
	if err != nil {
		return nil, err
	}

	opts := bimg.Options{
		Type:    bimg.JPEG,
		Quality: quality,
	}
	jpegImg, err := bimg.NewImage(img).Process(opts)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(jpegImg), nil
}

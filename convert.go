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

	// 動作検証の結果こちらは明示的にAutoRotateしないと動かなかった
	img, err := bimg.NewImage(out).AutoRotate()
	if err != nil {
		bimg.VipsCacheDropAll()
		return nil, err
	}

	opts := bimg.Options{
		Type:    bimg.JPEG,
		Quality: quality,
	}
	jpegImg, err := bimg.NewImage(img).Process(opts)
	// libvipsのキャッシュをクリアしてメモリリークを防ぐ
	bimg.VipsCacheDropAll()
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(jpegImg), nil
}

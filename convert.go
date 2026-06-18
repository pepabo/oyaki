package main

import (
	"bytes"

	"github.com/davidbyttow/govips/v2/vips"
)

// loadImage は libvips に画像を読み込む。
// oyaki は decode→re-encode の一方向パイプラインのみで random access を必要としない。
// EXIF orientation が 1(無回転)の画像は VIPS_ACCESS_SEQUENTIAL で読み込むことで、
// libvips が全画素をメモリ展開せずタイル単位でストリーミング処理し、大きな画像での
// RSS を抑えられる(OOM 対策)。
// 一方、回転/反転を伴う orientation の画像は sequential だと AutoRotate 時に
// "out of order read" でエラーになるため、random access にフォールバックする。
// ヘッダ読み込みは遅延評価で画素デコードを伴わないため、フォールバック時も
// 二重デコードは発生しない。
func loadImage(src []byte) (*vips.ImageRef, error) {
	params := vips.NewImportParams()
	params.Access.Set(vips.AccessSequential)

	img, err := vips.LoadImageFromBuffer(src, params)
	if err != nil {
		return nil, err
	}

	if img.Orientation() == 1 {
		return img, nil
	}

	// sequential で投機的にロードした img はここで使わず捨てる。
	// govips は GC/finalizer 経由でも解放するが、高負荷時は追いつかず OOM するため、
	// ネイティブメモリ(VipsImage)を即時解放してから random access で読み直す。
	img.Close()
	return vips.NewImageFromBuffer(src)
}

func convert(src []byte, q int) (*bytes.Buffer, error) {
	img, err := loadImage(src)
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

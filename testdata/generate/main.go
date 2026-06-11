//go:build ignore

// go run ./testdata/generate/main.go
//
// testdata に使うカラーパターン JPEG を生成する。
// 実行するたびに同一バイト列が得られる決定論的なパターンのみで構成し、
// 外部リソース・著作権上の問題が一切ない画像を作る。
//
// 生成するファイル:
//   testdata/color_pattern.jpg          — 800x600 RGBグラデーション, EXIF なし
//   testdata/color_pattern_rot6.jpg     — 600x800 RGBグラデーション + EXIF orientation=6
//     (orientation=6: 90度時計回りに回転して表示。AutoRotate で 800x600 になる)

package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
)

func main() {
	// 800x600 のカラーグラデーション画像
	wide := makeGradient(800, 600)
	// 600x800 のカラーグラデーション画像 (orientation=6 で表示時 800x600 になる)
	tall := makeGradient(600, 800)

	if err := writeJPEG("testdata/color_pattern.jpg", wide, 0); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote testdata/color_pattern.jpg")

	if err := writeJPEG("testdata/color_pattern_rot6.jpg", tall, 6); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote testdata/color_pattern_rot6.jpg")
}

// makeGradient は w×h の RGBグラデーション画像を生成する。
// 左上が赤、右下が青に向かって変化し、縦方向に緑が混ざる決定論的なパターン。
func makeGradient(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 255 / w),
				G: uint8(y * 255 / h),
				B: uint8((w - x) * 255 / w),
				A: 255,
			})
		}
	}
	return img
}

// writeJPEG は img を JPEG エンコードし、orientation > 0 なら EXIF APP1 を先頭に挿入して保存する。
func writeJPEG(path string, img image.Image, orientation int) error {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return err
	}
	data := buf.Bytes()

	if orientation > 0 {
		app1 := buildExifOrientation(uint16(orientation))
		// SOI (FF D8) の直後に APP1 を挿入
		patched := make([]byte, 0, len(data)+len(app1))
		patched = append(patched, data[:2]...)
		patched = append(patched, app1...)
		patched = append(patched, data[2:]...)
		data = patched
	}

	return os.WriteFile(path, data, 0644)
}

// buildExifOrientation は Orientation タグだけを持つ最小限の EXIF APP1 セグメントを返す。
func buildExifOrientation(orientation uint16) []byte {
	// TIFF (big-endian): IFD0 に Orientation エントリ 1 つ
	tiff := []byte{
		0x4D, 0x4D, 0x00, 0x2A, // "MM" + magic 42
		0x00, 0x00, 0x00, 0x08, // IFD0 offset = 8
		0x00, 0x01,             // 1 entry
		// tag=0x0112, type=SHORT(3), count=1, value=orientation
		0x01, 0x12, 0x00, 0x03, 0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, // value placeholder (2 bytes, big-endian)
		0x00, 0x00, // padding
		0x00, 0x00, 0x00, 0x00, // next IFD = null
	}
	binary.BigEndian.PutUint16(tiff[18:20], orientation)

	// APP1 = marker(2) + length(2) + "Exif\0\0"(6) + tiff
	dataLen := uint16(2 + 6 + len(tiff)) // length フィールド自身を含む
	app1 := []byte{
		0xFF, 0xE1,
		byte(dataLen >> 8), byte(dataLen),
		0x45, 0x78, 0x69, 0x66, 0x00, 0x00, // "Exif\0\0"
	}
	return append(app1, tiff...)
}

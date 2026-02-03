package convert

import (
	"bytes"
	"os"
	"testing"

	"github.com/h2non/bimg"
)

// テストで使用するためのbimg初期化
func init() {
	bimg.Initialize()
	bimg.VipsCacheSetMax(0)
	bimg.VipsCacheSetMaxMem(0)
}

func TestToJPEG(t *testing.T) {
	// テスト用の画像ファイルを読み込み
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		t.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	result, err := ToJPEG(f, 90)
	if err != nil {
		t.Fatalf("JPEG変換に失敗しました: %v", err)
	}

	if result.ContentType != "image/jpeg" {
		t.Errorf("期待されるContentType: image/jpeg, 実際: %s", result.ContentType)
	}

	if result.Size <= 0 {
		t.Errorf("変換後のサイズが0以下です: %d", result.Size)
	}

	if result.Data == nil || result.Data.Len() == 0 {
		t.Error("変換後のデータが空です")
	}
}

func TestToWebP(t *testing.T) {
	// テスト用の画像ファイルを読み込み
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		t.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	result, err := ToWebP(f, 90)
	if err != nil {
		t.Fatalf("WebP変換に失敗しました: %v", err)
	}

	if result.ContentType != "image/webp" {
		t.Errorf("期待されるContentType: image/webp, 実際: %s", result.ContentType)
	}

	if result.Size <= 0 {
		t.Errorf("変換後のサイズが0以下です: %d", result.Size)
	}

	if result.Data == nil || result.Data.Len() == 0 {
		t.Error("変換後のデータが空です")
	}
}

func TestJPEGConverterWithOptions(t *testing.T) {
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		t.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	converter := NewJPEGConverter()
	opts := Options{
		Quality:       50, // 低品質でテスト
		AutoRotate:    true,
		StripMetadata: true,
	}

	result, err := converter.Convert(f, opts)
	if err != nil {
		t.Fatalf("JPEG変換に失敗しました: %v", err)
	}

	if result.ContentType != "image/jpeg" {
		t.Errorf("期待されるContentType: image/jpeg, 実際: %s", result.ContentType)
	}
}

func TestWebPConverterWithOptions(t *testing.T) {
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		t.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	converter := NewWebPConverter()
	opts := Options{
		Quality:       75,
		AutoRotate:    false,
		StripMetadata: false,
	}

	result, err := converter.Convert(f, opts)
	if err != nil {
		t.Fatalf("WebP変換に失敗しました: %v", err)
	}

	if result.ContentType != "image/webp" {
		t.Errorf("期待されるContentType: image/webp, 実際: %s", result.ContentType)
	}
}

func TestInvalidInput(t *testing.T) {
	// 無効なデータでのテスト
	invalidData := bytes.NewBuffer([]byte("invalid image data"))

	_, err := ToJPEG(invalidData, 90)
	if err == nil {
		t.Error("無効なデータでもエラーが発生しませんでした")
	}

	invalidData = bytes.NewBuffer([]byte("invalid image data"))
	_, err = ToWebP(invalidData, 90)
	if err == nil {
		t.Error("無効なデータでもエラーが発生しませんでした")
	}
}

func TestQualityBounds(t *testing.T) {
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		t.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	// 最小品質
	result, err := ToJPEG(f, 1)
	if err != nil {
		t.Errorf("品質1での変換に失敗しました: %v", err)
	}
	if result == nil {
		t.Error("結果がnilです")
	}

	// ファイルポインターを先頭に戻す
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("ファイルのシークに失敗しました: %v", err)
	}

	// 最大品質
	result, err = ToJPEG(f, 100)
	if err != nil {
		t.Errorf("品質100での変換に失敗しました: %v", err)
	}
	if result == nil {
		t.Error("結果がnilです")
	}
}

package convert

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func BenchmarkToJPEG(b *testing.B) {
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		b.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			b.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	// データを再利用するため事前に読み込み
	src, err := io.ReadAll(f)
	if err != nil {
		b.Fatalf("テストデータの読み込みに失敗しました: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		srcBuf := bytes.NewBuffer(src)
		b.StartTimer()

		result, err := ToJPEG(srcBuf, 90)
		if err != nil {
			b.Fatalf("JPEG変換に失敗しました: %v", err)
		}
		result.Data.Reset() // メモリを解放
	}
}

func BenchmarkToWebP(b *testing.B) {
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		b.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			b.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	// データを再利用するため事前に読み込み
	src, err := io.ReadAll(f)
	if err != nil {
		b.Fatalf("テストデータの読み込みに失敗しました: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		srcBuf := bytes.NewBuffer(src)
		b.StartTimer()

		result, err := ToWebP(srcBuf, 90)
		if err != nil {
			b.Fatalf("WebP変換に失敗しました: %v", err)
		}
		result.Data.Reset() // メモリを解放
	}
}

func BenchmarkJPEGConverter(b *testing.B) {
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		b.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			b.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	src, err := io.ReadAll(f)
	if err != nil {
		b.Fatalf("テストデータの読み込みに失敗しました: %v", err)
	}

	converter := NewJPEGConverter()
	opts := Options{
		Quality:       90,
		AutoRotate:    true,
		StripMetadata: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		srcBuf := bytes.NewBuffer(src)
		b.StartTimer()

		result, err := converter.Convert(srcBuf, opts)
		if err != nil {
			b.Fatalf("JPEG変換に失敗しました: %v", err)
		}
		result.Data.Reset() // メモリを解放
	}
}

func BenchmarkWebPConverter(b *testing.B) {
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		b.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			b.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	src, err := io.ReadAll(f)
	if err != nil {
		b.Fatalf("テストデータの読み込みに失敗しました: %v", err)
	}

	converter := NewWebPConverter()
	opts := Options{
		Quality:       90,
		AutoRotate:    true,
		StripMetadata: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		srcBuf := bytes.NewBuffer(src)
		b.StartTimer()

		result, err := converter.Convert(srcBuf, opts)
		if err != nil {
			b.Fatalf("WebP変換に失敗しました: %v", err)
		}
		result.Data.Reset() // メモリを解放
	}
}

// 品質設定別のベンチマーク
func BenchmarkJPEGQuality50(b *testing.B) {
	benchmarkJPEGWithQuality(b, 50)
}

func BenchmarkJPEGQuality75(b *testing.B) {
	benchmarkJPEGWithQuality(b, 75)
}

func BenchmarkJPEGQuality90(b *testing.B) {
	benchmarkJPEGWithQuality(b, 90)
}

func BenchmarkJPEGQuality100(b *testing.B) {
	benchmarkJPEGWithQuality(b, 100)
}

func benchmarkJPEGWithQuality(b *testing.B, quality int) {
	f, err := os.Open("../testdata/oyaki.jpg")
	if err != nil {
		b.Fatalf("テストデータファイルを開けませんでした: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			b.Errorf("ファイルのクローズに失敗しました: %v", err)
		}
	}()

	src, err := io.ReadAll(f)
	if err != nil {
		b.Fatalf("テストデータの読み込みに失敗しました: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		srcBuf := bytes.NewBuffer(src)
		b.StartTimer()

		result, err := ToJPEG(srcBuf, quality)
		if err != nil {
			b.Fatalf("JPEG変換に失敗しました: %v", err)
		}
		result.Data.Reset() // メモリを解放
	}
}

package convert

import (
	"bytes"
	"io"
)

// Options は画像変換のオプションを表します
type Options struct {
	Quality       int  // 画像品質 (1-100)
	AutoRotate    bool // 自動回転
	StripMetadata bool // メタデータ削除
}

// Result は変換結果を表します
type Result struct {
	Data        *bytes.Buffer // 変換後の画像データ
	ContentType string        // MIMEタイプ
	Size        int           // データサイズ
}

// Converter は画像変換インターフェースです
type Converter interface {
	Convert(src io.Reader, opts Options) (*Result, error)
}

// NewJPEGConverter はJPEGコンバーターを作成します
func NewJPEGConverter() Converter {
	return &jpegConverter{}
}

// NewWebPConverter はWebPコンバーターを作成します
func NewWebPConverter() Converter {
	return &webpConverter{}
}

// ToJPEG は画像をJPEGに変換する利便性関数です
func ToJPEG(src io.Reader, quality int) (*Result, error) {
	converter := NewJPEGConverter()
	opts := Options{
		Quality:       quality,
		AutoRotate:    true,
		StripMetadata: false,
	}
	return converter.Convert(src, opts)
}

// ToWebP は画像をWebPに変換する利便性関数です
func ToWebP(src io.Reader, quality int) (*Result, error) {
	converter := NewWebPConverter()
	opts := Options{
		Quality:       quality,
		AutoRotate:    true,
		StripMetadata: true, // WebPではメタデータを削除
	}
	return converter.Convert(src, opts)
}

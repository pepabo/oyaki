package convert

import (
	"bytes"
	"io"

	"github.com/h2non/bimg"
)

// webpConverter はWebP変換を行います
type webpConverter struct{}

// Convert は画像をWebPに変換します
func (c *webpConverter) Convert(src io.Reader, opts Options) (*Result, error) {
	out, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	bimgOpts := bimg.Options{
		Type:         bimg.WEBP,
		Quality:      opts.Quality,
		NoAutoRotate: !opts.AutoRotate,
		// NoAutoRotateはデフォルトでfalseで、勝手にrotateしてくれる

		// Safariなどでは、bimgによってEXIFの回転処理を実施したあとにブラウザ側でEXIFを読んで再度回転してしまうことがあるので、
		// EXIFは削除する
		StripMetadata: opts.StripMetadata,
	}

	webpImg, err := bimg.NewImage(out).Process(bimgOpts)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(webpImg)
	result := &Result{
		Data:        buffer,
		ContentType: "image/webp",
		Size:        len(webpImg),
	}

	return result, nil
}

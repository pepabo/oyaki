package convert

import (
	"bytes"
	"io"

	"github.com/h2non/bimg"
)

// jpegConverter はJPEG変換を行います
type jpegConverter struct{}

// Convert は画像をJPEGに変換します
func (c *jpegConverter) Convert(src io.Reader, opts Options) (*Result, error) {
	img, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	if opts.AutoRotate {
		// 明示的にAutoRotateが必要
		img, err = bimg.NewImage(img).AutoRotate()
		if err != nil {
			return nil, err
		}
	}

	bimgOpts := bimg.Options{
		Type:          bimg.JPEG,
		Quality:       opts.Quality,
		StripMetadata: opts.StripMetadata,
	}

	jpegImg, err := bimg.NewImage(img).Process(bimgOpts)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(jpegImg)
	result := &Result{
		Data:        buffer,
		ContentType: "image/jpeg",
		Size:        len(jpegImg),
	}

	return result, nil
}

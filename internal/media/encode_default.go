//go:build !windows

package media

import (
	"bytes"
	"image"

	"github.com/chai2010/webp"
)

const (
	preferredImageExtension = ".webp"
	preferredImageMime      = "image/webp"
)

func encodeImageData(img image.Image, quality float32) ([]byte, string, error) {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: quality}); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), preferredImageExtension, nil
}

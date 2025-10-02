//go:build windows

package media

import (
	"bytes"
	"image"

	"github.com/disintegration/imaging"
)

const (
	preferredImageExtension = ".png"
	preferredImageMime      = "image/png"
)

func encodeImageData(img image.Image, _ float32) ([]byte, string, error) { // quality is ignored for PNG output
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.PNG); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), preferredImageExtension, nil
}

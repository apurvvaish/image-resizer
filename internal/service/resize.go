package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"

	"github.com/disintegration/imaging"
)

// ResizeWithPreserve resizes the given image.
// If height == 0, preserve aspect ratio and constrain to width.
// If width == 0, constrain to height.
// If both > 0, resize to exact width x height (may change aspect ratio).
func ResizeWithPreserve(img image.Image, width, height int) image.Image {
	if width <= 0 && height <= 0 {
		// nothing to do
		return img
	}
	if width > 0 && height > 0 {
		return imaging.Resize(img, width, height, imaging.Lanczos)
	}
	if width > 0 {
		return imaging.Resize(img, width, 0, imaging.Lanczos)
	}
	// height > 0
	return imaging.Resize(img, 0, height, imaging.Lanczos)
}

// EncodeImageToDataURI encodes image.Image to a base64 data URI using the requested mime type.
// Supported mime: "image/jpeg", "image/jpg", "image/png", "image/webp" — others fall back to PNG.
func EncodeImageToDataURI(img image.Image, mime string) (string, error) {
	var buf bytes.Buffer

	switch mime {
	case "image/jpeg", "image/jpg":
		if err := imaging.Encode(&buf, img, imaging.JPEG); err != nil {
			return "", fmt.Errorf("encode jpeg: %w", err)
		}
	case "image/png":
		if err := imaging.Encode(&buf, img, imaging.PNG); err != nil {
			return "", fmt.Errorf("encode png: %w", err)
		}
	default:
		// fallback to PNG
		if err := imaging.Encode(&buf, img, imaging.PNG); err != nil {
			return "", fmt.Errorf("encode png fallback: %w", err)
		}
		mime = "image/png"
	}

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURI := "data:" + mime + ";base64," + b64
	return dataURI, nil
}

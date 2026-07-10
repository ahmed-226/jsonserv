package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"golang.org/x/image/draw"

	"myserv/internal/config"
)

func CompressImage(base64Input string, cfg config.ImageFieldConfig) (string, error) {
	if cfg.MaxWidth <= 0 {
		cfg.MaxWidth = 128
	}
	if cfg.MaxHeight <= 0 {
		cfg.MaxHeight = 128
	}
	if cfg.Quality <= 0 {
		cfg.Quality = 30
	}
	if cfg.Format == "" {
		cfg.Format = "jpeg"
	}

	raw, mediaType, err := decodeBase64(base64Input)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("image decode: %w", err)
	}

	resized := resize(img, cfg.MaxWidth, cfg.MaxHeight)

	var buf bytes.Buffer
	switch cfg.Format {
	case "png":
		if err := png.Encode(&buf, resized); err != nil {
			return "", fmt.Errorf("png encode: %w", err)
		}
		mediaType = "image/png"
	default:
		if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: cfg.Quality}); err != nil {
			return "", fmt.Errorf("jpeg encode: %w", err)
		}
		mediaType = "image/jpeg"
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:" + mediaType + ";base64," + encoded, nil
}

func decodeBase64(input string) ([]byte, string, error) {
	mediaType := "image/jpeg"
	if idx := strings.Index(input, ","); idx != -1 {
		prefix := input[:idx]
		input = input[idx+1:]
		if strings.HasPrefix(prefix, "data:") {
			if semicolon := strings.Index(prefix, ";"); semicolon != -1 {
				mediaType = prefix[5:semicolon]
			}
		}
	}
	data, err := base64.StdEncoding.DecodeString(input)
	return data, mediaType, err
}

func resize(img image.Image, maxW, maxH int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxW && h <= maxH {
		return img
	}

	ratioW := float64(maxW) / float64(w)
	ratioH := float64(maxH) / float64(h)
	ratio := ratioW
	if ratioH < ratio {
		ratio = ratioH
	}

	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

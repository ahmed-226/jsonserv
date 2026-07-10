package image

import (
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"bytes"
	"testing"

	"myserv/internal/config"
)

func TestCompressImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 128, 255})
		}
	}

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	input := "data:image/jpeg;base64," + encoded

	cfg := config.ImageFieldConfig{
		MaxWidth:  64,
		MaxHeight: 64,
		Quality:   30,
		Format:    "jpeg",
	}

	result, err := CompressImage(input, cfg)
	if err != nil {
		t.Fatalf("CompressImage failed: %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	decoded, err := base64.StdEncoding.DecodeString(result[len("data:image/jpeg;base64,"):])
	if err != nil {
		t.Fatalf("failed to decode result: %v", err)
	}

	if len(decoded) >= buf.Len() {
		t.Errorf("compressed size (%d) should be smaller than original (%d)", len(decoded), buf.Len())
	}

	t.Logf("Original: %d bytes, Compressed: %d bytes (%.1f%% reduction)",
		buf.Len(), len(decoded), 100*(1-float64(len(decoded))/float64(buf.Len())))
}

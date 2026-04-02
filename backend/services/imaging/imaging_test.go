package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"
)

func createTestImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	return img
}

func createTestJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := createTestImage(w, h)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode test jpeg: %v", err)
	}
	return buf.Bytes()
}

func TestResize_Landscape(t *testing.T) {
	img := createTestImage(4000, 3000)
	resized := Resize(img, 300)
	bounds := resized.Bounds()
	if bounds.Dx() != 300 {
		t.Errorf("width = %d, want 300", bounds.Dx())
	}
	if bounds.Dy() != 225 {
		t.Errorf("height = %d, want 225", bounds.Dy())
	}
}

func TestResize_Portrait(t *testing.T) {
	img := createTestImage(3000, 4000)
	resized := Resize(img, 300)
	bounds := resized.Bounds()
	if bounds.Dx() != 225 {
		t.Errorf("width = %d, want 225", bounds.Dx())
	}
	if bounds.Dy() != 300 {
		t.Errorf("height = %d, want 300", bounds.Dy())
	}
}

func TestResize_AlreadySmall(t *testing.T) {
	img := createTestImage(200, 150)
	resized := Resize(img, 300)
	bounds := resized.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("size = %dx%d, want 200x150 (unchanged)", bounds.Dx(), bounds.Dy())
	}
}

func TestApplyOrientation(t *testing.T) {
	img := createTestImage(400, 300) // landscape

	tests := []struct {
		orient   int64
		wantW    int
		wantH    int
	}{
		{1, 400, 300}, // normal
		{2, 400, 300}, // flip H
		{3, 400, 300}, // rotate 180
		{4, 400, 300}, // flip V
		{5, 300, 400}, // transpose (swaps dimensions)
		{6, 300, 400}, // rotate 270 (swaps dimensions)
		{7, 300, 400}, // transverse (swaps dimensions)
		{8, 300, 400}, // rotate 90 (swaps dimensions)
	}

	for _, tt := range tests {
		orient := tt.orient
		result := ApplyOrientation(img, &orient)
		bounds := result.Bounds()
		if bounds.Dx() != tt.wantW || bounds.Dy() != tt.wantH {
			t.Errorf("orientation %d: size = %dx%d, want %dx%d",
				tt.orient, bounds.Dx(), bounds.Dy(), tt.wantW, tt.wantH)
		}
	}

	// nil orientation returns unchanged
	result := ApplyOrientation(img, nil)
	if result != img {
		t.Error("nil orientation should return original image")
	}
}

func TestEncodeJPEG_Roundtrip(t *testing.T) {
	img := createTestImage(640, 480)
	var buf bytes.Buffer
	if err := EncodeJPEG(&buf, img, 85); err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := jpeg.Decode(&buf)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 640 || bounds.Dy() != 480 {
		t.Errorf("size = %dx%d, want 640x480", bounds.Dx(), bounds.Dy())
	}
}

func TestDecodeImage_JPEG(t *testing.T) {
	data := createTestJPEG(t, 800, 600)
	r := bytes.NewReader(data)
	img, err := DecodeImage(r, "jpg")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 600 {
		t.Errorf("size = %dx%d, want 800x600", bounds.Dx(), bounds.Dy())
	}
}

func TestDecodeImage_UnsupportedFormat(t *testing.T) {
	_, err := DecodeImage(strings.NewReader("not an image"), "bmp")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestVariantBySize(t *testing.T) {
	v, ok := VariantBySize(SizeThumb)
	if !ok {
		t.Fatal("expected thumb variant")
	}
	if v.MaxDim != 300 {
		t.Errorf("thumb maxDim = %d, want 300", v.MaxDim)
	}

	v, ok = VariantBySize(SizePreview)
	if !ok {
		t.Fatal("expected preview variant")
	}
	if v.MaxDim != 1600 {
		t.Errorf("preview maxDim = %d, want 1600", v.MaxDim)
	}

	_, ok = VariantBySize("nonexistent")
	if ok {
		t.Error("expected no variant for nonexistent size")
	}
}

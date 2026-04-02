package imaging

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/tiff"
)

// Size represents a named output variant.
type Size string

const (
	SizeThumb   Size = "thumb"
	SizePreview Size = "preview"
)

// Variant holds the parameters for a generated image variant.
type Variant struct {
	Size        Size
	MaxDim      int
	JPEGQuality int
	Filename    string
}

// Variants defines the standard output variants.
var Variants = []Variant{
	{Size: SizeThumb, MaxDim: 300, JPEGQuality: 80, Filename: "thumb.jpg"},
	{Size: SizePreview, MaxDim: 1600, JPEGQuality: 85, Filename: "preview.jpg"},
}

// VariantBySize returns the variant for a given size.
func VariantBySize(s Size) (Variant, bool) {
	for _, v := range Variants {
		if v.Size == s {
			return v, true
		}
	}
	return Variant{}, false
}

// DecodeImage reads a source file and returns a Go image.Image.
// For RAW formats (arw, dng): extracts the embedded JPEG.
// For jpeg/png/tiff: decodes directly.
func DecodeImage(r io.ReadSeeker, format string) (image.Image, error) {
	f := strings.ToLower(format)
	switch f {
	case "arw", "dng":
		jpegReader, err := ExtractEmbeddedJPEG(r)
		if err != nil {
			return nil, fmt.Errorf("extract embedded jpeg: %w", err)
		}
		img, err := jpeg.Decode(jpegReader)
		if err != nil {
			return nil, fmt.Errorf("decode embedded jpeg: %w", err)
		}
		return img, nil
	case "jpg", "jpeg":
		img, err := jpeg.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("decode jpeg: %w", err)
		}
		return img, nil
	case "png":
		img, err := png.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("decode png: %w", err)
		}
		return img, nil
	case "tif", "tiff":
		img, _, err := image.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("decode tiff: %w", err)
		}
		return img, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", f)
	}
}

// Resize scales img so its longest edge is maxDim pixels, maintaining aspect ratio.
// Returns the original if already smaller than maxDim.
func Resize(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= maxDim && h <= maxDim {
		return img
	}
	return imaging.Fit(img, maxDim, maxDim, imaging.Lanczos)
}

// EncodeJPEG writes img as JPEG at the given quality.
func EncodeJPEG(w io.Writer, img image.Image, quality int) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
}

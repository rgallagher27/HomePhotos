package imaging

import (
	"image"

	"github.com/disintegration/imaging"
)

// ApplyOrientation rotates/flips the image according to the EXIF orientation value (1-8).
// Returns the image unchanged for orientation 1 or nil.
func ApplyOrientation(img image.Image, orientation *int64) image.Image {
	if orientation == nil {
		return img
	}
	switch *orientation {
	case 1:
		return img
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

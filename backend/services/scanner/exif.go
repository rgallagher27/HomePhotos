package scanner

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type EXIFData struct {
	Width        *int64
	Height       *int64
	CapturedAt   *time.Time
	CameraMake   string
	CameraModel  string
	LensModel    string
	FocalLength  *float64
	Aperture     *float64
	ShutterSpeed string
	ISO          *int64
	Orientation  *int64
	GPSLatitude  *float64
	GPSLongitude *float64
}

func ExtractEXIF(r io.Reader) (*EXIFData, error) {
	x, err := exif.Decode(r)
	if err != nil {
		// No EXIF data is not an error — return empty struct
		return &EXIFData{}, nil
	}

	data := &EXIFData{}

	if tag, err := x.Get(exif.PixelXDimension); err == nil {
		if v, err := tag.Int(0); err == nil {
			w := int64(v)
			data.Width = &w
		}
	}
	if tag, err := x.Get(exif.PixelYDimension); err == nil {
		if v, err := tag.Int(0); err == nil {
			h := int64(v)
			data.Height = &h
		}
	}

	if t, err := x.DateTime(); err == nil {
		data.CapturedAt = &t
	}

	if tag, err := x.Get(exif.Make); err == nil {
		data.CameraMake, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Model); err == nil {
		data.CameraModel, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.LensModel); err == nil {
		data.LensModel, _ = tag.StringVal()
	}

	if tag, err := x.Get(exif.FocalLength); err == nil {
		numer, denom, err := tag.Rat2(0)
		if err == nil && denom != 0 {
			fl := float64(numer) / float64(denom)
			data.FocalLength = &fl
		}
	}

	if tag, err := x.Get(exif.FNumber); err == nil {
		numer, denom, err := tag.Rat2(0)
		if err == nil && denom != 0 {
			ap := float64(numer) / float64(denom)
			data.Aperture = &ap
		}
	}

	if tag, err := x.Get(exif.ExposureTime); err == nil {
		numer, denom, err := tag.Rat2(0)
		if err == nil && denom != 0 {
			if numer == 1 {
				data.ShutterSpeed = "1/" + itoa(int(denom))
			} else {
				// Express as decimal fraction string
				val := float64(numer) / float64(denom)
				if val >= 1 {
					data.ShutterSpeed = ftoa(val) + "s"
				} else {
					data.ShutterSpeed = "1/" + itoa(int(math.Round(1.0/val)))
				}
			}
		}
	}

	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if v, err := tag.Int(0); err == nil {
			iso := int64(v)
			data.ISO = &iso
		}
	}

	if tag, err := x.Get(exif.Orientation); err == nil {
		if v, err := tag.Int(0); err == nil {
			o := int64(v)
			data.Orientation = &o
		}
	}

	if lat, lon, err := x.LatLong(); err == nil {
		data.GPSLatitude = &lat
		data.GPSLongitude = &lon
	}

	return data, nil
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func ftoa(f float64) string {
	return fmt.Sprintf("%.1f", f)
}

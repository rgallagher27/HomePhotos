package photo

import "time"

type CacheStatus string

const (
	CacheStatusPending CacheStatus = "pending"
	CacheStatusCached  CacheStatus = "cached"
	CacheStatusError   CacheStatus = "error"
)

type Photo struct {
	ID           int64
	FilePath     string
	FileName     string
	FileSize     int64
	FileMtime    time.Time
	Format       string
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
	Fingerprint  string
	ScannedAt    time.Time
	CacheStatus  CacheStatus
}

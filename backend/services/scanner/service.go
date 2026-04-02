package scanner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rgallagher/homephotos/domain/photo"
)

var ErrScanInProgress = errors.New("scan already in progress")

var supportedExtensions = map[string]string{
	".arw":  "arw",
	".dng":  "dng",
	".jpg":  "jpg",
	".jpeg": "jpg",
	".tif":  "tif",
	".tiff": "tif",
	".png":  "png",
}

type Status struct {
	State      string     `json:"status"`
	TotalFiles int        `json:"total_files"`
	Processed  int        `json:"processed"`
	Errors     int        `json:"errors"`
	StartedAt  *time.Time `json:"started_at"`
}

type Service struct {
	photos     photo.Repository
	sourcePath string
	mu         sync.Mutex
	status     atomic.Value
}

func New(photos photo.Repository, sourcePath string) *Service {
	s := &Service{
		photos:     photos,
		sourcePath: sourcePath,
	}
	s.status.Store(idleStatus())
	return s
}

func (s *Service) Status() Status {
	return s.status.Load().(Status)
}

func (s *Service) Run(ctx context.Context) error {
	if !s.mu.TryLock() {
		return ErrScanInProgress
	}
	defer s.mu.Unlock()

	now := time.Now()
	s.status.Store(Status{
		State:     "scanning",
		StartedAt: &now,
	})
	defer s.status.Store(idleStatus())

	return s.scan(ctx)
}

func (s *Service) scan(ctx context.Context) error {
	// Discover files
	type fileInfo struct {
		relPath string
		size    int64
		mtime   time.Time
		format  string
	}

	var files []fileInfo
	err := filepath.WalkDir(s.sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		format, ok := supportedExtensions[ext]
		if !ok {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(s.sourcePath, path)
		if err != nil {
			return nil
		}

		files = append(files, fileInfo{
			relPath: relPath,
			size:    info.Size(),
			mtime:   info.ModTime(),
			format:  format,
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk source: %w", err)
	}

	s.status.Store(Status{
		State:      "scanning",
		TotalFiles: len(files),
		StartedAt:  s.Status().StartedAt,
	})

	// Load existing fingerprints
	existing, err := s.photos.ListAllFingerprints(ctx)
	if err != nil {
		return fmt.Errorf("list fingerprints: %w", err)
	}

	// Process files
	activePaths := make([]string, 0, len(files))
	processed := 0
	scanErrors := 0

	for _, f := range files {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		activePaths = append(activePaths, f.relPath)
		fingerprint := computeFingerprint(f.relPath, f.size, f.mtime)

		if existingFP, exists := existing[f.relPath]; exists && existingFP == fingerprint {
			// Unchanged — skip
			processed++
			s.updateProgress(processed, scanErrors)
			continue
		}

		// New or changed — extract EXIF and upsert
		fullPath := filepath.Join(s.sourcePath, f.relPath)
		exifData := s.extractEXIF(fullPath)

		if _, exists := existing[f.relPath]; exists {
			// Update existing
			existingPhoto, err := s.photos.GetByFilePath(ctx, f.relPath)
			if err != nil {
				slog.Warn("get photo for update", "path", f.relPath, "error", err)
				scanErrors++
				processed++
				s.updateProgress(processed, scanErrors)
				continue
			}
			applyFileInfo(existingPhoto, f.size, f.mtime, f.format, fingerprint)
			applyEXIF(existingPhoto, exifData)
			existingPhoto.CacheStatus = photo.CacheStatusPending
			if err := s.photos.Update(ctx, existingPhoto); err != nil {
				slog.Warn("update photo", "path", f.relPath, "error", err)
				scanErrors++
			}
		} else {
			// Create new
			p := &photo.Photo{
				FilePath:    f.relPath,
				FileName:    filepath.Base(f.relPath),
				FileSize:    f.size,
				FileMtime:   f.mtime,
				Format:      f.format,
				Fingerprint: fingerprint,
				CacheStatus: photo.CacheStatusPending,
			}
			applyEXIF(p, exifData)
			if _, err := s.photos.Create(ctx, p); err != nil {
				slog.Warn("create photo", "path", f.relPath, "error", err)
				scanErrors++
			}
		}

		processed++
		s.updateProgress(processed, scanErrors)
	}

	// Remove orphaned records
	if len(activePaths) > 0 || len(existing) > 0 {
		deleted, err := s.photos.DeleteOrphaned(ctx, activePaths)
		if err != nil {
			slog.Warn("delete orphaned photos", "error", err)
		} else if deleted > 0 {
			slog.Info("removed orphaned photos", "count", deleted)
		}
	}

	return nil
}

func (s *Service) updateProgress(processed, errors int) {
	current := s.Status()
	s.status.Store(Status{
		State:      "scanning",
		TotalFiles: current.TotalFiles,
		Processed:  processed,
		Errors:     errors,
		StartedAt:  current.StartedAt,
	})
}

func (s *Service) extractEXIF(path string) *EXIFData {
	f, err := os.Open(path)
	if err != nil {
		return &EXIFData{}
	}
	defer f.Close()

	data, err := ExtractEXIF(f)
	if err != nil {
		return &EXIFData{}
	}
	return data
}

func computeFingerprint(path string, size int64, mtime time.Time) string {
	return fmt.Sprintf("%s|%d|%s", path, size, mtime.UTC().Format(time.RFC3339Nano))
}

func applyFileInfo(p *photo.Photo, size int64, mtime time.Time, format, fingerprint string) {
	p.FileSize = size
	p.FileMtime = mtime
	p.Format = format
	p.Fingerprint = fingerprint
}

func applyEXIF(p *photo.Photo, e *EXIFData) {
	p.Width = e.Width
	p.Height = e.Height
	p.CapturedAt = e.CapturedAt
	p.CameraMake = e.CameraMake
	p.CameraModel = e.CameraModel
	p.LensModel = e.LensModel
	p.FocalLength = e.FocalLength
	p.Aperture = e.Aperture
	p.ShutterSpeed = e.ShutterSpeed
	p.ISO = e.ISO
	p.Orientation = e.Orientation
	p.GPSLatitude = e.GPSLatitude
	p.GPSLongitude = e.GPSLongitude
}

func idleStatus() Status {
	return Status{State: "idle"}
}

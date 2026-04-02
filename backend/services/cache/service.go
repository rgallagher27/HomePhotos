package cache

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/rgallagher/homephotos/domain/photo"
	"github.com/rgallagher/homephotos/services/imaging"
)

// Service manages the image cache directory and orchestrates thumbnail generation.
type Service struct {
	sourcePath string
	cachePath  string
	photos     photo.Repository
}

// New creates a new cache service.
func New(photos photo.Repository, sourcePath, cachePath string) *Service {
	return &Service{
		sourcePath: sourcePath,
		cachePath:  cachePath,
		photos:     photos,
	}
}

// Photos returns the photo repository used by this service.
func (s *Service) Photos() photo.Repository {
	return s.photos
}

// CacheDir returns the cache directory for a given fingerprint.
// Layout: {cachePath}/{fingerprint[:2]}/{fingerprint}/
func (s *Service) CacheDir(fingerprint string) string {
	prefix := fingerprint
	if len(prefix) > 2 {
		prefix = prefix[:2]
	}
	return filepath.Join(s.cachePath, prefix, fingerprint)
}

// Has checks if a specific variant exists in cache.
func (s *Service) Has(fingerprint string, size imaging.Size) bool {
	path := s.Path(fingerprint, size)
	_, err := os.Stat(path)
	return err == nil
}

// Path returns the filesystem path for a cached variant.
func (s *Service) Path(fingerprint string, size imaging.Size) string {
	v, ok := imaging.VariantBySize(size)
	if !ok {
		return ""
	}
	return filepath.Join(s.CacheDir(fingerprint), v.Filename)
}

// Generate processes a single photo: reads source, decodes, applies orientation,
// resizes to thumb+preview, writes to cache dir, updates DB to 'cached'.
// On error: updates DB to 'error' and returns the error.
func (s *Service) Generate(ctx context.Context, p *photo.Photo) error {
	err := s.generate(ctx, p)
	if err != nil {
		slog.Warn("cache generate failed", "photo_id", p.ID, "path", p.FilePath, "error", err)
		if updateErr := s.photos.UpdateCacheStatus(ctx, p.ID, photo.CacheStatusError); updateErr != nil {
			slog.Error("failed to set error status", "photo_id", p.ID, "error", updateErr)
		}
		return err
	}
	return nil
}

func (s *Service) generate(ctx context.Context, p *photo.Photo) error {
	srcPath := filepath.Join(s.sourcePath, p.FilePath)
	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer f.Close()

	img, err := imaging.DecodeImage(f, p.Format)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	img = imaging.ApplyOrientation(img, p.Orientation)

	cacheDir := s.CacheDir(p.Fingerprint)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("mkdir cache: %w", err)
	}

	for _, v := range imaging.Variants {
		resized := imaging.Resize(img, v.MaxDim)
		outPath := filepath.Join(cacheDir, v.Filename)
		outFile, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", v.Filename, err)
		}
		if err := imaging.EncodeJPEG(outFile, resized, v.JPEGQuality); err != nil {
			outFile.Close()
			return fmt.Errorf("encode %s: %w", v.Filename, err)
		}
		outFile.Close()
	}

	if err := s.photos.UpdateCacheStatus(ctx, p.ID, photo.CacheStatusCached); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

// GenerateIfNeeded calls Generate only if the photo is not already cached.
func (s *Service) GenerateIfNeeded(ctx context.Context, p *photo.Photo) error {
	if p.CacheStatus == photo.CacheStatusCached && s.Has(p.Fingerprint, imaging.SizeThumb) {
		return nil
	}
	return s.Generate(ctx, p)
}

// SourceImageReader returns an io.ReadCloser for the source image.
// For RAW formats: extracts embedded JPEG. For JPEG: returns the file directly.
// Also returns the content type string.
func (s *Service) SourceImageReader(p *photo.Photo) (io.ReadCloser, string, error) {
	srcPath := filepath.Join(s.sourcePath, p.FilePath)
	f, err := os.Open(srcPath)
	if err != nil {
		return nil, "", fmt.Errorf("open source: %w", err)
	}

	switch p.Format {
	case "arw", "dng":
		reader, err := imaging.ExtractEmbeddedJPEG(f)
		if err != nil {
			f.Close()
			return nil, "", fmt.Errorf("extract embedded jpeg: %w", err)
		}
		// Wrap in a struct that closes the underlying file
		return &readerCloser{Reader: reader, Closer: f}, "image/jpeg", nil
	case "png":
		return f, "image/png", nil
	default:
		return f, "image/jpeg", nil
	}
}

type readerCloser struct {
	io.Reader
	io.Closer
}

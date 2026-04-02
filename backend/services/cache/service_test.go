package cache_test

import (
	"bytes"
	"context"
	"database/sql"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/domain/photo"
	"github.com/rgallagher/homephotos/services/cache"
	"github.com/rgallagher/homephotos/services/imaging"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)

	entries, err := os.ReadDir("../../database/sqlite/migrations")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		data, err := os.ReadFile(filepath.Join("../../database/sqlite/migrations", f))
		if err != nil {
			t.Fatalf("read migration %s: %v", f, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			t.Fatalf("exec migration %s: %v", f, err)
		}
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func createTestJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode test jpeg: %v", err)
	}
	return buf.Bytes()
}

func createSourceJPEG(t *testing.T, sourceDir, relPath string, w, h int) {
	t.Helper()
	data := createTestJPEG(t, w, h)
	fullPath := filepath.Join(sourceDir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		t.Fatalf("write source jpeg: %v", err)
	}
}

func newTestPhoto(filePath, fingerprint string) *photo.Photo {
	now := time.Now().Truncate(time.Second)
	return &photo.Photo{
		FilePath:    filePath,
		FileName:    filepath.Base(filePath),
		FileSize:    1024,
		FileMtime:   now,
		Format:      "jpg",
		CapturedAt:  &now,
		Fingerprint: fingerprint,
		CacheStatus: photo.CacheStatusPending,
	}
}

func TestGenerate_JPEG(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	sourceDir := t.TempDir()
	cacheDir := t.TempDir()

	createSourceJPEG(t, sourceDir, "test/photo.jpg", 800, 600)

	p := newTestPhoto("test/photo.jpg", "abcdef1234567890")
	created, err := repo.Create(ctx, p)
	if err != nil {
		t.Fatalf("create photo: %v", err)
	}

	svc := cache.New(repo, sourceDir, cacheDir)
	if err := svc.Generate(ctx, created); err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Verify thumb and preview exist
	thumbPath := svc.Path(created.Fingerprint, imaging.SizeThumb)
	if _, err := os.Stat(thumbPath); err != nil {
		t.Errorf("thumb not found at %s", thumbPath)
	}
	previewPath := svc.Path(created.Fingerprint, imaging.SizePreview)
	if _, err := os.Stat(previewPath); err != nil {
		t.Errorf("preview not found at %s", previewPath)
	}

	// Verify DB status updated
	got, _ := repo.GetByID(ctx, created.ID)
	if got.CacheStatus != photo.CacheStatusCached {
		t.Errorf("status = %q, want cached", got.CacheStatus)
	}

	// Verify cache dir structure
	expectedDir := filepath.Join(cacheDir, "ab", "abcdef1234567890")
	if svc.CacheDir(created.Fingerprint) != expectedDir {
		t.Errorf("cacheDir = %q, want %q", svc.CacheDir(created.Fingerprint), expectedDir)
	}
}

func TestGenerate_CorruptFile(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	sourceDir := t.TempDir()
	cacheDir := t.TempDir()

	// Write garbage bytes as source
	corruptPath := filepath.Join(sourceDir, "corrupt.jpg")
	os.WriteFile(corruptPath, []byte("not a jpeg at all"), 0o644)

	p := newTestPhoto("corrupt.jpg", "corrupt123")
	created, _ := repo.Create(ctx, p)

	svc := cache.New(repo, sourceDir, cacheDir)
	err := svc.Generate(ctx, created)
	if err == nil {
		t.Error("expected error for corrupt file")
	}

	// Verify status set to error
	got, _ := repo.GetByID(ctx, created.ID)
	if got.CacheStatus != photo.CacheStatusError {
		t.Errorf("status = %q, want error", got.CacheStatus)
	}
}

func TestHas(t *testing.T) {
	cacheDir := t.TempDir()
	svc := cache.New(nil, "", cacheDir)

	fp := "abcdef1234567890"
	if svc.Has(fp, imaging.SizeThumb) {
		t.Error("expected Has=false before creating file")
	}

	// Create the cache directory and file
	dir := svc.CacheDir(fp)
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "thumb.jpg"), []byte("fake"), 0o644)

	if !svc.Has(fp, imaging.SizeThumb) {
		t.Error("expected Has=true after creating file")
	}
	if svc.Has(fp, imaging.SizePreview) {
		t.Error("expected Has=false for missing preview")
	}
}

func TestCacheDir_Layout(t *testing.T) {
	svc := cache.New(nil, "", "/cache")
	dir := svc.CacheDir("a3f7b2c9e1d4506f")
	want := filepath.Join("/cache", "a3", "a3f7b2c9e1d4506f")
	if dir != want {
		t.Errorf("dir = %q, want %q", dir, want)
	}
}

func TestPath(t *testing.T) {
	svc := cache.New(nil, "", "/cache")
	fp := "abcdef1234567890"

	thumbPath := svc.Path(fp, imaging.SizeThumb)
	want := filepath.Join("/cache", "ab", fp, "thumb.jpg")
	if thumbPath != want {
		t.Errorf("thumb path = %q, want %q", thumbPath, want)
	}

	previewPath := svc.Path(fp, imaging.SizePreview)
	want = filepath.Join("/cache", "ab", fp, "preview.jpg")
	if previewPath != want {
		t.Errorf("preview path = %q, want %q", previewPath, want)
	}
}

func TestGenerateIfNeeded_AlreadyCached(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	sourceDir := t.TempDir()
	cacheDir := t.TempDir()

	createSourceJPEG(t, sourceDir, "cached.jpg", 400, 300)

	p := newTestPhoto("cached.jpg", "cached123fp")
	created, _ := repo.Create(ctx, p)

	svc := cache.New(repo, sourceDir, cacheDir)

	// Generate first time
	if err := svc.Generate(ctx, created); err != nil {
		t.Fatalf("first generate: %v", err)
	}

	// Re-fetch to get updated status
	created, _ = repo.GetByID(ctx, created.ID)

	// GenerateIfNeeded should be a no-op
	if err := svc.GenerateIfNeeded(ctx, created); err != nil {
		t.Fatalf("generate if needed: %v", err)
	}
}

func TestSourceImageReader_JPEG(t *testing.T) {
	sourceDir := t.TempDir()
	createSourceJPEG(t, sourceDir, "source.jpg", 200, 150)

	svc := cache.New(nil, sourceDir, "")
	p := &photo.Photo{FilePath: "source.jpg", Format: "jpg"}

	rc, contentType, err := svc.SourceImageReader(p)
	if err != nil {
		t.Fatalf("source reader: %v", err)
	}
	defer rc.Close()

	if contentType != "image/jpeg" {
		t.Errorf("content type = %q, want image/jpeg", contentType)
	}

	// Verify we can decode the image
	img, err := jpeg.Decode(rc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("size = %dx%d, want 200x150", bounds.Dx(), bounds.Dy())
	}
}

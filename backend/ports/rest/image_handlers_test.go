package rest

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/domain/photo"
)

func createSourceJPEG(t *testing.T, dir, relPath string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	fullPath := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write jpeg: %v", err)
	}
}

func TestGetPhotoImage_Unauthenticated(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)

	resp := doRequest(t, handler, "GET", "/api/v1/photos/1/image?size=thumb", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestGetPhotoImage_NotFound(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "GET", "/api/v1/photos/99999/image?size=thumb", "", token)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestGetPhotoImage_Thumb_PreCached(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	now := time.Now().Truncate(time.Second)
	p := &photo.Photo{
		FilePath:    "test/thumb.jpg",
		FileName:    "thumb.jpg",
		FileSize:    1024,
		FileMtime:   now,
		Format:      "jpg",
		CapturedAt:  &now,
		Fingerprint: "abcdef1234567890",
		CacheStatus: photo.CacheStatusCached,
	}
	created, _ := env.photos.Create(context.Background(), p)

	// Pre-populate cache with a thumb.jpg
	cacheDir := env.server.cache.CacheDir(created.Fingerprint)
	os.MkdirAll(cacheDir, 0o755)
	thumbData := createTestJPEGBytes(t, 300, 225)
	os.WriteFile(filepath.Join(cacheDir, "thumb.jpg"), thumbData, 0o644)

	resp := doRequest(t, handler, "GET", "/api/v1/photos/"+itoa64(created.ID)+"/image?size=thumb", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "image/jpeg" {
		t.Errorf("content-type = %q, want image/jpeg", ct)
	}

	cc := resp.Header.Get("Cache-Control")
	if cc != "public, max-age=31536000, immutable" {
		t.Errorf("cache-control = %q", cc)
	}
}

func TestGetPhotoImage_OnDemand(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	// Create source file
	createSourceJPEG(t, env.sourceDir, "ondemand.jpg", 800, 600)

	now := time.Now().Truncate(time.Second)
	p := &photo.Photo{
		FilePath:    "ondemand.jpg",
		FileName:    "ondemand.jpg",
		FileSize:    1024,
		FileMtime:   now,
		Format:      "jpg",
		CapturedAt:  &now,
		Fingerprint: "ondemand123fp",
		CacheStatus: photo.CacheStatusPending,
	}
	created, _ := env.photos.Create(context.Background(), p)

	resp := doRequest(t, handler, "GET", "/api/v1/photos/"+itoa64(created.ID)+"/image?size=thumb", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	// Verify DB status was updated
	got, _ := env.photos.GetByID(context.Background(), created.ID)
	if got.CacheStatus != photo.CacheStatusCached {
		t.Errorf("status = %q, want cached", got.CacheStatus)
	}
}

func TestGetPhotoImage_Full(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	createSourceJPEG(t, env.sourceDir, "full.jpg", 640, 480)

	now := time.Now().Truncate(time.Second)
	p := &photo.Photo{
		FilePath:    "full.jpg",
		FileName:    "full.jpg",
		FileSize:    1024,
		FileMtime:   now,
		Format:      "jpg",
		CapturedAt:  &now,
		Fingerprint: "fullres123fp",
		CacheStatus: photo.CacheStatusPending,
	}
	created, _ := env.photos.Create(context.Background(), p)

	resp := doRequest(t, handler, "GET", "/api/v1/photos/"+itoa64(created.ID)+"/image?size=full", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "image/jpeg" {
		t.Errorf("content-type = %q, want image/jpeg", ct)
	}
}

func createTestJPEGBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return buf.Bytes()
}

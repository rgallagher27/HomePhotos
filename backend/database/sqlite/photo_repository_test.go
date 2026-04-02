package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/domain/photo"
)

func newTestPhoto(filePath, fileName, format string) *photo.Photo {
	return &photo.Photo{
		FilePath:    filePath,
		FileName:    fileName,
		FileSize:    1024,
		FileMtime:   time.Now().Truncate(time.Second),
		Format:      format,
		Fingerprint: filePath + "|1024|" + time.Now().Format(time.RFC3339),
		CacheStatus: photo.CacheStatusPending,
	}
}

func ptr[T any](v T) *T { return &v }

func TestPhotoRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	p := &photo.Photo{
		FilePath:     "2025/test.jpg",
		FileName:     "test.jpg",
		FileSize:     5000,
		FileMtime:    now,
		Format:       "jpg",
		Width:        ptr(int64(1920)),
		Height:       ptr(int64(1080)),
		CapturedAt:   &now,
		CameraMake:   "Sony",
		CameraModel:  "ILCE-7RM5",
		LensModel:    "FE 24-70mm F2.8 GM II",
		FocalLength:  ptr(35.0),
		Aperture:     ptr(2.8),
		ShutterSpeed: "1/250",
		ISO:          ptr(int64(400)),
		Orientation:  ptr(int64(1)),
		GPSLatitude:  ptr(43.77),
		GPSLongitude: ptr(11.25),
		Fingerprint:  "abc123",
		CacheStatus:  photo.CacheStatusPending,
	}

	created, err := repo.Create(ctx, p)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if created.FilePath != "2025/test.jpg" {
		t.Errorf("file_path = %q, want %q", created.FilePath, "2025/test.jpg")
	}
	if created.CameraMake != "Sony" {
		t.Errorf("camera_make = %q, want %q", created.CameraMake, "Sony")
	}
	if created.Width == nil || *created.Width != 1920 {
		t.Errorf("width = %v, want 1920", created.Width)
	}
	if created.GPSLatitude == nil || *created.GPSLatitude != 43.77 {
		t.Errorf("gps_latitude = %v, want 43.77", created.GPSLatitude)
	}
	if created.ScannedAt.IsZero() {
		t.Error("expected non-zero scanned_at")
	}
}

func TestPhotoRepository_CreateDuplicate(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	p := newTestPhoto("dup/test.jpg", "test.jpg", "jpg")
	if _, err := repo.Create(ctx, p); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create(ctx, p)
	if err != photo.ErrDuplicatePath {
		t.Errorf("err = %v, want ErrDuplicatePath", err)
	}
}

func TestPhotoRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	if err != photo.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}

	p := newTestPhoto("get/test.jpg", "test.jpg", "jpg")
	created, err := repo.Create(ctx, p)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.FilePath != "get/test.jpg" {
		t.Errorf("file_path = %q, want %q", got.FilePath, "get/test.jpg")
	}
}

func TestPhotoRepository_GetByFilePath(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	_, err := repo.GetByFilePath(ctx, "nonexistent.jpg")
	if err != photo.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}

	p := newTestPhoto("path/test.arw", "test.arw", "arw")
	if _, err := repo.Create(ctx, p); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByFilePath(ctx, "path/test.arw")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.FileName != "test.arw" {
		t.Errorf("file_name = %q, want %q", got.FileName, "test.arw")
	}
}

func TestPhotoRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	p := newTestPhoto("upd/test.jpg", "test.jpg", "jpg")
	created, err := repo.Create(ctx, p)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	created.CameraModel = "ILCE-7RM5"
	created.Width = ptr(int64(9504))
	created.Fingerprint = "updated"
	created.CacheStatus = photo.CacheStatusPending

	if err := repo.Update(ctx, created); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.CameraModel != "ILCE-7RM5" {
		t.Errorf("camera_model = %q, want %q", got.CameraModel, "ILCE-7RM5")
	}
	if got.Width == nil || *got.Width != 9504 {
		t.Errorf("width = %v, want 9504", got.Width)
	}
	if got.Fingerprint != "updated" {
		t.Errorf("fingerprint = %q, want %q", got.Fingerprint, "updated")
	}
}

func TestPhotoRepository_ListAllFingerprints(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	p1 := newTestPhoto("a/1.jpg", "1.jpg", "jpg")
	p1.Fingerprint = "fp1"
	p2 := newTestPhoto("b/2.jpg", "2.jpg", "jpg")
	p2.Fingerprint = "fp2"
	repo.Create(ctx, p1)
	repo.Create(ctx, p2)

	fps, err := repo.ListAllFingerprints(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(fps) != 2 {
		t.Fatalf("len = %d, want 2", len(fps))
	}
	if fps["a/1.jpg"] != "fp1" {
		t.Errorf("fp[a/1.jpg] = %q, want %q", fps["a/1.jpg"], "fp1")
	}
	if fps["b/2.jpg"] != "fp2" {
		t.Errorf("fp[b/2.jpg] = %q, want %q", fps["b/2.jpg"], "fp2")
	}
}

func TestPhotoRepository_DeleteOrphaned(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	repo.Create(ctx, newTestPhoto("keep/1.jpg", "1.jpg", "jpg"))
	repo.Create(ctx, newTestPhoto("remove/2.jpg", "2.jpg", "jpg"))
	repo.Create(ctx, newTestPhoto("remove/3.jpg", "3.jpg", "jpg"))

	deleted, err := repo.DeleteOrphaned(ctx, []string{"keep/1.jpg"})
	if err != nil {
		t.Fatalf("delete orphaned: %v", err)
	}
	if deleted != 2 {
		t.Errorf("deleted = %d, want 2", deleted)
	}

	fps, _ := repo.ListAllFingerprints(ctx)
	if len(fps) != 1 {
		t.Errorf("remaining = %d, want 1", len(fps))
	}
}

func TestPhotoRepository_List_CapturedAtDesc(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)

	p1 := newTestPhoto("a.jpg", "a.jpg", "jpg")
	p1.CapturedAt = &t1
	p2 := newTestPhoto("b.jpg", "b.jpg", "jpg")
	p2.CapturedAt = &t2
	p3 := newTestPhoto("c.jpg", "c.jpg", "jpg")
	p3.CapturedAt = &t3

	repo.Create(ctx, p1)
	repo.Create(ctx, p2)
	repo.Create(ctx, p3)

	result, err := repo.List(ctx, photo.ListParams{
		SortBy:    "captured_at",
		SortOrder: "desc",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Photos) != 3 {
		t.Fatalf("len = %d, want 3", len(result.Photos))
	}
	if result.Photos[0].FileName != "c.jpg" {
		t.Errorf("first = %q, want c.jpg", result.Photos[0].FileName)
	}
	if result.Photos[2].FileName != "a.jpg" {
		t.Errorf("last = %q, want a.jpg", result.Photos[2].FileName)
	}
	if result.HasMore {
		t.Error("expected has_more = false")
	}
}

func TestPhotoRepository_List_FileNameAsc(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	repo.Create(ctx, newTestPhoto("z/charlie.jpg", "charlie.jpg", "jpg"))
	repo.Create(ctx, newTestPhoto("z/alpha.jpg", "alpha.jpg", "jpg"))
	repo.Create(ctx, newTestPhoto("z/bravo.jpg", "bravo.jpg", "jpg"))

	result, err := repo.List(ctx, photo.ListParams{
		SortBy:    "file_name",
		SortOrder: "asc",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Photos[0].FileName != "alpha.jpg" {
		t.Errorf("first = %q, want alpha.jpg", result.Photos[0].FileName)
	}
	if result.Photos[2].FileName != "charlie.jpg" {
		t.Errorf("last = %q, want charlie.jpg", result.Photos[2].FileName)
	}
}

func TestPhotoRepository_List_Filters(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	t1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	p1 := newTestPhoto("trips/italy/1.jpg", "1.jpg", "jpg")
	p1.CapturedAt = &t1
	p2 := newTestPhoto("trips/italy/2.arw", "2.arw", "arw")
	p2.CapturedAt = &t2
	p3 := newTestPhoto("home/3.jpg", "3.jpg", "jpg")
	p3.CapturedAt = &t2

	repo.Create(ctx, p1)
	repo.Create(ctx, p2)
	repo.Create(ctx, p3)

	// Filter by folder
	result, err := repo.List(ctx, photo.ListParams{Limit: 10, Folder: "trips/italy/"})
	if err != nil {
		t.Fatalf("list folder: %v", err)
	}
	if len(result.Photos) != 2 {
		t.Errorf("folder filter: len = %d, want 2", len(result.Photos))
	}

	// Filter by format
	result, err = repo.List(ctx, photo.ListParams{Limit: 10, Format: "arw"})
	if err != nil {
		t.Fatalf("list format: %v", err)
	}
	if len(result.Photos) != 1 {
		t.Errorf("format filter: len = %d, want 1", len(result.Photos))
	}

	// Filter by date range
	dateFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err = repo.List(ctx, photo.ListParams{Limit: 10, DateFrom: &dateFrom})
	if err != nil {
		t.Fatalf("list date: %v", err)
	}
	if len(result.Photos) != 2 {
		t.Errorf("date filter: len = %d, want 2", len(result.Photos))
	}
}

func TestPhotoRepository_List_CursorPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	names := []string{"a", "b", "c", "d", "e"}
	for i, name := range names {
		ts := time.Date(2025, 1, 1+i, 0, 0, 0, 0, time.UTC)
		p := newTestPhoto("page/"+name+".jpg", name+".jpg", "jpg")
		p.CapturedAt = &ts
		p.Fingerprint = "fp" + name
		if _, err := repo.Create(ctx, p); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	// Page 1: limit 2
	result, err := repo.List(ctx, photo.ListParams{
		SortBy:    "captured_at",
		SortOrder: "desc",
		Limit:     2,
	})
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if len(result.Photos) != 2 {
		t.Fatalf("page 1 len = %d, want 2", len(result.Photos))
	}
	if !result.HasMore {
		t.Error("page 1: expected has_more = true")
	}
	if result.NextCursor == "" {
		t.Fatal("page 1: expected non-empty cursor")
	}

	// Page 2
	result2, err := repo.List(ctx, photo.ListParams{
		SortBy:    "captured_at",
		SortOrder: "desc",
		Limit:     2,
		Cursor:    result.NextCursor,
	})
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(result2.Photos) != 2 {
		t.Fatalf("page 2 len = %d, want 2", len(result2.Photos))
	}
	if !result2.HasMore {
		t.Error("page 2: expected has_more = true")
	}

	// Page 3 (last)
	result3, err := repo.List(ctx, photo.ListParams{
		SortBy:    "captured_at",
		SortOrder: "desc",
		Limit:     2,
		Cursor:    result2.NextCursor,
	})
	if err != nil {
		t.Fatalf("page 3: %v", err)
	}
	if len(result3.Photos) != 1 {
		t.Fatalf("page 3 len = %d, want 1", len(result3.Photos))
	}
	if result3.HasMore {
		t.Error("page 3: expected has_more = false")
	}
}

func TestPhotoRepository_ListPending(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

	// Create photos with mixed statuses
	p1 := newTestPhoto("pending1.jpg", "pending1.jpg", "jpg")
	p1.CapturedAt = &t1
	p1.Fingerprint = "fp-pending1"
	created1, _ := repo.Create(ctx, p1)

	p2 := newTestPhoto("cached.jpg", "cached.jpg", "jpg")
	p2.CapturedAt = &t2
	p2.Fingerprint = "fp-cached"
	created2, _ := repo.Create(ctx, p2)
	repo.UpdateCacheStatus(ctx, created2.ID, photo.CacheStatusCached)

	p3 := newTestPhoto("pending2.jpg", "pending2.jpg", "jpg")
	p3.CapturedAt = &t3
	p3.Fingerprint = "fp-pending2"
	created3, _ := repo.Create(ctx, p3)

	p4 := newTestPhoto("error.jpg", "error.jpg", "jpg")
	p4.Fingerprint = "fp-error"
	created4, _ := repo.Create(ctx, p4)
	repo.UpdateCacheStatus(ctx, created4.ID, photo.CacheStatusError)

	// Should return only pending, ordered by captured_at DESC
	pending, err := repo.ListPending(ctx, 10)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("len = %d, want 2", len(pending))
	}
	// t3 (March) is more recent than t1 (January)
	if pending[0].ID != created3.ID {
		t.Errorf("first = %d, want %d (newest pending)", pending[0].ID, created3.ID)
	}
	if pending[1].ID != created1.ID {
		t.Errorf("second = %d, want %d (oldest pending)", pending[1].ID, created1.ID)
	}

	// Respects limit
	limited, err := repo.ListPending(ctx, 1)
	if err != nil {
		t.Fatalf("list pending limit: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("limited len = %d, want 1", len(limited))
	}
}

func TestPhotoRepository_UpdateCacheStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	p := newTestPhoto("status.jpg", "status.jpg", "jpg")
	p.Fingerprint = "fp-status"
	created, _ := repo.Create(ctx, p)

	if created.CacheStatus != photo.CacheStatusPending {
		t.Fatalf("initial status = %q, want pending", created.CacheStatus)
	}

	// Update to cached
	if err := repo.UpdateCacheStatus(ctx, created.ID, photo.CacheStatusCached); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := repo.GetByID(ctx, created.ID)
	if got.CacheStatus != photo.CacheStatusCached {
		t.Errorf("status = %q, want cached", got.CacheStatus)
	}

	// Update to error
	if err := repo.UpdateCacheStatus(ctx, created.ID, photo.CacheStatusError); err != nil {
		t.Fatalf("update error: %v", err)
	}

	got, _ = repo.GetByID(ctx, created.ID)
	if got.CacheStatus != photo.CacheStatusError {
		t.Errorf("status = %q, want error", got.CacheStatus)
	}
}

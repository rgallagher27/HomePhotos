package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/domain/photo"
)

func TestGetPhotos_Empty(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "GET", "/api/v1/photos", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body PhotoListResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Data) != 0 {
		t.Errorf("data len = %d, want 0", len(body.Data))
	}
	if body.HasMore {
		t.Error("expected has_more = false")
	}
}

func TestGetPhotos_WithPhotos(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	now := time.Now().Truncate(time.Second)
	names := []string{"a", "b", "c"}
	for _, name := range names {
		p := &photo.Photo{
			FilePath:    "test/" + name + ".jpg",
			FileName:    name + ".jpg",
			FileSize:    1024,
			FileMtime:   now,
			Format:      "jpg",
			CapturedAt:  &now,
			Fingerprint: "fp" + name,
			CacheStatus: photo.CacheStatusPending,
		}
		s.photos.Create(context.Background(), p)
	}

	resp := doRequest(t, handler, "GET", "/api/v1/photos", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body PhotoListResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Data) != 3 {
		t.Errorf("data len = %d, want 3", len(body.Data))
	}
}

func TestGetPhotos_Unauthenticated(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)

	resp := doRequest(t, handler, "GET", "/api/v1/photos", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestGetPhoto_Found(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	now := time.Now().Truncate(time.Second)
	p := &photo.Photo{
		FilePath:    "detail/test.arw",
		FileName:    "test.arw",
		FileSize:    50000000,
		FileMtime:   now,
		Format:      "arw",
		CapturedAt:  &now,
		CameraMake:  "Sony",
		CameraModel: "ILCE-7RM5",
		Fingerprint: "detail-fp",
		CacheStatus: photo.CacheStatusPending,
	}
	created, _ := s.photos.Create(context.Background(), p)

	resp := doRequest(t, handler, "GET", "/api/v1/photos/"+itoa64(created.ID), "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body PhotoDetailResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.FileName != "test.arw" {
		t.Errorf("file_name = %q, want %q", body.FileName, "test.arw")
	}
	if body.FileSizeBytes != 50000000 {
		t.Errorf("file_size_bytes = %d, want 50000000", body.FileSizeBytes)
	}
	if body.CameraMake == nil || *body.CameraMake != "Sony" {
		t.Errorf("camera_make = %v, want Sony", body.CameraMake)
	}
}

func TestGetPhoto_NotFound(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "GET", "/api/v1/photos/99999", "", token)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestGetPhoto_Unauthenticated(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)

	resp := doRequest(t, handler, "GET", "/api/v1/photos/1", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func itoa64(i int64) string {
	return fmt.Sprintf("%d", i)
}

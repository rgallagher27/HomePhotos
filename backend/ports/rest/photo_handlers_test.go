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

func TestGetPhotos_TagFilter(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	p1 := createTestPhotoViaRepo(t, env, "filter1.jpg")
	p2 := createTestPhotoViaRepo(t, env, "filter2.jpg")
	createTestPhotoViaRepo(t, env, "filter3.jpg") // no tags

	tag1 := createTestTagViaAPI(t, handler, token, "FilterA")
	tag2 := createTestTagViaAPI(t, handler, token, "FilterB")

	doRequest(t, handler, "POST", "/api/v1/photos/"+itoa64(p1.ID)+"/tags",
		`{"tag_ids":[`+itoa64(tag1.Id)+`]}`, token)
	doRequest(t, handler, "POST", "/api/v1/photos/"+itoa64(p2.ID)+"/tags",
		`{"tag_ids":[`+itoa64(tag1.Id)+`,`+itoa64(tag2.Id)+`]}`, token)

	// OR filter: tag1 → p1, p2
	resp := doRequest(t, handler, "GET", "/api/v1/photos?tags="+itoa64(tag1.Id), "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body PhotoListResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Data) != 2 {
		t.Errorf("or filter: len = %d, want 2", len(body.Data))
	}

	// AND filter: tag1 AND tag2 → only p2
	resp = doRequest(t, handler, "GET",
		"/api/v1/photos?tags="+itoa64(tag1.Id)+","+itoa64(tag2.Id)+"&tag_mode=and", "", token)
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Data) != 1 {
		t.Errorf("and filter: len = %d, want 1", len(body.Data))
	}
}

func TestGetPhoto_IncludesTags(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	p := createTestPhotoViaRepo(t, env, "detail-tags.jpg")
	tag1 := createTestTagViaAPI(t, handler, token, "DetailTag")

	doRequest(t, handler, "POST", "/api/v1/photos/"+itoa64(p.ID)+"/tags",
		`{"tag_ids":[`+itoa64(tag1.Id)+`]}`, token)

	resp := doRequest(t, handler, "GET", "/api/v1/photos/"+itoa64(p.ID), "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body PhotoDetailResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Tags == nil || len(*body.Tags) != 1 {
		t.Fatalf("tags = %v, want 1 tag", body.Tags)
	}
	if (*body.Tags)[0].Name != "DetailTag" {
		t.Errorf("tag name = %q, want DetailTag", (*body.Tags)[0].Name)
	}
}

func TestGetPhotos_IncludesTagsInListItems(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	p := createTestPhotoViaRepo(t, env, "list-tags.jpg")
	tag1 := createTestTagViaAPI(t, handler, token, "ListTag")

	doRequest(t, handler, "POST", "/api/v1/photos/"+itoa64(p.ID)+"/tags",
		`{"tag_ids":[`+itoa64(tag1.Id)+`]}`, token)

	resp := doRequest(t, handler, "GET", "/api/v1/photos", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body PhotoListResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Data) != 1 {
		t.Fatalf("data len = %d, want 1", len(body.Data))
	}
	if body.Data[0].Tags == nil || len(*body.Data[0].Tags) != 1 {
		t.Fatalf("tags = %v, want 1 tag", body.Data[0].Tags)
	}
	if (*body.Data[0].Tags)[0].Name != "ListTag" {
		t.Errorf("tag name = %q, want ListTag", (*body.Data[0].Tags)[0].Name)
	}
}

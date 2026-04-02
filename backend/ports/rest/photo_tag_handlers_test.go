package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/domain/photo"
)

func createTestPhotoViaRepo(t *testing.T, env *testEnv, path string) *photo.Photo {
	t.Helper()
	p, err := env.photos.Create(context.Background(), &photo.Photo{
		FilePath:    path,
		FileName:    path,
		FileSize:    1024,
		FileMtime:   time.Now().Truncate(time.Second),
		Format:      "jpg",
		Fingerprint: path + "|fp",
		CacheStatus: photo.CacheStatusPending,
	})
	if err != nil {
		t.Fatalf("create photo: %v", err)
	}
	return p
}

func createTestTagViaAPI(t *testing.T, handler http.Handler, token, name string) TagResponse {
	t.Helper()
	resp := doRequest(t, handler, "POST", "/api/v1/tags", `{"name":"`+name+`"}`, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create tag %s: status = %d", name, resp.StatusCode)
	}
	var tag TagResponse
	json.NewDecoder(resp.Body).Decode(&tag)
	return tag
}

func TestPostPhotoTags_Success(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	p := createTestPhotoViaRepo(t, env, "tagged.jpg")
	tag := createTestTagViaAPI(t, handler, token, "MyTag")

	resp := doRequest(t, handler, "POST", "/api/v1/photos/"+itoa64(p.ID)+"/tags",
		`{"tag_ids":[`+itoa64(tag.Id)+`]}`, token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

func TestPostPhotoTags_Idempotent(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	p := createTestPhotoViaRepo(t, env, "idempotent.jpg")
	tag := createTestTagViaAPI(t, handler, token, "MyTag")

	body := `{"tag_ids":[` + itoa64(tag.Id) + `]}`
	doRequest(t, handler, "POST", "/api/v1/photos/"+itoa64(p.ID)+"/tags", body, token)
	resp := doRequest(t, handler, "POST", "/api/v1/photos/"+itoa64(p.ID)+"/tags", body, token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

func TestPostPhotoTags_PhotoNotFound(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/photos/99999/tags",
		`{"tag_ids":[1]}`, token)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPostPhotoTags_Unauthenticated(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)

	resp := doRequest(t, handler, "POST", "/api/v1/photos/1/tags",
		`{"tag_ids":[1]}`, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestDeletePhotoTag_Success(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	p := createTestPhotoViaRepo(t, env, "remove.jpg")
	tag := createTestTagViaAPI(t, handler, token, "RemoveMe")

	doRequest(t, handler, "POST", "/api/v1/photos/"+itoa64(p.ID)+"/tags",
		`{"tag_ids":[`+itoa64(tag.Id)+`]}`, token)

	resp := doRequest(t, handler, "DELETE", "/api/v1/photos/"+itoa64(p.ID)+"/tags/"+itoa64(tag.Id), "", token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

func TestDeletePhotoTag_Idempotent(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	p := createTestPhotoViaRepo(t, env, "idem.jpg")

	resp := doRequest(t, handler, "DELETE", "/api/v1/photos/"+itoa64(p.ID)+"/tags/99999", "", token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

func TestPostBulkTag_Success(t *testing.T) {
	env := newTestEnv(t, true)
	handler := buildHandler(t, env.server)
	token := registerUser(t, handler, "admin", "password123")

	p1 := createTestPhotoViaRepo(t, env, "bulk1.jpg")
	p2 := createTestPhotoViaRepo(t, env, "bulk2.jpg")
	t1 := createTestTagViaAPI(t, handler, token, "BulkTag1")
	t2 := createTestTagViaAPI(t, handler, token, "BulkTag2")

	body := `{"photo_ids":[` + itoa64(p1.ID) + `,` + itoa64(p2.ID) + `],"tag_ids":[` + itoa64(t1.Id) + `,` + itoa64(t2.Id) + `]}`
	resp := doRequest(t, handler, "POST", "/api/v1/photos/bulk-tag", body, token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

func TestPostBulkTag_EmptyArrays(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/photos/bulk-tag",
		`{"photo_ids":[],"tag_ids":[1]}`, token)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestPostBulkTag_Unauthenticated(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)

	resp := doRequest(t, handler, "POST", "/api/v1/photos/bulk-tag",
		`{"photo_ids":[1],"tag_ids":[1]}`, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

package rest

import (
	"encoding/json"
	"net/http"
	"testing"
)

// Tag Group tests

func TestPostTagGroup_Success(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"People","sort_order":1}`, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var body TagGroupResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Name != "People" || body.SortOrder != 1 {
		t.Errorf("body = %+v", body)
	}
}

func TestPostTagGroup_Duplicate(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"People"}`, token)
	resp := doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"People"}`, token)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPostTagGroup_Forbidden(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	registerUser(t, handler, "admin", "password123")
	viewerToken := registerUser(t, handler, "viewer", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"People"}`, viewerToken)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}

func TestGetTagGroups(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"Bbb","sort_order":2}`, token)
	doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"Aaa","sort_order":1}`, token)

	resp := doRequest(t, handler, "GET", "/api/v1/tag-groups", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body TagGroupListResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Data) != 2 {
		t.Fatalf("count = %d, want 2", len(body.Data))
	}
	if body.Data[0].Name != "Aaa" {
		t.Errorf("first = %q, want Aaa (sorted by sort_order)", body.Data[0].Name)
	}
}

func TestPutTagGroup(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"Old"}`, token)
	var created TagGroupResponse
	json.NewDecoder(resp.Body).Decode(&created)

	resp = doRequest(t, handler, "PUT", "/api/v1/tag-groups/"+itoa64(created.Id), `{"name":"New","sort_order":5}`, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var updated TagGroupResponse
	json.NewDecoder(resp.Body).Decode(&updated)
	if updated.Name != "New" || updated.SortOrder != 5 {
		t.Errorf("updated = %+v", updated)
	}
}

func TestDeleteTagGroup(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"ToDelete"}`, token)
	var created TagGroupResponse
	json.NewDecoder(resp.Body).Decode(&created)

	resp = doRequest(t, handler, "DELETE", "/api/v1/tag-groups/"+itoa64(created.Id), "", token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

// Tag tests

func TestPostTag_Success(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	// Create group first
	resp := doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"Places"}`, token)
	var group TagGroupResponse
	json.NewDecoder(resp.Body).Decode(&group)

	resp = doRequest(t, handler, "POST", "/api/v1/tags",
		`{"name":"Beach","color":"#3498db","group_id":`+itoa64(group.Id)+`}`, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var tag TagResponse
	json.NewDecoder(resp.Body).Decode(&tag)
	if tag.Name != "Beach" || tag.Color == nil || *tag.Color != "#3498db" {
		t.Errorf("tag = %+v", tag)
	}
}

func TestPostTag_Ungrouped(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tags", `{"name":"Misc"}`, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var tag TagResponse
	json.NewDecoder(resp.Body).Decode(&tag)
	if tag.GroupId != nil {
		t.Errorf("group_id = %v, want nil", tag.GroupId)
	}
}

func TestPostTag_Duplicate(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	doRequest(t, handler, "POST", "/api/v1/tags", `{"name":"Dup"}`, token)
	resp := doRequest(t, handler, "POST", "/api/v1/tags", `{"name":"Dup"}`, token)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPostTag_Forbidden(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	registerUser(t, handler, "admin", "password123")
	viewerToken := registerUser(t, handler, "viewer", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tags", `{"name":"Test"}`, viewerToken)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}

func TestGetTags(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tag-groups", `{"name":"Group1"}`, token)
	var group TagGroupResponse
	json.NewDecoder(resp.Body).Decode(&group)

	doRequest(t, handler, "POST", "/api/v1/tags",
		`{"name":"Tag1","group_id":`+itoa64(group.Id)+`}`, token)
	doRequest(t, handler, "POST", "/api/v1/tags", `{"name":"Tag2"}`, token)

	resp = doRequest(t, handler, "GET", "/api/v1/tags", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body TagListResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Data) != 2 {
		t.Errorf("count = %d, want 2", len(body.Data))
	}
}

func TestPutTag(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tags", `{"name":"Old","color":"#000"}`, token)
	var created TagResponse
	json.NewDecoder(resp.Body).Decode(&created)

	resp = doRequest(t, handler, "PUT", "/api/v1/tags/"+itoa64(created.Id),
		`{"name":"New","color":"#fff"}`, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var updated TagResponse
	json.NewDecoder(resp.Body).Decode(&updated)
	if updated.Name != "New" || updated.Color == nil || *updated.Color != "#fff" {
		t.Errorf("updated = %+v", updated)
	}
}

func TestDeleteTag(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/tags", `{"name":"ToDelete"}`, token)
	var created TagResponse
	json.NewDecoder(resp.Body).Decode(&created)

	resp = doRequest(t, handler, "DELETE", "/api/v1/tags/"+itoa64(created.Id), "", token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

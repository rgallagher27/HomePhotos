package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rgallagher/homephotos/domain/tag"
)

// Tags

func (s *Server) GetTags(w http.ResponseWriter, r *http.Request) {
	if UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	tags, err := s.tags.ListTags(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tags")
		return
	}

	data := make([]TagResponse, len(tags))
	for i, t := range tags {
		data[i] = tagToResponse(t)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(TagListResponse{Data: data})
}

func (s *Server) PostTag(w http.ResponseWriter, r *http.Request) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t := &tag.Tag{
		Name:      req.Name,
		GroupID:   req.GroupId,
		CreatedBy: authUser.UserID,
	}
	if req.Color != nil {
		t.Color = *req.Color
	}

	created, err := s.tags.CreateTag(r.Context(), t)
	if err != nil {
		if errors.Is(err, tag.ErrDuplicateName) {
			writeError(w, http.StatusConflict, "tag name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create tag")
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tagToResponse(*created))
}

func (s *Server) PutTag(w http.ResponseWriter, r *http.Request, id int64) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	existing, err := s.tags.GetTagByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, tag.ErrNotFound) {
			writeError(w, http.StatusNotFound, "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch tag")
		return
	}

	var req UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Color != nil {
		existing.Color = *req.Color
	}
	if req.GroupId != nil {
		existing.GroupID = req.GroupId
	}

	if err := s.tags.UpdateTag(r.Context(), existing); err != nil {
		if errors.Is(err, tag.ErrDuplicateName) {
			writeError(w, http.StatusConflict, "tag name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update tag")
		return
	}

	updated, _ := s.tags.GetTagByID(r.Context(), id)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tagToResponse(*updated))
}

func (s *Server) DeleteTag(w http.ResponseWriter, r *http.Request, id int64) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	if _, err := s.tags.GetTagByID(r.Context(), id); err != nil {
		if errors.Is(err, tag.ErrNotFound) {
			writeError(w, http.StatusNotFound, "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch tag")
		return
	}

	if err := s.tags.DeleteTag(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete tag")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Tag Groups

func (s *Server) GetTagGroups(w http.ResponseWriter, r *http.Request) {
	if UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	groups, err := s.tags.ListGroups(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tag groups")
		return
	}

	data := make([]TagGroupResponse, len(groups))
	for i, g := range groups {
		data[i] = tagGroupToResponse(g)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(TagGroupListResponse{Data: data})
}

func (s *Server) PostTagGroup(w http.ResponseWriter, r *http.Request) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	var req CreateTagGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	g := &tag.TagGroup{Name: req.Name}
	if req.SortOrder != nil {
		g.SortOrder = *req.SortOrder
	}

	created, err := s.tags.CreateGroup(r.Context(), g)
	if err != nil {
		if errors.Is(err, tag.ErrDuplicateName) {
			writeError(w, http.StatusConflict, "tag group name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create tag group")
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tagGroupToResponse(*created))
}

func (s *Server) PutTagGroup(w http.ResponseWriter, r *http.Request, id int64) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	existing, err := s.tags.GetGroupByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, tag.ErrGroupNotFound) {
			writeError(w, http.StatusNotFound, "tag group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch tag group")
		return
	}

	var req UpdateTagGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}

	if err := s.tags.UpdateGroup(r.Context(), existing); err != nil {
		if errors.Is(err, tag.ErrDuplicateName) {
			writeError(w, http.StatusConflict, "tag group name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update tag group")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tagGroupToResponse(*existing))
}

func (s *Server) DeleteTagGroup(w http.ResponseWriter, r *http.Request, id int64) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	if _, err := s.tags.GetGroupByID(r.Context(), id); err != nil {
		if errors.Is(err, tag.ErrGroupNotFound) {
			writeError(w, http.StatusNotFound, "tag group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch tag group")
		return
	}

	if err := s.tags.DeleteGroup(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete tag group")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Converters

func tagToResponse(t tag.Tag) TagResponse {
	resp := TagResponse{
		Id:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}
	if t.Color != "" {
		resp.Color = &t.Color
	}
	if t.GroupID != nil {
		resp.GroupId = t.GroupID
	}
	if t.GroupName != "" {
		resp.GroupName = &t.GroupName
	}
	return resp
}

func tagGroupToResponse(g tag.TagGroup) TagGroupResponse {
	return TagGroupResponse{
		Id:        g.ID,
		Name:      g.Name,
		SortOrder: g.SortOrder,
		CreatedAt: g.CreatedAt,
	}
}

package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rgallagher/homephotos/domain/photo"
)

func (s *Server) PostPhotoTags(w http.ResponseWriter, r *http.Request, id int64) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if _, err := s.photos.GetByID(r.Context(), id); err != nil {
		if errors.Is(err, photo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "photo not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch photo")
		return
	}

	var req AssignTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.tags.AddPhotoTags(r.Context(), id, req.TagIds, authUser.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to assign tags")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) DeletePhotoTag(w http.ResponseWriter, r *http.Request, id int64, tagId int64) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if _, err := s.photos.GetByID(r.Context(), id); err != nil {
		if errors.Is(err, photo.ErrNotFound) {
			writeError(w, http.StatusNotFound, "photo not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch photo")
		return
	}

	if err := s.tags.RemovePhotoTag(r.Context(), id, tagId); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove tag")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) PostBulkTag(w http.ResponseWriter, r *http.Request) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req BulkTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.PhotoIds) == 0 || len(req.TagIds) == 0 {
		writeError(w, http.StatusBadRequest, "photo_ids and tag_ids must not be empty")
		return
	}

	if err := s.tags.BulkAddPhotoTags(r.Context(), req.PhotoIds, req.TagIds, authUser.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to bulk assign tags")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

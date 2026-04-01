package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rgallagher/homephotos/domain/user"
)

func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	users, err := s.users.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	data := make([]UserResponse, len(users))
	for i, u := range users {
		data[i] = userToResponse(&u)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(UserListResponse{Data: data})
}

func (s *Server) PutUserRole(w http.ResponseWriter, r *http.Request, id int64) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if authUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin access required")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role := user.Role(req.Role)
	if role != user.RoleAdmin && role != user.RoleViewer {
		writeError(w, http.StatusBadRequest, "invalid role")
		return
	}

	// Check user exists before updating
	if _, err := s.users.GetByID(r.Context(), id); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	if err := s.users.UpdateRole(r.Context(), id, role); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	u, err := s.users.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userToResponse(u))
}

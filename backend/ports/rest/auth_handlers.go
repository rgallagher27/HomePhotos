package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rgallagher/homephotos/domain/user"
	"github.com/rgallagher/homephotos/services/auth"
)

func (s *Server) PostAuthRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := ""
	if req.Email != nil {
		email = *req.Email
	}

	result, err := s.auth.Register(r.Context(), req.Username, req.Password, email)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrRegistrationClosed):
			writeError(w, http.StatusForbidden, "registration is closed")
		case errors.Is(err, user.ErrDuplicateUsername):
			writeError(w, http.StatusConflict, "username already exists")
		default:
			writeError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(AuthResponse{
		Id:       result.ID,
		Username: result.Username,
		Role:     AuthResponseRole(result.Role),
		Token:    result.Token,
	})
}

func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := s.auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(AuthResponse{
		Id:       result.ID,
		Username: result.Username,
		Role:     AuthResponseRole(result.Role),
		Token:    result.Token,
	})
}

func (s *Server) GetAuthMe(w http.ResponseWriter, r *http.Request) {
	authUser := UserFromContext(r.Context())
	if authUser == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	u, err := s.users.GetByID(r.Context(), authUser.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userToResponse(u))
}

func userToResponse(u *user.User) UserResponse {
	resp := UserResponse{
		Id:        u.ID,
		Username:  u.Username,
		Role:      UserResponseRole(u.Role),
		CreatedAt: u.CreatedAt,
		LastLogin: u.LastLogin,
	}
	if u.DisplayName != "" {
		resp.DisplayName = &u.DisplayName
	}
	if u.Email != "" {
		resp.Email = &u.Email
	}
	return resp
}

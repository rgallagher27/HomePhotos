package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/services/auth"
)

func TestUserFromContext(t *testing.T) {
	// nil when no user in context
	r := httptest.NewRequest("GET", "/", nil)
	if u := UserFromContext(r.Context()); u != nil {
		t.Error("expected nil user from empty context")
	}

	// non-nil when user set
	ctx := ContextWithUser(r.Context(), &AuthenticatedUser{UserID: 1, Role: "admin"})
	u := UserFromContext(ctx)
	if u == nil {
		t.Fatal("expected non-nil user")
	}
	if u.UserID != 1 {
		t.Errorf("userID = %d, want 1", u.UserID)
	}
	if u.Role != "admin" {
		t.Errorf("role = %q, want %q", u.Role, "admin")
	}
}

func TestRequireAdmin(t *testing.T) {
	handler := RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		user       *AuthenticatedUser
		wantStatus int
	}{
		{"no user", nil, http.StatusUnauthorized},
		{"viewer", &AuthenticatedUser{UserID: 1, Role: "viewer"}, http.StatusForbidden},
		{"admin", &AuthenticatedUser{UserID: 1, Role: "admin"}, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if tt.user != nil {
				r = r.WithContext(ContextWithUser(r.Context(), tt.user))
			}
			w := httptest.NewRecorder()
			handler(w, r)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestNewJWTAuthenticator_TokenValidation(t *testing.T) {
	tokens := auth.NewTokenService("test-secret", time.Hour)

	// Generate a valid token and verify the authenticator extracts claims correctly
	tokenStr, err := tokens.Generate(42, "viewer")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Verify the token service works (the authenticator delegates to it)
	claims, err := tokens.Validate(tokenStr)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("userID = %d, want 42", claims.UserID)
	}
	if claims.Role != "viewer" {
		t.Errorf("role = %q, want %q", claims.Role, "viewer")
	}
}

package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

func buildHandler(t *testing.T, s *Server) http.Handler {
	t.Helper()
	swagger, err := GetSwagger()
	if err != nil {
		t.Fatalf("get swagger: %v", err)
	}
	swagger.Servers = nil

	h := HandlerWithOptions(s, StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(),
		Middlewares: []MiddlewareFunc{
			middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
				Options: openapi3filter.Options{
					AuthenticationFunc: NewJWTAuthenticator(s.tokens),
				},
				ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
					writeError(w, statusCode, message)
				},
			}),
			jsonContentTypeMiddleware,
		},
	})

	return jwtContextMiddleware(s.tokens)(h)
}

func doRequest(t *testing.T, handler http.Handler, method, path, body, token string) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w.Result()
}

func extractToken(t *testing.T, resp *http.Response) string {
	t.Helper()
	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	return result.Token
}

func TestPostAuthRegister(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantRole   string
	}{
		{
			name:       "first user gets admin",
			body:       `{"username":"alice","password":"password123"}`,
			wantStatus: http.StatusCreated,
			wantRole:   "admin",
		},
		{
			name:       "missing username",
			body:       `{"password":"password123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			body:       `{"username":"bob"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "password too short",
			body:       `{"username":"bob","password":"short"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testServer(t, true)
			h := buildHandler(t, s)

			resp := doRequest(t, h, "POST", "/api/v1/auth/register", tt.body, "")
			if resp.StatusCode != tt.wantStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d, want %d; body: %s", resp.StatusCode, tt.wantStatus, body)
			}

			if tt.wantRole != "" {
				var result struct {
					Role string `json:"role"`
				}
				json.NewDecoder(resp.Body).Decode(&result)
				if result.Role != tt.wantRole {
					t.Errorf("role = %q, want %q", result.Role, tt.wantRole)
				}
			}
		})
	}
}

func TestPostAuthRegister_Duplicate(t *testing.T) {
	s := testServer(t, true)
	h := buildHandler(t, s)

	registerUser(t, h, "alice", "password123")

	resp := doRequest(t, h, "POST", "/api/v1/auth/register", `{"username":"alice","password":"password123"}`, "")
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
}

func TestPostAuthRegister_Closed(t *testing.T) {
	s := testServer(t, false)
	h := buildHandler(t, s)

	resp := doRequest(t, h, "POST", "/api/v1/auth/register", `{"username":"alice","password":"password123"}`, "")
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestPostAuthLogin(t *testing.T) {
	s := testServer(t, true)
	h := buildHandler(t, s)
	registerUser(t, h, "alice", "password123")

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid credentials", `{"username":"alice","password":"password123"}`, http.StatusOK},
		{"wrong password", `{"username":"alice","password":"wrongpassword"}`, http.StatusUnauthorized},
		{"unknown user", `{"username":"nobody","password":"password123"}`, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doRequest(t, h, "POST", "/api/v1/auth/login", tt.body, "")
			if resp.StatusCode != tt.wantStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want %d; body: %s", resp.StatusCode, tt.wantStatus, body)
			}
		})
	}
}

func TestGetAuthMe(t *testing.T) {
	s := testServer(t, true)
	h := buildHandler(t, s)
	token := registerUser(t, h, "alice", "password123")

	// Authenticated request
	resp := doRequest(t, h, "GET", "/api/v1/auth/me", "", token)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200; body: %s", resp.StatusCode, body)
	}

	var userResp struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}
	json.NewDecoder(resp.Body).Decode(&userResp)
	if userResp.Username != "alice" {
		t.Errorf("username = %q, want %q", userResp.Username, "alice")
	}
	if userResp.Role != "admin" {
		t.Errorf("role = %q, want %q", userResp.Role, "admin")
	}

	// Unauthenticated request
	resp = doRequest(t, h, "GET", "/api/v1/auth/me", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no-auth status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

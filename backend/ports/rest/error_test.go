package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		wantCode   string
	}{
		{"bad request", http.StatusBadRequest, "invalid input", "BAD_REQUEST"},
		{"not found", http.StatusNotFound, "user not found", "NOT_FOUND"},
		{"unauthorized", http.StatusUnauthorized, "invalid token", "UNAUTHORIZED"},
		{"internal error", http.StatusInternalServerError, "something broke", "INTERNAL_SERVER_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeError(w, tt.statusCode, tt.message)

			if w.Code != tt.statusCode {
				t.Errorf("status = %d, want %d", w.Code, tt.statusCode)
			}

			var resp struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
			}
			if resp.Error.Message != tt.message {
				t.Errorf("message = %q, want %q", resp.Error.Message, tt.message)
			}
		})
	}
}

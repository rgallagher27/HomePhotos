package auth

import (
	"testing"
	"time"
)

func TestTokenService(t *testing.T) {
	ts := NewTokenService("test-secret-key", time.Hour)

	tests := []struct {
		name    string
		userID  int64
		role    string
		wantErr bool
	}{
		{"admin token", 1, "admin", false},
		{"viewer token", 42, "viewer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ts.Generate(tt.userID, tt.role)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			if token == "" {
				t.Fatal("expected non-empty token")
			}

			claims, err := ts.Validate(token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if claims.UserID != tt.userID {
					t.Errorf("userID = %d, want %d", claims.UserID, tt.userID)
				}
				if claims.Role != tt.role {
					t.Errorf("role = %q, want %q", claims.Role, tt.role)
				}
			}
		})
	}
}

func TestTokenService_Expired(t *testing.T) {
	ts := NewTokenService("test-secret", -time.Hour) // negative expiry = already expired

	token, err := ts.Generate(1, "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = ts.Validate(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestTokenService_InvalidSignature(t *testing.T) {
	ts1 := NewTokenService("secret-one", time.Hour)
	ts2 := NewTokenService("secret-two", time.Hour)

	token, _ := ts1.Generate(1, "admin")

	_, err := ts2.Validate(token)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestTokenService_Malformed(t *testing.T) {
	ts := NewTokenService("secret", time.Hour)

	_, err := ts.Validate("not-a-jwt")
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

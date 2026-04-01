package auth

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/domain/user"
)

// fakeRepo implements user.Repository in-memory for testing.
type fakeRepo struct {
	mu    sync.Mutex
	users map[int64]*user.User
	nextID int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: make(map[int64]*user.User), nextID: 1}
}

func (r *fakeRepo) Create(_ context.Context, u *user.User) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.users {
		if existing.Username == u.Username {
			return nil, user.ErrDuplicateUsername
		}
	}
	u.ID = r.nextID
	r.nextID++
	u.CreatedAt = time.Now()
	clone := *u
	r.users[clone.ID] = &clone
	return &clone, nil
}

func (r *fakeRepo) GetByUsername(_ context.Context, username string) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Username == username {
			clone := *u
			return &clone, nil
		}
	}
	return nil, user.ErrNotFound
}

func (r *fakeRepo) GetByID(_ context.Context, id int64) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	clone := *u
	return &clone, nil
}

func (r *fakeRepo) List(_ context.Context) ([]user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []user.User
	for _, u := range r.users {
		result = append(result, *u)
	}
	return result, nil
}

func (r *fakeRepo) UpdateRole(_ context.Context, id int64, role user.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return user.ErrNotFound
	}
	u.Role = role
	return nil
}

func (r *fakeRepo) UpdateLastLogin(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return user.ErrNotFound
	}
	now := time.Now()
	u.LastLogin = &now
	return nil
}

func (r *fakeRepo) Count(_ context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.users)), nil
}

func newTestService(registrationOpen bool) (*Service, *fakeRepo) {
	repo := newFakeRepo()
	tokens := NewTokenService("test-secret", time.Hour)
	svc := New(repo, tokens, 4, registrationOpen) // cost 4 for fast tests
	return svc, repo
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Service, *fakeRepo)
		username string
		password string
		wantRole string
		wantErr  error
	}{
		{
			name:     "first user gets admin",
			username: "alice",
			password: "password123",
			wantRole: "admin",
		},
		{
			name: "second user gets viewer",
			setup: func(svc *Service, _ *fakeRepo) {
				svc.Register(context.Background(), "first", "password123", "")
			},
			username: "second",
			password: "password123",
			wantRole: "viewer",
		},
		{
			name: "duplicate username",
			setup: func(svc *Service, _ *fakeRepo) {
				svc.Register(context.Background(), "alice", "password123", "")
			},
			username: "alice",
			password: "password123",
			wantErr:  user.ErrDuplicateUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestService(true)
			if tt.setup != nil {
				tt.setup(svc, repo)
			}

			result, err := svc.Register(context.Background(), tt.username, tt.password, "")
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Role != tt.wantRole {
				t.Errorf("role = %q, want %q", result.Role, tt.wantRole)
			}
			if result.Token == "" {
				t.Error("expected non-empty token")
			}
			if result.Username != tt.username {
				t.Errorf("username = %q, want %q", result.Username, tt.username)
			}
		})
	}
}

func TestService_RegisterClosed(t *testing.T) {
	svc, _ := newTestService(false)
	_, err := svc.Register(context.Background(), "alice", "password123", "")
	if err != ErrRegistrationClosed {
		t.Errorf("err = %v, want ErrRegistrationClosed", err)
	}
}

func TestService_Login(t *testing.T) {
	svc, _ := newTestService(true)
	svc.Register(context.Background(), "alice", "password123", "")

	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{"valid credentials", "alice", "password123", nil},
		{"wrong password", "alice", "wrongpassword", ErrInvalidCredentials},
		{"unknown user", "nobody", "password123", ErrInvalidCredentials},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.Login(context.Background(), tt.username, tt.password)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Token == "" {
				t.Error("expected non-empty token")
			}
			if result.Username != tt.username {
				t.Errorf("username = %q, want %q", result.Username, tt.username)
			}
		})
	}
}

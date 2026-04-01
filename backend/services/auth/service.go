package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/rgallagher/homephotos/domain/user"
)

type Result struct {
	ID       int64
	Username string
	Role     string
	Token    string
}

type Service struct {
	users            user.Repository
	tokens           *TokenService
	bcryptCost       int
	registrationOpen bool
}

func New(users user.Repository, tokens *TokenService, bcryptCost int, registrationOpen bool) *Service {
	return &Service{
		users:            users,
		tokens:           tokens,
		bcryptCost:       bcryptCost,
		registrationOpen: registrationOpen,
	}
}

func (s *Service) Register(ctx context.Context, username, password, email string) (*Result, error) {
	if !s.registrationOpen {
		return nil, ErrRegistrationClosed
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	count, err := s.users.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	role := user.RoleViewer
	if count == 0 {
		role = user.RoleAdmin
	}

	created, err := s.users.Create(ctx, &user.User{
		Username:     username,
		PasswordHash: string(hash),
		Email:        email,
		Role:         role,
	})
	if err != nil {
		return nil, err
	}

	token, err := s.tokens.Generate(created.ID, string(created.Role))
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &Result{
		ID:       created.ID,
		Username: created.Username,
		Role:     string(created.Role),
		Token:    token,
	}, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*Result, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	_ = s.users.UpdateLastLogin(ctx, u.ID)

	token, err := s.tokens.Generate(u.ID, string(u.Role))
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &Result{
		ID:       u.ID,
		Username: u.Username,
		Role:     string(u.Role),
		Token:    token,
	}, nil
}

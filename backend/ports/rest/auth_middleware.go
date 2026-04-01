package rest

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/rgallagher/homephotos/services/auth"
)

type contextKey string

const authenticatedUserKey contextKey = "authenticated_user"

type AuthenticatedUser struct {
	UserID int64
	Role   string
}

func UserFromContext(ctx context.Context) *AuthenticatedUser {
	u, _ := ctx.Value(authenticatedUserKey).(*AuthenticatedUser)
	return u
}

func ContextWithUser(ctx context.Context, u *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, authenticatedUserKey, u)
}

// NewJWTAuthenticator returns an openapi3filter.AuthenticationFunc that
// validates Bearer tokens. It only checks token validity — it does not
// set request context. Use jwtContextMiddleware to inject the user into context.
func NewJWTAuthenticator(tokens *auth.TokenService) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		if input.SecuritySchemeName != "BearerAuth" {
			return fmt.Errorf("unsupported security scheme: %s", input.SecuritySchemeName)
		}

		authHeader := input.RequestValidationInput.Request.Header.Get("Authorization")
		if authHeader == "" {
			return fmt.Errorf("missing Authorization header")
		}

		tokenStr, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			return fmt.Errorf("invalid Authorization header format")
		}

		_, err := tokens.Validate(tokenStr)
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}

		return nil
	}
}

// jwtContextMiddleware extracts the JWT from the Authorization header (if present)
// and injects the AuthenticatedUser into the request context for downstream handlers.
// This runs on every request but only sets context when a valid token is found.
func jwtContextMiddleware(tokens *auth.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if tokenStr, found := strings.CutPrefix(authHeader, "Bearer "); found {
				if claims, err := tokens.Validate(tokenStr); err == nil {
					ctx := ContextWithUser(r.Context(), &AuthenticatedUser{
						UserID: claims.UserID,
						Role:   claims.Role,
					})
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u == nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if u.Role != "admin" {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next(w, r)
	}
}

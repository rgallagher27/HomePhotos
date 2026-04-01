package rest

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rgallagher/homephotos/config"
	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/services/auth"
	sloghttp "github.com/samber/slog-http"
)

func NewRestServer(ctx context.Context, cfg config.Config) (*http.Server, error) {
	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return initServer(cfg, db)
}

func initServer(cfg config.Config, db *sql.DB) (*http.Server, error) {
	swagger, err := GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("get swagger: %w", err)
	}
	swagger.Servers = nil

	tokens := auth.NewTokenService(cfg.JWTSecret, 24*time.Hour)
	userRepo := sqlite.NewUserRepository(db)
	authSvc := auth.New(userRepo, tokens, 12, cfg.RegistrationOpen)

	server := NewServer(db, authSvc, tokens, userRepo)

	h := HandlerWithOptions(server, StdHTTPServerOptions{
		BaseRouter: http.NewServeMux(),
		Middlewares: []MiddlewareFunc{
			middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
				Options: openapi3filter.Options{
					AuthenticationFunc: NewJWTAuthenticator(tokens),
				},
				ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
					writeError(w, statusCode, message)
				},
			}),
			jsonContentTypeMiddleware,
		},
	})

	var handler http.Handler = h
	handler = jwtContextMiddleware(tokens)(handler)
	handler = corsMiddleware(handler)
	handler = sloghttp.Recovery(handler)
	handler = sloghttp.New(slog.Default())(handler)

	return &http.Server{
		Handler: handler,
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
	}, nil
}

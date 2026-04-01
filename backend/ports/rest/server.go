package rest

import (
	"database/sql"

	"github.com/rgallagher/homephotos/domain/user"
	"github.com/rgallagher/homephotos/services/auth"
)

type Server struct {
	db     *sql.DB
	auth   *auth.Service
	tokens *auth.TokenService
	users  user.Repository
}

func NewServer(db *sql.DB, authSvc *auth.Service, tokens *auth.TokenService, users user.Repository) *Server {
	return &Server{
		db:     db,
		auth:   authSvc,
		tokens: tokens,
		users:  users,
	}
}

var _ ServerInterface = (*Server)(nil)

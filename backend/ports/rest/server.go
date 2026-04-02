package rest

import (
	"database/sql"

	"github.com/rgallagher/homephotos/domain/photo"
	"github.com/rgallagher/homephotos/domain/user"
	"github.com/rgallagher/homephotos/services/auth"
	"github.com/rgallagher/homephotos/services/scanner"
)

type Server struct {
	db      *sql.DB
	auth    *auth.Service
	tokens  *auth.TokenService
	users   user.Repository
	photos  photo.Repository
	scanner *scanner.Service
}

func NewServer(db *sql.DB, authSvc *auth.Service, tokens *auth.TokenService, users user.Repository, photos photo.Repository, scannerSvc *scanner.Service) *Server {
	return &Server{
		db:      db,
		auth:    authSvc,
		tokens:  tokens,
		users:   users,
		photos:  photos,
		scanner: scannerSvc,
	}
}

var _ ServerInterface = (*Server)(nil)

package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathan-osman/certy/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Server provides the web interface for interacting with the CA and
// certificate functions in the storage package.
type Server struct {
	server  http.Server
	logger  zerolog.Logger
	storage *storage.Storage
}

// New create a new Server instance.
func New(addr string, st *storage.Storage) *Server {
	var (
		r = gin.New()
		s = &Server{
			server: http.Server{
				Addr:    addr,
				Handler: r,
			},
			logger:  log.With().Str("package", "server").Logger(),
			storage: st,
		}
	)

	// Listen for connections in a separate goroutine
	go func() {
		defer s.logger.Info().Msg("server stopped")
		s.logger.Info().Msg("server started")
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Msg(err.Error())
		}
	}()

	return s
}

// Close shuts down the server.
func (s *Server) Close() {
	s.server.Shutdown(context.Background())
}

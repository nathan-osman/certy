package server

import (
	"context"
	"embed"
	"errors"
	"log/slog"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/nathan-osman/certy/storage"
	loader "github.com/nathan-osman/pongo2-embed-loader"
	"gitlab.com/go-box/pongo2gin/v6"
)

var (
	//go:embed static
	staticFS embed.FS

	//go:embed templates
	tmplFS embed.FS

	formMethods = []string{
		http.MethodGet,
		http.MethodPost,
	}
)

func init() {
	pongo2.RegisterFilter("formatBytes", formatBytes)
	pongo2.RegisterFilter("formatDate", formatDate)
}

// Server provides the web interface for interacting with the CA and
// certificate functions in the storage package.
type Server struct {
	server  http.Server
	logger  *slog.Logger
	storage *storage.Storage
}

// New create a new Server instance.
func New(cfg *Config) (*Server, error) {
	var (
		r = gin.New()
		s = &Server{
			server: http.Server{
				Addr:    cfg.Addr,
				Handler: r,
			},
			logger:  cfg.Logger,
			storage: cfg.Storage,
		}
	)

	// Configure the logger
	if s.logger == nil {
		s.logger = slog.Default()
	}
	s.logger = s.logger.With("package", "server")

	// Load the template set
	tmplSet := pongo2.NewSet("", &loader.Loader{
		Content: tmplFS,
	})

	// Render HTML templates with pongo
	r.HTMLRender = pongo2gin.New(pongo2gin.RenderOptions{
		TemplateSet: tmplSet,
	})

	// Handle errors gracefully
	r.Use(gin.CustomRecovery(s.errorHandler))

	// TODO: not do this
	r.GET("/", s.index)
	r.GET("/view", s.certView)
	r.Match(formMethods, "/new", s.certNew)

	// Static files
	f, err := static.EmbedFolder(staticFS, "static")
	if err != nil {
		return nil, err
	}
	r.Use(static.Serve("/static", f))

	// Listen for connections in a separate goroutine
	go func() {
		defer s.logger.Info("server stopped")
		s.logger.Info("server started")
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error(err.Error())
		}
	}()

	return s, nil
}

// Close shuts down the server.
func (s *Server) Close() {
	s.server.Shutdown(context.Background())
}

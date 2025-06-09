package server

import (
	"context"
	"embed"
	"errors"
	"net/http"
	"path"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/nathan-osman/certy/storage"
	loader "github.com/nathan-osman/pongo2-embed-loader"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/go-box/pongo2gin/v6"
)

// Server provides the web interface for interacting with the CA and
// certificate functions in the storage package.
type Server struct {
	server  http.Server
	logger  zerolog.Logger
	storage *storage.Storage
}

var (
	//go:embed static
	staticFS embed.FS

	//go:embed templates
	tmplFS embed.FS
)

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix, filepath string) bool {
	f, err := e.Open(path.Join(prefix, filepath))
	if err != nil {
		return false
	}
	f.Close()
	return true
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

	// Load the template set
	tmplSet := pongo2.NewSet("", &loader.Loader{
		Content: tmplFS,
	})

	// Render HTML templates with pongo
	r.HTMLRender = pongo2gin.New(pongo2gin.RenderOptions{
		TemplateSet: tmplSet,
	})

	// Interface pages
	r.GET("/", s.index)
	r.GET("/new", s.newCAGET)
	r.POST("/new", s.newCAPOST)
	r.GET("/:uuid", s.viewCAGET)

	// Static files
	r.Use(static.Serve("/", embedFileSystem{http.FS(staticFS)}))

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

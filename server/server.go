package server

import (
	"context"
	"embed"
	"errors"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/nathan-osman/certy/storage"
	loader "github.com/nathan-osman/pongo2-embed-loader"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/go-box/pongo2gin/v6"
)

const (
	contextCAName = "ca_name"
	contextCACert = "ca_cert"
)

var (
	//go:embed static
	staticFS embed.FS

	//go:embed templates
	tmplFS embed.FS
)

func init() {
	pongo2.RegisterFilter("formatBytes", formatBytes)
	pongo2.RegisterFilter("formatDate", formatDate)
}

// Server provides the web interface for interacting with the CA and
// certificate functions in the storage package.
type Server struct {
	server  http.Server
	logger  zerolog.Logger
	storage *storage.Storage
}

// New create a new Server instance.
func New(addr string, st *storage.Storage) (*Server, error) {
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

	groupCA := r.Group("/:uuid")
	{
		groupCA.Use(func(c *gin.Context) {
			u := c.Param("uuid")
			v, err := s.storage.LoadCA(u)
			if err != nil {
				panic(err)
			}
			c.Set(contextCAName, u)
			c.Set(contextCACert, v)
			c.Next()
		})
		groupCA.GET("", s.viewCAGET)
		groupCA.GET("/new", s.caNewGET)
		groupCA.POST("/new", s.caNewPOST)
	}

	// Static files
	f, err := static.EmbedFolder(staticFS, "static")
	if err != nil {
		return nil, err
	}
	r.Use(static.Serve("/static", f))

	// Listen for connections in a separate goroutine
	go func() {
		defer s.logger.Info().Msg("server stopped")
		s.logger.Info().Msg("server started")
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Msg(err.Error())
		}
	}()

	return s, nil
}

// Close shuts down the server.
func (s *Server) Close() {
	s.server.Shutdown(context.Background())
}

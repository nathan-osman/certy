package server

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"runtime"
	"runtime/debug"

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
	pongo2.RegisterFilter("formatDuration", formatDuration)
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

	// Switch to release mode when debug mode is enabled
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

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

	// If debug mode is enabled, use the templates directly from the
	// filesystem; otherwise, use the built-in ones
	var tmplLoader pongo2.TemplateLoader
	if cfg.Debug {
		tmplLoader = pongo2.MustNewLocalFileSystemLoader(
			"server/templates",
		)
	} else {
		subFS, err := fs.Sub(tmplFS, "templates")
		if err != nil {
			return nil, err
		}
		tmplLoader = &loader.Loader{Content: subFS}
	}

	// Load build info
	b, _ := debug.ReadBuildInfo()

	// Create the template set
	tmplSet := pongo2.NewSet("", tmplLoader)
	tmplSet.Globals["GOOS"] = runtime.GOOS
	tmplSet.Globals["GOARCH"] = runtime.GOARCH
	tmplSet.Globals["BuildInfo"] = b

	// Enable auto-reload if debug is enabled
	if cfg.Debug {
		tmplSet.Debug = true
		tmplSet.Globals["DEBUG"] = true
	}

	// Render HTML templates with pongo
	r.HTMLRender = pongo2gin.New(pongo2gin.RenderOptions{
		TemplateSet: tmplSet,
	})

	// Handle errors gracefully
	r.Use(gin.CustomRecovery(s.errorHandler))

	// Routes for the server
	r.GET("/", s.index)
	r.GET("/view", s.certView)
	r.POST("/validate", s.certValidate)
	r.POST("/action", s.certAction)
	r.Match(formMethods, "/new", s.certNew)
	r.Match(formMethods, "/pkcs12", s.certPKCS12)
	r.Match(formMethods, "/delete", s.certDelete)

	// Static files (use FS if running in debug)
	var serveFS static.ServeFileSystem
	if cfg.Debug {
		serveFS = static.LocalFile("server/static", false)
	} else {
		f, err := static.EmbedFolder(staticFS, "static")
		if err != nil {
			return nil, err
		}
		serveFS = f
	}
	r.Use(static.Serve("/static", serveFS))

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

package server

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"runtime"
	"runtime/debug"
	"slices"

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

	splitPathRegExp = regexp.MustCompile(
		`/([0-9a-f]{12}(?:/[0-9a-f]{12})*)(?:/(\w+))?`,
	)

	methodsGet     = []string{http.MethodGet}
	methodsPost    = []string{http.MethodPost}
	methodsGetPost = []string{http.MethodGet, http.MethodPost}
)

func init() {
	pongo2.RegisterFilter("formatBytes", formatBytes)
	pongo2.RegisterFilter("formatDate", formatDate)
	pongo2.RegisterFilter("formatDuration", formatDuration)
}

type internalRoute struct {
	methods []string
	handler func(*gin.Context, string)
}

// Server provides the web interface for interacting with the CA and
// certificate functions in the storage package.
type Server struct {
	server  http.Server
	logger  *slog.Logger
	storage *storage.Storage
	routes  map[string]internalRoute
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
		ContentType: "text/html; charset=utf-8",
	})

	// Handle errors gracefully
	r.Use(gin.CustomRecovery(s.errorHandler))

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

	// In order to provide URLs of the format:
	//
	// / [root] / [intermediate] / [leaf] / [action]
	//
	// ...some custom routing logic is required (unfortunately). This is all
	// handled in routePath(), which calls the appropriate handler.
	r.GET("/*path", s.routePath)
	r.POST("/*path", s.routePath)

	// Populate the route map
	s.routes = map[string]internalRoute{
		"": {
			methods: methodsGet,
			handler: s.certView,
		},
		"validate": {
			methods: methodsPost,
			handler: s.certValidate,
		},
		"export": {
			methods: methodsPost,
			handler: s.certExport,
		},
		"new": {
			methods: methodsGetPost,
			handler: s.certNew,
		},
		"pkcs12": {
			methods: methodsGetPost,
			handler: s.certPKCS12,
		},
		"delete": {
			methods: methodsGetPost,
			handler: s.certDelete,
		},
	}

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

func (s *Server) routePath(c *gin.Context) {

	p := c.Param("path")

	// Show the home page for GET /
	if p == "/" && c.Request.Method == http.MethodGet {
		s.index(c)
		return
	}

	// Page for creating new root certificates
	if p == "/new" && slices.Contains(methodsGetPost, c.Request.Method) {
		s.certNew(c, "")
		return
	}

	// Split the path into / [cert] / [action]
	v := splitPathRegExp.FindStringSubmatch(p)
	if len(v) < 2 {
		s.e404Handler(c)
		return
	}

	// Check if the action is in the route slice
	r, ok := s.routes[v[2]]
	if !ok {
		s.e404Handler(c)
		return
	}

	// Check if the method is allowed
	if !slices.Contains(r.methods, c.Request.Method) {
		s.errorHandler(c, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	// Route the request
	r.handler(c, v[1])
}

// Close shuts down the server.
func (s *Server) Close() {
	s.server.Shutdown(context.Background())
}

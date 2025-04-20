package server

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathan-osman/certy/storage"
)

var (
	//go:embed templates/*
	tmplFS embed.FS
)

func render(c *gin.Context, name string, data any) {
	t, err := template.ParseFS(tmplFS, "templates/base.html", name)
	if err != nil {
		panic(err)
	}
	if err := t.Execute(c.Writer, data); err != nil {
		panic(err)
	}
}

func (s *Server) index(c *gin.Context) {
	render(c, "templates/index.html", gin.H{
		"Certificates": s.storage.ListCAs(),
	})
}

func (s *Server) newCAGET(c *gin.Context) {
	render(c, "templates/new_ca.html", gin.H{})
}

func (s *Server) newCAPOST(c *gin.Context) {
	params := &storage.CreateCAParams{}
	if err := c.ShouldBind(params); err != nil {
		panic(err)
	}
	if err := s.storage.CreateCA(params); err != nil {
		panic(err)
	}
	c.Redirect(http.StatusFound, "/")
}

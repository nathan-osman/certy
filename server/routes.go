package server

import (
	"embed"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
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
		"Year": time.Now().Year(),
	})
}

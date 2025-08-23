package server

import (
	"fmt"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
	"github.com/nathan-osman/certy/storage"
)

func (s *Server) errorHandler(c *gin.Context, err any) {
	msg := "an unknown error has occurred"
	switch v := err.(type) {
	case error:
		msg = v.Error()
	}
	c.HTML(http.StatusInternalServerError, "templates/error.html", pongo2.Context{
		"msg": msg,
	})
	c.Abort()
}

func (s *Server) index(c *gin.Context) {
	c.HTML(http.StatusOK, "templates/index.html", pongo2.Context{
		"refs": s.storage.GetRootCertificates(),
	})
}

func (s *Server) certView(c *gin.Context) {
	p := c.Query("cert")
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	c.HTML(http.StatusOK, "templates/cert_view.html", pongo2.Context{
		"path": p,
		"cert": v,
	})
}

func (s *Server) certNew(c *gin.Context) {
	var (
		p    = c.Query("cert")
		cert *storage.Certificate
		form = &storage.CreateCertificateParams{}
	)
	if p != "" {
		v, err := s.storage.GetCertificate(p)
		if err != nil {
			panic(err)
		}
		cert = v
	}
	if c.Request.Method == http.MethodPost {
		if err := c.ShouldBind(form); err != nil {
			panic(err)
		}
		v, err := s.storage.CreateCertificate(p, form)
		if err != nil {
			panic(err)
		}
		c.Redirect(
			http.StatusSeeOther,
			fmt.Sprintf("/view?cert=%s", v),
		)
		return
	}
	c.HTML(http.StatusOK, "templates/cert_new.html", pongo2.Context{
		"cert": cert,
		"form": form,
	})
}

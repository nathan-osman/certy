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
		"entries": s.storage.ListCAs(),
	})
}

func (s *Server) newCAGET(c *gin.Context) {
	c.HTML(http.StatusOK, "templates/new_ca.html", pongo2.Context{})
}

func (s *Server) newCAPOST(c *gin.Context) {
	params := &storage.CreateCAParams{}
	if err := c.ShouldBind(params); err != nil {
		panic(err)
	}
	u, err := s.storage.CreateCA(params)
	if err != nil {
		panic(err)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/%s", u))
}

func (s *Server) viewCAGET(c *gin.Context) {
	c.HTML(http.StatusOK, "templates/view_ca.html", pongo2.Context{
		"name":        c.GetString(contextCAName),
		"certificate": certFromContext(c, contextCACert),
	})
}

func (s *Server) caNewCertGET(c *gin.Context) {
	c.HTML(http.StatusOK, "templates/new_cert.html", pongo2.Context{
		"name":        c.GetString(contextCAName),
		"certificate": certFromContext(c, contextCACert),
	})
}

func (s *Server) caNewCertPOST(c *gin.Context) {
	var (
		parentName = c.GetString(contextCAName)
		params     = &storage.CreateCertParams{}
	)
	if err := c.ShouldBind(params); err != nil {
		panic(err)
	}
	u, err := s.storage.CreateCert(parentName, params)
	if err != nil {
		panic(err)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf(
		"/%s/%s",
		parentName,
		u,
	))
}

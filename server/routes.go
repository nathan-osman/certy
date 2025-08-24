package server

import (
	"fmt"
	"net/http"
	"strconv"

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

func (s *Server) certPKCS12(c *gin.Context) {
	var (
		p    = c.Query("cert")
		cert *storage.Certificate
		form = &storage.ExportCertificatePKCS12Params{}
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
		b, err := s.storage.ExportCertificatePKCS12(p, form)
		if err != nil {
			panic(err)
		}
		c.Header(
			"Content-Disposition",
			fmt.Sprintf(
				`attachment; filename="%s.p12"`,
				cert.X509.Subject.CommonName,
			),
		)
		c.Header("Content-Length", strconv.Itoa(len(b)))
		c.Data(http.StatusOK, "application/x-pkcs12", b)
		return
	}
	c.HTML(http.StatusOK, "templates/cert_pkcs12.html", pongo2.Context{
		"cert": cert,
		"form": form,
	})
}

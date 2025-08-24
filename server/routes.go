package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
	"github.com/nathan-osman/certy/storage"
)

var (
	errInvalidAction = errors.New("invalid action specified")
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
		"path":           p,
		"cert":           v,
		"combineAddress": combineAddress,
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
	} else {
		if cert != nil {
			sub := cert.X509.Subject
			form.Organization = ifPresent(sub.Organization)
			form.OrganizationalUnit = ifPresent(sub.OrganizationalUnit)
			form.Country = ifPresent(sub.Country)
			form.Province = ifPresent(sub.Province)
			form.Locality = ifPresent(sub.Locality)
			form.StreetAddress = ifPresent(sub.StreetAddress)
			form.PostalCode = ifPresent(sub.PostalCode)
		} else {
			form.CanSign = true
		}
	}
	c.HTML(http.StatusOK, "templates/cert_new.html", pongo2.Context{
		"cert": cert,
		"form": form,
		"path": p,
		"page": "New",
	})
}

func (s *Server) certAction(c *gin.Context) {
	var (
		p    = c.Query("cert")
		cert *storage.Certificate
	)
	if p != "" {
		v, err := s.storage.GetCertificate(p)
		if err != nil {
			panic(err)
		}
		cert = v
	}
	var (
		b         []byte
		err       error
		suffix    string
		extension string
	)
	switch c.Query("action") {
	case "export_cert_pem":
		b, err = s.storage.ExportCertificatePEM(p)
		extension = "crt"
	case "export_chain_pem":
		b, err = s.storage.ExportCertificateChainPEM(p)
		suffix = "-chain"
		extension = "pem"
	case "export_pub_key":
		b, err = s.storage.ExportPublicKeyPEM(p)
		extension = "pub"
	case "export_priv_key":
		b, err = s.storage.ExportPrivateKeyPEM(p)
		extension = "key"
	default:
		panic(errInvalidAction)
	}
	if err != nil {
		panic(err)
	}
	downloadCert(c, "application/x-pem-file", b, cert, suffix, extension)
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
		downloadCert(c, "application/x-pkcs12", b, cert, "", "p12")
		return
	}
	c.HTML(http.StatusOK, "templates/cert_pkcs12.html", pongo2.Context{
		"cert": cert,
		"form": form,
		"path": p,
		"page": "Export",
	})
}

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
	c.HTML(http.StatusInternalServerError, "error.html", pongo2.Context{
		"title": "Something Went Wrong",
		"desc":  "An error was encountered while trying to display the page",
		"msg":   msg,
	})
	c.Abort()
}

func (s *Server) index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", pongo2.Context{
		"title": "Root Certificates",
		"desc":  "View root certificates currently managed by Certy",
		"refs":  s.storage.GetRootCertificates(),
	})
}

func (s *Server) certView(c *gin.Context) {
	p := c.Query("cert")
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	c.HTML(http.StatusOK, "cert_view.html", pongo2.Context{
		"title":          v.X509.Subject.CommonName,
		"desc":           "View and manage this certificate and its children",
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
	var desc string
	if cert != nil {
		desc = fmt.Sprintf(
			"Create a new certificate signed by %s",
			cert.X509.Subject.CommonName,
		)
	} else {
		desc = "Create a new root certificate"
	}
	c.HTML(http.StatusOK, "cert_new.html", pongo2.Context{
		"title": "New Certificate",
		"desc":  desc,
		"cert":  cert,
		"form":  form,
		"path":  p,
		"page":  "New",
	})
}

func (s *Server) certValidate(c *gin.Context) {
	p := c.Query("cert")
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	var msg string
	if err := s.storage.ValidateCertificate(
		c.Query("cert"),
	); err != nil {
		msg = err.Error()
	}
	c.HTML(http.StatusOK, "cert_validate.html", pongo2.Context{
		"title": "Validation Results",
		"desc":  "The results of your certificate validation are shown below",
		"cert":  v,
		"path":  p,
		"msg":   msg,
		"page":  "Validation",
	})
}

func (s *Server) certAction(c *gin.Context) {
	p := c.Query("cert")
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	var (
		b         []byte
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
	downloadCert(c, "application/x-pem-file", b, v, suffix, extension)
}

func (s *Server) certPKCS12(c *gin.Context) {
	var (
		p    = c.Query("cert")
		form = &storage.ExportCertificatePKCS12Params{}
	)
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	if c.Request.Method == http.MethodPost {
		if err := c.ShouldBind(form); err != nil {
			panic(err)
		}
		b, err := s.storage.ExportCertificatePKCS12(p, form)
		if err != nil {
			panic(err)
		}
		downloadCert(c, "application/x-pkcs12", b, v, "", "p12")
		return
	}
	desc := fmt.Sprintf(
		"Export %s and its private key in PKCS#12 format",
		v.X509.Subject.CommonName,
	)
	c.HTML(http.StatusOK, "cert_pkcs12.html", pongo2.Context{
		"title": "Export PKCS#12",
		"desc":  desc,
		"cert":  v,
		"form":  form,
		"path":  p,
		"page":  "Export",
	})
}

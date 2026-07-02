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
	errInvalidFmt = errors.New("invalid format specified")
)

func (s *Server) e404Handler(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", pongo2.Context{
		"title": "Page Not Found",
		"desc":  "The page you are attempting to view does not exist",
	})
}

func (s *Server) errorHandler(c *gin.Context, err any) {
	msg := "an unknown error has occurred"
	switch v := err.(type) {
	case string:
		msg = v
	case error:
		msg = v.Error()
	}
	c.HTML(http.StatusInternalServerError, "error.html", pongo2.Context{
		"title": "Something Went Wrong",
		"desc":  "An error was encountered while trying to display the page",
		"msg":   msg,
		"stack": captureStack(),
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

func (s *Server) certView(c *gin.Context, p string) {
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	c.HTML(http.StatusOK, "cert_view.html", pongo2.Context{
		"title":          v.X509.Subject.CommonName,
		"desc":           "View and manage this certificate and its children",
		"cert":           v,
		"combineAddress": combineAddress,
	})
}

func (s *Server) certNew(c *gin.Context, p string) {
	var (
		cert *storage.Certificate
		form = &storage.CreateCertificateParams{
			KeySize: 2048,
		}
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
			fmt.Sprintf("/%s", v.Path),
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
			form.AllowChaining = true
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
		"page":  "New Certificate",
	})
}

func (s *Server) certValidate(c *gin.Context, p string) {
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	r, err := s.storage.ValidateCertificate(v.Path)
	if err != nil {
		panic(err)
	}
	c.HTML(http.StatusOK, "cert_validate.html", pongo2.Context{
		"title":   "Validation Results",
		"desc":    "The results of your certificate validation are shown below",
		"cert":    v,
		"results": r,
		"page":    "Validation",
	})
}

func (s *Server) certExport(c *gin.Context, p string) {
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	var (
		b         []byte
		suffix    string
		extension string
		mime      = "application/x-pem-file"
	)
	switch c.Query("f") {
	case "cert_pem":
		b, err = s.storage.ExportCertificatePEM(p)
		extension = "pem"
	case "cert_der":
		b, err = s.storage.ExportCertificateDER(p)
		extension = "cer"
		mime = "application/pkix-cert"
	case "cert_pkcs7":
		b, err = s.storage.ExportCertificatePKCS7(p)
		extension = "p7b"
		mime = "application/x-pkcs7-certificates"
	case "chain_pem":
		b, err = s.storage.ExportCertificateChainPEM(p)
		suffix = "-chain"
		extension = "pem"
	case "pub_key":
		b, err = s.storage.ExportPublicKeyPEM(p)
		extension = "pub"
	case "priv_key":
		b, err = s.storage.ExportPrivateKeyPEM(p)
		extension = "key"
	default:
		panic(errInvalidFmt)
	}
	if err != nil {
		panic(err)
	}
	downloadCert(c, mime, b, v, suffix, extension)
}

func (s *Server) certPKCS12(c *gin.Context, p string) {
	form := &storage.ExportCertificatePKCS12Params{}
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
		"page":  "Export PKCS#12",
	})
}

func (s *Server) certDelete(c *gin.Context, p string) {
	v, err := s.storage.GetCertificate(p)
	if err != nil {
		panic(err)
	}
	if c.Request.Method == http.MethodPost {
		if err := s.storage.DeleteCertificate(p); err != nil {
			panic(err)
		}
		if len(v.Parents) != 0 {
			c.Redirect(
				http.StatusSeeOther,
				fmt.Sprintf(
					"/%s",
					v.Parents[len(v.Parents)-1].Path,
				),
			)
		} else {
			c.Redirect(http.StatusSeeOther, "/")
		}
		return
	}
	c.HTML(http.StatusOK, "cert_delete.html", pongo2.Context{
		"title": fmt.Sprintf("Delete %s", v.X509.Subject.CommonName),
		"desc":  "Delete certificate and private key",
		"cert":  v,
		"page":  "Delete",
	})
}

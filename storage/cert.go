package storage

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	typeCertificate = "CERTIFICATE"

	filenameCert = "cert.pem"
)

var (
	errNotACert         = errors.New("file is not a PEM-encoded certificate")
	errCertDoesNotExist = errors.New("certificate does not exist")
)

type storageCert struct {
	id          string
	parent      *storageCert
	fingerprint string
	cert        *x509.Certificate
	children    map[string]*storageCert
}

func (s *Storage) loadCerts(
	dir string,
	parent *storageCert,
) (map[string]*storageCert, error) {
	certs := map[string]*storageCert{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return certs, nil
		}
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			c, err := s.loadCert(
				filepath.Join(dir, e.Name()),
				parent,
			)
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}
			certs[c.id] = c
		}
	}
	return certs, nil
}

func (s *Storage) loadCert(
	dir string,
	parent *storageCert,
) (*storageCert, error) {
	b, err := os.ReadFile(filepath.Join(dir, filenameCert))
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != typeCertificate {
		return nil, errNotACert
	}
	x, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	var (
		h = sha256.Sum256(x.Raw)
		c = &storageCert{
			id:          hex.EncodeToString(h[:12]),
			parent:      parent,
			fingerprint: hex.EncodeToString(h[:]),
			cert:        x,
		}
	)
	certs, err := s.loadCerts(dir, c)
	if err != nil {
		return nil, err
	}
	c.children = certs
	return c, nil
}

func (s *Storage) getCert(certPath string) (*storageCert, string, error) {
	var (
		parts = strings.Split(certPath, "/")
		c     *storageCert
		m     = s.rootCerts
		d     = s.certDir
	)
	for _, p := range parts {
		v, ok := m[p]
		if !ok {
			return nil, "", errCertDoesNotExist
		}
		c = v
		m = v.children
		d = filepath.Join(d, v.id)
	}
	return c, d, nil
}

func (s *Storage) createCertificate(
	dir string,
	template, parent *x509.Certificate,
	publicKey *rsa.PublicKey,
	privateKey *rsa.PrivateKey,
) error {
	c, err := x509.CreateCertificate(
		rand.Reader,
		template,
		parent,
		publicKey,
		privateKey,
	)
	if err != nil {
		return err
	}
	b := pem.EncodeToMemory(&pem.Block{
		Type:  typeCertificate,
		Bytes: c,
	})
	f, err := os.OpenFile(
		filepath.Join(dir, filenameCert),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		return err
	}
	return nil
}

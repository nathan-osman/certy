package storage

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

const typeCertificate = "CERTIFICATE"

var errNotACerticate = errors.New("file is not a PEM-encoded certificate")

func loadCertificate(filename string) (*x509.Certificate, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != typeCertificate {
		return nil, errNotACerticate
	}
	return x509.ParseCertificate(block.Bytes)
}

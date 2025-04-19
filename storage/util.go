package storage

import (
	"errors"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

const (
	typeCertificate = "CERTIFICATE"
	typePrivateKey  = "RSA PRIVATE KEY"
)

var (
	errNotACerticate = errors.New("file is not a PEM-encoded certificate")
)

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

func generatePrivateKey(filename string) (*rsa.PrivateKey, error) {
	p, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	b := pem.EncodeToMemory(&pem.Block{
		Type:  typePrivateKey,
		Bytes: x509.MarshalPKCS1PrivateKey(p),
	})
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		return nil, err
	}
	return p, nil
}

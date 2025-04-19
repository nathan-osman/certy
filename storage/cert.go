package storage

import (
	"crypto/rand"
	"crypto/rsa"
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

func createCertificate(
	filename string,
	template, parent *x509.Certificate,
	key *rsa.PrivateKey,
) error {
	c, err := x509.CreateCertificate(
		rand.Reader,
		template,
		parent,
		&key.PublicKey,
		key,
	)
	if err != nil {
		return err
	}
	b := pem.EncodeToMemory(&pem.Block{
		Type:  typeCertificate,
		Bytes: c,
	})
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		return err
	}
	return nil
}

package storage

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

const (
	typePrivateKey = "PRIVATE KEY"
	typePublicKey  = "PUBLIC KEY"

	filenamePrivateKey = "key.pem"
)

var (
	errNotAPrivateKey = errors.New("file is not a PKCS#8 private key")
	errNotAnRSAKey    = errors.New("file is not an RSA private key")
)

func generatePrivateKey(filename string) (*rsa.PrivateKey, error) {
	p, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	b, err := x509.MarshalPKCS8PrivateKey(p)
	if err != nil {
		return nil, err
	}
	block := pem.EncodeToMemory(&pem.Block{
		Type:  typePrivateKey,
		Bytes: b,
	})
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.Write(block); err != nil {
		return nil, err
	}
	return p, nil
}

func loadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != typePrivateKey {
		return nil, errNotAPrivateKey
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	v, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, errNotAnRSAKey
	}
	return v, err
}

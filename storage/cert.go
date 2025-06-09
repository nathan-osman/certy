package storage

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strconv"
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

func (s *Storage) createNextSerial(path string) (int, string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return 0, "", err
	}
	highestSeen := 0
	for _, e := range entries {
		if e.IsDir() {
			v, err := strconv.Atoi(e.Name())
			if err == nil {
				highestSeen = max(highestSeen, v)
			}
		}
	}
	var (
		serial    = highestSeen + 1
		serialStr = strconv.Itoa(serial)
	)
	if err := os.MkdirAll(filepath.Join(s.dataDir, path, serialStr), 0700); err != nil {
		return 0, "", err
	}
	return serial, serialStr, nil
}

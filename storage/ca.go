package storage

import (
	"crypto/x509"
	"os"
	"path/filepath"
)

func (s *Storage) loadCA(dir string) (*x509.Certificate, error) {
	return loadCertificate(filepath.Join(s.dataDir, dir, "cert.pem"))
}

func (s *Storage) loadCAs() ([]*x509.Certificate, error) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, err
	}
	certs := []*x509.Certificate{}
	for _, e := range entries {
		if e.IsDir() {
			c, err := s.loadCA(e.Name())
			if err != nil {
				s.logger.Warn().Msgf("unable to load CA %s", e.Name())
				continue
			}
			certs = append(certs, c)
		}
	}
	return certs, nil
}

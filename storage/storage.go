package storage

import (
	"crypto/x509"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Internally, the directory structure looks something like this:
// - cas/
//   - [UUID]/
//     - cert.pem
//     - key.pem
//     - [SERIAL]/
//       - cert.pem
//       - key.pem

// Storage provides an abstraction to the certificate data stored on disk.
type Storage struct {
	mutex   sync.RWMutex
	logger  zerolog.Logger
	dataDir string
	cas     map[string]*x509.Certificate
}

// New creates a new Storage instance.
func New(dataDir string) (*Storage, error) {
	s := &Storage{
		logger:  log.With().Str("package", "storage").Logger(),
		dataDir: filepath.Join("cas", dataDir),
	}
	certs, err := s.loadCAs()
	if err != nil {
		return nil, err
	}
	s.cas = certs
	return s, nil
}

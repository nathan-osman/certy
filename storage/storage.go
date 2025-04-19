package storage

import (
	"crypto/x509"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Internally, the directory structure looks something like this:
// - ca/
//   - [UUID]/
//     - meta.json
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
	cas     []*x509.Certificate
}

// New creates a new Storage instance.
func New(dataDir string) (*Storage, error) {
	s := &Storage{
		dataDir: dataDir,
		logger:  log.With().Str("package", "storage").Logger(),
	}
	certs, err := s.loadCAs()
	if err != nil {
		return nil, err
	}
	s.cas = certs
	return s, nil
}

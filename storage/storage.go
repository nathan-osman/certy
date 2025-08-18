package storage

import (
	"log/slog"
	"path/filepath"
	"sync"
)

// Internally, the directory structure looks something like this:
//
// - certs/
//   - [SHA-256]/
//     - cert.pem
//     - key.pem
//     - serial
//     - [SHA-256]/
//       - cert.pem
//       - key.pem
//
// A few things to note:
//   - this structure can be arbitrarily deep
//   - serial is present in all intermediate certificates
//   - certificates are identified by their path in the hierarchy:
//     [SHA-256 of root]/[SHA-256 of intermediate]/[SHA-256]

// Storage provides an abstraction to the certificate data stored on disk.
// All public methods are safe for use in multiple goroutines.
type Storage struct {
	mutex     sync.RWMutex
	logger    *slog.Logger
	certDir   string
	rootCerts map[string]*storageCert
}

// New creates a new Storage instance.
func New(cfg *Config) (*Storage, error) {
	s := &Storage{
		logger:  cfg.Logger,
		certDir: filepath.Join(cfg.DataDir, "certs"),
	}
	if s.logger == nil {
		s.logger = slog.Default()
	}
	s.logger = s.logger.With("package", "storage")
	certs, err := s.loadCerts(s.certDir, nil)
	if err != nil {
		return nil, err
	}
	s.rootCerts = certs
	return s, nil
}

package storage

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func (s *Storage) loadCA(dir string) (*x509.Certificate, error) {
	return loadCertificate(filepath.Join(s.dataDir, dir, "cert.pem"))
}

func (s *Storage) loadCAs() (map[string]*x509.Certificate, error) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]*x509.Certificate{}, nil
		}
		return nil, err
	}
	certs := map[string]*x509.Certificate{}
	for _, e := range entries {
		if e.IsDir() {
			c, err := s.loadCA(e.Name())
			if err != nil {
				s.logger.Warn().Msgf("unable to load CA %s", e.Name())
				continue
			}
			certs[e.Name()] = c
		}
	}
	return certs, nil
}

// TODO: sort this list deterministically?

// ListCAs returns a list of CAs loaded from disk. This function is
// thread-safe.
func (s *Storage) ListCAs() []*x509.Certificate {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	cas := []*x509.Certificate{}
	for _, v := range s.cas {
		cas = append(cas, v)
	}
	return cas
}

// CreateCAParams provides CreateCA with parameters for the new CA.
type CreateCAParams struct {
	CommonName    string        `form:"common_name"`
	Organization  string        `form:"organization"`
	Validity      time.Duration `form:"validity"`
	UsageCertSign bool          `form:"usage_cert_sign"`
}

// TODO: delete the intermediate files if something fails in this method
// maybe using defer?

// CreateCA creates a new CA on disk using the provided data.
func (s *Storage) CreateCA(params *CreateCAParams) error {
	var (
		uuid = uuid.New().String()
		dir  = filepath.Join(s.dataDir, uuid)
	)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	p, err := generatePrivateKey(filepath.Join(dir, "key.pem"))
	if err != nil {
		return err
	}
	var (
		n = time.Now()
		c = &x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				CommonName:   params.CommonName,
				Organization: []string{params.Organization},
			},
			NotBefore:             n,
			NotAfter:              n.Add(params.Validity),
			BasicConstraintsValid: true,
			IsCA:                  true,
		}
	)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := createCertificate(
		filepath.Join(dir, "cert.pem"),
		c, c, p,
	); err != nil {
		return err
	}
	s.cas[uuid] = c
	return nil
}

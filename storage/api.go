package storage

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

// Ref stores an ID and certificate pair.
type Ref struct {
	ID   string
	X509 *x509.Certificate
}

// Certificate represents an X.509 certificate in a format suitable for
// rendering to templates.
type Certificate struct {
	Parents     []*Ref
	Fingerprint string
	X509        *x509.Certificate
	Children    []*Ref
}

func parentList(p *storageCert) []*Ref {
	parents := []*Ref{}
	for p != nil {
		parents = append([]*Ref{
			&Ref{
				ID:   p.id,
				X509: p.cert,
			},
		}, parents...)
		p = p.parent
	}
	return parents
}

func childList(m map[string]*storageCert) []*Ref {
	children := []*Ref{}
	for k, v := range m {
		children = append(children, &Ref{
			ID:   k,
			X509: v.cert,
		})
	}
	return children
}

func convertCert(cert *storageCert) *Certificate {
	return &Certificate{
		Parents:     parentList(cert.parent),
		Fingerprint: cert.fingerprint,
		X509:        cert.cert,
		Children:    childList(cert.children),
	}
}

// GetRootCertificates returns the root certificates.
func (s *Storage) GetRootCertificates() []*Ref {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return childList(s.rootCerts)
}

// GetCertificate attempts to return a certificate by its path.
func (s *Storage) GetCertificate(certPath string) (*Certificate, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, _, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	return convertCert(c), nil
}

// ExportCertificatePEM exports the specified certificate as a PEM-encoded
// file.
func (s *Storage) ExportCertificatePEM(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, _, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  typeCertificate,
		Bytes: c.cert.Raw,
	}), nil
}

// ExportCertificateChainPEM exports the specified certificate and its parents
// as a PEM-encoded file.
func (s *Storage) ExportCertificateChainPEM(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, _, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	var b []byte
	for c != nil {
		v := pem.EncodeToMemory(&pem.Block{
			Type:  typeCertificate,
			Bytes: c.cert.Raw,
		})
		b = append(v, b...)
		c = c.parent
	}
	return b, nil
}

// ExportCertificatePKCS12Params provides ExportCertificatePKCS12 with
// parameters for exporting a certificate and its private key.
type ExportCertificatePKCS12Params struct {
	Password string
}

// ExportCertificatePKCS12 exports the specified certificate and its private
// key and CAs as a PKCS#12 file.
func (s *Storage) ExportCertificatePKCS12(
	certPath string,
	params *ExportCertificatePKCS12Params,
) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, d, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	k, err := loadPrivateKey(filepath.Join(d, filenamePrivateKey))
	if err != nil {
		return nil, err
	}
	var (
		v     = c.parent
		certs = []*x509.Certificate{}
	)
	for v != nil {
		certs = append([]*x509.Certificate{v.cert}, certs...)
		v = v.parent
	}
	return pkcs12.Modern.Encode(k, c.cert, certs, params.Password)
}

// ExportPublicKeyPEM exports the public key of the specified certificate as a
// PEM-encoded file.
func (s *Storage) ExportPublicKeyPEM(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, d, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	k, err := loadPrivateKey(filepath.Join(d, filenamePrivateKey))
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  typePublicKey,
		Bytes: x509.MarshalPKCS1PublicKey(&k.PublicKey),
	}), nil
}

// ExportPrivateKeyPEM exports the private key of the specified certificate as
// a PEM-encoded file.
func (s *Storage) ExportPrivateKeyPEM(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, d, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(d, filenamePrivateKey))
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  typePrivateKey,
		Bytes: b,
	}), nil
}

// CreateCertificateParams provides CreateCertificate with parameters for
// creating a new X.509 certificate and private key.
type CreateCertificateParams struct {
	CommonName   string
	Organization string
	Validity     string
}

// CreateCertificate creates a new certificate & private key. The certificate
// path to the newly created certificate is returned upon success.
func (s *Storage) CreateCertificate(
	certPath string,
	params *CreateCertificateParams,
) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Begin by loading the parent certificate (if supplied)
	var (
		p         *storageCert
		parentDir = s.certDir
	)
	if certPath != "" {
		v, vDir, err := s.getCert(certPath)
		if err != nil {
			return "", err
		}
		p = v
		parentDir = vDir
	}

	// The directory for the certificate and private key needs to be created
	// before we know the certificate's ID (fingerprint), so we create a
	// temporary directory and then rename it afterwards; note that the defer
	// call to remove the directory will be a no-op on success since the
	// directory will no longer exist under its temp name
	d, err := os.MkdirTemp(parentDir, "temp")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(d)

	// Parse the validity duration
	v, err := parseDuration(params.Validity)
	if err != nil {
		return "", err
	}

	// Generate a new private key
	k, err := generatePrivateKey(filepath.Join(d, filenamePrivateKey))
	if err != nil {
		return "", err
	}

	// Use the new key if this is a root CA; otherwise, load the parent's
	certPrivateKey := k
	if p != nil {
		k, err := loadPrivateKey(filepath.Join(parentDir, filenamePrivateKey))
		if err != nil {
			return "", err
		}
		certPrivateKey = k
	}

	// Use 1 for the serial if this is a root CA; otherwise, allocate the next
	// serial number from the parent
	var serial int64 = 1
	if p != nil {
		v, err := s.allocNextSerial(parentDir)
		if err != nil {
			return "", err
		}
		serial = v
	}

	// Create the certificate template
	var (
		n    = time.Now()
		cert = &x509.Certificate{
			SerialNumber: big.NewInt(serial),
			Subject: pkix.Name{
				CommonName:   params.CommonName,
				Organization: []string{params.Organization},
			},
			NotBefore:             n,
			NotAfter:              n.Add(v),
			KeyUsage:              x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			BasicConstraintsValid: true,
			IsCA:                  p == nil,
		}
	)

	// Use the certificate as its own parent if this is a root CA; otherwise,
	// use the parent's certificate
	parentCert := cert
	if p != nil {
		parentCert = p.cert
	}

	// FINALLY, create the actual certificate
	if err := s.createCertificate(
		d,
		cert,
		parentCert,
		&k.PublicKey,
		certPrivateKey,
	); err != nil {
		return "", err
	}

	// ...and load it from disk
	c, err := s.loadCert(d, p)
	if err != nil {
		return "", err
	}

	// The order of the next two tasks is important - the rename should be the
	// last action that can fail (return error) since (basically) everything
	// up until this point will be destroyed by the defer RemoveAll() call
	// above on failure; and adding the storageCert to its parent should only
	// be done when the layout on disk is complete

	// Rename the directory to the certificate's ID
	if err := os.Rename(d, filepath.Join(parentDir, c.id)); err != nil {
		return "", err
	}

	// ...and add it to the internal map
	if p == nil {
		s.rootCerts[c.id] = c
	} else {
		p.children[c.id] = c
	}

	// Return the path to the new certificate
	if certPath == "" {
		return c.id, nil
	}
	return fmt.Sprintf("%s/%s", certPath, c.id), nil
}

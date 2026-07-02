package storage

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"go.mozilla.org/pkcs7"
	"software.sslmate.com/src/go-pkcs12"
)

// Ref stores an ID and certificate data.
type Ref struct {
	ID   string
	Path string
	X509 *x509.Certificate
}

// Private key holds information about a certificate's private key.
type PrivateKey struct {
	Size int
}

// Certificate represents an X.509 certificate in a format suitable for
// rendering to templates.
type Certificate struct {
	ID          string
	Path        string
	Parents     []*Ref
	Fingerprint string
	X509        *x509.Certificate
	Children    []*Ref
	PrivateKey  *PrivateKey
}

// IsExpired indicates whether the certificate is expired or not.
func (c *Certificate) IsExpired() bool {
	return c.X509.NotAfter.Before(time.Now())
}

// MaySign indicates whether this certificate may sign others. This does not
// necessarily mean that it *can* sign certificates, if for example the
// private key does not exist.
func (c *Certificate) MaySign() bool {
	return c.X509.IsCA && c.X509.KeyUsage&x509.KeyUsageCertSign != 0
}

// CanSign indicates whether this certificate has the ability to sign others
// (a private key exists on disk).
func (c *Certificate) CanSign() bool {
	return c.MaySign() && c.PrivateKey != nil
}

// KeyUsage provides a human-friendly list of possible uses.
func (c *Certificate) KeyUsage() []string {
	usages := []string{}
	if c.X509.KeyUsage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "certificate signing")
	}
	if c.X509.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "digital signature")
	}
	if c.X509.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "key encipherment")
	}
	for _, u := range c.X509.ExtKeyUsage {
		switch u {
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "client auth")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "server auth")
		}
	}
	return usages
}

func parentList(p *storageCert) []*Ref {
	var (
		parents = []*Ref{}
		v       = p
	)
	for v != nil {
		parents = append([]*Ref{
			{
				ID:   v.id,
				Path: v.vPath,
				X509: v.cert,
			},
		}, parents...)
		v = v.parent
	}
	return parents
}

func childList(m map[string]*storageCert) []*Ref {
	children := []*Ref{}
	for k, v := range m {
		children = append(children, &Ref{
			ID:   k,
			Path: v.vPath,
			X509: v.cert,
		})
	}
	return children
}

func convertCert(cert *storageCert, key *rsa.PrivateKey) *Certificate {
	c := &Certificate{
		ID:          cert.id,
		Path:        cert.vPath,
		Parents:     parentList(cert.parent),
		Fingerprint: cert.fingerprint,
		X509:        cert.cert,
		Children:    childList(cert.children),
	}
	if key != nil {
		c.PrivateKey = &PrivateKey{
			Size: key.Size() * 8,
		}
	}
	return c
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
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	var key *rsa.PrivateKey
	if c.hasKey {
		k, err := loadPrivateKey(filepath.Join(c.fPath, filenamePrivateKey))
		if err != nil {
			return nil, err
		}
		key = k
	}
	return convertCert(c, key), nil
}

// ValidationResult indicates the validity of a single certificate in a chain
// represented by Err being nil or not.
type ValidationResult struct {
	X509 *x509.Certificate
	Err  string
}

// ValidateCertificate attempts to validate the specified certificate. The
// result is returned as a slice indicating the validity of each link in the
// chain of trust.
func (s *Storage) ValidateCertificate(certPath string) ([]*ValidationResult, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	var (
		results    []*ValidationResult
		pRoot      = x509.NewCertPool()
		pImed      = x509.NewCertPool()
		foundError bool
	)
	for i, c := range c.chain() {
		result := &ValidationResult{
			X509: c.cert,
		}
		if i == 0 {
			pRoot.AddCert(c.cert)
		}
		if _, err := c.cert.Verify(x509.VerifyOptions{
			Roots:         pRoot,
			Intermediates: pImed,
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		}); err != nil {
			if foundError {
				result.Err = "cannot validate because parent certificate failed to validate"
			} else {
				result.Err = err.Error()
				foundError = true
			}
		}
		results = append(results, result)
		if i > 0 {
			pImed.AddCert(c.cert)
		}
	}
	return results, nil
}

// ExportCertificatePEM exports the specified certificate as a PEM-encoded
// file.
func (s *Storage) ExportCertificatePEM(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  typeCertificate,
		Bytes: c.cert.Raw,
	}), nil
}

// ExportCertificateDER exports the specified certificate in DER format.
func (s *Storage) ExportCertificateDER(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	return c.cert.Raw, nil
}

// ExportCertificatePKCS7 exports the specified certificate in PKCS#7 format.
func (s *Storage) ExportCertificatePKCS7(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	d, err := pkcs7.NewSignedData(nil)
	if err != nil {
		return nil, err
	}
	d.AddCertificate(c.cert)
	return d.Finish()
}

// ExportCertificateChainPEM exports the specified certificate and its parents
// (not including the root) as a PEM-encoded file.
func (s *Storage) ExportCertificateChainPEM(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	var b []byte
	for c != nil && c.parent != nil {
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
	Password  string
	UseLegacy bool
}

// ExportCertificatePKCS12 exports the specified certificate and its private
// key and CAs as a PKCS#12 file.
func (s *Storage) ExportCertificatePKCS12(
	certPath string,
	params *ExportCertificatePKCS12Params,
) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	k, err := loadPrivateKey(filepath.Join(c.fPath, filenamePrivateKey))
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
	var encoder *pkcs12.Encoder
	if params.UseLegacy {
		encoder = pkcs12.LegacyDES
	} else {
		encoder = pkcs12.Modern
	}
	return encoder.Encode(k, c.cert, certs, params.Password)
}

// ExportPublicKeyPEM exports the public key of the specified certificate as a
// PEM-encoded file.
func (s *Storage) ExportPublicKeyPEM(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	k, err := loadPrivateKey(filepath.Join(c.fPath, filenamePrivateKey))
	if err != nil {
		return nil, err
	}
	b, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  typePublicKey,
		Bytes: b,
	}), nil
}

// ExportPrivateKeyPEM exports the private key of the specified certificate as
// a PEM-encoded file.
func (s *Storage) ExportPrivateKeyPEM(certPath string) ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	c, err := s.getCert(certPath)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(c.fPath, filenamePrivateKey))
	if err != nil {
		return nil, err
	}
	return b, nil
}

// CreateCertificateParams provides CreateCertificate with parameters for
// creating a new X.509 certificate and private key.
type CreateCertificateParams struct {
	CommonName         string
	Organization       string
	OrganizationalUnit string
	Country            string
	Province           string
	Locality           string
	StreetAddress      string
	PostalCode         string
	Validity           string
	CanSign            bool
	AllowChaining      bool
	CodeSigning        bool
	ClientAuth         bool
	ServerAuth         bool
	SANs               string
	KeySize            int
}

// CreateCertificate creates a new certificate & private key. The newly
// created certificate is returned upon success.
func (s *Storage) CreateCertificate(
	certPath string,
	params *CreateCertificateParams,
) (*Certificate, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Begin by loading the parent certificate (if supplied)
	var (
		p         *storageCert
		parentDir = s.certDir
	)
	if certPath != "" {
		v, err := s.getCert(certPath)
		if err != nil {
			return nil, err
		}
		p = v
		parentDir = v.fPath
	}

	// The directory for the certificate and private key needs to be created
	// before we know the certificate's ID (fingerprint), so we create a
	// temporary directory and then rename it afterwards; note that the defer
	// call to remove the directory will be a no-op on success since the
	// directory will no longer exist under its temp name
	d, err := os.MkdirTemp(parentDir, "temp")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(d)

	// Parse the validity duration
	v, err := parseDuration(params.Validity)
	if err != nil {
		return nil, err
	}

	// Generate a new private key
	k, err := generatePrivateKey(
		filepath.Join(d, filenamePrivateKey),
		params.KeySize,
	)
	if err != nil {
		return nil, err
	}

	// Use the new key if this is a root CA; otherwise, load the parent's
	certPrivateKey := k
	if p != nil {
		k, err := loadPrivateKey(filepath.Join(parentDir, filenamePrivateKey))
		if err != nil {
			return nil, err
		}
		certPrivateKey = k
	}

	// Use 1 for the serial if this is a root CA; otherwise, allocate the next
	// serial number from the parent
	var serial int64 = 1
	if p != nil {
		v, err := s.allocNextSerial(parentDir)
		if err != nil {
			return nil, err
		}
		serial = v
	}

	// Create the certificate template
	var (
		n    = time.Now()
		cert = &x509.Certificate{
			SerialNumber: big.NewInt(serial),
			Subject: pkix.Name{
				Country:            ifProvided(params.Country),
				Organization:       ifProvided(params.Organization),
				OrganizationalUnit: ifProvided(params.OrganizationalUnit),
				Locality:           ifProvided(params.Locality),
				Province:           ifProvided(params.Province),
				StreetAddress:      ifProvided(params.StreetAddress),
				PostalCode:         ifProvided(params.PostalCode),
				CommonName:         params.CommonName,
			},
			NotBefore:             n,
			NotAfter:              n.Add(v),
			BasicConstraintsValid: true,
			IsCA:                  params.CanSign,
		}
	)

	// Use the certificate as its own parent if this is a root CA; otherwise,
	// use the parent's certificate
	parentCert := cert
	if p != nil {
		parentCert = p.cert
	}

	// Set the flags
	if params.CanSign {
		cert.KeyUsage |= x509.KeyUsageCertSign
	}
	if cert.IsCA && !params.AllowChaining {
		cert.MaxPathLenZero = true
	}
	if params.CodeSigning {
		cert.KeyUsage |= x509.KeyUsageDigitalSignature
		cert.ExtKeyUsage = append(
			cert.ExtKeyUsage,
			x509.ExtKeyUsageCodeSigning,
		)
	}
	if params.ClientAuth {
		cert.KeyUsage |= x509.KeyUsageDigitalSignature
		cert.ExtKeyUsage = append(
			cert.ExtKeyUsage,
			x509.ExtKeyUsageClientAuth,
		)
	}
	if params.ServerAuth {
		cert.KeyUsage |=
			x509.KeyUsageDigitalSignature |
				x509.KeyUsageKeyEncipherment
		cert.ExtKeyUsage = append(
			cert.ExtKeyUsage,
			x509.ExtKeyUsageServerAuth,
		)
	}

	// If SANs were provided (usually required for web servers), include them
	// as well; check each value to see if it is an IP address or domain
	if params.SANs != "" {
		for _, v := range strings.FieldsFunc(
			params.SANs,
			func(c rune) bool {
				return c == ',' || unicode.IsSpace(c)
			},
		) {
			i := net.ParseIP(v)
			if i != nil {
				cert.IPAddresses = append(cert.IPAddresses, i)
			} else {
				cert.DNSNames = append(cert.DNSNames, v)
			}
		}
	}

	// FINALLY, create the actual certificate
	if err := s.createCertificate(
		d,
		cert,
		parentCert,
		&k.PublicKey,
		certPrivateKey,
	); err != nil {
		return nil, err
	}

	// ...and load it from disk
	c, err := s.loadCert(d, p)
	if err != nil {
		return nil, err
	}

	// The order of the next two tasks is important - the rename should be the
	// last action that can fail (return error) since (basically) everything
	// up until this point will be destroyed by the defer RemoveAll() call
	// above on failure; and adding the storageCert to its parent should only
	// be done when the layout on disk is complete

	// Rename the directory to the certificate's ID
	newDir := filepath.Join(parentDir, c.id)
	if err := os.Rename(d, newDir); err != nil {
		return nil, err
	}

	// ...update the certificate's fPath to point to the new directory...
	c.fPath = newDir

	// ...and add it to the internal map
	if p == nil {
		s.rootCerts[c.id] = c
	} else {
		p.children[c.id] = c
	}

	// Return the new certificate
	return convertCert(c, k), nil
}

// DeleteCertificate removes a certificate and its private key from disk. Note
// that this will also delete all stored certificates and private keys signed
// by it. This will not revoke the certificate.
func (s *Storage) DeleteCertificate(
	certPath string,
) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find the certificate and its directory on disk
	c, err := s.getCert(certPath)
	if err != nil {
		return err
	}

	// Remove it from disk
	if err := os.RemoveAll(c.fPath); err != nil {
		return err
	}

	// Remove it from the internal map
	if c.parent != nil {
		delete(c.parent.children, c.id)
	} else {
		delete(s.rootCerts, c.id)
	}

	// Successfully deleted
	return nil
}

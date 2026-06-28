package storage

import (
	"bytes"
	"encoding/pem"
	"errors"
	"testing"
)

const (
	rootCertCN  = "Root CA"
	childCertCN = "service.example.test"
	childCertIP = "127.0.0.1"
)

func newTestStorage(t *testing.T, dataDir string) *Storage {
	t.Helper()
	s, err := New(&Config{
		DataDir: dataDir,
	})
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	return s
}

func TestStorageCreatesPersistsExportsAndDeletesCertificates(t *testing.T) {
	var (
		dataDir = t.TempDir()
		s       = newTestStorage(t, dataDir)
	)

	// Create a root certificate
	_, err := s.CreateCertificate("", &CreateCertificateParams{
		CommonName:    rootCertCN,
		Validity:      "1h",
		CanSign:       true,
		AllowChaining: true,
	})
	if err != nil {
		t.Fatalf("create root certificate: %v", err)
	}

	// ...and confirm it was added to the internal map
	rootCerts := s.GetRootCertificates()
	if len(rootCerts) != 1 {
		t.Fatalf("root certificate count = %d, want 1", len(rootCerts))
	}

	// Compare the root CA to the input
	rootCert, err := s.GetCertificate(rootCerts[0].Path)
	if err != nil {
		t.Fatalf("get root certificate: %v", err)
	}
	if rootCert.X509.Subject.CommonName != rootCertCN {
		t.Fatalf(
			"root common name = %q, want %q",
			rootCert.X509.Subject.CommonName,
			rootCertCN,
		)
	}
	if !rootCert.CanSign() {
		t.Fatal("root certificate should be able to sign")
	}

	// Create a child of the root CA
	c, err := s.CreateCertificate(rootCert.Path, &CreateCertificateParams{
		CommonName: childCertCN,
		Validity:   "30m",
		ServerAuth: true,
		SANs:       childCertCN + " " + childCertIP,
	})
	if err != nil {
		t.Fatalf("create child certificate: %v", err)
	}

	// Retrieve the newly created child certificate
	childCert, err := s.GetCertificate(c.Path)
	if err != nil {
		t.Fatalf("get child certificate: %v", err)
	}

	// Confirm its parents, attributes, & validity
	if len(childCert.Parents) != 1 || childCert.Parents[0].ID != rootCert.ID {
		t.Fatalf("child parents = %#v, want root %q", childCert.Parents, rootCert.ID)
	}
	if childCert.CanSign() {
		t.Fatal("leaf certificate should not be able to sign")
	}
	if err := s.ValidateCertificate(childCert.Path); err != nil {
		t.Fatalf("validate child certificate: %v", err)
	}

	// Attempt to export the child certificate in PEM format
	pemBytes, err := s.ExportCertificatePEM(childCert.Path)
	if err != nil {
		t.Fatalf("export child certificate PEM: %v", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != typeCertificate {
		t.Fatalf("exported PEM block = %#v, want certificate block", block)
	}

	// Attempt to export the child certificate's public key
	publicKey, err := s.ExportPublicKeyPEM(childCert.Path)
	if err != nil {
		t.Fatalf("export child public key: %v", err)
	}
	if !bytes.Contains(publicKey, []byte("BEGIN PUBLIC KEY")) {
		t.Fatalf("public key export does not contain a PEM public key: %q", publicKey)
	}

	// Reload the storage directory and make sure everything is still there
	s = newTestStorage(t, dataDir)
	if _, err := s.GetCertificate(childCert.Path); err != nil {
		t.Fatalf("get child certificate after reload: %v", err)
	}

	// Delete the root certificate
	if err := s.DeleteCertificate(rootCert.Path); err != nil {
		t.Fatalf("delete root certificate: %v", err)
	}

	// Make sure the root and child are gone
	if _, err := s.GetCertificate(rootCert.Path); !errors.Is(err, errCertDoesNotExist) {
		t.Fatalf("get deleted root error = %v, want %v", err, errCertDoesNotExist)
	}
	if _, err := s.GetCertificate(rootCert.Path); !errors.Is(err, errCertDoesNotExist) {
		t.Fatalf("get deleted child error = %v, want %v", err, errCertDoesNotExist)
	}
}

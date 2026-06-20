package storage

import (
	"bytes"
	"encoding/pem"
	"errors"
	"testing"
)

func TestStorageCreatesPersistsExportsAndDeletesCertificates(t *testing.T) {
	dataDir := t.TempDir()
	s := newTestStorage(t, dataDir)

	rootID, err := s.CreateCertificate("", &CreateCertificateParams{
		CommonName:    "Root CA",
		Organization:  "Example Org",
		Validity:      "1h",
		CanSign:       true,
		AllowChaining: true,
	})
	if err != nil {
		t.Fatalf("create root certificate: %v", err)
	}

	roots := s.GetRootCertificates()
	if len(roots) != 1 {
		t.Fatalf("root certificate count = %d, want 1", len(roots))
	}
	if roots[0].ID != rootID {
		t.Fatalf("root ID = %q, want %q", roots[0].ID, rootID)
	}

	root, err := s.GetCertificate(rootID)
	if err != nil {
		t.Fatalf("get root certificate: %v", err)
	}
	if root.X509.Subject.CommonName != "Root CA" {
		t.Fatalf("root common name = %q, want %q", root.X509.Subject.CommonName, "Root CA")
	}
	if !root.CanSign() {
		t.Fatal("root certificate should be able to sign")
	}

	childID, err := s.CreateCertificate(rootID, &CreateCertificateParams{
		CommonName: "service.example.test",
		Validity:   "30m",
		ServerAuth: true,
		SANs:       "service.example.test 127.0.0.1",
	})
	if err != nil {
		t.Fatalf("create child certificate: %v", err)
	}

	child, err := s.GetCertificate(childID)
	if err != nil {
		t.Fatalf("get child certificate: %v", err)
	}
	if len(child.Parents) != 1 || child.Parents[0].ID != rootID {
		t.Fatalf("child parents = %#v, want root %q", child.Parents, rootID)
	}
	if child.CanSign() {
		t.Fatal("leaf certificate should not be able to sign")
	}
	if err := s.ValidateCertificate(childID); err != nil {
		t.Fatalf("validate child certificate: %v", err)
	}

	pemBytes, err := s.ExportCertificatePEM(childID)
	if err != nil {
		t.Fatalf("export child certificate PEM: %v", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != typeCertificate {
		t.Fatalf("exported PEM block = %#v, want certificate block", block)
	}

	publicKey, err := s.ExportPublicKeyPEM(childID)
	if err != nil {
		t.Fatalf("export child public key: %v", err)
	}
	if !bytes.Contains(publicKey, []byte("BEGIN PUBLIC KEY")) {
		t.Fatalf("public key export does not contain a PEM public key: %q", publicKey)
	}

	reloaded := newTestStorage(t, dataDir)
	if _, err := reloaded.GetCertificate(childID); err != nil {
		t.Fatalf("get child certificate after reload: %v", err)
	}

	if err := reloaded.DeleteCertificate(rootID); err != nil {
		t.Fatalf("delete root certificate: %v", err)
	}
	if _, err := reloaded.GetCertificate(rootID); !errors.Is(err, errCertDoesNotExist) {
		t.Fatalf("get deleted root error = %v, want %v", err, errCertDoesNotExist)
	}
	if _, err := reloaded.GetCertificate(childID); !errors.Is(err, errCertDoesNotExist) {
		t.Fatalf("get deleted child error = %v, want %v", err, errCertDoesNotExist)
	}
}

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

package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllocNextSerialStartsAtOneAndIncrements(t *testing.T) {
	t.Parallel()

	s := &Storage{}
	dir := t.TempDir()

	first, err := s.allocNextSerial(dir)
	if err != nil {
		t.Fatalf("allocNextSerial first call returned error: %v", err)
	}
	if first != 1 {
		t.Fatalf("first serial = %d, want 1", first)
	}

	second, err := s.allocNextSerial(dir)
	if err != nil {
		t.Fatalf("allocNextSerial second call returned error: %v", err)
	}
	if second != 2 {
		t.Fatalf("second serial = %d, want 2", second)
	}

	serialFile := filepath.Join(dir, filenameSerial)
	contents, err := os.ReadFile(serialFile)
	if err != nil {
		t.Fatalf("read serial file: %v", err)
	}
	if string(contents) != "2" {
		t.Fatalf("serial file = %q, want %q", contents, "2")
	}
}

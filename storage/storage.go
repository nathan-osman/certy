package storage

// Internally, the directory structure looks something like this:
// - ca/
//   - [UUID]/
//     - cert.pem
//     - key.pem
//     - [UUID]/
//       - cert.pem
//       - key.pem

// Storage provides an abstraction to the certificate data stored on disk.
type Storage struct {
	dataDir string
}

// New creates a new Storage instance.
func New(dataDir string) *Storage {
	return &Storage{
		dataDir: dataDir,
	}
}

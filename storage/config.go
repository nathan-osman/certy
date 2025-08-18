package storage

import (
	"log/slog"
)

// Config provides configuration for Storage.
type Config struct {

	// DataDir specifies where the data for the application should be stored.
	// An empty value indicates the current directory.
	DataDir string

	// Logger can be used to capture log messages.
	Logger *slog.Logger
}

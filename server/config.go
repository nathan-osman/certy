package server

import (
	"log/slog"

	"github.com/nathan-osman/certy/storage"
)

// Config provides configuration for Server.
type Config struct {

	// Addr is the address the server should listen on.
	Addr string

	// Logger can be used to capture log messages.
	Logger *slog.Logger

	// Storage is a pointer to a Storage instance
	Storage *storage.Storage
}

package minivmm

import (
	"os"
	"path/filepath"
)

const (
	// EnvDir is a environment variable key.
	EnvDir = "VMM_DIR"
	// EnvPort is a environment variable key.
	EnvPort = "VMM_LISTEN_PORT"
	// EnvTLS is a environment variable key.
	EnvTLS = "VMM_USE_TLS"
	// EnvOrigin is a environment variable key.
	EnvOrigin = "VMM_ORIGIN"
	// EnvOIDC is a environment variable key.
	EnvOIDC = "VMM_OIDC_URL"
	// EnvAgents is a environment variable key.
	EnvAgents = "VMM_AGENTS"
	// EnvNoAuth is a environment variable key.
	EnvNoAuth = "VMM_NO_AUTH"
	// EnvCorsOrigins is a environment variable key.
	EnvCorsOrigins = "VMM_CORS_ALLOWED_ORIGINS"
	// EnvNameServers is a environment variable key.
	EnvNameServers = "VMM_NAME_SERVERS"
	// EnvNoKvm is a environment variable key.
	EnvNoKvm = "VMM_NO_KVM"
)

var (
	// ForwardDir is a directory path for the fowarder's metadata files.
	ForwardDir = filepath.Join(os.Getenv(EnvDir), "forwards")
	// VMDir is a directory path for the files associated to virtual machines.
	VMDir = filepath.Join(os.Getenv(EnvDir), "vms")
	// ImageDir is a directory path for the base image files.
	ImageDir = filepath.Join(os.Getenv(EnvDir), "images")
)

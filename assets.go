package minivmm

import (
	"embed"
	"io/fs"
)

//go:embed web/dist
var assets embed.FS

// GetAssets return the assets of Web UI.
func GetAssets() fs.FS {
	s, _ := fs.Sub(assets, "web/dist")
	return s
}

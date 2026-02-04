// Package data provides embedded game data and utilities for loading it.
package data

import "embed"

// dataFS embeds all JSON files from the data directory at build time.
//
//go:embed *.json
var dataFS embed.FS

// FS returns the embedded filesystem containing game data.
func FS() embed.FS {
	return dataFS
}

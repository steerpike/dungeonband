// Package gamedata provides embedded game data and utilities for loading it.
package gamedata

import "embed"

// dataFS embeds all JSON files from this directory at build time.
//
//go:embed *.json
var dataFS embed.FS

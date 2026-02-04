package gamedata

import (
	"encoding/json"
	"fmt"
)

// Load reads and unmarshals a JSON file from the embedded filesystem.
func Load[T any](filename string) (T, error) {
	var result T

	content, err := dataFS.ReadFile(filename)
	if err != nil {
		return result, fmt.Errorf("failed to read embedded file %s: %w", filename, err)
	}

	if err := json.Unmarshal(content, &result); err != nil {
		return result, fmt.Errorf("failed to parse JSON from %s: %w", filename, err)
	}

	return result, nil
}

// MustLoad reads and unmarshals a JSON file, panicking on error.
// Use this for data that must be present for the game to function.
func MustLoad[T any](filename string) T {
	result, err := Load[T](filename)
	if err != nil {
		panic(err)
	}
	return result
}

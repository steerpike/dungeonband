package data

import "github.com/gdamore/tcell/v2"

// EnemyDef defines an enemy type loaded from JSON.
type EnemyDef struct {
	ID          string `json:"id"`          // Unique identifier (e.g., "goblin")
	Name        string `json:"name"`        // Display name (e.g., "Goblin")
	Glyph       string `json:"glyph"`       // Single character for rendering (e.g., "g")
	Color       string `json:"color"`       // Hex color code (e.g., "#00FF00")
	HP          int    `json:"hp"`          // Base hit points
	Attack      int    `json:"attack"`      // Base attack power
	Defense     int    `json:"defense"`     // Base defense value
	SpawnWeight int    `json:"spawnWeight"` // Relative spawn frequency (higher = more common)
}

// GlyphRune returns the glyph as a rune for rendering.
func (e *EnemyDef) GlyphRune() rune {
	if len(e.Glyph) == 0 {
		return '?'
	}
	return rune(e.Glyph[0])
}

// TCellColor returns the color as a tcell.Color.
func (e *EnemyDef) TCellColor() tcell.Color {
	color, err := ParseHexColor(e.Color)
	if err != nil {
		return tcell.ColorWhite // fallback
	}
	return color
}

// EnemiesFile represents the structure of enemies.json.
type EnemiesFile struct {
	Enemies []EnemyDef `json:"enemies"`
}

// LoadEnemies loads enemy definitions from the embedded enemies.json file.
func LoadEnemies() ([]EnemyDef, error) {
	file, err := Load[EnemiesFile]("enemies.json")
	if err != nil {
		return nil, err
	}
	return file.Enemies, nil
}

// MustLoadEnemies loads enemy definitions, panicking on error.
func MustLoadEnemies() []EnemyDef {
	enemies, err := LoadEnemies()
	if err != nil {
		panic(err)
	}
	return enemies
}

// Package world provides dungeon generation and map management.
package world

// Tile represents a single map tile.
type Tile rune

const (
	// TileWall represents an impassable wall tile.
	TileWall Tile = '#'
	// TileFloor represents a passable floor tile.
	TileFloor Tile = '.'
)

// IsPassable returns true if the tile can be walked on.
func (t Tile) IsPassable() bool {
	return t == TileFloor
}

// Rune returns the tile's display character.
func (t Tile) Rune() rune {
	return rune(t)
}

// Package entity provides game entities like the party and monsters.
package entity

import (
	"github.com/gdamore/tcell/v2"

	"github.com/samdwyer/dungeonband/internal/gamedata"
)

// EnemyType represents a type of enemy.
// Deprecated: Use EnemyDef from data package instead.
type EnemyType int

const (
	EnemyGoblin EnemyType = iota
	EnemyOrc
	EnemySkeleton
)

// String returns the enemy type name.
func (t EnemyType) String() string {
	switch t {
	case EnemyGoblin:
		return "Goblin"
	case EnemyOrc:
		return "Orc"
	case EnemySkeleton:
		return "Skeleton"
	default:
		return "Unknown"
	}
}

// Symbol returns the default display symbol for an enemy type.
func (t EnemyType) Symbol() rune {
	switch t {
	case EnemyGoblin:
		return 'g'
	case EnemyOrc:
		return 'o'
	case EnemySkeleton:
		return 's'
	default:
		return '?'
	}
}

// Enemy represents a hostile creature in the dungeon.
type Enemy struct {
	Def       *gamedata.EnemyDef // Reference to the enemy definition (nil for legacy enemies)
	Name      string             // Enemy name (e.g., "Goblin Scout")
	Type      EnemyType          // Type of enemy (deprecated, use Def)
	Symbol    rune               // Display symbol
	X, Y      int                // Position in the dungeon
	RoomIndex int                // Index of the room this enemy is in (-1 if not in a room)
	HP        int                // Current hit points
	MaxHP     int                // Maximum hit points
}

// NewEnemy creates a new enemy of the given type at the specified position.
// Deprecated: Use NewEnemyFromDef instead.
func NewEnemy(enemyType EnemyType, x, y, roomIndex int) *Enemy {
	return &Enemy{
		Name:      enemyType.String(),
		Type:      enemyType,
		Symbol:    enemyType.Symbol(),
		X:         x,
		Y:         y,
		RoomIndex: roomIndex,
		HP:        10, // Default HP
		MaxHP:     10,
	}
}

// NewEnemyFromDef creates a new enemy from a data-driven definition.
func NewEnemyFromDef(def *gamedata.EnemyDef, x, y, roomIndex int) *Enemy {
	return &Enemy{
		Def:       def,
		Name:      def.Name,
		Symbol:    def.GlyphRune(),
		X:         x,
		Y:         y,
		RoomIndex: roomIndex,
		HP:        def.HP,
		MaxHP:     def.HP,
	}
}

// Position returns the enemy's current x, y coordinates.
func (e *Enemy) Position() (int, int) {
	return e.X, e.Y
}

// Color returns the tcell color for this enemy.
// Uses the EnemyDef color if available, otherwise falls back to type-based colors.
func (e *Enemy) Color() tcell.Color {
	if e.Def != nil {
		return e.Def.TCellColor()
	}
	// Fallback for legacy enemies
	switch e.Type {
	case EnemyGoblin:
		return tcell.ColorGreen
	case EnemyOrc:
		return tcell.ColorRed
	case EnemySkeleton:
		return tcell.ColorWhite
	default:
		return tcell.ColorPurple
	}
}

// Attack returns the enemy's attack power.
func (e *Enemy) Attack() int {
	if e.Def != nil {
		return e.Def.Attack
	}
	return 2 // Default
}

// Defense returns the enemy's defense value.
func (e *Enemy) Defense() int {
	if e.Def != nil {
		return e.Def.Defense
	}
	return 1 // Default
}

// ID returns the enemy's unique type identifier.
func (e *Enemy) ID() string {
	if e.Def != nil {
		return e.Def.ID
	}
	return e.Type.String()
}

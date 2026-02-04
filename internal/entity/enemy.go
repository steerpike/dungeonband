// Package entity provides game entities like the party and monsters.
package entity

// EnemyType represents a type of enemy.
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
	Name      string    // Enemy name (e.g., "Goblin Scout")
	Type      EnemyType // Type of enemy
	Symbol    rune      // Display symbol
	X, Y      int       // Position in the dungeon
	RoomIndex int       // Index of the room this enemy is in (-1 if not in a room)
}

// NewEnemy creates a new enemy of the given type at the specified position.
func NewEnemy(enemyType EnemyType, x, y, roomIndex int) *Enemy {
	return &Enemy{
		Name:      enemyType.String(),
		Type:      enemyType,
		Symbol:    enemyType.Symbol(),
		X:         x,
		Y:         y,
		RoomIndex: roomIndex,
	}
}

// Position returns the enemy's current x, y coordinates.
func (e *Enemy) Position() (int, int) {
	return e.X, e.Y
}

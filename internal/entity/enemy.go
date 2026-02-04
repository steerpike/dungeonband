// Package entity provides game entities like the party and monsters.
package entity

import (
	"github.com/gdamore/tcell/v2"

	"github.com/samdwyer/dungeonband/internal/combat"
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
	MP        int                // Current mana points
	MaxMP     int                // Maximum mana points

	activeStatusEffects []combat.StatusEffect
}

// NewEnemy creates a new enemy of the given type at the specified position.
// Deprecated: Use NewEnemyFromDef instead.
func NewEnemy(enemyType EnemyType, x, y, roomIndex int) *Enemy {
	return &Enemy{
		Name:                enemyType.String(),
		Type:                enemyType,
		Symbol:              enemyType.Symbol(),
		X:                   x,
		Y:                   y,
		RoomIndex:           roomIndex,
		HP:                  10, // Default HP
		MaxHP:               10,
		MP:                  0,
		MaxMP:               0,
		activeStatusEffects: []combat.StatusEffect{},
	}
}

// NewEnemyFromDef creates a new enemy from a data-driven definition.
func NewEnemyFromDef(def *gamedata.EnemyDef, x, y, roomIndex int) *Enemy {
	return &Enemy{
		Def:                 def,
		Name:                def.Name,
		Symbol:              def.GlyphRune(),
		X:                   x,
		Y:                   y,
		RoomIndex:           roomIndex,
		HP:                  def.HP,
		MaxHP:               def.HP,
		MP:                  0, // Enemies don't use MP currently
		MaxMP:               0,
		activeStatusEffects: []combat.StatusEffect{},
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

// =============================================================================
// Combatant interface implementation
// =============================================================================

// GetName returns the enemy's name.
func (e *Enemy) GetName() string { return e.Name }

// IsAlive returns true if the enemy has HP remaining.
func (e *Enemy) IsAlive() bool { return e.HP > 0 }

// GetHP returns current HP.
func (e *Enemy) GetHP() int { return e.HP }

// GetMaxHP returns maximum HP.
func (e *Enemy) GetMaxHP() int { return e.MaxHP }

// GetMP returns current MP.
func (e *Enemy) GetMP() int { return e.MP }

// GetMaxMP returns maximum MP.
func (e *Enemy) GetMaxMP() int { return e.MaxMP }

// GetAttack returns attack stat.
func (e *Enemy) GetAttack() int { return e.Attack() }

// GetDefense returns defense stat.
func (e *Enemy) GetDefense() int { return e.Defense() }

// GetMagic returns magic stat (enemies default to 0).
func (e *Enemy) GetMagic() int { return 0 }

// TakeDamage reduces HP and returns actual damage taken.
func (e *Enemy) TakeDamage(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if actual > e.HP {
		actual = e.HP
	}
	e.HP -= actual
	return actual
}

// Heal restores HP and returns actual amount healed.
func (e *Enemy) Heal(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if e.HP+actual > e.MaxHP {
		actual = e.MaxHP - e.HP
	}
	e.HP += actual
	return actual
}

// SpendMP reduces MP and returns false if insufficient.
func (e *Enemy) SpendMP(amount int) bool {
	if e.MP < amount {
		return false
	}
	e.MP -= amount
	return true
}

// RestoreMP restores MP and returns actual amount restored.
func (e *Enemy) RestoreMP(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if e.MP+actual > e.MaxMP {
		actual = e.MaxMP - e.MP
	}
	e.MP += actual
	return actual
}

// GetAbilityIDs returns the list of ability IDs this enemy can use.
func (e *Enemy) GetAbilityIDs() []string {
	if e.Def != nil {
		return e.Def.Abilities
	}
	return []string{"attack"} // Default to basic attack
}

// GetStatusEffects returns active status effects.
func (e *Enemy) GetStatusEffects() []combat.StatusEffect {
	return e.activeStatusEffects
}

// AddStatusEffect adds or replaces a status effect.
func (e *Enemy) AddStatusEffect(effect combat.StatusEffect) {
	for i, existing := range e.activeStatusEffects {
		if existing.Type == effect.Type {
			e.activeStatusEffects[i] = effect
			return
		}
	}
	e.activeStatusEffects = append(e.activeStatusEffects, effect)
}

// RemoveStatusEffect removes a status effect by type.
func (e *Enemy) RemoveStatusEffect(effectType gamedata.StatusEffectType) {
	for i, existing := range e.activeStatusEffects {
		if existing.Type == effectType {
			e.activeStatusEffects = append(e.activeStatusEffects[:i], e.activeStatusEffects[i+1:]...)
			return
		}
	}
}

// TickStatusEffects processes turn-based status effects.
func (e *Enemy) TickStatusEffects() []combat.StatusTick {
	var ticks []combat.StatusTick
	remaining := []combat.StatusEffect{}

	for _, effect := range e.activeStatusEffects {
		tick := combat.StatusTick{Type: effect.Type}

		switch effect.Type {
		case gamedata.StatusPoison:
			tick.Amount = e.TakeDamage(effect.Power)
		case gamedata.StatusRegen:
			tick.Amount = e.Heal(effect.Power)
		}

		effect.RemainingTurns--
		if effect.RemainingTurns <= 0 {
			tick.Ended = true
		} else {
			remaining = append(remaining, effect)
		}
		ticks = append(ticks, tick)
	}

	e.activeStatusEffects = remaining
	return ticks
}

// Ensure Enemy implements combat.Combatant
var _ combat.Combatant = (*Enemy)(nil)

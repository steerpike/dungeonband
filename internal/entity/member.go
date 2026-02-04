// Package entity provides game entities like the party and monsters.
package entity

import (
	"github.com/samdwyer/dungeonband/internal/combat"
	"github.com/samdwyer/dungeonband/internal/gamedata"
)

// Class represents an adventurer's class.
type Class int

const (
	ClassWarrior Class = iota
	ClassRogue
	ClassWizard
	ClassCleric
)

// String returns the class name.
func (c Class) String() string {
	switch c {
	case ClassWarrior:
		return "Warrior"
	case ClassRogue:
		return "Rogue"
	case ClassWizard:
		return "Wizard"
	case ClassCleric:
		return "Cleric"
	default:
		return "Unknown"
	}
}

// ID returns the class identifier for data lookup.
func (c Class) ID() string {
	switch c {
	case ClassWarrior:
		return "warrior"
	case ClassRogue:
		return "rogue"
	case ClassWizard:
		return "wizard"
	case ClassCleric:
		return "cleric"
	default:
		return "unknown"
	}
}

// Symbol returns the default display symbol for a class.
func (c Class) Symbol() rune {
	switch c {
	case ClassWarrior:
		return 'W'
	case ClassRogue:
		return 'R'
	case ClassWizard:
		return 'Z'
	case ClassCleric:
		return 'C'
	default:
		return '?'
	}
}

// Member represents an individual party member.
type Member struct {
	Name   string // Character name
	Class  Class  // Character class
	Symbol rune   // Display symbol (defaults to class symbol)
	X, Y   int    // Position (absolute in combat, relative in formation)

	// Combat stats
	HP, MaxHP           int
	MP, MaxMP           int
	Attack              int
	Defense             int
	Magic               int
	AbilityIDs          []string
	activeStatusEffects []combat.StatusEffect
}

// NewMember creates a new party member with the given name and class.
// Stats are set to default values; use InitFromClassDef to load from data.
func NewMember(name string, class Class) *Member {
	return &Member{
		Name:                name,
		Class:               class,
		Symbol:              class.Symbol(),
		HP:                  20, // Default stats
		MaxHP:               20,
		MP:                  10,
		MaxMP:               10,
		Attack:              5,
		Defense:             3,
		Magic:               3,
		AbilityIDs:          []string{"attack", "defend"},
		activeStatusEffects: []combat.StatusEffect{},
	}
}

// InitFromClassDef initializes member stats from a class definition.
func (m *Member) InitFromClassDef(def *gamedata.ClassDef) {
	if def == nil {
		return
	}
	m.HP = def.HP
	m.MaxHP = def.HP
	m.MP = def.MP
	m.MaxMP = def.MP
	m.Attack = def.Attack
	m.Defense = def.Defense
	m.Magic = def.Magic
	m.AbilityIDs = make([]string, len(def.Abilities))
	copy(m.AbilityIDs, def.Abilities)
}

// SetPosition updates the member's position.
func (m *Member) SetPosition(x, y int) {
	m.X = x
	m.Y = y
}

// =============================================================================
// Combatant interface implementation
// =============================================================================

// GetName returns the member's name.
func (m *Member) GetName() string { return m.Name }

// IsAlive returns true if the member has HP remaining.
func (m *Member) IsAlive() bool { return m.HP > 0 }

// GetHP returns current HP.
func (m *Member) GetHP() int { return m.HP }

// GetMaxHP returns maximum HP.
func (m *Member) GetMaxHP() int { return m.MaxHP }

// GetMP returns current MP.
func (m *Member) GetMP() int { return m.MP }

// GetMaxMP returns maximum MP.
func (m *Member) GetMaxMP() int { return m.MaxMP }

// GetAttack returns attack stat.
func (m *Member) GetAttack() int { return m.Attack }

// GetDefense returns defense stat.
func (m *Member) GetDefense() int { return m.Defense }

// GetMagic returns magic stat.
func (m *Member) GetMagic() int { return m.Magic }

// TakeDamage reduces HP and returns actual damage taken.
func (m *Member) TakeDamage(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if actual > m.HP {
		actual = m.HP
	}
	m.HP -= actual
	return actual
}

// Heal restores HP and returns actual amount healed.
func (m *Member) Heal(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if m.HP+actual > m.MaxHP {
		actual = m.MaxHP - m.HP
	}
	m.HP += actual
	return actual
}

// SpendMP reduces MP and returns false if insufficient.
func (m *Member) SpendMP(amount int) bool {
	if m.MP < amount {
		return false
	}
	m.MP -= amount
	return true
}

// RestoreMP restores MP and returns actual amount restored.
func (m *Member) RestoreMP(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if m.MP+actual > m.MaxMP {
		actual = m.MaxMP - m.MP
	}
	m.MP += actual
	return actual
}

// GetAbilityIDs returns the list of ability IDs this member can use.
func (m *Member) GetAbilityIDs() []string {
	return m.AbilityIDs
}

// GetStatusEffects returns active status effects.
func (m *Member) GetStatusEffects() []combat.StatusEffect {
	return m.activeStatusEffects
}

// AddStatusEffect adds or replaces a status effect.
func (m *Member) AddStatusEffect(effect combat.StatusEffect) {
	// Replace existing effect of same type
	for i, existing := range m.activeStatusEffects {
		if existing.Type == effect.Type {
			m.activeStatusEffects[i] = effect
			return
		}
	}
	m.activeStatusEffects = append(m.activeStatusEffects, effect)
}

// RemoveStatusEffect removes a status effect by type.
func (m *Member) RemoveStatusEffect(effectType gamedata.StatusEffectType) {
	for i, existing := range m.activeStatusEffects {
		if existing.Type == effectType {
			m.activeStatusEffects = append(m.activeStatusEffects[:i], m.activeStatusEffects[i+1:]...)
			return
		}
	}
}

// TickStatusEffects processes turn-based status effects.
func (m *Member) TickStatusEffects() []combat.StatusTick {
	var ticks []combat.StatusTick
	remaining := []combat.StatusEffect{}

	for _, effect := range m.activeStatusEffects {
		tick := combat.StatusTick{Type: effect.Type}

		switch effect.Type {
		case gamedata.StatusPoison:
			tick.Amount = m.TakeDamage(effect.Power)
		case gamedata.StatusRegen:
			tick.Amount = m.Heal(effect.Power)
		}

		effect.RemainingTurns--
		if effect.RemainingTurns <= 0 {
			tick.Ended = true
		} else {
			remaining = append(remaining, effect)
		}
		ticks = append(ticks, tick)
	}

	m.activeStatusEffects = remaining
	return ticks
}

// Ensure Member implements combat.Combatant
var _ combat.Combatant = (*Member)(nil)

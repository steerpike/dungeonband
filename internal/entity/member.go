// Package entity provides game entities like the party and monsters.
package entity

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
}

// NewMember creates a new party member with the given name and class.
func NewMember(name string, class Class) *Member {
	return &Member{
		Name:   name,
		Class:  class,
		Symbol: class.Symbol(),
	}
}

// SetPosition updates the member's position.
func (m *Member) SetPosition(x, y int) {
	m.X = x
	m.Y = y
}

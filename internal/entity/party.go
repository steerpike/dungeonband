// Package entity provides game entities like the party and monsters.
package entity

import "github.com/samdwyer/dungeonband/internal/gamedata"

// Party represents the player's party of adventurers.
// In explore mode, the party is displayed as a single symbol.
// In combat mode, individual members are displayed.
type Party struct {
	X, Y    int       // Current position in the dungeon (party center)
	Symbol  rune      // Display symbol ('&' in explore mode)
	Members []*Member // Individual party members
}

// NewParty creates a new party at the given position with default members.
func NewParty(x, y int) *Party {
	return &Party{
		X:      x,
		Y:      y,
		Symbol: '&',
		Members: []*Member{
			NewMember("Aldric", ClassWarrior),
			NewMember("Shade", ClassRogue),
			NewMember("Zephyr", ClassWizard),
			NewMember("Celeste", ClassCleric),
		},
	}
}

// NewPartyWithClassData creates a new party with members initialized from class definitions.
func NewPartyWithClassData(x, y int, classRegistry *gamedata.ClassRegistry) *Party {
	party := NewParty(x, y)

	// Initialize each member with their class data
	for _, member := range party.Members {
		classDef := classRegistry.GetByID(member.Class.ID())
		if classDef != nil {
			member.InitFromClassDef(classDef)
		}
	}

	return party
}

// Move updates the party position by the given delta.
func (p *Party) Move(dx, dy int) {
	p.X += dx
	p.Y += dy
}

// Position returns the current x, y coordinates.
func (p *Party) Position() (int, int) {
	return p.X, p.Y
}

// AliveMemberCount returns the number of members with HP > 0.
func (p *Party) AliveMemberCount() int {
	count := 0
	for _, m := range p.Members {
		if m.IsAlive() {
			count++
		}
	}
	return count
}

// GetAliveMember returns the nth alive member (0-indexed), or nil.
func (p *Party) GetAliveMember(index int) *Member {
	current := 0
	for _, m := range p.Members {
		if m.IsAlive() {
			if current == index {
				return m
			}
			current++
		}
	}
	return nil
}

// IsDefeated returns true if all party members are dead.
func (p *Party) IsDefeated() bool {
	return p.AliveMemberCount() == 0
}

// Package entity provides game entities like the party and monsters.
package entity

// Party represents the player's party of adventurers.
// In explore mode, the party is displayed as a single symbol.
type Party struct {
	X, Y   int  // Current position in the dungeon
	Symbol rune // Display symbol ('&' in explore mode)
}

// NewParty creates a new party at the given position.
func NewParty(x, y int) *Party {
	return &Party{
		X:      x,
		Y:      y,
		Symbol: '&',
	}
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

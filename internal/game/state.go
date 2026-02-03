// Package game provides the main game loop and state management.
package game

// State represents the current game state.
type State int

const (
	// StateExplore is the default exploration mode where party moves as one unit.
	StateExplore State = iota
	// StateCombat is the tactical combat mode where party members act individually.
	StateCombat
)

// String returns a human-readable state name.
func (s State) String() string {
	switch s {
	case StateExplore:
		return "explore"
	case StateCombat:
		return "combat"
	default:
		return "unknown"
	}
}

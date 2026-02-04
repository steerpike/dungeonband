package game

import (
	"testing"

	"github.com/samdwyer/dungeonband/internal/entity"
	"github.com/samdwyer/dungeonband/internal/gamedata"
)

func TestCombatPhaseString(t *testing.T) {
	tests := []struct {
		phase    CombatPhase
		expected string
	}{
		{PhasePlayerTurn, "player_turn"},
		{PhaseEnemyTurn, "enemy_turn"},
		{PhaseVictory, "victory"},
		{PhaseDefeat, "defeat"},
		{CombatPhase(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.phase.String()
		if got != tt.expected {
			t.Errorf("CombatPhase(%d).String() = %q, want %q", tt.phase, got, tt.expected)
		}
	}
}

func TestNewCombatState(t *testing.T) {
	enemies := []*entity.Enemy{
		entity.NewEnemy(entity.EnemyGoblin, 5, 5, 1),
		entity.NewEnemy(entity.EnemyOrc, 6, 5, 1),
	}

	cs := NewCombatState(enemies)

	if cs.Phase != PhasePlayerTurn {
		t.Errorf("NewCombatState().Phase = %v, want PhasePlayerTurn", cs.Phase)
	}
	if len(cs.Enemies) != 2 {
		t.Errorf("NewCombatState().Enemies length = %d, want 2", len(cs.Enemies))
	}
	if cs.ActiveMemberIndex != 0 {
		t.Errorf("NewCombatState().ActiveMemberIndex = %d, want 0", cs.ActiveMemberIndex)
	}
	if cs.TurnCount != 0 {
		t.Errorf("NewCombatState().TurnCount = %d, want 0", cs.TurnCount)
	}
	if cs.LastMessage != "Combat begins!" {
		t.Errorf("NewCombatState().LastMessage = %q, want %q", cs.LastMessage, "Combat begins!")
	}
}

func TestCombatStateAliveEnemyCount(t *testing.T) {
	enemies := []*entity.Enemy{
		entity.NewEnemy(entity.EnemyGoblin, 5, 5, 1),
		entity.NewEnemy(entity.EnemyOrc, 6, 5, 1),
	}

	cs := NewCombatState(enemies)

	// Initially all alive
	if got := cs.AliveEnemyCount(); got != 2 {
		t.Errorf("AliveEnemyCount() = %d, want 2", got)
	}

	// Kill one enemy
	enemies[0].TakeDamage(1000)

	if got := cs.AliveEnemyCount(); got != 1 {
		t.Errorf("AliveEnemyCount() after kill = %d, want 1", got)
	}

	// Kill all
	enemies[1].TakeDamage(1000)

	if got := cs.AliveEnemyCount(); got != 0 {
		t.Errorf("AliveEnemyCount() all dead = %d, want 0", got)
	}
}

func TestCombatStateGetFirstAliveEnemy(t *testing.T) {
	enemies := []*entity.Enemy{
		entity.NewEnemy(entity.EnemyGoblin, 5, 5, 1),
		entity.NewEnemy(entity.EnemyOrc, 6, 5, 1),
	}

	cs := NewCombatState(enemies)

	// First alive should be first enemy
	first := cs.GetFirstAliveEnemy()
	if first != enemies[0] {
		t.Error("GetFirstAliveEnemy() should return first enemy")
	}

	// Kill first enemy
	enemies[0].TakeDamage(1000)

	// Now first alive should be second enemy
	first = cs.GetFirstAliveEnemy()
	if first != enemies[1] {
		t.Error("GetFirstAliveEnemy() should return second enemy after first dies")
	}

	// Kill all
	enemies[1].TakeDamage(1000)

	first = cs.GetFirstAliveEnemy()
	if first != nil {
		t.Error("GetFirstAliveEnemy() should return nil when all dead")
	}
}

func TestCombatStateGetAliveEnemy(t *testing.T) {
	// Create enemies using data-driven approach
	goblinDef := &gamedata.EnemyDef{
		ID:        "goblin",
		Name:      "Goblin",
		HP:        15,
		Attack:    5,
		Defense:   2,
		Abilities: []string{"attack"},
	}
	orcDef := &gamedata.EnemyDef{
		ID:        "orc",
		Name:      "Orc",
		HP:        25,
		Attack:    8,
		Defense:   4,
		Abilities: []string{"attack"},
	}

	enemies := []*entity.Enemy{
		entity.NewEnemyFromDef(goblinDef, 5, 5, 1),
		entity.NewEnemyFromDef(orcDef, 6, 5, 1),
	}

	cs := NewCombatState(enemies)

	// Get by index
	if e := cs.GetAliveEnemy(0); e != enemies[0] {
		t.Error("GetAliveEnemy(0) should return first enemy")
	}
	if e := cs.GetAliveEnemy(1); e != enemies[1] {
		t.Error("GetAliveEnemy(1) should return second enemy")
	}
	if e := cs.GetAliveEnemy(2); e != nil {
		t.Error("GetAliveEnemy(2) should return nil (out of bounds)")
	}

	// Kill first enemy
	enemies[0].TakeDamage(1000)

	// Index 0 should now be second enemy
	if e := cs.GetAliveEnemy(0); e != enemies[1] {
		t.Error("GetAliveEnemy(0) should return second enemy after first dies")
	}
	if e := cs.GetAliveEnemy(1); e != nil {
		t.Error("GetAliveEnemy(1) should return nil after first dies")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{-5, "-5"},
		{-123, "-123"},
	}

	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.expected {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

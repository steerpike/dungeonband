package game

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/samdwyer/dungeonband/internal/combat"
	"github.com/samdwyer/dungeonband/internal/entity"
	"github.com/samdwyer/dungeonband/internal/gamedata"
	"github.com/samdwyer/dungeonband/internal/telemetry"
)

// CombatPhase represents the current phase of combat.
type CombatPhase int

const (
	// PhasePlayerTurn - waiting for player to select an ability
	PhasePlayerTurn CombatPhase = iota
	// PhaseEnemyTurn - enemies are taking their turns
	PhaseEnemyTurn
	// PhaseVictory - all enemies defeated
	PhaseVictory
	// PhaseDefeat - all party members defeated
	PhaseDefeat
)

// String returns a human-readable phase name.
func (p CombatPhase) String() string {
	switch p {
	case PhasePlayerTurn:
		return "player_turn"
	case PhaseEnemyTurn:
		return "enemy_turn"
	case PhaseVictory:
		return "victory"
	case PhaseDefeat:
		return "defeat"
	default:
		return "unknown"
	}
}

// CombatState holds all state for an active combat encounter.
type CombatState struct {
	Phase             CombatPhase
	Enemies           []*entity.Enemy
	ActiveMemberIndex int                  // Which party member is acting (0-3)
	ActiveEnemyIndex  int                  // Which enemy is acting
	TurnCount         int                  // Total turns taken
	LastMessage       string               // Message to display from last action
	SelectedAbility   *gamedata.AbilityDef // Ability selected by current actor
}

// NewCombatState creates a new combat state for an encounter.
func NewCombatState(enemies []*entity.Enemy) *CombatState {
	return &CombatState{
		Phase:             PhasePlayerTurn,
		Enemies:           enemies,
		ActiveMemberIndex: 0,
		ActiveEnemyIndex:  0,
		TurnCount:         0,
		LastMessage:       "Combat begins!",
	}
}

// AliveEnemyCount returns the number of enemies still alive.
func (cs *CombatState) AliveEnemyCount() int {
	count := 0
	for _, e := range cs.Enemies {
		if e.IsAlive() {
			count++
		}
	}
	return count
}

// GetFirstAliveEnemy returns the first alive enemy, or nil.
func (cs *CombatState) GetFirstAliveEnemy() *entity.Enemy {
	for _, e := range cs.Enemies {
		if e.IsAlive() {
			return e
		}
	}
	return nil
}

// GetAliveEnemy returns the nth alive enemy (0-indexed), or nil.
func (cs *CombatState) GetAliveEnemy(index int) *entity.Enemy {
	current := 0
	for _, e := range cs.Enemies {
		if e.IsAlive() {
			if current == index {
				return e
			}
			current++
		}
	}
	return nil
}

// =============================================================================
// Combat Loop Methods on Game
// =============================================================================

// initCombatState initializes combat state when entering combat.
func (g *Game) initCombatState(ctx context.Context) {
	tracer := telemetry.Tracer("combat")
	_, span := tracer.Start(ctx, "combat.start")
	span.SetAttributes(
		attribute.Int("party_size", g.party.AliveMemberCount()),
		attribute.Int("enemy_count", len(g.combatEnemies)),
	)
	span.End()

	g.combatState = NewCombatState(g.combatEnemies)

	// Find first alive member
	g.combatState.ActiveMemberIndex = 0
	for i, m := range g.party.Members {
		if m.IsAlive() {
			g.combatState.ActiveMemberIndex = i
			break
		}
	}
}

// executeCombatTurn executes the current actor's turn with the selected ability.
func (g *Game) executeCombatTurn(ctx context.Context, ability *gamedata.AbilityDef, user combat.Combatant, target combat.Combatant) {
	if g.effectResolver == nil || ability == nil {
		return
	}

	tracer := telemetry.Tracer("combat")
	ctx, span := tracer.Start(ctx, "combat.turn")
	span.SetAttributes(
		attribute.String("actor", user.GetName()),
		attribute.String("ability", ability.ID),
		attribute.String("target", target.GetName()),
		attribute.Int("turn", g.combatState.TurnCount),
	)
	defer span.End()

	// Resolve the ability
	result := g.effectResolver.Resolve(ability, user, target)

	// Build message
	if result.Success {
		if result.Damage > 0 {
			g.combatState.LastMessage = result.Message + " " +
				target.GetName() + " takes " + itoa(result.Damage) + " damage!"
			span.SetAttributes(attribute.Int("damage", result.Damage))
		} else if result.Healing > 0 {
			g.combatState.LastMessage = result.Message + " " +
				target.GetName() + " heals " + itoa(result.Healing) + " HP!"
			span.SetAttributes(attribute.Int("healing", result.Healing))
		} else {
			g.combatState.LastMessage = result.Message
		}
		if result.StatusAdded != "" {
			span.SetAttributes(attribute.String("status_applied", string(result.StatusAdded)))
		}
	} else {
		g.combatState.LastMessage = result.Message
		span.SetAttributes(attribute.Bool("failed", true))
	}

	g.combatState.TurnCount++
}

// advanceToNextPartyMember moves to the next alive party member, or to enemy phase.
func (g *Game) advanceToNextPartyMember() {
	// Find next alive member after current
	for i := g.combatState.ActiveMemberIndex + 1; i < len(g.party.Members); i++ {
		if g.party.Members[i].IsAlive() {
			g.combatState.ActiveMemberIndex = i
			return
		}
	}

	// No more party members, switch to enemy turn
	g.combatState.Phase = PhaseEnemyTurn
	g.combatState.ActiveEnemyIndex = 0
}

// executeEnemyTurns executes all enemy turns in sequence.
func (g *Game) executeEnemyTurns(ctx context.Context) {
	for _, enemy := range g.combatState.Enemies {
		if !enemy.IsAlive() {
			continue
		}

		// Simple AI: pick a random ability and random alive party member
		ability := g.selectEnemyAbility(enemy)
		target := g.selectEnemyTarget(enemy, ability)

		if ability != nil && target != nil {
			g.executeCombatTurn(ctx, ability, enemy, target)
		}

		// Check for party defeat after each enemy turn
		if g.party.IsDefeated() {
			g.combatState.Phase = PhaseDefeat
			g.combatState.LastMessage = "Your party has been defeated!"
			return
		}
	}

	// All enemies done, check victory or start new round
	if g.combatState.AliveEnemyCount() == 0 {
		g.combatState.Phase = PhaseVictory
		g.combatState.LastMessage = "Victory! All enemies defeated!"
	} else {
		// Start new round with first alive party member
		g.combatState.Phase = PhasePlayerTurn
		for i, m := range g.party.Members {
			if m.IsAlive() {
				g.combatState.ActiveMemberIndex = i
				break
			}
		}
	}
}

// selectEnemyAbility picks an ability for an enemy to use.
func (g *Game) selectEnemyAbility(enemy *entity.Enemy) *gamedata.AbilityDef {
	if g.abilityRegistry == nil {
		return nil
	}

	abilityIDs := enemy.GetAbilityIDs()
	if len(abilityIDs) == 0 {
		return nil
	}

	// Simple AI: pick a random ability that the enemy can use
	// Shuffle and find first usable
	for _, idx := range g.rng.Perm(len(abilityIDs)) {
		ability := g.abilityRegistry.GetByID(abilityIDs[idx])
		if ability != nil && enemy.GetMP() >= ability.MPCost {
			return ability
		}
	}

	// Fallback to first ability (usually "attack" which has 0 MP cost)
	return g.abilityRegistry.GetByID(abilityIDs[0])
}

// selectEnemyTarget picks a target for an enemy ability.
func (g *Game) selectEnemyTarget(enemy *entity.Enemy, ability *gamedata.AbilityDef) combat.Combatant {
	if ability == nil {
		return nil
	}

	switch ability.TargetType {
	case gamedata.TargetSelf:
		return enemy
	case gamedata.TargetSingleEnemy, gamedata.TargetAllEnemies:
		// For enemies, "enemy" means party members
		// Pick random alive party member, preferring lowest HP
		return g.selectLowestHPPartyMember()
	case gamedata.TargetSingleAlly, gamedata.TargetAllAllies:
		// For enemies, "ally" means other enemies
		// Pick lowest HP ally (for healing)
		return g.selectLowestHPEnemy()
	default:
		return g.selectLowestHPPartyMember()
	}
}

// selectLowestHPPartyMember returns the alive party member with lowest HP.
func (g *Game) selectLowestHPPartyMember() *entity.Member {
	var lowest *entity.Member
	for _, m := range g.party.Members {
		if m.IsAlive() {
			if lowest == nil || m.GetHP() < lowest.GetHP() {
				lowest = m
			}
		}
	}
	return lowest
}

// selectLowestHPEnemy returns the alive enemy with lowest HP.
func (g *Game) selectLowestHPEnemy() *entity.Enemy {
	var lowest *entity.Enemy
	for _, e := range g.combatState.Enemies {
		if e.IsAlive() {
			if lowest == nil || e.GetHP() < lowest.GetHP() {
				lowest = e
			}
		}
	}
	return lowest
}

// checkCombatEnd checks if combat should end and updates phase accordingly.
func (g *Game) checkCombatEnd() bool {
	if g.party.IsDefeated() {
		g.combatState.Phase = PhaseDefeat
		g.combatState.LastMessage = "Your party has been defeated!"
		return true
	}
	if g.combatState.AliveEnemyCount() == 0 {
		g.combatState.Phase = PhaseVictory
		g.combatState.LastMessage = "Victory! All enemies defeated!"
		return true
	}
	return false
}

// endCombat handles combat ending (victory or defeat).
func (g *Game) endCombat(ctx context.Context, outcome string) {
	tracer := telemetry.Tracer("combat")
	_, span := tracer.Start(ctx, "combat.end")
	span.SetAttributes(
		attribute.String("outcome", outcome),
		attribute.Int("turns_taken", g.combatState.TurnCount),
		attribute.Int("party_hp_remaining", g.totalPartyHP()),
	)
	span.End()

	// Remove dead enemies from the dungeon
	if outcome == "victory" {
		g.removeDeadEnemies()
	}
}

// totalPartyHP returns the sum of all party members' current HP.
func (g *Game) totalPartyHP() int {
	total := 0
	for _, m := range g.party.Members {
		total += m.GetHP()
	}
	return total
}

// removeDeadEnemies removes defeated enemies from the game.
func (g *Game) removeDeadEnemies() {
	alive := make([]*entity.Enemy, 0, len(g.enemies))
	for _, e := range g.enemies {
		if e.IsAlive() {
			alive = append(alive, e)
		}
	}
	g.enemies = alive
}

// itoa is a simple int to string helper.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	digits := ""
	for i > 0 {
		digits = string(rune('0'+i%10)) + digits
		i /= 10
	}
	return digits
}

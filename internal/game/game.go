package game

import (
	"context"
	"log"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"go.opentelemetry.io/otel/attribute"

	"github.com/samdwyer/dungeonband/internal/combat"
	"github.com/samdwyer/dungeonband/internal/entity"
	"github.com/samdwyer/dungeonband/internal/gamedata"
	"github.com/samdwyer/dungeonband/internal/telemetry"
	"github.com/samdwyer/dungeonband/internal/ui"
	"github.com/samdwyer/dungeonband/internal/world"
)

// Game holds the entire game state.
type Game struct {
	screen          *ui.Screen
	renderer        *ui.Renderer
	dungeon         *world.Dungeon
	party           *entity.Party
	enemies         []*entity.Enemy
	enemyRegistry   *gamedata.EnemyRegistry
	classRegistry   *gamedata.ClassRegistry
	abilityRegistry *gamedata.AbilityRegistry
	effectResolver  *combat.EffectResolver
	state           State
	running         bool
	rng             *rand.Rand
	seed            int64

	// Combat state
	combatEnemies     []*entity.Enemy // Enemies in the current combat encounter
	activeMemberIndex int             // Index of the party member whose turn it is
	combatState       *CombatState    // Full combat state for turn-based combat
}

// New creates a new game instance with the given configuration.
func New(cfg Config) (*Game, error) {
	screen, err := ui.NewScreen()
	if err != nil {
		return nil, err
	}

	// Load enemy registry from embedded data
	enemyRegistry, err := gamedata.LoadEnemyRegistry()
	if err != nil {
		log.Printf("Warning: failed to load enemy registry: %v (using legacy spawning)", err)
	}

	// Load class registry
	classRegistry, err := gamedata.LoadClassRegistry()
	if err != nil {
		log.Printf("Warning: failed to load class registry: %v (using default stats)", err)
	}

	// Load ability registry
	abilityRegistry, err := gamedata.LoadAbilityRegistry()
	if err != nil {
		log.Printf("Warning: failed to load ability registry: %v", err)
	}

	var effectResolver *combat.EffectResolver
	if abilityRegistry != nil {
		effectResolver = combat.NewEffectResolver(abilityRegistry)
	}

	return &Game{
		screen:          screen,
		renderer:        ui.NewRenderer(screen),
		enemyRegistry:   enemyRegistry,
		classRegistry:   classRegistry,
		abilityRegistry: abilityRegistry,
		effectResolver:  effectResolver,
		state:           StateExplore,
		running:         true,
		rng:             rand.New(rand.NewSource(cfg.Seed)),
		seed:            cfg.Seed,
	}, nil
}

// Run executes the main game loop.
func (g *Game) Run(ctx context.Context) error {
	tracer := telemetry.Tracer("game")

	// Initialize game (traced)
	ctx, initSpan := tracer.Start(ctx, "game.init")

	// Generate dungeon with the game's RNG for reproducibility
	g.dungeon = world.NewDungeon(world.DefaultWidth, world.DefaultHeight, g.rng)
	g.dungeon.Generate(ctx)

	// Place party in first room's center
	if len(g.dungeon.Rooms) > 0 {
		startX, startY := g.dungeon.Rooms[0].Center()

		// Create party with class data if available
		if g.classRegistry != nil {
			g.party = entity.NewPartyWithClassData(startX, startY, g.classRegistry)
		} else {
			g.party = entity.NewParty(startX, startY)
		}

		// Spawn enemies in rooms (skip room 0 - starting room)
		g.spawnEnemies()

		initSpan.SetAttributes(
			attribute.Int("dungeon.rooms", len(g.dungeon.Rooms)),
			attribute.Int("party.start_x", startX),
			attribute.Int("party.start_y", startY),
			attribute.Int("enemy_count", len(g.enemies)),
			attribute.Int64("seed", g.seed),
		)
	} else {
		// Fallback: place in center of map
		if g.classRegistry != nil {
			g.party = entity.NewPartyWithClassData(g.dungeon.Width/2, g.dungeon.Height/2, g.classRegistry)
		} else {
			g.party = entity.NewParty(g.dungeon.Width/2, g.dungeon.Height/2)
		}
		initSpan.SetAttributes(
			attribute.Int("dungeon.rooms", 0),
			attribute.String("warning", "no rooms generated, using fallback position"),
			attribute.Int("enemy_count", 0),
			attribute.Int64("seed", g.seed),
		)
	}

	initSpan.End()

	// Main game loop
	for g.running {
		// Render current state
		if g.state == StateCombat {
			combatInfo := g.buildCombatInfo()
			g.renderer.RenderWithCombat(g.dungeon, g.party, g.enemies, ui.GameState(g.state), g.seed, combatInfo)
		} else {
			g.renderer.Render(g.dungeon, g.party, g.enemies, ui.GameState(g.state), g.seed)
		}

		// Handle input (blocking)
		g.handleInput(ctx)
	}

	// Cleanup
	g.screen.Close()
	return nil
}

// handleInput processes a single input event.
func (g *Game) handleInput(ctx context.Context) {
	ev := g.screen.PollEvent()

	switch ev := ev.(type) {
	case *tcell.EventKey:
		g.handleKeyEvent(ctx, ev)
	case *tcell.EventResize:
		g.screen.Sync()
	}
}

// handleKeyEvent processes keyboard input.
func (g *Game) handleKeyEvent(ctx context.Context, ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		if g.state == StateCombat {
			// If victory or defeat, always allow exit
			if g.combatState != nil && (g.combatState.Phase == PhaseVictory || g.combatState.Phase == PhaseDefeat) {
				g.handleCombatEnd(ctx)
			} else {
				// Exit combat mode (flee)
				g.transitionState(ctx, StateExplore, "manual")
			}
		} else {
			// Quit game from explore mode
			g.running = false
		}

	case tcell.KeyCtrlC:
		g.running = false

	case tcell.KeyUp:
		if g.state == StateExplore {
			g.tryMove(ctx, 0, -1)
		}
	case tcell.KeyDown:
		if g.state == StateExplore {
			g.tryMove(ctx, 0, 1)
		}
	case tcell.KeyLeft:
		if g.state == StateExplore {
			g.tryMove(ctx, -1, 0)
		}
	case tcell.KeyRight:
		if g.state == StateExplore {
			g.tryMove(ctx, 1, 0)
		}

	case tcell.KeyRune:
		r := ev.Rune()

		// In combat victory/defeat, any key continues
		if g.state == StateCombat && g.combatState != nil {
			if g.combatState.Phase == PhaseVictory || g.combatState.Phase == PhaseDefeat {
				g.handleCombatEnd(ctx)
				return
			}
		}

		// Handle number keys for ability selection in combat
		if g.state == StateCombat && r >= '1' && r <= '9' {
			g.handleCombatAbilitySelection(ctx, int(r-'1'))
			return
		}

		switch r {
		case 'q', 'Q':
			g.running = false
		case 'c', 'C':
			if g.state == StateExplore {
				g.transitionState(ctx, StateCombat, "manual")
			}
		case 'h':
			if g.state == StateExplore {
				g.tryMove(ctx, -1, 0)
			}
		case 'j':
			if g.state == StateExplore {
				g.tryMove(ctx, 0, 1)
			}
		case 'k':
			if g.state == StateExplore {
				g.tryMove(ctx, 0, -1)
			}
		case 'l':
			if g.state == StateExplore {
				g.tryMove(ctx, 1, 0)
			}
		}
	}
}

// handleCombatAbilitySelection handles when player presses a number key in combat.
func (g *Game) handleCombatAbilitySelection(ctx context.Context, abilityIndex int) {
	// Only handle input during player turn
	if g.combatState == nil || g.combatState.Phase != PhasePlayerTurn {
		return
	}

	activeMember := g.getActiveMember()
	if activeMember == nil || g.abilityRegistry == nil {
		return
	}

	abilityIDs := activeMember.GetAbilityIDs()
	if abilityIndex >= len(abilityIDs) {
		return // Invalid selection
	}

	ability := g.abilityRegistry.GetByID(abilityIDs[abilityIndex])
	if ability == nil {
		return
	}

	// Check if can use (enough MP)
	if activeMember.GetMP() < ability.MPCost {
		g.combatState.LastMessage = "Not enough MP!"
		return
	}

	// Select target based on ability type
	var target combat.Combatant
	if ability.IsOffensive() {
		// Target first alive enemy
		target = g.combatState.GetFirstAliveEnemy()
	} else {
		// Target self for defensive/healing abilities
		target = activeMember
	}

	if target == nil {
		return
	}

	// Execute the turn
	g.executeCombatTurn(ctx, ability, activeMember, target)

	// Check for combat end (victory)
	if g.checkCombatEnd() {
		return
	}

	// Advance to next party member or enemy phase
	g.advanceToNextPartyMember()

	// If it's now enemy phase, execute all enemy turns
	if g.combatState.Phase == PhaseEnemyTurn {
		g.executeEnemyTurns(ctx)
	}
}

// tryMove attempts to move the party by the given delta.
func (g *Game) tryMove(ctx context.Context, dx, dy int) {
	newX := g.party.X + dx
	newY := g.party.Y + dy

	if g.dungeon.IsPassable(newX, newY) {
		g.party.Move(dx, dy)
	}
}

// Close cleans up game resources.
func (g *Game) Close() {
	if g.screen != nil {
		g.screen.Close()
	}
}

// transitionState changes the game state and records telemetry.
func (g *Game) transitionState(ctx context.Context, newState State, trigger string) {
	if g.state == newState {
		return // No change
	}

	tracer := telemetry.Tracer("game")
	_, span := tracer.Start(ctx, "game.state_change")
	span.SetAttributes(
		attribute.String("from_state", g.state.String()),
		attribute.String("to_state", newState.String()),
		attribute.String("trigger", trigger),
	)
	span.End()

	// Handle state-specific setup
	if newState == StateCombat {
		g.enterCombat(ctx)
	} else if g.state == StateCombat {
		g.exitCombat()
	}

	g.state = newState
}

// enterCombat sets up combat state.
func (g *Game) enterCombat(ctx context.Context) {
	// Find enemies in the same room as the party
	partyRoomIndex := g.dungeon.RoomIndexAt(g.party.X, g.party.Y)
	g.combatEnemies = nil
	for _, enemy := range g.enemies {
		if enemy.RoomIndex == partyRoomIndex && enemy.IsAlive() {
			g.combatEnemies = append(g.combatEnemies, enemy)
		}
	}
	g.activeMemberIndex = 0

	// Initialize full combat state with telemetry
	g.initCombatState(ctx)
}

// exitCombat cleans up combat state.
func (g *Game) exitCombat() {
	g.combatEnemies = nil
	g.activeMemberIndex = 0
}

// getActiveMember returns the current active party member in combat.
func (g *Game) getActiveMember() *entity.Member {
	return g.party.GetAliveMember(g.activeMemberIndex)
}

// buildCombatInfo creates the combat UI information for rendering.
func (g *Game) buildCombatInfo() *ui.CombatInfo {
	if g.combatState == nil {
		return nil
	}

	// Use combatState's active member index for consistency
	activeMember := g.party.GetAliveMember(g.combatState.ActiveMemberIndex)
	if activeMember == nil {
		return nil
	}

	// Build ability info list
	var abilities []ui.AbilityInfo
	if g.abilityRegistry != nil {
		for _, abilityID := range activeMember.GetAbilityIDs() {
			abilityDef := g.abilityRegistry.GetByID(abilityID)
			if abilityDef != nil {
				canUse := activeMember.GetMP() >= abilityDef.MPCost
				abilities = append(abilities, ui.AbilityInfo{
					Name:   abilityDef.Name,
					MPCost: abilityDef.MPCost,
					CanUse: canUse,
				})
			}
		}
	}

	return &ui.CombatInfo{
		ActiveMember: activeMember,
		Abilities:    abilities,
		Enemies:      g.combatState.Enemies,
		Message:      g.combatState.LastMessage,
	}
}

// spawnEnemies populates the dungeon with enemies.
// Spawns 1-3 enemies per room, skipping room 0 (starting room).
// Uses the enemy registry for weighted spawning if available.
func (g *Game) spawnEnemies() {
	for roomIndex := 1; roomIndex < len(g.dungeon.Rooms); roomIndex++ {
		// 1-3 enemies per room
		count := 1 + g.rng.Intn(3)

		for i := 0; i < count; i++ {
			// Find a random position in the room
			x, y := g.dungeon.RandomPointInRoom(roomIndex)
			if x >= 0 && y >= 0 {
				var enemy *entity.Enemy

				// Use registry if available, otherwise fall back to legacy spawning
				if g.enemyRegistry != nil {
					def := g.enemyRegistry.SpawnRandom(g.rng)
					if def != nil {
						enemy = entity.NewEnemyFromDef(def, x, y, roomIndex)
					}
				}

				// Fallback to legacy spawning if registry not available or failed
				if enemy == nil {
					enemyTypes := []entity.EnemyType{
						entity.EnemyGoblin,
						entity.EnemyOrc,
						entity.EnemySkeleton,
					}
					enemyType := enemyTypes[g.rng.Intn(len(enemyTypes))]
					enemy = entity.NewEnemy(enemyType, x, y, roomIndex)
				}

				g.enemies = append(g.enemies, enemy)
			}
		}
	}
}

// handleCombatEnd processes the end of combat (victory or defeat).
func (g *Game) handleCombatEnd(ctx context.Context) {
	if g.combatState == nil {
		return
	}

	if g.combatState.Phase == PhaseVictory {
		g.endCombat(ctx, "victory")
		g.transitionState(ctx, StateExplore, "victory")
	} else if g.combatState.Phase == PhaseDefeat {
		g.endCombat(ctx, "defeat")
		// For now, just return to explore - could add game over screen later
		g.transitionState(ctx, StateExplore, "defeat")
	}
}

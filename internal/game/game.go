package game

import (
	"context"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"go.opentelemetry.io/otel/attribute"

	"github.com/samdwyer/dungeonband/internal/entity"
	"github.com/samdwyer/dungeonband/internal/telemetry"
	"github.com/samdwyer/dungeonband/internal/ui"
	"github.com/samdwyer/dungeonband/internal/world"
)

// Game holds the entire game state.
type Game struct {
	screen   *ui.Screen
	renderer *ui.Renderer
	dungeon  *world.Dungeon
	party    *entity.Party
	enemies  []*entity.Enemy
	state    State
	running  bool
	rng      *rand.Rand
}

// New creates a new game instance.
func New() (*Game, error) {
	screen, err := ui.NewScreen()
	if err != nil {
		return nil, err
	}

	return &Game{
		screen:   screen,
		renderer: ui.NewRenderer(screen),
		state:    StateExplore,
		running:  true,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Run executes the main game loop.
func (g *Game) Run(ctx context.Context) error {
	tracer := telemetry.Tracer("game")

	// Initialize game (traced)
	ctx, initSpan := tracer.Start(ctx, "game.init")

	// Generate dungeon
	g.dungeon = world.NewDungeon(world.DefaultWidth, world.DefaultHeight)
	g.dungeon.Generate(ctx)

	// Place party in first room's center
	if len(g.dungeon.Rooms) > 0 {
		startX, startY := g.dungeon.Rooms[0].Center()
		g.party = entity.NewParty(startX, startY)

		// Spawn enemies in rooms (skip room 0 - starting room)
		g.spawnEnemies()

		initSpan.SetAttributes(
			attribute.Int("dungeon.rooms", len(g.dungeon.Rooms)),
			attribute.Int("party.start_x", startX),
			attribute.Int("party.start_y", startY),
			attribute.Int("enemy_count", len(g.enemies)),
		)
	} else {
		// Fallback: place in center of map
		g.party = entity.NewParty(g.dungeon.Width/2, g.dungeon.Height/2)
		initSpan.SetAttributes(
			attribute.Int("dungeon.rooms", 0),
			attribute.String("warning", "no rooms generated, using fallback position"),
			attribute.Int("enemy_count", 0),
		)
	}

	initSpan.End()

	// Main game loop
	for g.running {
		// Render current state
		g.renderer.Render(g.dungeon, g.party, g.enemies, ui.GameState(g.state))

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
			// Exit combat mode
			g.transitionState(ctx, StateExplore, "manual")
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
		switch ev.Rune() {
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

	g.state = newState
}

// spawnEnemies populates the dungeon with enemies.
// Spawns 1-3 enemies per room, skipping room 0 (starting room).
func (g *Game) spawnEnemies() {
	enemyTypes := []entity.EnemyType{
		entity.EnemyGoblin,
		entity.EnemyOrc,
		entity.EnemySkeleton,
	}

	for roomIndex := 1; roomIndex < len(g.dungeon.Rooms); roomIndex++ {
		// 1-3 enemies per room
		count := 1 + g.rng.Intn(3)

		for i := 0; i < count; i++ {
			// Pick random enemy type
			enemyType := enemyTypes[g.rng.Intn(len(enemyTypes))]

			// Find a random position in the room
			x, y := g.dungeon.RandomPointInRoom(roomIndex)
			if x >= 0 && y >= 0 {
				enemy := entity.NewEnemy(enemyType, x, y, roomIndex)
				g.enemies = append(g.enemies, enemy)
			}
		}
	}
}

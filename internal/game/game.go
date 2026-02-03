package game

import (
	"context"

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
	state    State
	running  bool
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

		initSpan.SetAttributes(
			attribute.Int("dungeon.rooms", len(g.dungeon.Rooms)),
			attribute.Int("party.start_x", startX),
			attribute.Int("party.start_y", startY),
		)
	} else {
		// Fallback: place in center of map
		g.party = entity.NewParty(g.dungeon.Width/2, g.dungeon.Height/2)
		initSpan.SetAttributes(
			attribute.Int("dungeon.rooms", 0),
			attribute.String("warning", "no rooms generated, using fallback position"),
		)
	}

	initSpan.End()

	// Main game loop
	for g.running {
		// Render current state
		g.renderer.Render(g.dungeon, g.party)

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
	case tcell.KeyEscape, tcell.KeyCtrlC:
		g.running = false

	case tcell.KeyUp:
		g.tryMove(ctx, 0, -1)
	case tcell.KeyDown:
		g.tryMove(ctx, 0, 1)
	case tcell.KeyLeft:
		g.tryMove(ctx, -1, 0)
	case tcell.KeyRight:
		g.tryMove(ctx, 1, 0)

	case tcell.KeyRune:
		// Handle character keys if needed in future
		switch ev.Rune() {
		case 'q', 'Q':
			g.running = false
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

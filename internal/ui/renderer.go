package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/samdwyer/dungeonband/internal/entity"
	"github.com/samdwyer/dungeonband/internal/world"
)

// Renderer handles drawing the game to the screen.
type Renderer struct {
	screen *Screen
}

// NewRenderer creates a new renderer for the given screen.
func NewRenderer(screen *Screen) *Renderer {
	return &Renderer{screen: screen}
}

// Render draws the dungeon and party to the screen.
func (r *Renderer) Render(dungeon *world.Dungeon, party *entity.Party) {
	r.screen.Clear()

	// Draw dungeon tiles
	for y := 0; y < dungeon.Height; y++ {
		for x := 0; x < dungeon.Width; x++ {
			tile := dungeon.GetTile(x, y)
			style := r.getTileStyle(tile)
			r.screen.SetContent(x, y, tile.Rune(), style)
		}
	}

	// Draw party on top
	partyStyle := tcell.StyleDefault.
		Foreground(tcell.ColorYellow).
		Bold(true)
	r.screen.SetContent(party.X, party.Y, party.Symbol, partyStyle)

	r.screen.Show()
}

// getTileStyle returns the appropriate style for a tile type.
func (r *Renderer) getTileStyle(tile world.Tile) tcell.Style {
	switch tile {
	case world.TileWall:
		return tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
	case world.TileFloor:
		return tcell.StyleDefault.Foreground(tcell.ColorGray)
	default:
		return tcell.StyleDefault
	}
}

// RenderMessage displays a message at the bottom of the screen.
func (r *Renderer) RenderMessage(msg string, y int) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for i, ch := range msg {
		r.screen.SetContent(i, y, ch, style)
	}
}

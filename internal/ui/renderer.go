package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/samdwyer/dungeonband/internal/entity"
	"github.com/samdwyer/dungeonband/internal/world"
)

// GameState represents the current game state for rendering purposes.
type GameState int

const (
	StateExplore GameState = iota
	StateCombat
)

// AbilityInfo holds display information for an ability in the combat UI.
type AbilityInfo struct {
	Name   string
	MPCost int
	CanUse bool // false if not enough MP
}

// CombatInfo holds all information needed to render the combat UI.
type CombatInfo struct {
	ActiveMember *entity.Member  // The party member whose turn it is
	Abilities    []AbilityInfo   // Available abilities for the active member
	Enemies      []*entity.Enemy // Enemies in combat
	Message      string          // Current combat message
}

// Renderer handles drawing the game to the screen.
type Renderer struct {
	screen *Screen
}

// NewRenderer creates a new renderer for the given screen.
func NewRenderer(screen *Screen) *Renderer {
	return &Renderer{screen: screen}
}

// Render draws the dungeon and party to the screen based on game state.
func (r *Renderer) Render(dungeon *world.Dungeon, party *entity.Party, enemies []*entity.Enemy, state GameState, seed int64) {
	r.RenderWithCombat(dungeon, party, enemies, state, seed, nil)
}

// RenderWithCombat draws the game with optional combat UI information.
func (r *Renderer) RenderWithCombat(dungeon *world.Dungeon, party *entity.Party, enemies []*entity.Enemy, state GameState, seed int64, combatInfo *CombatInfo) {
	r.screen.Clear()

	// Determine which room the party is in (for visibility)
	partyRoomIndex := dungeon.RoomIndexAt(party.X, party.Y)

	// Draw dungeon tiles
	for y := 0; y < dungeon.Height; y++ {
		for x := 0; x < dungeon.Width; x++ {
			tile := dungeon.GetTile(x, y)
			style := r.getTileStyle(tile)
			r.screen.SetContent(x, y, tile.Rune(), style)
		}
	}

	// Draw enemies (only those in the same room as party)
	r.renderEnemies(enemies, partyRoomIndex)

	// Draw party based on state
	if state == StateCombat {
		r.renderCombatFormation(dungeon, party, combatInfo)
	} else {
		r.renderExploreParty(party)
	}

	// Draw state indicator in top-left
	r.renderStateIndicator(state)

	// Draw seed in top-right
	r.renderSeed(dungeon.Width, seed)

	// Draw combat UI panel if in combat
	if state == StateCombat && combatInfo != nil {
		r.renderCombatUI(dungeon.Height, combatInfo)
	}

	r.screen.Show()
}

// renderExploreParty draws the party as a single symbol in explore mode.
func (r *Renderer) renderExploreParty(party *entity.Party) {
	partyStyle := tcell.StyleDefault.
		Foreground(tcell.ColorYellow).
		Bold(true)
	r.screen.SetContent(party.X, party.Y, party.Symbol, partyStyle)
}

// renderCombatFormation draws individual party members spread on tiles.
func (r *Renderer) renderCombatFormation(dungeon *world.Dungeon, party *entity.Party, combatInfo *CombatInfo) {
	// Find valid positions for formation around party position
	positions := r.findFormationPositions(dungeon, party.X, party.Y, len(party.Members))

	// Place members at positions
	for i, member := range party.Members {
		if i < len(positions) {
			pos := positions[i]
			member.SetPosition(pos.x, pos.y)
			style := r.getMemberStyle(member.Class)

			// Highlight active member
			if combatInfo != nil && combatInfo.ActiveMember == member {
				style = style.Background(tcell.ColorDarkBlue)
			}

			// Dim dead members
			if !member.IsAlive() {
				style = tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
			}

			r.screen.SetContent(pos.x, pos.y, member.Symbol, style)
		}
	}
}

// position represents a coordinate pair.
type position struct {
	x, y int
}

// findFormationPositions finds valid tiles for party members around center.
// Tries 2x2 formation first, falls back to line formation in corridors.
func (r *Renderer) findFormationPositions(dungeon *world.Dungeon, centerX, centerY, count int) []position {
	// Priority order for 2x2 formation (relative to center):
	// [0][1]  = NW, NE (front row - Warrior, Rogue)
	// [2][3]  = SW, SE (back row - Wizard, Cleric)
	offsets2x2 := []position{
		{-1, 0}, {0, 0}, // Front row (same Y as party, left and center)
		{-1, 1}, {0, 1}, // Back row (below party)
	}

	// Try 2x2 formation
	positions := make([]position, 0, count)
	for _, off := range offsets2x2 {
		x, y := centerX+off.x, centerY+off.y
		if dungeon.IsPassable(x, y) {
			positions = append(positions, position{x, y})
			if len(positions) >= count {
				return positions
			}
		}
	}

	// If we got enough positions, return them
	if len(positions) >= count {
		return positions
	}

	// Fall back to line formation - search in expanding rings
	positions = r.findLineFormation(dungeon, centerX, centerY, count)
	return positions
}

// findLineFormation finds positions in a line or scattered pattern.
func (r *Renderer) findLineFormation(dungeon *world.Dungeon, centerX, centerY, count int) []position {
	positions := make([]position, 0, count)
	visited := make(map[position]bool)

	// Start with center
	if dungeon.IsPassable(centerX, centerY) {
		positions = append(positions, position{centerX, centerY})
		visited[position{centerX, centerY}] = true
	}

	// Expand outward in cardinal directions first, then diagonals
	directions := []position{
		{0, -1}, {0, 1}, {-1, 0}, {1, 0}, // Cardinals
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1}, // Diagonals
	}

	for radius := 1; radius <= 3 && len(positions) < count; radius++ {
		for _, dir := range directions {
			x, y := centerX+dir.x*radius, centerY+dir.y*radius
			pos := position{x, y}
			if !visited[pos] && dungeon.IsPassable(x, y) {
				positions = append(positions, pos)
				visited[pos] = true
				if len(positions) >= count {
					return positions
				}
			}
		}
	}

	return positions
}

// getMemberStyle returns the style for a party member based on class.
func (r *Renderer) getMemberStyle(class entity.Class) tcell.Style {
	switch class {
	case entity.ClassWarrior:
		return tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	case entity.ClassRogue:
		return tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	case entity.ClassWizard:
		return tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true)
	case entity.ClassCleric:
		return tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true)
	default:
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
}

// renderStateIndicator draws the current state in the top-left corner.
func (r *Renderer) renderStateIndicator(state GameState) {
	var text string
	var style tcell.Style

	if state == StateCombat {
		text = "COMBAT"
		style = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	} else {
		text = "EXPLORE"
		style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	}

	for i, ch := range text {
		r.screen.SetContent(i, 0, ch, style)
	}
}

// renderSeed draws the seed value in the top-right corner.
func (r *Renderer) renderSeed(screenWidth int, seed int64) {
	text := fmt.Sprintf("Seed:%d", seed)
	style := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)

	// Position at top-right
	startX := screenWidth - len(text)
	if startX < 0 {
		startX = 0
	}

	for i, ch := range text {
		r.screen.SetContent(startX+i, 0, ch, style)
	}
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

// renderEnemies draws enemies that are visible to the party.
// Only enemies in the same room as the party are rendered.
func (r *Renderer) renderEnemies(enemies []*entity.Enemy, partyRoomIndex int) {
	for _, enemy := range enemies {
		// Only show enemies in the same room as the party
		if enemy.RoomIndex == partyRoomIndex && partyRoomIndex >= 0 {
			style := tcell.StyleDefault.Foreground(enemy.Color())
			r.screen.SetContent(enemy.X, enemy.Y, enemy.Symbol, style)
		}
	}
}

// renderCombatUI draws the combat UI panel below the dungeon.
func (r *Renderer) renderCombatUI(startY int, info *CombatInfo) {
	if info == nil || info.ActiveMember == nil {
		return
	}

	y := startY + 1

	// Draw active member info
	memberLine := fmt.Sprintf("%s's turn | HP: %d/%d | MP: %d/%d",
		info.ActiveMember.Name,
		info.ActiveMember.HP, info.ActiveMember.MaxHP,
		info.ActiveMember.MP, info.ActiveMember.MaxMP,
	)
	r.renderText(0, y, memberLine, tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true))
	y++

	// Draw separator
	r.renderText(0, y, "--- Abilities (press 1-9 to select) ---", tcell.StyleDefault.Foreground(tcell.ColorGray))
	y++

	// Draw abilities
	for i, ability := range info.Abilities {
		if i >= 9 {
			break // Only show first 9 abilities
		}

		var line string
		if ability.MPCost > 0 {
			line = fmt.Sprintf("[%d] %s (%d MP)", i+1, ability.Name, ability.MPCost)
		} else {
			line = fmt.Sprintf("[%d] %s", i+1, ability.Name)
		}

		style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if !ability.CanUse {
			style = tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
		}
		r.renderText(0, y, line, style)
		y++
	}

	y++

	// Draw enemies in combat
	if len(info.Enemies) > 0 {
		r.renderText(0, y, "--- Enemies ---", tcell.StyleDefault.Foreground(tcell.ColorGray))
		y++
		for _, enemy := range info.Enemies {
			if enemy.IsAlive() {
				enemyLine := fmt.Sprintf("%s HP: %d/%d", enemy.Name, enemy.HP, enemy.MaxHP)
				r.renderText(0, y, enemyLine, tcell.StyleDefault.Foreground(enemy.Color()))
				y++
			}
		}
	}

	// Draw combat message
	if info.Message != "" {
		y++
		r.renderText(0, y, info.Message, tcell.StyleDefault.Foreground(tcell.ColorAqua))
	}
}

// renderText draws a string at the given position.
func (r *Renderer) renderText(x, y int, text string, style tcell.Style) {
	for i, ch := range text {
		r.screen.SetContent(x+i, y, ch, style)
	}
}

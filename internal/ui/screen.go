// Package ui provides terminal rendering using tcell.
package ui

import "github.com/gdamore/tcell/v2"

// Screen wraps tcell.Screen with a simplified interface.
type Screen struct {
	screen tcell.Screen
}

// NewScreen creates and initializes a new terminal screen.
func NewScreen() (*Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	s.EnableMouse()
	s.Clear()
	return &Screen{screen: s}, nil
}

// Close finalizes the screen and restores terminal state.
func (s *Screen) Close() {
	s.screen.Fini()
}

// PollEvent waits for and returns the next terminal event.
func (s *Screen) PollEvent() tcell.Event {
	return s.screen.PollEvent()
}

// Clear clears the screen buffer.
func (s *Screen) Clear() {
	s.screen.Clear()
}

// Show flushes the screen buffer to the terminal.
func (s *Screen) Show() {
	s.screen.Show()
}

// SetContent sets a single cell's content at the given position.
func (s *Screen) SetContent(x, y int, r rune, style tcell.Style) {
	s.screen.SetContent(x, y, r, nil, style)
}

// Size returns the current terminal dimensions.
func (s *Screen) Size() (width, height int) {
	return s.screen.Size()
}

// Sync forces a complete redraw of the screen.
func (s *Screen) Sync() {
	s.screen.Sync()
}

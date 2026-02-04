package gamedata

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ParseHexColor converts a hex color string (e.g., "#FF0000" or "FF0000") to a tcell.Color.
func ParseHexColor(hex string) (tcell.Color, error) {
	// Remove leading # if present
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return tcell.ColorDefault, fmt.Errorf("invalid hex color length: %s", hex)
	}

	// Parse RGB components
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid red component in %s: %w", hex, err)
	}

	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid green component in %s: %w", hex, err)
	}

	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return tcell.ColorDefault, fmt.Errorf("invalid blue component in %s: %w", hex, err)
	}

	return tcell.NewRGBColor(int32(r), int32(g), int32(b)), nil
}

// MustParseHexColor converts a hex color string to tcell.Color, panicking on error.
func MustParseHexColor(hex string) tcell.Color {
	color, err := ParseHexColor(hex)
	if err != nil {
		panic(err)
	}
	return color
}

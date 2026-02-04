package gamedata

// ClassDef defines a playable class loaded from JSON.
type ClassDef struct {
	ID        string   `json:"id"`        // Unique identifier matching entity.Class (e.g., "warrior")
	Name      string   `json:"name"`      // Display name (e.g., "Warrior")
	Symbol    string   `json:"symbol"`    // Single character for rendering (e.g., "W")
	HP        int      `json:"hp"`        // Base hit points
	MP        int      `json:"mp"`        // Base mana points
	Attack    int      `json:"attack"`    // Base attack power
	Defense   int      `json:"defense"`   // Base defense value
	Magic     int      `json:"magic"`     // Base magic power
	Abilities []string `json:"abilities"` // List of ability IDs this class can use
}

// SymbolRune returns the symbol as a rune for rendering.
func (c *ClassDef) SymbolRune() rune {
	if len(c.Symbol) == 0 {
		return '?'
	}
	return rune(c.Symbol[0])
}

// ClassesFile represents the structure of classes.json.
type ClassesFile struct {
	Classes []ClassDef `json:"classes"`
}

// LoadClasses loads class definitions from the embedded classes.json file.
func LoadClasses() ([]ClassDef, error) {
	file, err := Load[ClassesFile]("classes.json")
	if err != nil {
		return nil, err
	}
	return file.Classes, nil
}

// MustLoadClasses loads class definitions, panicking on error.
func MustLoadClasses() []ClassDef {
	classes, err := LoadClasses()
	if err != nil {
		panic(err)
	}
	return classes
}

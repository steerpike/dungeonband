package gamedata

import (
	"math/rand"
	"testing"
)

func TestLoadEnemies(t *testing.T) {
	enemies, err := LoadEnemies()
	if err != nil {
		t.Fatalf("Failed to load enemies: %v", err)
	}

	if len(enemies) != 3 {
		t.Errorf("Expected 3 enemies, got %d", len(enemies))
	}

	// Verify expected enemies exist
	expectedIDs := map[string]bool{"goblin": false, "orc": false, "skeleton": false}
	for _, e := range enemies {
		if _, ok := expectedIDs[e.ID]; ok {
			expectedIDs[e.ID] = true
		}
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Expected enemy %q not found", id)
		}
	}
}

func TestEnemyRegistry(t *testing.T) {
	registry, err := LoadEnemyRegistry()
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	if registry.Count() != 3 {
		t.Errorf("Expected 3 enemy types, got %d", registry.Count())
	}

	// Test GetByID
	goblin := registry.GetByID("goblin")
	if goblin == nil {
		t.Error("Goblin not found by ID")
	} else if goblin.Name != "Goblin" {
		t.Errorf("Expected name 'Goblin', got %q", goblin.Name)
	}

	// Test weighted spawning is deterministic with same seed
	rng1 := rand.New(rand.NewSource(12345))
	rng2 := rand.New(rand.NewSource(12345))

	spawns1 := make([]string, 10)
	spawns2 := make([]string, 10)

	for i := 0; i < 10; i++ {
		spawns1[i] = registry.SpawnRandom(rng1).ID
		spawns2[i] = registry.SpawnRandom(rng2).ID
	}

	for i := 0; i < 10; i++ {
		if spawns1[i] != spawns2[i] {
			t.Errorf("Spawn %d mismatch: %s != %s", i, spawns1[i], spawns2[i])
		}
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"#FF0000", true},
		{"FF0000", true},
		{"#00FF00", true},
		{"#0000FF", true},
		{"#FFFFFF", true},
		{"#000000", true},
		{"invalid", false},
		{"#FFF", false}, // Too short
	}

	for _, tt := range tests {
		_, err := ParseHexColor(tt.input)
		if tt.valid && err != nil {
			t.Errorf("ParseHexColor(%q) should be valid, got error: %v", tt.input, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ParseHexColor(%q) should be invalid, got no error", tt.input)
		}
	}
}

func TestEnemyDefMethods(t *testing.T) {
	def := EnemyDef{
		ID:          "test",
		Name:        "Test Enemy",
		Glyph:       "T",
		Color:       "#FF0000",
		HP:          10,
		Attack:      5,
		Defense:     2,
		SpawnWeight: 50,
	}

	if def.GlyphRune() != 'T' {
		t.Errorf("Expected glyph 'T', got %c", def.GlyphRune())
	}

	color := def.TCellColor()
	if color == 0 {
		t.Error("TCellColor returned zero color")
	}
}

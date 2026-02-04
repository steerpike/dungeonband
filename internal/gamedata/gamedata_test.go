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

func TestEnemyAbilities(t *testing.T) {
	enemies, err := LoadEnemies()
	if err != nil {
		t.Fatalf("Failed to load enemies: %v", err)
	}

	// All enemies should have at least one ability
	for _, e := range enemies {
		if len(e.Abilities) == 0 {
			t.Errorf("Enemy %q has no abilities", e.ID)
		}
	}

	// Verify specific enemies have expected abilities
	registry, err := LoadEnemyRegistry()
	if err != nil {
		t.Fatalf("Failed to load enemy registry: %v", err)
	}

	goblin := registry.GetByID("goblin")
	if goblin == nil {
		t.Fatal("Goblin not found")
	}
	hasAttack := false
	for _, a := range goblin.Abilities {
		if a == "attack" {
			hasAttack = true
			break
		}
	}
	if !hasAttack {
		t.Error("Goblin should have 'attack' ability")
	}

	skeleton := registry.GetByID("skeleton")
	if skeleton == nil {
		t.Fatal("Skeleton not found")
	}
	hasBoneThrow := false
	for _, a := range skeleton.Abilities {
		if a == "bone_throw" {
			hasBoneThrow = true
			break
		}
	}
	if !hasBoneThrow {
		t.Error("Skeleton should have 'bone_throw' ability")
	}
}

func TestLoadAbilities(t *testing.T) {
	abilities, err := LoadAbilities()
	if err != nil {
		t.Fatalf("Failed to load abilities: %v", err)
	}

	if len(abilities) < 5 {
		t.Errorf("Expected at least 5 abilities, got %d", len(abilities))
	}

	// Verify expected abilities exist
	expectedIDs := map[string]bool{
		"attack":        false,
		"defend":        false,
		"fireball":      false,
		"heal":          false,
		"poison_strike": false,
	}
	for _, a := range abilities {
		if _, ok := expectedIDs[a.ID]; ok {
			expectedIDs[a.ID] = true
		}
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Expected ability %q not found", id)
		}
	}
}

func TestAbilityRegistry(t *testing.T) {
	registry, err := LoadAbilityRegistry()
	if err != nil {
		t.Fatalf("Failed to load ability registry: %v", err)
	}

	if registry.Count() < 5 {
		t.Errorf("Expected at least 5 abilities, got %d", registry.Count())
	}

	// Test GetByID
	fireball := registry.GetByID("fireball")
	if fireball == nil {
		t.Fatal("Fireball not found by ID")
	}
	if fireball.Name != "Fireball" {
		t.Errorf("Expected name 'Fireball', got %q", fireball.Name)
	}
	if fireball.EffectType != EffectDamage {
		t.Errorf("Expected effectType 'damage', got %q", fireball.EffectType)
	}
	if fireball.DamageType != DamageMagical {
		t.Errorf("Expected damageType 'magical', got %q", fireball.DamageType)
	}

	// Test GetMultiple
	ids := []string{"attack", "heal", "nonexistent"}
	abilities := registry.GetMultiple(ids)
	if len(abilities) != 2 {
		t.Errorf("Expected 2 abilities from GetMultiple, got %d", len(abilities))
	}

	// Test NeedsTarget
	if !fireball.NeedsTarget() {
		t.Error("Fireball should need a target")
	}
	defend := registry.GetByID("defend")
	if defend.NeedsTarget() {
		t.Error("Defend should not need a target (self-target)")
	}

	// Test IsOffensive
	if !fireball.IsOffensive() {
		t.Error("Fireball should be offensive")
	}
	heal := registry.GetByID("heal")
	if heal.IsOffensive() {
		t.Error("Heal should not be offensive")
	}
}

func TestLoadClasses(t *testing.T) {
	classes, err := LoadClasses()
	if err != nil {
		t.Fatalf("Failed to load classes: %v", err)
	}

	if len(classes) != 4 {
		t.Errorf("Expected 4 classes, got %d", len(classes))
	}

	// Verify expected classes exist
	expectedIDs := map[string]bool{
		"warrior": false,
		"rogue":   false,
		"wizard":  false,
		"cleric":  false,
	}
	for _, c := range classes {
		if _, ok := expectedIDs[c.ID]; ok {
			expectedIDs[c.ID] = true
		}
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Expected class %q not found", id)
		}
	}
}

func TestClassRegistry(t *testing.T) {
	registry, err := LoadClassRegistry()
	if err != nil {
		t.Fatalf("Failed to load class registry: %v", err)
	}

	if registry.Count() != 4 {
		t.Errorf("Expected 4 classes, got %d", registry.Count())
	}

	// Test GetByID
	warrior := registry.GetByID("warrior")
	if warrior == nil {
		t.Fatal("Warrior not found by ID")
	}
	if warrior.Name != "Warrior" {
		t.Errorf("Expected name 'Warrior', got %q", warrior.Name)
	}
	if warrior.HP != 30 {
		t.Errorf("Expected HP 30, got %d", warrior.HP)
	}

	// Verify warrior has expected abilities
	hasAttack := false
	hasPowerAttack := false
	for _, a := range warrior.Abilities {
		if a == "attack" {
			hasAttack = true
		}
		if a == "power_attack" {
			hasPowerAttack = true
		}
	}
	if !hasAttack {
		t.Error("Warrior should have 'attack' ability")
	}
	if !hasPowerAttack {
		t.Error("Warrior should have 'power_attack' ability")
	}

	// Test cleric has healing abilities
	cleric := registry.GetByID("cleric")
	if cleric == nil {
		t.Fatal("Cleric not found by ID")
	}
	hasHeal := false
	hasGroupHeal := false
	for _, a := range cleric.Abilities {
		if a == "heal" {
			hasHeal = true
		}
		if a == "group_heal" {
			hasGroupHeal = true
		}
	}
	if !hasHeal {
		t.Error("Cleric should have 'heal' ability")
	}
	if !hasGroupHeal {
		t.Error("Cleric should have 'group_heal' ability")
	}
}

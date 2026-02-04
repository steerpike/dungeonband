package gamedata

import (
	"errors"
	"math/rand"
)

// EnemyRegistry holds loaded enemy definitions and provides spawning utilities.
type EnemyRegistry struct {
	enemies     []EnemyDef
	totalWeight int
}

// NewEnemyRegistry creates a registry from loaded enemy definitions.
func NewEnemyRegistry(enemies []EnemyDef) *EnemyRegistry {
	totalWeight := 0
	for _, e := range enemies {
		totalWeight += e.SpawnWeight
	}
	return &EnemyRegistry{
		enemies:     enemies,
		totalWeight: totalWeight,
	}
}

// LoadEnemyRegistry loads and creates a registry from the embedded enemies.json.
func LoadEnemyRegistry() (*EnemyRegistry, error) {
	enemies, err := LoadEnemies()
	if err != nil {
		return nil, err
	}
	if len(enemies) == 0 {
		return nil, errors.New("no enemies loaded from enemies.json")
	}
	return NewEnemyRegistry(enemies), nil
}

// MustLoadEnemyRegistry loads a registry, panicking on error.
func MustLoadEnemyRegistry() *EnemyRegistry {
	registry, err := LoadEnemyRegistry()
	if err != nil {
		panic(err)
	}
	return registry
}

// SpawnRandom selects a random enemy definition using weighted probability.
// Enemies with higher spawnWeight are more likely to be selected.
func (r *EnemyRegistry) SpawnRandom(rng *rand.Rand) *EnemyDef {
	if r.totalWeight <= 0 || len(r.enemies) == 0 {
		return nil
	}

	// Pick a random value in the total weight range
	roll := rng.Intn(r.totalWeight)

	// Find which enemy this roll corresponds to
	cumulative := 0
	for i := range r.enemies {
		cumulative += r.enemies[i].SpawnWeight
		if roll < cumulative {
			return &r.enemies[i]
		}
	}

	// Fallback (shouldn't happen)
	return &r.enemies[0]
}

// GetByID returns the enemy definition with the given ID, or nil if not found.
func (r *EnemyRegistry) GetByID(id string) *EnemyDef {
	for i := range r.enemies {
		if r.enemies[i].ID == id {
			return &r.enemies[i]
		}
	}
	return nil
}

// All returns all enemy definitions.
func (r *EnemyRegistry) All() []EnemyDef {
	return r.enemies
}

// Count returns the number of enemy types in the registry.
func (r *EnemyRegistry) Count() int {
	return len(r.enemies)
}

// =============================================================================
// AbilityRegistry
// =============================================================================

// AbilityRegistry holds loaded ability definitions and provides lookup utilities.
type AbilityRegistry struct {
	abilities map[string]*AbilityDef
	all       []AbilityDef
}

// NewAbilityRegistry creates a registry from loaded ability definitions.
func NewAbilityRegistry(abilities []AbilityDef) *AbilityRegistry {
	registry := &AbilityRegistry{
		abilities: make(map[string]*AbilityDef),
		all:       abilities,
	}
	for i := range abilities {
		registry.abilities[abilities[i].ID] = &abilities[i]
	}
	return registry
}

// LoadAbilityRegistry loads and creates a registry from the embedded abilities.json.
func LoadAbilityRegistry() (*AbilityRegistry, error) {
	abilities, err := LoadAbilities()
	if err != nil {
		return nil, err
	}
	if len(abilities) == 0 {
		return nil, errors.New("no abilities loaded from abilities.json")
	}
	return NewAbilityRegistry(abilities), nil
}

// MustLoadAbilityRegistry loads a registry, panicking on error.
func MustLoadAbilityRegistry() *AbilityRegistry {
	registry, err := LoadAbilityRegistry()
	if err != nil {
		panic(err)
	}
	return registry
}

// GetByID returns the ability definition with the given ID, or nil if not found.
func (r *AbilityRegistry) GetByID(id string) *AbilityDef {
	return r.abilities[id]
}

// GetMultiple returns ability definitions for a list of IDs.
// Missing IDs are silently skipped.
func (r *AbilityRegistry) GetMultiple(ids []string) []*AbilityDef {
	result := make([]*AbilityDef, 0, len(ids))
	for _, id := range ids {
		if ability := r.abilities[id]; ability != nil {
			result = append(result, ability)
		}
	}
	return result
}

// All returns all ability definitions.
func (r *AbilityRegistry) All() []AbilityDef {
	return r.all
}

// Count returns the number of abilities in the registry.
func (r *AbilityRegistry) Count() int {
	return len(r.all)
}

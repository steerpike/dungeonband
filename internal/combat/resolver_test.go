package combat

import (
	"testing"

	"github.com/samdwyer/dungeonband/internal/gamedata"
)

// mockCombatant is a test implementation of the Combatant interface.
type mockCombatant struct {
	name          string
	hp, maxHP     int
	mp, maxMP     int
	attack        int
	defense       int
	magic         int
	abilityIDs    []string
	statusEffects []StatusEffect
}

func newMockCombatant(name string, hp, mp, attack, defense, magic int) *mockCombatant {
	return &mockCombatant{
		name:          name,
		hp:            hp,
		maxHP:         hp,
		mp:            mp,
		maxMP:         mp,
		attack:        attack,
		defense:       defense,
		magic:         magic,
		abilityIDs:    []string{},
		statusEffects: []StatusEffect{},
	}
}

func (m *mockCombatant) GetName() string         { return m.name }
func (m *mockCombatant) IsAlive() bool           { return m.hp > 0 }
func (m *mockCombatant) GetHP() int              { return m.hp }
func (m *mockCombatant) GetMaxHP() int           { return m.maxHP }
func (m *mockCombatant) GetMP() int              { return m.mp }
func (m *mockCombatant) GetMaxMP() int           { return m.maxMP }
func (m *mockCombatant) GetAttack() int          { return m.attack }
func (m *mockCombatant) GetDefense() int         { return m.defense }
func (m *mockCombatant) GetMagic() int           { return m.magic }
func (m *mockCombatant) GetAbilityIDs() []string { return m.abilityIDs }

func (m *mockCombatant) TakeDamage(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if actual > m.hp {
		actual = m.hp
	}
	m.hp -= actual
	return actual
}

func (m *mockCombatant) Heal(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if m.hp+actual > m.maxHP {
		actual = m.maxHP - m.hp
	}
	m.hp += actual
	return actual
}

func (m *mockCombatant) SpendMP(amount int) bool {
	if m.mp < amount {
		return false
	}
	m.mp -= amount
	return true
}

func (m *mockCombatant) RestoreMP(amount int) int {
	if amount <= 0 {
		return 0
	}
	actual := amount
	if m.mp+actual > m.maxMP {
		actual = m.maxMP - m.mp
	}
	m.mp += actual
	return actual
}

func (m *mockCombatant) GetStatusEffects() []StatusEffect {
	return m.statusEffects
}

func (m *mockCombatant) AddStatusEffect(effect StatusEffect) {
	// Replace existing effect of same type
	for i, existing := range m.statusEffects {
		if existing.Type == effect.Type {
			m.statusEffects[i] = effect
			return
		}
	}
	m.statusEffects = append(m.statusEffects, effect)
}

func (m *mockCombatant) RemoveStatusEffect(effectType gamedata.StatusEffectType) {
	for i, existing := range m.statusEffects {
		if existing.Type == effectType {
			m.statusEffects = append(m.statusEffects[:i], m.statusEffects[i+1:]...)
			return
		}
	}
}

func (m *mockCombatant) TickStatusEffects() []StatusTick {
	var ticks []StatusTick
	remaining := []StatusEffect{}

	for _, effect := range m.statusEffects {
		tick := StatusTick{Type: effect.Type}

		switch effect.Type {
		case gamedata.StatusPoison:
			tick.Amount = m.TakeDamage(effect.Power)
		case gamedata.StatusRegen:
			tick.Amount = m.Heal(effect.Power)
		}

		effect.RemainingTurns--
		if effect.RemainingTurns <= 0 {
			tick.Ended = true
		} else {
			remaining = append(remaining, effect)
		}
		ticks = append(ticks, tick)
	}

	m.statusEffects = remaining
	return ticks
}

func TestResolveDamagePhysical(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	// Attacker: 8 attack, Target: 3 defense
	// Attack ability: basePower 5
	// Expected: 5 + 8 - 3 = 10 damage
	attacker := newMockCombatant("Warrior", 30, 0, 8, 6, 0)
	target := newMockCombatant("Goblin", 15, 0, 2, 3, 0)

	attack := registry.GetByID("attack")
	if attack == nil {
		t.Fatal("attack ability not found")
	}

	result := resolver.Resolve(attack, attacker, target)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Message)
	}
	if result.Damage != 10 {
		t.Errorf("Expected 10 damage, got %d", result.Damage)
	}
	if target.GetHP() != 5 {
		t.Errorf("Expected target HP 5, got %d", target.GetHP())
	}
}

func TestResolveDamagePhysicalMinimum(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	// Attacker: 2 attack, Target: 10 defense
	// Attack ability: basePower 0
	// Expected: 0 + 2 - 10 = -8 -> min 1 damage
	attacker := newMockCombatant("Weak", 10, 0, 2, 0, 0)
	target := newMockCombatant("Tank", 50, 0, 0, 10, 0)

	attack := registry.GetByID("attack")
	result := resolver.Resolve(attack, attacker, target)

	if result.Damage != 1 {
		t.Errorf("Expected minimum 1 damage, got %d", result.Damage)
	}
}

func TestResolveDamageMagical(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	// Wizard: 10 magic
	// Fireball: basePower 12, magical damage
	// Expected: 12 + 10 = 22 damage (defense ignored)
	wizard := newMockCombatant("Wizard", 15, 20, 2, 2, 10)
	target := newMockCombatant("Armored Orc", 30, 0, 4, 8, 0) // High defense shouldn't matter

	fireball := registry.GetByID("fireball")
	if fireball == nil {
		t.Fatal("fireball ability not found")
	}

	result := resolver.Resolve(fireball, wizard, target)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Message)
	}
	// Fireball: 12 base + 10 magic = 22
	if result.Damage != 22 {
		t.Errorf("Expected 22 magical damage, got %d", result.Damage)
	}
	if target.GetHP() != 8 {
		t.Errorf("Expected target HP 8, got %d", target.GetHP())
	}
}

func TestResolveHeal(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	// Cleric: 8 magic
	// Heal: basePower 10
	// Expected: 10 + 8 = 18 healing
	cleric := newMockCombatant("Cleric", 22, 15, 4, 4, 8)
	wounded := newMockCombatant("Wounded Warrior", 30, 0, 8, 6, 0)
	wounded.hp = 10 // Simulate damage taken

	heal := registry.GetByID("heal")
	if heal == nil {
		t.Fatal("heal ability not found")
	}

	result := resolver.Resolve(heal, cleric, wounded)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Message)
	}
	// Heal: 10 base + 8 magic = 18 healing
	if result.Healing != 18 {
		t.Errorf("Expected 18 healing, got %d", result.Healing)
	}
	if wounded.GetHP() != 28 {
		t.Errorf("Expected target HP 28, got %d", wounded.GetHP())
	}
}

func TestResolveHealCapped(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	// Healing should be capped at max HP
	cleric := newMockCombatant("Cleric", 22, 15, 4, 4, 8)
	slightlyWounded := newMockCombatant("Warrior", 30, 0, 8, 6, 0)
	slightlyWounded.hp = 28 // Only 2 HP missing

	heal := registry.GetByID("heal")
	result := resolver.Resolve(heal, cleric, slightlyWounded)

	// Should only heal 2, even though heal amount would be 18
	if result.Healing != 2 {
		t.Errorf("Expected 2 healing (capped), got %d", result.Healing)
	}
	if slightlyWounded.GetHP() != 30 {
		t.Errorf("Expected target HP 30 (max), got %d", slightlyWounded.GetHP())
	}
}

func TestResolveInsufficientMP(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	// Wizard with no MP tries to cast fireball
	wizard := newMockCombatant("Wizard", 15, 0, 2, 2, 10) // 0 MP
	target := newMockCombatant("Goblin", 10, 0, 2, 1, 0)

	fireball := registry.GetByID("fireball")
	result := resolver.Resolve(fireball, wizard, target)

	if result.Success {
		t.Error("Expected failure due to insufficient MP")
	}
	if target.GetHP() != 10 {
		t.Error("Target should not have taken damage")
	}
}

func TestResolveMPCost(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	wizard := newMockCombatant("Wizard", 15, 20, 2, 2, 10)
	target := newMockCombatant("Goblin", 10, 0, 2, 1, 0)

	fireball := registry.GetByID("fireball")
	mpBefore := wizard.GetMP()

	resolver.Resolve(fireball, wizard, target)

	mpAfter := wizard.GetMP()
	expectedCost := fireball.MPCost
	if mpBefore-mpAfter != expectedCost {
		t.Errorf("Expected MP cost %d, actual cost %d", expectedCost, mpBefore-mpAfter)
	}
}

func TestResolvePoisonStrike(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	rogue := newMockCombatant("Rogue", 20, 5, 6, 3, 2)
	target := newMockCombatant("Orc", 15, 0, 4, 2, 0)

	poisonStrike := registry.GetByID("poison_strike")
	if poisonStrike == nil {
		t.Fatal("poison_strike ability not found")
	}

	result := resolver.Resolve(poisonStrike, rogue, target)

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Message)
	}
	// Should deal damage AND apply poison
	if result.Damage == 0 {
		t.Error("Expected damage from poison_strike")
	}
	if result.StatusAdded != gamedata.StatusPoison {
		t.Error("Expected poison status to be applied")
	}

	// Verify poison is on the target
	effects := target.GetStatusEffects()
	hasPoison := false
	for _, e := range effects {
		if e.Type == gamedata.StatusPoison {
			hasPoison = true
			break
		}
	}
	if !hasPoison {
		t.Error("Target should have poison status effect")
	}
}

func TestResolveDefend(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	warrior := newMockCombatant("Warrior", 30, 0, 8, 6, 0)

	defend := registry.GetByID("defend")
	if defend == nil {
		t.Fatal("defend ability not found")
	}

	result := resolver.Resolve(defend, warrior, warrior) // Self-target

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Message)
	}
	if result.StatusAdded != gamedata.StatusDefenseUp {
		t.Errorf("Expected defense_up status, got %s", result.StatusAdded)
	}
}

func TestStatusEffectTick(t *testing.T) {
	target := newMockCombatant("Victim", 20, 0, 0, 0, 0)

	// Apply poison: 3 damage per turn, 2 turns
	target.AddStatusEffect(StatusEffect{
		Type:           gamedata.StatusPoison,
		RemainingTurns: 2,
		Power:          3,
	})

	// First tick
	ticks := target.TickStatusEffects()
	if len(ticks) != 1 {
		t.Fatalf("Expected 1 tick, got %d", len(ticks))
	}
	if ticks[0].Amount != 3 {
		t.Errorf("Expected 3 poison damage, got %d", ticks[0].Amount)
	}
	if ticks[0].Ended {
		t.Error("Poison should not have ended yet")
	}
	if target.GetHP() != 17 {
		t.Errorf("Expected HP 17, got %d", target.GetHP())
	}

	// Second tick - should end
	ticks = target.TickStatusEffects()
	if len(ticks) != 1 {
		t.Fatalf("Expected 1 tick, got %d", len(ticks))
	}
	if !ticks[0].Ended {
		t.Error("Poison should have ended")
	}
	if target.GetHP() != 14 {
		t.Errorf("Expected HP 14, got %d", target.GetHP())
	}

	// Third tick - no more effects
	ticks = target.TickStatusEffects()
	if len(ticks) != 0 {
		t.Error("Expected no ticks after poison ended")
	}
}

func TestCalculateDamagePreview(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	// Attacker: 8 attack, Target: 3 defense
	// Attack: basePower 5
	// Expected: 5 + 8 - 3 = 10 damage
	attacker := newMockCombatant("Warrior", 30, 0, 8, 6, 0)
	target := newMockCombatant("Goblin", 15, 0, 2, 3, 0)

	attack := registry.GetByID("attack")
	damage := resolver.CalculateDamage(attack, attacker, target)

	// Should calculate but not apply
	if damage != 10 {
		t.Errorf("Expected preview damage 10, got %d", damage)
	}
	if target.GetHP() != 15 {
		t.Error("Preview should not have damaged target")
	}
}

func TestCanUse(t *testing.T) {
	registry := gamedata.MustLoadAbilityRegistry()
	resolver := NewEffectResolver(registry)

	wizard := newMockCombatant("Wizard", 15, 5, 2, 2, 10)
	fireball := registry.GetByID("fireball")

	// Should be able to use fireball (5 MP cost, have 5 MP)
	if !resolver.CanUse(fireball, wizard) {
		t.Error("Should be able to use fireball with exactly enough MP")
	}

	// Spend some MP
	wizard.SpendMP(1)

	// Should NOT be able to use fireball now (5 MP cost, have 4 MP)
	if resolver.CanUse(fireball, wizard) {
		t.Error("Should not be able to use fireball with insufficient MP")
	}
}

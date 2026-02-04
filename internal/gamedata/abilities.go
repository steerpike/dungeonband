package gamedata

// =============================================================================
// ABILITY SYSTEM DESIGN
// =============================================================================
//
// Overview:
// ---------
// Abilities are composable, data-driven actions that can be used by both party
// members and enemies during combat. They are defined in JSON and loaded at
// game startup.
//
// Core Concepts:
// --------------
//
// 1. EffectType - What the ability does:
//    - damage: Reduces target HP
//    - heal: Restores target HP
//    - buff: Applies positive status effect
//    - debuff: Applies negative status effect
//
// 2. TargetType - Who the ability affects:
//    - self: The caster only
//    - single_enemy: One enemy (requires selection)
//    - all_enemies: All enemies in combat
//    - single_ally: One ally (requires selection)
//    - all_allies: All party members
//
// 3. DamageType - For damage/heal abilities:
//    - physical: Reduced by Defense stat
//    - magical: Ignores Defense (or uses Magic Defense if we add it)
//    - true: Cannot be reduced
//
// 4. StatusEffect - For buff/debuff abilities:
//    - poison: Damage over time
//    - regen: Heal over time
//    - defense_up: Increased defense
//    - defense_down: Decreased defense
//    - attack_up: Increased attack
//    - attack_down: Decreased attack
//
// JSON Schema:
// ------------
// {
//   "id": "fireball",
//   "name": "Fireball",
//   "description": "Hurls a ball of fire at the enemy",
//   "effectType": "damage",
//   "targetType": "single_enemy",
//   "damageType": "magical",
//   "basePower": 15,
//   "mpCost": 5,
//   "cooldown": 0,
//   "statusEffect": null,
//   "statusDuration": 0
// }
//
// Damage Calculation:
// -------------------
// Physical: damage = basePower + attacker.Attack - target.Defense (min 1)
// Magical:  damage = basePower + attacker.Magic (min 1)
// True:     damage = basePower
//
// Healing: heal = basePower + caster.Magic (or flat basePower)
//
// Integration Points:
// -------------------
// 1. EnemyDef.Abilities []string - list of ability IDs enemy can use
// 2. ClassDef.Abilities []string - list of ability IDs class starts with
// 3. Member gains HP, MP, Attack, Defense, Magic stats
// 4. Enemy already has HP, Attack, Defense (add MP, Magic)
// 5. Combat system resolves abilities turn by turn
//
// Turn Order:
// -----------
// Simple: Party members act first (in order), then enemies (in order)
// Future: Speed-based initiative system
//
// Combat Flow:
// ------------
// 1. Enter combat (triggered by proximity to enemy or manual)
// 2. Display combat UI showing party and enemies with HP bars
// 3. For each party member's turn:
//    a. Show available abilities
//    b. Player selects ability (1-9 keys or menu)
//    c. If ability needs target, player selects target
//    d. Resolve ability effect
//    e. Check for victory (all enemies dead) or defeat (all party dead)
// 4. For each enemy's turn:
//    a. AI selects ability (random or weighted by situation)
//    b. AI selects target (random or lowest HP)
//    c. Resolve ability effect
//    d. Check for victory/defeat
// 5. Repeat until combat ends
//
// Telemetry:
// ----------
// - combat.start: party_size, enemy_count, room_index
// - combat.turn: actor_name, ability_id, target_name, damage/heal amount
// - combat.end: outcome (victory/defeat/flee), turns_taken, party_hp_remaining

// EffectType represents what an ability does.
type EffectType string

const (
	EffectDamage EffectType = "damage"
	EffectHeal   EffectType = "heal"
	EffectBuff   EffectType = "buff"
	EffectDebuff EffectType = "debuff"
)

// TargetType represents who an ability can target.
type TargetType string

const (
	TargetSelf        TargetType = "self"
	TargetSingleEnemy TargetType = "single_enemy"
	TargetAllEnemies  TargetType = "all_enemies"
	TargetSingleAlly  TargetType = "single_ally"
	TargetAllAllies   TargetType = "all_allies"
)

// DamageType represents how damage is calculated.
type DamageType string

const (
	DamagePhysical DamageType = "physical"
	DamageMagical  DamageType = "magical"
	DamageTrue     DamageType = "true"
)

// StatusEffectType represents status effects that can be applied.
type StatusEffectType string

const (
	StatusNone        StatusEffectType = ""
	StatusPoison      StatusEffectType = "poison"
	StatusRegen       StatusEffectType = "regen"
	StatusDefenseUp   StatusEffectType = "defense_up"
	StatusDefenseDown StatusEffectType = "defense_down"
	StatusAttackUp    StatusEffectType = "attack_up"
	StatusAttackDown  StatusEffectType = "attack_down"
)

// AbilityDef defines an ability loaded from JSON.
type AbilityDef struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	EffectType     EffectType       `json:"effectType"`
	TargetType     TargetType       `json:"targetType"`
	DamageType     DamageType       `json:"damageType,omitempty"`
	BasePower      int              `json:"basePower"`
	MPCost         int              `json:"mpCost"`
	Cooldown       int              `json:"cooldown"`
	StatusEffect   StatusEffectType `json:"statusEffect,omitempty"`
	StatusDuration int              `json:"statusDuration,omitempty"`
	StatusPower    int              `json:"statusPower,omitempty"` // For DoT/HoT effects
}

// NeedsTarget returns true if the ability requires target selection.
func (a *AbilityDef) NeedsTarget() bool {
	return a.TargetType == TargetSingleEnemy || a.TargetType == TargetSingleAlly
}

// IsOffensive returns true if the ability targets enemies.
func (a *AbilityDef) IsOffensive() bool {
	return a.TargetType == TargetSingleEnemy || a.TargetType == TargetAllEnemies
}

// AbilitiesFile represents the structure of abilities.json.
type AbilitiesFile struct {
	Abilities []AbilityDef `json:"abilities"`
}

// LoadAbilities loads ability definitions from the embedded abilities.json file.
func LoadAbilities() ([]AbilityDef, error) {
	file, err := Load[AbilitiesFile]("abilities.json")
	if err != nil {
		return nil, err
	}
	return file.Abilities, nil
}

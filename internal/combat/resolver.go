// Package combat provides the turn-based combat system for DungeonBand.
package combat

import (
	"github.com/samdwyer/dungeonband/internal/gamedata"
)

// Combatant is the interface for any entity that can participate in combat.
// Both party members and enemies implement this interface.
type Combatant interface {
	// Identity
	GetName() string
	IsAlive() bool

	// Stats
	GetHP() int
	GetMaxHP() int
	GetMP() int
	GetMaxMP() int
	GetAttack() int
	GetDefense() int
	GetMagic() int

	// Mutations
	TakeDamage(amount int) int // Returns actual damage taken
	Heal(amount int) int       // Returns actual amount healed
	SpendMP(amount int) bool   // Returns false if insufficient MP
	RestoreMP(amount int) int  // Returns actual amount restored

	// Abilities
	GetAbilityIDs() []string

	// Status effects
	GetStatusEffects() []StatusEffect
	AddStatusEffect(effect StatusEffect)
	RemoveStatusEffect(effectType gamedata.StatusEffectType)
	TickStatusEffects() []StatusTick // Process turn-based effects, returns what happened
}

// StatusEffect represents an active status effect on a combatant.
type StatusEffect struct {
	Type           gamedata.StatusEffectType
	RemainingTurns int
	Power          int // For DoT/HoT: damage/heal per turn
}

// StatusTick represents what happened when a status effect was processed.
type StatusTick struct {
	Type   gamedata.StatusEffectType
	Amount int  // Damage taken or healing received
	Ended  bool // True if the effect expired
}

// EffectResult contains the outcome of resolving an ability.
type EffectResult struct {
	Success     bool
	Damage      int                       // For damage abilities
	Healing     int                       // For heal abilities
	StatusAdded gamedata.StatusEffectType // For buff/debuff abilities
	Message     string                    // Human-readable description
}

// EffectResolver calculates and applies ability effects.
type EffectResolver struct {
	abilityRegistry *gamedata.AbilityRegistry
}

// NewEffectResolver creates a new effect resolver.
func NewEffectResolver(abilityRegistry *gamedata.AbilityRegistry) *EffectResolver {
	return &EffectResolver{
		abilityRegistry: abilityRegistry,
	}
}

// Resolve applies an ability from the user to the target(s) and returns results.
// For multi-target abilities, this should be called once per target.
func (r *EffectResolver) Resolve(ability *gamedata.AbilityDef, user Combatant, target Combatant) EffectResult {
	if ability == nil {
		return EffectResult{Success: false, Message: "Invalid ability"}
	}

	// Check MP cost
	if ability.MPCost > 0 && user.GetMP() < ability.MPCost {
		return EffectResult{
			Success: false,
			Message: user.GetName() + " doesn't have enough MP!",
		}
	}

	// Spend MP
	if ability.MPCost > 0 {
		user.SpendMP(ability.MPCost)
	}

	switch ability.EffectType {
	case gamedata.EffectDamage:
		return r.resolveDamage(ability, user, target)
	case gamedata.EffectHeal:
		return r.resolveHeal(ability, user, target)
	case gamedata.EffectBuff, gamedata.EffectDebuff:
		return r.resolveStatusEffect(ability, user, target)
	default:
		return EffectResult{Success: false, Message: "Unknown ability effect type"}
	}
}

// CanUse checks if a combatant can use an ability (has enough MP).
func (r *EffectResolver) CanUse(ability *gamedata.AbilityDef, user Combatant) bool {
	if ability == nil {
		return false
	}
	return user.GetMP() >= ability.MPCost
}

// resolveDamage handles damage-type abilities.
func (r *EffectResolver) resolveDamage(ability *gamedata.AbilityDef, user Combatant, target Combatant) EffectResult {
	var damage int

	switch ability.DamageType {
	case gamedata.DamagePhysical:
		// Physical: basePower + attacker.Attack - target.Defense (min 1)
		damage = ability.BasePower + user.GetAttack() - target.GetDefense()
		if damage < 1 {
			damage = 1
		}
	case gamedata.DamageMagical:
		// Magical: basePower + attacker.Magic (min 1)
		damage = ability.BasePower + user.GetMagic()
		if damage < 1 {
			damage = 1
		}
	case gamedata.DamageTrue:
		// True: basePower (unmitigated)
		damage = ability.BasePower
	default:
		// Fallback to physical calculation
		damage = ability.BasePower + user.GetAttack() - target.GetDefense()
		if damage < 1 {
			damage = 1
		}
	}

	// Apply damage to target
	actualDamage := target.TakeDamage(damage)

	// Check if ability also applies a status effect (e.g., poison_strike)
	result := EffectResult{
		Success: true,
		Damage:  actualDamage,
		Message: user.GetName() + " uses " + ability.Name + " on " + target.GetName() + "!",
	}

	if ability.StatusEffect != "" && ability.StatusEffect != gamedata.StatusNone {
		effect := StatusEffect{
			Type:           ability.StatusEffect,
			RemainingTurns: ability.StatusDuration,
			Power:          ability.StatusPower,
		}
		target.AddStatusEffect(effect)
		result.StatusAdded = ability.StatusEffect
	}

	return result
}

// resolveHeal handles heal-type abilities.
func (r *EffectResolver) resolveHeal(ability *gamedata.AbilityDef, user Combatant, target Combatant) EffectResult {
	// Healing: basePower + caster.Magic
	healAmount := ability.BasePower + user.GetMagic()
	if healAmount < 1 {
		healAmount = 1
	}

	actualHealing := target.Heal(healAmount)

	result := EffectResult{
		Success: true,
		Healing: actualHealing,
		Message: user.GetName() + " uses " + ability.Name + " on " + target.GetName() + "!",
	}

	// Check if heal also applies a status effect (e.g., regen)
	if ability.StatusEffect != "" && ability.StatusEffect != gamedata.StatusNone {
		effect := StatusEffect{
			Type:           ability.StatusEffect,
			RemainingTurns: ability.StatusDuration,
			Power:          ability.StatusPower,
		}
		target.AddStatusEffect(effect)
		result.StatusAdded = ability.StatusEffect
	}

	return result
}

// resolveStatusEffect handles buff and debuff abilities.
func (r *EffectResolver) resolveStatusEffect(ability *gamedata.AbilityDef, user Combatant, target Combatant) EffectResult {
	if ability.StatusEffect == "" || ability.StatusEffect == gamedata.StatusNone {
		return EffectResult{
			Success: false,
			Message: ability.Name + " has no status effect defined",
		}
	}

	effect := StatusEffect{
		Type:           ability.StatusEffect,
		RemainingTurns: ability.StatusDuration,
		Power:          ability.StatusPower,
	}
	target.AddStatusEffect(effect)

	return EffectResult{
		Success:     true,
		StatusAdded: ability.StatusEffect,
		Message:     user.GetName() + " uses " + ability.Name + " on " + target.GetName() + "!",
	}
}

// CalculateDamage calculates damage without applying it (for AI/preview).
func (r *EffectResolver) CalculateDamage(ability *gamedata.AbilityDef, user Combatant, target Combatant) int {
	if ability == nil || ability.EffectType != gamedata.EffectDamage {
		return 0
	}

	var damage int
	switch ability.DamageType {
	case gamedata.DamagePhysical:
		damage = ability.BasePower + user.GetAttack() - target.GetDefense()
	case gamedata.DamageMagical:
		damage = ability.BasePower + user.GetMagic()
	case gamedata.DamageTrue:
		damage = ability.BasePower
	default:
		damage = ability.BasePower + user.GetAttack() - target.GetDefense()
	}
	if damage < 1 {
		damage = 1
	}
	return damage
}

// CalculateHealing calculates healing without applying it (for AI/preview).
func (r *EffectResolver) CalculateHealing(ability *gamedata.AbilityDef, user Combatant) int {
	if ability == nil || ability.EffectType != gamedata.EffectHeal {
		return 0
	}
	healing := ability.BasePower + user.GetMagic()
	if healing < 1 {
		healing = 1
	}
	return healing
}

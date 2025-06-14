package services

// Difficulty level constants
const (
	difficultyEasy   = "easy"
	difficultyMedium = "medium"
	difficultyHard   = "hard"
	difficultyDeadly = "deadly"
)

// Size constants
const (
	sizeSmall  = "small"
	sizeMedium = "medium"
	sizeLarge  = "large"
)

// Action type constants  
const (
	actionTypeSpell       = "spell"
	actionTypeAbility     = "ability"
	actionTypeAttack      = "attack"
	actionTypeMelee       = "melee"
	actionTypeRanged      = "ranged"
	actionTypeMovement    = "movement"
	actionTypeKillingBlow = "killing_blow"
)

// Outcome constants
const (
	outcomeHit        = "hit"
	outcomeMiss       = "miss"
	outcomeCritical   = "critical"
	outcomeNeutral    = "neutral"
	outcomeKillingBlow = "killing_blow"
)

// Category constants
const (
	categoryAbility = "ability"
	categorySpell   = "spell"
	categoryItem    = "item"
)

// Character class constants
const (
	classWarlock = "warlock"
	classWizard  = "wizard"
	classFighter = "fighter"
	classRogue   = "rogue"
	classCleric  = "cleric"
	classRanger  = "ranger"
)

// Encounter type constants
const (
	encounterTypeCombat      = "combat"
	encounterTypeSocial      = "social"
	encounterTypeExploration = "exploration"
	encounterTypePuzzle      = "puzzle"
)

// Dice constants
const (
	diceD4  = "1d4"
	diceD6  = "1d6" 
	diceD8  = "1d8"
	diceD10 = "1d10"
	diceD12 = "1d12"
	diceD20 = "1d20"
)

// Common formula constants
const (
	defaultDamageFormula = "1d6"
)
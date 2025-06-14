package constants

// Character classes
const (
	ClassBarbarian = "barbarian"
	ClassBard      = "bard"
	ClassCleric    = "cleric"
	ClassDruid     = "druid"
	ClassFighter   = "fighter"
	ClassPaladin   = "paladin"
	ClassRanger    = "ranger"
	ClassSorcerer  = "sorcerer"
	ClassWizard    = "wizard"
	// ClassWarlock already defined in strings.go
)

// Size categories
const (
	SizeSmall  = "small"
	SizeMedium = "medium"
	// SizeLarge already defined in strings.go
)

// Rarity levels
const (
	RarityUncommon = "uncommon"
	RarityRare     = "rare"
	// RarityCommon and RarityVeryRare already defined in strings.go
)

// Action and ability types
const (
	ActionAttack      = "attack"
	ActionHeal        = "heal"
	ActionHit         = "hit"
	ActionCritical    = "critical"
	ActionKillingBlow = "killing_blow"
	ActionCombat      = "combat"
	ActionAbility     = "ability"
	ActionTypeSpell   = "spell"
	ActionTypeAbility = "ability"
	// ActionSpell can use CategorySpell from strings.go
)

// Difficulty levels
const (
	DifficultyEasy   = "easy"
	DifficultyDeadly = "deadly"
	// DifficultyHard already defined in strings.go
)

// Terrain and climate types
const (
	TerrainSwamp    = "swamp"
	TerrainDesert   = "desert"
	TerrainCoastal  = "coastal"
	TerrainOutdoor  = "outdoor"
	ClimateCold     = "cold"
	ClimateArid     = "arid"
	// TerrainMountainous, TerrainForest, ClimateTropical already defined
)

// Economic status
const (
	EconomicPoor = "poor"
)

// Political/Social types
const (
	ApproachDiplomatic = "diplomatic"
	ApproachMilitary   = "military"
	ActionDiplomacy    = "diplomacy"
	ActionTrade        = "trade"
	AspectSocialStructure = "social_structure"
	AspectCustoms        = "customs"
)

// Additional combat outcomes
const (
	OutcomeVictory         = "victory"
	OutcomeDecisiveVictory = "decisive_victory"
	OutcomeHit             = "hit"
	OutcomeKillingBlow     = "killing_blow"
	// OutcomeCostlyVictory, OutcomeDefeat already defined
	// RelationNeutral can be used for neutral outcome
)

// More dice types
const (
	DiceD4  = "1d4"
	DiceD6  = "1d6"
	DiceD10 = "1d10"
	DiceD20 = "1d20"
	// DiceD8, DiceD12, DiceD100 already defined
)

// Special constants
const (
	MockProvider         = "mock"
	FormatJSON           = "json"
	CategoryAbility      = "ability"
	DefaultDamageFormula = "1d6"
	EncounterTypeCombat  = "combat"
	// TrueString already defined in abilities.go
)
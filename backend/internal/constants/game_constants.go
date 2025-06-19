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
	ClassWarlock   = "warlock"
	ClassRogue     = "rogue"
)

// Size categories
const (
	SizeSmall  = "small"
	SizeMedium = "medium"
	SizeLarge  = "large"
)

// Rarity levels
const (
	RarityCommon   = "common"
	RarityUncommon = "uncommon"
	RarityRare     = "rare"
	RarityVeryRare = "very_rare"
	RarityLegendary = "legendary"
)

// Action and ability types
const (
	ActionAttack         = "attack"
	ActionHeal           = "heal"
	ActionHit            = "hit"
	ActionCritical       = "critical"
	ActionKillingBlow    = "killing_blow"
	ActionCombat         = "combat"
	ActionAbility        = "ability"
	ActionTypeSpell      = "spell"
	ActionTypeAbility    = "ability"
	ActionTypeBonusAction = "bonus_action"
	ActionTypeRetreat    = "retreat"
	// ActionSpell can use CategorySpell from strings.go
)

// Difficulty levels
const (
	DifficultyEasy   = "easy"
	DifficultyMedium = "medium"
	DifficultyHard   = "hard"
	DifficultyDeadly = "deadly"
)

// Terrain and climate types
const (
	TerrainSwamp       = "swamp"
	TerrainDesert      = "desert"
	TerrainCoastal     = "coastal"
	TerrainOutdoor     = "outdoor"
	TerrainMountainous = "mountainous"
	TerrainForest      = "forest"
	ClimateCold        = "cold"
	ClimateArid        = "arid"
	ClimateTropical    = "tropical"
)

// Map types
const (
	MapTypeDungeon = "dungeon"
	MapTypeWilderness = "wilderness"
	MapTypeUrban = "urban"
)

// Economic status
const (
	EconomicPoor = "poor"
)

// Political/Social types
const (
	ApproachDiplomatic    = "diplomatic"
	ApproachMilitary      = "military"
	ActionDiplomacy       = "diplomacy"
	ActionTrade           = "trade"
	AspectSocialStructure = "social_structure"
	AspectCustoms         = "customs"
)

// Additional combat outcomes
const (
	OutcomeVictory         = "victory"
	OutcomeDecisiveVictory = "decisive_victory"
	OutcomeHit             = "hit"
	OutcomeKillingBlow     = "killing_blow"
	OutcomeCostlyVictory   = "costly_victory"
	OutcomeDefeat          = "defeat"
	// RelationNeutral can be used for neutral outcome
)

// More dice types
const (
	DiceD4   = "1d4"
	DiceD6   = "1d6"
	DiceD8   = "1d8"
	DiceD10  = "1d10"
	DiceD12  = "1d12"
	DiceD20  = "1d20"
	DiceD100 = "1d100"
)

// Special constants
const (
	MockProvider         = "mock"
	FormatJSON           = "json"
	CategoryAbility      = "ability"
	CategorySpell        = "spell"
	DefaultDamageFormula = "1d6"
	EncounterTypeCombat  = "combat"
	// TrueString already defined in abilities.go
)

// Operator constants
const (
	OperatorEquals      = "equals"
	OperatorContains    = "contains"
	OperatorGreaterThan = "greater_than"
	OperatorLessThan    = "less_than"
	OperatorIn          = "in"
)

// Time constants
const (
	TimeNight = "night"
	TimeDay   = "day"
	TimeDawn  = "dawn"
	TimeDusk  = "dusk"
)

// Encounter status constants
const (
	EncounterStatusPlanned    = "planned"
	EncounterStatusInProgress = "in_progress"
	EncounterStatusCompleted  = "completed"
)

// Objective constants
const (
	ObjectiveDefeatAll = "defeat_all"
	ObjectiveSurvive   = "survive"
	ObjectiveReach     = "reach"
	ObjectiveProtect   = "protect"
)

// Relation constants
const (
	RelationNeutral = "neutral"
	RelationAlly    = "ally"
	RelationEnemy   = "enemy"
)

// Culture values
const (
	CultureValues     = "values"
	CultureInfluence  = "cultural_influence"
	EventConflict     = "conflict"
)

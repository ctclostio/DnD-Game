package models

import (
	"time"

	"github.com/google/uuid"
)

// AINPC represents an AI-generated NPC with personality and dialogue
type AINPC struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	GameSessionID      uuid.UUID              `json:"gameSessionId" db:"game_session_id"`
	Name               string                 `json:"name" db:"name"`
	Race               string                 `json:"race" db:"race"`
	Occupation         string                 `json:"occupation" db:"occupation"`
	PersonalityTraits  []string               `json:"personalityTraits" db:"personality_traits"`
	Appearance         string                 `json:"appearance" db:"appearance"`
	VoiceDescription   string                 `json:"voiceDescription" db:"voice_description"`
	Motivations        string                 `json:"motivations" db:"motivations"`
	Secrets            string                 `json:"secrets" db:"secrets"`
	DialogueStyle      string                 `json:"dialogueStyle" db:"dialogue_style"`
	RelationshipToParty string                `json:"relationshipToParty" db:"relationship_to_party"`
	StatBlock          map[string]interface{} `json:"statBlock,omitempty" db:"stat_block"`
	GeneratedDialogue  []DialogueEntry        `json:"generatedDialogue" db:"generated_dialogue"`
	CreatedBy          uuid.UUID              `json:"createdBy" db:"created_by"`
	IsRecurring        bool                   `json:"isRecurring" db:"is_recurring"`
	LastSeenSession    *uuid.UUID             `json:"lastSeenSession,omitempty" db:"last_seen_session"`
	Notes              string                 `json:"notes" db:"notes"`
	CreatedAt          time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time              `json:"updatedAt" db:"updated_at"`
}

// DialogueEntry represents a single piece of dialogue from an NPC
type DialogueEntry struct {
	Context   string    `json:"context"`
	Dialogue  string    `json:"dialogue"`
	Timestamp time.Time `json:"timestamp"`
}

// AILocation represents an AI-generated location
type AILocation struct {
	ID                  uuid.UUID      `json:"id" db:"id"`
	GameSessionID       uuid.UUID      `json:"gameSessionId" db:"game_session_id"`
	Name                string         `json:"name" db:"name"`
	Type                string         `json:"type" db:"type"`
	Description         string         `json:"description" db:"description"`
	Atmosphere          string         `json:"atmosphere" db:"atmosphere"`
	NotableFeatures     []string       `json:"notableFeatures" db:"notable_features"`
	NPCsPresent         []uuid.UUID    `json:"npcsPresent" db:"npcs_present"`
	AvailableActions    []string       `json:"availableActions" db:"available_actions"`
	SecretsAndHidden    []SecretDetail `json:"secretsAndHidden" db:"secrets_and_hidden"`
	EnvironmentalEffects string        `json:"environmentalEffects" db:"environmental_effects"`
	CreatedBy           uuid.UUID      `json:"createdBy" db:"created_by"`
	ParentLocationID    *uuid.UUID     `json:"parentLocationId,omitempty" db:"parent_location_id"`
	IsDiscovered        bool           `json:"isDiscovered" db:"is_discovered"`
	CreatedAt           time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time      `json:"updatedAt" db:"updated_at"`
}

// SecretDetail represents a hidden element in a location
type SecretDetail struct {
	Description   string `json:"description"`
	DiscoveryDC   int    `json:"discoveryDC"`
	DiscoveryHint string `json:"discoveryHint"`
}

// AINarration represents combat narration or dramatic moments
type AINarration struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	GameSessionID  uuid.UUID              `json:"gameSessionId" db:"game_session_id"`
	Type           string                 `json:"type" db:"type"`
	Context        map[string]interface{} `json:"context" db:"context"`
	Narration      string                 `json:"narration" db:"narration"`
	IntensityLevel int                    `json:"intensityLevel" db:"intensity_level"`
	Tags           []string               `json:"tags" db:"tags"`
	CreatedBy      uuid.UUID              `json:"createdBy" db:"created_by"`
	UsedCount      int                    `json:"usedCount" db:"used_count"`
	CreatedAt      time.Time              `json:"createdAt" db:"created_at"`
}

// AIStoryElement represents plot twists and story hooks
type AIStoryElement struct {
	ID                 uuid.UUID      `json:"id" db:"id"`
	GameSessionID      uuid.UUID      `json:"gameSessionId" db:"game_session_id"`
	Type               string         `json:"type" db:"type"`
	Title              string         `json:"title" db:"title"`
	Description        string         `json:"description" db:"description"`
	Context            map[string]interface{} `json:"context" db:"context"`
	ImpactLevel        string         `json:"impactLevel" db:"impact_level"`
	SuggestedTiming    string         `json:"suggestedTiming" db:"suggested_timing"`
	Prerequisites      []string       `json:"prerequisites" db:"prerequisites"`
	Consequences       []string       `json:"consequences" db:"consequences"`
	ForeshadowingHints []string       `json:"foreshadowingHints" db:"foreshadowing_hints"`
	CreatedBy          uuid.UUID      `json:"createdBy" db:"created_by"`
	Used               bool           `json:"used" db:"used"`
	UsedAt             *time.Time     `json:"usedAt,omitempty" db:"used_at"`
	CreatedAt          time.Time      `json:"createdAt" db:"created_at"`
}

// AIEnvironmentalHazard represents environmental challenges
type AIEnvironmentalHazard struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	GameSessionID    uuid.UUID              `json:"gameSessionId" db:"game_session_id"`
	LocationID       *uuid.UUID             `json:"locationId,omitempty" db:"location_id"`
	Name             string                 `json:"name" db:"name"`
	Description      string                 `json:"description" db:"description"`
	TriggerCondition string                 `json:"triggerCondition" db:"trigger_condition"`
	EffectDescription string                `json:"effectDescription" db:"effect_description"`
	MechanicalEffects map[string]interface{} `json:"mechanicalEffects" db:"mechanical_effects"`
	DifficultyClass  int                    `json:"difficultyClass" db:"difficulty_class"`
	DamageFormula    string                 `json:"damageFormula" db:"damage_formula"`
	AvoidanceHints   string                 `json:"avoidanceHints" db:"avoidance_hints"`
	IsTrap           bool                   `json:"isTrap" db:"is_trap"`
	IsNatural        bool                   `json:"isNatural" db:"is_natural"`
	ResetCondition   string                 `json:"resetCondition" db:"reset_condition"`
	CreatedBy        uuid.UUID              `json:"createdBy" db:"created_by"`
	IsActive         bool                   `json:"isActive" db:"is_active"`
	TriggeredCount   int                    `json:"triggeredCount" db:"triggered_count"`
	CreatedAt        time.Time              `json:"createdAt" db:"created_at"`
}

// DMAssistantHistory tracks all DM assistant interactions
type DMAssistantHistory struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	GameSessionID  uuid.UUID              `json:"gameSessionId" db:"game_session_id"`
	UserID         uuid.UUID              `json:"userId" db:"user_id"`
	RequestType    string                 `json:"requestType" db:"request_type"`
	RequestContext map[string]interface{} `json:"requestContext" db:"request_context"`
	Prompt         string                 `json:"prompt" db:"prompt"`
	Response       string                 `json:"response" db:"response"`
	Feedback       string                 `json:"feedback" db:"feedback"`
	CreatedAt      time.Time              `json:"createdAt" db:"created_at"`
}

// Request types for DM Assistant
const (
	RequestTypeNPCDialogue        = "npc_dialogue"
	RequestTypeLocationDesc       = "location_description"
	RequestTypeCombatNarration    = "combat_narration"
	RequestTypeDeathDescription   = "death_description"
	RequestTypePlotTwist          = "plot_twist"
	RequestTypeEnvironmentalHazard = "environmental_hazard"
	RequestTypeStoryHook          = "story_hook"
)

// Location types
const (
	LocationTypeTavern     = "tavern"
	LocationTypeDungeon    = "dungeon"
	LocationTypeShop       = "shop"
	LocationTypeWilderness = "wilderness"
	LocationTypeCity       = "city"
	LocationTypeTemple     = "temple"
	LocationTypeCastle     = "castle"
)

// Narration types
const (
	NarrationTypeCombatHit      = "combat_hit"
	NarrationTypeCombatMiss     = "combat_miss"
	NarrationTypeCombatCritical = "combat_critical"
	NarrationTypeDeath          = "death"
	NarrationTypeDramatic       = "dramatic_moment"
)

// Story element types
const (
	StoryElementPlotTwist    = "plot_twist"
	StoryElementStoryHook    = "story_hook"
	StoryElementRevelation   = "revelation"
	StoryElementComplication = "complication"
)

// Impact levels
const (
	ImpactLevelMinor           = "minor"
	ImpactLevelModerate        = "moderate"
	ImpactLevelMajor           = "major"
	ImpactLevelCampaignChanging = "campaign-changing"
)

// DMAssistantRequest represents a request to the DM Assistant
type DMAssistantRequest struct {
	Type           string                 `json:"type" validate:"required"`
	GameSessionID  string                 `json:"gameSessionId" validate:"required"`
	Context        map[string]interface{} `json:"context"`
	Parameters     map[string]interface{} `json:"parameters"`
	StreamResponse bool                   `json:"streamResponse"`
}

// NPCDialogueRequest for generating NPC dialogue
type NPCDialogueRequest struct {
	NPCName        string   `json:"npcName"`
	NPCPersonality []string `json:"npcPersonality"`
	DialogueStyle  string   `json:"dialogueStyle"`
	Situation      string   `json:"situation"`
	PlayerInput    string   `json:"playerInput"`
	PreviousContext string  `json:"previousContext"`
}

// LocationDescriptionRequest for generating location descriptions
type LocationDescriptionRequest struct {
	LocationType    string   `json:"locationType"`
	LocationName    string   `json:"locationName"`
	Atmosphere      string   `json:"atmosphere"`
	SpecialFeatures []string `json:"specialFeatures"`
	TimeOfDay       string   `json:"timeOfDay"`
	Weather         string   `json:"weather"`
}

// CombatNarrationRequest for combat descriptions
type CombatNarrationRequest struct {
	AttackerName   string `json:"attackerName"`
	TargetName     string `json:"targetName"`
	ActionType     string `json:"actionType"`
	WeaponOrSpell  string `json:"weaponOrSpell"`
	Damage         int    `json:"damage"`
	IsHit          bool   `json:"isHit"`
	IsCritical     bool   `json:"isCritical"`
	TargetHP       int    `json:"targetHP"`
	TargetMaxHP    int    `json:"targetMaxHP"`
}

// EnvironmentRequest for generating environment descriptions
type EnvironmentRequest struct {
	Location    string   `json:"location"`
	Type        string   `json:"type"`
	Atmosphere  string   `json:"atmosphere"`
	Features    []string `json:"features"`
	TimeOfDay   string   `json:"timeOfDay,omitempty"`
	Weather     string   `json:"weather,omitempty"`
	KeyFeatures []string `json:"keyFeatures,omitempty"`
}

// EnvironmentDescription represents generated environment details
type EnvironmentDescription struct {
	Description       string          `json:"description"`
	SensoryDetails    SensoryDetails  `json:"sensoryDetails"`
	NotableFeatures   []string        `json:"notableFeatures"`
	PossibleActions   []string        `json:"possibleActions"`
	HiddenElements    []string        `json:"hiddenElements"`
	PointsOfInterest  []string        `json:"pointsOfInterest,omitempty"`
	PotentialHazards  []string        `json:"potentialHazards,omitempty"`
}

// SensoryDetails represents the sensory aspects of an environment
type SensoryDetails struct {
	Sight string `json:"sight"`
	Sound string `json:"sound"`
	Smell string `json:"smell"`
	Touch string `json:"touch"`
	Taste string `json:"taste"`
}

// PlotHookRequest for generating plot hooks
type PlotHookRequest struct {
	CurrentSituation string   `json:"currentSituation"`
	PlayerGoals      []string `json:"playerGoals"`
	WorldEvents      []string `json:"worldEvents"`
	NPCMotivations   []string `json:"npcMotivations"`
	Theme            string   `json:"theme,omitempty"`
	PartyLevel       int      `json:"partyLevel,omitempty"`
	Setting          string   `json:"setting,omitempty"`
	PartyComposition []string `json:"partyComposition,omitempty"`
}

// PlotHook represents a generated plot hook
type PlotHook struct {
	Title             string    `json:"title"`
	Hook              string    `json:"hook"`
	Background        string    `json:"background,omitempty"`
	KeyNPCs           []PlotNPC `json:"keyNPCs,omitempty"`
	InitialClues      []string  `json:"initialClues,omitempty"`
	PotentialRewards  []string  `json:"potentialRewards,omitempty"`
	Escalation        string    `json:"escalation,omitempty"`
	NPCInvolved       []string  `json:"npcInvolved"`
	Urgency           string    `json:"urgency"`
	Reward            string    `json:"reward"`
	Consequences      string    `json:"consequences"`
}

// PlotNPC represents an NPC involved in a plot hook
type PlotNPC struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	Motivation string `json:"motivation"`
}

// RulingRequest for DM rulings
type RulingRequest struct {
	Situation      string `json:"situation"`
	RuleInQuestion string `json:"ruleInQuestion"`
	PlayerAction   string `json:"playerAction"`
	Context        string `json:"context"`
	RulesContext   string `json:"rulesContext,omitempty"`
	PlayerIntent   string `json:"playerIntent,omitempty"`
}

// RulingSuggestion represents a suggested ruling
type RulingSuggestion struct {
	Ruling                string   `json:"ruling"`
	Reasoning             string   `json:"reasoning"`
	RuleReference         string   `json:"ruleReference"`
	Alternative           string   `json:"alternative"`
	Precedent             string   `json:"precedent,omitempty"`
	Alternatives          []string `json:"alternatives,omitempty"`
	BalanceConsiderations string   `json:"balanceConsiderations,omitempty"`
}

// TreasureRequest for generating treasure
type TreasureRequest struct {
	ChallengeRating int    `json:"challengeRating"`
	TreasureType    string `json:"treasureType"`
	PartyLevel      int    `json:"partyLevel"`
	Context         string `json:"context"`
	PartySize       int    `json:"partySize,omitempty"`
}

// TreasureHoard represents generated treasure
type TreasureHoard struct {
	Currency    map[string]int   `json:"currency"`
	Items       []string         `json:"items"`
	MagicItems  []string         `json:"magicItems"`
	SpecialItems []string        `json:"specialItems"`
	TotalValue  int              `json:"totalValue"`
	Coins       CoinageBreakdown `json:"coins,omitempty"`
	Gems        []Gem            `json:"gems,omitempty"`
	ArtObjects  []ArtObject      `json:"artObjects,omitempty"`
	MagicItemDetails []MagicItem `json:"magicItemDetails,omitempty"`
}

// CoinageBreakdown represents the breakdown of coins in treasure
type CoinageBreakdown struct {
	Copper   int `json:"copper"`
	Silver   int `json:"silver"`
	Gold     int `json:"gold"`
	Platinum int `json:"platinum"`
}

// Gem represents a gem found in treasure
type Gem struct {
	Name        string `json:"name"`
	Value       int    `json:"value"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description"`
}

// ArtObject represents an art object found in treasure
type ArtObject struct {
	Name        string `json:"name"`
	Value       int    `json:"value"`
	Description string `json:"description"`
}

// MagicItem represents a magic item with detailed properties
type MagicItem struct {
	Name        string   `json:"name"`
	Rarity      string   `json:"rarity"`
	Description string   `json:"description"`
	Properties  []string `json:"properties"`
}
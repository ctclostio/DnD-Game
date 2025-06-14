package models

import (
	"time"
)

// Encounter represents a planned or active encounter in the game.
type Encounter struct {
	ID                     string                 `json:"id" db:"id"`
	GameSessionID          string                 `json:"gameSessionId" db:"game_session_id"`
	CreatedBy              string                 `json:"createdBy" db:"created_by"`
	Name                   string                 `json:"name" db:"name"`
	Description            string                 `json:"description" db:"description"`
	Location               string                 `json:"location" db:"location"`
	EncounterType          string                 `json:"encounterType" db:"encounter_type"`
	Difficulty             string                 `json:"difficulty" db:"difficulty"`
	ChallengeRating        float64                `json:"challengeRating" db:"challenge_rating"`
	NarrativeContext       string                 `json:"narrativeContext" db:"narrative_context"`
	EnvironmentalFeatures  []string               `json:"environmentalFeatures" db:"environmental_features"`
	StoryHooks             []string               `json:"storyHooks" db:"story_hooks"`
	PartyLevel             int                    `json:"partyLevel" db:"party_level"`
	PartySize              int                    `json:"partySize" db:"party_size"`
	PartyComposition       map[string]interface{} `json:"partyComposition" db:"party_composition"`
	Enemies                []EncounterEnemy       `json:"enemies" db:"enemies"`
	TotalXP                int                    `json:"totalXp" db:"total_xp"`
	AdjustedXP             int                    `json:"adjustedXp" db:"adjusted_xp"`
	EnemyTactics           *TacticalInfo          `json:"enemyTactics,omitempty" db:"enemy_tactics"`
	EnvironmentalHazards   []EnvironmentalHazard  `json:"environmentalHazards,omitempty" db:"environmental_hazards"`
	TerrainFeatures        []TerrainFeature       `json:"terrainFeatures,omitempty" db:"terrain_features"`
	SocialSolutions        []Solution             `json:"socialSolutions,omitempty" db:"social_solutions"`
	StealthOptions         []Solution             `json:"stealthOptions,omitempty" db:"stealth_options"`
	EnvironmentalSolutions []Solution             `json:"environmentalSolutions,omitempty" db:"environmental_solutions"`
	ScalingOptions         *ScalingOptions        `json:"scalingOptions,omitempty" db:"scaling_options"`
	ReinforcementWaves     []ReinforcementWave    `json:"reinforcementWaves,omitempty" db:"reinforcement_waves"`
	EscapeRoutes           []EscapeRoute          `json:"escapeRoutes,omitempty" db:"escape_routes"`
	Status                 string                 `json:"status" db:"status"`
	StartedAt              *time.Time             `json:"startedAt,omitempty" db:"started_at"`
	CompletedAt            *time.Time             `json:"completedAt,omitempty" db:"completed_at"`
	Outcome                string                 `json:"outcome,omitempty" db:"outcome"`
	CreatedAt              time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt              time.Time              `json:"updatedAt" db:"updated_at"`
}

// EncounterEnemy represents an enemy in an encounter.
type EncounterEnemy struct {
	ID                string                 `json:"id" db:"id"`
	EncounterID       string                 `json:"encounterId" db:"encounter_id"`
	NPCID             *string                `json:"npcId,omitempty" db:"npc_id"`
	Name              string                 `json:"name" db:"name"`
	Type              string                 `json:"type" db:"type"`
	Size              string                 `json:"size" db:"size"`
	ChallengeRating   float64                `json:"challengeRating" db:"challenge_rating"`
	HitPoints         int                    `json:"hitPoints" db:"hit_points"`
	ArmorClass        int                    `json:"armorClass" db:"armor_class"`
	Stats             map[string]interface{} `json:"stats" db:"stats"`
	Abilities         []Ability              `json:"abilities" db:"abilities"`
	Actions           []Action               `json:"actions" db:"actions"`
	LegendaryActions  []Action               `json:"legendaryActions,omitempty" db:"legendary_actions"`
	PersonalityTraits []string               `json:"personalityTraits" db:"personality_traits"`
	Ideal             string                 `json:"ideal" db:"ideal"`
	Bond              string                 `json:"bond" db:"bond"`
	Flaw              string                 `json:"flaw" db:"flaw"`
	Tactics           string                 `json:"tactics" db:"tactics"`
	MoraleThreshold   int                    `json:"moraleThreshold" db:"morale_threshold"`
	InitialPosition   *GridPosition          `json:"initialPosition,omitempty" db:"initial_position"`
	CurrentPosition   *GridPosition          `json:"currentPosition,omitempty" db:"current_position"`
	Conditions        []string               `json:"conditions" db:"conditions"`
	IsAlive           bool                   `json:"isAlive" db:"is_alive"`
	Fled              bool                   `json:"fled" db:"fled"`
	Quantity          int                    `json:"quantity"`
	Role              string                 `json:"role"` // tank, damage, support, etc.
}

// TacticalInfo contains AI-generated tactical suggestions.
type TacticalInfo struct {
	GeneralStrategy   string            `json:"generalStrategy"`
	PriorityTargets   []string          `json:"priorityTargets"`
	Positioning       string            `json:"positioning"`
	CombatPhases      []CombatPhase     `json:"combatPhases"`
	RetreatConditions string            `json:"retreatConditions"`
	SpecialTactics    map[string]string `json:"specialTactics"`
}

// CombatPhase represents different phases of combat with different tactics.
type CombatPhase struct {
	Name      string   `json:"name"`
	Trigger   string   `json:"trigger"`
	Tactics   string   `json:"tactics"`
	Abilities []string `json:"abilities"`
}

// EnvironmentalHazard represents a hazard in the encounter.
type EnvironmentalHazard struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Trigger     string `json:"trigger"`
	Effect      string `json:"effect"`
	SaveDC      int    `json:"saveDc"`
	Damage      string `json:"damage"`
}

// TerrainFeature represents terrain that affects combat.
type TerrainFeature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Effect      string `json:"effect"`
	Location    string `json:"location"`
}

// Solution represents a non-combat solution option.
type Solution struct {
	Method       string   `json:"method"`
	Description  string   `json:"description"`
	Requirements []string `json:"requirements"`
	DC           int      `json:"dc"`
	Consequences string   `json:"consequences"`
}

// ScalingOptions for dynamic difficulty.
type ScalingOptions struct {
	Easy   ScalingAdjustment `json:"easy"`
	Medium ScalingAdjustment `json:"medium"`
	Hard   ScalingAdjustment `json:"hard"`
	Deadly ScalingAdjustment `json:"deadly"`
}

// ScalingAdjustment represents how to adjust an encounter.
type ScalingAdjustment struct {
	AddEnemies     []string `json:"addEnemies,omitempty"`
	RemoveEnemies  []string `json:"removeEnemies,omitempty"`
	HPModifier     int      `json:"hpModifier,omitempty"`
	DamageModifier int      `json:"damageModifier,omitempty"`
	AddHazards     []string `json:"addHazards,omitempty"`
	AddTerrain     []string `json:"addTerrain,omitempty"`
	AddObjectives  []string `json:"addObjectives,omitempty"`
}

// ReinforcementWave for mid-combat additions.
type ReinforcementWave struct {
	Round        int              `json:"round"`
	Trigger      string           `json:"trigger"`
	Enemies      []EncounterEnemy `json:"enemies"`
	Entrance     string           `json:"entrance"`
	Announcement string           `json:"announcement"`
}

// EscapeRoute for tactical retreats.
type EscapeRoute struct {
	Direction   string `json:"direction"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	Consequence string `json:"consequence"`
}

// Note: Position type is defined in combat.go

// Ability represents a special ability.
type Ability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Recharge    string `json:"recharge,omitempty"`
}

// Action represents something a creature can do.
type Action struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AttackBonus int    `json:"attackBonus,omitempty"`
	Damage      string `json:"damage,omitempty"`
	SaveDC      int    `json:"saveDc,omitempty"`
	SaveType    string `json:"saveType,omitempty"`
}

// EncounterObjective represents goals for the encounter.
type EncounterObjective struct {
	ID                string                 `json:"id" db:"id"`
	EncounterID       string                 `json:"encounterId" db:"encounter_id"`
	Type              string                 `json:"type" db:"type"`
	Description       string                 `json:"description" db:"description"`
	SuccessConditions map[string]interface{} `json:"successConditions" db:"success_conditions"`
	FailureConditions map[string]interface{} `json:"failureConditions,omitempty" db:"failure_conditions"`
	XPReward          int                    `json:"xpReward" db:"xp_reward"`
	GoldReward        int                    `json:"goldReward" db:"gold_reward"`
	ItemRewards       []ItemReward           `json:"itemRewards,omitempty" db:"item_rewards"`
	StoryRewards      []string               `json:"storyRewards,omitempty" db:"story_rewards"`
	IsCompleted       bool                   `json:"isCompleted" db:"is_completed"`
	IsFailed          bool                   `json:"isFailed" db:"is_failed"`
	CompletedAt       *time.Time             `json:"completedAt,omitempty" db:"completed_at"`
	CreatedAt         time.Time              `json:"createdAt" db:"created_at"`
}

// ItemReward represents an item reward from an objective.
type ItemReward struct {
	ItemID      string `json:"itemId"`
	ItemName    string `json:"itemName"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description"`
}

// EncounterEvent tracks what happens during an encounter.
type EncounterEvent struct {
	ID               string                 `json:"id" db:"id"`
	EncounterID      string                 `json:"encounterId" db:"encounter_id"`
	RoundNumber      int                    `json:"roundNumber" db:"round_number"`
	EventType        string                 `json:"eventType" db:"event_type"`
	ActorType        string                 `json:"actorType" db:"actor_type"`
	ActorID          *string                `json:"actorId,omitempty" db:"actor_id"`
	ActorName        string                 `json:"actorName" db:"actor_name"`
	Description      string                 `json:"description" db:"description"`
	MechanicalEffect map[string]interface{} `json:"mechanicalEffect,omitempty" db:"mechanical_effect"`
	AISuggestion     string                 `json:"aiSuggestion,omitempty" db:"ai_suggestion"`
	SuggestionUsed   bool                   `json:"suggestionUsed" db:"suggestion_used"`
	CreatedAt        time.Time              `json:"createdAt" db:"created_at"`
}

// EncounterTemplate for reusable encounters.
type EncounterTemplate struct {
	ID                    string                 `json:"id" db:"id"`
	CreatedBy             *string                `json:"createdBy,omitempty" db:"created_by"`
	Name                  string                 `json:"name" db:"name"`
	Description           string                 `json:"description" db:"description"`
	Tags                  []string               `json:"tags" db:"tags"`
	EncounterType         string                 `json:"encounterType" db:"encounter_type"`
	MinLevel              int                    `json:"minLevel" db:"min_level"`
	MaxLevel              int                    `json:"maxLevel" db:"max_level"`
	EnvironmentTypes      []string               `json:"environmentTypes" db:"environment_types"`
	EnemyGroups           []EnemyGroup           `json:"enemyGroups" db:"enemy_groups"`
	ScalingFormula        map[string]interface{} `json:"scalingFormula" db:"scaling_formula"`
	TacticalNotes         string                 `json:"tacticalNotes" db:"tactical_notes"`
	EnvironmentalFeatures map[string]interface{} `json:"environmentalFeatures" db:"environmental_features"`
	ObjectiveOptions      []ObjectiveOption      `json:"objectiveOptions" db:"objective_options"`
	IsPublic              bool                   `json:"isPublic" db:"is_public"`
	TimesUsed             int                    `json:"timesUsed" db:"times_used"`
	Rating                float64                `json:"rating" db:"rating"`
	CreatedAt             time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time              `json:"updatedAt" db:"updated_at"`
}

// EnemyGroup for templates.
type EnemyGroup struct {
	Name         string `json:"name"`
	MinQuantity  int    `json:"minQuantity"`
	MaxQuantity  int    `json:"maxQuantity"`
	Role         string `json:"role"`
	ScalingNotes string `json:"scalingNotes"`
}

// ObjectiveOption for templates.
type ObjectiveOption struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
}

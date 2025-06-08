package models

import (
	"time"

	"github.com/google/uuid"
)

// CombatAnalytics represents detailed combat statistics
type CombatAnalytics struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	CombatID          uuid.UUID  `json:"combat_id" db:"combat_id"`
	GameSessionID     uuid.UUID  `json:"game_session_id" db:"game_session_id"`
	CombatDuration    int        `json:"combat_duration" db:"combat_duration"`
	TotalDamageDealt  int        `json:"total_damage_dealt" db:"total_damage_dealt"`
	TotalHealingDone  int        `json:"total_healing_done" db:"total_healing_done"`
	KillingBlows      JSONB      `json:"killing_blows" db:"killing_blows"`
	CombatSummary     JSONB      `json:"combat_summary" db:"combat_summary"`
	MVPID             string     `json:"mvp_id" db:"mvp_id"`
	MVPType           string     `json:"mvp_type" db:"mvp_type"`
	TacticalRating    int        `json:"tactical_rating" db:"tactical_rating"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// CombatSummary represents a summary of combat analytics
type CombatSummary struct {
	CombatID         uuid.UUID              `json:"combat_id"`
	Duration         int                    `json:"duration"`
	TotalRounds      int                    `json:"total_rounds"`
	TotalDamage      int                    `json:"total_damage"`
	TotalHealing     int                    `json:"total_healing"`
	ParticipantStats []CombatantAnalytics   `json:"participant_stats"`
	MVP              *CombatantAnalytics    `json:"mvp,omitempty"`
	Highlights       []string               `json:"highlights"`
}

// CombatantAnalytics represents individual performance metrics
type CombatantAnalytics struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	CombatAnalyticsID    uuid.UUID  `json:"combat_analytics_id" db:"combat_analytics_id"`
	CombatantID          string     `json:"combatant_id" db:"combatant_id"`
	CombatantType        string     `json:"combatant_type" db:"combatant_type"`
	CombatantName        string     `json:"combatant_name" db:"combatant_name"`
	DamageDealt          int        `json:"damage_dealt" db:"damage_dealt"`
	DamageTaken          int        `json:"damage_taken" db:"damage_taken"`
	HealingDone          int        `json:"healing_done" db:"healing_done"`
	HealingReceived      int        `json:"healing_received" db:"healing_received"`
	AttacksMade          int        `json:"attacks_made" db:"attacks_made"`
	AttacksHit           int        `json:"attacks_hit" db:"attacks_hit"`
	AttacksMissed        int        `json:"attacks_missed" db:"attacks_missed"`
	CriticalHits         int        `json:"critical_hits" db:"critical_hits"`
	CriticalMisses       int        `json:"critical_misses" db:"critical_misses"`
	SavesMade            int        `json:"saves_made" db:"saves_made"`
	SavesFailed          int        `json:"saves_failed" db:"saves_failed"`
	RoundsSurvived       int        `json:"rounds_survived" db:"rounds_survived"`
	FinalHP              int        `json:"final_hp" db:"final_hp"`
	ConditionsSuffered   JSONB      `json:"conditions_suffered" db:"conditions_suffered"`
	AbilitiesUsed        JSONB      `json:"abilities_used" db:"abilities_used"`
	TacticalDecisions    JSONB      `json:"tactical_decisions" db:"tactical_decisions"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}

// AutoCombatResolution represents a quick combat resolution
type AutoCombatResolution struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	GameSessionID        uuid.UUID  `json:"game_session_id" db:"game_session_id"`
	EncounterDifficulty  string     `json:"encounter_difficulty" db:"encounter_difficulty"`
	PartyComposition     JSONB      `json:"party_composition" db:"party_composition"`
	EnemyComposition     JSONB      `json:"enemy_composition" db:"enemy_composition"`
	ResolutionType       string     `json:"resolution_type" db:"resolution_type"`
	Outcome              string     `json:"outcome" db:"outcome"`
	RoundsSimulated      int        `json:"rounds_simulated" db:"rounds_simulated"`
	PartyResourcesUsed   JSONB      `json:"party_resources_used" db:"party_resources_used"`
	LootGenerated        JSONB      `json:"loot_generated" db:"loot_generated"`
	ExperienceAwarded    int        `json:"experience_awarded" db:"experience_awarded"`
	NarrativeSummary     string     `json:"narrative_summary" db:"narrative_summary"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}

// BattleMap represents a generated tactical map
type BattleMap struct {
	ID                  uuid.UUID   `json:"id" db:"id"`
	CombatID            *uuid.UUID  `json:"combat_id,omitempty" db:"combat_id"`
	GameSessionID       uuid.UUID   `json:"game_session_id" db:"game_session_id"`
	LocationDescription string      `json:"location_description" db:"location_description"`
	MapType             string      `json:"map_type" db:"map_type"`
	GridSizeX           int         `json:"grid_size_x" db:"grid_size_x"`
	GridSizeY           int         `json:"grid_size_y" db:"grid_size_y"`
	TerrainFeatures     JSONB       `json:"terrain_features" db:"terrain_features"`
	ObstaclePositions   JSONB       `json:"obstacle_positions" db:"obstacle_positions"`
	CoverPositions      JSONB       `json:"cover_positions" db:"cover_positions"`
	HazardZones         JSONB       `json:"hazard_zones" db:"hazard_zones"`
	SpawnPoints         JSONB       `json:"spawn_points" db:"spawn_points"`
	TacticalNotes       JSONB       `json:"tactical_notes" db:"tactical_notes"`
	VisualTheme         string      `json:"visual_theme" db:"visual_theme"`
	CreatedAt           time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at" db:"updated_at"`
}

// SmartInitiativeRule represents initiative bonuses and special rules
type SmartInitiativeRule struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	GameSessionID         uuid.UUID  `json:"game_session_id" db:"game_session_id"`
	EntityID              string     `json:"entity_id" db:"entity_id"`
	EntityType            string     `json:"entity_type" db:"entity_type"`
	BaseInitiativeBonus   int        `json:"base_initiative_bonus" db:"base_initiative_bonus"`
	AdvantageOnInitiative bool       `json:"advantage_on_initiative" db:"advantage_on_initiative"`
	AlertFeat             bool       `json:"alert_feat" db:"alert_feat"`
	SpecialRules          JSONB      `json:"special_rules" db:"special_rules"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// CombatActionLog represents detailed action tracking
type CombatActionLog struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	CombatID          uuid.UUID  `json:"combat_id" db:"combat_id"`
	RoundNumber       int        `json:"round_number" db:"round_number"`
	TurnNumber        int        `json:"turn_number" db:"turn_number"`
	ActorID           string     `json:"actor_id" db:"actor_id"`
	ActorType         string     `json:"actor_type" db:"actor_type"`
	ActionType        string     `json:"action_type" db:"action_type"`
	TargetID          *string    `json:"target_id,omitempty" db:"target_id"`
	TargetType        *string    `json:"target_type,omitempty" db:"target_type"`
	RollResults       JSONB      `json:"roll_results" db:"roll_results"`
	Outcome           string     `json:"outcome" db:"outcome"`
	DamageDealt       int        `json:"damage_dealt" db:"damage_dealt"`
	ConditionsApplied JSONB      `json:"conditions_applied" db:"conditions_applied"`
	ResourcesUsed     JSONB      `json:"resources_used" db:"resources_used"`
	PositionData      JSONB      `json:"position_data" db:"position_data"`
	Timestamp         time.Time  `json:"timestamp" db:"timestamp"`
}

// Request/Response types

type AutoResolveRequest struct {
	EncounterDifficulty string          `json:"encounter_difficulty" binding:"required"`
	EnemyTypes          []EnemyInfo     `json:"enemy_types" binding:"required"`
	TerrainType         string          `json:"terrain_type"`
	UseResources        bool            `json:"use_resources"` // Whether to use spell slots, etc.
}

type EnemyInfo struct {
	Name  string `json:"name"`
	CR    string `json:"cr"`
	Count int    `json:"count"`
}

type GenerateBattleMapRequest struct {
	LocationDescription string   `json:"location_description" binding:"required"`
	MapType             string   `json:"map_type"` // dungeon, outdoor, urban, special
	DesiredSize         string   `json:"desired_size"` // small, medium, large
	IncludeHazards      bool     `json:"include_hazards"`
	TerrainComplexity   string   `json:"terrain_complexity"` // simple, moderate, complex
}

type SmartInitiativeRequest struct {
	CombatID    uuid.UUID             `json:"combat_id" binding:"required"`
	Combatants  []InitiativeCombatant `json:"combatants" binding:"required"`
}

type InitiativeCombatant struct {
	ID               string `json:"id"`
	Type             string `json:"type"` // character or npc
	Name             string `json:"name"`
	DexterityModifier int   `json:"dexterity_modifier"`
}

type CombatAnalyticsReport struct {
	Analytics          *CombatAnalytics      `json:"analytics"`
	CombatantReports   []*CombatantReport    `json:"combatant_reports"`
	TacticalAnalysis   *TacticalAnalysis     `json:"tactical_analysis"`
	Recommendations    []string              `json:"recommendations"`
}

type CombatantReport struct {
	Analytics       *CombatantAnalytics `json:"analytics"`
	PerformanceRating string            `json:"performance_rating"` // excellent, good, fair, poor
	Highlights      []string           `json:"highlights"`
}

type TacticalAnalysis struct {
	PositioningScore    int      `json:"positioning_score"` // 1-10
	ResourceManagement  int      `json:"resource_management"` // 1-10
	TargetPrioritization int     `json:"target_prioritization"` // 1-10
	TeamworkScore       int      `json:"teamwork_score"` // 1-10
	MissedOpportunities []string `json:"missed_opportunities"`
}

// Battle Map structures

type BattleMapTerrainFeature struct {
	Type        string    `json:"type"` // wall, pillar, tree, water, etc.
	Position    GridPosition  `json:"position"`
	Size        Size      `json:"size"`
	Properties  []string  `json:"properties"` // blocks_movement, blocks_sight, difficult_terrain
}

// Note: GridPosition is defined in combat.go

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type HazardZone struct {
	Type        string     `json:"type"` // fire, acid, spike_pit, etc.
	Area        []GridPosition `json:"area"`
	DamageType  string     `json:"damage_type"`
	DamageDice  string     `json:"damage_dice"`
	SaveDC      int        `json:"save_dc"`
	SaveType    string     `json:"save_type"`
}

type TacticalNote struct {
	Position    GridPosition `json:"position"`
	Note        string   `json:"note"`
	Importance  string   `json:"importance"` // high, medium, low
}

// InitiativeEntry represents a combatant's initiative result
type InitiativeEntry struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Initiative int    `json:"initiative"`
	Roll       int    `json:"roll"`
	Bonus      int    `json:"bonus"`
}

// PlayerCombatStats represents a player's combat statistics across all sessions
type PlayerCombatStats struct {
	PlayerID         uuid.UUID `json:"player_id"`
	CharacterID      uuid.UUID `json:"character_id"`
	CharacterName    string    `json:"character_name"`
	TotalCombats     int       `json:"total_combats"`
	TotalDamageDealt int       `json:"total_damage_dealt"`
	TotalDamageTaken int       `json:"total_damage_taken"`
	TotalHealing     int       `json:"total_healing"`
	TotalKills       int       `json:"total_kills"`
	AverageAccuracy  float64   `json:"average_accuracy"`
	MVPCount         int       `json:"mvp_count"`
}

// SessionCombatStats represents combat statistics for a game session
type SessionCombatStats struct {
	SessionID        uuid.UUID `json:"session_id"`
	TotalCombats     int       `json:"total_combats"`
	AverageDuration  int       `json:"average_duration"`
	TotalDamageDealt int       `json:"total_damage_dealt"`
	TotalHealing     int       `json:"total_healing"`
	PlayerDeaths     int       `json:"player_deaths"`
	EnemyDeaths      int       `json:"enemy_deaths"`
}

// CombatTrends represents trends in combat over time
type CombatTrends struct {
	SessionID            uuid.UUID `json:"session_id"`
	AverageCombatLength  float64   `json:"average_combat_length"`
	DifficultyTrend      string    `json:"difficulty_trend"` // increasing, decreasing, stable
	PlayerPerformance    string    `json:"player_performance"` // improving, declining, stable
	MostEffectivePlayer  string    `json:"most_effective_player"`
	MostTargetedPlayer   string    `json:"most_targeted_player"`
	PopularStrategies    []string  `json:"popular_strategies"`
}
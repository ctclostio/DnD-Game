package models

import (
	"time"

	"github.com/google/uuid"
)

// StoryArc represents a narrative arc in the campaign
type StoryArc struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	GameSessionID   uuid.UUID       `json:"game_session_id" db:"game_session_id"`
	Title           string          `json:"title" db:"title"`
	Description     string          `json:"description" db:"description"`
	ArcType         string          `json:"arc_type" db:"arc_type"` // main_quest, side_quest, character_arc
	Status          string          `json:"status" db:"status"`       // active, completed, abandoned, foreshadowed
	ParentArcID     *uuid.UUID      `json:"parent_arc_id,omitempty" db:"parent_arc_id"`
	ImportanceLevel int             `json:"importance_level" db:"importance_level"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	Metadata        JSONB           `json:"metadata" db:"metadata"`
}

// SessionMemory represents the record of a game session
type SessionMemory struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	GameSessionID     uuid.UUID       `json:"game_session_id" db:"game_session_id"`
	SessionNumber     int             `json:"session_number" db:"session_number"`
	SessionDate       time.Time       `json:"session_date" db:"session_date"`
	RecapSummary      string          `json:"recap_summary" db:"recap_summary"`
	KeyEvents         JSONB           `json:"key_events" db:"key_events"`
	NPCsEncountered   JSONB           `json:"npcs_encountered" db:"npcs_encountered"`
	DecisionsMade     JSONB           `json:"decisions_made" db:"decisions_made"`
	ItemsAcquired     JSONB           `json:"items_acquired" db:"items_acquired"`
	LocationsVisited  JSONB           `json:"locations_visited" db:"locations_visited"`
	CombatEncounters  JSONB           `json:"combat_encounters" db:"combat_encounters"`
	PlotDevelopments  JSONB           `json:"plot_developments" db:"plot_developments"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// PlotThread represents an ongoing narrative thread
type PlotThread struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	GameSessionID        uuid.UUID       `json:"game_session_id" db:"game_session_id"`
	StoryArcID           *uuid.UUID      `json:"story_arc_id,omitempty" db:"story_arc_id"`
	ThreadType           string          `json:"thread_type" db:"thread_type"` // mystery, conflict, relationship, prophecy
	Title                string          `json:"title" db:"title"`
	Description          string          `json:"description" db:"description"`
	Status               string          `json:"status" db:"status"` // active, resolved, dormant, abandoned
	TensionLevel         int             `json:"tension_level" db:"tension_level"`
	IntroducedSession    *int            `json:"introduced_session,omitempty" db:"introduced_session"`
	ResolvedSession      *int            `json:"resolved_session,omitempty" db:"resolved_session"`
	RelatedNPCs          JSONB           `json:"related_npcs" db:"related_npcs"`
	RelatedLocations     JSONB           `json:"related_locations" db:"related_locations"`
	ForeshadowingHints   JSONB           `json:"foreshadowing_hints" db:"foreshadowing_hints"`
	ResolutionConditions JSONB           `json:"resolution_conditions" db:"resolution_conditions"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// ForeshadowingElement represents a hint or clue about future events
type ForeshadowingElement struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	GameSessionID        uuid.UUID       `json:"game_session_id" db:"game_session_id"`
	PlotThreadID         *uuid.UUID      `json:"plot_thread_id,omitempty" db:"plot_thread_id"`
	StoryArcID           *uuid.UUID      `json:"story_arc_id,omitempty" db:"story_arc_id"`
	ElementType          string          `json:"element_type" db:"element_type"` // prophecy, rumor, symbol, dream, omen
	Content              string          `json:"content" db:"content"`
	SubtletyLevel        int             `json:"subtlety_level" db:"subtlety_level"`
	Revealed             bool            `json:"revealed" db:"revealed"`
	RevealedSession      *int            `json:"revealed_session,omitempty" db:"revealed_session"`
	PlacementSuggestions JSONB           `json:"placement_suggestions" db:"placement_suggestions"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// CampaignTimeline represents an event in the campaign's chronology
type CampaignTimeline struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	GameSessionID     uuid.UUID       `json:"game_session_id" db:"game_session_id"`
	SessionMemoryID   *uuid.UUID      `json:"session_memory_id,omitempty" db:"session_memory_id"`
	EventDate         time.Time       `json:"event_date" db:"event_date"`
	RealSessionDate   time.Time       `json:"real_session_date" db:"real_session_date"`
	EventType         string          `json:"event_type" db:"event_type"` // combat, roleplay, discovery, decision
	EventTitle        string          `json:"event_title" db:"event_title"`
	EventDescription  string          `json:"event_description" db:"event_description"`
	ImpactLevel       int             `json:"impact_level" db:"impact_level"`
	RelatedArcs       JSONB           `json:"related_arcs" db:"related_arcs"`
	RelatedThreads    JSONB           `json:"related_threads" db:"related_threads"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
}

// NPCRelationship tracks how NPCs relate to the party and each other
type NPCRelationship struct {
	ID                    uuid.UUID       `json:"id" db:"id"`
	GameSessionID         uuid.UUID       `json:"game_session_id" db:"game_session_id"`
	NPCID                 uuid.UUID       `json:"npc_id" db:"npc_id"`
	TargetType            string          `json:"target_type" db:"target_type"` // character, npc, faction
	TargetID              uuid.UUID       `json:"target_id" db:"target_id"`
	RelationshipType      string          `json:"relationship_type" db:"relationship_type"` // ally, enemy, neutral, rival
	RelationshipScore     int             `json:"relationship_score" db:"relationship_score"` // -100 to 100
	LastInteractionSession *int           `json:"last_interaction_session,omitempty" db:"last_interaction_session"`
	InteractionHistory    JSONB           `json:"interaction_history" db:"interaction_history"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`
}

// Campaign-related request/response types

type CreateStoryArcRequest struct {
	Title           string          `json:"title" binding:"required"`
	Description     string          `json:"description"`
	ArcType         string          `json:"arc_type" binding:"required"`
	ParentArcID     *uuid.UUID      `json:"parent_arc_id,omitempty"`
	ImportanceLevel int             `json:"importance_level"`
}

type UpdateStoryArcRequest struct {
	Title           *string         `json:"title,omitempty"`
	Description     *string         `json:"description,omitempty"`
	Status          *string         `json:"status,omitempty"`
	ImportanceLevel *int            `json:"importance_level,omitempty"`
	Metadata        *JSONB          `json:"metadata,omitempty"`
}

type CreateSessionMemoryRequest struct {
	SessionNumber    int             `json:"session_number" binding:"required"`
	SessionDate      time.Time       `json:"session_date" binding:"required"`
	KeyEvents        []KeyEvent      `json:"key_events"`
	NPCsEncountered  []string        `json:"npcs_encountered"`
	DecisionsMade    []Decision      `json:"decisions_made"`
	ItemsAcquired    []string        `json:"items_acquired"`
	LocationsVisited []string        `json:"locations_visited"`
}

type KeyEvent struct {
	Time        string `json:"time"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

type Decision struct {
	Context string `json:"context"`
	Choice  string `json:"choice"`
	Outcome string `json:"outcome"`
}

type GenerateRecapRequest struct {
	SessionCount int `json:"session_count"` // How many sessions to include
}

type GenerateStoryArcRequest struct {
	Context      string   `json:"context"`      // Current campaign context
	PlayerGoals  []string `json:"player_goals"` // What players want to achieve
	ArcType      string   `json:"arc_type"`     // main_quest, side_quest, etc.
	Complexity   string   `json:"complexity"`   // simple, moderate, complex
}

type GenerateForeshadowingRequest struct {
	PlotThreadID  *uuid.UUID `json:"plot_thread_id,omitempty"`
	StoryArcID    *uuid.UUID `json:"story_arc_id,omitempty"`
	SubtletyLevel int        `json:"subtlety_level"` // 1-10
	ElementType   string     `json:"element_type"`   // prophecy, rumor, symbol, etc.
}

// AI-generated content structures

type GeneratedStoryArc struct {
	Title               string                `json:"title"`
	Description         string                `json:"description"`
	ArcType             string                `json:"arc_type"`
	ImportanceLevel     int                   `json:"importance_level"`
	KeyMilestones       []Milestone           `json:"key_milestones"`
	PotentialConflicts  []Conflict            `json:"potential_conflicts"`
	NPCsInvolved        []NPCInvolvement      `json:"npcs_involved"`
	ExpectedDuration    string                `json:"expected_duration"`
	PossibleResolutions []Resolution          `json:"possible_resolutions"`
	Connections         []ArcConnection       `json:"connections"`
}

type Milestone struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Trigger     string `json:"trigger"`
}

type Conflict struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Stakes      string `json:"stakes"`
}

type NPCInvolvement struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Motivation  string `json:"motivation"`
}

type Resolution struct {
	Type         string `json:"type"`
	Description  string `json:"description"`
	Consequences string `json:"consequences"`
}

type ArcConnection struct {
	ToArc       string `json:"to_arc"`
	Relationship string `json:"relationship"`
}

type GeneratedRecap struct {
	Summary          string            `json:"summary"`
	KeyEvents        []string          `json:"key_events"`
	UnresolvedThreads []string         `json:"unresolved_threads"`
	NPCUpdates       []NPCUpdate       `json:"npc_updates"`
	Cliffhanger      string            `json:"cliffhanger"`
	NextSessionHooks []string          `json:"next_session_hooks"`
}

type NPCUpdate struct {
	Name   string `json:"name"`
	Update string `json:"update"`
}

type GeneratedForeshadowing struct {
	Content              string            `json:"content"`
	ElementType          string            `json:"element_type"`
	SubtletyLevel        int               `json:"subtlety_level"`
	PlacementSuggestions []string          `json:"placement_suggestions"`
	RevealTiming         string            `json:"reveal_timing"`
	ConnectionHints      []string          `json:"connection_hints"`
}
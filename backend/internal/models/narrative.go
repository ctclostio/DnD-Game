package models

import (
	"time"
)

// NarrativeProfile tracks player storytelling preferences and patterns.
type NarrativeProfile struct {
	ID              string                 `json:"id" db:"id"`
	UserID          string                 `json:"user_id" db:"user_id"`
	CharacterID     string                 `json:"character_id" db:"character_id"`
	Preferences     StoryPreferences       `json:"preferences" db:"preferences"`
	DecisionHistory []DecisionRecord       `json:"decision_history" db:"decision_history"`
	PlayStyle       string                 `json:"play_style" db:"play_style"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	Analytics       map[string]interface{} `json:"analytics" db:"analytics"`
}

// StoryPreferences defines what kinds of stories resonate with a player.
type StoryPreferences struct {
	Themes           []string `json:"themes"`           // "redemption", "revenge", "discovery", etc.
	Tone             []string `json:"tone"`             // "dark", "heroic", "comedic", "tragic"
	Complexity       int      `json:"complexity"`       // 1-5 scale
	MoralAlignment   string   `json:"moral_alignment"`  // How they typically resolve moral dilemmas
	PacingPreference string   `json:"pacing"`           // "fast", "moderate", "slow-burn"
	CombatNarrative  float64  `json:"combat_narrative"` // 0-1, how much combat vs roleplay
}

// DecisionRecord tracks a significant player decision.
type DecisionRecord struct {
	Timestamp       time.Time              `json:"timestamp"`
	Context         string                 `json:"context"`
	Decision        string                 `json:"decision"`
	Alternatives    []string               `json:"alternatives"`
	Consequences    []string               `json:"consequences"`
	EmotionalWeight float64                `json:"emotional_weight"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// BackstoryElement represents a piece of character backstory that can be woven into narratives.
type BackstoryElement struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"character_id"`
	Type        string    `json:"type"` // "origin", "trauma", "goal", "relationship", "secret"
	Content     string    `json:"content"`
	Weight      float64   `json:"weight"`      // How important this is to the character
	Used        bool      `json:"used"`        // Has this been incorporated into a story
	UsageCount  int       `json:"usage_count"` // How many times referenced
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
}

// PersonalizedNarrative represents a story event tailored to a specific player.
type PersonalizedNarrative struct {
	ID                 string                 `json:"id"`
	BaseEventID        string                 `json:"base_event_id"`
	CharacterID        string                 `json:"character_id"`
	PersonalizedHooks  []NarrativeHook        `json:"personalized_hooks"`
	BackstoryCallbacks []BackstoryIntegration `json:"backstory_callbacks"`
	EmotionalResonance float64                `json:"emotional_resonance"`
	PredictedImpact    []PredictedImpact      `json:"predicted_impact"`
	Metadata           map[string]interface{} `json:"metadata"`
	GeneratedAt        time.Time              `json:"generated_at"`
}

// NarrativeHook is a story element designed to engage a specific player.
type NarrativeHook struct {
	Type        string  `json:"type"` // "moral_dilemma", "personal_connection", "mystery", "challenge"
	Content     string  `json:"content"`
	Relevance   float64 `json:"relevance"`              // How well this matches player preferences
	BackstoryID string  `json:"backstory_id,omitempty"` // If this hooks into backstory
}

// BackstoryIntegration shows how a backstory element is woven into current events.
type BackstoryIntegration struct {
	BackstoryElementID string `json:"backstory_element_id"`
	IntegrationType    string `json:"integration_type"` // "direct_reference", "thematic_echo", "consequence", "parallel"
	NarrativeText      string `json:"narrative_text"`
	Subtlety           int    `json:"subtlety"` // 1-5, how obvious the connection is
}

// PredictedImpact forecasts how this narrative might affect the player.
type PredictedImpact struct {
	Type        string  `json:"type"` // "emotional", "mechanical", "story_progression"
	Description string  `json:"description"`
	Likelihood  float64 `json:"likelihood"`
	Magnitude   float64 `json:"magnitude"`
}

// ConsequenceEvent represents a ripple effect from a player action.
type ConsequenceEvent struct {
	ID                string                 `json:"id" db:"id"`
	TriggerActionID   string                 `json:"trigger_action_id" db:"trigger_action_id"`
	TriggerType       string                 `json:"trigger_type" db:"trigger_type"`
	Description       string                 `json:"description" db:"description"`
	Severity          int                    `json:"severity" db:"severity"` // 1-10 scale
	Delay             string                 `json:"delay" db:"delay"`       // "immediate", "short", "medium", "long"
	ActualTriggerTime *time.Time             `json:"actual_trigger_time,omitempty" db:"actual_trigger_time"`
	AffectedEntities  []AffectedEntity       `json:"affected_entities" db:"affected_entities"`
	CascadeEffects    []CascadeEffect        `json:"cascade_effects" db:"cascade_effects"`
	Status            string                 `json:"status" db:"status"` // "pending", "triggered", "resolved", "prevented"
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
}

// AffectedEntity represents something impacted by a consequence.
type AffectedEntity struct {
	EntityType     string                 `json:"entity_type"` // "npc", "faction", "location", "economy", "reputation"
	EntityID       string                 `json:"entity_id"`
	EntityName     string                 `json:"entity_name"`
	ImpactType     string                 `json:"impact_type"`
	ImpactSeverity int                    `json:"impact_severity"`
	Description    string                 `json:"description"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CascadeEffect represents secondary consequences.
type CascadeEffect struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Probability float64    `json:"probability"`
	Timeline    string     `json:"timeline"`
	Triggered   bool       `json:"triggered"`
	TriggerTime *time.Time `json:"trigger_time,omitempty"`
}

// PerspectiveNarrative represents the same event from different viewpoints.
type PerspectiveNarrative struct {
	ID              string                 `json:"id" db:"id"`
	EventID         string                 `json:"event_id" db:"event_id"`
	PerspectiveType string                 `json:"perspective_type" db:"perspective_type"` // "npc", "faction", "deity", "historical"
	SourceID        string                 `json:"source_id" db:"source_id"`               // ID of NPC/faction/etc
	SourceName      string                 `json:"source_name" db:"source_name"`
	Narrative       string                 `json:"narrative" db:"narrative"`
	Bias            string                 `json:"bias" db:"bias"`               // "positive", "negative", "neutral", "conflicted"
	TruthLevel      float64                `json:"truth_level" db:"truth_level"` // 0-1, how accurate this perspective is
	HiddenDetails   []string               `json:"hidden_details" db:"hidden_details"`
	Contradictions  []Contradiction        `json:"contradictions" db:"contradictions"`
	EmotionalTone   string                 `json:"emotional_tone" db:"emotional_tone"`
	CulturalContext map[string]interface{} `json:"cultural_context" db:"cultural_context"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// Contradiction represents conflicting information between perspectives.
type Contradiction struct {
	OtherPerspectiveID string `json:"other_perspective_id"`
	ConflictingDetail  string `json:"conflicting_detail"`
	ThisVersion        string `json:"this_version"`
	OtherVersion       string `json:"other_version"`
	TruthValue         string `json:"truth_value"` // "this_true", "other_true", "both_partial", "neither_true"
}

// NarrativeEvent represents a significant event in the game world from a narrative perspective.
type NarrativeEvent struct {
	ID                string                 `json:"id" db:"id"`
	Type              string                 `json:"type" db:"type"`
	Name              string                 `json:"name" db:"name"`
	Description       string                 `json:"description" db:"description"`
	Location          string                 `json:"location" db:"location"`
	Timestamp         time.Time              `json:"timestamp" db:"timestamp"`
	Participants      []string               `json:"participants" db:"participants"`
	Witnesses         []string               `json:"witnesses" db:"witnesses"`
	ImmediateEffects  []string               `json:"immediate_effects" db:"immediate_effects"`
	PotentialRipples  []ConsequenceEvent     `json:"potential_ripples"`
	Perspectives      []PerspectiveNarrative `json:"perspectives"`
	PlayerInvolvement map[string]string      `json:"player_involvement" db:"player_involvement"`
	Status            string                 `json:"status" db:"status"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
}

// NarrativeMemory stores the AI's understanding of ongoing narratives.
type NarrativeMemory struct {
	ID              string                 `json:"id" db:"id"`
	SessionID       string                 `json:"session_id" db:"session_id"`
	CharacterID     string                 `json:"character_id" db:"character_id"`
	MemoryType      string                 `json:"memory_type" db:"memory_type"` // "decision", "consequence", "relationship", "discovery"
	Content         string                 `json:"content" db:"content"`
	EmotionalWeight float64                `json:"emotional_weight" db:"emotional_weight"`
	Connections     []string               `json:"connections" db:"connections"` // IDs of related memories
	Active          bool                   `json:"active" db:"active"`
	LastReferenced  time.Time              `json:"last_referenced" db:"last_referenced"`
	ReferenceCount  int                    `json:"reference_count" db:"reference_count"`
	Tags            []string               `json:"tags" db:"tags"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// PlayerAction represents a significant action taken by a player.
type PlayerAction struct {
	ID                    string                 `json:"id" db:"id"`
	SessionID             string                 `json:"session_id" db:"session_id"`
	CharacterID           string                 `json:"character_id" db:"character_id"`
	ActionType            string                 `json:"action_type" db:"action_type"`
	TargetType            string                 `json:"target_type" db:"target_type"`
	TargetID              string                 `json:"target_id" db:"target_id"`
	ActionDescription     string                 `json:"action_description" db:"action_description"`
	MoralWeight           string                 `json:"moral_weight" db:"moral_weight"`
	ImmediateResult       string                 `json:"immediate_result" db:"immediate_result"`
	PotentialConsequences int                    `json:"potential_consequences" db:"potential_consequences"`
	Timestamp             time.Time              `json:"timestamp" db:"timestamp"`
	Metadata              map[string]interface{} `json:"metadata" db:"metadata"`
}

// NarrativeThread represents a connected series of story events.
type NarrativeThread struct {
	ID                  string                 `json:"id" db:"id"`
	Name                string                 `json:"name" db:"name"`
	Description         string                 `json:"description" db:"description"`
	ThreadType          string                 `json:"thread_type" db:"thread_type"`
	Status              string                 `json:"status" db:"status"`
	ConnectedEvents     []string               `json:"connected_events" db:"connected_events"`
	KeyParticipants     []string               `json:"key_participants" db:"key_participants"`
	TensionLevel        float64                `json:"tension_level" db:"tension_level"`
	ResolutionProximity float64                `json:"resolution_proximity" db:"resolution_proximity"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
	ResolvedAt          *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`
}

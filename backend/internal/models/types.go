package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB represents a JSONB database type that can be used with PostgreSQL.
type JSONB json.RawMessage

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*j = JSONB(v)
		return nil
	case string:
		*j = JSONB([]byte(v))
		return nil
	default:
		return errors.New("cannot scan unknown type into JSONB")
	}
}

// MarshalJSON implements json.Marshaler for JSONB
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler for JSONB
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("models.JSONB: UnmarshalJSON on nil pointer")
	}
	*j = JSONB(data)
	return nil
}

// Rule represents a game rule that can be executed.
type Rule struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Type        string                 `json:"type" db:"type"`
	Trigger     string                 `json:"trigger" db:"trigger"`
	Conditions  []RuleCondition        `json:"conditions" db:"conditions"`
	Actions     []RuleAction           `json:"actions" db:"actions"`
	Priority    int                    `json:"priority" db:"priority"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// RuleAction represents an action that a rule can perform.
type RuleAction struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
}

// DiceRollResult represents the result of a dice roll.
type DiceRollResult struct {
	Total    int    `json:"total"`
	Rolls    []int  `json:"rolls"`
	Modifier int    `json:"modifier"`
	Notation string `json:"notation"`
	Purpose  string `json:"purpose,omitempty"`
	Critical bool   `json:"critical"`
	Fumble   bool   `json:"fumble"`
}

// PerspectiveSource represents the source of a narrative perspective.
type PerspectiveSource struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"` // character, npc, location, object, narrator
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Background      string                 `json:"background,omitempty"`
	Motivations     []string               `json:"motivations,omitempty"`
	Relationships   map[string]string      `json:"relationships,omitempty"`
	CulturalContext map[string]interface{} `json:"cultural_context,omitempty"`
}

// FactionDecision represents a decision made by a faction.
type FactionDecision struct {
	ID           string                 `json:"id" db:"id"`
	FactionID    string                 `json:"faction_id" db:"faction_id"`
	DecisionType string                 `json:"decision_type" db:"decision_type"`
	Context      map[string]interface{} `json:"context" db:"context"`
	Options      []DecisionOption       `json:"options" db:"options"`
	ChosenOption *DecisionOption        `json:"chosen_option,omitempty" db:"chosen_option"`
	Reasoning    string                 `json:"reasoning" db:"reasoning"`
	Confidence   float64                `json:"confidence" db:"confidence"`
	Timestamp    int64                  `json:"timestamp" db:"timestamp"`
}

// DecisionOption represents a possible choice in a faction decision.
type DecisionOption struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Benefits     []string               `json:"benefits"`
	Risks        []string               `json:"risks"`
	Requirements map[string]interface{} `json:"requirements"`
	Score        float64                `json:"score"`
}

// FactionDecisionResult represents the outcome of a faction decision.
type FactionDecisionResult struct {
	DecisionID    string                 `json:"decision_id"`
	Success       bool                   `json:"success"`
	Consequences  []string               `json:"consequences"`
	ImpactMetrics map[string]interface{} `json:"impact_metrics"`
	NextActions   []string               `json:"next_actions"`
}

// PlayerInteraction represents a player's interaction with the game world.
type PlayerInteraction struct {
	ID        string                 `json:"id"`
	PlayerID  string                 `json:"player_id"`
	Type      string                 `json:"type"`
	Target    string                 `json:"target"`
	Action    string                 `json:"action"`
	Context   map[string]interface{} `json:"context"`
	Timestamp int64                  `json:"timestamp"`
}

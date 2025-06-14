package models

import (
	"time"
)

// NPC represents a non-player character or monster.
type NPC struct {
	ID                  string         `json:"id" db:"id"`
	GameSessionID       string         `json:"gameSessionId" db:"game_session_id"`
	Name                string         `json:"name" db:"name"`
	Type                string         `json:"type" db:"type"` // humanoid, beast, undead, etc.
	Size                string         `json:"size" db:"size"` // tiny, small, medium, large, huge, gargantuan
	Alignment           string         `json:"alignment" db:"alignment"`
	ArmorClass          int            `json:"armorClass" db:"armor_class"`
	HitPoints           int            `json:"hitPoints" db:"hit_points"`
	MaxHitPoints        int            `json:"maxHitPoints" db:"max_hit_points"`
	Speed               map[string]int `json:"speed" db:"speed"` // walk, fly, swim, etc.
	Attributes          Attributes     `json:"attributes" db:"attributes"`
	SavingThrows        SavingThrows   `json:"savingThrows" db:"saving_throws"`
	Skills              []Skill        `json:"skills" db:"skills"`
	DamageResistances   []string       `json:"damageResistances" db:"damage_resistances"`
	DamageImmunities    []string       `json:"damageImmunities" db:"damage_immunities"`
	ConditionImmunities []string       `json:"conditionImmunities" db:"condition_immunities"`
	Senses              map[string]int `json:"senses" db:"senses"` // darkvision, blindsight, etc.
	Languages           []string       `json:"languages" db:"languages"`
	ChallengeRating     float64        `json:"challengeRating" db:"challenge_rating"`
	ExperiencePoints    int            `json:"experiencePoints" db:"experience_points"`
	Abilities           []NPCAbility   `json:"abilities" db:"abilities"`
	Actions             []NPCAction    `json:"actions" db:"actions"`
	LegendaryActions    int            `json:"legendaryActions" db:"legendary_actions"`
	IsTemplate          bool           `json:"isTemplate" db:"is_template"` // If true, this is a template that can be copied
	CreatedBy           string         `json:"createdBy" db:"created_by"`   // User ID of creator (DM)
	CreatedAt           time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time      `json:"updatedAt" db:"updated_at"`
}

// NPCAbility represents a special ability or trait.
type NPCAbility struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NPCAction represents an action the NPC can take.
type NPCAction struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // action, bonus_action, reaction, legendary
	Description string `json:"description"`
	AttackBonus int    `json:"attackBonus,omitempty"`
	Damage      string `json:"damage,omitempty"` // e.g., "2d6+3"
	DamageType  string `json:"damageType,omitempty"`
	Range       string `json:"range,omitempty"`
	SaveDC      int    `json:"saveDC,omitempty"`
	SaveType    string `json:"saveType,omitempty"` // STR, DEX, CON, INT, WIS, CHA
}

// NPCTemplate represents a template for creating NPCs.
type NPCTemplate struct {
	ID                  string         `json:"id" db:"id"`
	Name                string         `json:"name" db:"name"`
	Source              string         `json:"source" db:"source"` // MM, custom, etc.
	Type                string         `json:"type" db:"type"`
	Size                string         `json:"size" db:"size"`
	Alignment           string         `json:"alignment" db:"alignment"`
	ArmorClass          int            `json:"armorClass" db:"armor_class"`
	HitDice             string         `json:"hitDice" db:"hit_dice"` // e.g., "2d8+2"
	Speed               map[string]int `json:"speed" db:"speed"`
	Attributes          Attributes     `json:"attributes" db:"attributes"`
	SavingThrows        SavingThrows   `json:"savingThrows" db:"saving_throws"`
	Skills              []Skill        `json:"skills" db:"skills"`
	DamageResistances   []string       `json:"damageResistances" db:"damage_resistances"`
	DamageImmunities    []string       `json:"damageImmunities" db:"damage_immunities"`
	ConditionImmunities []string       `json:"conditionImmunities" db:"condition_immunities"`
	Senses              map[string]int `json:"senses" db:"senses"`
	Languages           []string       `json:"languages" db:"languages"`
	ChallengeRating     float64        `json:"challengeRating" db:"challenge_rating"`
	Abilities           []NPCAbility   `json:"abilities" db:"abilities"`
	Actions             []NPCAction    `json:"actions" db:"actions"`
	LegendaryActions    int            `json:"legendaryActions" db:"legendary_actions"`
	CreatedAt           time.Time      `json:"createdAt" db:"created_at"`
}

// NPCSearchFilter represents search criteria for NPCs.
type NPCSearchFilter struct {
	GameSessionID    string  `json:"gameSessionId"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	Size             string  `json:"size"`
	MinCR            float64 `json:"minCR"`
	MaxCR            float64 `json:"maxCR"`
	IncludeTemplates bool    `json:"includeTemplates"`
}

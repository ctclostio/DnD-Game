package models

import (
	"time"

	"github.com/google/uuid"
)

// CustomRace represents a player-created race using AI generation
type CustomRace struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	UserPrompt  string    `json:"userPrompt" db:"user_prompt"`

	// Ability Score Improvements
	AbilityScoreIncreases map[string]int `json:"abilityScoreIncreases" db:"ability_score_increases"`

	// Basic attributes
	Size  string `json:"size" db:"size"`
	Speed int    `json:"speed" db:"speed"`

	// Features and traits
	Traits    []RacialTrait `json:"traits" db:"traits"`
	Languages []string      `json:"languages" db:"languages"`

	// Special abilities
	Darkvision  int      `json:"darkvision" db:"darkvision"`
	Resistances []string `json:"resistances" db:"resistances"`
	Immunities  []string `json:"immunities" db:"immunities"`

	// Proficiencies
	SkillProficiencies  []string `json:"skillProficiencies" db:"skill_proficiencies"`
	ToolProficiencies   []string `json:"toolProficiencies" db:"tool_proficiencies"`
	WeaponProficiencies []string `json:"weaponProficiencies" db:"weapon_proficiencies"`
	ArmorProficiencies  []string `json:"armorProficiencies" db:"armor_proficiencies"`

	// Metadata
	CreatedBy      uuid.UUID  `json:"createdBy" db:"created_by"`
	ApprovedBy     *uuid.UUID `json:"approvedBy,omitempty" db:"approved_by"`
	ApprovalStatus string     `json:"approvalStatus" db:"approval_status"`
	ApprovalNotes  *string    `json:"approvalNotes,omitempty" db:"approval_notes"`
	BalanceScore   *int       `json:"balanceScore,omitempty" db:"balance_score"`

	// Usage tracking
	TimesUsed int  `json:"timesUsed" db:"times_used"`
	IsPublic  bool `json:"isPublic" db:"is_public"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// RacialTrait represents a special ability or feature of a race
type RacialTrait struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CustomRaceRequest represents a request to create a custom race
type CustomRaceRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"required,min=10,max=1000"`
}

// CustomRaceGenerationResult represents the AI-generated race data
type CustomRaceGenerationResult struct {
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	AbilityScoreIncreases map[string]int `json:"abilityScoreIncreases"`
	Size                  string         `json:"size"`
	Speed                 int            `json:"speed"`
	Traits                []RacialTrait  `json:"traits"`
	Languages             []string       `json:"languages"`
	Darkvision            int            `json:"darkvision"`
	Resistances           []string       `json:"resistances"`
	Immunities            []string       `json:"immunities"`
	SkillProficiencies    []string       `json:"skillProficiencies"`
	ToolProficiencies     []string       `json:"toolProficiencies"`
	WeaponProficiencies   []string       `json:"weaponProficiencies"`
	ArmorProficiencies    []string       `json:"armorProficiencies"`
	BalanceScore          int            `json:"balanceScore"`
	BalanceExplanation    string         `json:"balanceExplanation"`
}

// ApprovalStatus constants
const (
	ApprovalStatusPending        = "pending"
	ApprovalStatusApproved       = "approved"
	ApprovalStatusRejected       = "rejected"
	ApprovalStatusRevisionNeeded = "revision_needed"
)

// ValidSizes for D&D races
var ValidSizes = []string{"Tiny", "Small", "Medium", "Large", "Huge", "Gargantuan"}

// ValidDamageTypes for resistances and immunities
var ValidDamageTypes = []string{
	"acid", "bludgeoning", "cold", "fire", "force", "lightning",
	"necrotic", "piercing", "poison", "psychic", "radiant", "slashing", "thunder",
}

// ValidSkills for proficiencies
var ValidSkills = []string{
	"Acrobatics", "Animal Handling", "Arcana", "Athletics", "Deception",
	"History", "Insight", "Intimidation", "Investigation", "Medicine",
	"Nature", "Perception", "Performance", "Persuasion", "Religion",
	"Sleight of Hand", "Stealth", "Survival",
}

// ValidLanguages in D&D
var ValidLanguages = []string{
	"Common", "Dwarvish", "Elvish", "Giant", "Gnomish", "Goblin",
	"Halfling", "Orc", "Abyssal", "Celestial", "Draconic", "Deep Speech",
	"Infernal", "Primordial", "Sylvan", "Undercommon",
}

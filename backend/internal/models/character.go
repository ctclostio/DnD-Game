package models

import "time"

type Character struct {
	ID                  string                 `json:"id" db:"id"`
	UserID              string                 `json:"userId" db:"user_id"`
	Name                string                 `json:"name" db:"name"`
	Race                string                 `json:"race" db:"race"`
	Subrace             string                 `json:"subrace,omitempty" db:"subrace"`
	CustomRaceID        *string                `json:"customRaceId,omitempty" db:"custom_race_id"`
	Class               string                 `json:"class" db:"class"`
	Subclass            string                 `json:"subclass,omitempty" db:"subclass"`
	CustomClassID       *string                `json:"customClassId,omitempty" db:"custom_class_id"`
	Background          string                 `json:"background" db:"background"`
	Alignment           string                 `json:"alignment" db:"alignment"`
	Level               int                    `json:"level" db:"level"`
	ExperiencePoints    int                    `json:"experiencePoints" db:"experience_points"`
	HitPoints           int                    `json:"hitPoints" db:"hit_points"`
	MaxHitPoints        int                    `json:"maxHitPoints" db:"max_hit_points"`
	TempHitPoints       int                    `json:"tempHitPoints" db:"temp_hit_points"`
	HitDice             string                 `json:"hitDice" db:"hit_dice"`
	ArmorClass          int                    `json:"armorClass" db:"armor_class"`
	Initiative          int                    `json:"initiative" db:"initiative"`
	Speed               int                    `json:"speed" db:"speed"`
	ProficiencyBonus    int                    `json:"proficiencyBonus" db:"proficiency_bonus"`
	Attributes          Attributes             `json:"attributes" db:"attributes"`
	SavingThrows        SavingThrows           `json:"savingThrows" db:"saving_throws"`
	Skills              []Skill                `json:"skills" db:"skills"`
	Proficiencies       Proficiencies          `json:"proficiencies" db:"proficiencies"`
	Features            []Feature              `json:"features" db:"features"`
	Equipment           []Item                 `json:"equipment" db:"equipment"`
	Spells              SpellData              `json:"spells" db:"spells"`
	Resources           map[string]interface{} `json:"resources" db:"resources"`
	CarryCapacity       float64                `json:"carryCapacity" db:"carry_capacity"`
	CurrentWeight       float64                `json:"currentWeight" db:"current_weight"`
	AttunementSlotsUsed int                    `json:"attunementSlotsUsed" db:"attunement_slots_used"`
	AttunementSlotsMax  int                    `json:"attunementSlotsMax" db:"attunement_slots_max"`
	CreatedAt           time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time              `json:"updatedAt" db:"updated_at"`
}

type Attributes struct {
	Strength     int `json:"strength" db:"strength"`
	Dexterity    int `json:"dexterity" db:"dexterity"`
	Constitution int `json:"constitution" db:"constitution"`
	Intelligence int `json:"intelligence" db:"intelligence"`
	Wisdom       int `json:"wisdom" db:"wisdom"`
	Charisma     int `json:"charisma" db:"charisma"`
}

type Skill struct {
	Name        string `json:"name" db:"name"`
	Modifier    int    `json:"modifier" db:"modifier"`
	Proficiency bool   `json:"proficiency" db:"proficiency"`
}

type Spell struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Level       int    `json:"level" db:"level"`
	School      string `json:"school" db:"school"`
	CastingTime string `json:"castingTime" db:"casting_time"`
	Range       string `json:"range" db:"range"`
	Components  string `json:"components" db:"components"`
	Duration    string `json:"duration" db:"duration"`
	Description string `json:"description" db:"description"`
	Prepared    bool   `json:"prepared,omitempty" db:"prepared"`
}

type SavingThrows struct {
	Strength     SavingThrow `json:"strength" db:"strength"`
	Dexterity    SavingThrow `json:"dexterity" db:"dexterity"`
	Constitution SavingThrow `json:"constitution" db:"constitution"`
	Intelligence SavingThrow `json:"intelligence" db:"intelligence"`
	Wisdom       SavingThrow `json:"wisdom" db:"wisdom"`
	Charisma     SavingThrow `json:"charisma" db:"charisma"`
}

type SavingThrow struct {
	Modifier    int  `json:"modifier" db:"modifier"`
	Proficiency bool `json:"proficiency" db:"proficiency"`
}

type Proficiencies struct {
	Armor     []string `json:"armor" db:"armor"`
	Weapons   []string `json:"weapons" db:"weapons"`
	Tools     []string `json:"tools" db:"tools"`
	Languages []string `json:"languages" db:"languages"`
}

type Feature struct {
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Level       int    `json:"level" db:"level"`
	Source      string `json:"source" db:"source"`
}

type SpellData struct {
	SpellcastingAbility string      `json:"spellcastingAbility,omitempty" db:"spellcasting_ability"`
	SpellSaveDC         int         `json:"spellSaveDC,omitempty" db:"spell_save_dc"`
	SpellAttackBonus    int         `json:"spellAttackBonus,omitempty" db:"spell_attack_bonus"`
	SpellSlots          []SpellSlot `json:"spellSlots,omitempty" db:"spell_slots"`
	SpellsKnown         []Spell     `json:"spellsKnown,omitempty" db:"spells_known"`
	CantripsKnown       int         `json:"cantripsKnown,omitempty" db:"cantrips_known"`
}

type SpellSlot struct {
	Level     int `json:"level" db:"level"`
	Total     int `json:"total" db:"total"`
	Remaining int `json:"remaining" db:"remaining"`
}

// CustomClass represents a player-created custom class
type CustomClass struct {
	ID                       string                 `json:"id" db:"id"`
	UserID                   string                 `json:"userId" db:"user_id"`
	Name                     string                 `json:"name" db:"name"`
	Description              string                 `json:"description" db:"description"`
	HitDie                   int                    `json:"hitDie" db:"hit_die"`
	PrimaryAbility           string                 `json:"primaryAbility" db:"primary_ability"`
	SavingThrowProficiencies []string               `json:"savingThrowProficiencies" db:"saving_throw_proficiencies"`
	SkillProficiencies       []string               `json:"skillProficiencies" db:"skill_proficiencies"`
	SkillChoices             int                    `json:"skillChoices" db:"skill_choices"`
	StartingEquipment        string                 `json:"startingEquipment" db:"starting_equipment"`
	ArmorProficiencies       []string               `json:"armorProficiencies" db:"armor_proficiencies"`
	WeaponProficiencies      []string               `json:"weaponProficiencies" db:"weapon_proficiencies"`
	ToolProficiencies        []string               `json:"toolProficiencies,omitempty" db:"tool_proficiencies"`
	ClassFeatures            []ClassFeature         `json:"classFeatures" db:"class_features"`
	SubclassName             string                 `json:"subclassName,omitempty" db:"subclass_name"`
	SubclassLevel            int                    `json:"subclassLevel,omitempty" db:"subclass_level"`
	Subclasses               []Subclass             `json:"subclasses,omitempty" db:"subclasses"`
	SpellcastingAbility      string                 `json:"spellcastingAbility,omitempty" db:"spellcasting_ability"`
	SpellList                []string               `json:"spellList,omitempty" db:"spell_list"`
	SpellsKnownProgression   []int                  `json:"spellsKnownProgression,omitempty" db:"spells_known_progression"`
	CantripsKnownProgression []int                  `json:"cantripsKnownProgression,omitempty" db:"cantrips_known_progression"`
	SpellSlotsProgression    map[string]interface{} `json:"spellSlotsProgression,omitempty" db:"spell_slots_progression"`
	RitualCasting            bool                   `json:"ritualCasting" db:"ritual_casting"`
	SpellcastingFocus        string                 `json:"spellcastingFocus,omitempty" db:"spellcasting_focus"`
	BalanceScore             int                    `json:"balanceScore" db:"balance_score"`
	PowerLevel               string                 `json:"powerLevel" db:"power_level"`
	IsApproved               bool                   `json:"isApproved" db:"is_approved"`
	ApprovedBy               *string                `json:"approvedBy,omitempty" db:"approved_by"`
	ApprovedAt               *time.Time             `json:"approvedAt,omitempty" db:"approved_at"`
	DMNotes                  string                 `json:"dmNotes,omitempty" db:"dm_notes"`
	CreatedAt                time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt                time.Time              `json:"updatedAt" db:"updated_at"`
}

type ClassFeature struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UsesPerRest string `json:"usesPerRest,omitempty"`
	RestType    string `json:"restType,omitempty"`
	Passive     bool   `json:"passive"`
}

type Subclass struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Features    []ClassFeature `json:"features"`
}

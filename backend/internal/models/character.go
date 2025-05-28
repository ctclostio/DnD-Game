package models

import "time"

type Character struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Race         string    `json:"race"`
	Class        string    `json:"class"`
	Level        int       `json:"level"`
	ExperiencePoints int   `json:"experiencePoints"`
	HitPoints    int       `json:"hitPoints"`
	MaxHitPoints int       `json:"maxHitPoints"`
	ArmorClass   int       `json:"armorClass"`
	Speed        int       `json:"speed"`
	Attributes   Attributes `json:"attributes"`
	Skills       []Skill    `json:"skills"`
	Equipment    []Item     `json:"equipment"`
	Spells       []Spell    `json:"spells"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type Attributes struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

type Skill struct {
	Name        string `json:"name"`
	Modifier    int    `json:"modifier"`
	Proficiency bool   `json:"proficiency"`
}

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Weight      float64 `json:"weight"`
	Value       int    `json:"value"`
	Properties  map[string]interface{} `json:"properties"`
}

type Spell struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	School      string `json:"school"`
	CastingTime string `json:"castingTime"`
	Range       string `json:"range"`
	Components  string `json:"components"`
	Duration    string `json:"duration"`
	Description string `json:"description"`
}
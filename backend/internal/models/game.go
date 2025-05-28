package models

import "time"

type GameSession struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	DungeonMaster string    `json:"dungeonMaster"`
	Players     []Player    `json:"players"`
	Status      string      `json:"status"` // active, paused, completed
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type Player struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CharacterID string   `json:"characterId"`
	IsOnline   bool      `json:"isOnline"`
	JoinedAt   time.Time `json:"joinedAt"`
}

type DiceRoll struct {
	ID         string    `json:"id"`
	PlayerID   string    `json:"playerId"`
	DiceType   string    `json:"diceType"` // d4, d6, d8, d10, d12, d20, d100
	Count      int       `json:"count"`
	Modifier   int       `json:"modifier"`
	Results    []int     `json:"results"`
	Total      int       `json:"total"`
	Purpose    string    `json:"purpose"` // attack, damage, skill check, etc.
	Timestamp  time.Time `json:"timestamp"`
}

type GameEvent struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"sessionId"`
	Type      string                 `json:"type"` // roll, message, combat, etc.
	PlayerID  string                 `json:"playerId"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}
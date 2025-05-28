package models

import "time"

type GameSession struct {
	ID          string      `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	DMUserID    string      `json:"dmUserId" db:"dm_user_id"`
	Players     []Player    `json:"players" db:"-"` // Not stored directly in game_sessions table
	Status      string      `json:"status" db:"status"` // active, paused, completed
	CreatedAt   time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time   `json:"updatedAt" db:"updated_at"`
}

type Player struct {
	ID          string    `json:"id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	CharacterID string    `json:"characterId" db:"character_id"`
	IsOnline    bool      `json:"isOnline" db:"is_online"`
	JoinedAt    time.Time `json:"joinedAt" db:"joined_at"`
}

type DiceRoll struct {
	ID            string    `json:"id" db:"id"`
	GameSessionID string    `json:"gameSessionId" db:"game_session_id"`
	UserID        string    `json:"userId" db:"user_id"`
	DiceType      string    `json:"diceType" db:"dice_type"` // d4, d6, d8, d10, d12, d20, d100
	Count         int       `json:"count" db:"count"`
	Modifier      int       `json:"modifier" db:"modifier"`
	Results       []int     `json:"results" db:"results"`
	Total         int       `json:"total" db:"total"`
	Purpose       string    `json:"purpose" db:"purpose"` // attack, damage, skill check, etc.
	RollNotation  string    `json:"rollNotation" db:"roll_notation"` // e.g., "2d20+5"
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
}

type GameEvent struct {
	ID        string                 `json:"id" db:"id"`
	SessionID string                 `json:"sessionId" db:"session_id"`
	Type      string                 `json:"type" db:"type"` // roll, message, combat, etc.
	PlayerID  string                 `json:"playerId" db:"player_id"`
	Data      map[string]interface{} `json:"data" db:"data"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}
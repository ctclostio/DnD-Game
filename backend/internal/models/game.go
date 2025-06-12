package models

import "time"

// GameStatus represents the status of a game session
type GameStatus string

const (
	GameStatusPending   GameStatus = "pending"
	GameStatusActive    GameStatus = "active"
	GameStatusPaused    GameStatus = "paused"
	GameStatusCompleted GameStatus = "completed"
)

type GameSession struct {
	ID                    string                 `json:"id" db:"id"`
	DMID                  string                 `json:"dmId" db:"dm_user_id"`
	Name                  string                 `json:"name" db:"name"`
	Description           string                 `json:"description" db:"description"`
	Code                  string                 `json:"code" db:"code"`
	IsActive              bool                   `json:"isActive" db:"is_active"`
	Status                GameStatus             `json:"status" db:"status"`
	State                 map[string]interface{} `json:"state" db:"session_state"`
	MaxPlayers            int                    `json:"maxPlayers" db:"max_players"`
	IsPublic              bool                   `json:"isPublic" db:"is_public"`
	RequiresInvite        bool                   `json:"requiresInvite" db:"requires_invite"`
	AllowedCharacterLevel int                    `json:"allowedCharacterLevel" db:"allowed_character_level"`
	CreatedAt             time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time              `json:"updatedAt" db:"updated_at"`
	StartedAt             *time.Time             `json:"startedAt,omitempty" db:"started_at"`
	EndedAt               *time.Time             `json:"endedAt,omitempty" db:"ended_at"`
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
	Purpose       string    `json:"purpose" db:"purpose"`            // attack, damage, skill check, etc.
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

// GameInvite represents an invitation to join a game session
type GameInvite struct {
	ID           string     `json:"id" db:"id"`
	SessionID    string     `json:"sessionId" db:"session_id"`
	InviterID    string     `json:"inviterId" db:"inviter_id"`
	InviteeEmail string     `json:"inviteeEmail" db:"invitee_email"`
	InviteeID    *string    `json:"inviteeId,omitempty" db:"invitee_id"`
	Code         string     `json:"code" db:"code"`
	ExpiresAt    time.Time  `json:"expiresAt" db:"expires_at"`
	UsedAt       *time.Time `json:"usedAt,omitempty" db:"used_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

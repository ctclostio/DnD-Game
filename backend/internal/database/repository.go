package database

import (
	"context"

	"github.com/your-username/dnd-game/backend/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
}

// CharacterRepository defines the interface for character data operations
type CharacterRepository interface {
	Create(ctx context.Context, character *models.Character) error
	GetByID(ctx context.Context, id string) (*models.Character, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Character, error)
	Update(ctx context.Context, character *models.Character) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.Character, error)
}

// GameSessionRepository defines the interface for game session data operations
type GameSessionRepository interface {
	Create(ctx context.Context, session *models.GameSession) error
	GetByID(ctx context.Context, id string) (*models.GameSession, error)
	GetByDMUserID(ctx context.Context, dmUserID string) ([]*models.GameSession, error)
	GetByParticipantUserID(ctx context.Context, userID string) ([]*models.GameSession, error)
	Update(ctx context.Context, session *models.GameSession) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.GameSession, error)
	
	// Participant management
	AddParticipant(ctx context.Context, sessionID, userID string, characterID *string) error
	RemoveParticipant(ctx context.Context, sessionID, userID string) error
	GetParticipants(ctx context.Context, sessionID string) ([]*models.GameParticipant, error)
	UpdateParticipantOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error
}

// DiceRollRepository defines the interface for dice roll data operations
type DiceRollRepository interface {
	Create(ctx context.Context, roll *models.DiceRoll) error
	GetByID(ctx context.Context, id string) (*models.DiceRoll, error)
	GetByGameSession(ctx context.Context, sessionID string, offset, limit int) ([]*models.DiceRoll, error)
	GetByUser(ctx context.Context, userID string, offset, limit int) ([]*models.DiceRoll, error)
	GetByGameSessionAndUser(ctx context.Context, sessionID, userID string, offset, limit int) ([]*models.DiceRoll, error)
	Delete(ctx context.Context, id string) error
}

// Repositories aggregates all repository interfaces
type Repositories struct {
	Users        UserRepository
	Characters   CharacterRepository
	GameSessions GameSessionRepository
	DiceRolls    DiceRollRepository
	NPCs         NPCRepository
	Inventory    *InventoryRepository
}
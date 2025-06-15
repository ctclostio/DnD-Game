package interfaces

import (
	"context"
	
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// GameSessionInterface handles core game session operations
type GameSessionInterface interface {
	Create(ctx context.Context, session *models.GameSession) error
	GetByID(ctx context.Context, id string) (*models.GameSession, error)
	GetByDMUserID(ctx context.Context, dmUserID string) ([]*models.GameSession, error)
	GetByParticipantUserID(ctx context.Context, userID string) ([]*models.GameSession, error)
	Update(ctx context.Context, session *models.GameSession) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.GameSession, error)
}

// GameParticipantInterface manages session participants
type GameParticipantInterface interface {
	AddParticipant(ctx context.Context, sessionID, userID string, characterID *string) error
	RemoveParticipant(ctx context.Context, sessionID, userID string) error
	GetParticipants(ctx context.Context, sessionID string) ([]*models.GameParticipant, error)
	UpdateParticipantOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error
}

// LegacyGameSessionRepository maintains backward compatibility
// This interface combines all the focused interfaces
// It will be deprecated once all code is updated to use specific interfaces
type LegacyGameSessionRepository interface {
	GameSessionInterface
	GameParticipantInterface
}
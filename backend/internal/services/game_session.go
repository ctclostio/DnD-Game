package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
)

type GameSessionService struct {
	repo database.GameSessionRepository
}

func NewGameSessionService(repo database.GameSessionRepository) *GameSessionService {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	
	return &GameSessionService{
		repo: repo,
	}
}

// generateSessionCode generates a unique 6-character alphanumeric code
func (s *GameSessionService) generateSessionCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}

// CreateSession creates a new game session
func (s *GameSessionService) CreateSession(ctx context.Context, session *models.GameSession) error {
	// Validate session
	if session.Name == "" {
		return fmt.Errorf("session name is required")
	}
	if session.DMID == "" {
		return fmt.Errorf("dungeon master user ID is required")
	}
	
	// Set default status
	if session.Status == "" {
		session.Status = models.GameStatusActive
	}
	
	// Generate unique code if not provided
	if session.Code == "" {
		// Generate codes until we find a unique one
		// In production, you'd check against the database
		session.Code = s.generateSessionCode()
	}
	
	// Set IsActive to true by default
	session.IsActive = true
	
	// Initialize empty state if nil
	if session.State == nil {
		session.State = make(map[string]interface{})
	}
	
	// Create session
	if err := s.repo.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	
	// Add DM as a participant
	if err := s.repo.AddParticipant(ctx, session.ID, session.DMID, nil); err != nil {
		return fmt.Errorf("failed to add DM as participant: %w", err)
	}
	
	return nil
}

// GetSessionByID retrieves a game session by ID
func (s *GameSessionService) GetSessionByID(ctx context.Context, id string) (*models.GameSession, error) {
	return s.repo.GetByID(ctx, id)
}

// GetSessionsByDM retrieves all sessions for a dungeon master
func (s *GameSessionService) GetSessionsByDM(ctx context.Context, dmUserID string) ([]*models.GameSession, error) {
	return s.repo.GetByDMUserID(ctx, dmUserID)
}

// GetSessionsByPlayer retrieves all sessions where user is a participant
func (s *GameSessionService) GetSessionsByPlayer(ctx context.Context, userID string) ([]*models.GameSession, error) {
	return s.repo.GetByParticipantUserID(ctx, userID)
}

// UpdateSession updates a game session
func (s *GameSessionService) UpdateSession(ctx context.Context, session *models.GameSession) error {
	// Validate session ID
	if session.ID == "" {
		return fmt.Errorf("session ID is required")
	}
	
	// Check if session exists
	existing, err := s.repo.GetByID(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}
	
	// Preserve DM user ID and created at
	session.DMID = existing.DMID
	session.CreatedAt = existing.CreatedAt
	
	return s.repo.Update(ctx, session)
}

// DeleteSession deletes a game session
func (s *GameSessionService) DeleteSession(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// JoinSession adds a player to a game session
func (s *GameSessionService) JoinSession(ctx context.Context, sessionID, userID string, characterID *string) error {
	// Validate input
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	
	// Check if session exists
	_, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}
	
	// Add participant
	return s.repo.AddParticipant(ctx, sessionID, userID, characterID)
}

// LeaveSession removes a player from a game session
func (s *GameSessionService) LeaveSession(ctx context.Context, sessionID, userID string) error {
	// Validate input
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	
	// Check if session exists
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}
	
	// Don't allow DM to leave
	if session.DMID == userID {
		return fmt.Errorf("dungeon master cannot leave the session")
	}
	
	// Remove participant
	return s.repo.RemoveParticipant(ctx, sessionID, userID)
}

// UpdatePlayerOnlineStatus updates the online status of a player in a session
func (s *GameSessionService) UpdatePlayerOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error {
	return s.repo.UpdateParticipantOnlineStatus(ctx, sessionID, userID, isOnline)
}

// GetSessionParticipants retrieves all participants for a game session
func (s *GameSessionService) GetSessionParticipants(ctx context.Context, sessionID string) ([]*models.GameParticipant, error) {
	return s.repo.GetParticipants(ctx, sessionID)
}

// ValidateUserInSession checks if a user is a participant in a session
func (s *GameSessionService) ValidateUserInSession(ctx context.Context, sessionID, userID string) error {
	participants, err := s.repo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}
	
	for _, p := range participants {
		if p.UserID == userID {
			return nil
		}
	}
	
	// Also check if user is the DM
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}
	
	if session.DMID == userID {
		return nil
	}
	
	return fmt.Errorf("user is not a participant in this session")
}

// GetGameSession gets a game session by ID (alias for GetSessionByID)
func (s *GameSessionService) GetGameSession(ctx context.Context, id string) (*models.GameSession, error) {
	return s.GetSessionByID(ctx, id)
}

// GetSession gets a game session by ID (alias for GetSessionByID)
func (s *GameSessionService) GetSession(ctx context.Context, id string) (*models.GameSession, error) {
	return s.GetSessionByID(ctx, id)
}
package services

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Error messages
const (
	errSessionIDRequired = "session ID is required"
)

type GameSessionService struct {
	repo     database.GameSessionRepository
	charRepo database.CharacterRepository
	userRepo database.UserRepository
}

func NewGameSessionService(repo database.GameSessionRepository) *GameSessionService {
	return &GameSessionService{
		repo: repo,
	}
}

// SetCharacterRepository sets the character repository (optional for backward compatibility)
func (s *GameSessionService) SetCharacterRepository(charRepo database.CharacterRepository) {
	s.charRepo = charRepo
}

// SetUserRepository sets the user repository (optional for backward compatibility)
func (s *GameSessionService) SetUserRepository(userRepo database.UserRepository) {
	s.userRepo = userRepo
}

// generateSessionCode generates a cryptographically secure 6-character alphanumeric code
func (s *GameSessionService) generateSessionCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6
	
	// Use crypto/rand for secure random generation
	bytes := make([]byte, codeLength)
	if _, err := rand.Read(bytes); err != nil {
		// In the unlikely event of an error, panic as this is a critical security component
		panic(fmt.Sprintf("failed to generate secure session code: %v", err))
	}
	
	// Convert random bytes to characters from our charset
	for i := range bytes {
		bytes[i] = chars[bytes[i]%byte(len(chars))]
	}
	
	return string(bytes)
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

	// Set default status to active so the session can be used immediately.
	// If the caller didn't specify a status, also default IsActive to true so
	// the session can be joined right away.
	if session.Status == "" {
		session.Status = models.GameStatusPending
	}

	// Set security defaults
	if session.MaxPlayers == 0 {
		session.MaxPlayers = 6 // Default D&D party size
	}
	if session.MaxPlayers < 2 {
		return fmt.Errorf("max players must be at least 2")
	}
	if session.MaxPlayers > 10 {
		return fmt.Errorf("max players cannot exceed 10")
	}

	// Default to private sessions that require invites
	if !session.IsPublic {
		session.RequiresInvite = true
	}

	// Generate unique code if not provided
	if session.Code == "" {
		// TODO: Check uniqueness against database
		session.Code = s.generateSessionCode()
	}

	// Sessions should remain as specified by the caller; most will be active

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
		return fmt.Errorf(errSessionIDRequired)
	}

	// Check if session exists
	existing, err := s.repo.GetByID(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve session %s for update: %w", session.ID, err)
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

// JoinSession adds a player to a game session with comprehensive security checks
func (s *GameSessionService) JoinSession(ctx context.Context, sessionID, userID string, characterID *string) error {
	// Validate input
	if err := s.validateJoinInput(sessionID, userID); err != nil {
		return err
	}

	// Get session and check if it can be joined
	session, err := s.getJoinableSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Check participant status and capacity
	if err := s.validateParticipantStatus(ctx, sessionID, userID, session); err != nil {
		return err
	}

	// Validate character if provided
	if err := s.validateCharacterForSession(ctx, userID, characterID, session); err != nil {
		return err
	}

	// Add participant
	return s.repo.AddParticipant(ctx, sessionID, userID, characterID)
}

// validateJoinInput validates the basic input parameters
func (s *GameSessionService) validateJoinInput(sessionID, userID string) error {
	if sessionID == "" {
		return fmt.Errorf(errSessionIDRequired)
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	return nil
}

// getJoinableSession retrieves and validates session can be joined
func (s *GameSessionService) getJoinableSession(ctx context.Context, sessionID string) (*models.GameSession, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session %s for joining: %w", sessionID, err)
	}

	// Security check: If a session is marked active but flagged inactive,
	// joining should fail. Sessions in other states may allow joining.
	if session.Status == models.GameStatusActive && !session.IsActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Security check: Session must not be completed
	if session.Status == models.GameStatusCompleted {
		return nil, fmt.Errorf("cannot join completed session")
	}

	return session, nil
}

// validateParticipantStatus checks if user can join based on participant rules
func (s *GameSessionService) validateParticipantStatus(ctx context.Context, sessionID, userID string, session *models.GameSession) error {
	participants, err := s.repo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to check participants: %w", err)
	}

	currentPlayerCount := 0
	for _, p := range participants {
		if p.UserID == userID {
			return fmt.Errorf("you are already in this session")
		}
		// Count non-DM players
		if p.UserID != session.DMID {
			currentPlayerCount++
		}
	}

	// Security check: Session capacity
	if session.MaxPlayers > 0 && currentPlayerCount >= session.MaxPlayers-1 { // -1 because DM doesn't count
		return fmt.Errorf("session is full (max %d players)", session.MaxPlayers-1)
	}

	return nil
}

// validateCharacterForSession validates character ownership and requirements
func (s *GameSessionService) validateCharacterForSession(ctx context.Context, userID string, characterID *string, session *models.GameSession) error {
	if characterID == nil || *characterID == "" || s.charRepo == nil {
		return nil
	}

	character, err := s.charRepo.GetByID(ctx, *characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	if character.UserID != userID {
		return fmt.Errorf("you don't own this character")
	}

	// Check character level requirements if set
	if session.AllowedCharacterLevel > 0 && character.Level > session.AllowedCharacterLevel {
		return fmt.Errorf("character level %d exceeds session limit of %d",
			character.Level, session.AllowedCharacterLevel)
	}

	return nil
}

// LeaveSession removes a player from a game session
func (s *GameSessionService) LeaveSession(ctx context.Context, sessionID, userID string) error {
	// Validate input
	if sessionID == "" {
		return fmt.Errorf(errSessionIDRequired)
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Check if session exists
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to retrieve session %s for leaving: %w", sessionID, err)
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
		return fmt.Errorf("failed to retrieve session %s for user validation: %w", sessionID, err)
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

// KickPlayer removes a player from the session (DM only operation)
func (s *GameSessionService) KickPlayer(ctx context.Context, sessionID, playerID string) error {
	// Validate input
	if sessionID == "" {
		return fmt.Errorf(errSessionIDRequired)
	}
	if playerID == "" {
		return fmt.Errorf("player ID is required")
	}

	// Check if session exists
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to retrieve session %s for kicking player: %w", sessionID, err)
	}

	// Security check: Cannot kick the DM
	if session.DMID == playerID {
		return fmt.Errorf("cannot kick the dungeon master")
	}

	// Security check: Player must be in the session
	participants, err := s.repo.GetParticipants(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get participants: %w", err)
	}

	found := false
	for _, p := range participants {
		if p.UserID == playerID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("player is not in this session")
	}

	// Remove participant
	return s.repo.RemoveParticipant(ctx, sessionID, playerID)
}

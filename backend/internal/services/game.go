package services

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Error message constants
const (
	errMsgSessionNotFound = "session not found"
)

type GameService struct {
	mu       sync.RWMutex
	sessions map[string]*models.GameSession
	events   map[string][]*models.GameEvent
}

func NewGameService() *GameService {
	return &GameService{
		sessions: make(map[string]*models.GameSession),
		events:   make(map[string][]*models.GameEvent),
	}
}

func (s *GameService) CreateSession(session *models.GameSession) (*models.GameSession, error) {
	session.ID = generateID()
	session.Status = models.GameStatusActive
	session.CreatedAt = time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	s.events[session.ID] = make([]*models.GameEvent, 0)

	return session, nil
}

func (s *GameService) GetSessionByID(id string) (*models.GameSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[id]
	if !exists {
		return nil, errors.New(errMsgSessionNotFound)
	}
	return session, nil
}

func (s *GameService) AddPlayerToSession(sessionID, _ string, player *models.Player) error {
	s.mu.RLock()
	_, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return errors.New(errMsgSessionNotFound)
	}

	player.JoinedAt = time.Now()
	// TODO: Store player in database
	// session.Players = append(session.Players, *player)

	return nil
}

func (s *GameService) RecordGameEvent(event *models.GameEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[event.SessionID]; !exists {
		return errors.New(errMsgSessionNotFound)
	}

	event.ID = generateID()
	event.Timestamp = time.Now()

	s.events[event.SessionID] = append(s.events[event.SessionID], event)
	return nil
}

func (s *GameService) GetSessionEvents(sessionID string) ([]*models.GameEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events, exists := s.events[sessionID]
	if !exists {
		return nil, errors.New(errMsgSessionNotFound)
	}
	return events, nil
}

func generateID() string {
	return uuid.New().String()
}

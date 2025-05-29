package services

import (
	"errors"
	"time"
	"github.com/google/uuid"
	"github.com/your-username/dnd-game/backend/internal/models"
)

type GameService struct {
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
	
	s.sessions[session.ID] = session
	s.events[session.ID] = make([]*models.GameEvent, 0)
	
	return session, nil
}

func (s *GameService) GetSessionByID(id string) (*models.GameSession, error) {
	session, exists := s.sessions[id]
	if !exists {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (s *GameService) AddPlayerToSession(sessionID, playerID string, player *models.Player) error {
	_, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}
	
	player.JoinedAt = time.Now()
	// TODO: Store player in database
	// session.Players = append(session.Players, *player)
	
	return nil
}

func (s *GameService) RecordGameEvent(event *models.GameEvent) error {
	if _, exists := s.sessions[event.SessionID]; !exists {
		return errors.New("session not found")
	}
	
	event.ID = generateID()
	event.Timestamp = time.Now()
	
	s.events[event.SessionID] = append(s.events[event.SessionID], event)
	return nil
}

func (s *GameService) GetSessionEvents(sessionID string) ([]*models.GameEvent, error) {
	events, exists := s.events[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}
	return events, nil
}

func generateID() string {
	return uuid.New().String()
}
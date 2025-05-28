package services

import (
	"errors"
	"time"
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
	session.Status = "active"
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	
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
	session, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}
	
	player.JoinedAt = time.Now()
	session.Players = append(session.Players, *player)
	session.UpdatedAt = time.Now()
	
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
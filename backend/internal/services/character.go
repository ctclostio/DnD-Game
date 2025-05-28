package services

import (
	"errors"
	"time"
	"github.com/your-username/dnd-game/backend/internal/models"
)

type CharacterService struct {
	// In-memory storage for now, replace with database later
	characters map[string]*models.Character
}

func NewCharacterService() *CharacterService {
	return &CharacterService{
		characters: make(map[string]*models.Character),
	}
}

func (s *CharacterService) GetAllCharacters() ([]*models.Character, error) {
	result := make([]*models.Character, 0, len(s.characters))
	for _, char := range s.characters {
		result = append(result, char)
	}
	return result, nil
}

func (s *CharacterService) GetCharacterByID(id string) (*models.Character, error) {
	char, exists := s.characters[id]
	if !exists {
		return nil, errors.New("character not found")
	}
	return char, nil
}

func (s *CharacterService) CreateCharacter(char *models.Character) (*models.Character, error) {
	char.ID = generateID()
	char.CreatedAt = time.Now()
	char.UpdatedAt = time.Now()
	
	// Set default values
	if char.Level == 0 {
		char.Level = 1
	}
	if char.MaxHitPoints == 0 {
		char.MaxHitPoints = 10 + getModifier(char.Attributes.Constitution)
	}
	char.HitPoints = char.MaxHitPoints
	
	s.characters[char.ID] = char
	return char, nil
}

func (s *CharacterService) UpdateCharacter(char *models.Character) (*models.Character, error) {
	existing, exists := s.characters[char.ID]
	if !exists {
		return nil, errors.New("character not found")
	}
	
	char.CreatedAt = existing.CreatedAt
	char.UpdatedAt = time.Now()
	s.characters[char.ID] = char
	return char, nil
}

func generateID() string {
	// Simple ID generation, replace with UUID in production
	return time.Now().Format("20060102150405")
}

func getModifier(ability int) int {
	return (ability - 10) / 2
}
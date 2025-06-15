package testutil

import (
	"context"
	"testing"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockFactory creates mock objects for testing
type MockFactory struct {
	t *testing.T
}

// NewMockFactory creates a new mock factory
func NewMockFactory(t *testing.T) *MockFactory {
	return &MockFactory{t: t}
}

// MockCharacterRepository creates a mock character repository
type MockCharacterRepository struct {
	mock.Mock
}

func (m *MockCharacterRepository) Create(char *models.Character) error {
	args := m.Called(char)
	return args.Error(0)
}

func (m *MockCharacterRepository) GetByID(id int64) (*models.Character, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterRepository) GetByUserID(userID int64) ([]*models.Character, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

func (m *MockCharacterRepository) Update(char *models.Character) error {
	args := m.Called(char)
	return args.Error(0)
}

func (m *MockCharacterRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCharacterRepository) UpdateHP(id int64, hp int) error {
	args := m.Called(id, hp)
	return args.Error(0)
}

// MockUserRepository creates a mock user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id int64) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// MockGameSessionRepository creates a mock game session repository
type MockGameSessionRepository struct {
	mock.Mock
}

func (m *MockGameSessionRepository) Create(session *models.GameSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetByID(id int64) (*models.GameSession, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetByCode(code string) (*models.GameSession, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) Update(session *models.GameSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetActiveByUserID(userID int64) ([]*models.GameSession, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

// MockCombatRepository creates a mock combat repository
type MockCombatRepository struct {
	mock.Mock
}

func (m *MockCombatRepository) Create(combat *models.Combat) error {
	args := m.Called(combat)
	return args.Error(0)
}

func (m *MockCombatRepository) GetByID(id int64) (*models.Combat, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatRepository) GetBySessionID(sessionID int64) (*models.Combat, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatRepository) Update(combat *models.Combat) error {
	args := m.Called(combat)
	return args.Error(0)
}

// MockInventoryRepository creates a mock inventory repository
type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) CreateItem(item *models.InventoryItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetItemByID(id int64) (*models.InventoryItem, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventoryItem), args.Error(1)
}

func (m *MockInventoryRepository) GetItemsByCharacterID(charID int64) ([]*models.InventoryItem, error) {
	args := m.Called(charID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.InventoryItem), args.Error(1)
}

func (m *MockInventoryRepository) UpdateItem(item *models.InventoryItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockInventoryRepository) DeleteItem(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockDiceRollRepository creates a mock dice roll repository
type MockDiceRollRepository struct {
	mock.Mock
}

func (m *MockDiceRollRepository) Create(roll *models.DiceRoll) error {
	args := m.Called(roll)
	return args.Error(0)
}

func (m *MockDiceRollRepository) GetByID(id int64) (*models.DiceRoll, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollRepository) GetBySessionID(sessionID int64, limit int) ([]*models.DiceRoll, error) {
	args := m.Called(sessionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

// Service Mocks

// MockCharacterService creates a mock character service
type MockCharacterService struct {
	mock.Mock
}

func (m *MockCharacterService) Create(char *models.Character) error {
	args := m.Called(char)
	return args.Error(0)
}

func (m *MockCharacterService) GetByID(id int64) (*models.Character, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterService) GetByUserID(userID int64) ([]*models.Character, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

func (m *MockCharacterService) Update(char *models.Character) error {
	args := m.Called(char)
	return args.Error(0)
}

func (m *MockCharacterService) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCharacterService) LevelUp(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCharacterService) TakeDamage(id int64, damage int) error {
	args := m.Called(id, damage)
	return args.Error(0)
}

func (m *MockCharacterService) Heal(id int64, healing int) error {
	args := m.Called(id, healing)
	return args.Error(0)
}

// MockCombatService creates a mock combat service
type MockCombatService struct {
	mock.Mock
}

func (m *MockCombatService) StartCombat(sessionID int64, participants []models.Combatant) (*models.Combat, error) {
	args := m.Called(sessionID, participants)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatService) GetCombat(id int64) (*models.Combat, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatService) NextTurn(combatID int64) error {
	args := m.Called(combatID)
	return args.Error(0)
}

func (m *MockCombatService) EndCombat(combatID int64) error {
	args := m.Called(combatID)
	return args.Error(0)
}

// MockLLMProvider creates a mock LLM provider
type MockLLMProvider struct {
	mock.Mock
}

func (m *MockLLMProvider) GenerateCompletion(ctx context.Context, prompt, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) GenerateContent(ctx context.Context, prompt, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) GenerateResponse(prompt string, options map[string]interface{}) (string, error) {
	args := m.Called(prompt, options)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) GenerateCharacterBackstory(character *models.Character) (string, error) {
	args := m.Called(character)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) GenerateNPCDialog(npc *models.NPC, context string) (string, error) {
	args := m.Called(npc, context)
	return args.String(0), args.Error(1)
}

// MockWebSocketHub creates a mock WebSocket hub
type MockWebSocketHub struct {
	mock.Mock
}

func (m *MockWebSocketHub) Broadcast(message interface{}) {
	m.Called(message)
}

func (m *MockWebSocketHub) SendToUser(userID int64, message interface{}) {
	m.Called(userID, message)
}

func (m *MockWebSocketHub) SendToSession(sessionID int64, message interface{}) {
	m.Called(sessionID, message)
}

// ExpectationSetters provides helper methods to set up mock expectations

// SetupCharacterRepoExpectations sets up common character repository expectations
func SetupCharacterRepoExpectations(repo *MockCharacterRepository, char *models.Character) {
	repo.On("GetByID", char.ID).Return(char, nil)
	repo.On("Update", char).Return(nil)
	repo.On("UpdateHP", char.ID, mock.AnythingOfType("int")).Return(nil)
}

// SetupUserRepoExpectations sets up common user repository expectations
func SetupUserRepoExpectations(repo *MockUserRepository, user *models.User) {
	repo.On("GetByID", user.ID).Return(user, nil)
	repo.On("GetByUsername", user.Username).Return(user, nil)
	repo.On("GetByEmail", user.Email).Return(user, nil)
}

// SetupGameSessionRepoExpectations sets up common game session repository expectations
func SetupGameSessionRepoExpectations(repo *MockGameSessionRepository, session *models.GameSession) {
	repo.On("GetByID", session.ID).Return(session, nil)
	repo.On("Update", session).Return(nil)
}

// SetupCombatRepoExpectations sets up common combat repository expectations
func SetupCombatRepoExpectations(repo *MockCombatRepository, combat *models.Combat) {
	repo.On("GetByID", combat.ID).Return(combat, nil)
	repo.On("GetBySessionID", combat.GameSessionID).Return(combat, nil)
	repo.On("Update", combat).Return(nil)
}

// MockBehaviorPresets provides preset mock behaviors

// PresetSuccessfulCharacterCreation sets up mocks for successful character creation
func PresetSuccessfulCharacterCreation(
	userRepo *MockUserRepository,
	charRepo *MockCharacterRepository,
	user *models.User,
	_ *models.Character,
) {
	userRepo.On("GetByID", user.ID).Return(user, nil)
	charRepo.On("Create", mock.AnythingOfType("*models.Character")).Return(nil)
}

// PresetSuccessfulCombatStart sets up mocks for successful combat start
func PresetSuccessfulCombatStart(
	sessionRepo *MockGameSessionRepository,
	combatRepo *MockCombatRepository,
	session *models.GameSession,
) {
	sessionRepo.On("GetByID", session.ID).Return(session, nil)
	combatRepo.On("Create", mock.AnythingOfType("*models.Combat")).Return(nil)
	sessionRepo.On("Update", mock.AnythingOfType("*models.GameSession")).Return(nil)
}

// PresetNotFound sets up mocks for not found scenarios
func PresetNotFound(repo interface{}, method string, id int64) {
	switch r := repo.(type) {
	case *MockCharacterRepository:
		r.On(method, id).Return(nil, models.ErrNotFound)
	case *MockUserRepository:
		r.On(method, id).Return(nil, models.ErrNotFound)
	case *MockGameSessionRepository:
		r.On(method, id).Return(nil, models.ErrNotFound)
	case *MockCombatRepository:
		r.On(method, id).Return(nil, models.ErrNotFound)
	}
}

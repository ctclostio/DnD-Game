package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// MockJWTManager is a mock implementation of auth.JWTManager
type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateTokenPair(userID, username, email, role string) (*auth.TokenPair, error) {
	args := m.Called(userID, username, email, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(tokenString string, expectedType auth.TokenType) (*auth.Claims, error) {
	args := m.Called(tokenString, expectedType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockJWTManager) RefreshToken(refreshToken string) (*auth.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

// MockLLMProvider is a mock implementation of LLMProvider interface
type MockLLMProvider struct {
	mock.Mock
}

func (m *MockLLMProvider) GenerateContent(ctx context.Context, prompt, system string) (string, error) {
	args := m.Called(ctx, prompt, system)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) GenerateJSON(ctx context.Context, prompt, system string, schema interface{}) (string, error) {
	args := m.Called(ctx, prompt, system, schema)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) StreamContent(ctx context.Context, prompt, system string) (<-chan string, <-chan error) {
	args := m.Called(ctx, prompt, system)
	return args.Get(0).(<-chan string), args.Get(1).(<-chan error)
}

func (m *MockLLMProvider) GenerateCompletion(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	args := m.Called(ctx, prompt, systemPrompt)
	return args.String(0), args.Error(1)
}

// MockCustomClassRepository is a mock implementation
type MockCustomClassRepository struct {
	mock.Mock
}

func (m *MockCustomClassRepository) Create(class *models.CustomClass) error {
	args := m.Called(class)
	return args.Error(0)
}

func (m *MockCustomClassRepository) GetByID(id string) (*models.CustomClass, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CustomClass), args.Error(1)
}

func (m *MockCustomClassRepository) GetByUserID(userID string, includeUnapproved bool) ([]*models.CustomClass, error) {
	args := m.Called(userID, includeUnapproved)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CustomClass), args.Error(1)
}

func (m *MockCustomClassRepository) GetApproved() ([]*models.CustomClass, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CustomClass), args.Error(1)
}

func (m *MockCustomClassRepository) Update(class *models.CustomClass) error {
	args := m.Called(class)
	return args.Error(0)
}

func (m *MockCustomClassRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCustomClassRepository) Approve(id string, approvedBy string) error {
	args := m.Called(id, approvedBy)
	return args.Error(0)
}

// MockCombatService is a mock implementation of combat service
type MockCombatService struct {
	mock.Mock
}

func (m *MockCombatService) StartCombat(ctx context.Context, sessionID string, participants []models.Combatant) (*models.Combat, error) {
	args := m.Called(ctx, sessionID, participants)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatService) GetCombatState(ctx context.Context, combatID string) (*models.Combat, error) {
	args := m.Called(ctx, combatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatService) ExecuteAction(ctx context.Context, combatID string, action models.CombatAction) (*models.Combat, error) {
	args := m.Called(ctx, combatID, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Combat), args.Error(1)
}

func (m *MockCombatService) EndCombat(ctx context.Context, combatID string) error {
	args := m.Called(ctx, combatID)
	return args.Error(0)
}

// Helper function to create a test user
func CreateTestUser(id, username, email, role string) *models.User {
	return &models.User{
		ID:           id,
		Username:     username,
		Email:        email,
		Role:         role,
		PasswordHash: "hashed_password",
	}
}

// Helper function to create a test character
func CreateTestCharacter(id, userID, name, race, class string) *models.Character {
	return &models.Character{
		ID:           id,
		UserID:       userID,
		Name:         name,
		Race:         race,
		Class:        class,
		Level:        1,
		HitPoints:    10,
		MaxHitPoints: 10,
		ArmorClass:   15,
		Speed:        30,
		Attributes: models.Attributes{
			Strength:     16,
			Dexterity:    14,
			Constitution: 13,
			Intelligence: 12,
			Wisdom:       15,
			Charisma:     10,
		},
	}
}

// Helper function to create a test dice roll
func CreateTestDiceRoll(sessionID, userID, notation string) *models.DiceRoll {
	return &models.DiceRoll{
		GameSessionID: sessionID,
		UserID:        userID,
		RollNotation:  notation,
		Purpose:       "Test roll",
	}
}

// Helper function to create test item
func CreateTestItem(id, name string, itemType models.ItemType, value int, weight float64) *models.Item {
	return &models.Item{
		ID:         id,
		Name:       name,
		Type:       itemType,
		Value:      value,
		Weight:     weight,
		Properties: make(map[string]interface{}),
	}
}

// Helper function to create test inventory item
func CreateTestInventoryItem(characterID, itemID string, quantity int, equipped, attuned bool, item *models.Item) *models.InventoryItem {
	return &models.InventoryItem{
		CharacterID: characterID,
		ItemID:      itemID,
		Quantity:    quantity,
		Equipped:    equipped,
		Attuned:     attuned,
		Item:        item,
	}
}

// Helper function to create a test game session
func CreateTestGameSession(id, dmID, name string) *models.GameSession {
	return &models.GameSession{
		ID:          id,
		DMID:        dmID,
		Name:        name,
		Description: "Test session",
		Status:      models.GameStatusActive,
	}
}

// MockDiceRollService is a mock implementation of DiceRollService
type MockDiceRollService struct {
	mock.Mock
}

func (m *MockDiceRollService) RollDice(ctx context.Context, roll *models.DiceRoll) error {
	args := m.Called(ctx, roll)
	return args.Error(0)
}

func (m *MockDiceRollService) SetGameSessionRepo(repo database.GameSessionRepository) {
	m.Called(repo)
}

func (m *MockDiceRollService) GetRollByID(ctx context.Context, id string) (*models.DiceRoll, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollService) GetRollsBySession(ctx context.Context, sessionID string, offset, limit int) ([]*models.DiceRoll, error) {
	args := m.Called(ctx, sessionID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollService) GetRollsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollService) GetRollsBySessionAndUser(ctx context.Context, sessionID, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	args := m.Called(ctx, sessionID, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollService) DeleteRoll(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDiceRollService) SimulateRoll(notation string) (*models.DiceRoll, error) {
	args := m.Called(notation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiceRoll), args.Error(1)
}

// MockAIDMAssistantService is a mock implementation of AIDMAssistantInterface
type MockAIDMAssistantService struct {
	mock.Mock
}

func (m *MockAIDMAssistantService) GenerateNPCDialogue(ctx context.Context, req models.NPCDialogueRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockAIDMAssistantService) GenerateLocationDescription(ctx context.Context, req models.LocationDescriptionRequest) (*models.AILocation, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AILocation), args.Error(1)
}

func (m *MockAIDMAssistantService) GenerateCombatNarration(ctx context.Context, req models.CombatNarrationRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockAIDMAssistantService) GeneratePlotTwist(ctx context.Context, currentContext map[string]interface{}) (*models.AIStoryElement, error) {
	args := m.Called(ctx, currentContext)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AIStoryElement), args.Error(1)
}

func (m *MockAIDMAssistantService) GenerateEnvironmentalHazard(ctx context.Context, locationType string, difficulty int) (*models.AIEnvironmentalHazard, error) {
	args := m.Called(ctx, locationType, difficulty)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AIEnvironmentalHazard), args.Error(1)
}

func (m *MockAIDMAssistantService) GenerateNPC(ctx context.Context, role string, context map[string]interface{}) (*models.AINPC, error) {
	args := m.Called(ctx, role, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AINPC), args.Error(1)
}

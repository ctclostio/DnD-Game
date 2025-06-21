// Package mocks provides mock implementations of repository interfaces for testing
package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Helper functions to reduce code duplication in mock implementations

func handleErrorReturn(args mock.Arguments, index int) error {
	return args.Error(index)
}

func handleSingleReturn[T any](args mock.Arguments, valueIndex, errorIndex int) (*T, error) {
	if args.Get(valueIndex) == nil {
		return nil, args.Error(errorIndex)
	}
	return args.Get(valueIndex).(*T), args.Error(errorIndex)
}

func handleSliceReturn[T any](args mock.Arguments, valueIndex, errorIndex int) ([]*T, error) {
	if args.Get(valueIndex) == nil {
		return nil, args.Error(errorIndex)
	}
	return args.Get(valueIndex).([]*T), args.Error(errorIndex)
}

// MockUserRepository is a mock implementation of database.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	// Create a new user record
	args := m.Called(ctx, user)
	return handleErrorReturn(args, 0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	return handleSingleReturn[models.User](args, 0, 1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	return handleSingleReturn[models.User](args, 0, 1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return handleSingleReturn[models.User](args, 0, 1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	// Update existing user record
	args := m.Called(ctx, user)
	return handleErrorReturn(args, 0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return handleErrorReturn(args, 0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, offset, limit)
	return handleSliceReturn[models.User](args, 0, 1)
}

// MockCharacterRepository is a mock implementation of database.CharacterRepository
type MockCharacterRepository struct {
	mock.Mock
}

func (m *MockCharacterRepository) Create(ctx context.Context, character *models.Character) error {
	// Create a new character record
	args := m.Called(ctx, character)
	return handleErrorReturn(args, 0)
}

func (m *MockCharacterRepository) GetByID(ctx context.Context, id string) (*models.Character, error) {
	args := m.Called(ctx, id)
	return handleSingleReturn[models.Character](args, 0, 1)
}

func (m *MockCharacterRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Character, error) {
	args := m.Called(ctx, userID)
	return handleSliceReturn[models.Character](args, 0, 1)
}

func (m *MockCharacterRepository) Update(ctx context.Context, character *models.Character) error {
	// Update existing character record
	args := m.Called(ctx, character)
	return handleErrorReturn(args, 0)
}

func (m *MockCharacterRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return handleErrorReturn(args, 0)
}

func (m *MockCharacterRepository) List(ctx context.Context, offset, limit int) ([]*models.Character, error) {
	args := m.Called(ctx, offset, limit)
	return handleSliceReturn[models.Character](args, 0, 1)
}

// MockDiceRollRepository is a mock implementation of database.DiceRollRepository
type MockDiceRollRepository struct {
	mock.Mock
}

func (m *MockDiceRollRepository) Create(ctx context.Context, roll *models.DiceRoll) error {
	args := m.Called(ctx, roll)
	return args.Error(0)
}

func (m *MockDiceRollRepository) GetByID(ctx context.Context, id string) (*models.DiceRoll, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollRepository) GetByGameSession(ctx context.Context, sessionID string, offset, limit int) ([]*models.DiceRoll, error) {
	args := m.Called(ctx, sessionID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollRepository) GetByUser(ctx context.Context, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollRepository) GetByGameSessionAndUser(ctx context.Context, sessionID, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	args := m.Called(ctx, sessionID, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiceRoll), args.Error(1)
}

func (m *MockDiceRollRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockGameSessionRepository is a mock implementation of database.GameSessionRepository
type MockGameSessionRepository struct {
	mock.Mock
}

func (m *MockGameSessionRepository) Create(ctx context.Context, session *models.GameSession) error {
	// Create a new game session record
	args := m.Called(ctx, session)
	return handleErrorReturn(args, 0)
}

func (m *MockGameSessionRepository) GetByID(ctx context.Context, id string) (*models.GameSession, error) {
	args := m.Called(ctx, id)
	return handleSingleReturn[models.GameSession](args, 0, 1)
}

func (m *MockGameSessionRepository) GetByDMUserID(ctx context.Context, dmUserID string) ([]*models.GameSession, error) {
	args := m.Called(ctx, dmUserID)
	return handleSliceReturn[models.GameSession](args, 0, 1)
}

func (m *MockGameSessionRepository) GetByParticipantUserID(ctx context.Context, userID string) ([]*models.GameSession, error) {
	args := m.Called(ctx, userID)
	return handleSliceReturn[models.GameSession](args, 0, 1)
}

func (m *MockGameSessionRepository) Update(ctx context.Context, session *models.GameSession) error {
	args := m.Called(ctx, session)
	return handleErrorReturn(args, 0)
}

func (m *MockGameSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return handleErrorReturn(args, 0)
}

func (m *MockGameSessionRepository) List(ctx context.Context, offset, limit int) ([]*models.GameSession, error) {
	args := m.Called(ctx, offset, limit)
	return handleSliceReturn[models.GameSession](args, 0, 1)
}

func (m *MockGameSessionRepository) AddParticipant(ctx context.Context, sessionID, userID string, characterID *string) error {
	args := m.Called(ctx, sessionID, userID, characterID)
	return handleErrorReturn(args, 0)
}

func (m *MockGameSessionRepository) RemoveParticipant(ctx context.Context, sessionID, userID string) error {
	args := m.Called(ctx, sessionID, userID)
	return handleErrorReturn(args, 0)
}

func (m *MockGameSessionRepository) GetParticipants(ctx context.Context, sessionID string) ([]*models.GameParticipant, error) {
	args := m.Called(ctx, sessionID)
	return handleSliceReturn[models.GameParticipant](args, 0, 1)
}

func (m *MockGameSessionRepository) UpdateParticipantOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error {
	args := m.Called(ctx, sessionID, userID, isOnline)
	return handleErrorReturn(args, 0)
}

// MockInventoryRepository is a mock implementation of database.InventoryRepository
type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) CreateItem(item *models.Item) error {
	args := m.Called(item)
	return handleErrorReturn(args, 0)
}

func (m *MockInventoryRepository) GetItem(itemID string) (*models.Item, error) {
	args := m.Called(itemID)
	return handleSingleReturn[models.Item](args, 0, 1)
}

func (m *MockInventoryRepository) GetItemsByType(itemType models.ItemType) ([]*models.Item, error) {
	args := m.Called(itemType)
	return handleSliceReturn[models.Item](args, 0, 1)
}

func (m *MockInventoryRepository) AddItemToInventory(characterID, itemID string, quantity int) error {
	args := m.Called(characterID, itemID, quantity)
	return handleErrorReturn(args, 0)
}

func (m *MockInventoryRepository) RemoveItemFromInventory(characterID, itemID string, quantity int) error {
	args := m.Called(characterID, itemID, quantity)
	return handleErrorReturn(args, 0)
}

func (m *MockInventoryRepository) GetCharacterInventory(characterID string) ([]*models.InventoryItem, error) {
	args := m.Called(characterID)
	return handleSliceReturn[models.InventoryItem](args, 0, 1)
}

func (m *MockInventoryRepository) EquipItem(characterID, itemID string, equip bool) error {
	args := m.Called(characterID, itemID, equip)
	return handleErrorReturn(args, 0)
}

func (m *MockInventoryRepository) AttuneItem(characterID, itemID string) error {
	args := m.Called(characterID, itemID)
	return handleErrorReturn(args, 0)
}

func (m *MockInventoryRepository) UnattuneItem(characterID, itemID string) error {
	args := m.Called(characterID, itemID)
	return handleErrorReturn(args, 0)
}

func (m *MockInventoryRepository) GetCharacterCurrency(characterID string) (*models.Currency, error) {
	args := m.Called(characterID)
	return handleSingleReturn[models.Currency](args, 0, 1)
}

func (m *MockInventoryRepository) CreateCharacterCurrency(currency *models.Currency) error {
	args := m.Called(currency)
	return handleErrorReturn(args, 0)
}

func (m *MockInventoryRepository) UpdateCharacterCurrency(currency *models.Currency) error {
	args := m.Called(currency)
	return handleErrorReturn(args, 0)
}

func (m *MockInventoryRepository) GetCharacterWeight(characterID string) (*models.InventoryWeight, error) {
	args := m.Called(characterID)
	return handleSingleReturn[models.InventoryWeight](args, 0, 1)
}

// MockRefreshTokenRepository is a mock implementation of database.RefreshTokenRepository
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(userID, tokenID, token string, expiresAt time.Time) error {
	args := m.Called(userID, tokenID, token, expiresAt)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) ValidateAndGet(token string) (*database.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) Revoke(tokenID string) error {
	args := m.Called(tokenID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) CleanupExpired() error {
	args := m.Called()
	return args.Error(0)
}

// MockNPCRepository is a mock implementation of database.NPCRepository
type MockNPCRepository struct {
	mock.Mock
}

func (m *MockNPCRepository) Create(ctx context.Context, npc *models.NPC) error {
	args := m.Called(ctx, npc)
	return args.Error(0)
}

func (m *MockNPCRepository) GetByID(ctx context.Context, id string) (*models.NPC, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NPC), args.Error(1)
}

func (m *MockNPCRepository) GetByGameSession(ctx context.Context, gameSessionID string) ([]*models.NPC, error) {
	args := m.Called(ctx, gameSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NPC), args.Error(1)
}

func (m *MockNPCRepository) Update(ctx context.Context, npc *models.NPC) error {
	args := m.Called(ctx, npc)
	return args.Error(0)
}

func (m *MockNPCRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNPCRepository) Search(ctx context.Context, filter *models.NPCSearchFilter) ([]*models.NPC, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NPC), args.Error(1)
}

func (m *MockNPCRepository) GetTemplates(ctx context.Context) ([]*models.NPCTemplate, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NPCTemplate), args.Error(1)
}

func (m *MockNPCRepository) GetTemplateByID(ctx context.Context, id string) (*models.NPCTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NPCTemplate), args.Error(1)
}

func (m *MockNPCRepository) CreateFromTemplate(ctx context.Context, templateID, gameSessionID, createdBy string) (*models.NPC, error) {
	args := m.Called(ctx, templateID, gameSessionID, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NPC), args.Error(1)
}

// MockCustomRaceRepository is a mock implementation of database.CustomRaceRepository
type MockCustomRaceRepository struct {
	mock.Mock
}

func (m *MockCustomRaceRepository) Create(ctx context.Context, race *models.CustomRace) error {
	// Create a new custom race record
	args := m.Called(ctx, race)
	return args.Error(0)
}

func (m *MockCustomRaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CustomRace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CustomRace), args.Error(1)
}

func (m *MockCustomRaceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.CustomRace, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CustomRace), args.Error(1)
}

func (m *MockCustomRaceRepository) GetPublicRaces(ctx context.Context) ([]*models.CustomRace, error) {
	// Get all publicly available custom races
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CustomRace), args.Error(1)
}

func (m *MockCustomRaceRepository) GetPendingApproval(ctx context.Context) ([]*models.CustomRace, error) {
	// Get custom races awaiting approval
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CustomRace), args.Error(1)
}

func (m *MockCustomRaceRepository) Update(ctx context.Context, race *models.CustomRace) error {
	args := m.Called(ctx, race)
	return args.Error(0)
}

func (m *MockCustomRaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomRaceRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return handleErrorReturn(args, 0)
}

// MockCampaignRepository is a mock implementation of database.CampaignRepository
type MockCampaignRepository struct {
	mock.Mock
}

func (m *MockCampaignRepository) CreateStoryArc(arc *models.StoryArc) error {
	args := m.Called(arc)
	return handleErrorReturn(args, 0)
}

func (m *MockCampaignRepository) GetStoryArc(id uuid.UUID) (*models.StoryArc, error) {
	args := m.Called(id)
	return handleSingleReturn[models.StoryArc](args, 0, 1)
}

func (m *MockCampaignRepository) GetStoryArcsBySession(sessionID uuid.UUID) ([]*models.StoryArc, error) {
	args := m.Called(sessionID)
	return handleSliceReturn[models.StoryArc](args, 0, 1)
}

func (m *MockCampaignRepository) UpdateStoryArc(id uuid.UUID, updates map[string]interface{}) error {
	// Update story arc attributes
	args := m.Called(id, updates)
	return handleErrorReturn(args, 0)
}

func (m *MockCampaignRepository) DeleteStoryArc(id uuid.UUID) error {
	args := m.Called(id)
	return handleErrorReturn(args, 0)
}

func (m *MockCampaignRepository) CreateSessionMemory(memory *models.SessionMemory) error {
	args := m.Called(memory)
	return handleErrorReturn(args, 0)
}

func (m *MockCampaignRepository) GetSessionMemory(id uuid.UUID) (*models.SessionMemory, error) {
	args := m.Called(id)
	return handleSingleReturn[models.SessionMemory](args, 0, 1)
}

func (m *MockCampaignRepository) GetSessionMemories(sessionID uuid.UUID, limit int) ([]*models.SessionMemory, error) {
	args := m.Called(sessionID, limit)
	return handleSliceReturn[models.SessionMemory](args, 0, 1)
}

func (m *MockCampaignRepository) GetLatestSessionMemory(sessionID uuid.UUID) (*models.SessionMemory, error) {
	args := m.Called(sessionID)
	return handleSingleReturn[models.SessionMemory](args, 0, 1)
}

func (m *MockCampaignRepository) UpdateSessionMemory(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return handleErrorReturn(args, 0)
}

func (m *MockCampaignRepository) CreatePlotThread(thread *models.PlotThread) error {
	args := m.Called(thread)
	return handleErrorReturn(args, 0)
}

func (m *MockCampaignRepository) GetPlotThread(id uuid.UUID) (*models.PlotThread, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlotThread), args.Error(1)
}

func (m *MockCampaignRepository) GetPlotThreadsBySession(sessionID uuid.UUID) ([]*models.PlotThread, error) {
	// Get all plot threads for a session
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlotThread), args.Error(1)
}

func (m *MockCampaignRepository) GetActivePlotThreads(sessionID uuid.UUID) ([]*models.PlotThread, error) {
	// Get only active plot threads for a session
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlotThread), args.Error(1)
}

func (m *MockCampaignRepository) UpdatePlotThread(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockCampaignRepository) DeletePlotThread(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCampaignRepository) CreateForeshadowingElement(element *models.ForeshadowingElement) error {
	args := m.Called(element)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetForeshadowingElement(id uuid.UUID) (*models.ForeshadowingElement, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ForeshadowingElement), args.Error(1)
}

func (m *MockCampaignRepository) GetUnrevealedForeshadowing(sessionID uuid.UUID) ([]*models.ForeshadowingElement, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ForeshadowingElement), args.Error(1)
}

func (m *MockCampaignRepository) RevealForeshadowing(id uuid.UUID, sessionNumber int) error {
	args := m.Called(id, sessionNumber)
	return args.Error(0)
}

func (m *MockCampaignRepository) CreateTimelineEvent(event *models.CampaignTimeline) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetTimelineEvents(sessionID uuid.UUID, startDate, endDate time.Time) ([]*models.CampaignTimeline, error) {
	args := m.Called(sessionID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CampaignTimeline), args.Error(1)
}

func (m *MockCampaignRepository) CreateOrUpdateNPCRelationship(relationship *models.NPCRelationship) error {
	args := m.Called(relationship)
	return args.Error(0)
}

func (m *MockCampaignRepository) GetNPCRelationships(sessionID, npcID uuid.UUID) ([]*models.NPCRelationship, error) {
	args := m.Called(sessionID, npcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NPCRelationship), args.Error(1)
}

func (m *MockCampaignRepository) UpdateRelationshipScore(sessionID, npcID, targetID uuid.UUID, scoreDelta int) error {
	args := m.Called(sessionID, npcID, targetID, scoreDelta)
	return args.Error(0)
}

// MockRuleBuilderRepository is a mock implementation of database.RuleBuilderRepository
type MockRuleBuilderRepository struct {
	mock.Mock
}

func (m *MockRuleBuilderRepository) GetRuleTemplates(userID, category string, isPublic bool) ([]models.RuleTemplate, error) {
	args := m.Called(userID, category, isPublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RuleTemplate), args.Error(1)
}

func (m *MockRuleBuilderRepository) GetRuleTemplate(templateID string) (*models.RuleTemplate, error) {
	args := m.Called(templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RuleTemplate), args.Error(1)
}

func (m *MockRuleBuilderRepository) CreateRuleTemplate(template *models.RuleTemplate) error {
	// Create a new rule template
	args := m.Called(template)
	return args.Error(0)
}

func (m *MockRuleBuilderRepository) UpdateRuleTemplate(templateID string, updates map[string]interface{}) error {
	args := m.Called(templateID, updates)
	return args.Error(0)
}

func (m *MockRuleBuilderRepository) DeleteRuleTemplate(templateID string) error {
	// Delete existing rule template by ID
	args := m.Called(templateID)
	return args.Error(0)
}

func (m *MockRuleBuilderRepository) GetNodeTemplates() ([]models.NodeTemplate, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.NodeTemplate), args.Error(1)
}

func (m *MockRuleBuilderRepository) GetRuleInstance(instanceID string) (*models.RuleInstance, error) {
	args := m.Called(instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RuleInstance), args.Error(1)
}

func (m *MockRuleBuilderRepository) DeactivateRuleInstance(instanceID string) error {
	args := m.Called(instanceID)
	return args.Error(0)
}

func (m *MockRuleBuilderRepository) CreateActiveRule(rule *models.ActiveRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockRuleBuilderRepository) GetActiveRules(gameSessionID, characterID string) ([]models.ActiveRule, error) {
	args := m.Called(gameSessionID, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActiveRule), args.Error(1)
}

func (m *MockRuleBuilderRepository) GetRuleExecutionHistory(gameSessionID, characterID string, limit int) ([]models.RuleExecution, error) {
	args := m.Called(gameSessionID, characterID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RuleExecution), args.Error(1)
}

func (m *MockRuleBuilderRepository) IncrementUsageCount(templateID string) error {
	args := m.Called(templateID)
	return args.Error(0)
}

// MockDMAssistantRepository is a mock implementation of database.DMAssistantRepository
type MockDMAssistantRepository struct {
	mock.Mock
}

func (m *MockDMAssistantRepository) SaveHistory(ctx context.Context, history *models.DMAssistantHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) SaveNPC(ctx context.Context, npc *models.AINPC) error {
	args := m.Called(ctx, npc)
	return args.Error(0)
}

func (m *MockDMAssistantRepository) GetNPCByID(ctx context.Context, id uuid.UUID) (*models.AINPC, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AINPC), args.Error(1)
}

func (m *MockDMAssistantRepository) GetNPCsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AINPC, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AINPC), args.Error(1)
}

func (m *MockDMAssistantRepository) AddNPCDialog(ctx context.Context, npcID uuid.UUID, entry models.DialogEntry) error {
	args := m.Called(ctx, npcID, entry)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) SaveLocation(ctx context.Context, location *models.AILocation) error {
	// Save AI-generated location data
	args := m.Called(ctx, location)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) GetLocationByID(ctx context.Context, id uuid.UUID) (*models.AILocation, error) {
	args := m.Called(ctx, id)
	return handleSingleReturn[models.AILocation](args, 0, 1)
}

func (m *MockDMAssistantRepository) GetLocationsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AILocation, error) {
	args := m.Called(ctx, sessionID)
	return handleSliceReturn[models.AILocation](args, 0, 1)
}

func (m *MockDMAssistantRepository) SaveStoryElement(ctx context.Context, element *models.AIStoryElement) error {
	// Save AI-generated story element
	args := m.Called(ctx, element)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) GetUnusedStoryElements(ctx context.Context, sessionID uuid.UUID) ([]*models.AIStoryElement, error) {
	args := m.Called(ctx, sessionID)
	return handleSliceReturn[models.AIStoryElement](args, 0, 1)
}

func (m *MockDMAssistantRepository) MarkStoryElementUsed(ctx context.Context, elementID uuid.UUID) error {
	args := m.Called(ctx, elementID)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) SaveNarration(ctx context.Context, narration *models.AINarration) error {
	args := m.Called(ctx, narration)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) SaveEnvironmentalHazard(ctx context.Context, hazard *models.AIEnvironmentalHazard) error {
	args := m.Called(ctx, hazard)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) GetActiveHazardsByLocation(ctx context.Context, locationID uuid.UUID) ([]*models.AIEnvironmentalHazard, error) {
	args := m.Called(ctx, locationID)
	return handleSliceReturn[models.AIEnvironmentalHazard](args, 0, 1)
}

func (m *MockDMAssistantRepository) TriggerHazard(ctx context.Context, hazardID uuid.UUID) error {
	args := m.Called(ctx, hazardID)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) GetHistoryBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.DMAssistantHistory, error) {
	args := m.Called(ctx, sessionID, limit)
	return handleSliceReturn[models.DMAssistantHistory](args, 0, 1)
}

func (m *MockDMAssistantRepository) UpdateNPC(ctx context.Context, npc *models.AINPC) error {
	args := m.Called(ctx, npc)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) UpdateLocation(ctx context.Context, location *models.AILocation) error {
	args := m.Called(ctx, location)
	return handleErrorReturn(args, 0)
}

func (m *MockDMAssistantRepository) GetNarrationsByType(ctx context.Context, sessionID uuid.UUID, narrationType string) ([]*models.AINarration, error) {
	args := m.Called(ctx, sessionID, narrationType)
	return handleSliceReturn[models.AINarration](args, 0, 1)
}

package database

import (
	"context"
	"time"

	"github.com/your-username/dnd-game/backend/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
}

// CharacterRepository defines the interface for character data operations
type CharacterRepository interface {
	Create(ctx context.Context, character *models.Character) error
	GetByID(ctx context.Context, id string) (*models.Character, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Character, error)
	Update(ctx context.Context, character *models.Character) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.Character, error)
}

// GameSessionRepository defines the interface for game session data operations
type GameSessionRepository interface {
	Create(ctx context.Context, session *models.GameSession) error
	GetByID(ctx context.Context, id string) (*models.GameSession, error)
	GetByDMUserID(ctx context.Context, dmUserID string) ([]*models.GameSession, error)
	GetByParticipantUserID(ctx context.Context, userID string) ([]*models.GameSession, error)
	Update(ctx context.Context, session *models.GameSession) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.GameSession, error)
	
	// Participant management
	AddParticipant(ctx context.Context, sessionID, userID string, characterID *string) error
	RemoveParticipant(ctx context.Context, sessionID, userID string) error
	GetParticipants(ctx context.Context, sessionID string) ([]*models.GameParticipant, error)
	UpdateParticipantOnlineStatus(ctx context.Context, sessionID, userID string, isOnline bool) error
}

// DiceRollRepository defines the interface for dice roll data operations
type DiceRollRepository interface {
	Create(ctx context.Context, roll *models.DiceRoll) error
	GetByID(ctx context.Context, id string) (*models.DiceRoll, error)
	GetByGameSession(ctx context.Context, sessionID string, offset, limit int) ([]*models.DiceRoll, error)
	GetByUser(ctx context.Context, userID string, offset, limit int) ([]*models.DiceRoll, error)
	GetByGameSessionAndUser(ctx context.Context, sessionID, userID string, offset, limit int) ([]*models.DiceRoll, error)
	Delete(ctx context.Context, id string) error
}

// InventoryRepository defines the interface for inventory data operations
type InventoryRepository interface {
	// Item operations
	CreateItem(item *models.Item) error
	GetItem(itemID string) (*models.Item, error)
	GetItemsByType(itemType models.ItemType) ([]*models.Item, error)
	
	// Inventory operations
	AddItemToInventory(characterID, itemID string, quantity int) error
	RemoveItemFromInventory(characterID, itemID string, quantity int) error
	GetCharacterInventory(characterID string) ([]*models.InventoryItem, error)
	EquipItem(characterID, itemID string, equip bool) error
	AttuneItem(characterID, itemID string) error
	UnattuneItem(characterID, itemID string) error
	
	// Currency operations
	GetCharacterCurrency(characterID string) (*models.Currency, error)
	CreateCharacterCurrency(currency *models.Currency) error
	UpdateCharacterCurrency(currency *models.Currency) error
	
	// Weight operations
	GetCharacterWeight(characterID string) (*models.InventoryWeight, error)
}

// RefreshTokenRepository defines the interface for refresh token data operations
type RefreshTokenRepository interface {
	Create(userID, tokenID string, token string, expiresAt time.Time) error
	ValidateAndGet(token string) (*RefreshToken, error)
	Revoke(tokenID string) error
	RevokeAllForUser(userID string) error
	CleanupExpired() error
}

// NPCRepository defines the interface for NPC data operations
type NPCRepository interface {
	Create(ctx context.Context, npc *models.NPC) error
	GetByID(ctx context.Context, id string) (*models.NPC, error)
	GetByGameSession(ctx context.Context, gameSessionID string) ([]*models.NPC, error)
	Update(ctx context.Context, npc *models.NPC) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, filter models.NPCSearchFilter) ([]*models.NPC, error)
	
	// Template operations
	GetTemplates(ctx context.Context) ([]*models.NPCTemplate, error)
	GetTemplateByID(ctx context.Context, id string) (*models.NPCTemplate, error)
	CreateFromTemplate(ctx context.Context, templateID, gameSessionID, createdBy string) (*models.NPC, error)
}

// Repositories aggregates all repository interfaces
type Repositories struct {
	Users            UserRepository
	Characters       CharacterRepository
	GameSessions     GameSessionRepository
	DiceRolls        DiceRollRepository
	NPCs             NPCRepository
	Inventory        InventoryRepository
	RefreshTokens    RefreshTokenRepository
	CustomRaces      CustomRaceRepository
	CustomClasses    *CustomClassRepository
	DMAssistant      DMAssistantRepository
	Encounters       *EncounterRepository
	Campaign         CampaignRepository
	CombatAnalytics  CombatAnalyticsRepository
	WorldBuilding    *WorldBuildingRepository
	Narrative        *NarrativeRepository
}
package handlers

import (
	"context"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/websocket"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// HandlersV2 holds all HTTP handlers with logging support
type HandlersV2 struct {
	userService         *services.UserService
	characterService    *services.CharacterService
	gameService         *services.GameSessionService
	diceService         *services.DiceRollService
	combatService       *services.CombatService
	npcService          *services.NPCService
	inventoryService    *services.InventoryService
	encounterService    *services.EncounterService
	customRaceService   *services.CustomRaceService
	dmAssistantService  *services.DMAssistantService
	campaignService     *services.CampaignService
	combatAutomation    *services.CombatAutomationService
	combatAnalytics     *services.CombatAnalyticsService
	settlementGen       *services.SettlementGeneratorService
	factionSystem       *services.FactionSystemService
	worldEventEngine    *services.WorldEventEngineService
	economicSim         *services.EconomicSimulatorService
	ruleEngine          *services.RuleEngine
	balanceAnalyzer     *services.AIBalanceAnalyzer
	conditionalReality  *services.ConditionalRealitySystem
	narrativeEngine     *services.NarrativeEngine
	jwtManager          *auth.JWTManager
	refreshTokenService *services.RefreshTokenService
	websocketHub        *websocket.Hub
	log                 *logger.LoggerV2
}

// New creates a new handlers instance with logging
func New(svc *services.Services, hub *websocket.Hub, log *logger.LoggerV2) *HandlersV2 {
	return &HandlersV2{
		userService:         svc.Users,
		characterService:    svc.Characters,
		gameService:         svc.GameSessions,
		diceService:         svc.DiceRolls,
		combatService:       svc.Combat,
		npcService:          svc.NPCs,
		inventoryService:    svc.Inventory,
		encounterService:    svc.Encounters,
		customRaceService:   svc.CustomRaces,
		dmAssistantService:  svc.DMAssistant,
		campaignService:     svc.Campaign,
		combatAutomation:    svc.CombatAutomation,
		combatAnalytics:     svc.CombatAnalytics,
		settlementGen:       svc.SettlementGen,
		factionSystem:       svc.FactionSystem,
		worldEventEngine:    svc.WorldEventEngine,
		economicSim:         svc.EconomicSim,
		ruleEngine:          svc.RuleEngine,
		balanceAnalyzer:     svc.BalanceAnalyzer,
		conditionalReality:  svc.ConditionalReality,
		narrativeEngine:     svc.NarrativeEngine,
		jwtManager:          svc.JWTManager,
		refreshTokenService: svc.RefreshTokens,
		websocketHub:        hub,
		log:                 log.WithOperation("handlers", ""),
	}
}

// Service-specific handler getters with proper logging context

// AuthHandler returns auth handler with logging
func (h *HandlersV2) AuthHandler() *AuthHandlerV2 {
	return &AuthHandlerV2{
		userService:         h.userService,
		jwtManager:          h.jwtManager,
		refreshTokenService: h.refreshTokenService,
		log:                 h.log.WithOperation("auth", "handler"),
	}
}

// CharacterHandler returns character handler with logging
func (h *HandlersV2) CharacterHandler() *CharacterHandlerV2WithLogging {
	return &CharacterHandlerV2WithLogging{
		characterService: h.characterService,
		log:              h.log.WithOperation("character", "handler"),
	}
}

// GameHandler returns game session handler with logging
func (h *HandlersV2) GameHandler() *GameHandlerV2 {
	return &GameHandlerV2{
		gameService:  h.gameService,
		websocketHub: h.websocketHub,
		log:          h.log.WithOperation("game", "handler"),
	}
}

// CombatHandler returns combat handler with logging
func (h *HandlersV2) CombatHandler() *CombatHandlerV2 {
	return &CombatHandlerV2{
		combatService:    h.combatService,
		combatAutomation: h.combatAutomation,
		combatAnalytics:  h.combatAnalytics,
		log:              h.log.WithOperation("combat", "handler"),
	}
}

// DMAssistantHandler returns DM assistant handler with logging
func (h *HandlersV2) DMAssistantHandler() *DMAssistantHandlerV2 {
	return &DMAssistantHandlerV2{
		dmAssistantService: h.dmAssistantService,
		log:                h.log.WithOperation("dm_assistant", "handler"),
	}
}

// InventoryHandler returns inventory handler with logging
func (h *HandlersV2) InventoryHandler() *InventoryHandlerV2 {
	return &InventoryHandlerV2{
		inventoryService: h.inventoryService,
		log:              h.log.WithOperation("inventory", "handler"),
	}
}

// WebSocketHandler returns WebSocket handler with logging
func (h *HandlersV2) WebSocketHandler() *websocket.HandlerV2 {
	return websocket.NewHandlerV2(h.websocketHub, h.jwtManager, h.log.WithOperation("websocket", "handler"))
}

// Example handler structures (implement similarly for all handlers)

// AuthHandlerV2 handles authentication with logging
type AuthHandlerV2 struct {
	userService         *services.UserService
	jwtManager          *auth.JWTManager
	refreshTokenService *services.RefreshTokenService
	log                 *logger.LoggerV2
}

// CharacterHandlerV2WithLogging handles character operations with logging
type CharacterHandlerV2WithLogging struct {
	characterService CharacterService
	log              *logger.LoggerV2
}

// GameHandlerV2 handles game sessions with logging
type GameHandlerV2 struct {
	gameService  *services.GameSessionService
	websocketHub *websocket.Hub
	log          *logger.LoggerV2
}

// CombatHandlerV2 handles combat with logging
type CombatHandlerV2 struct {
	combatService    *services.CombatService
	combatAutomation *services.CombatAutomationService
	combatAnalytics  *services.CombatAnalyticsService
	log              *logger.LoggerV2
}

// DMAssistantHandlerV2 handles DM assistant features with logging
type DMAssistantHandlerV2 struct {
	dmAssistantService *services.DMAssistantService
	log                *logger.LoggerV2
}

// InventoryHandlerV2 handles inventory with logging
type InventoryHandlerV2 struct {
	inventoryService *services.InventoryService
	log              *logger.LoggerV2
}

// CharacterService interface for dependency injection
type CharacterService interface {
	GetAllCharacters(ctx context.Context, userID string) ([]*models.Character, error)
	GetCharacterByID(ctx context.Context, id string) (*models.Character, error)
	CreateCharacter(ctx context.Context, character *models.Character) error
	UpdateCharacter(ctx context.Context, character *models.Character) error
	DeleteCharacter(ctx context.Context, id string) error
	LevelUp(ctx context.Context, characterID string, hitPointIncrease int, attributeIncrease string) (*models.Character, error)
}

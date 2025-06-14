package services

import (
	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
)

// Services aggregates all service interfaces
type Services struct {
	DB                 *database.DB
	Users              *UserService
	Characters         *CharacterService
	GameSessions       *GameSessionService
	DiceRolls          *DiceRollService
	Combat             *CombatService
	NPCs               *NPCService
	Inventory          *InventoryService
	CustomRaces        *CustomRaceService
	DMAssistant        *DMAssistantService
	Encounters         *EncounterService
	Campaign           *CampaignService
	CombatAutomation   *CombatAutomationService
	CombatAnalytics    *CombatAnalyticsService
	SettlementGen      *SettlementGeneratorService
	FactionSystem      *FactionSystemService
	WorldEventEngine   *WorldEventEngineService
	EconomicSim        *EconomicSimulatorService
	RuleEngine         *RuleEngine
	BalanceAnalyzer    *AIBalanceAnalyzer
	ConditionalReality *ConditionalRealitySystem
	JWTManager         *auth.JWTManager
	RefreshTokens      *RefreshTokenService
	Config             *config.Config
	NarrativeEngine    *NarrativeEngine
	WorldBuilding      interface{} // TODO: Add proper world building service
	RuleBuilder        interface{} // TODO: Add proper rule builder service
	AICampaignManager  *AICampaignManager
	BattleMapGen       *AIBattleMapGenerator
}

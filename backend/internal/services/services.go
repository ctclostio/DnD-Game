package services

import (
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/config"
)

// Services aggregates all service interfaces
type Services struct {
	Users         *UserService
	Characters    *CharacterService
	GameSessions  *GameSessionService
	DiceRolls     *DiceRollService
	Combat        *CombatService
	NPCs          *NPCService
	Inventory     *InventoryService
	JWTManager    *auth.JWTManager
	RefreshTokens *RefreshTokenService
	Config        *config.Config
}
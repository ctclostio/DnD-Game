package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// DMAssistantService handles DM assistance operations
type DMAssistantService struct {
	repo        database.DMAssistantRepository
	aiAssistant AIDMAssistantInterface
}

// NewDMAssistantService creates a new DM assistant service
func NewDMAssistantService(repo database.DMAssistantRepository, aiAssistant AIDMAssistantInterface) *DMAssistantService {
	return &DMAssistantService{
		repo:        repo,
		aiAssistant: aiAssistant,
	}
}

// ProcessRequest handles a DM assistant request
func (s *DMAssistantService) ProcessRequest(ctx context.Context, userID uuid.UUID, req models.DMAssistantRequest) (interface{}, error) {
	gameSessionID, err := uuid.Parse(req.GameSessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid game session ID: %w", err)
	}

	// Record the request
	historyEntry := &models.DMAssistantHistory{
		ID:             uuid.New(),
		GameSessionID:  gameSessionID,
		UserID:         userID,
		RequestType:    req.Type,
		RequestContext: req.Context,
		CreatedAt:      time.Now(),
	}

	var result interface{}
	var prompt string

	switch req.Type {
	case models.RequestTypeNPCDialog:
		npcReq, err := s.parseNPCDialogRequest(req.Parameters)
		if err != nil {
			return nil, err
		}
		prompt = fmt.Sprintf("NPC: %s, Player: %s", npcReq.NPCName, npcReq.PlayerInput)

		dialog, err := s.aiAssistant.GenerateNPCDialog(ctx, npcReq)
		if err != nil {
			return nil, err
		}

		historyEntry.Response = dialog
		result = map[string]string{"dialog": dialog}

	case models.RequestTypeLocationDesc:
		locReq, err := s.parseLocationRequest(req.Parameters)
		if err != nil {
			return nil, err
		}
		prompt = fmt.Sprintf("Location: %s (%s)", locReq.LocationName, locReq.LocationType)

		location, err := s.aiAssistant.GenerateLocationDescription(ctx, locReq)
		if err != nil {
			return nil, err
		}

		location.GameSessionID = gameSessionID
		location.CreatedBy = userID

		// Save to database
		if err := s.repo.SaveLocation(ctx, location); err != nil {
			return nil, fmt.Errorf("failed to save location: %w", err)
		}

		historyEntry.Response = location.Description
		result = location

	case models.RequestTypeCombatNarration:
		combatReq := s.parseCombatRequest(req.Parameters)
		prompt = fmt.Sprintf("Combat: %s vs %s", combatReq.AttackerName, combatReq.TargetName)

		narration, err := s.aiAssistant.GenerateCombatNarration(ctx, combatReq)
		if err != nil {
			return nil, err
		}

		// Save narration for reuse
		narrationEntry := &models.AINarration{
			ID:            uuid.New(),
			GameSessionID: gameSessionID,
			Type:          s.getCombatNarrationType(combatReq),
			Context:       req.Context,
			Narration:     narration,
			CreatedBy:     userID,
			CreatedAt:     time.Now(),
		}

		if err := s.repo.SaveNarration(ctx, narrationEntry); err != nil {
			return nil, fmt.Errorf("failed to save narration: %w", err)
		}

		historyEntry.Response = narration
		result = map[string]string{"narration": narration}

	case models.RequestTypePlotTwist:
		plotTwist, err := s.aiAssistant.GeneratePlotTwist(ctx, req.Context)
		if err != nil {
			return nil, err
		}

		plotTwist.GameSessionID = gameSessionID
		plotTwist.CreatedBy = userID
		plotTwist.CreatedAt = time.Now()

		if err := s.repo.SaveStoryElement(ctx, plotTwist); err != nil {
			return nil, fmt.Errorf("failed to save plot twist: %w", err)
		}

		prompt = "Generate plot twist"
		historyEntry.Response = plotTwist.Description
		result = plotTwist

	case models.RequestTypeEnvironmentalHazard:
		locationType, _ := req.Parameters["locationType"].(string)
		difficulty, _ := req.Parameters["difficulty"].(float64)

		hazard, err := s.aiAssistant.GenerateEnvironmentalHazard(ctx, locationType, int(difficulty))
		if err != nil {
			return nil, err
		}

		hazard.GameSessionID = gameSessionID
		hazard.CreatedBy = userID
		hazard.CreatedAt = time.Now()

		if locationID, ok := req.Parameters["locationId"].(string); ok {
			locID, _ := uuid.Parse(locationID)
			hazard.LocationID = &locID
		}

		if err := s.repo.SaveEnvironmentalHazard(ctx, hazard); err != nil {
			return nil, fmt.Errorf("failed to save hazard: %w", err)
		}

		prompt = fmt.Sprintf("Hazard for %s (difficulty %d)", locationType, int(difficulty))
		historyEntry.Response = hazard.Description
		result = hazard

	default:
		return nil, fmt.Errorf("unknown request type: %s", req.Type)
	}

	// Save history
	historyEntry.Prompt = prompt
	if err := s.repo.SaveHistory(ctx, historyEntry); err != nil {
		// Log error but don't fail the request
		logger.WithContext(ctx).WithError(err).Error().Msg("Failed to save history")
	}

	return result, nil
}

// GetNPCByID retrieves an NPC by ID
func (s *DMAssistantService) GetNPCByID(ctx context.Context, npcID uuid.UUID) (*models.AINPC, error) {
	return s.repo.GetNPCByID(ctx, npcID)
}

// GetNPCsBySession retrieves all NPCs for a game session
func (s *DMAssistantService) GetNPCsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AINPC, error) {
	return s.repo.GetNPCsBySession(ctx, sessionID)
}

// GetLocationByID retrieves a location by ID
func (s *DMAssistantService) GetLocationByID(ctx context.Context, locationID uuid.UUID) (*models.AILocation, error) {
	return s.repo.GetLocationByID(ctx, locationID)
}

// GetLocationsBySession retrieves all locations for a game session
func (s *DMAssistantService) GetLocationsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AILocation, error) {
	return s.repo.GetLocationsBySession(ctx, sessionID)
}

// GetUnusedStoryElements retrieves unused story elements for a session
func (s *DMAssistantService) GetUnusedStoryElements(ctx context.Context, sessionID uuid.UUID) ([]*models.AIStoryElement, error) {
	return s.repo.GetUnusedStoryElements(ctx, sessionID)
}

// MarkStoryElementUsed marks a story element as used
func (s *DMAssistantService) MarkStoryElementUsed(ctx context.Context, elementID uuid.UUID) error {
	return s.repo.MarkStoryElementUsed(ctx, elementID)
}

// GetActiveHazards retrieves active environmental hazards
func (s *DMAssistantService) GetActiveHazards(ctx context.Context, locationID uuid.UUID) ([]*models.AIEnvironmentalHazard, error) {
	return s.repo.GetActiveHazardsByLocation(ctx, locationID)
}

// TriggerHazard marks a hazard as triggered
func (s *DMAssistantService) TriggerHazard(ctx context.Context, hazardID uuid.UUID) error {
	return s.repo.TriggerHazard(ctx, hazardID)
}

// Helper methods

func (s *DMAssistantService) parseNPCDialogRequest(params map[string]interface{}) (*models.NPCDialogRequest, error) {
	req := &models.NPCDialogRequest{}

	if name, ok := params["npcName"].(string); ok {
		req.NPCName = name
	} else {
		return nil, fmt.Errorf("npcName is required")
	}

	if personality, ok := params["npcPersonality"].([]interface{}); ok {
		for _, trait := range personality {
			if t, ok := trait.(string); ok {
				req.NPCPersonality = append(req.NPCPersonality, t)
			}
		}
	}

	req.DialogStyle, _ = params["dialogStyle"].(string)
	req.Situation, _ = params["situation"].(string)
	req.PlayerInput, _ = params["playerInput"].(string)
	req.PreviousContext, _ = params["previousContext"].(string)

	return req, nil
}

func (s *DMAssistantService) parseLocationRequest(params map[string]interface{}) (*models.LocationDescriptionRequest, error) {
	req := &models.LocationDescriptionRequest{}

	if locType, ok := params["locationType"].(string); ok {
		req.LocationType = locType
	} else {
		return nil, fmt.Errorf("locationType is required")
	}

	req.LocationName, _ = params["locationName"].(string)
	req.Atmosphere, _ = params["atmosphere"].(string)
	req.TimeOfDay, _ = params["timeOfDay"].(string)
	req.Weather, _ = params["weather"].(string)

	if features, ok := params["specialFeatures"].([]interface{}); ok {
		for _, feature := range features {
			if f, ok := feature.(string); ok {
				req.SpecialFeatures = append(req.SpecialFeatures, f)
			}
		}
	}

	return req, nil
}

func (s *DMAssistantService) parseCombatRequest(params map[string]interface{}) *models.CombatNarrationRequest {
	req := &models.CombatNarrationRequest{}

	req.AttackerName, _ = params["attackerName"].(string)
	req.TargetName, _ = params["targetName"].(string)
	req.ActionType, _ = params["actionType"].(string)
	req.WeaponOrSpell, _ = params["weaponOrSpell"].(string)

	if damage, ok := params["damage"].(float64); ok {
		req.Damage = int(damage)
	}

	req.IsHit, _ = params["isHit"].(bool)
	req.IsCritical, _ = params["isCritical"].(bool)

	if hp, ok := params["targetHP"].(float64); ok {
		req.TargetHP = int(hp)
	}
	if maxHP, ok := params["targetMaxHP"].(float64); ok {
		req.TargetMaxHP = int(maxHP)
	}

	return req
}

func (s *DMAssistantService) getCombatNarrationType(req *models.CombatNarrationRequest) string {
	if !req.IsHit {
		return models.NarrationTypeCombatMiss
	}
	if req.IsCritical {
		return models.NarrationTypeCombatCritical
	}
	if req.TargetHP <= 0 {
		return models.NarrationTypeDeath
	}
	return models.NarrationTypeCombatHit
}

// CreateNPC generates and saves a new NPC
func (s *DMAssistantService) CreateNPC(ctx context.Context, sessionID, userID uuid.UUID, role string, context map[string]interface{}) (*models.AINPC, error) {
	npc, err := s.aiAssistant.GenerateNPC(ctx, role, context)
	if err != nil {
		return nil, fmt.Errorf("failed to generate NPC: %w", err)
	}

	npc.GameSessionID = sessionID
	npc.CreatedBy = userID
	npc.CreatedAt = time.Now()
	npc.UpdatedAt = time.Now()

	if err := s.repo.SaveNPC(ctx, npc); err != nil {
		return nil, fmt.Errorf("failed to save NPC: %w", err)
	}

	return npc, nil
}

// UpdateNPCDialog adds new dialog to an NPC's history
func (s *DMAssistantService) UpdateNPCDialog(ctx context.Context, npcID uuid.UUID, dialog, context string) error {
	entry := models.DialogEntry{
		Context:   context,
		Dialog:    dialog,
		Timestamp: time.Now(),
	}

	return s.repo.AddNPCDialog(ctx, npcID, entry)
}

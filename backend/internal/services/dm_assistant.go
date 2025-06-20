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

	// Create history entry
	historyEntry := s.createHistoryEntry(gameSessionID, userID, req)

	// Process request based on type
	result, prompt, err := s.processRequestByType(ctx, gameSessionID, userID, req)
	if err != nil {
		return nil, err
	}

	// Update history entry with response
	historyEntry.Prompt = prompt
	s.extractResponseForHistory(historyEntry, result)

	// Save history
	s.saveRequestHistory(ctx, historyEntry)

	return result, nil
}

// createHistoryEntry creates a new history entry for the request
func (s *DMAssistantService) createHistoryEntry(gameSessionID, userID uuid.UUID, req models.DMAssistantRequest) *models.DMAssistantHistory {
	return &models.DMAssistantHistory{
		ID:             uuid.New(),
		GameSessionID:  gameSessionID,
		UserID:         userID,
		RequestType:    req.Type,
		RequestContext: req.Context,
		CreatedAt:      time.Now(),
	}
}

// processRequestByType handles the request based on its type
func (s *DMAssistantService) processRequestByType(ctx context.Context, gameSessionID, userID uuid.UUID, req models.DMAssistantRequest) (interface{}, string, error) {
	switch req.Type {
	case models.RequestTypeNPCDialog:
		return s.handleNPCDialog(ctx, req)
	case models.RequestTypeLocationDesc:
		return s.handleLocationDescription(ctx, gameSessionID, userID, req)
	case models.RequestTypeCombatNarration:
		return s.handleCombatNarration(ctx, gameSessionID, userID, req)
	case models.RequestTypePlotTwist:
		return s.handlePlotTwist(ctx, gameSessionID, userID, req)
	case models.RequestTypeEnvironmentalHazard:
		return s.handleEnvironmentalHazard(ctx, gameSessionID, userID, req)
	default:
		return nil, "", fmt.Errorf("unknown request type: %s", req.Type)
	}
}

// handleNPCDialog processes NPC dialog generation requests
func (s *DMAssistantService) handleNPCDialog(ctx context.Context, req models.DMAssistantRequest) (interface{}, string, error) {
	npcReq, err := s.parseNPCDialogRequest(req.Parameters)
	if err != nil {
		return nil, "", err
	}

	prompt := fmt.Sprintf("NPC: %s, Player: %s", npcReq.NPCName, npcReq.PlayerInput)
	
	dialog, err := s.aiAssistant.GenerateNPCDialog(ctx, npcReq)
	if err != nil {
		return nil, "", err
	}

	result := map[string]string{"dialog": dialog}
	return result, prompt, nil
}

// handleLocationDescription processes location description generation requests
func (s *DMAssistantService) handleLocationDescription(ctx context.Context, gameSessionID, userID uuid.UUID, req models.DMAssistantRequest) (interface{}, string, error) {
	locReq, err := s.parseLocationRequest(req.Parameters)
	if err != nil {
		return nil, "", err
	}

	prompt := fmt.Sprintf("Location: %s (%s)", locReq.LocationName, locReq.LocationType)

	location, err := s.aiAssistant.GenerateLocationDescription(ctx, locReq)
	if err != nil {
		return nil, "", err
	}

	location.GameSessionID = gameSessionID
	location.CreatedBy = userID

	// Save to database
	if err := s.repo.SaveLocation(ctx, location); err != nil {
		return nil, "", fmt.Errorf("failed to save location: %w", err)
	}

	return location, prompt, nil
}

// handleCombatNarration processes combat narration generation requests
func (s *DMAssistantService) handleCombatNarration(ctx context.Context, gameSessionID, userID uuid.UUID, req models.DMAssistantRequest) (interface{}, string, error) {
	combatReq := s.parseCombatRequest(req.Parameters)
	prompt := fmt.Sprintf("Combat: %s vs %s", combatReq.AttackerName, combatReq.TargetName)

	narration, err := s.aiAssistant.GenerateCombatNarration(ctx, combatReq)
	if err != nil {
		return nil, "", err
	}

	// Save narration for reuse
	narrationEntry := s.createNarrationEntry(gameSessionID, userID, req.Context, combatReq, narration)

	if err := s.repo.SaveNarration(ctx, narrationEntry); err != nil {
		return nil, "", fmt.Errorf("failed to save narration: %w", err)
	}

	result := map[string]string{"narration": narration}
	return result, prompt, nil
}

// createNarrationEntry creates a narration entry
func (s *DMAssistantService) createNarrationEntry(gameSessionID, userID uuid.UUID, context map[string]interface{}, combatReq *models.CombatNarrationRequest, narration string) *models.AINarration {
	return &models.AINarration{
		ID:            uuid.New(),
		GameSessionID: gameSessionID,
		Type:          s.getCombatNarrationType(combatReq),
		Context:       context,
		Narration:     narration,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
	}
}

// handlePlotTwist processes plot twist generation requests
func (s *DMAssistantService) handlePlotTwist(ctx context.Context, gameSessionID, userID uuid.UUID, req models.DMAssistantRequest) (interface{}, string, error) {
	plotTwist, err := s.aiAssistant.GeneratePlotTwist(ctx, req.Context)
	if err != nil {
		return nil, "", err
	}

	plotTwist.GameSessionID = gameSessionID
	plotTwist.CreatedBy = userID
	plotTwist.CreatedAt = time.Now()

	if err := s.repo.SaveStoryElement(ctx, plotTwist); err != nil {
		return nil, "", fmt.Errorf("failed to save plot twist: %w", err)
	}

	prompt := "Generate plot twist"
	return plotTwist, prompt, nil
}

// handleEnvironmentalHazard processes environmental hazard generation requests
func (s *DMAssistantService) handleEnvironmentalHazard(ctx context.Context, gameSessionID, userID uuid.UUID, req models.DMAssistantRequest) (interface{}, string, error) {
	locationType, _ := req.Parameters["locationType"].(string)
	difficulty, _ := req.Parameters["difficulty"].(float64)

	hazard, err := s.aiAssistant.GenerateEnvironmentalHazard(ctx, locationType, int(difficulty))
	if err != nil {
		return nil, "", err
	}

	s.setHazardProperties(hazard, gameSessionID, userID, req.Parameters)

	if err := s.repo.SaveEnvironmentalHazard(ctx, hazard); err != nil {
		return nil, "", fmt.Errorf("failed to save hazard: %w", err)
	}

	prompt := fmt.Sprintf("Hazard for %s (difficulty %d)", locationType, int(difficulty))
	return hazard, prompt, nil
}

// setHazardProperties sets common properties on a hazard
func (s *DMAssistantService) setHazardProperties(hazard *models.AIEnvironmentalHazard, gameSessionID, userID uuid.UUID, parameters map[string]interface{}) {
	hazard.GameSessionID = gameSessionID
	hazard.CreatedBy = userID
	hazard.CreatedAt = time.Now()

	if locationID, ok := parameters["locationId"].(string); ok {
		locID, _ := uuid.Parse(locationID)
		hazard.LocationID = &locID
	}
}

// extractResponseForHistory extracts the response text from various result types
func (s *DMAssistantService) extractResponseForHistory(historyEntry *models.DMAssistantHistory, result interface{}) {
	switch v := result.(type) {
	case map[string]string:
		if narration, exists := v["narration"]; exists {
			historyEntry.Response = narration
		} else if dialog, exists := v["dialog"]; exists {
			historyEntry.Response = dialog
		}
	case *models.AILocation:
		historyEntry.Response = v.Description
	case *models.AIStoryElement:
		historyEntry.Response = v.Description
	case *models.AIEnvironmentalHazard:
		historyEntry.Response = v.Description
	}
}

// saveRequestHistory saves the history entry with response information
func (s *DMAssistantService) saveRequestHistory(ctx context.Context, historyEntry *models.DMAssistantHistory) {
	if err := s.repo.SaveHistory(ctx, historyEntry); err != nil {
		// Log error but don't fail the request
		logger.WithContext(ctx).WithError(err).Error().Msg("Failed to save history")
	}
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

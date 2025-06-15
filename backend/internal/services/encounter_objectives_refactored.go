package services

import (
	"context"
	"fmt"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// ObjectiveEvaluator defines the interface for evaluating objective completion
type ObjectiveEvaluator interface {
	// Evaluate checks if the objective is completed or failed
	Evaluate(ctx context.Context, objective *models.EncounterObjective, encounter *models.Encounter) (completed bool, failed bool, err error)
	// GetType returns the objective type this evaluator handles
	GetType() string
}

// ObjectiveEvaluatorRegistry manages all objective evaluators
type ObjectiveEvaluatorRegistry struct {
	evaluators map[string]ObjectiveEvaluator
}

// NewObjectiveEvaluatorRegistry creates a new registry with all evaluators
func NewObjectiveEvaluatorRegistry() *ObjectiveEvaluatorRegistry {
	registry := &ObjectiveEvaluatorRegistry{
		evaluators: make(map[string]ObjectiveEvaluator),
	}
	
	// Register all evaluators
	registry.Register(NewDefeatAllEvaluator())
	registry.Register(NewSurviveRoundsEvaluator())
	registry.Register(NewProtectNPCEvaluator())
	registry.Register(NewReachLocationEvaluator())
	registry.Register(NewCollectItemsEvaluator())
	registry.Register(NewDefeatSpecificEvaluator())
	registry.Register(NewPreventActionEvaluator())
	
	return registry
}

// Register adds a new evaluator to the registry
func (r *ObjectiveEvaluatorRegistry) Register(evaluator ObjectiveEvaluator) {
	r.evaluators[evaluator.GetType()] = evaluator
}

// GetEvaluator returns the evaluator for a specific objective type
func (r *ObjectiveEvaluatorRegistry) GetEvaluator(objectiveType string) (ObjectiveEvaluator, bool) {
	evaluator, exists := r.evaluators[objectiveType]
	return evaluator, exists
}

// DefeatAllEvaluator checks if all enemies are defeated
type DefeatAllEvaluator struct{}

func NewDefeatAllEvaluator() *DefeatAllEvaluator {
	return &DefeatAllEvaluator{}
}

func (e *DefeatAllEvaluator) GetType() string {
	return constants.ObjectiveDefeatAll
}

func (e *DefeatAllEvaluator) Evaluate(ctx context.Context, objective *models.EncounterObjective, encounter *models.Encounter) (bool, bool, error) {
	for i := range encounter.Enemies {
		enemy := &encounter.Enemies[i]
		if enemy.IsAlive && !enemy.Fled {
			return false, false, nil // Not completed, not failed
		}
	}
	return true, false, nil // All defeated - completed
}

// SurviveRoundsEvaluator checks if the party survived a certain number of rounds
type SurviveRoundsEvaluator struct {
	roundTracker RoundTracker
}

type RoundTracker interface {
	GetCurrentRound(encounterID string) int
}

func NewSurviveRoundsEvaluator() *SurviveRoundsEvaluator {
	return &SurviveRoundsEvaluator{}
}

func (e *SurviveRoundsEvaluator) GetType() string {
	return "survive_rounds"
}

func (e *SurviveRoundsEvaluator) Evaluate(ctx context.Context, objective *models.EncounterObjective, encounter *models.Encounter) (bool, bool, error) {
	targetRounds, ok := objective.SuccessConditions["target_rounds"].(int)
	if !ok {
		return false, false, fmt.Errorf("survive_rounds objective missing target_rounds condition")
	}
	
	currentRound := encounter.CurrentRound
	if e.roundTracker != nil {
		currentRound = e.roundTracker.GetCurrentRound(encounter.ID)
	}
	
	// Check if party is defeated
	allDefeated := true
	for i := range encounter.PartyMembers {
		if encounter.PartyMembers[i].IsAlive {
			allDefeated = false
			break
		}
	}
	
	if allDefeated {
		return false, true, nil // Failed - party defeated
	}
	
	return currentRound >= targetRounds, false, nil
}

// ProtectNPCEvaluator checks if a specific NPC is still alive
type ProtectNPCEvaluator struct {
	npcService NPCServiceInterface
}

type NPCServiceInterface interface {
	GetNPCStatus(ctx context.Context, npcID string) (*models.NPCStatus, error)
}

func NewProtectNPCEvaluator() *ProtectNPCEvaluator {
	return &ProtectNPCEvaluator{}
}

func (e *ProtectNPCEvaluator) GetType() string {
	return "protect_npc"
}

func (e *ProtectNPCEvaluator) Evaluate(ctx context.Context, objective *models.EncounterObjective, encounter *models.Encounter) (bool, bool, error) {
	npcID, ok := objective.SuccessConditions["npc_id"].(string)
	if !ok {
		return false, false, fmt.Errorf("protect_npc objective missing npc_id condition")
	}
	
	// Check in encounter NPCs first
	for i := range encounter.NPCs {
		if encounter.NPCs[i].ID == npcID {
			if !encounter.NPCs[i].IsAlive {
				return false, true, nil // Failed - NPC dead
			}
			// Still alive, check if encounter is complete
			return encounter.Status == "completed", false, nil
		}
	}
	
	// If not in encounter, check with NPC service
	if e.npcService != nil {
		status, err := e.npcService.GetNPCStatus(ctx, npcID)
		if err != nil {
			return false, false, err
		}
		if !status.IsAlive {
			return false, true, nil // Failed
		}
	}
	
	return encounter.Status == "completed", false, nil
}

// ReachLocationEvaluator checks if the party reached a specific location
type ReachLocationEvaluator struct{}

func NewReachLocationEvaluator() *ReachLocationEvaluator {
	return &ReachLocationEvaluator{}
}

func (e *ReachLocationEvaluator) GetType() string {
	return "reach_location"
}

func (e *ReachLocationEvaluator) Evaluate(ctx context.Context, objective *models.EncounterObjective, encounter *models.Encounter) (bool, bool, error) {
	targetLocation, ok := objective.SuccessConditions["location_id"].(string)
	if !ok {
		return false, false, fmt.Errorf("reach_location objective missing location_id condition")
	}
	
	// Check if any party member reached the location
	for i := range encounter.PartyMembers {
		if encounter.PartyMembers[i].CurrentLocation == targetLocation {
			return true, false, nil
		}
	}
	
	return false, false, nil
}

// CollectItemsEvaluator checks if specific items were collected
type CollectItemsEvaluator struct{}

func NewCollectItemsEvaluator() *CollectItemsEvaluator {
	return &CollectItemsEvaluator{}
}

func (e *CollectItemsEvaluator) GetType() string {
	return "collect_items"
}

func (e *CollectItemsEvaluator) Evaluate(ctx context.Context, objective *models.EncounterObjective, encounter *models.Encounter) (bool, bool, error) {
	requiredItems, ok := objective.SuccessConditions["items"].([]string)
	if !ok {
		return false, false, fmt.Errorf("collect_items objective missing items condition")
	}
	
	collectedItems := encounter.CollectedItems
	for _, required := range requiredItems {
		found := false
		for _, collected := range collectedItems {
			if collected == required {
				found = true
				break
			}
		}
		if !found {
			return false, false, nil // Not all items collected yet
		}
	}
	
	return true, false, nil // All items collected
}

// DefeatSpecificEvaluator checks if specific enemies were defeated
type DefeatSpecificEvaluator struct{}

func NewDefeatSpecificEvaluator() *DefeatSpecificEvaluator {
	return &DefeatSpecificEvaluator{}
}

func (e *DefeatSpecificEvaluator) GetType() string {
	return "defeat_specific"
}

func (e *DefeatSpecificEvaluator) Evaluate(ctx context.Context, objective *models.EncounterObjective, encounter *models.Encounter) (bool, bool, error) {
	targets, ok := objective.SuccessConditions["target_ids"].([]string)
	if !ok {
		return false, false, fmt.Errorf("defeat_specific objective missing target_ids condition")
	}
	
	for _, targetID := range targets {
		found := false
		defeated := false
		
		for i := range encounter.Enemies {
			if encounter.Enemies[i].ID == targetID {
				found = true
				defeated = !encounter.Enemies[i].IsAlive || encounter.Enemies[i].Fled
				break
			}
		}
		
		if !found || !defeated {
			return false, false, nil // Target not found or not defeated
		}
	}
	
	return true, false, nil // All targets defeated
}

// PreventActionEvaluator checks if a specific action was prevented
type PreventActionEvaluator struct{}

func NewPreventActionEvaluator() *PreventActionEvaluator {
	return &PreventActionEvaluator{}
}

func (e *PreventActionEvaluator) GetType() string {
	return "prevent_action"
}

func (e *PreventActionEvaluator) Evaluate(ctx context.Context, objective *models.EncounterObjective, encounter *models.Encounter) (bool, bool, error) {
	preventedAction, ok := objective.SuccessConditions["action"].(string)
	if !ok {
		return false, false, fmt.Errorf("prevent_action objective missing action condition")
	}
	
	// Check if the action occurred
	for _, event := range encounter.Events {
		if event.Type == preventedAction {
			return false, true, nil // Failed - action occurred
		}
	}
	
	// If encounter is complete and action didn't occur, success
	if encounter.Status == "completed" {
		return true, false, nil
	}
	
	return false, false, nil // Still in progress
}

// ObjectiveManager handles objective evaluation using the registry
type ObjectiveManager struct {
	evaluatorRegistry *ObjectiveEvaluatorRegistry
	repository        EncounterRepositoryInterface
	rewardService     RewardServiceInterface
}

type EncounterRepositoryInterface interface {
	GetObjectives(encounterID string) ([]*models.EncounterObjective, error)
	GetByID(encounterID string) (*models.Encounter, error)
	CompleteObjective(objectiveID string) error
	FailObjective(objectiveID string) error
}

type RewardServiceInterface interface {
	AwardObjectiveRewards(ctx context.Context, objective *models.EncounterObjective, sessionID string) error
}

// NewObjectiveManager creates a new objective manager
func NewObjectiveManager(
	repo EncounterRepositoryInterface,
	rewardService RewardServiceInterface,
) *ObjectiveManager {
	return &ObjectiveManager{
		evaluatorRegistry: NewObjectiveEvaluatorRegistry(),
		repository:        repo,
		rewardService:     rewardService,
	}
}

// CheckObjectives is the refactored version of the original function
func (m *ObjectiveManager) CheckObjectives(ctx context.Context, encounterID string) error {
	objectives, err := m.repository.GetObjectives(encounterID)
	if err != nil {
		return fmt.Errorf("failed to get objectives: %w", err)
	}
	
	encounter, err := m.repository.GetByID(encounterID)
	if err != nil {
		return fmt.Errorf("encounter not found: %w", err)
	}
	
	for _, objective := range objectives {
		if err := m.evaluateObjective(ctx, objective, encounter); err != nil {
			// Log error but continue checking other objectives
			continue
		}
	}
	
	return nil
}

// evaluateObjective checks a single objective
func (m *ObjectiveManager) evaluateObjective(
	ctx context.Context,
	objective *models.EncounterObjective,
	encounter *models.Encounter,
) error {
	// Skip already completed or failed objectives
	if objective.IsCompleted || objective.IsFailed {
		return nil
	}
	
	// Get the appropriate evaluator
	evaluator, exists := m.evaluatorRegistry.GetEvaluator(objective.Type)
	if !exists {
		return fmt.Errorf("unknown objective type: %s", objective.Type)
	}
	
	// Evaluate the objective
	completed, failed, err := evaluator.Evaluate(ctx, objective, encounter)
	if err != nil {
		return fmt.Errorf("failed to evaluate objective %s: %w", objective.ID, err)
	}
	
	// Update objective status
	if completed {
		if err := m.repository.CompleteObjective(objective.ID); err != nil {
			return fmt.Errorf("failed to complete objective: %w", err)
		}
		
		if err := m.rewardService.AwardObjectiveRewards(ctx, objective, encounter.GameSessionID); err != nil {
			// Log error but don't fail the whole operation
		}
	} else if failed {
		if err := m.repository.FailObjective(objective.ID); err != nil {
			return fmt.Errorf("failed to fail objective: %w", err)
		}
	}
	
	return nil
}

// RegisterCustomEvaluator allows adding custom objective types
func (m *ObjectiveManager) RegisterCustomEvaluator(evaluator ObjectiveEvaluator) {
	m.evaluatorRegistry.Register(evaluator)
}

// Example of how to use in the EncounterService
func (s *EncounterService) CheckObjectivesRefactored(ctx context.Context, encounterID string) error {
	manager := NewObjectiveManager(s.repo, s)
	return manager.CheckObjectives(ctx, encounterID)
}
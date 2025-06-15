package services

import (
	"context"
	"fmt"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

type EncounterService struct {
	repo             *database.EncounterRepository
	encounterBuilder *AIEncounterBuilder
	combatService    *CombatService
}

func NewEncounterService(repo *database.EncounterRepository, builder *AIEncounterBuilder, combat *CombatService) *EncounterService {
	return &EncounterService{
		repo:             repo,
		encounterBuilder: builder,
		combatService:    combat,
	}
}

// GenerateEncounter creates a new AI-generated encounter
func (s *EncounterService) GenerateEncounter(ctx context.Context, req *EncounterRequest, gameSessionID, userID string) (*models.Encounter, error) {
	// Generate the encounter using AI
	encounter, err := s.encounterBuilder.GenerateEncounter(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encounter: %w", err)
	}

	// Set metadata
	encounter.GameSessionID = gameSessionID
	encounter.CreatedBy = userID
	encounter.Status = constants.EncounterStatusPlanned

	// Save to database
	if err := s.repo.Create(encounter); err != nil {
		return nil, fmt.Errorf("failed to save encounter: %w", err)
	}

	// Create default objectives based on encounter type
	s.createDefaultObjectives(encounter)

	return encounter, nil
}

// GetEncounter retrieves an encounter by ID
func (s *EncounterService) GetEncounter(_ context.Context, encounterID string) (*models.Encounter, error) {
	encounter, err := s.repo.GetByID(encounterID)
	if err != nil {
		return nil, fmt.Errorf("encounter not found: %w", err)
	}

	// Load objectives
	objectives, _ := s.repo.GetObjectives(encounterID)
	_ = objectives // placeholder for future objective handling

	return encounter, nil
}

// GetEncountersBySession retrieves all encounters for a game session
func (s *EncounterService) GetEncountersBySession(_ context.Context, gameSessionID string) ([]*models.Encounter, error) {
	return s.repo.GetByGameSession(gameSessionID)
}

// StartEncounter begins an encounter
func (s *EncounterService) StartEncounter(_ context.Context, encounterID string) error {
	encounter, err := s.repo.GetByID(encounterID)
	if err != nil {
		return fmt.Errorf("encounter not found: %w", err)
	}

	if encounter.Status != constants.EncounterStatusPlanned {
		return fmt.Errorf("encounter already started or completed")
	}

	// Update status
	if err := s.repo.StartEncounter(encounterID); err != nil {
		return fmt.Errorf("failed to start encounter: %w", err)
	}

	// Log event
	event := &models.EncounterEvent{
		EncounterID: encounterID,
		RoundNumber: 0,
		EventType:   "encounter_start",
		ActorType:   "system",
		ActorName:   "System",
		Description: fmt.Sprintf("Encounter '%s' has begun!", encounter.Name),
	}
	_ = s.repo.CreateEvent(event)

	return nil
}

// CompleteEncounter marks an encounter as completed
func (s *EncounterService) CompleteEncounter(_ context.Context, encounterID string, outcome string) error {
	if err := s.repo.CompleteEncounter(encounterID, outcome); err != nil {
		return fmt.Errorf("failed to complete encounter: %w", err)
	}

	// Log event
	event := &models.EncounterEvent{
		EncounterID: encounterID,
		RoundNumber: 0,
		EventType:   "encounter_complete",
		ActorType:   "system",
		ActorName:   "System",
		Description: fmt.Sprintf("Encounter completed: %s", outcome),
	}
	_ = s.repo.CreateEvent(event)

	return nil
}

// ScaleEncounter adjusts the difficulty of an encounter
func (s *EncounterService) ScaleEncounter(ctx context.Context, encounterID string, newDifficulty string) (*models.Encounter, error) {
	encounter, err := s.repo.GetByID(encounterID)
	if err != nil {
		return nil, fmt.Errorf("encounter not found: %w", err)
	}

	if encounter.ScalingOptions == nil {
		return encounter, fmt.Errorf("encounter has no scaling options")
	}

	// Apply scaling based on new difficulty
	var adjustment models.ScalingAdjustment
	switch newDifficulty {
	case difficultyEasy:
		adjustment = encounter.ScalingOptions.Easy
	case difficultyMedium:
		adjustment = encounter.ScalingOptions.Medium
	case difficultyHard:
		adjustment = encounter.ScalingOptions.Hard
	case difficultyDeadly:
		adjustment = encounter.ScalingOptions.Deadly
	default:
		return encounter, fmt.Errorf("invalid difficulty: %s", newDifficulty)
	}

	// Apply adjustments
	s.applyScalingAdjustment(encounter, &adjustment)

	// Recalculate XP
	s.encounterBuilder.calculateEncounterXP(encounter)

	// Update in database
	encounter.Difficulty = newDifficulty
	// TODO: Update encounter in database

	// Log event
	event := &models.EncounterEvent{
		EncounterID: encounterID,
		RoundNumber: 0,
		EventType:   "difficulty_scaled",
		ActorType:   "system",
		ActorName:   "DM",
		Description: fmt.Sprintf("Encounter difficulty changed to %s", newDifficulty),
	}
	_ = s.repo.CreateEvent(event)

	return encounter, nil
}

// GetTacticalSuggestion provides AI-generated tactical advice
func (s *EncounterService) GetTacticalSuggestion(ctx context.Context, encounterID, situation string) (string, error) {
	encounter, err := s.repo.GetByID(encounterID)
	if err != nil {
		return "", fmt.Errorf("encounter not found: %w", err)
	}

	suggestion, err := s.encounterBuilder.GenerateTacticalSuggestion(ctx, encounter, situation)
	if err != nil {
		return "", fmt.Errorf("failed to generate suggestion: %w", err)
	}

	// Log the suggestion
	event := &models.EncounterEvent{
		EncounterID:  encounterID,
		RoundNumber:  0, // Should be passed in
		EventType:    "tactical_suggestion",
		ActorType:    "system",
		ActorName:    "AI Tactician",
		Description:  situation,
		AISuggestion: suggestion,
	}
	_ = s.repo.CreateEvent(event)

	return suggestion, nil
}

// LogCombatEvent records an event during combat
func (s *EncounterService) LogCombatEvent(ctx context.Context, encounterID string, round int, eventType, actorType, actorID, actorName, description string, mechanicalEffect map[string]interface{}) error {
	event := &models.EncounterEvent{
		EncounterID:      encounterID,
		RoundNumber:      round,
		EventType:        eventType,
		ActorType:        actorType,
		ActorID:          &actorID,
		ActorName:        actorName,
		Description:      description,
		MechanicalEffect: mechanicalEffect,
	}

	return s.repo.CreateEvent(event)
}

// GetEncounterEvents retrieves recent events for an encounter
func (s *EncounterService) GetEncounterEvents(ctx context.Context, encounterID string, limit int) ([]*models.EncounterEvent, error) {
	return s.repo.GetEvents(encounterID, limit)
}

// UpdateEnemyStatus updates an enemy's status during combat
func (s *EncounterService) UpdateEnemyStatus(ctx context.Context, enemyID string, updates map[string]interface{}) error {
	return s.repo.UpdateEnemyStatus(enemyID, updates)
}

// TriggerReinforcements activates a reinforcement wave
func (s *EncounterService) TriggerReinforcements(ctx context.Context, encounterID string, waveIndex int) error {
	encounter, err := s.repo.GetByID(encounterID)
	if err != nil {
		return fmt.Errorf("encounter not found: %w", err)
	}

	if waveIndex >= len(encounter.ReinforcementWaves) {
		return fmt.Errorf("invalid reinforcement wave index")
	}

	wave := encounter.ReinforcementWaves[waveIndex]

	// Add reinforcement enemies to the encounter
	for i := range wave.Enemies {
		wave.Enemies[i].EncounterID = encounterID
		if err := s.repo.CreateEncounterEnemy(&wave.Enemies[i]); err != nil {
			return fmt.Errorf("failed to add reinforcement: %w", err)
		}
	}

	// Log event
	event := &models.EncounterEvent{
		EncounterID: encounterID,
		RoundNumber: wave.Round,
		EventType:   "reinforcements_arrived",
		ActorType:   "environment",
		ActorName:   "Battlefield",
		Description: wave.Announcement,
	}
	_ = s.repo.CreateEvent(event)

	return nil
}

// CheckObjectives evaluates and updates objective completion
func (s *EncounterService) CheckObjectives(ctx context.Context, encounterID string) error {
	objectives, err := s.repo.GetObjectives(encounterID)
	if err != nil {
		return fmt.Errorf("failed to get objectives: %w", err)
	}

	encounter, err := s.repo.GetByID(encounterID)
	if err != nil {
		return fmt.Errorf("encounter not found: %w", err)
	}

	for _, objective := range objectives {
		if objective.IsCompleted || objective.IsFailed {
			continue
		}

		// Check success conditions based on type
		completed := false
		failed := false

		switch objective.Type {
		case constants.ObjectiveDefeatAll:
			// Check if all enemies are defeated
			allDefeated := true
			for i := range encounter.Enemies {
				if encounter.Enemies[i].IsAlive && !encounter.Enemies[i].Fled {
					allDefeated = false
					break
				}
			}
			completed = allDefeated

		case "survive_rounds":
			// This would need round tracking
			// completed = currentRound >= targetRounds

		case "protect_npc":
			// Check if protected NPC is still alive
			// This would need NPC tracking

			// Add more objective types as needed
		}

		if completed {
			_ = s.repo.CompleteObjective(objective.ID)
			s.awardObjectiveRewards(ctx, objective, encounter.GameSessionID)
		} else if failed {
			_ = s.repo.FailObjective(objective.ID)
		}
	}

	return nil
}

// Helper functions

func (s *EncounterService) createDefaultObjectives(encounter *models.Encounter) {
	// Create primary objective based on encounter type
	primaryObjective := &models.EncounterObjective{
		EncounterID: encounter.ID,
		Type:        "custom",
		Description: fmt.Sprintf("Complete the %s encounter", encounter.EncounterType),
		XPReward:    encounter.TotalXP,
		GoldReward:  encounter.TotalXP / 10,
	}

	switch encounter.EncounterType {
	case constants.EncounterTypeCombat:
		primaryObjective.Type = constants.ObjectiveDefeatAll
		primaryObjective.Description = "Defeat all enemies"
	case "social":
		primaryObjective.Type = "negotiate"
		primaryObjective.Description = "Successfully negotiate or persuade"
	case "exploration":
		primaryObjective.Type = "reach_location"
		primaryObjective.Description = "Explore and discover the area"
	case "puzzle":
		primaryObjective.Type = "solve_puzzle"
		primaryObjective.Description = "Solve the puzzle or riddle"
	}

	_ = s.repo.CreateObjective(primaryObjective)

	// Create optional objectives for non-combat resolution
	if encounter.EncounterType == constants.EncounterTypeCombat && len(encounter.SocialSolutions) > 0 {
		bonusObjective := &models.EncounterObjective{
			EncounterID:  encounter.ID,
			Type:         "custom",
			Description:  "Resolve the encounter without violence",
			XPReward:     encounter.TotalXP / 4,
			StoryRewards: []string{"Peaceful resolution bonus"},
		}
		_ = s.repo.CreateObjective(bonusObjective)
	}
}

func (s *EncounterService) applyScalingAdjustment(encounter *models.Encounter, adjustment *models.ScalingAdjustment) {
	// Remove enemies - logic to remove specified enemies could be added here

	// Add enemies - placeholder for future logic

	// Adjust HP for all enemies
	if adjustment.HPModifier != 0 {
		for i := range encounter.Enemies {
			encounter.Enemies[i].HitPoints += adjustment.HPModifier
			if encounter.Enemies[i].HitPoints < 1 {
				encounter.Enemies[i].HitPoints = 1
			}
		}
	}

	// Adjust damage (would need to modify actions)
	_ = adjustment.DamageModifier // Placeholder for future damage modification

	// Add hazards - future enhancement

	// Add terrain - future enhancement
}

func (s *EncounterService) awardObjectiveRewards(ctx context.Context, objective *models.EncounterObjective, gameSessionID string) {
	// Award XP to party members
	// This would integrate with character service to add XP

	// Award gold
	// This would integrate with inventory service

	// Award items
	_ = objective.ItemRewards // Placeholder for item rewards integration

	// Log the rewards
	event := &models.EncounterEvent{
		EncounterID: objective.EncounterID,
		RoundNumber: 0,
		EventType:   "rewards_granted",
		ActorType:   "system",
		ActorName:   "System",
		Description: fmt.Sprintf("Objective completed: %s. Rewards: %d XP, %d gold",
			objective.Description, objective.XPReward, objective.GoldReward),
	}
	_ = s.repo.CreateEvent(event)
}

package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// CustomRaceService handles custom race operations
type CustomRaceService struct {
	repo        database.CustomRaceRepository
	aiGenerator AIRaceGeneratorInterface
}

// NewCustomRaceService creates a new custom race service
func NewCustomRaceService(repo database.CustomRaceRepository, aiGenerator AIRaceGeneratorInterface) *CustomRaceService {
	return &CustomRaceService{
		repo:        repo,
		aiGenerator: aiGenerator,
	}
}

// CreateCustomRace generates and stores a new custom race
func (s *CustomRaceService) CreateCustomRace(ctx context.Context, userID uuid.UUID, request models.CustomRaceRequest) (*models.CustomRace, error) {
	// Generate the race using AI
	generatedRace, err := s.aiGenerator.GenerateCustomRace(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate custom race: %w", err)
	}

	// Create the custom race model
	customRace := &models.CustomRace{
		ID:                    uuid.New(),
		Name:                  generatedRace.Name,
		Description:           generatedRace.Description,
		UserPrompt:            fmt.Sprintf("Name: %s\nDescription: %s", request.Name, request.Description),
		AbilityScoreIncreases: generatedRace.AbilityScoreIncreases,
		Size:                  generatedRace.Size,
		Speed:                 generatedRace.Speed,
		Traits:                generatedRace.Traits,
		Languages:             generatedRace.Languages,
		Darkvision:            generatedRace.Darkvision,
		Resistances:           generatedRace.Resistances,
		Immunities:            generatedRace.Immunities,
		SkillProficiencies:    generatedRace.SkillProficiencies,
		ToolProficiencies:     generatedRace.ToolProficiencies,
		WeaponProficiencies:   generatedRace.WeaponProficiencies,
		ArmorProficiencies:    generatedRace.ArmorProficiencies,
		CreatedBy:             userID,
		ApprovalStatus:        models.ApprovalStatusPending,
		BalanceScore:          &generatedRace.BalanceScore,
		TimesUsed:             0,
		IsPublic:              false,
	}

	// Check if the balance score indicates it needs automatic approval
	if generatedRace.BalanceScore <= 7 {
		// Auto-approve well-balanced races
		customRace.ApprovalStatus = models.ApprovalStatusApproved
		approvalNote := fmt.Sprintf("Auto-approved: Balance score %d. %s", generatedRace.BalanceScore, generatedRace.BalanceExplanation)
		customRace.ApprovalNotes = &approvalNote
	}

	// Save to database
	if err := s.repo.Create(ctx, customRace); err != nil {
		return nil, fmt.Errorf("failed to save custom race: %w", err)
	}

	return customRace, nil
}

// GetCustomRace retrieves a custom race by ID
func (s *CustomRaceService) GetCustomRace(ctx context.Context, raceID uuid.UUID) (*models.CustomRace, error) {
	race, err := s.repo.GetByID(ctx, raceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("custom race not found")
		}
		return nil, fmt.Errorf("failed to get custom race: %w", err)
	}

	return race, nil
}

// GetUserCustomRaces retrieves all custom races created by a user
func (s *CustomRaceService) GetUserCustomRaces(ctx context.Context, userID uuid.UUID) ([]*models.CustomRace, error) {
	races, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user custom races: %w", err)
	}

	return races, nil
}

// GetPublicCustomRaces retrieves all approved public custom races
func (s *CustomRaceService) GetPublicCustomRaces(ctx context.Context) ([]*models.CustomRace, error) {
	races, err := s.repo.GetPublicRaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get public custom races: %w", err)
	}

	return races, nil
}

// ApproveCustomRace approves a custom race (DM only)
func (s *CustomRaceService) ApproveCustomRace(ctx context.Context, raceID uuid.UUID, approverID uuid.UUID, notes string) error {
	race, err := s.repo.GetByID(ctx, raceID)
	if err != nil {
		return fmt.Errorf("failed to get custom race: %w", err)
	}

	race.ApprovalStatus = models.ApprovalStatusApproved
	race.ApprovedBy = &approverID
	race.ApprovalNotes = &notes

	if err := s.repo.Update(ctx, race); err != nil {
		return fmt.Errorf("failed to update custom race: %w", err)
	}

	return nil
}

// RejectCustomRace rejects a custom race (DM only)
func (s *CustomRaceService) RejectCustomRace(ctx context.Context, raceID uuid.UUID, approverID uuid.UUID, notes string) error {
	race, err := s.repo.GetByID(ctx, raceID)
	if err != nil {
		return fmt.Errorf("failed to get custom race: %w", err)
	}

	race.ApprovalStatus = models.ApprovalStatusRejected
	race.ApprovedBy = &approverID
	race.ApprovalNotes = &notes

	if err := s.repo.Update(ctx, race); err != nil {
		return fmt.Errorf("failed to update custom race: %w", err)
	}

	return nil
}

// RequestRevision requests changes to a custom race (DM only)
func (s *CustomRaceService) RequestRevision(ctx context.Context, raceID uuid.UUID, approverID uuid.UUID, notes string) error {
	race, err := s.repo.GetByID(ctx, raceID)
	if err != nil {
		return fmt.Errorf("failed to get custom race: %w", err)
	}

	race.ApprovalStatus = models.ApprovalStatusRevisionNeeded
	race.ApprovedBy = &approverID
	race.ApprovalNotes = &notes

	if err := s.repo.Update(ctx, race); err != nil {
		return fmt.Errorf("failed to update custom race: %w", err)
	}

	return nil
}

// MakePublic makes a custom race available to all players
func (s *CustomRaceService) MakePublic(ctx context.Context, raceID uuid.UUID, userID uuid.UUID) error {
	race, err := s.repo.GetByID(ctx, raceID)
	if err != nil {
		return fmt.Errorf("failed to get custom race: %w", err)
	}

	// Only the creator or an approver can make it public
	if race.CreatedBy != userID && (race.ApprovedBy == nil || *race.ApprovedBy != userID) {
		return fmt.Errorf("unauthorized to make this race public")
	}

	// Only approved races can be made public
	if race.ApprovalStatus != models.ApprovalStatusApproved {
		return fmt.Errorf("only approved races can be made public")
	}

	race.IsPublic = true

	if err := s.repo.Update(ctx, race); err != nil {
		return fmt.Errorf("failed to update custom race: %w", err)
	}

	return nil
}

// IncrementUsage increments the usage counter for a custom race
func (s *CustomRaceService) IncrementUsage(ctx context.Context, raceID uuid.UUID) error {
	return s.repo.IncrementUsage(ctx, raceID)
}

// ValidateCustomRaceForCharacter validates if a custom race can be used for a character
func (s *CustomRaceService) ValidateCustomRaceForCharacter(ctx context.Context, raceID uuid.UUID, userID uuid.UUID) error {
	race, err := s.repo.GetByID(ctx, raceID)
	if err != nil {
		return fmt.Errorf("custom race not found")
	}

	// Check if the user can use this race
	canUse := false

	// Creator can always use their own race
	if race.CreatedBy == userID {
		canUse = true
	}

	// Anyone can use public, approved races
	if race.IsPublic && race.ApprovalStatus == models.ApprovalStatusApproved {
		canUse = true
	}

	// Check if specifically approved for this user (future enhancement)
	// This could be extended to allow DMs to approve races for specific players

	if !canUse {
		return fmt.Errorf("you don't have permission to use this custom race")
	}

	// Check approval status
	if race.ApprovalStatus != models.ApprovalStatusApproved {
		return fmt.Errorf("this custom race has not been approved yet")
	}

	return nil
}

// GetCustomRaceStats formats custom race data for character creation
func (s *CustomRaceService) GetCustomRaceStats(ctx context.Context, raceID uuid.UUID) (map[string]interface{}, error) {
	race, err := s.repo.GetByID(ctx, raceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom race: %w", err)
	}

	// Format the race data similar to standard races
	stats := map[string]interface{}{
		"name":                  race.Name,
		"description":           race.Description,
		"abilityScoreIncreases": race.AbilityScoreIncreases,
		"size":                  race.Size,
		"speed":                 race.Speed,
		"traits":                race.Traits,
		"languages":             race.Languages,
		"darkvision":            race.Darkvision,
		"resistances":           race.Resistances,
		"immunities":            race.Immunities,
		"skillProficiencies":    race.SkillProficiencies,
		"toolProficiencies":     race.ToolProficiencies,
		"weaponProficiencies":   race.WeaponProficiencies,
		"armorProficiencies":    race.ArmorProficiencies,
		"isCustom":              true,
		"customRaceId":          race.ID,
	}

	return stats, nil
}

// GetPendingApproval retrieves all custom races pending DM approval
func (s *CustomRaceService) GetPendingApproval(ctx context.Context) ([]*models.CustomRace, error) {
	races, err := s.repo.GetPendingApproval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending custom races: %w", err)
	}

	return races, nil
}

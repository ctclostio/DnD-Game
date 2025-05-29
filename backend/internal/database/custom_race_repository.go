package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// CustomRaceRepository defines the interface for custom race database operations
type CustomRaceRepository interface {
	Create(ctx context.Context, race *models.CustomRace) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.CustomRace, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.CustomRace, error)
	GetPublicRaces(ctx context.Context) ([]*models.CustomRace, error)
	GetPendingApproval(ctx context.Context) ([]*models.CustomRace, error)
	Update(ctx context.Context, race *models.CustomRace) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementUsage(ctx context.Context, id uuid.UUID) error
}

// customRaceRepository implements CustomRaceRepository
type customRaceRepository struct {
	db *sqlx.DB
}

// NewCustomRaceRepository creates a new custom race repository
func NewCustomRaceRepository(db *sqlx.DB) CustomRaceRepository {
	return &customRaceRepository{db: db}
}

// Create inserts a new custom race
func (r *customRaceRepository) Create(ctx context.Context, race *models.CustomRace) error {
	// Marshal JSON fields
	abilityScoresJSON, err := json.Marshal(race.AbilityScoreIncreases)
	if err != nil {
		return fmt.Errorf("failed to marshal ability scores: %w", err)
	}

	traitsJSON, err := json.Marshal(race.Traits)
	if err != nil {
		return fmt.Errorf("failed to marshal traits: %w", err)
	}

	languagesJSON, err := json.Marshal(race.Languages)
	if err != nil {
		return fmt.Errorf("failed to marshal languages: %w", err)
	}

	resistancesJSON, err := json.Marshal(race.Resistances)
	if err != nil {
		return fmt.Errorf("failed to marshal resistances: %w", err)
	}

	immunitiesJSON, err := json.Marshal(race.Immunities)
	if err != nil {
		return fmt.Errorf("failed to marshal immunities: %w", err)
	}

	skillProfJSON, err := json.Marshal(race.SkillProficiencies)
	if err != nil {
		return fmt.Errorf("failed to marshal skill proficiencies: %w", err)
	}

	toolProfJSON, err := json.Marshal(race.ToolProficiencies)
	if err != nil {
		return fmt.Errorf("failed to marshal tool proficiencies: %w", err)
	}

	weaponProfJSON, err := json.Marshal(race.WeaponProficiencies)
	if err != nil {
		return fmt.Errorf("failed to marshal weapon proficiencies: %w", err)
	}

	armorProfJSON, err := json.Marshal(race.ArmorProficiencies)
	if err != nil {
		return fmt.Errorf("failed to marshal armor proficiencies: %w", err)
	}

	query := `
		INSERT INTO custom_races (
			id, name, description, user_prompt,
			ability_score_increases, size, speed,
			traits, languages, darkvision,
			resistances, immunities,
			skill_proficiencies, tool_proficiencies,
			weapon_proficiencies, armor_proficiencies,
			created_by, approval_status, balance_score,
			approval_notes, is_public
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21
		)`

	_, err = r.db.ExecContext(ctx, query,
		race.ID, race.Name, race.Description, race.UserPrompt,
		abilityScoresJSON, race.Size, race.Speed,
		traitsJSON, languagesJSON, race.Darkvision,
		resistancesJSON, immunitiesJSON,
		skillProfJSON, toolProfJSON,
		weaponProfJSON, armorProfJSON,
		race.CreatedBy, race.ApprovalStatus, race.BalanceScore,
		race.ApprovalNotes, race.IsPublic,
	)

	return err
}

// GetByID retrieves a custom race by ID
func (r *customRaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CustomRace, error) {
	var race models.CustomRace
	var (
		abilityScoresJSON   []byte
		traitsJSON          []byte
		languagesJSON       []byte
		resistancesJSON     []byte
		immunitiesJSON      []byte
		skillProfJSON       []byte
		toolProfJSON        []byte
		weaponProfJSON      []byte
		armorProfJSON       []byte
	)

	query := `
		SELECT 
			id, name, description, user_prompt,
			ability_score_increases, size, speed,
			traits, languages, darkvision,
			resistances, immunities,
			skill_proficiencies, tool_proficiencies,
			weapon_proficiencies, armor_proficiencies,
			created_by, approved_by, approval_status,
			approval_notes, balance_score,
			times_used, is_public,
			created_at, updated_at
		FROM custom_races
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&race.ID, &race.Name, &race.Description, &race.UserPrompt,
		&abilityScoresJSON, &race.Size, &race.Speed,
		&traitsJSON, &languagesJSON, &race.Darkvision,
		&resistancesJSON, &immunitiesJSON,
		&skillProfJSON, &toolProfJSON,
		&weaponProfJSON, &armorProfJSON,
		&race.CreatedBy, &race.ApprovedBy, &race.ApprovalStatus,
		&race.ApprovalNotes, &race.BalanceScore,
		&race.TimesUsed, &race.IsPublic,
		&race.CreatedAt, &race.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if err := unmarshalRaceJSON(&race, abilityScoresJSON, traitsJSON, languagesJSON,
		resistancesJSON, immunitiesJSON, skillProfJSON, toolProfJSON,
		weaponProfJSON, armorProfJSON); err != nil {
		return nil, err
	}

	return &race, nil
}

// GetByUserID retrieves all custom races created by a user
func (r *customRaceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.CustomRace, error) {
	query := `
		SELECT 
			id, name, description, user_prompt,
			ability_score_increases, size, speed,
			traits, languages, darkvision,
			resistances, immunities,
			skill_proficiencies, tool_proficiencies,
			weapon_proficiencies, armor_proficiencies,
			created_by, approved_by, approval_status,
			approval_notes, balance_score,
			times_used, is_public,
			created_at, updated_at
		FROM custom_races
		WHERE created_by = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCustomRaces(rows)
}

// GetPublicRaces retrieves all approved public custom races
func (r *customRaceRepository) GetPublicRaces(ctx context.Context) ([]*models.CustomRace, error) {
	query := `
		SELECT 
			id, name, description, user_prompt,
			ability_score_increases, size, speed,
			traits, languages, darkvision,
			resistances, immunities,
			skill_proficiencies, tool_proficiencies,
			weapon_proficiencies, armor_proficiencies,
			created_by, approved_by, approval_status,
			approval_notes, balance_score,
			times_used, is_public,
			created_at, updated_at
		FROM custom_races
		WHERE is_public = true AND approval_status = 'approved'
		ORDER BY times_used DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCustomRaces(rows)
}

// GetPendingApproval retrieves all custom races pending DM approval
func (r *customRaceRepository) GetPendingApproval(ctx context.Context) ([]*models.CustomRace, error) {
	query := `
		SELECT 
			id, name, description, user_prompt,
			ability_score_increases, size, speed,
			traits, languages, darkvision,
			resistances, immunities,
			skill_proficiencies, tool_proficiencies,
			weapon_proficiencies, armor_proficiencies,
			created_by, approved_by, approval_status,
			approval_notes, balance_score,
			times_used, is_public,
			created_at, updated_at
		FROM custom_races
		WHERE approval_status = 'pending'
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCustomRaces(rows)
}

// Update updates a custom race
func (r *customRaceRepository) Update(ctx context.Context, race *models.CustomRace) error {
	// Marshal JSON fields
	abilityScoresJSON, err := json.Marshal(race.AbilityScoreIncreases)
	if err != nil {
		return fmt.Errorf("failed to marshal ability scores: %w", err)
	}

	traitsJSON, err := json.Marshal(race.Traits)
	if err != nil {
		return fmt.Errorf("failed to marshal traits: %w", err)
	}

	languagesJSON, err := json.Marshal(race.Languages)
	if err != nil {
		return fmt.Errorf("failed to marshal languages: %w", err)
	}

	resistancesJSON, err := json.Marshal(race.Resistances)
	if err != nil {
		return fmt.Errorf("failed to marshal resistances: %w", err)
	}

	immunitiesJSON, err := json.Marshal(race.Immunities)
	if err != nil {
		return fmt.Errorf("failed to marshal immunities: %w", err)
	}

	skillProfJSON, err := json.Marshal(race.SkillProficiencies)
	if err != nil {
		return fmt.Errorf("failed to marshal skill proficiencies: %w", err)
	}

	toolProfJSON, err := json.Marshal(race.ToolProficiencies)
	if err != nil {
		return fmt.Errorf("failed to marshal tool proficiencies: %w", err)
	}

	weaponProfJSON, err := json.Marshal(race.WeaponProficiencies)
	if err != nil {
		return fmt.Errorf("failed to marshal weapon proficiencies: %w", err)
	}

	armorProfJSON, err := json.Marshal(race.ArmorProficiencies)
	if err != nil {
		return fmt.Errorf("failed to marshal armor proficiencies: %w", err)
	}

	query := `
		UPDATE custom_races SET
			name = $2, description = $3, user_prompt = $4,
			ability_score_increases = $5, size = $6, speed = $7,
			traits = $8, languages = $9, darkvision = $10,
			resistances = $11, immunities = $12,
			skill_proficiencies = $13, tool_proficiencies = $14,
			weapon_proficiencies = $15, armor_proficiencies = $16,
			approved_by = $17, approval_status = $18,
			approval_notes = $19, balance_score = $20,
			is_public = $21, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query,
		race.ID, race.Name, race.Description, race.UserPrompt,
		abilityScoresJSON, race.Size, race.Speed,
		traitsJSON, languagesJSON, race.Darkvision,
		resistancesJSON, immunitiesJSON,
		skillProfJSON, toolProfJSON,
		weaponProfJSON, armorProfJSON,
		race.ApprovedBy, race.ApprovalStatus,
		race.ApprovalNotes, race.BalanceScore,
		race.IsPublic,
	)

	return err
}

// Delete removes a custom race
func (r *customRaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM custom_races WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// IncrementUsage increments the usage counter for a custom race
func (r *customRaceRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE custom_races SET times_used = times_used + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Helper functions

func scanCustomRaces(rows *sql.Rows) ([]*models.CustomRace, error) {
	var races []*models.CustomRace

	for rows.Next() {
		var race models.CustomRace
		var (
			abilityScoresJSON   []byte
			traitsJSON          []byte
			languagesJSON       []byte
			resistancesJSON     []byte
			immunitiesJSON      []byte
			skillProfJSON       []byte
			toolProfJSON        []byte
			weaponProfJSON      []byte
			armorProfJSON       []byte
		)

		err := rows.Scan(
			&race.ID, &race.Name, &race.Description, &race.UserPrompt,
			&abilityScoresJSON, &race.Size, &race.Speed,
			&traitsJSON, &languagesJSON, &race.Darkvision,
			&resistancesJSON, &immunitiesJSON,
			&skillProfJSON, &toolProfJSON,
			&weaponProfJSON, &armorProfJSON,
			&race.CreatedBy, &race.ApprovedBy, &race.ApprovalStatus,
			&race.ApprovalNotes, &race.BalanceScore,
			&race.TimesUsed, &race.IsPublic,
			&race.CreatedAt, &race.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := unmarshalRaceJSON(&race, abilityScoresJSON, traitsJSON, languagesJSON,
			resistancesJSON, immunitiesJSON, skillProfJSON, toolProfJSON,
			weaponProfJSON, armorProfJSON); err != nil {
			return nil, err
		}

		races = append(races, &race)
	}

	return races, rows.Err()
}

func unmarshalRaceJSON(race *models.CustomRace, abilityScores, traits, languages,
	resistances, immunities, skillProf, toolProf, weaponProf, armorProf []byte) error {

	if err := json.Unmarshal(abilityScores, &race.AbilityScoreIncreases); err != nil {
		return fmt.Errorf("failed to unmarshal ability scores: %w", err)
	}

	if err := json.Unmarshal(traits, &race.Traits); err != nil {
		return fmt.Errorf("failed to unmarshal traits: %w", err)
	}

	if err := json.Unmarshal(languages, &race.Languages); err != nil {
		return fmt.Errorf("failed to unmarshal languages: %w", err)
	}

	if err := json.Unmarshal(resistances, &race.Resistances); err != nil {
		return fmt.Errorf("failed to unmarshal resistances: %w", err)
	}

	if err := json.Unmarshal(immunities, &race.Immunities); err != nil {
		return fmt.Errorf("failed to unmarshal immunities: %w", err)
	}

	if err := json.Unmarshal(skillProf, &race.SkillProficiencies); err != nil {
		return fmt.Errorf("failed to unmarshal skill proficiencies: %w", err)
	}

	if err := json.Unmarshal(toolProf, &race.ToolProficiencies); err != nil {
		return fmt.Errorf("failed to unmarshal tool proficiencies: %w", err)
	}

	if err := json.Unmarshal(weaponProf, &race.WeaponProficiencies); err != nil {
		return fmt.Errorf("failed to unmarshal weapon proficiencies: %w", err)
	}

	if err := json.Unmarshal(armorProf, &race.ArmorProficiencies); err != nil {
		return fmt.Errorf("failed to unmarshal armor proficiencies: %w", err)
	}

	return nil
}
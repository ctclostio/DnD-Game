package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// characterRepository implements CharacterRepository interface
type characterRepository struct {
	db *DB
}

// NewCharacterRepository creates a new character repository
func NewCharacterRepository(db *DB) CharacterRepository {
	return &characterRepository{db: db}
}

// scanCharacter is a helper to scan a single Character row with JSON fields
func (r *characterRepository) scanCharacter(row RowScanner) (*models.Character, error) {
	var character models.Character
	var attributesJSON, skillsJSON, equipmentJSON, spellsJSON []byte

	err := row.Scan(
		&character.ID, &character.UserID, &character.Name, &character.Race,
		&character.Class, &character.Level, &character.ExperiencePoints,
		&character.HitPoints, &character.MaxHitPoints, &character.ArmorClass,
		&character.Speed, &attributesJSON, &skillsJSON, &equipmentJSON,
		&spellsJSON, &character.CreatedAt, &character.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan character: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(attributesJSON, &character.Attributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}
	if err := json.Unmarshal(skillsJSON, &character.Skills); err != nil {
		return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
	}
	if err := json.Unmarshal(equipmentJSON, &character.Equipment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
	}
	if err := json.Unmarshal(spellsJSON, &character.Spells); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spells: %w", err)
	}

	return &character, nil
}

// Create creates a new character
func (r *characterRepository) Create(ctx context.Context, character *models.Character) error {
	// Generate ID if not provided (for SQLite compatibility)
	if character.ID == "" {
		character.ID = uuid.New().String()
	}

	// Convert complex types to JSON
	attributesJSON, err := json.Marshal(character.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	skillsJSON, err := json.Marshal(character.Skills)
	if err != nil {
		return fmt.Errorf("failed to marshal skills: %w", err)
	}

	equipmentJSON, err := json.Marshal(character.Equipment)
	if err != nil {
		return fmt.Errorf("failed to marshal equipment: %w", err)
	}

	spellsJSON, err := json.Marshal(character.Spells)
	if err != nil {
		return fmt.Errorf("failed to marshal spells: %w", err)
	}

	query := `
		INSERT INTO characters (
			id, user_id, name, race, class, level, experience_points,
			hit_points, max_hit_points, armor_class, speed,
			attributes, skills, equipment, spells
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	err = r.db.QueryRowContextRebind(ctx, query,
		character.ID, character.UserID, character.Name, character.Race, character.Class,
		character.Level, character.ExperiencePoints, character.HitPoints,
		character.MaxHitPoints, character.ArmorClass, character.Speed,
		attributesJSON, skillsJSON, equipmentJSON, spellsJSON).
		Scan(&character.ID, &character.CreatedAt, &character.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// GetByID retrieves a character by ID
func (r *characterRepository) GetByID(ctx context.Context, id string) (*models.Character, error) {
	var character models.Character
	var attributesJSON, skillsJSON, equipmentJSON, spellsJSON []byte

	query := `
		SELECT id, user_id, name, race, COALESCE(subrace, ''), class, COALESCE(subclass, ''),
			   COALESCE(background, ''), COALESCE(alignment, ''), level, experience_points,
			   hit_points, max_hit_points, armor_class, speed,
			   attributes, skills, equipment, spells, created_at, updated_at
		FROM characters
		WHERE id = ?`

	err := r.db.QueryRowContextRebind(ctx, query, id).Scan(
		&character.ID, &character.UserID, &character.Name, &character.Race, &character.Subrace,
		&character.Class, &character.Subclass, &character.Background, &character.Alignment,
		&character.Level, &character.ExperiencePoints,
		&character.HitPoints, &character.MaxHitPoints, &character.ArmorClass,
		&character.Speed, &attributesJSON, &skillsJSON, &equipmentJSON,
		&spellsJSON, &character.CreatedAt, &character.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf(constants.ErrCharacterNotFound)
		}
		return nil, fmt.Errorf("failed to scan character data: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(attributesJSON, &character.Attributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}
	if err := json.Unmarshal(skillsJSON, &character.Skills); err != nil {
		return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
	}
	if err := json.Unmarshal(equipmentJSON, &character.Equipment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal equipment: %w", err)
	}
	if err := json.Unmarshal(spellsJSON, &character.Spells); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spells: %w", err)
	}

	// Initialize fields that might not be in the database
	// Note: We don't need to check if SavingThrows is empty since it only contains basic types
	// Initialize empty slices and maps
	if character.Proficiencies.Languages == nil {
		character.Proficiencies.Languages = []string{}
	}
	if character.Proficiencies.Tools == nil {
		character.Proficiencies.Tools = []string{}
	}
	if character.Proficiencies.Weapons == nil {
		character.Proficiencies.Weapons = []string{}
	}
	if character.Proficiencies.Armor == nil {
		character.Proficiencies.Armor = []string{}
	}
	if character.Features == nil {
		character.Features = []models.Feature{}
	}
	if character.Resources == nil {
		character.Resources = make(map[string]interface{})
	}

	return &character, nil
}

// GetByUserID retrieves all characters for a user
func (r *characterRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Character, error) {
	query := `
		SELECT id, user_id, name, race, class, level, experience_points,
			   hit_points, max_hit_points, armor_class, speed,
			   attributes, skills, equipment, spells, created_at, updated_at
		FROM characters
		WHERE user_id = ?
		ORDER BY created_at DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get characters by user id: %w", err)
	}
	defer func() { _ = rows.Close() }()

	characters := make([]*models.Character, 0, 10)
	charactersPtr, err := ScanRowsGeneric(rows, r.scanCharacter)
	if err != nil {
		return nil, err
	}
	for _, c := range charactersPtr {
		characters = append(characters, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating characters: %w", err)
	}

	return characters, nil
}

// Update updates a character
func (r *characterRepository) Update(ctx context.Context, character *models.Character) error {
	// Convert complex types to JSON
	attributesJSON, err := json.Marshal(character.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	skillsJSON, err := json.Marshal(character.Skills)
	if err != nil {
		return fmt.Errorf("failed to marshal skills: %w", err)
	}

	equipmentJSON, err := json.Marshal(character.Equipment)
	if err != nil {
		return fmt.Errorf("failed to marshal equipment: %w", err)
	}

	spellsJSON, err := json.Marshal(character.Spells)
	if err != nil {
		return fmt.Errorf("failed to marshal spells: %w", err)
	}

	// Use ? placeholders and rebind for database compatibility
	query := `
		UPDATE characters
		SET name = ?, race = ?, class = ?, level = ?, experience_points = ?,
			hit_points = ?, max_hit_points = ?, armor_class = ?, speed = ?,
			attributes = ?, skills = ?, equipment = ?, spells = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := r.db.ExecContextRebind(ctx, query,
		character.Name, character.Race, character.Class,
		character.Level, character.ExperiencePoints, character.HitPoints,
		character.MaxHitPoints, character.ArmorClass, character.Speed,
		attributesJSON, skillsJSON, equipmentJSON, spellsJSON,
		character.ID)
	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(constants.ErrCharacterNotFound)
	}

	return nil
}

// Delete deletes a character
func (r *characterRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM characters WHERE id = ?`

	result, err := r.db.ExecContextRebind(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(constants.ErrCharacterNotFound)
	}

	return nil
}

// List retrieves a paginated list of characters
func (r *characterRepository) List(ctx context.Context, offset, limit int) ([]*models.Character, error) {
	query := `
		SELECT id, user_id, name, race, class, level, experience_points,
			   hit_points, max_hit_points, armor_class, speed,
			   attributes, skills, equipment, spells, created_at, updated_at
		FROM characters
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	characters := make([]*models.Character, 0, 10)
	charactersPtr, err := ScanRowsGeneric(rows, r.scanCharacter)
	if err != nil {
		return nil, err
	}
	for _, c := range charactersPtr {
		characters = append(characters, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating characters: %w", err)
	}

	return characters, nil
}

package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/your-username/dnd-game/backend/internal/models"
)

// characterRepository implements CharacterRepository interface
type characterRepository struct {
	db *DB
}

// NewCharacterRepository creates a new character repository
func NewCharacterRepository(db *DB) CharacterRepository {
	return &characterRepository{db: db}
}

// Create creates a new character
func (r *characterRepository) Create(ctx context.Context, character *models.Character) error {
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
			user_id, name, race, class, level, experience_points,
			hit_points, max_hit_points, armor_class, speed,
			attributes, skills, equipment, spells
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	err = r.db.QueryRowContext(ctx, query,
		character.UserID, character.Name, character.Race, character.Class,
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
		SELECT id, user_id, name, race, class, level, experience_points,
			   hit_points, max_hit_points, armor_class, speed,
			   attributes, skills, equipment, spells, created_at, updated_at
		FROM characters
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&character.ID, &character.UserID, &character.Name, &character.Race,
		&character.Class, &character.Level, &character.ExperiencePoints,
		&character.HitPoints, &character.MaxHitPoints, &character.ArmorClass,
		&character.Speed, &attributesJSON, &skillsJSON, &equipmentJSON,
		&spellsJSON, &character.CreatedAt, &character.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("character not found")
		}
		return nil, fmt.Errorf("failed to get character by id: %w", err)
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

// GetByUserID retrieves all characters for a user
func (r *characterRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Character, error) {
	query := `
		SELECT id, user_id, name, race, class, level, experience_points,
			   hit_points, max_hit_points, armor_class, speed,
			   attributes, skills, equipment, spells, created_at, updated_at
		FROM characters
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get characters by user id: %w", err)
	}
	defer rows.Close()

	var characters []*models.Character
	for rows.Next() {
		var character models.Character
		var attributesJSON, skillsJSON, equipmentJSON, spellsJSON []byte

		err := rows.Scan(
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

		characters = append(characters, &character)
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

	query := `
		UPDATE characters
		SET name = $2, race = $3, class = $4, level = $5, experience_points = $6,
			hit_points = $7, max_hit_points = $8, armor_class = $9, speed = $10,
			attributes = $11, skills = $12, equipment = $13, spells = $14,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`

	err = r.db.QueryRowContext(ctx, query,
		character.ID, character.Name, character.Race, character.Class,
		character.Level, character.ExperiencePoints, character.HitPoints,
		character.MaxHitPoints, character.ArmorClass, character.Speed,
		attributesJSON, skillsJSON, equipmentJSON, spellsJSON).
		Scan(&character.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("character not found")
		}
		return fmt.Errorf("failed to update character: %w", err)
	}

	return nil
}

// Delete deletes a character
func (r *characterRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM characters WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("character not found")
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
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}
	defer rows.Close()

	var characters []*models.Character
	for rows.Next() {
		var character models.Character
		var attributesJSON, skillsJSON, equipmentJSON, spellsJSON []byte

		err := rows.Scan(
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

		characters = append(characters, &character)
	}

	return characters, nil
}
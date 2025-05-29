package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"dnd-backend/internal/models"
	"github.com/lib/pq"
)

type CustomClassRepository struct {
	db *sql.DB
}

func NewCustomClassRepository(db *sql.DB) *CustomClassRepository {
	return &CustomClassRepository{db: db}
}

func (r *CustomClassRepository) Create(class *models.CustomClass) error {
	// Convert complex types to JSON
	classFeatures, err := json.Marshal(class.ClassFeatures)
	if err != nil {
		return fmt.Errorf("failed to marshal class features: %w", err)
	}

	subclasses, err := json.Marshal(class.Subclasses)
	if err != nil {
		return fmt.Errorf("failed to marshal subclasses: %w", err)
	}

	spellSlots, err := json.Marshal(class.SpellSlotsProgression)
	if err != nil {
		return fmt.Errorf("failed to marshal spell slots: %w", err)
	}

	query := `
		INSERT INTO custom_classes (
			user_id, name, description, hit_die, primary_ability,
			saving_throw_proficiencies, skill_proficiencies, skill_choices,
			starting_equipment, armor_proficiencies, weapon_proficiencies,
			tool_proficiencies, class_features, subclass_name, subclass_level,
			subclasses, spellcasting_ability, spell_list, spells_known_progression,
			cantrips_known_progression, spell_slots_progression, ritual_casting,
			spellcasting_focus, balance_score, power_level, dm_notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		) RETURNING id, created_at, updated_at`

	err = r.db.QueryRow(
		query,
		class.UserID,
		class.Name,
		class.Description,
		class.HitDie,
		class.PrimaryAbility,
		pq.Array(class.SavingThrowProficiencies),
		pq.Array(class.SkillProficiencies),
		class.SkillChoices,
		class.StartingEquipment,
		pq.Array(class.ArmorProficiencies),
		pq.Array(class.WeaponProficiencies),
		pq.Array(class.ToolProficiencies),
		classFeatures,
		class.SubclassName,
		class.SubclassLevel,
		subclasses,
		class.SpellcastingAbility,
		pq.Array(class.SpellList),
		pq.Array(class.SpellsKnownProgression),
		pq.Array(class.CantripsKnownProgression),
		spellSlots,
		class.RitualCasting,
		class.SpellcastingFocus,
		class.BalanceScore,
		class.PowerLevel,
		class.DMNotes,
	).Scan(&class.ID, &class.CreatedAt, &class.UpdatedAt)

	return err
}

func (r *CustomClassRepository) GetByID(id string) (*models.CustomClass, error) {
	query := `
		SELECT id, user_id, name, description, hit_die, primary_ability,
			saving_throw_proficiencies, skill_proficiencies, skill_choices,
			starting_equipment, armor_proficiencies, weapon_proficiencies,
			tool_proficiencies, class_features, subclass_name, subclass_level,
			subclasses, spellcasting_ability, spell_list, spells_known_progression,
			cantrips_known_progression, spell_slots_progression, ritual_casting,
			spellcasting_focus, balance_score, power_level, is_approved,
			approved_by, approved_at, dm_notes, created_at, updated_at
		FROM custom_classes
		WHERE id = $1`

	var class models.CustomClass
	var classFeatures, subclasses, spellSlots []byte

	err := r.db.QueryRow(query, id).Scan(
		&class.ID,
		&class.UserID,
		&class.Name,
		&class.Description,
		&class.HitDie,
		&class.PrimaryAbility,
		pq.Array(&class.SavingThrowProficiencies),
		pq.Array(&class.SkillProficiencies),
		&class.SkillChoices,
		&class.StartingEquipment,
		pq.Array(&class.ArmorProficiencies),
		pq.Array(&class.WeaponProficiencies),
		pq.Array(&class.ToolProficiencies),
		&classFeatures,
		&class.SubclassName,
		&class.SubclassLevel,
		&subclasses,
		&class.SpellcastingAbility,
		pq.Array(&class.SpellList),
		pq.Array(&class.SpellsKnownProgression),
		pq.Array(&class.CantripsKnownProgression),
		&spellSlots,
		&class.RitualCasting,
		&class.SpellcastingFocus,
		&class.BalanceScore,
		&class.PowerLevel,
		&class.IsApproved,
		&class.ApprovedBy,
		&class.ApprovedAt,
		&class.DMNotes,
		&class.CreatedAt,
		&class.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(classFeatures, &class.ClassFeatures); err != nil {
		return nil, fmt.Errorf("failed to unmarshal class features: %w", err)
	}

	if err := json.Unmarshal(subclasses, &class.Subclasses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subclasses: %w", err)
	}

	if err := json.Unmarshal(spellSlots, &class.SpellSlotsProgression); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spell slots: %w", err)
	}

	return &class, nil
}

func (r *CustomClassRepository) GetByUserID(userID string, includeUnapproved bool) ([]*models.CustomClass, error) {
	query := `
		SELECT id, name, description, hit_die, primary_ability,
			balance_score, power_level, is_approved, created_at
		FROM custom_classes
		WHERE user_id = $1`

	if !includeUnapproved {
		query += " AND is_approved = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.CustomClass
	for rows.Next() {
		var class models.CustomClass
		err := rows.Scan(
			&class.ID,
			&class.Name,
			&class.Description,
			&class.HitDie,
			&class.PrimaryAbility,
			&class.BalanceScore,
			&class.PowerLevel,
			&class.IsApproved,
			&class.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		classes = append(classes, &class)
	}

	return classes, nil
}

func (r *CustomClassRepository) GetApproved() ([]*models.CustomClass, error) {
	query := `
		SELECT id, name, description, hit_die, primary_ability,
			balance_score, power_level, created_at
		FROM custom_classes
		WHERE is_approved = true
		ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.CustomClass
	for rows.Next() {
		var class models.CustomClass
		err := rows.Scan(
			&class.ID,
			&class.Name,
			&class.Description,
			&class.HitDie,
			&class.PrimaryAbility,
			&class.BalanceScore,
			&class.PowerLevel,
			&class.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		classes = append(classes, &class)
	}

	return classes, nil
}

func (r *CustomClassRepository) Approve(id, approverID string) error {
	query := `
		UPDATE custom_classes
		SET is_approved = true, approved_by = $2, approved_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := r.db.Exec(query, id, approverID)
	return err
}

func (r *CustomClassRepository) Delete(id, userID string) error {
	query := `DELETE FROM custom_classes WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, id, userID)
	return err
}
package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/lib/pq"
)

type CustomClassRepository struct {
	db *DB
}

func NewCustomClassRepository(db *DB) *CustomClassRepository {
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
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		) RETURNING id, created_at, updated_at`

	err = r.db.QueryRowRebind(
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
		WHERE id = ?`

	var class models.CustomClass
	var classFeatures, subclasses, spellSlots []byte

	err := r.db.QueryRowRebind(query, id).Scan(
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
		WHERE user_id = ?`

	if !includeUnapproved {
		query += " AND is_approved = true"
	}

	query += " ORDER BY created_at DESC"

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	classes := make([]*models.CustomClass, 0, 20)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating custom classes: %w", err)
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

	rows, err := r.db.QueryContext(context.Background(), r.db.Rebind(query))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	classes := make([]*models.CustomClass, 0, 50)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating custom classes: %w", err)
	}

	return classes, nil
}

func (r *CustomClassRepository) Approve(id, approverID string) error {
	query := `
		UPDATE custom_classes
		SET is_approved = true, approved_by = ?, approved_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := r.db.ExecRebind(query, approverID, id)
	return err
}

func (r *CustomClassRepository) Delete(id, userID string) error {
	query := `DELETE FROM custom_classes WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecRebind(query, id, userID)
	return err
}

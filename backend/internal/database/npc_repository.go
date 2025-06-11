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

type npcRepository struct {
	db *sqlx.DB
}

// NewNPCRepository creates a new NPC repository
func NewNPCRepository(db *sqlx.DB) NPCRepository {
	return &npcRepository{db: db}
}

func (r *npcRepository) Create(ctx context.Context, npc *models.NPC) error {
	if npc.ID == "" {
		npc.ID = uuid.New().String()
	}

	// Convert complex fields to JSON
	speedJSON, _ := json.Marshal(npc.Speed)
	attributesJSON, _ := json.Marshal(npc.Attributes)
	savingThrowsJSON, _ := json.Marshal(npc.SavingThrows)
	skillsJSON, _ := json.Marshal(npc.Skills)
	sensesJSON, _ := json.Marshal(npc.Senses)
	abilitiesJSON, _ := json.Marshal(npc.Abilities)
	actionsJSON, _ := json.Marshal(npc.Actions)

	query := `
		INSERT INTO npcs (
			id, game_session_id, name, type, size, alignment,
			armor_class, hit_points, max_hit_points, speed,
			attributes, saving_throws, skills,
			damage_resistances, damage_immunities, condition_immunities,
			senses, languages, challenge_rating, experience_points,
			abilities, actions, legendary_actions, is_template, created_by
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?
		)`

	// Convert arrays to JSON for SQLite compatibility
	damageResistancesJSON, _ := json.Marshal(npc.DamageResistances)
	damageImmunitiesJSON, _ := json.Marshal(npc.DamageImmunities)
	conditionImmunitiesJSON, _ := json.Marshal(npc.ConditionImmunities)
	languagesJSON, _ := json.Marshal(npc.Languages)
	
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		npc.ID, npc.GameSessionID, npc.Name, npc.Type, npc.Size, npc.Alignment,
		npc.ArmorClass, npc.HitPoints, npc.MaxHitPoints, speedJSON,
		attributesJSON, savingThrowsJSON, skillsJSON,
		damageResistancesJSON, damageImmunitiesJSON, conditionImmunitiesJSON,
		sensesJSON, languagesJSON, npc.ChallengeRating, npc.ExperiencePoints,
		abilitiesJSON, actionsJSON, npc.LegendaryActions, npc.IsTemplate, npc.CreatedBy,
	)

	return err
}

func (r *npcRepository) GetByID(ctx context.Context, id string) (*models.NPC, error) {
	query := `
		SELECT 
			id, game_session_id, name, type, size, alignment,
			armor_class, hit_points, max_hit_points, speed,
			attributes, saving_throws, skills,
			damage_resistances, damage_immunities, condition_immunities,
			senses, languages, challenge_rating, experience_points,
			abilities, actions, legendary_actions, is_template, created_by,
			created_at, updated_at
		FROM npcs
		WHERE id = ?`

	var npc models.NPC
	var speedJSON, attributesJSON, savingThrowsJSON, skillsJSON []byte
	var sensesJSON, abilitiesJSON, actionsJSON []byte
	var damageResistancesJSON, damageImmunitiesJSON, conditionImmunitiesJSON, languagesJSON []byte

	query = r.db.Rebind(query)
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&npc.ID, &npc.GameSessionID, &npc.Name, &npc.Type, &npc.Size, &npc.Alignment,
		&npc.ArmorClass, &npc.HitPoints, &npc.MaxHitPoints, &speedJSON,
		&attributesJSON, &savingThrowsJSON, &skillsJSON,
		&damageResistancesJSON, &damageImmunitiesJSON, &conditionImmunitiesJSON,
		&sensesJSON, &languagesJSON, &npc.ChallengeRating, &npc.ExperiencePoints,
		&abilitiesJSON, &actionsJSON, &npc.LegendaryActions, &npc.IsTemplate, &npc.CreatedBy,
		&npc.CreatedAt, &npc.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("NPC not found")
		}
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal(speedJSON, &npc.Speed)
	json.Unmarshal(attributesJSON, &npc.Attributes)
	json.Unmarshal(savingThrowsJSON, &npc.SavingThrows)
	json.Unmarshal(skillsJSON, &npc.Skills)
	json.Unmarshal(sensesJSON, &npc.Senses)
	json.Unmarshal(abilitiesJSON, &npc.Abilities)
	json.Unmarshal(actionsJSON, &npc.Actions)
	
	// Unmarshal array fields stored as JSON
	json.Unmarshal(damageResistancesJSON, &npc.DamageResistances)
	json.Unmarshal(damageImmunitiesJSON, &npc.DamageImmunities)
	json.Unmarshal(conditionImmunitiesJSON, &npc.ConditionImmunities)
	json.Unmarshal(languagesJSON, &npc.Languages)

	return &npc, nil
}

func (r *npcRepository) GetByGameSession(ctx context.Context, gameSessionID string) ([]*models.NPC, error) {
	query := `
		SELECT 
			id, game_session_id, name, type, size, alignment,
			armor_class, hit_points, max_hit_points, speed,
			attributes, saving_throws, skills,
			damage_resistances, damage_immunities, condition_immunities,
			senses, languages, challenge_rating, experience_points,
			abilities, actions, legendary_actions, is_template, created_by,
			created_at, updated_at
		FROM npcs
		WHERE game_session_id = ? AND is_template = false
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query), gameSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var npcs []*models.NPC
	for rows.Next() {
		npc, err := r.scanNPC(rows)
		if err != nil {
			return nil, err
		}
		npcs = append(npcs, npc)
	}

	return npcs, rows.Err()
}

func (r *npcRepository) Update(ctx context.Context, npc *models.NPC) error {
	// Convert complex fields to JSON
	speedJSON, _ := json.Marshal(npc.Speed)
	attributesJSON, _ := json.Marshal(npc.Attributes)
	savingThrowsJSON, _ := json.Marshal(npc.SavingThrows)
	skillsJSON, _ := json.Marshal(npc.Skills)
	sensesJSON, _ := json.Marshal(npc.Senses)
	abilitiesJSON, _ := json.Marshal(npc.Abilities)
	actionsJSON, _ := json.Marshal(npc.Actions)

	query := `
		UPDATE npcs SET
			name = ?, type = ?, size = ?, alignment = ?,
			armor_class = ?, hit_points = ?, max_hit_points = ?, speed = ?,
			attributes = ?, saving_throws = ?, skills = ?,
			damage_resistances = ?, damage_immunities = ?, condition_immunities = ?,
			senses = ?, languages = ?, challenge_rating = ?, experience_points = ?,
			abilities = ?, actions = ?, legendary_actions = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	// Convert arrays to JSON for SQLite compatibility
	damageResistancesJSON, _ := json.Marshal(npc.DamageResistances)
	damageImmunitiesJSON, _ := json.Marshal(npc.DamageImmunities)
	conditionImmunitiesJSON, _ := json.Marshal(npc.ConditionImmunities)
	languagesJSON, _ := json.Marshal(npc.Languages)
	
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		npc.Name, npc.Type, npc.Size, npc.Alignment,
		npc.ArmorClass, npc.HitPoints, npc.MaxHitPoints, speedJSON,
		attributesJSON, savingThrowsJSON, skillsJSON,
		damageResistancesJSON, damageImmunitiesJSON, conditionImmunitiesJSON,
		sensesJSON, languagesJSON, npc.ChallengeRating, npc.ExperiencePoints,
		abilitiesJSON, actionsJSON, npc.LegendaryActions, npc.ID,
	)

	return err
}

func (r *npcRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM npcs WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *npcRepository) Search(ctx context.Context, filter models.NPCSearchFilter) ([]*models.NPC, error) {
	query := `
		SELECT 
			id, game_session_id, name, type, size, alignment,
			armor_class, hit_points, max_hit_points, speed,
			attributes, saving_throws, skills,
			damage_resistances, damage_immunities, condition_immunities,
			senses, languages, challenge_rating, experience_points,
			abilities, actions, legendary_actions, is_template, created_by,
			created_at, updated_at
		FROM npcs
		WHERE 1=1`

	args := []interface{}{}

	if filter.GameSessionID != "" {
		query += " AND game_session_id = ?"
		args = append(args, filter.GameSessionID)
	}

	if filter.Name != "" {
		query += " AND name ILIKE ?"
		args = append(args, "%"+filter.Name+"%")
	}

	if filter.Type != "" {
		query += " AND type = ?"
		args = append(args, filter.Type)
	}

	if filter.Size != "" {
		query += " AND size = ?"
		args = append(args, filter.Size)
	}

	if filter.MinCR > 0 {
		query += " AND challenge_rating >= ?"
		args = append(args, filter.MinCR)
	}

	if filter.MaxCR > 0 {
		query += " AND challenge_rating <= ?"
		args = append(args, filter.MaxCR)
	}

	if !filter.IncludeTemplates {
		query += " AND is_template = false"
	}

	query += " ORDER BY name"

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var npcs []*models.NPC
	for rows.Next() {
		npc, err := r.scanNPC(rows)
		if err != nil {
			return nil, err
		}
		npcs = append(npcs, npc)
	}

	return npcs, rows.Err()
}

func (r *npcRepository) GetTemplates(ctx context.Context) ([]*models.NPCTemplate, error) {
	query := `
		SELECT 
			id, name, source, type, size, alignment,
			armor_class, hit_dice, speed,
			attributes, saving_throws, skills,
			damage_resistances, damage_immunities, condition_immunities,
			senses, languages, challenge_rating,
			abilities, actions, legendary_actions,
			created_at
		FROM npc_templates
		ORDER BY challenge_rating, name`

	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*models.NPCTemplate
	for rows.Next() {
		template, err := r.scanNPCTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, rows.Err()
}

func (r *npcRepository) GetTemplateByID(ctx context.Context, id string) (*models.NPCTemplate, error) {
	query := `
		SELECT 
			id, name, source, type, size, alignment,
			armor_class, hit_dice, speed,
			attributes, saving_throws, skills,
			damage_resistances, damage_immunities, condition_immunities,
			senses, languages, challenge_rating,
			abilities, actions, legendary_actions,
			created_at
		FROM npc_templates
		WHERE id = ?`

	var template models.NPCTemplate
	var speedJSON, attributesJSON, savingThrowsJSON, skillsJSON []byte
	var sensesJSON, abilitiesJSON, actionsJSON []byte
	var damageResistancesJSON, damageImmunitiesJSON, conditionImmunitiesJSON, languagesJSON []byte

	query = r.db.Rebind(query)
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&template.ID, &template.Name, &template.Source, &template.Type, &template.Size, &template.Alignment,
		&template.ArmorClass, &template.HitDice, &speedJSON,
		&attributesJSON, &savingThrowsJSON, &skillsJSON,
		&damageResistancesJSON, &damageImmunitiesJSON, &conditionImmunitiesJSON,
		&sensesJSON, &languagesJSON, &template.ChallengeRating,
		&abilitiesJSON, &actionsJSON, &template.LegendaryActions,
		&template.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal(speedJSON, &template.Speed)
	json.Unmarshal(attributesJSON, &template.Attributes)
	json.Unmarshal(savingThrowsJSON, &template.SavingThrows)
	json.Unmarshal(skillsJSON, &template.Skills)
	json.Unmarshal(sensesJSON, &template.Senses)
	json.Unmarshal(abilitiesJSON, &template.Abilities)
	json.Unmarshal(actionsJSON, &template.Actions)
	
	// Unmarshal array fields stored as JSON
	json.Unmarshal(damageResistancesJSON, &template.DamageResistances)
	json.Unmarshal(damageImmunitiesJSON, &template.DamageImmunities)
	json.Unmarshal(conditionImmunitiesJSON, &template.ConditionImmunities)
	json.Unmarshal(languagesJSON, &template.Languages)

	return &template, nil
}

func (r *npcRepository) CreateFromTemplate(ctx context.Context, templateID, gameSessionID, createdBy string) (*models.NPC, error) {
	template, err := r.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Calculate hit points from hit dice
	// This is simplified - in reality we'd parse the dice notation
	hitPoints := 10 // Default

	npc := &models.NPC{
		ID:                  uuid.New().String(),
		GameSessionID:       gameSessionID,
		Name:                template.Name,
		Type:                template.Type,
		Size:                template.Size,
		Alignment:           template.Alignment,
		ArmorClass:          template.ArmorClass,
		HitPoints:           hitPoints,
		MaxHitPoints:        hitPoints,
		Speed:               template.Speed,
		Attributes:          template.Attributes,
		SavingThrows:        template.SavingThrows,
		Skills:              template.Skills,
		DamageResistances:   template.DamageResistances,
		DamageImmunities:    template.DamageImmunities,
		ConditionImmunities: template.ConditionImmunities,
		Senses:              template.Senses,
		Languages:           template.Languages,
		ChallengeRating:     template.ChallengeRating,
		Abilities:           template.Abilities,
		Actions:             template.Actions,
		LegendaryActions:    template.LegendaryActions,
		IsTemplate:          false,
		CreatedBy:           createdBy,
	}

	// Calculate experience points based on CR
	npc.ExperiencePoints = r.calculateXPFromCR(template.ChallengeRating)

	err = r.Create(ctx, npc)
	if err != nil {
		return nil, err
	}

	return npc, nil
}

// Helper functions

func (r *npcRepository) scanNPC(scanner interface{ Scan(...interface{}) error }) (*models.NPC, error) {
	var npc models.NPC
	var speedJSON, attributesJSON, savingThrowsJSON, skillsJSON []byte
	var sensesJSON, abilitiesJSON, actionsJSON []byte
	var damageResistancesJSON, damageImmunitiesJSON, conditionImmunitiesJSON, languagesJSON []byte

	err := scanner.Scan(
		&npc.ID, &npc.GameSessionID, &npc.Name, &npc.Type, &npc.Size, &npc.Alignment,
		&npc.ArmorClass, &npc.HitPoints, &npc.MaxHitPoints, &speedJSON,
		&attributesJSON, &savingThrowsJSON, &skillsJSON,
		&damageResistancesJSON, &damageImmunitiesJSON, &conditionImmunitiesJSON,
		&sensesJSON, &languagesJSON, &npc.ChallengeRating, &npc.ExperiencePoints,
		&abilitiesJSON, &actionsJSON, &npc.LegendaryActions, &npc.IsTemplate, &npc.CreatedBy,
		&npc.CreatedAt, &npc.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal(speedJSON, &npc.Speed)
	json.Unmarshal(attributesJSON, &npc.Attributes)
	json.Unmarshal(savingThrowsJSON, &npc.SavingThrows)
	json.Unmarshal(skillsJSON, &npc.Skills)
	json.Unmarshal(sensesJSON, &npc.Senses)
	json.Unmarshal(abilitiesJSON, &npc.Abilities)
	json.Unmarshal(actionsJSON, &npc.Actions)
	
	// Unmarshal array fields stored as JSON
	json.Unmarshal(damageResistancesJSON, &npc.DamageResistances)
	json.Unmarshal(damageImmunitiesJSON, &npc.DamageImmunities)
	json.Unmarshal(conditionImmunitiesJSON, &npc.ConditionImmunities)
	json.Unmarshal(languagesJSON, &npc.Languages)

	return &npc, nil
}

func (r *npcRepository) scanNPCTemplate(scanner interface{ Scan(...interface{}) error }) (*models.NPCTemplate, error) {
	var template models.NPCTemplate
	var speedJSON, attributesJSON, savingThrowsJSON, skillsJSON []byte
	var sensesJSON, abilitiesJSON, actionsJSON []byte
	var damageResistancesJSON, damageImmunitiesJSON, conditionImmunitiesJSON, languagesJSON []byte

	err := scanner.Scan(
		&template.ID, &template.Name, &template.Source, &template.Type, &template.Size, &template.Alignment,
		&template.ArmorClass, &template.HitDice, &speedJSON,
		&attributesJSON, &savingThrowsJSON, &skillsJSON,
		&damageResistancesJSON, &damageImmunitiesJSON, &conditionImmunitiesJSON,
		&sensesJSON, &languagesJSON, &template.ChallengeRating,
		&abilitiesJSON, &actionsJSON, &template.LegendaryActions,
		&template.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal(speedJSON, &template.Speed)
	json.Unmarshal(attributesJSON, &template.Attributes)
	json.Unmarshal(savingThrowsJSON, &template.SavingThrows)
	json.Unmarshal(skillsJSON, &template.Skills)
	json.Unmarshal(sensesJSON, &template.Senses)
	json.Unmarshal(abilitiesJSON, &template.Abilities)
	json.Unmarshal(actionsJSON, &template.Actions)
	
	// Unmarshal array fields stored as JSON
	json.Unmarshal(damageResistancesJSON, &template.DamageResistances)
	json.Unmarshal(damageImmunitiesJSON, &template.DamageImmunities)
	json.Unmarshal(conditionImmunitiesJSON, &template.ConditionImmunities)
	json.Unmarshal(languagesJSON, &template.Languages)

	return &template, nil
}

func (r *npcRepository) calculateXPFromCR(cr float64) int {
	// D&D 5e XP by Challenge Rating
	xpByCR := map[float64]int{
		0:    10,
		0.125: 25,
		0.25: 50,
		0.5:  100,
		1:    200,
		2:    450,
		3:    700,
		4:    1100,
		5:    1800,
		6:    2300,
		7:    2900,
		8:    3900,
		9:    5000,
		10:   5900,
		11:   7200,
		12:   8400,
		13:   10000,
		14:   11500,
		15:   13000,
		16:   15000,
		17:   18000,
		18:   20000,
		19:   22000,
		20:   25000,
		21:   33000,
		22:   41000,
		23:   50000,
		24:   62000,
		25:   75000,
		26:   90000,
		27:   105000,
		28:   120000,
		29:   135000,
		30:   155000,
	}

	if xp, ok := xpByCR[cr]; ok {
		return xp
	}
	return 0
}
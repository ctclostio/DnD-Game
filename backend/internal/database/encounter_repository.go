package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/lib/pq"
)

type EncounterRepository struct {
	db *DB
}

func NewEncounterRepository(db *DB) *EncounterRepository {
	return &EncounterRepository{db: db}
}

func (r *EncounterRepository) Create(encounter *models.Encounter) error {
	// Convert complex types to JSON
	partyComposition, _ := json.Marshal(encounter.PartyComposition)
	enemies, _ := json.Marshal(encounter.Enemies)
	enemyTactics, _ := json.Marshal(encounter.EnemyTactics)
	environmentalHazards, _ := json.Marshal(encounter.EnvironmentalHazards)
	terrainFeatures, _ := json.Marshal(encounter.TerrainFeatures)
	socialSolutions, _ := json.Marshal(encounter.SocialSolutions)
	stealthOptions, _ := json.Marshal(encounter.StealthOptions)
	environmentalSolutions, _ := json.Marshal(encounter.EnvironmentalSolutions)
	scalingOptions, _ := json.Marshal(encounter.ScalingOptions)
	reinforcementWaves, _ := json.Marshal(encounter.ReinforcementWaves)
	escapeRoutes, _ := json.Marshal(encounter.EscapeRoutes)

	query := `
		INSERT INTO encounters (
			game_session_id, created_by, name, description, location,
			encounter_type, difficulty, challenge_rating, narrative_context,
			environmental_features, story_hooks, party_level, party_size,
			party_composition, enemies, total_xp, adjusted_xp,
			enemy_tactics, environmental_hazards, terrain_features,
			social_solutions, stealth_options, environmental_solutions,
			scaling_options, reinforcement_waves, escape_routes
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContextRebind(
		context.Background(), query,
		encounter.GameSessionID,
		encounter.CreatedBy,
		encounter.Name,
		encounter.Description,
		encounter.Location,
		encounter.EncounterType,
		encounter.Difficulty,
		encounter.ChallengeRating,
		encounter.NarrativeContext,
		pq.Array(encounter.EnvironmentalFeatures),
		pq.Array(encounter.StoryHooks),
		encounter.PartyLevel,
		encounter.PartySize,
		partyComposition,
		enemies,
		encounter.TotalXP,
		encounter.AdjustedXP,
		enemyTactics,
		environmentalHazards,
		terrainFeatures,
		socialSolutions,
		stealthOptions,
		environmentalSolutions,
		scalingOptions,
		reinforcementWaves,
		escapeRoutes,
	).Scan(&encounter.ID, &encounter.CreatedAt, &encounter.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create encounter: %w", err)
	}

	// Create encounter enemies
	for _, enemy := range encounter.Enemies {
		enemy.EncounterID = encounter.ID
		if err := r.CreateEncounterEnemy(&enemy); err != nil {
			return fmt.Errorf("failed to create encounter enemy: %w", err)
		}
	}

	return nil
}

// CreateEncounterEnemy adds a new enemy to an encounter (exported for reinforcements)
func (r *EncounterRepository) CreateEncounterEnemy(enemy *models.EncounterEnemy) error {
	stats, _ := json.Marshal(enemy.Stats)
	abilities, _ := json.Marshal(enemy.Abilities)
	actions, _ := json.Marshal(enemy.Actions)
	legendaryActions, _ := json.Marshal(enemy.LegendaryActions)
	initialPosition, _ := json.Marshal(enemy.InitialPosition)
	currentPosition, _ := json.Marshal(enemy.CurrentPosition)

	query := `
		INSERT INTO encounter_enemies (
			encounter_id, npc_id, name, type, size, challenge_rating,
			hit_points, armor_class, stats, abilities, actions,
			legendary_actions, personality_traits, ideal, bond, flaw,
			tactics, morale_threshold, initial_position, current_position,
			conditions, is_alive, fled
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		) RETURNING id`

	err := r.db.QueryRowContextRebind(
		context.Background(), query,
		enemy.EncounterID,
		enemy.NPCID,
		enemy.Name,
		enemy.Type,
		enemy.Size,
		enemy.ChallengeRating,
		enemy.HitPoints,
		enemy.ArmorClass,
		stats,
		abilities,
		actions,
		legendaryActions,
		pq.Array(enemy.PersonalityTraits),
		enemy.Ideal,
		enemy.Bond,
		enemy.Flaw,
		enemy.Tactics,
		enemy.MoraleThreshold,
		initialPosition,
		currentPosition,
		pq.Array(enemy.Conditions),
		enemy.IsAlive,
		enemy.Fled,
	).Scan(&enemy.ID)

	return err
}

func (r *EncounterRepository) GetByID(id string) (*models.Encounter, error) {
	query := `
		SELECT id, game_session_id, created_by, name, description, location,
			encounter_type, difficulty, challenge_rating, narrative_context,
			environmental_features, story_hooks, party_level, party_size,
			party_composition, enemies, total_xp, adjusted_xp,
			enemy_tactics, environmental_hazards, terrain_features,
			social_solutions, stealth_options, environmental_solutions,
			scaling_options, reinforcement_waves, escape_routes,
			status, started_at, completed_at, outcome,
			created_at, updated_at
		FROM encounters
		WHERE id = ?`

	var encounter models.Encounter
	var partyComposition, enemies, enemyTactics, environmentalHazards,
		terrainFeatures, socialSolutions, stealthOptions,
		environmentalSolutions, scalingOptions, reinforcementWaves,
		escapeRoutes []byte

	err := r.db.QueryRowContextRebind(context.Background(), query, id).Scan(
		&encounter.ID,
		&encounter.GameSessionID,
		&encounter.CreatedBy,
		&encounter.Name,
		&encounter.Description,
		&encounter.Location,
		&encounter.EncounterType,
		&encounter.Difficulty,
		&encounter.ChallengeRating,
		&encounter.NarrativeContext,
		pq.Array(&encounter.EnvironmentalFeatures),
		pq.Array(&encounter.StoryHooks),
		&encounter.PartyLevel,
		&encounter.PartySize,
		&partyComposition,
		&enemies,
		&encounter.TotalXP,
		&encounter.AdjustedXP,
		&enemyTactics,
		&environmentalHazards,
		&terrainFeatures,
		&socialSolutions,
		&stealthOptions,
		&environmentalSolutions,
		&scalingOptions,
		&reinforcementWaves,
		&escapeRoutes,
		&encounter.Status,
		&encounter.StartedAt,
		&encounter.CompletedAt,
		&encounter.Outcome,
		&encounter.CreatedAt,
		&encounter.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	_ = json.Unmarshal(partyComposition, &encounter.PartyComposition)
	_ = json.Unmarshal(enemies, &encounter.Enemies)
	_ = json.Unmarshal(enemyTactics, &encounter.EnemyTactics)
	_ = json.Unmarshal(environmentalHazards, &encounter.EnvironmentalHazards)
	_ = json.Unmarshal(terrainFeatures, &encounter.TerrainFeatures)
	_ = json.Unmarshal(socialSolutions, &encounter.SocialSolutions)
	_ = json.Unmarshal(stealthOptions, &encounter.StealthOptions)
	_ = json.Unmarshal(environmentalSolutions, &encounter.EnvironmentalSolutions)
	_ = json.Unmarshal(scalingOptions, &encounter.ScalingOptions)
	_ = json.Unmarshal(reinforcementWaves, &encounter.ReinforcementWaves)
	_ = json.Unmarshal(escapeRoutes, &encounter.EscapeRoutes)

	// Load encounter enemies
	enemiesList, err := r.getEncounterEnemies(id)
	if err == nil {
		encounter.Enemies = enemiesList
	}

	return &encounter, nil
}

func (r *EncounterRepository) getEncounterEnemies(encounterID string) ([]models.EncounterEnemy, error) {
	query := `
		SELECT id, encounter_id, npc_id, name, type, size, challenge_rating,
			hit_points, armor_class, stats, abilities, actions,
			legendary_actions, personality_traits, ideal, bond, flaw,
			tactics, morale_threshold, initial_position, current_position,
			conditions, is_alive, fled
		FROM encounter_enemies
		WHERE encounter_id = ?`

	rows, err := r.db.QueryContextRebind(context.Background(), query, encounterID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	enemies := make([]models.EncounterEnemy, 0, 10)
	for rows.Next() {
		var enemy models.EncounterEnemy
		var stats, abilities, actions, legendaryActions,
			initialPosition, currentPosition []byte

		err := rows.Scan(
			&enemy.ID,
			&enemy.EncounterID,
			&enemy.NPCID,
			&enemy.Name,
			&enemy.Type,
			&enemy.Size,
			&enemy.ChallengeRating,
			&enemy.HitPoints,
			&enemy.ArmorClass,
			&stats,
			&abilities,
			&actions,
			&legendaryActions,
			pq.Array(&enemy.PersonalityTraits),
			&enemy.Ideal,
			&enemy.Bond,
			&enemy.Flaw,
			&enemy.Tactics,
			&enemy.MoraleThreshold,
			&initialPosition,
			&currentPosition,
			pq.Array(&enemy.Conditions),
			&enemy.IsAlive,
			&enemy.Fled,
		)

		if err != nil {
			continue
		}

		// Unmarshal JSON fields
		_ = json.Unmarshal(stats, &enemy.Stats)
		_ = json.Unmarshal(abilities, &enemy.Abilities)
		_ = json.Unmarshal(actions, &enemy.Actions)
		_ = json.Unmarshal(legendaryActions, &enemy.LegendaryActions)
		_ = json.Unmarshal(initialPosition, &enemy.InitialPosition)
		_ = json.Unmarshal(currentPosition, &enemy.CurrentPosition)

		enemies = append(enemies, enemy)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return enemies, nil
}

func (r *EncounterRepository) GetByGameSession(gameSessionID string) ([]*models.Encounter, error) {
	query := `
		SELECT id, name, description, location, encounter_type,
			difficulty, challenge_rating, status, created_at
		FROM encounters
		WHERE game_session_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContextRebind(context.Background(), query, gameSessionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	encounters := make([]*models.Encounter, 0, 20)
	for rows.Next() {
		var encounter models.Encounter
		err := rows.Scan(
			&encounter.ID,
			&encounter.Name,
			&encounter.Description,
			&encounter.Location,
			&encounter.EncounterType,
			&encounter.Difficulty,
			&encounter.ChallengeRating,
			&encounter.Status,
			&encounter.CreatedAt,
		)
		if err != nil {
			continue
		}
		encounters = append(encounters, &encounter)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return encounters, nil
}

func (r *EncounterRepository) UpdateStatus(id string, status string) error {
	query := `UPDATE encounters SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContextRebind(context.Background(), query, status, id)
	return err
}

func (r *EncounterRepository) StartEncounter(id string) error {
	query := `
		UPDATE encounters 
		SET status = 'active', started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := r.db.ExecContextRebind(context.Background(), query, id)
	return err
}

func (r *EncounterRepository) CompleteEncounter(id string, outcome string) error {
	query := `
		UPDATE encounters 
		SET status = 'completed', completed_at = CURRENT_TIMESTAMP, 
			outcome = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := r.db.ExecContextRebind(context.Background(), query, outcome, id)
	return err
}

func (r *EncounterRepository) CreateEvent(event *models.EncounterEvent) error {
	mechanicalEffect, _ := json.Marshal(event.MechanicalEffect)

	query := `
		INSERT INTO encounter_events (
			encounter_id, round_number, event_type, actor_type,
			actor_id, actor_name, description, mechanical_effect,
			ai_suggestion, suggestion_used
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		) RETURNING id, created_at`

	err := r.db.QueryRowContextRebind(
		context.Background(), query,
		event.EncounterID,
		event.RoundNumber,
		event.EventType,
		event.ActorType,
		event.ActorID,
		event.ActorName,
		event.Description,
		mechanicalEffect,
		event.AISuggestion,
		event.SuggestionUsed,
	).Scan(&event.ID, &event.CreatedAt)

	return err
}

func (r *EncounterRepository) GetEvents(encounterID string, limit int) ([]*models.EncounterEvent, error) {
	query := `
		SELECT id, encounter_id, round_number, event_type, actor_type,
			actor_id, actor_name, description, mechanical_effect,
			ai_suggestion, suggestion_used, created_at
		FROM encounter_events
		WHERE encounter_id = ?
		ORDER BY created_at DESC
		LIMIT ?`

	rows, err := r.db.QueryContextRebind(context.Background(), query, encounterID, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	events := make([]*models.EncounterEvent, 0, 50)
	for rows.Next() {
		var event models.EncounterEvent
		var mechanicalEffect []byte

		err := rows.Scan(
			&event.ID,
			&event.EncounterID,
			&event.RoundNumber,
			&event.EventType,
			&event.ActorType,
			&event.ActorID,
			&event.ActorName,
			&event.Description,
			&mechanicalEffect,
			&event.AISuggestion,
			&event.SuggestionUsed,
			&event.CreatedAt,
		)

		if err != nil {
			continue
		}

		_ = json.Unmarshal(mechanicalEffect, &event.MechanicalEffect)
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EncounterRepository) UpdateEnemyStatus(enemyID string, updates map[string]interface{}) error {
	// Build dynamic update query
	setClause := ""
	args := []interface{}{}

	for key, value := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = ?", key)
		args = append(args, value)
	}

	args = append(args, enemyID)
	query := fmt.Sprintf("UPDATE encounter_enemies SET %s WHERE id = ?", setClause)
	_, err := r.db.ExecContextRebind(context.Background(), query, args...)
	return err
}

func (r *EncounterRepository) CreateObjective(objective *models.EncounterObjective) error {
	successConditions, _ := json.Marshal(objective.SuccessConditions)
	failureConditions, _ := json.Marshal(objective.FailureConditions)
	itemRewards, _ := json.Marshal(objective.ItemRewards)

	query := `
		INSERT INTO encounter_objectives (
			encounter_id, type, description, success_conditions,
			failure_conditions, xp_reward, gold_reward, item_rewards,
			story_rewards
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?
		) RETURNING id, created_at`

	err := r.db.QueryRowContextRebind(
		context.Background(), query,
		objective.EncounterID,
		objective.Type,
		objective.Description,
		successConditions,
		failureConditions,
		objective.XPReward,
		objective.GoldReward,
		itemRewards,
		pq.Array(objective.StoryRewards),
	).Scan(&objective.ID, &objective.CreatedAt)

	return err
}

func (r *EncounterRepository) GetObjectives(encounterID string) ([]*models.EncounterObjective, error) {
	query := `
		SELECT id, encounter_id, type, description, success_conditions,
			failure_conditions, xp_reward, gold_reward, item_rewards,
			story_rewards, is_completed, is_failed, completed_at, created_at
		FROM encounter_objectives
		WHERE encounter_id = ?`

	rows, err := r.db.QueryContextRebind(context.Background(), query, encounterID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	objectives := make([]*models.EncounterObjective, 0, 10)
	for rows.Next() {
		var objective models.EncounterObjective
		var successConditions, failureConditions, itemRewards []byte

		err := rows.Scan(
			&objective.ID,
			&objective.EncounterID,
			&objective.Type,
			&objective.Description,
			&successConditions,
			&failureConditions,
			&objective.XPReward,
			&objective.GoldReward,
			&itemRewards,
			pq.Array(&objective.StoryRewards),
			&objective.IsCompleted,
			&objective.IsFailed,
			&objective.CompletedAt,
			&objective.CreatedAt,
		)

		if err != nil {
			continue
		}

		_ = json.Unmarshal(successConditions, &objective.SuccessConditions)
		_ = json.Unmarshal(failureConditions, &objective.FailureConditions)
		_ = json.Unmarshal(itemRewards, &objective.ItemRewards)

		objectives = append(objectives, &objective)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return objectives, nil
}

func (r *EncounterRepository) CompleteObjective(id string) error {
	query := `
		UPDATE encounter_objectives 
		SET is_completed = true, completed_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := r.db.ExecContextRebind(context.Background(), query, id)
	return err
}

func (r *EncounterRepository) FailObjective(id string) error {
	query := `
		UPDATE encounter_objectives 
		SET is_failed = true, completed_at = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := r.db.ExecContextRebind(context.Background(), query, id)
	return err
}

package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"dnd-backend/internal/models"
	"github.com/lib/pq"
)

type EncounterRepository struct {
	db *sql.DB
}

func NewEncounterRepository(db *sql.DB) *EncounterRepository {
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
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		) RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
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
		if err := r.createEncounterEnemy(&enemy); err != nil {
			return fmt.Errorf("failed to create encounter enemy: %w", err)
		}
	}

	return nil
}

func (r *EncounterRepository) createEncounterEnemy(enemy *models.EncounterEnemy) error {
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
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		) RETURNING id`

	err := r.db.QueryRow(
		query,
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
		WHERE id = $1`

	var encounter models.Encounter
	var partyComposition, enemies, enemyTactics, environmentalHazards,
		terrainFeatures, socialSolutions, stealthOptions,
		environmentalSolutions, scalingOptions, reinforcementWaves,
		escapeRoutes []byte

	err := r.db.QueryRow(query, id).Scan(
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
	json.Unmarshal(partyComposition, &encounter.PartyComposition)
	json.Unmarshal(enemies, &encounter.Enemies)
	json.Unmarshal(enemyTactics, &encounter.EnemyTactics)
	json.Unmarshal(environmentalHazards, &encounter.EnvironmentalHazards)
	json.Unmarshal(terrainFeatures, &encounter.TerrainFeatures)
	json.Unmarshal(socialSolutions, &encounter.SocialSolutions)
	json.Unmarshal(stealthOptions, &encounter.StealthOptions)
	json.Unmarshal(environmentalSolutions, &encounter.EnvironmentalSolutions)
	json.Unmarshal(scalingOptions, &encounter.ScalingOptions)
	json.Unmarshal(reinforcementWaves, &encounter.ReinforcementWaves)
	json.Unmarshal(escapeRoutes, &encounter.EscapeRoutes)

	// Load encounter enemies
	enemies_list, err := r.getEncounterEnemies(id)
	if err == nil {
		encounter.Enemies = enemies_list
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
		WHERE encounter_id = $1`

	rows, err := r.db.Query(query, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enemies []models.EncounterEnemy
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
		json.Unmarshal(stats, &enemy.Stats)
		json.Unmarshal(abilities, &enemy.Abilities)
		json.Unmarshal(actions, &enemy.Actions)
		json.Unmarshal(legendaryActions, &enemy.LegendaryActions)
		json.Unmarshal(initialPosition, &enemy.InitialPosition)
		json.Unmarshal(currentPosition, &enemy.CurrentPosition)

		enemies = append(enemies, enemy)
	}

	return enemies, nil
}

func (r *EncounterRepository) GetByGameSession(gameSessionID string) ([]*models.Encounter, error) {
	query := `
		SELECT id, name, description, location, encounter_type,
			difficulty, challenge_rating, status, created_at
		FROM encounters
		WHERE game_session_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, gameSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var encounters []*models.Encounter
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

	return encounters, nil
}

func (r *EncounterRepository) UpdateStatus(id string, status string) error {
	query := `UPDATE encounters SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(query, id, status)
	return err
}

func (r *EncounterRepository) StartEncounter(id string) error {
	query := `
		UPDATE encounters 
		SET status = 'active', started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *EncounterRepository) CompleteEncounter(id string, outcome string) error {
	query := `
		UPDATE encounters 
		SET status = 'completed', completed_at = CURRENT_TIMESTAMP, 
			outcome = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1`
	_, err := r.db.Exec(query, id, outcome)
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
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at`

	err := r.db.QueryRow(
		query,
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
		WHERE encounter_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(query, encounterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.EncounterEvent
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

		json.Unmarshal(mechanicalEffect, &event.MechanicalEffect)
		events = append(events, &event)
	}

	return events, nil
}

func (r *EncounterRepository) UpdateEnemyStatus(enemyID string, updates map[string]interface{}) error {
	// Build dynamic update query
	setClause := ""
	args := []interface{}{enemyID}
	argCount := 2

	for key, value := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argCount)
		args = append(args, value)
		argCount++
	}

	query := fmt.Sprintf("UPDATE encounter_enemies SET %s WHERE id = $1", setClause)
	_, err := r.db.Exec(query, args...)
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
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, created_at`

	err := r.db.QueryRow(
		query,
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
		WHERE encounter_id = $1`

	rows, err := r.db.Query(query, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objectives []*models.EncounterObjective
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

		json.Unmarshal(successConditions, &objective.SuccessConditions)
		json.Unmarshal(failureConditions, &objective.FailureConditions)
		json.Unmarshal(itemRewards, &objective.ItemRewards)

		objectives = append(objectives, &objective)
	}

	return objectives, nil
}

func (r *EncounterRepository) CompleteObjective(id string) error {
	query := `
		UPDATE encounter_objectives 
		SET is_completed = true, completed_at = CURRENT_TIMESTAMP 
		WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *EncounterRepository) FailObjective(id string) error {
	query := `
		UPDATE encounter_objectives 
		SET is_failed = true, completed_at = CURRENT_TIMESTAMP 
		WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
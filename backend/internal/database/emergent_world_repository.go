package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/security"
)

// EmergentWorldRepository handles database operations for the emergent world system
type EmergentWorldRepository struct {
	db *DB
}

// NewEmergentWorldRepository creates a new emergent world repository
func NewEmergentWorldRepository(db *DB) *EmergentWorldRepository {
	return &EmergentWorldRepository{db: db}
}

// World State Methods

// GetWorldState retrieves the current world state for a session
func (r *EmergentWorldRepository) GetWorldState(sessionID string) (*models.WorldState, error) {
	query := `
		SELECT id, session_id, current_time, last_simulated, world_data, 
		       is_active, created_at, updated_at
		FROM world_states
		WHERE session_id = ? AND is_active = true
		LIMIT 1
	`

	var state models.WorldState
	var worldDataJSON []byte

	err := r.db.QueryRowContextRebind(context.Background(), query, sessionID).Scan(
		&state.ID,
		&state.SessionID,
		&state.CurrentTime,
		&state.LastSimulated,
		&worldDataJSON,
		&state.IsActive,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create new world state if none exists
		return r.createWorldState(sessionID)
	} else if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(worldDataJSON, &state.WorldData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal world data: %w", err)
	}

	return &state, nil
}

// createWorldState creates a new world state
func (r *EmergentWorldRepository) createWorldState(sessionID string) (*models.WorldState, error) {
	state := &models.WorldState{
		ID:            generateUUID(),
		SessionID:     sessionID,
		CurrentTime:   time.Now(),
		LastSimulated: time.Now(),
		WorldData:     make(map[string]interface{}),
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	worldDataJSON, _ := json.Marshal(state.WorldData)

	query := `
		INSERT INTO world_states (
			id, session_id, current_time, last_simulated, 
			world_data, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		state.ID,
		state.SessionID,
		state.CurrentTime,
		state.LastSimulated,
		worldDataJSON,
		state.IsActive,
		state.CreatedAt,
		state.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return state, nil
}

// UpdateWorldState updates an existing world state
func (r *EmergentWorldRepository) UpdateWorldState(state *models.WorldState) error {
	worldDataJSON, err := json.Marshal(state.WorldData)
	if err != nil {
		return fmt.Errorf("failed to marshal world data: %w", err)
	}

	state.UpdatedAt = time.Now()

	query := `
		UPDATE world_states
		SET current_time = ?, last_simulated = ?, world_data = ?, 
		    updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContextRebind(context.Background(), query,
		state.CurrentTime,
		state.LastSimulated,
		worldDataJSON,
		state.UpdatedAt,
		state.ID,
	)

	return err
}

// NPC Goal Methods

// CreateNPCGoal creates a new NPC goal
func (r *EmergentWorldRepository) CreateNPCGoal(goal *models.NPCGoal) error {
	parametersJSON, err := json.Marshal(goal.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	query := `
		INSERT INTO npc_goals (
			id, npc_id, goal_type, priority, description,
			progress, parameters, status, started_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContextRebind(context.Background(), query,
		goal.ID,
		goal.NPCID,
		goal.GoalType,
		goal.Priority,
		goal.Description,
		goal.Progress,
		parametersJSON,
		goal.Status,
		goal.StartedAt,
		goal.CompletedAt,
	)

	return err
}

// GetNPCGoals retrieves all goals for an NPC
func (r *EmergentWorldRepository) GetNPCGoals(npcID string) ([]models.NPCGoal, error) {
	query := `
		SELECT id, npc_id, goal_type, priority, description,
		       progress, parameters, status, started_at, completed_at
		FROM npc_goals
		WHERE npc_id = ?
		ORDER BY priority DESC, started_at DESC
	`

	rows, err := r.db.QueryContextRebind(context.Background(), query, npcID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	goals := make([]models.NPCGoal, 0, 10)
	for rows.Next() {
		var goal models.NPCGoal
		var parametersJSON []byte

		err := rows.Scan(
			&goal.ID,
			&goal.NPCID,
			&goal.GoalType,
			&goal.Priority,
			&goal.Description,
			&goal.Progress,
			&parametersJSON,
			&goal.Status,
			&goal.StartedAt,
			&goal.CompletedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(parametersJSON, &goal.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		goals = append(goals, goal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return goals, nil
}

// UpdateNPCGoal updates an existing NPC goal
func (r *EmergentWorldRepository) UpdateNPCGoal(goal *models.NPCGoal) error {
	parametersJSON, err := json.Marshal(goal.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	query := `
		UPDATE npc_goals
		SET goal_type = ?, priority = ?, description = ?,
		    progress = ?, parameters = ?, status = ?, completed_at = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContextRebind(context.Background(), query,
		goal.GoalType,
		goal.Priority,
		goal.Description,
		goal.Progress,
		parametersJSON,
		goal.Status,
		goal.CompletedAt,
		goal.ID,
	)

	return err
}

// NPC Schedule Methods

// CreateNPCSchedule creates a new NPC schedule entry
func (r *EmergentWorldRepository) CreateNPCSchedule(schedule *models.NPCSchedule) error {
	parametersJSON, err := json.Marshal(schedule.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	query := `
		INSERT INTO npc_schedules (
			id, npc_id, time_of_day, activity, location, parameters
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContextRebind(context.Background(), query,
		schedule.ID,
		schedule.NPCID,
		schedule.TimeOfDay,
		schedule.Activity,
		schedule.Location,
		parametersJSON,
	)

	return err
}

// GetNPCSchedule retrieves the schedule for an NPC
func (r *EmergentWorldRepository) GetNPCSchedule(npcID string) ([]models.NPCSchedule, error) {
	query := `
		SELECT id, npc_id, time_of_day, activity, location, parameters
		FROM npc_schedules
		WHERE npc_id = ?
		ORDER BY 
			CASE time_of_day
				WHEN 'morning' THEN 1
				WHEN 'afternoon' THEN 2
				WHEN 'evening' THEN 3
				WHEN 'night' THEN 4
			END
	`

	rows, err := r.db.QueryContextRebind(context.Background(), query, npcID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	schedules := make([]models.NPCSchedule, 0, 24)
	for rows.Next() {
		var schedule models.NPCSchedule
		var parametersJSON []byte

		err := rows.Scan(
			&schedule.ID,
			&schedule.NPCID,
			&schedule.TimeOfDay,
			&schedule.Activity,
			&schedule.Location,
			&parametersJSON,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(parametersJSON, &schedule.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

// Faction Personality Methods

// CreateFactionPersonality creates a new faction personality
func (r *EmergentWorldRepository) CreateFactionPersonality(personality *models.FactionPersonality) error {
	traitsJSON, _ := json.Marshal(personality.Traits)
	valuesJSON, _ := json.Marshal(personality.Values)
	memoriesJSON, _ := json.Marshal(personality.Memories)
	decisionWeightsJSON, _ := json.Marshal(personality.DecisionWeights)
	learningDataJSON, _ := json.Marshal(personality.LearningData)

	query := `
		INSERT INTO faction_personalities (
			id, faction_id, traits, values, memories,
			current_mood, decision_weights, learning_data, last_learning_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		personality.ID,
		personality.FactionID,
		traitsJSON,
		valuesJSON,
		memoriesJSON,
		personality.CurrentMood,
		decisionWeightsJSON,
		learningDataJSON,
		personality.LastLearningTime,
	)

	return err
}

// GetFactionPersonality retrieves a faction's personality
func (r *EmergentWorldRepository) GetFactionPersonality(factionID string) (*models.FactionPersonality, error) {
	query := `
		SELECT id, faction_id, traits, values, memories,
		       current_mood, decision_weights, learning_data, last_learning_time
		FROM faction_personalities
		WHERE faction_id = ?
		LIMIT 1
	`

	var personality models.FactionPersonality
	var traitsJSON, valuesJSON, memoriesJSON, decisionWeightsJSON, learningDataJSON []byte

	err := r.db.QueryRowContextRebind(context.Background(), query, factionID).Scan(
		&personality.ID,
		&personality.FactionID,
		&traitsJSON,
		&valuesJSON,
		&memoriesJSON,
		&personality.CurrentMood,
		&decisionWeightsJSON,
		&learningDataJSON,
		&personality.LastLearningTime,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	_ = json.Unmarshal(traitsJSON, &personality.Traits)
	_ = json.Unmarshal(valuesJSON, &personality.Values)
	_ = json.Unmarshal(memoriesJSON, &personality.Memories)
	_ = json.Unmarshal(decisionWeightsJSON, &personality.DecisionWeights)
	_ = json.Unmarshal(learningDataJSON, &personality.LearningData)

	return &personality, nil
}

// UpdateFactionPersonality updates a faction personality
func (r *EmergentWorldRepository) UpdateFactionPersonality(personality *models.FactionPersonality) error {
	traitsJSON, _ := json.Marshal(personality.Traits)
	valuesJSON, _ := json.Marshal(personality.Values)
	memoriesJSON, _ := json.Marshal(personality.Memories)
	decisionWeightsJSON, _ := json.Marshal(personality.DecisionWeights)
	learningDataJSON, _ := json.Marshal(personality.LearningData)

	query := `
		UPDATE faction_personalities
		SET traits = ?, values = ?, memories = ?,
		    current_mood = ?, decision_weights = ?, 
		    learning_data = ?, last_learning_time = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		traitsJSON,
		valuesJSON,
		memoriesJSON,
		personality.CurrentMood,
		decisionWeightsJSON,
		learningDataJSON,
		personality.LastLearningTime,
		personality.ID,
	)

	return err
}

// Faction Agenda Methods

// CreateFactionAgenda creates a new faction agenda
func (r *EmergentWorldRepository) CreateFactionAgenda(agenda *models.FactionAgenda) error {
	stagesJSON, _ := json.Marshal(agenda.Stages)
	parametersJSON, _ := json.Marshal(agenda.Parameters)

	query := `
		INSERT INTO faction_agendas (
			id, faction_id, agenda_type, title, description,
			priority, stages, progress, status, parameters, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		agenda.ID,
		agenda.FactionID,
		agenda.AgendaType,
		agenda.Title,
		agenda.Description,
		agenda.Priority,
		stagesJSON,
		agenda.Progress,
		agenda.Status,
		parametersJSON,
		agenda.CreatedAt,
	)

	return err
}

// GetFactionAgendas retrieves all agendas for a faction
func (r *EmergentWorldRepository) GetFactionAgendas(factionID string) ([]models.FactionAgenda, error) {
	query := `
		SELECT id, faction_id, agenda_type, title, description,
		       priority, stages, progress, status, parameters, created_at
		FROM faction_agendas
		WHERE faction_id = ?
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.QueryContextRebind(context.Background(), query, factionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	agendas := make([]models.FactionAgenda, 0, 10)
	for rows.Next() {
		var agenda models.FactionAgenda
		var stagesJSON, parametersJSON []byte

		err := rows.Scan(
			&agenda.ID,
			&agenda.FactionID,
			&agenda.AgendaType,
			&agenda.Title,
			&agenda.Description,
			&agenda.Priority,
			&stagesJSON,
			&agenda.Progress,
			&agenda.Status,
			&parametersJSON,
			&agenda.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(stagesJSON, &agenda.Stages)
		_ = json.Unmarshal(parametersJSON, &agenda.Parameters)

		agendas = append(agendas, agenda)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return agendas, nil
}

// UpdateFactionAgenda updates a faction agenda
func (r *EmergentWorldRepository) UpdateFactionAgenda(agenda *models.FactionAgenda) error {
	stagesJSON, _ := json.Marshal(agenda.Stages)
	parametersJSON, _ := json.Marshal(agenda.Parameters)

	query := `
		UPDATE faction_agendas
		SET agenda_type = ?, title = ?, description = ?,
		    priority = ?, stages = ?, progress = ?,
		    status = ?, parameters = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		agenda.AgendaType,
		agenda.Title,
		agenda.Description,
		agenda.Priority,
		stagesJSON,
		agenda.Progress,
		agenda.Status,
		parametersJSON,
		agenda.ID,
	)

	return err
}

// Culture Methods

// CreateCulture creates a new procedural culture
func (r *EmergentWorldRepository) CreateCulture(culture *models.ProceduralCulture) error {
	languageJSON, _ := json.Marshal(culture.Language)
	customsJSON, _ := json.Marshal(culture.Customs)
	artStyleJSON, _ := json.Marshal(culture.ArtStyle)
	beliefSystemJSON, _ := json.Marshal(culture.BeliefSystem)
	valuesJSON, _ := json.Marshal(culture.Values)
	taboosJSON, _ := json.Marshal(culture.Taboos)
	greetingsJSON, _ := json.Marshal(culture.Greetings)
	architectureJSON, _ := json.Marshal(culture.Architecture)
	cuisineJSON, _ := json.Marshal(culture.Cuisine)
	musicStyleJSON, _ := json.Marshal(culture.MusicStyle)
	clothingStyleJSON, _ := json.Marshal(culture.ClothingStyle)
	namingConventionsJSON, _ := json.Marshal(culture.NamingConventions)
	socialStructureJSON, _ := json.Marshal(culture.SocialStructure)
	metadataJSON, _ := json.Marshal(culture.Metadata)

	query := `
		INSERT INTO procedural_cultures (
			id, name, language, customs, art_style, belief_system,
			values, taboos, greetings, architecture, cuisine,
			music_style, clothing_style, naming_conventions,
			social_structure, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		culture.ID,
		culture.Name,
		languageJSON,
		customsJSON,
		artStyleJSON,
		beliefSystemJSON,
		valuesJSON,
		taboosJSON,
		greetingsJSON,
		architectureJSON,
		cuisineJSON,
		musicStyleJSON,
		clothingStyleJSON,
		namingConventionsJSON,
		socialStructureJSON,
		metadataJSON,
		culture.CreatedAt,
	)

	return err
}

// GetCulture retrieves a culture by ID
func (r *EmergentWorldRepository) GetCulture(cultureID string) (*models.ProceduralCulture, error) {
	query := `
		SELECT id, name, language, customs, art_style, belief_system,
		       values, taboos, greetings, architecture, cuisine,
		       music_style, clothing_style, naming_conventions,
		       social_structure, metadata, created_at
		FROM procedural_cultures
		WHERE id = ?
	`

	var culture models.ProceduralCulture
	var languageJSON, customsJSON, artStyleJSON, beliefSystemJSON []byte
	var valuesJSON, taboosJSON, greetingsJSON, architectureJSON []byte
	var cuisineJSON, musicStyleJSON, clothingStyleJSON []byte
	var namingConventionsJSON, socialStructureJSON, metadataJSON []byte

	err := r.db.QueryRowContextRebind(context.Background(), query, cultureID).Scan(
		&culture.ID,
		&culture.Name,
		&languageJSON,
		&customsJSON,
		&artStyleJSON,
		&beliefSystemJSON,
		&valuesJSON,
		&taboosJSON,
		&greetingsJSON,
		&architectureJSON,
		&cuisineJSON,
		&musicStyleJSON,
		&clothingStyleJSON,
		&namingConventionsJSON,
		&socialStructureJSON,
		&metadataJSON,
		&culture.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal all JSON fields
	_ = json.Unmarshal(languageJSON, &culture.Language)
	_ = json.Unmarshal(customsJSON, &culture.Customs)
	_ = json.Unmarshal(artStyleJSON, &culture.ArtStyle)
	_ = json.Unmarshal(beliefSystemJSON, &culture.BeliefSystem)
	_ = json.Unmarshal(valuesJSON, &culture.Values)
	_ = json.Unmarshal(taboosJSON, &culture.Taboos)
	_ = json.Unmarshal(greetingsJSON, &culture.Greetings)
	_ = json.Unmarshal(architectureJSON, &culture.Architecture)
	_ = json.Unmarshal(cuisineJSON, &culture.Cuisine)
	_ = json.Unmarshal(musicStyleJSON, &culture.MusicStyle)
	_ = json.Unmarshal(clothingStyleJSON, &culture.ClothingStyle)
	_ = json.Unmarshal(namingConventionsJSON, &culture.NamingConventions)
	_ = json.Unmarshal(socialStructureJSON, &culture.SocialStructure)
	_ = json.Unmarshal(metadataJSON, &culture.Metadata)

	return &culture, nil
}

// GetCulturesBySession retrieves all cultures for a session
func (r *EmergentWorldRepository) GetCulturesBySession(sessionID string) ([]*models.ProceduralCulture, error) {
	query := `
		SELECT id, name, language, customs, art_style, belief_system,
		       values, taboos, greetings, architecture, cuisine,
		       music_style, clothing_style, naming_conventions,
		       social_structure, metadata, created_at
		FROM procedural_cultures
		WHERE metadata->>'session_id' = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContextRebind(context.Background(), query, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	cultures := make([]*models.ProceduralCulture, 0, 20)
	for rows.Next() {
		culture := &models.ProceduralCulture{}
		var languageJSON, customsJSON, artStyleJSON, beliefSystemJSON []byte
		var valuesJSON, taboosJSON, greetingsJSON, architectureJSON []byte
		var cuisineJSON, musicStyleJSON, clothingStyleJSON []byte
		var namingConventionsJSON, socialStructureJSON, metadataJSON []byte

		err := rows.Scan(
			&culture.ID,
			&culture.Name,
			&languageJSON,
			&customsJSON,
			&artStyleJSON,
			&beliefSystemJSON,
			&valuesJSON,
			&taboosJSON,
			&greetingsJSON,
			&architectureJSON,
			&cuisineJSON,
			&musicStyleJSON,
			&clothingStyleJSON,
			&namingConventionsJSON,
			&socialStructureJSON,
			&metadataJSON,
			&culture.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		_ = json.Unmarshal(languageJSON, &culture.Language)
		_ = json.Unmarshal(customsJSON, &culture.Customs)
		_ = json.Unmarshal(artStyleJSON, &culture.ArtStyle)
		_ = json.Unmarshal(beliefSystemJSON, &culture.BeliefSystem)
		_ = json.Unmarshal(valuesJSON, &culture.Values)
		_ = json.Unmarshal(taboosJSON, &culture.Taboos)
		_ = json.Unmarshal(greetingsJSON, &culture.Greetings)
		_ = json.Unmarshal(architectureJSON, &culture.Architecture)
		_ = json.Unmarshal(cuisineJSON, &culture.Cuisine)
		_ = json.Unmarshal(musicStyleJSON, &culture.MusicStyle)
		_ = json.Unmarshal(clothingStyleJSON, &culture.ClothingStyle)
		_ = json.Unmarshal(namingConventionsJSON, &culture.NamingConventions)
		_ = json.Unmarshal(socialStructureJSON, &culture.SocialStructure)
		_ = json.Unmarshal(metadataJSON, &culture.Metadata)

		cultures = append(cultures, culture)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cultures, nil
}

// UpdateCulture updates a procedural culture
func (r *EmergentWorldRepository) UpdateCulture(culture *models.ProceduralCulture) error {
	// Marshal only the fields that might change
	valuesJSON, _ := json.Marshal(culture.Values)
	customsJSON, _ := json.Marshal(culture.Customs)
	socialStructureJSON, _ := json.Marshal(culture.SocialStructure)
	metadataJSON, _ := json.Marshal(culture.Metadata)

	query := `
		UPDATE procedural_cultures
		SET values = ?, customs = ?, social_structure = ?, metadata = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		valuesJSON,
		customsJSON,
		socialStructureJSON,
		metadataJSON,
		culture.ID,
	)

	return err
}

// World Event Methods

// CreateWorldEvent creates a new world event
func (r *EmergentWorldRepository) CreateWorldEvent(event *models.EmergentWorldEvent) error {
	impactJSON, _ := json.Marshal(event.Impact)
	affectedEntitiesJSON, _ := json.Marshal(event.AffectedEntities)
	consequencesJSON, _ := json.Marshal(event.Consequences)

	query := `
		INSERT INTO emergent_world_events (
			id, session_id, event_type, title, description,
			impact, affected_entities, consequences,
			is_player_visible, occurred_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		event.ID,
		event.SessionID,
		event.EventType,
		event.Title,
		event.Description,
		impactJSON,
		affectedEntitiesJSON,
		consequencesJSON,
		event.IsPlayerVisible,
		event.OccurredAt,
	)

	return err
}

// GetWorldEvents retrieves world events for a session
func (r *EmergentWorldRepository) GetWorldEvents(sessionID string, limit int, onlyPlayerVisible bool) ([]models.EmergentWorldEvent, error) {
	query := `
		SELECT id, session_id, event_type, title, description,
		       impact, affected_entities, consequences,
		       is_player_visible, occurred_at
		FROM emergent_world_events
		WHERE session_id = ?
	`

	args := []interface{}{sessionID}

	if onlyPlayerVisible {
		query += " AND is_player_visible = true"
	}

	query += " ORDER BY occurred_at DESC"

	if limit > 0 {
		query += constants.LimitClause
		args = append(args, limit)
	}

	rows, err := r.db.QueryContextRebind(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	events := make([]models.EmergentWorldEvent, 0, limit)
	for rows.Next() {
		var event models.EmergentWorldEvent
		var impactJSON, affectedEntitiesJSON, consequencesJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.SessionID,
			&event.EventType,
			&event.Title,
			&event.Description,
			&impactJSON,
			&affectedEntitiesJSON,
			&consequencesJSON,
			&event.IsPlayerVisible,
			&event.OccurredAt,
		)

		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(impactJSON, &event.Impact)
		_ = json.Unmarshal(affectedEntitiesJSON, &event.AffectedEntities)
		_ = json.Unmarshal(consequencesJSON, &event.Consequences)

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// Simulation Log Methods

// CreateSimulationLog creates a new simulation log entry
func (r *EmergentWorldRepository) CreateSimulationLog(log *models.SimulationLog) error {
	detailsJSON, _ := json.Marshal(log.Details)

	query := `
		INSERT INTO simulation_logs (
			id, session_id, simulation_type, start_time, end_time,
			events_created, details, success, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContextRebind(context.Background(), query,
		log.ID,
		log.SessionID,
		log.SimulationType,
		log.StartTime,
		log.EndTime,
		log.EventsCreated,
		detailsJSON,
		log.Success,
		log.ErrorMessage,
	)

	return err
}

// GetSimulationLogs retrieves simulation logs for a session
func (r *EmergentWorldRepository) GetSimulationLogs(sessionID string, limit int) ([]models.SimulationLog, error) {
	query := `
		SELECT id, session_id, simulation_type, start_time, end_time,
		       events_created, details, success, error_message
		FROM simulation_logs
		WHERE session_id = ?
		ORDER BY start_time DESC
	`

	args := []interface{}{sessionID}

	if limit > 0 {
		query += constants.LimitClause
		args = append(args, limit)
	}

	rows, err := r.db.QueryContextRebind(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	logs := make([]models.SimulationLog, 0, limit)
	for rows.Next() {
		var log models.SimulationLog
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.SessionID,
			&log.SimulationType,
			&log.StartTime,
			&log.EndTime,
			&log.EventsCreated,
			&detailsJSON,
			&log.Success,
			&log.ErrorMessage,
		)

		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(detailsJSON, &log.Details)
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// Helper function to generate UUID
func generateUUID() string {
	// Use cryptographically secure ID generation
	id, err := security.GenerateSecureID()
	if err != nil {
		// Fallback to timestamp-based ID if secure generation fails
		// This ensures the function never fails, but logs should be added in production
		return fmt.Sprintf("%d-%x", time.Now().UnixNano(), time.Now().Unix())
	}
	return id
}

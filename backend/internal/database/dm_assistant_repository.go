package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// DMAssistantRepository defines the interface for DM Assistant database operations
type DMAssistantRepository interface {
	// NPC operations
	SaveNPC(ctx context.Context, npc *models.AINPC) error
	GetNPCByID(ctx context.Context, id uuid.UUID) (*models.AINPC, error)
	GetNPCsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AINPC, error)
	UpdateNPC(ctx context.Context, npc *models.AINPC) error
	AddNPCDialogue(ctx context.Context, npcID uuid.UUID, dialogue models.DialogueEntry) error

	// Location operations
	SaveLocation(ctx context.Context, location *models.AILocation) error
	GetLocationByID(ctx context.Context, id uuid.UUID) (*models.AILocation, error)
	GetLocationsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AILocation, error)
	UpdateLocation(ctx context.Context, location *models.AILocation) error

	// Narration operations
	SaveNarration(ctx context.Context, narration *models.AINarration) error
	GetNarrationsByType(ctx context.Context, sessionID uuid.UUID, narrationType string) ([]*models.AINarration, error)

	// Story element operations
	SaveStoryElement(ctx context.Context, element *models.AIStoryElement) error
	GetUnusedStoryElements(ctx context.Context, sessionID uuid.UUID) ([]*models.AIStoryElement, error)
	MarkStoryElementUsed(ctx context.Context, elementID uuid.UUID) error

	// Environmental hazard operations
	SaveEnvironmentalHazard(ctx context.Context, hazard *models.AIEnvironmentalHazard) error
	GetActiveHazardsByLocation(ctx context.Context, locationID uuid.UUID) ([]*models.AIEnvironmentalHazard, error)
	TriggerHazard(ctx context.Context, hazardID uuid.UUID) error

	// History operations
	SaveHistory(ctx context.Context, history *models.DMAssistantHistory) error
	GetHistoryBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.DMAssistantHistory, error)
}

// dmAssistantRepository implements DMAssistantRepository
type dmAssistantRepository struct {
	db *sqlx.DB
}

// NewDMAssistantRepository creates a new DM assistant repository
func NewDMAssistantRepository(db *sqlx.DB) DMAssistantRepository {
	return &dmAssistantRepository{db: db}
}

// NPC operations

func (r *dmAssistantRepository) SaveNPC(ctx context.Context, npc *models.AINPC) error {
	personalityJSON, err := json.Marshal(npc.PersonalityTraits)
	if err != nil {
		return fmt.Errorf("failed to marshal personality traits: %w", err)
	}

	statBlockJSON, err := json.Marshal(npc.StatBlock)
	if err != nil {
		return fmt.Errorf("failed to marshal stat block: %w", err)
	}

	dialogueJSON, err := json.Marshal(npc.GeneratedDialogue)
	if err != nil {
		return fmt.Errorf("failed to marshal dialogue: %w", err)
	}

	query := `
		INSERT INTO ai_npcs (
			id, game_session_id, name, race, occupation,
			personality_traits, appearance, voice_description,
			motivations, secrets, dialogue_style,
			relationship_to_party, stat_block, generated_dialogue,
			created_by, is_recurring, notes
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?
		)`

	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query,
		npc.ID, npc.GameSessionID, npc.Name, npc.Race, npc.Occupation,
		personalityJSON, npc.Appearance, npc.VoiceDescription,
		npc.Motivations, npc.Secrets, npc.DialogueStyle,
		npc.RelationshipToParty, statBlockJSON, dialogueJSON,
		npc.CreatedBy, npc.IsRecurring, npc.Notes,
	)

	return err
}

func (r *dmAssistantRepository) GetNPCByID(ctx context.Context, id uuid.UUID) (*models.AINPC, error) {
	var npc models.AINPC
	var personalityJSON, statBlockJSON, dialogueJSON []byte

	query := `
		SELECT 
			id, game_session_id, name, race, occupation,
			personality_traits, appearance, voice_description,
			motivations, secrets, dialogue_style,
			relationship_to_party, stat_block, generated_dialogue,
			created_by, is_recurring, last_seen_session, notes,
			created_at, updated_at
		FROM ai_npcs
		WHERE id = ?`

	query = r.db.Rebind(query)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&npc.ID, &npc.GameSessionID, &npc.Name, &npc.Race, &npc.Occupation,
		&personalityJSON, &npc.Appearance, &npc.VoiceDescription,
		&npc.Motivations, &npc.Secrets, &npc.DialogueStyle,
		&npc.RelationshipToParty, &statBlockJSON, &dialogueJSON,
		&npc.CreatedBy, &npc.IsRecurring, &npc.LastSeenSession, &npc.Notes,
		&npc.CreatedAt, &npc.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(personalityJSON, &npc.PersonalityTraits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal personality traits: %w", err)
	}

	if err := json.Unmarshal(statBlockJSON, &npc.StatBlock); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stat block: %w", err)
	}

	if err := json.Unmarshal(dialogueJSON, &npc.GeneratedDialogue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dialogue: %w", err)
	}

	return &npc, nil
}

func (r *dmAssistantRepository) GetNPCsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AINPC, error) {
	query := `
		SELECT 
			id, game_session_id, name, race, occupation,
			personality_traits, appearance, voice_description,
			motivations, secrets, dialogue_style,
			relationship_to_party, stat_block, generated_dialogue,
			created_by, is_recurring, last_seen_session, notes,
			created_at, updated_at
		FROM ai_npcs
		WHERE game_session_id = ?
		ORDER BY created_at DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	npcs := make([]*models.AINPC, 0, 20)
	for rows.Next() {
		npc, err := r.scanNPC(rows)
		if err != nil {
			return nil, err
		}
		npcs = append(npcs, npc)
	}

	return npcs, rows.Err()
}

func (r *dmAssistantRepository) UpdateNPC(ctx context.Context, npc *models.AINPC) error {
	personalityJSON, _ := json.Marshal(npc.PersonalityTraits)
	statBlockJSON, _ := json.Marshal(npc.StatBlock)
	dialogueJSON, _ := json.Marshal(npc.GeneratedDialogue)

	query := `
		UPDATE ai_npcs SET
			name = ?, race = ?, occupation = ?,
			personality_traits = ?, appearance = ?,
			voice_description = ?, motivations = ?,
			secrets = ?, dialogue_style = ?,
			relationship_to_party = ?, stat_block = ?,
			generated_dialogue = ?, is_recurring = ?,
			last_seen_session = ?, notes = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		npc.Name, npc.Race, npc.Occupation,
		personalityJSON, npc.Appearance, npc.VoiceDescription,
		npc.Motivations, npc.Secrets, npc.DialogueStyle,
		npc.RelationshipToParty, statBlockJSON, dialogueJSON,
		npc.IsRecurring, npc.LastSeenSession, npc.Notes,
		npc.ID,
	)

	return err
}

func (r *dmAssistantRepository) AddNPCDialogue(ctx context.Context, npcID uuid.UUID, dialogue models.DialogueEntry) error {
	// First get existing dialogue
	npc, err := r.GetNPCByID(ctx, npcID)
	if err != nil {
		return err
	}

	// Add new dialogue entry
	npc.GeneratedDialogue = append(npc.GeneratedDialogue, dialogue)

	// Update the NPC
	return r.UpdateNPC(ctx, npc)
}

// Location operations

func (r *dmAssistantRepository) SaveLocation(ctx context.Context, location *models.AILocation) error {
	featuresJSON, _ := json.Marshal(location.NotableFeatures)
	npcsJSON, _ := json.Marshal(location.NPCsPresent)
	actionsJSON, _ := json.Marshal(location.AvailableActions)
	secretsJSON, _ := json.Marshal(location.SecretsAndHidden)

	query := `
		INSERT INTO ai_locations (
			id, game_session_id, name, type, description,
			atmosphere, notable_features, npcs_present,
			available_actions, secrets_and_hidden,
			environmental_effects, created_by,
			parent_location_id, is_discovered
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?
		)`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		location.ID, location.GameSessionID, location.Name,
		location.Type, location.Description, location.Atmosphere,
		featuresJSON, npcsJSON, actionsJSON, secretsJSON,
		location.EnvironmentalEffects, location.CreatedBy,
		location.ParentLocationID, location.IsDiscovered,
	)

	return err
}

func (r *dmAssistantRepository) GetLocationByID(ctx context.Context, id uuid.UUID) (*models.AILocation, error) {
	var location models.AILocation
	var featuresJSON, npcsJSON, actionsJSON, secretsJSON []byte

	query := `
		SELECT 
			id, game_session_id, name, type, description,
			atmosphere, notable_features, npcs_present,
			available_actions, secrets_and_hidden,
			environmental_effects, created_by,
			parent_location_id, is_discovered,
			created_at, updated_at
		FROM ai_locations
		WHERE id = ?`

	query = r.db.Rebind(query)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&location.ID, &location.GameSessionID, &location.Name,
		&location.Type, &location.Description, &location.Atmosphere,
		&featuresJSON, &npcsJSON, &actionsJSON, &secretsJSON,
		&location.EnvironmentalEffects, &location.CreatedBy,
		&location.ParentLocationID, &location.IsDiscovered,
		&location.CreatedAt, &location.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	_ = json.Unmarshal(featuresJSON, &location.NotableFeatures)
	_ = json.Unmarshal(npcsJSON, &location.NPCsPresent)
	_ = json.Unmarshal(actionsJSON, &location.AvailableActions)
	_ = json.Unmarshal(secretsJSON, &location.SecretsAndHidden)

	return &location, nil
}

func (r *dmAssistantRepository) GetLocationsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.AILocation, error) {
	query := `
		SELECT 
			id, game_session_id, name, type, description,
			atmosphere, notable_features, npcs_present,
			available_actions, secrets_and_hidden,
			environmental_effects, created_by,
			parent_location_id, is_discovered,
			created_at, updated_at
		FROM ai_locations
		WHERE game_session_id = ?
		ORDER BY created_at DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	locations := make([]*models.AILocation, 0, 20)
	for rows.Next() {
		location, err := r.scanLocation(rows)
		if err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	return locations, rows.Err()
}

func (r *dmAssistantRepository) UpdateLocation(ctx context.Context, location *models.AILocation) error {
	featuresJSON, _ := json.Marshal(location.NotableFeatures)
	npcsJSON, _ := json.Marshal(location.NPCsPresent)
	actionsJSON, _ := json.Marshal(location.AvailableActions)
	secretsJSON, _ := json.Marshal(location.SecretsAndHidden)

	query := `
		UPDATE ai_locations SET
			name = ?, type = ?, description = ?,
			atmosphere = ?, notable_features = ?,
			npcs_present = ?, available_actions = ?,
			secrets_and_hidden = ?, environmental_effects = ?,
			parent_location_id = ?, is_discovered = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		location.Name, location.Type,
		location.Description, location.Atmosphere,
		featuresJSON, npcsJSON, actionsJSON, secretsJSON,
		location.EnvironmentalEffects, location.ParentLocationID,
		location.IsDiscovered,
		location.ID,
	)

	return err
}

// Other operations implementation...

func (r *dmAssistantRepository) SaveNarration(ctx context.Context, narration *models.AINarration) error {
	contextJSON, _ := json.Marshal(narration.Context)
	tagsJSON, _ := json.Marshal(narration.Tags)

	query := `
		INSERT INTO ai_narrations (
			id, game_session_id, type, context,
			narration, intensity_level, tags,
			created_by, used_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		narration.ID, narration.GameSessionID, narration.Type,
		contextJSON, narration.Narration, narration.IntensityLevel,
		tagsJSON, narration.CreatedBy, narration.UsedCount,
	)

	return err
}

func (r *dmAssistantRepository) GetNarrationsByType(ctx context.Context, sessionID uuid.UUID, narrationType string) ([]*models.AINarration, error) {
	query := `
		SELECT id, game_session_id, type, context,
			narration, intensity_level, tags,
			created_by, used_count, created_at
		FROM ai_narrations
		WHERE game_session_id = ? AND type = ?
		ORDER BY created_at DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, sessionID, narrationType)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	narrations := make([]*models.AINarration, 0, 10)
	for rows.Next() {
		var n models.AINarration
		var contextJSON, tagsJSON []byte

		err := rows.Scan(
			&n.ID, &n.GameSessionID, &n.Type, &contextJSON,
			&n.Narration, &n.IntensityLevel, &tagsJSON,
			&n.CreatedBy, &n.UsedCount, &n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(contextJSON, &n.Context)
		_ = json.Unmarshal(tagsJSON, &n.Tags)

		narrations = append(narrations, &n)
	}

	return narrations, rows.Err()
}

func (r *dmAssistantRepository) SaveStoryElement(ctx context.Context, element *models.AIStoryElement) error {
	contextJSON, _ := json.Marshal(element.Context)
	prereqJSON, _ := json.Marshal(element.Prerequisites)
	consequencesJSON, _ := json.Marshal(element.Consequences)
	hintsJSON, _ := json.Marshal(element.ForeshadowingHints)

	query := `
		INSERT INTO ai_story_elements (
			id, game_session_id, type, title, description,
			context, impact_level, suggested_timing,
			prerequisites, consequences, foreshadowing_hints,
			created_by, used
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?
		)`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		element.ID, element.GameSessionID, element.Type,
		element.Title, element.Description, contextJSON,
		element.ImpactLevel, element.SuggestedTiming,
		prereqJSON, consequencesJSON, hintsJSON,
		element.CreatedBy, element.Used,
	)

	return err
}

func (r *dmAssistantRepository) GetUnusedStoryElements(ctx context.Context, sessionID uuid.UUID) ([]*models.AIStoryElement, error) {
	query := `
		SELECT 
			id, game_session_id, type, title, description,
			context, impact_level, suggested_timing,
			prerequisites, consequences, foreshadowing_hints,
			created_by, used, used_at, created_at
		FROM ai_story_elements
		WHERE game_session_id = ? AND used = false
		ORDER BY created_at DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	elements := make([]*models.AIStoryElement, 0, 20)
	for rows.Next() {
		element, err := r.scanStoryElement(rows)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}

	return elements, rows.Err()
}

func (r *dmAssistantRepository) MarkStoryElementUsed(ctx context.Context, elementID uuid.UUID) error {
	query := `
		UPDATE ai_story_elements 
		SET used = true, used_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, elementID)
	return err
}

func (r *dmAssistantRepository) SaveEnvironmentalHazard(ctx context.Context, hazard *models.AIEnvironmentalHazard) error {
	mechanicalJSON, _ := json.Marshal(hazard.MechanicalEffects)

	query := `
		INSERT INTO ai_environmental_hazards (
			id, game_session_id, location_id, name, description,
			trigger_condition, effect_description, mechanical_effects,
			difficulty_class, damage_formula, avoidance_hints,
			is_trap, is_natural, reset_condition,
			created_by, is_active, triggered_count
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?
		)`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		hazard.ID, hazard.GameSessionID, hazard.LocationID,
		hazard.Name, hazard.Description, hazard.TriggerCondition,
		hazard.EffectDescription, mechanicalJSON,
		hazard.DifficultyClass, hazard.DamageFormula,
		hazard.AvoidanceHints, hazard.IsTrap, hazard.IsNatural,
		hazard.ResetCondition, hazard.CreatedBy,
		hazard.IsActive, hazard.TriggeredCount,
	)

	return err
}

func (r *dmAssistantRepository) GetActiveHazardsByLocation(ctx context.Context, locationID uuid.UUID) ([]*models.AIEnvironmentalHazard, error) {
	query := `
		SELECT 
			id, game_session_id, location_id, name, description,
			trigger_condition, effect_description, mechanical_effects,
			difficulty_class, damage_formula, avoidance_hints,
			is_trap, is_natural, reset_condition,
			created_by, is_active, triggered_count, created_at
		FROM ai_environmental_hazards
		WHERE location_id = ? AND is_active = true
		ORDER BY created_at DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	hazards := make([]*models.AIEnvironmentalHazard, 0, 10)
	for rows.Next() {
		hazard, err := r.scanHazard(rows)
		if err != nil {
			return nil, err
		}
		hazards = append(hazards, hazard)
	}

	return hazards, rows.Err()
}

func (r *dmAssistantRepository) TriggerHazard(ctx context.Context, hazardID uuid.UUID) error {
	query := `
		UPDATE ai_environmental_hazards 
		SET triggered_count = triggered_count + 1
		WHERE id = ?`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, hazardID)
	return err
}

func (r *dmAssistantRepository) SaveHistory(ctx context.Context, history *models.DMAssistantHistory) error {
	contextJSON, _ := json.Marshal(history.RequestContext)

	query := `
		INSERT INTO dm_assistant_history (
			id, game_session_id, user_id, request_type,
			request_context, prompt, response, feedback
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query,
		history.ID, history.GameSessionID, history.UserID,
		history.RequestType, contextJSON, history.Prompt,
		history.Response, history.Feedback,
	)

	return err
}

func (r *dmAssistantRepository) GetHistoryBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.DMAssistantHistory, error) {
	query := `
		SELECT 
			id, game_session_id, user_id, request_type,
			request_context, prompt, response, feedback,
			created_at
		FROM dm_assistant_history
		WHERE game_session_id = ?
		ORDER BY created_at DESC
		LIMIT ?`

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	history := make([]*models.DMAssistantHistory, 0, limit)
	for rows.Next() {
		var h models.DMAssistantHistory
		var contextJSON []byte

		err := rows.Scan(
			&h.ID, &h.GameSessionID, &h.UserID, &h.RequestType,
			&contextJSON, &h.Prompt, &h.Response, &h.Feedback,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(contextJSON, &h.RequestContext)
		history = append(history, &h)
	}

	return history, rows.Err()
}

// Helper scan functions

func (r *dmAssistantRepository) scanNPC(rows *sql.Rows) (*models.AINPC, error) {
	var npc models.AINPC
	var personalityJSON, statBlockJSON, dialogueJSON []byte

	err := rows.Scan(
		&npc.ID, &npc.GameSessionID, &npc.Name, &npc.Race, &npc.Occupation,
		&personalityJSON, &npc.Appearance, &npc.VoiceDescription,
		&npc.Motivations, &npc.Secrets, &npc.DialogueStyle,
		&npc.RelationshipToParty, &statBlockJSON, &dialogueJSON,
		&npc.CreatedBy, &npc.IsRecurring, &npc.LastSeenSession, &npc.Notes,
		&npc.CreatedAt, &npc.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(personalityJSON, &npc.PersonalityTraits)
	_ = json.Unmarshal(statBlockJSON, &npc.StatBlock)
	_ = json.Unmarshal(dialogueJSON, &npc.GeneratedDialogue)

	return &npc, nil
}

func (r *dmAssistantRepository) scanLocation(rows *sql.Rows) (*models.AILocation, error) {
	var location models.AILocation
	var featuresJSON, npcsJSON, actionsJSON, secretsJSON []byte

	err := rows.Scan(
		&location.ID, &location.GameSessionID, &location.Name,
		&location.Type, &location.Description, &location.Atmosphere,
		&featuresJSON, &npcsJSON, &actionsJSON, &secretsJSON,
		&location.EnvironmentalEffects, &location.CreatedBy,
		&location.ParentLocationID, &location.IsDiscovered,
		&location.CreatedAt, &location.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(featuresJSON, &location.NotableFeatures)
	_ = json.Unmarshal(npcsJSON, &location.NPCsPresent)
	_ = json.Unmarshal(actionsJSON, &location.AvailableActions)
	_ = json.Unmarshal(secretsJSON, &location.SecretsAndHidden)

	return &location, nil
}

func (r *dmAssistantRepository) scanStoryElement(rows *sql.Rows) (*models.AIStoryElement, error) {
	var element models.AIStoryElement
	var contextJSON, prereqJSON, consequencesJSON, hintsJSON []byte

	err := rows.Scan(
		&element.ID, &element.GameSessionID, &element.Type,
		&element.Title, &element.Description, &contextJSON,
		&element.ImpactLevel, &element.SuggestedTiming,
		&prereqJSON, &consequencesJSON, &hintsJSON,
		&element.CreatedBy, &element.Used, &element.UsedAt,
		&element.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(contextJSON, &element.Context)
	_ = json.Unmarshal(prereqJSON, &element.Prerequisites)
	_ = json.Unmarshal(consequencesJSON, &element.Consequences)
	_ = json.Unmarshal(hintsJSON, &element.ForeshadowingHints)

	return &element, nil
}

func (r *dmAssistantRepository) scanHazard(rows *sql.Rows) (*models.AIEnvironmentalHazard, error) {
	var hazard models.AIEnvironmentalHazard
	var mechanicalJSON []byte

	err := rows.Scan(
		&hazard.ID, &hazard.GameSessionID, &hazard.LocationID,
		&hazard.Name, &hazard.Description, &hazard.TriggerCondition,
		&hazard.EffectDescription, &mechanicalJSON,
		&hazard.DifficultyClass, &hazard.DamageFormula,
		&hazard.AvoidanceHints, &hazard.IsTrap, &hazard.IsNatural,
		&hazard.ResetCondition, &hazard.CreatedBy,
		&hazard.IsActive, &hazard.TriggeredCount, &hazard.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(mechanicalJSON, &hazard.MechanicalEffects)

	return &hazard, nil
}

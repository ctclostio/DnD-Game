package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/your-username/dnd-game/backend/internal/models"
)

// NarrativeRepository handles all narrative-related database operations
type NarrativeRepository struct {
	db *sqlx.DB
}

// NewNarrativeRepository creates a new narrative repository
func NewNarrativeRepository(db *sqlx.DB) *NarrativeRepository {
	return &NarrativeRepository{db: db}
}

// CreateNarrativeProfile creates a new narrative profile for a character
func (r *NarrativeRepository) CreateNarrativeProfile(profile *models.NarrativeProfile) error {
	profile.ID = uuid.New().String()
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	preferencesJSON, err := json.Marshal(profile.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	decisionHistoryJSON, err := json.Marshal(profile.DecisionHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal decision history: %w", err)
	}

	analyticsJSON, err := json.Marshal(profile.Analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	query := `
		INSERT INTO narrative_profiles (
			id, user_id, character_id, preferences, decision_history,
			play_style, analytics, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err = r.db.Exec(
		query,
		profile.ID,
		profile.UserID,
		profile.CharacterID,
		preferencesJSON,
		decisionHistoryJSON,
		profile.PlayStyle,
		analyticsJSON,
		profile.CreatedAt,
		profile.UpdatedAt,
	)

	return err
}

// GetNarrativeProfile retrieves a narrative profile by character ID
func (r *NarrativeRepository) GetNarrativeProfile(characterID string) (*models.NarrativeProfile, error) {
	var profile models.NarrativeProfile
	var preferencesJSON, decisionHistoryJSON, analyticsJSON []byte

	query := `
		SELECT id, user_id, character_id, preferences, decision_history,
			   play_style, analytics, created_at, updated_at
		FROM narrative_profiles
		WHERE character_id = ?`

	query = r.db.Rebind(query)
	err := r.db.QueryRow(query, characterID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.CharacterID,
		&preferencesJSON,
		&decisionHistoryJSON,
		&profile.PlayStyle,
		&analyticsJSON,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(preferencesJSON, &profile.Preferences); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
	}

	if err := json.Unmarshal(decisionHistoryJSON, &profile.DecisionHistory); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decision history: %w", err)
	}

	if err := json.Unmarshal(analyticsJSON, &profile.Analytics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analytics: %w", err)
	}

	return &profile, nil
}

// UpdateNarrativeProfile updates an existing narrative profile
func (r *NarrativeRepository) UpdateNarrativeProfile(profile *models.NarrativeProfile) error {
	profile.UpdatedAt = time.Now()

	preferencesJSON, err := json.Marshal(profile.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	decisionHistoryJSON, err := json.Marshal(profile.DecisionHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal decision history: %w", err)
	}

	analyticsJSON, err := json.Marshal(profile.Analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	query := `
		UPDATE narrative_profiles 
		SET preferences = ?, decision_history = ?, play_style = ?,
			analytics = ?, updated_at = ?
		WHERE id = ?`

	query = r.db.Rebind(query)
	_, err = r.db.Exec(
		query,
		preferencesJSON,
		decisionHistoryJSON,
		profile.PlayStyle,
		analyticsJSON,
		profile.UpdatedAt,
		profile.ID,
	)

	return err
}

// CreateBackstoryElement creates a new backstory element
func (r *NarrativeRepository) CreateBackstoryElement(element *models.BackstoryElement) error {
	element.ID = uuid.New().String()
	element.CreatedAt = time.Now()

	query := `
		INSERT INTO backstory_elements (
			id, character_id, type, content, weight, used,
			usage_count, tags, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err := r.db.Exec(
		query,
		element.ID,
		element.CharacterID,
		element.Type,
		element.Content,
		element.Weight,
		element.Used,
		element.UsageCount,
		pq.Array(element.Tags),
		element.CreatedAt,
	)

	return err
}

// GetBackstoryElements retrieves all backstory elements for a character
func (r *NarrativeRepository) GetBackstoryElements(characterID string) ([]models.BackstoryElement, error) {
	query := `
		SELECT id, character_id, type, content, weight, used,
			   usage_count, tags, created_at
		FROM backstory_elements
		WHERE character_id = ?
		ORDER BY weight DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.Query(query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elements []models.BackstoryElement
	for rows.Next() {
		var element models.BackstoryElement
		err := rows.Scan(
			&element.ID,
			&element.CharacterID,
			&element.Type,
			&element.Content,
			&element.Weight,
			&element.Used,
			&element.UsageCount,
			pq.Array(&element.Tags),
			&element.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}

	return elements, rows.Err()
}

// CreatePlayerAction records a player action
func (r *NarrativeRepository) CreatePlayerAction(action *models.PlayerAction) error {
	action.ID = uuid.New().String()
	action.Timestamp = time.Now()

	metadataJSON, err := json.Marshal(action.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO player_actions (
			id, session_id, character_id, action_type, target_type,
			target_id, action_description, moral_weight, immediate_result,
			potential_consequences, timestamp, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err = r.db.Exec(
		query,
		action.ID,
		action.SessionID,
		action.CharacterID,
		action.ActionType,
		action.TargetType,
		action.TargetID,
		action.ActionDescription,
		action.MoralWeight,
		action.ImmediateResult,
		action.PotentialConsequences,
		action.Timestamp,
		metadataJSON,
	)

	return err
}

// CreateConsequenceEvent creates a new consequence event
func (r *NarrativeRepository) CreateConsequenceEvent(consequence *models.ConsequenceEvent) error {
	consequence.ID = uuid.New().String()
	consequence.CreatedAt = time.Now()

	affectedEntitiesJSON, err := json.Marshal(consequence.AffectedEntities)
	if err != nil {
		return fmt.Errorf("failed to marshal affected entities: %w", err)
	}

	cascadeEffectsJSON, err := json.Marshal(consequence.CascadeEffects)
	if err != nil {
		return fmt.Errorf("failed to marshal cascade effects: %w", err)
	}

	metadataJSON, err := json.Marshal(consequence.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO consequence_events (
			id, trigger_action_id, trigger_type, description, severity,
			delay, actual_trigger_time, affected_entities, cascade_effects,
			status, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err = r.db.Exec(
		query,
		consequence.ID,
		consequence.TriggerActionID,
		consequence.TriggerType,
		consequence.Description,
		consequence.Severity,
		consequence.Delay,
		consequence.ActualTriggerTime,
		affectedEntitiesJSON,
		cascadeEffectsJSON,
		consequence.Status,
		metadataJSON,
		consequence.CreatedAt,
	)

	return err
}

// GetPendingConsequences retrieves consequences ready to trigger
func (r *NarrativeRepository) GetPendingConsequences(sessionID string, currentTime time.Time) ([]models.ConsequenceEvent, error) {
	// Calculate time thresholds for database-agnostic queries
	shortThreshold := currentTime.Add(-1 * time.Hour)
	mediumThreshold := currentTime.Add(-24 * time.Hour)
	longThreshold := currentTime.Add(-7 * 24 * time.Hour)

	query := `
		SELECT ce.id, ce.trigger_action_id, ce.trigger_type, ce.description,
			   ce.severity, ce.delay, ce.actual_trigger_time, ce.affected_entities,
			   ce.cascade_effects, ce.status, ce.metadata, ce.created_at
		FROM consequence_events ce
		JOIN player_actions pa ON ce.trigger_action_id = pa.id
		WHERE pa.session_id = ? 
		  AND ce.status = 'pending'
		  AND (
			(ce.delay = 'immediate') OR
			(ce.delay = 'short' AND ce.created_at < ?) OR
			(ce.delay = 'medium' AND ce.created_at < ?) OR
			(ce.delay = 'long' AND ce.created_at < ?)
		  )
		ORDER BY ce.severity DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.Query(query, sessionID, shortThreshold, mediumThreshold, longThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var consequences []models.ConsequenceEvent
	for rows.Next() {
		var consequence models.ConsequenceEvent
		var affectedEntitiesJSON, cascadeEffectsJSON, metadataJSON []byte

		err := rows.Scan(
			&consequence.ID,
			&consequence.TriggerActionID,
			&consequence.TriggerType,
			&consequence.Description,
			&consequence.Severity,
			&consequence.Delay,
			&consequence.ActualTriggerTime,
			&affectedEntitiesJSON,
			&cascadeEffectsJSON,
			&consequence.Status,
			&metadataJSON,
			&consequence.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(affectedEntitiesJSON, &consequence.AffectedEntities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal affected entities: %w", err)
		}

		if err := json.Unmarshal(cascadeEffectsJSON, &consequence.CascadeEffects); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cascade effects: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &consequence.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		consequences = append(consequences, consequence)
	}

	return consequences, rows.Err()
}

// CreateNarrativeEvent creates a new narrative event
func (r *NarrativeRepository) CreateNarrativeEvent(event *models.NarrativeEvent) error {
	event.ID = uuid.New().String()
	event.Timestamp = time.Now()

	playerInvolvementJSON, err := json.Marshal(event.PlayerInvolvement)
	if err != nil {
		return fmt.Errorf("failed to marshal player involvement: %w", err)
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO world_events (
			id, type, name, description, location, timestamp,
			participants, witnesses, immediate_effects, player_involvement,
			status, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err = r.db.Exec(
		query,
		event.ID,
		event.Type,
		event.Name,
		event.Description,
		event.Location,
		event.Timestamp,
		pq.Array(event.Participants),
		pq.Array(event.Witnesses),
		pq.Array(event.ImmediateEffects),
		playerInvolvementJSON,
		event.Status,
		metadataJSON,
	)

	return err
}

// CreatePerspectiveNarrative creates a new perspective on an event
func (r *NarrativeRepository) CreatePerspectiveNarrative(perspective *models.PerspectiveNarrative) error {
	perspective.ID = uuid.New().String()
	perspective.CreatedAt = time.Now()

	contradictionsJSON, err := json.Marshal(perspective.Contradictions)
	if err != nil {
		return fmt.Errorf("failed to marshal contradictions: %w", err)
	}

	culturalContextJSON, err := json.Marshal(perspective.CulturalContext)
	if err != nil {
		return fmt.Errorf("failed to marshal cultural context: %w", err)
	}

	query := `
		INSERT INTO perspective_narratives (
			id, event_id, perspective_type, source_id, source_name,
			narrative, bias, truth_level, hidden_details, contradictions,
			emotional_tone, cultural_context, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err = r.db.Exec(
		query,
		perspective.ID,
		perspective.EventID,
		perspective.PerspectiveType,
		perspective.SourceID,
		perspective.SourceName,
		perspective.Narrative,
		perspective.Bias,
		perspective.TruthLevel,
		pq.Array(perspective.HiddenDetails),
		contradictionsJSON,
		perspective.EmotionalTone,
		culturalContextJSON,
		perspective.CreatedAt,
	)

	return err
}

// GetEventPerspectives retrieves all perspectives for an event
func (r *NarrativeRepository) GetEventPerspectives(eventID string) ([]models.PerspectiveNarrative, error) {
	query := `
		SELECT id, event_id, perspective_type, source_id, source_name,
			   narrative, bias, truth_level, hidden_details, contradictions,
			   emotional_tone, cultural_context, created_at
		FROM perspective_narratives
		WHERE event_id = ?
		ORDER BY created_at`

	query = r.db.Rebind(query)
	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perspectives []models.PerspectiveNarrative
	for rows.Next() {
		var perspective models.PerspectiveNarrative
		var contradictionsJSON, culturalContextJSON []byte

		err := rows.Scan(
			&perspective.ID,
			&perspective.EventID,
			&perspective.PerspectiveType,
			&perspective.SourceID,
			&perspective.SourceName,
			&perspective.Narrative,
			&perspective.Bias,
			&perspective.TruthLevel,
			pq.Array(&perspective.HiddenDetails),
			&contradictionsJSON,
			&perspective.EmotionalTone,
			&culturalContextJSON,
			&perspective.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(contradictionsJSON, &perspective.Contradictions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal contradictions: %w", err)
		}

		if err := json.Unmarshal(culturalContextJSON, &perspective.CulturalContext); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cultural context: %w", err)
		}

		perspectives = append(perspectives, perspective)
	}

	return perspectives, rows.Err()
}

// CreateNarrativeMemory stores a narrative memory
func (r *NarrativeRepository) CreateNarrativeMemory(memory *models.NarrativeMemory) error {
	memory.ID = uuid.New().String()
	memory.CreatedAt = time.Now()
	memory.LastReferenced = time.Now()

	metadataJSON, err := json.Marshal(memory.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO narrative_memories (
			id, session_id, character_id, memory_type, content,
			emotional_weight, connections, active, last_referenced,
			reference_count, tags, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err = r.db.Exec(
		query,
		memory.ID,
		memory.SessionID,
		memory.CharacterID,
		memory.MemoryType,
		memory.Content,
		memory.EmotionalWeight,
		pq.Array(memory.Connections),
		memory.Active,
		memory.LastReferenced,
		memory.ReferenceCount,
		pq.Array(memory.Tags),
		metadataJSON,
		memory.CreatedAt,
	)

	return err
}

// GetActiveMemories retrieves active memories for a character
func (r *NarrativeRepository) GetActiveMemories(characterID string, limit int) ([]models.NarrativeMemory, error) {
	query := `
		SELECT id, session_id, character_id, memory_type, content,
			   emotional_weight, connections, active, last_referenced,
			   reference_count, tags, metadata, created_at
		FROM narrative_memories
		WHERE character_id = ? AND active = true
		ORDER BY emotional_weight DESC, last_referenced DESC
		LIMIT ?`

	query = r.db.Rebind(query)
	rows, err := r.db.Query(query, characterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []models.NarrativeMemory
	for rows.Next() {
		var memory models.NarrativeMemory
		var metadataJSON []byte

		err := rows.Scan(
			&memory.ID,
			&memory.SessionID,
			&memory.CharacterID,
			&memory.MemoryType,
			&memory.Content,
			&memory.EmotionalWeight,
			pq.Array(&memory.Connections),
			&memory.Active,
			&memory.LastReferenced,
			&memory.ReferenceCount,
			pq.Array(&memory.Tags),
			&metadataJSON,
			&memory.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &memory.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		memories = append(memories, memory)
	}

	return memories, rows.Err()
}

// UpdateConsequenceStatus updates the status of a consequence event
func (r *NarrativeRepository) UpdateConsequenceStatus(consequenceID string, status string, triggerTime *time.Time) error {
	query := `
		UPDATE consequence_events
		SET status = ?, actual_trigger_time = ?
		WHERE id = ?`

	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, status, triggerTime, consequenceID)
	return err
}

// IncrementBackstoryUsage marks a backstory element as used
func (r *NarrativeRepository) IncrementBackstoryUsage(elementID string) error {
	query := `
		UPDATE backstory_elements
		SET used = true, usage_count = usage_count + 1
		WHERE id = ?`

	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, elementID)
	return err
}

// CreateNarrativeThread creates a new narrative thread
func (r *NarrativeRepository) CreateNarrativeThread(thread *models.NarrativeThread) error {
	thread.ID = uuid.New().String()
	thread.CreatedAt = time.Now()
	thread.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(thread.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO narrative_threads (
			id, name, description, thread_type, status,
			connected_events, key_participants, tension_level,
			resolution_proximity, created_at, updated_at, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	query = r.db.Rebind(query)
	_, err = r.db.Exec(
		query,
		thread.ID,
		thread.Name,
		thread.Description,
		thread.ThreadType,
		thread.Status,
		pq.Array(thread.ConnectedEvents),
		pq.Array(thread.KeyParticipants),
		thread.TensionLevel,
		thread.ResolutionProximity,
		thread.CreatedAt,
		thread.UpdatedAt,
		metadataJSON,
	)

	return err
}

// GetActiveNarrativeThreads retrieves all active narrative threads
func (r *NarrativeRepository) GetActiveNarrativeThreads() ([]models.NarrativeThread, error) {
	query := `
		SELECT id, name, description, thread_type, status,
			   connected_events, key_participants, tension_level,
			   resolution_proximity, created_at, updated_at, 
			   resolved_at, metadata
		FROM narrative_threads
		WHERE status IN ('active', 'dormant')
		ORDER BY tension_level DESC`

	query = r.db.Rebind(query)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []models.NarrativeThread
	for rows.Next() {
		var thread models.NarrativeThread
		var metadataJSON []byte
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&thread.ID,
			&thread.Name,
			&thread.Description,
			&thread.ThreadType,
			&thread.Status,
			pq.Array(&thread.ConnectedEvents),
			pq.Array(&thread.KeyParticipants),
			&thread.TensionLevel,
			&thread.ResolutionProximity,
			&thread.CreatedAt,
			&thread.UpdatedAt,
			&resolvedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		if resolvedAt.Valid {
			thread.ResolvedAt = &resolvedAt.Time
		}

		if err := json.Unmarshal(metadataJSON, &thread.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		threads = append(threads, thread)
	}

	return threads, rows.Err()
}

// UpdatePlayerAction updates an existing player action
func (r *NarrativeRepository) UpdatePlayerAction(action *models.PlayerAction) error {
	// TODO: Implement update logic
	return nil
}

// GetWorldEvent retrieves a narrative event by ID
func (r *NarrativeRepository) GetWorldEvent(eventID string) (*models.NarrativeEvent, error) {
	// TODO: Implement retrieval logic
	return &models.NarrativeEvent{}, nil
}

// CreateWorldEvent creates a new narrative event
func (r *NarrativeRepository) CreateWorldEvent(event *models.NarrativeEvent) error {
	// TODO: Implement creation logic
	return nil
}

// CreatePersonalizedNarrative saves a personalized narrative
func (r *NarrativeRepository) CreatePersonalizedNarrative(narrative *models.PersonalizedNarrative) error {
	// TODO: Implement creation logic
	return nil
}


// Add NarrativeThread model if not in models package
type NarrativeThread struct {
	ID                  string                 `json:"id" db:"id"`
	Name                string                 `json:"name" db:"name"`
	Description         string                 `json:"description" db:"description"`
	ThreadType          string                 `json:"thread_type" db:"thread_type"`
	Status              string                 `json:"status" db:"status"`
	ConnectedEvents     []string               `json:"connected_events" db:"connected_events"`
	KeyParticipants     []string               `json:"key_participants" db:"key_participants"`
	TensionLevel        float64                `json:"tension_level" db:"tension_level"`
	ResolutionProximity float64                `json:"resolution_proximity" db:"resolution_proximity"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
	ResolvedAt          *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`
}
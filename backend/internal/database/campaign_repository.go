package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CampaignRepository interface {
	// Story Arc methods.
	CreateStoryArc(arc *models.StoryArc) error
	GetStoryArc(id uuid.UUID) (*models.StoryArc, error)
	GetStoryArcsBySession(sessionID uuid.UUID) ([]*models.StoryArc, error)
	UpdateStoryArc(id uuid.UUID, updates map[string]interface{}) error
	DeleteStoryArc(id uuid.UUID) error

	// Session Memory methods.
	CreateSessionMemory(memory *models.SessionMemory) error
	GetSessionMemory(id uuid.UUID) (*models.SessionMemory, error)
	GetSessionMemories(sessionID uuid.UUID, limit int) ([]*models.SessionMemory, error)
	GetLatestSessionMemory(sessionID uuid.UUID) (*models.SessionMemory, error)
	UpdateSessionMemory(id uuid.UUID, updates map[string]interface{}) error

	// Plot Thread methods.
	CreatePlotThread(thread *models.PlotThread) error
	GetPlotThread(id uuid.UUID) (*models.PlotThread, error)
	GetPlotThreadsBySession(sessionID uuid.UUID) ([]*models.PlotThread, error)
	GetActivePlotThreads(sessionID uuid.UUID) ([]*models.PlotThread, error)
	UpdatePlotThread(id uuid.UUID, updates map[string]interface{}) error
	DeletePlotThread(id uuid.UUID) error

	// Foreshadowing methods.
	CreateForeshadowingElement(element *models.ForeshadowingElement) error
	GetForeshadowingElement(id uuid.UUID) (*models.ForeshadowingElement, error)
	GetUnrevealedForeshadowing(sessionID uuid.UUID) ([]*models.ForeshadowingElement, error)
	RevealForeshadowing(id uuid.UUID, sessionNumber int) error

	// Timeline methods.
	CreateTimelineEvent(event *models.CampaignTimeline) error
	GetTimelineEvents(sessionID uuid.UUID, startDate, endDate time.Time) ([]*models.CampaignTimeline, error)

	// NPC Relationship methods.
	CreateOrUpdateNPCRelationship(relationship *models.NPCRelationship) error
	GetNPCRelationships(sessionID uuid.UUID, npcID uuid.UUID) ([]*models.NPCRelationship, error)
	UpdateRelationshipScore(sessionID, npcID, targetID uuid.UUID, scoreDelta int) error
}

type campaignRepository struct {
	db *sqlx.DB
}

func NewCampaignRepository(db *sqlx.DB) CampaignRepository {
	return &campaignRepository{db: db}
}

// Story Arc methods.
func (r *campaignRepository) CreateStoryArc(arc *models.StoryArc) error {
	query := `
		INSERT INTO story_arcs (
			id, game_session_id, title, description, arc_type, 
			status, parent_arc_id, importance_level, metadata
		) VALUES (
			:id, :game_session_id, :title, :description, :arc_type,
			:status, :parent_arc_id, :importance_level, :metadata
		)`

	if arc.ID == uuid.Nil {
		arc.ID = uuid.New()
	}
	if arc.Status == "" {
		arc.Status = constants.StatusActive
	}
	if arc.ImportanceLevel == 0 {
		arc.ImportanceLevel = 5
	}

	_, err := r.db.NamedExec(query, arc)
	return err
}

func (r *campaignRepository) GetStoryArc(id uuid.UUID) (*models.StoryArc, error) {
	var arc models.StoryArc
	query := `SELECT * FROM story_arcs WHERE id = ?`
	query = r.db.Rebind(query)
	err := r.db.Get(&arc, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("story arc not found")
	}
	return &arc, err
}

func (r *campaignRepository) GetStoryArcsBySession(sessionID uuid.UUID) ([]*models.StoryArc, error) {
	var arcs []*models.StoryArc
	query := `
		SELECT * FROM story_arcs 
		WHERE game_session_id = ? 
		ORDER BY importance_level DESC, created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.Select(&arcs, query, sessionID)
	return arcs, err
}

func (r *campaignRepository) UpdateStoryArc(id uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	query, args := buildUpdateQuery("story_arcs", id, updates)
	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *campaignRepository) DeleteStoryArc(id uuid.UUID) error {
	query := `DELETE FROM story_arcs WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, id)
	return err
}

// Session Memory methods.
func (r *campaignRepository) CreateSessionMemory(memory *models.SessionMemory) error {
	query := `
		INSERT INTO session_memories (
			id, game_session_id, session_number, session_date, recap_summary,
			key_events, npcs_encountered, decisions_made, items_acquired,
			locations_visited, combat_encounters, plot_developments
		) VALUES (
			:id, :game_session_id, :session_number, :session_date, :recap_summary,
			:key_events, :npcs_encountered, :decisions_made, :items_acquired,
			:locations_visited, :combat_encounters, :plot_developments
		)`

	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}

	_, err := r.db.NamedExec(query, memory)
	return err
}

func (r *campaignRepository) GetSessionMemory(id uuid.UUID) (*models.SessionMemory, error) {
	var memory models.SessionMemory
	query := `SELECT * FROM session_memories WHERE id = ?`
	query = r.db.Rebind(query)
	err := r.db.Get(&memory, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session memory not found")
	}
	return &memory, err
}

func (r *campaignRepository) GetSessionMemories(sessionID uuid.UUID, limit int) ([]*models.SessionMemory, error) {
	var memories []*models.SessionMemory
	query := `
		SELECT * FROM session_memories 
		WHERE game_session_id = ? 
		ORDER BY session_date DESC
		LIMIT ?`
	query = r.db.Rebind(query)
	err := r.db.Select(&memories, query, sessionID, limit)
	return memories, err
}

func (r *campaignRepository) GetLatestSessionMemory(sessionID uuid.UUID) (*models.SessionMemory, error) {
	var memory models.SessionMemory
	query := `
		SELECT * FROM session_memories 
		WHERE game_session_id = ? 
		ORDER BY session_date DESC 
		LIMIT 1`
	query = r.db.Rebind(query)
	err := r.db.Get(&memory, query, sessionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &memory, err
}

func (r *campaignRepository) UpdateSessionMemory(id uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	query, args := buildUpdateQuery("session_memories", id, updates)
	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, args...)
	return err
}

// Plot Thread methods.
func (r *campaignRepository) CreatePlotThread(thread *models.PlotThread) error {
	query := `
		INSERT INTO plot_threads (
			id, game_session_id, story_arc_id, thread_type, title,
			description, status, tension_level, introduced_session,
			related_npcs, related_locations, foreshadowing_hints,
			resolution_conditions
		) VALUES (
			:id, :game_session_id, :story_arc_id, :thread_type, :title,
			:description, :status, :tension_level, :introduced_session,
			:related_npcs, :related_locations, :foreshadowing_hints,
			:resolution_conditions
		)`

	if thread.ID == uuid.Nil {
		thread.ID = uuid.New()
	}
	if thread.Status == "" {
		thread.Status = "active"
	}
	if thread.TensionLevel == 0 {
		thread.TensionLevel = 5
	}

	_, err := r.db.NamedExec(query, thread)
	return err
}

func (r *campaignRepository) GetPlotThread(id uuid.UUID) (*models.PlotThread, error) {
	var thread models.PlotThread
	query := `SELECT * FROM plot_threads WHERE id = ?`
	query = r.db.Rebind(query)
	err := r.db.Get(&thread, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plot thread not found")
	}
	return &thread, err
}

func (r *campaignRepository) GetPlotThreadsBySession(sessionID uuid.UUID) ([]*models.PlotThread, error) {
	var threads []*models.PlotThread
	query := `
		SELECT * FROM plot_threads 
		WHERE game_session_id = ? 
		ORDER BY tension_level DESC, created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.Select(&threads, query, sessionID)
	return threads, err
}

func (r *campaignRepository) GetActivePlotThreads(sessionID uuid.UUID) ([]*models.PlotThread, error) {
	var threads []*models.PlotThread
	query := `
		SELECT * FROM plot_threads 
		WHERE game_session_id = ? AND status = 'active'
		ORDER BY tension_level DESC, created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.Select(&threads, query, sessionID)
	return threads, err
}

func (r *campaignRepository) UpdatePlotThread(id uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	query, args := buildUpdateQuery("plot_threads", id, updates)
	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *campaignRepository) DeletePlotThread(id uuid.UUID) error {
	query := `DELETE FROM plot_threads WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, id)
	return err
}

// Foreshadowing methods.
func (r *campaignRepository) CreateForeshadowingElement(element *models.ForeshadowingElement) error {
	query := `
		INSERT INTO foreshadowing_elements (
			id, game_session_id, plot_thread_id, story_arc_id,
			element_type, content, subtlety_level, revealed,
			placement_suggestions
		) VALUES (
			:id, :game_session_id, :plot_thread_id, :story_arc_id,
			:element_type, :content, :subtlety_level, :revealed,
			:placement_suggestions
		)`

	if element.ID == uuid.Nil {
		element.ID = uuid.New()
	}
	if element.SubtletyLevel == 0 {
		element.SubtletyLevel = 5
	}

	_, err := r.db.NamedExec(query, element)
	return err
}

func (r *campaignRepository) GetForeshadowingElement(id uuid.UUID) (*models.ForeshadowingElement, error) {
	var element models.ForeshadowingElement
	query := `SELECT * FROM foreshadowing_elements WHERE id = ?`
	query = r.db.Rebind(query)
	err := r.db.Get(&element, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("foreshadowing element not found")
	}
	return &element, err
}

func (r *campaignRepository) GetUnrevealedForeshadowing(sessionID uuid.UUID) ([]*models.ForeshadowingElement, error) {
	var elements []*models.ForeshadowingElement
	query := `
		SELECT * FROM foreshadowing_elements 
		WHERE game_session_id = ? AND revealed = false
		ORDER BY subtlety_level ASC, created_at ASC`
	query = r.db.Rebind(query)
	err := r.db.Select(&elements, query, sessionID)
	return elements, err
}

func (r *campaignRepository) RevealForeshadowing(id uuid.UUID, sessionNumber int) error {
	query := `
		UPDATE foreshadowing_elements 
		SET revealed = true, revealed_session = ?, updated_at = ?
		WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, sessionNumber, time.Now(), id)
	return err
}

// Timeline methods.
func (r *campaignRepository) CreateTimelineEvent(event *models.CampaignTimeline) error {
	query := `
		INSERT INTO campaign_timeline (
			id, game_session_id, session_memory_id, event_date,
			real_session_date, event_type, event_title, event_description,
			impact_level, related_arcs, related_threads
		) VALUES (
			:id, :game_session_id, :session_memory_id, :event_date,
			:real_session_date, :event_type, :event_title, :event_description,
			:impact_level, :related_arcs, :related_threads
		)`

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.ImpactLevel == 0 {
		event.ImpactLevel = 5
	}

	_, err := r.db.NamedExec(query, event)
	return err
}

func (r *campaignRepository) GetTimelineEvents(sessionID uuid.UUID, startDate, endDate time.Time) ([]*models.CampaignTimeline, error) {
	var events []*models.CampaignTimeline
	query := `
		SELECT * FROM campaign_timeline 
		WHERE game_session_id = ? 
		AND event_date BETWEEN ? AND ?
		ORDER BY event_date ASC`
	query = r.db.Rebind(query)
	err := r.db.Select(&events, query, sessionID, startDate, endDate)
	return events, err
}

// NPC Relationship methods.
func (r *campaignRepository) CreateOrUpdateNPCRelationship(relationship *models.NPCRelationship) error {
	if relationship.ID == uuid.Nil {
		relationship.ID = uuid.New()
	}

	// First, try to find existing relationship.
	var existingID uuid.UUID
	checkQuery := `
		SELECT id FROM npc_relationships 
		WHERE game_session_id = ? AND npc_id = ? AND target_id = ?`
	checkQuery = r.db.Rebind(checkQuery)
	err := r.db.Get(&existingID, checkQuery, relationship.GameSessionID, relationship.NPCID, relationship.TargetID)

	if err == sql.ErrNoRows {
		// Insert new relationship.
		insertQuery := `
			INSERT INTO npc_relationships (
				id, game_session_id, npc_id, target_type, target_id,
				relationship_type, relationship_score, last_interaction_session,
				interaction_history
			) VALUES (
				:id, :game_session_id, :npc_id, :target_type, :target_id,
				:relationship_type, :relationship_score, :last_interaction_session,
				:interaction_history
			)`
		_, err = r.db.NamedExec(insertQuery, relationship)
		return err
	} else if err != nil {
		return err
	}

	// Update existing relationship.
	updateQuery := `
		UPDATE npc_relationships SET
			relationship_type = ?,
			relationship_score = ?,
			last_interaction_session = ?,
			interaction_history = ?,
			updated_at = ?
		WHERE id = ?`
	updateQuery = r.db.Rebind(updateQuery)
	_, err = r.db.Exec(updateQuery,
		relationship.RelationshipType,
		relationship.RelationshipScore,
		relationship.LastInteractionSession,
		relationship.InteractionHistory,
		time.Now(),
		existingID,
	)
	return err
}

func (r *campaignRepository) GetNPCRelationships(sessionID uuid.UUID, npcID uuid.UUID) ([]*models.NPCRelationship, error) {
	var relationships []*models.NPCRelationship
	query := `
		SELECT * FROM npc_relationships 
		WHERE game_session_id = ? AND npc_id = ?
		ORDER BY relationship_score DESC`
	query = r.db.Rebind(query)
	err := r.db.Select(&relationships, query, sessionID, npcID)
	return relationships, err
}

func (r *campaignRepository) UpdateRelationshipScore(sessionID, npcID, targetID uuid.UUID, scoreDelta int) error {
	query := `
		UPDATE npc_relationships 
		SET relationship_score = GREATEST(-100, LEAST(100, relationship_score + ?)),
		    updated_at = CURRENT_TIMESTAMP
		WHERE game_session_id = ? AND npc_id = ? AND target_id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.Exec(query, scoreDelta, sessionID, npcID, targetID)
	return err
}

// Helper function to build dynamic update queries.
func buildUpdateQuery(table string, id uuid.UUID, updates map[string]interface{}) (string, []interface{}) {
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)

	for column, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", column))
		args = append(args, value)
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		table,
		joinStrings(setClauses, ", "),
	)

	return query, args
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

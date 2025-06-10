package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/your-username/dnd-game/backend/internal/models"
)

type CombatAnalyticsRepository interface {
	// Combat Analytics methods
	CreateCombatAnalytics(analytics *models.CombatAnalytics) error
	GetCombatAnalytics(combatID uuid.UUID) (*models.CombatAnalytics, error)
	GetCombatAnalyticsBySession(sessionID uuid.UUID) ([]*models.CombatAnalytics, error)
	UpdateCombatAnalytics(id uuid.UUID, updates map[string]interface{}) error

	// Combatant Analytics methods
	CreateCombatantAnalytics(analytics *models.CombatantAnalytics) error
	GetCombatantAnalytics(combatAnalyticsID uuid.UUID) ([]*models.CombatantAnalytics, error)
	UpdateCombatantAnalytics(id uuid.UUID, updates map[string]interface{}) error

	// Auto Combat Resolution methods
	CreateAutoCombatResolution(resolution *models.AutoCombatResolution) error
	GetAutoCombatResolution(id uuid.UUID) (*models.AutoCombatResolution, error)
	GetAutoCombatResolutionsBySession(sessionID uuid.UUID) ([]*models.AutoCombatResolution, error)

	// Battle Map methods
	CreateBattleMap(battleMap *models.BattleMap) error
	GetBattleMap(id uuid.UUID) (*models.BattleMap, error)
	GetBattleMapByCombat(combatID uuid.UUID) (*models.BattleMap, error)
	GetBattleMapsBySession(sessionID uuid.UUID) ([]*models.BattleMap, error)
	UpdateBattleMap(id uuid.UUID, updates map[string]interface{}) error

	// Smart Initiative methods
	CreateOrUpdateInitiativeRule(rule *models.SmartInitiativeRule) error
	GetInitiativeRule(sessionID uuid.UUID, entityID string) (*models.SmartInitiativeRule, error)
	GetInitiativeRulesBySession(sessionID uuid.UUID) ([]*models.SmartInitiativeRule, error)

	// Combat Action Log methods
	CreateCombatAction(action *models.CombatActionLog) error
	GetCombatActions(combatID uuid.UUID) ([]*models.CombatActionLog, error)
	GetCombatActionsByRound(combatID uuid.UUID, roundNumber int) ([]*models.CombatActionLog, error)
}

type combatAnalyticsRepository struct {
	db *sqlx.DB
}

func NewCombatAnalyticsRepository(db *sqlx.DB) CombatAnalyticsRepository {
	return &combatAnalyticsRepository{db: db}
}

// Combat Analytics methods

func (r *combatAnalyticsRepository) CreateCombatAnalytics(analytics *models.CombatAnalytics) error {
	query := `
		INSERT INTO combat_analytics (
			id, combat_id, game_session_id, combat_duration,
			total_damage_dealt, total_healing_done, killing_blows,
			combat_summary, mvp_id, mvp_type, tactical_rating
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if analytics.ID == uuid.Nil {
		analytics.ID = uuid.New()
	}
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()

	query = r.db.Rebind(query)
	_, err := r.db.Exec(query,
		analytics.ID,
		analytics.CombatID,
		analytics.GameSessionID,
		analytics.CombatDuration,
		analytics.TotalDamageDealt,
		analytics.TotalHealingDone,
		analytics.KillingBlows,
		analytics.CombatSummary,
		analytics.MVPID,
		analytics.MVPType,
		analytics.TacticalRating)
	return err
}

func (r *combatAnalyticsRepository) GetCombatAnalytics(combatID uuid.UUID) (*models.CombatAnalytics, error) {
	var analytics models.CombatAnalytics
	query := `SELECT * FROM combat_analytics WHERE combat_id = ?`
	query = r.db.Rebind(query)
	err := r.db.Get(&analytics, query, combatID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("combat analytics not found")
	}
	return &analytics, err
}

func (r *combatAnalyticsRepository) GetCombatAnalyticsBySession(sessionID uuid.UUID) ([]*models.CombatAnalytics, error) {
	var analytics []*models.CombatAnalytics
	query := `
		SELECT * FROM combat_analytics 
		WHERE game_session_id = ? 
		ORDER BY created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.Select(&analytics, query, sessionID)
	return analytics, err
}

func (r *combatAnalyticsRepository) UpdateCombatAnalytics(id uuid.UUID, updates map[string]interface{}) error {
	// TODO: Implement database-agnostic update query builder
	// For now, this method is not implemented due to buildUpdateQuery using PostgreSQL syntax
	return fmt.Errorf("UpdateCombatAnalytics not implemented - pending SQL migration")
}

// Combatant Analytics methods

func (r *combatAnalyticsRepository) CreateCombatantAnalytics(analytics *models.CombatantAnalytics) error {
	query := `
		INSERT INTO combatant_analytics (
			id, combat_analytics_id, combatant_id, combatant_type, combatant_name,
			damage_dealt, damage_taken, healing_done, healing_received,
			attacks_made, attacks_hit, attacks_missed, critical_hits, critical_misses,
			saves_made, saves_failed, rounds_survived, final_hp,
			conditions_suffered, abilities_used, tactical_decisions
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if analytics.ID == uuid.Nil {
		analytics.ID = uuid.New()
	}
	analytics.CreatedAt = time.Now()

	query = r.db.Rebind(query)
	_, err := r.db.Exec(query,
		analytics.ID,
		analytics.CombatAnalyticsID,
		analytics.CombatantID,
		analytics.CombatantType,
		analytics.CombatantName,
		analytics.DamageDealt,
		analytics.DamageTaken,
		analytics.HealingDone,
		analytics.HealingReceived,
		analytics.AttacksMade,
		analytics.AttacksHit,
		analytics.AttacksMissed,
		analytics.CriticalHits,
		analytics.CriticalMisses,
		analytics.SavesMade,
		analytics.SavesFailed,
		analytics.RoundsSurvived,
		analytics.FinalHP,
		analytics.ConditionsSuffered,
		analytics.AbilitiesUsed,
		analytics.TacticalDecisions)
	return err
}

func (r *combatAnalyticsRepository) GetCombatantAnalytics(combatAnalyticsID uuid.UUID) ([]*models.CombatantAnalytics, error) {
	var analytics []*models.CombatantAnalytics
	query := `
		SELECT * FROM combatant_analytics 
		WHERE combat_analytics_id = ? 
		ORDER BY damage_dealt DESC`
	query = r.db.Rebind(query)
	err := r.db.Select(&analytics, query, combatAnalyticsID)
	return analytics, err
}

func (r *combatAnalyticsRepository) UpdateCombatantAnalytics(id uuid.UUID, updates map[string]interface{}) error {
	// TODO: Implement database-agnostic update query builder
	// For now, this method is not implemented due to buildUpdateQuery using PostgreSQL syntax
	return fmt.Errorf("UpdateCombatantAnalytics not implemented - pending SQL migration")
}

// Auto Combat Resolution methods

func (r *combatAnalyticsRepository) CreateAutoCombatResolution(resolution *models.AutoCombatResolution) error {
	query := `
		INSERT INTO auto_combat_resolutions (
			id, game_session_id, encounter_difficulty, party_composition,
			enemy_composition, resolution_type, outcome, rounds_simulated,
			party_resources_used, loot_generated, experience_awarded,
			narrative_summary
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if resolution.ID == uuid.Nil {
		resolution.ID = uuid.New()
	}
	resolution.CreatedAt = time.Now()

	query = r.db.Rebind(query)
	_, err := r.db.Exec(query,
		resolution.ID,
		resolution.GameSessionID,
		resolution.EncounterDifficulty,
		resolution.PartyComposition,
		resolution.EnemyComposition,
		resolution.ResolutionType,
		resolution.Outcome,
		resolution.RoundsSimulated,
		resolution.PartyResourcesUsed,
		resolution.LootGenerated,
		resolution.ExperienceAwarded,
		resolution.NarrativeSummary)
	return err
}

func (r *combatAnalyticsRepository) GetAutoCombatResolution(id uuid.UUID) (*models.AutoCombatResolution, error) {
	var resolution models.AutoCombatResolution
	query := `SELECT * FROM auto_combat_resolutions WHERE id = ?`
	query = r.db.Rebind(query)
	err := r.db.Get(&resolution, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("auto combat resolution not found")
	}
	return &resolution, err
}

func (r *combatAnalyticsRepository) GetAutoCombatResolutionsBySession(sessionID uuid.UUID) ([]*models.AutoCombatResolution, error) {
	var resolutions []*models.AutoCombatResolution
	query := `
		SELECT * FROM auto_combat_resolutions 
		WHERE game_session_id = ? 
		ORDER BY created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.Select(&resolutions, query, sessionID)
	return resolutions, err
}

// Battle Map methods

func (r *combatAnalyticsRepository) CreateBattleMap(battleMap *models.BattleMap) error {
	query := `
		INSERT INTO battle_maps (
			id, combat_id, game_session_id, location_description,
			map_type, grid_size_x, grid_size_y, terrain_features,
			obstacle_positions, cover_positions, hazard_zones,
			spawn_points, tactical_notes, visual_theme
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if battleMap.ID == uuid.Nil {
		battleMap.ID = uuid.New()
	}
	battleMap.CreatedAt = time.Now()
	battleMap.UpdatedAt = time.Now()

	query = r.db.Rebind(query)
	_, err := r.db.Exec(query,
		battleMap.ID,
		battleMap.CombatID,
		battleMap.GameSessionID,
		battleMap.LocationDescription,
		battleMap.MapType,
		battleMap.GridSizeX,
		battleMap.GridSizeY,
		battleMap.TerrainFeatures,
		battleMap.ObstaclePositions,
		battleMap.CoverPositions,
		battleMap.HazardZones,
		battleMap.SpawnPoints,
		battleMap.TacticalNotes,
		battleMap.VisualTheme)
	return err
}

func (r *combatAnalyticsRepository) GetBattleMap(id uuid.UUID) (*models.BattleMap, error) {
	var battleMap models.BattleMap
	query := `SELECT * FROM battle_maps WHERE id = ?`
	query = r.db.Rebind(query)
	err := r.db.Get(&battleMap, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("battle map not found")
	}
	return &battleMap, err
}

func (r *combatAnalyticsRepository) GetBattleMapByCombat(combatID uuid.UUID) (*models.BattleMap, error) {
	var battleMap models.BattleMap
	query := `SELECT * FROM battle_maps WHERE combat_id = ? ORDER BY created_at DESC LIMIT 1`
	query = r.db.Rebind(query)
	err := r.db.Get(&battleMap, query, combatID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &battleMap, err
}

func (r *combatAnalyticsRepository) GetBattleMapsBySession(sessionID uuid.UUID) ([]*models.BattleMap, error) {
	var maps []*models.BattleMap
	query := `
		SELECT * FROM battle_maps 
		WHERE game_session_id = ? 
		ORDER BY created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.Select(&maps, query, sessionID)
	return maps, err
}

func (r *combatAnalyticsRepository) UpdateBattleMap(id uuid.UUID, updates map[string]interface{}) error {
	// TODO: Implement database-agnostic update query builder
	// For now, this method is not implemented due to buildUpdateQuery using PostgreSQL syntax
	return fmt.Errorf("UpdateBattleMap not implemented - pending SQL migration")
}

// Smart Initiative methods

func (r *combatAnalyticsRepository) CreateOrUpdateInitiativeRule(rule *models.SmartInitiativeRule) error {
	// Check if rule exists
	var existingID string
	checkQuery := `SELECT id FROM smart_initiative_rules WHERE game_session_id = ? AND entity_id = ?`
	checkQuery = r.db.Rebind(checkQuery)
	err := r.db.QueryRow(checkQuery, rule.GameSessionID, rule.EntityID).Scan(&existingID)
	
	if err == sql.ErrNoRows {
		// Insert new rule
		if rule.ID == uuid.Nil {
			rule.ID = uuid.New()
		}
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()
		
		insertQuery := `
			INSERT INTO smart_initiative_rules (
				id, game_session_id, entity_id, entity_type,
				base_initiative_bonus, advantage_on_initiative,
				alert_feat, special_rules, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
		insertQuery = r.db.Rebind(insertQuery)
		_, err = r.db.Exec(insertQuery,
			rule.ID,
			rule.GameSessionID,
			rule.EntityID,
			rule.EntityType,
			rule.BaseInitiativeBonus,
			rule.AdvantageOnInitiative,
			rule.AlertFeat,
			rule.SpecialRules,
			rule.CreatedAt,
			rule.UpdatedAt)
		return err
	} else if err != nil {
		return err
	}
	
	// Update existing rule
	rule.UpdatedAt = time.Now()
	updateQuery := `
		UPDATE smart_initiative_rules SET
			entity_type = ?,
			base_initiative_bonus = ?,
			advantage_on_initiative = ?,
			alert_feat = ?,
			special_rules = ?,
			updated_at = ?
		WHERE game_session_id = ? AND entity_id = ?`
	
	updateQuery = r.db.Rebind(updateQuery)
	_, err = r.db.Exec(updateQuery,
		rule.EntityType,
		rule.BaseInitiativeBonus,
		rule.AdvantageOnInitiative,
		rule.AlertFeat,
		rule.SpecialRules,
		rule.UpdatedAt,
		rule.GameSessionID,
		rule.EntityID)
	return err
}

func (r *combatAnalyticsRepository) GetInitiativeRule(sessionID uuid.UUID, entityID string) (*models.SmartInitiativeRule, error) {
	var rule models.SmartInitiativeRule
	query := `SELECT * FROM smart_initiative_rules WHERE game_session_id = ? AND entity_id = ?`
	query = r.db.Rebind(query)
	err := r.db.Get(&rule, query, sessionID, entityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &rule, err
}

func (r *combatAnalyticsRepository) GetInitiativeRulesBySession(sessionID uuid.UUID) ([]*models.SmartInitiativeRule, error) {
	var rules []*models.SmartInitiativeRule
	query := `
		SELECT * FROM smart_initiative_rules 
		WHERE game_session_id = ? 
		ORDER BY entity_type, entity_id`
	query = r.db.Rebind(query)
	err := r.db.Select(&rules, query, sessionID)
	return rules, err
}

// Combat Action Log methods

func (r *combatAnalyticsRepository) CreateCombatAction(action *models.CombatActionLog) error {
	query := `
		INSERT INTO combat_action_log (
			id, combat_id, round_number, turn_number,
			actor_id, actor_type, action_type,
			target_id, target_type, roll_results,
			outcome, damage_dealt, conditions_applied,
			resources_used, position_data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if action.ID == uuid.Nil {
		action.ID = uuid.New()
	}
	action.Timestamp = time.Now()

	query = r.db.Rebind(query)
	_, err := r.db.Exec(query,
		action.ID,
		action.CombatID,
		action.RoundNumber,
		action.TurnNumber,
		action.ActorID,
		action.ActorType,
		action.ActionType,
		action.TargetID,
		action.TargetType,
		action.RollResults,
		action.Outcome,
		action.DamageDealt,
		action.ConditionsApplied,
		action.ResourcesUsed,
		action.PositionData)
	return err
}

func (r *combatAnalyticsRepository) GetCombatActions(combatID uuid.UUID) ([]*models.CombatActionLog, error) {
	var actions []*models.CombatActionLog
	query := `
		SELECT * FROM combat_action_log 
		WHERE combat_id = ? 
		ORDER BY round_number, turn_number`
	query = r.db.Rebind(query)
	err := r.db.Select(&actions, query, combatID)
	return actions, err
}

func (r *combatAnalyticsRepository) GetCombatActionsByRound(combatID uuid.UUID, roundNumber int) ([]*models.CombatActionLog, error) {
	var actions []*models.CombatActionLog
	query := `
		SELECT * FROM combat_action_log 
		WHERE combat_id = ? AND round_number = ?
		ORDER BY turn_number`
	query = r.db.Rebind(query)
	err := r.db.Select(&actions, query, combatID, roundNumber)
	return actions, err
}
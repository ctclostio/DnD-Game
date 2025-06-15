package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// RuleBuilderRepository handles database operations for rule builder
type RuleBuilderRepository struct {
	db *DB
}

// NewRuleBuilderRepository creates a new rule builder repository
func NewRuleBuilderRepository(db *DB) *RuleBuilderRepository {
	return &RuleBuilderRepository{db: db}
}

// scanActiveRule is a helper to scan a single ActiveRule row
func (r *RuleBuilderRepository) scanActiveRule(row RowScanner) (*models.ActiveRule, error) {
	var rule models.ActiveRule
	var compiledLogicJSON, parametersJSON []byte

	dest := []interface{}{
		&rule.ID,
		&rule.TemplateID,
		&rule.GameSessionID,
		&rule.CharacterID,
		&compiledLogicJSON,
		&parametersJSON,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	}

	jsonFields := map[int]JSONFieldUnmarshaler{
		4: UnmarshalJSONWithError(&rule.CompiledLogic, "compiled logic"),
		5: UnmarshalJSONWithError(&rule.Parameters, "parameters"),
	}

	if err := ScanWithJSON(row, dest, jsonFields); err != nil {
		return nil, err
	}

	return &rule, nil
}

// scanRuleExecution is a helper to scan a single RuleExecution row
func (r *RuleBuilderRepository) scanRuleExecution(row RowScanner) (*models.RuleExecution, error) {
	var execution models.RuleExecution
	var triggerContextJSON, executionResultJSON []byte

	dest := []interface{}{
		&execution.ID,
		&execution.RuleID,
		&execution.GameSessionID,
		&execution.CharacterID,
		&triggerContextJSON,
		&executionResultJSON,
		&execution.Success,
		&execution.ErrorMessage,
		&execution.ExecutedAt,
	}

	jsonFields := map[int]JSONFieldUnmarshaler{
		4: UnmarshalJSONWithError(&execution.TriggerContext, "trigger context"),
		5: UnmarshalJSONWithError(&execution.ExecutionResult, "execution result"),
	}

	if err := ScanWithJSON(row, dest, jsonFields); err != nil {
		return nil, err
	}

	return &execution, nil
}

// Rule Template Methods

// CreateRuleTemplate creates a new rule template
func (r *RuleBuilderRepository) CreateRuleTemplate(template *models.RuleTemplate) error {
	query := `
		INSERT INTO rule_templates (
			id, name, description, category, complexity,
			logic_graph, parameters, conditional_modifiers,
			tags, is_public, created_by, 
			average_rating, usage_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	logicGraphJSON, err := json.Marshal(template.LogicGraph)
	if err != nil {
		return fmt.Errorf("failed to marshal logic graph: %w", err)
	}

	parametersJSON, err := json.Marshal(template.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	condModsJSON, err := json.Marshal(template.ConditionalModifiers)
	if err != nil {
		return fmt.Errorf("failed to marshal conditional modifiers: %w", err)
	}

	tagsJSON, err := json.Marshal(template.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	_, err = r.db.ExecRebind(query,
		template.ID,
		template.Name,
		template.Description,
		template.Category,
		template.Complexity,
		logicGraphJSON,
		parametersJSON,
		condModsJSON,
		tagsJSON,
		template.IsPublic,
		template.CreatedByID,
		template.AverageRating,
		template.UsageCount,
		template.CreatedAt,
		template.UpdatedAt,
	)

	return err
}

// GetRuleTemplate gets a rule template by ID
func (r *RuleBuilderRepository) GetRuleTemplate(templateID string) (*models.RuleTemplate, error) {
	query := `
		SELECT id, name, description, category, complexity,
			   logic_graph, parameters, conditional_modifiers,
			   tags, is_public, created_by, 
			   average_rating, usage_count, created_at, updated_at
		FROM rule_templates
		WHERE id = ?
	`

	var template models.RuleTemplate
	var logicGraphJSON, parametersJSON, condModsJSON, tagsJSON []byte

	err := r.db.QueryRowRebind(query, templateID).Scan(
		&template.ID,
		&template.Name,
		&template.Description,
		&template.Category,
		&template.Complexity,
		&logicGraphJSON,
		&parametersJSON,
		&condModsJSON,
		&tagsJSON,
		&template.IsPublic,
		&template.CreatedByID,
		&template.AverageRating,
		&template.UsageCount,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(logicGraphJSON, &template.LogicGraph); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logic graph: %w", err)
	}

	if err := json.Unmarshal(parametersJSON, &template.Parameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	if err := json.Unmarshal(condModsJSON, &template.ConditionalModifiers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditional modifiers: %w", err)
	}

	if err := json.Unmarshal(tagsJSON, &template.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &template, nil
}

// GetRuleTemplates gets rule templates with filters
func (r *RuleBuilderRepository) GetRuleTemplates(userID, category string, isPublic bool) ([]models.RuleTemplate, error) {
	query := `
		SELECT id, name, description, category, complexity,
			   logic_graph, parameters, conditional_modifiers,
			   tags, is_public, created_by, 
			   average_rating, usage_count, created_at, updated_at
		FROM rule_templates
		WHERE 1=1
	`

	args := []interface{}{}

	if isPublic {
		query += " AND is_public = ?"
		args = append(args, true)
	} else if userID != "" {
		query += " AND (is_public = true OR created_by = ?)"
		args = append(args, userID)
	}

	if category != "" && category != "all" {
		query += " AND category = ?"
		args = append(args, category)
	}

	query += " ORDER BY usage_count DESC, average_rating DESC"

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	templates := make([]models.RuleTemplate, 0, 10)
	for rows.Next() {
		var template models.RuleTemplate
		var logicGraphJSON, parametersJSON, condModsJSON, tagsJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Description,
			&template.Category,
			&template.Complexity,
			&logicGraphJSON,
			&parametersJSON,
			&condModsJSON,
			&tagsJSON,
			&template.IsPublic,
			&template.CreatedByID,
			&template.AverageRating,
			&template.UsageCount,
			&template.CreatedAt,
			&template.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(logicGraphJSON, &template.LogicGraph); err != nil {
			return nil, fmt.Errorf("failed to unmarshal logic graph: %w", err)
		}

		if err := json.Unmarshal(parametersJSON, &template.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		if err := json.Unmarshal(condModsJSON, &template.ConditionalModifiers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditional modifiers: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &template.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return templates, nil
}

// UpdateRuleTemplate updates a rule template
func (r *RuleBuilderRepository) UpdateRuleTemplate(templateID string, updates map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE rule_templates SET updated_at = ?"
	args := []interface{}{time.Now()}

	for key, value := range updates {
		query += fmt.Sprintf(", %s = ?", key)

		// Handle JSON fields
		switch key {
		case "logic_graph", "parameters", "conditional_modifiers", "tags":
			jsonData, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal %s: %w", key, err)
			}
			args = append(args, jsonData)
		default:
			args = append(args, value)
		}
	}

	query += " WHERE id = ?"
	args = append(args, templateID)

	_, err := r.db.ExecRebind(query, args...)
	return err
}

// DeleteRuleTemplate deletes a rule template
func (r *RuleBuilderRepository) DeleteRuleTemplate(templateID string) error {
	_, err := r.db.ExecRebind("DELETE FROM rule_templates WHERE id = ?", templateID)
	return err
}

// IncrementUsageCount increments the usage count for a template
func (r *RuleBuilderRepository) IncrementUsageCount(templateID string) error {
	_, err := r.db.ExecRebind(
		"UPDATE rule_templates SET usage_count = usage_count + 1 WHERE id = ?",
		templateID,
	)
	return err
}

// Node Template Methods

// GetNodeTemplates gets all available node templates
func (r *RuleBuilderRepository) GetNodeTemplates() ([]models.NodeTemplate, error) {
	query := `
		SELECT id, node_type, subtype, category, name, description,
			   icon, color, input_ports, output_ports, default_properties
		FROM node_templates
		ORDER BY category, name
	`

	rows, err := r.db.QueryContext(context.Background(), r.db.Rebind(query))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	templates := make([]models.NodeTemplate, 0, 10)
	for rows.Next() {
		var template models.NodeTemplate
		var inputPortsJSON, outputPortsJSON, defaultPropsJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.NodeType,
			&template.Subtype,
			&template.Category,
			&template.Name,
			&template.Description,
			&template.Icon,
			&template.Color,
			&inputPortsJSON,
			&outputPortsJSON,
			&defaultPropsJSON,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(inputPortsJSON, &template.InputPorts); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input ports: %w", err)
		}

		if err := json.Unmarshal(outputPortsJSON, &template.OutputPorts); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output ports: %w", err)
		}

		if err := json.Unmarshal(defaultPropsJSON, &template.DefaultProperties); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default properties: %w", err)
		}

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return templates, nil
}

// Active Rule Methods

// CreateActiveRule creates a new active rule instance
func (r *RuleBuilderRepository) CreateActiveRule(rule *models.ActiveRule) error {
	query := `
		INSERT INTO active_rules (
			id, template_id, game_session_id, character_id,
			compiled_logic, parameters, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	compiledLogicJSON, err := MarshalJSONField(rule.CompiledLogic, "compiled logic")
	if err != nil {
		return err
	}

	parametersJSON, err := MarshalJSONField(rule.Parameters, "parameters")
	if err != nil {
		return err
	}

	_, err = r.db.ExecRebind(query,
		rule.ID,
		rule.TemplateID,
		rule.GameSessionID,
		rule.CharacterID,
		compiledLogicJSON,
		parametersJSON,
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	return err
}

// GetActiveRules gets active rules for a session or character
func (r *RuleBuilderRepository) GetActiveRules(gameSessionID, characterID string) ([]models.ActiveRule, error) {
	query := `
		SELECT id, template_id, game_session_id, character_id,
			   compiled_logic, parameters, is_active,
			   created_at, updated_at
		FROM active_rules
		WHERE is_active = true
	`

	args := []interface{}{}

	if gameSessionID != "" {
		query += constants.AndGameSessionIDClause
		args = append(args, gameSessionID)
	}

	if characterID != "" {
		query += constants.AndCharacterIDClause
		args = append(args, characterID)
	}

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	rulesPtr, err := ScanRowsGeneric(rows, r.scanActiveRule)
	if err != nil {
		return nil, err
	}

	// Convert []*models.ActiveRule to []models.ActiveRule
	rules := make([]models.ActiveRule, len(rulesPtr))
	for i, r := range rulesPtr {
		rules[i] = *r
	}
	return rules, nil
}

// DeactivateRule deactivates an active rule
func (r *RuleBuilderRepository) DeactivateRule(ruleID string) error {
	_, err := r.db.ExecRebind(
		"UPDATE active_rules SET is_active = false, updated_at = ? WHERE id = ?",
		time.Now(),
		ruleID,
	)
	return err
}

// Execution History Methods

// CreateRuleExecution creates a new rule execution record
func (r *RuleBuilderRepository) CreateRuleExecution(execution *models.RuleExecution) error {
	query := `
		INSERT INTO rule_executions (
			id, rule_id, game_session_id, character_id,
			trigger_context, execution_result, success,
			error_message, executed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	triggerContextJSON, err := MarshalJSONField(execution.TriggerContext, "trigger context")
	if err != nil {
		return err
	}

	executionResultJSON, err := MarshalJSONField(execution.ExecutionResult, "execution result")
	if err != nil {
		return err
	}

	_, err = r.db.ExecRebind(query,
		execution.ID,
		execution.RuleID,
		execution.GameSessionID,
		execution.CharacterID,
		triggerContextJSON,
		executionResultJSON,
		execution.Success,
		execution.ErrorMessage,
		execution.ExecutedAt,
	)

	return err
}

// GetRuleExecutionHistory gets rule execution history
func (r *RuleBuilderRepository) GetRuleExecutionHistory(gameSessionID, characterID string, limit int) ([]models.RuleExecution, error) {
	query := `
		SELECT id, rule_id, game_session_id, character_id,
			   trigger_context, execution_result, success,
			   error_message, executed_at
		FROM rule_executions
		WHERE 1=1
	`

	args := []interface{}{}

	if gameSessionID != "" {
		query += constants.AndGameSessionIDClause
		args = append(args, gameSessionID)
	}

	if characterID != "" {
		query += constants.AndCharacterIDClause
		args = append(args, characterID)
	}

	query += " ORDER BY executed_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	executionsPtr, err := ScanRowsGeneric(rows, r.scanRuleExecution)
	if err != nil {
		return nil, err
	}

	// Convert []*models.RuleExecution to []models.RuleExecution
	executions := make([]models.RuleExecution, len(executionsPtr))
	for i, e := range executionsPtr {
		executions[i] = *e
	}

	return executions, nil
}

// Conditional Modifier Methods

// GetConditionalModifiers gets conditional modifiers for a rule
func (r *RuleBuilderRepository) GetConditionalModifiers(ruleID string) ([]models.ConditionalModifier, error) {
	query := `
		SELECT cm.*
		FROM conditional_modifiers cm
		JOIN rule_templates rt ON rt.id = ?
		WHERE cm.id = ANY(rt.conditional_modifiers)
		ORDER BY cm.priority DESC
	`

	rows, err := r.db.QueryContext(context.Background(), r.db.Rebind(query), ruleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	modifiers := make([]models.ConditionalModifier, 0, 10)
	for rows.Next() {
		var modifier models.ConditionalModifier
		var modifiersJSON []byte

		err := rows.Scan(
			&modifier.ID,
			&modifier.Name,
			&modifier.ContextType,
			&modifier.ContextValue,
			&modifiersJSON,
			&modifier.Priority,
			&modifier.Description,
		)

		if err != nil {
			return nil, err
		}

		// Unmarshal modifiers
		if err := json.Unmarshal(modifiersJSON, &modifier.Modifiers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal modifiers: %w", err)
		}

		modifiers = append(modifiers, modifier)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return modifiers, nil
}

// GetRuleInstance retrieves a rule instance by ID
func (r *RuleBuilderRepository) GetRuleInstance(instanceID string) (*models.RuleInstance, error) {
	query := `
		SELECT id, template_id, game_session_id, owner_id,
			   parameter_values, is_active, created_at, updated_at
		FROM rule_instances
		WHERE id = ?
	`

	var instance models.RuleInstance
	var parameterValuesJSON []byte

	err := r.db.QueryRowRebind(query, instanceID).Scan(
		&instance.ID,
		&instance.TemplateID,
		&instance.SessionID,
		&instance.OwnerID,
		&parameterValuesJSON,
		&instance.IsActive,
		&instance.CreatedAt,
		&instance.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal parameter values
	if err := json.Unmarshal(parameterValuesJSON, &instance.ParameterValues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameter values: %w", err)
	}

	return &instance, nil
}

// DeactivateRuleInstance deactivates a rule instance
func (r *RuleBuilderRepository) DeactivateRuleInstance(instanceID string) error {
	query := `
		UPDATE rule_instances
		SET is_active = false, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecRebind(query, time.Now(), instanceID)
	return err
}

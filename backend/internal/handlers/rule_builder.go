package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
	"github.com/gorilla/mux"
)

// Rule Template Handlers.
// GetRuleTemplates handles GET /api/rules/templates.
func (h *Handlers) GetRuleTemplates(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r.Context())

	// Get query parameters.
	category := r.URL.Query().Get("category")
	isPublic := r.URL.Query().Get("public") == "true"

	templates, err := h.ruleEngine.GetRuleTemplates(userID, category, isPublic)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, templates)
}

// GetRuleTemplate handles GET /api/rules/templates/{id}.
func (h *Handlers) GetRuleTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := mux.Vars(r)["id"]

	template, err := h.ruleEngine.GetRuleTemplate(templateID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	response.JSON(w, r, http.StatusOK, template)
}

// CreateRuleTemplate handles POST /api/rules/templates.
func (h *Handlers) CreateRuleTemplate(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r.Context())

	var template models.RuleTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	template.CreatedByID = userID

	createdTemplate, err := h.ruleEngine.CreateRuleTemplate(&template)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, createdTemplate)
}

// UpdateRuleTemplate handles PUT /api/rules/templates/{id}.
func (h *Handlers) UpdateRuleTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := mux.Vars(r)["id"]
	userID, _ := auth.GetUserIDFromContext(r.Context())

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Verify ownership.
	template, err := h.ruleEngine.GetRuleTemplate(templateID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	if template.CreatedByID != userID {
		response.Forbidden(w, r, "You can only update your own rule templates")
		return
	}

	updatedTemplate, err := h.ruleEngine.UpdateRuleTemplate(templateID, updates)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, updatedTemplate)
}

// DeleteRuleTemplate handles DELETE /api/rules/templates/{id}.
func (h *Handlers) DeleteRuleTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := mux.Vars(r)["id"]
	userID, _ := auth.GetUserIDFromContext(r.Context())

	// Verify ownership.
	template, err := h.ruleEngine.GetRuleTemplate(templateID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	if template.CreatedByID != userID {
		response.Forbidden(w, r, "You can only delete your own rule templates")
		return
	}

	if err := h.ruleEngine.DeleteRuleTemplate(templateID); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusNoContent, nil)
}

// CompileRuleTemplate handles POST /api/rules/templates/{id}/compile.
func (h *Handlers) CompileRuleTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := mux.Vars(r)["id"]

	var compileRequest struct {
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&compileRequest); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Get the template first.
	template, err := h.ruleEngine.GetRuleTemplate(templateID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	// Compile the template.
	compiled, err := h.ruleEngine.CompileRule(template)
	if err != nil {
		response.BadRequest(w, r, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, compiled)
}

// ValidateRuleTemplate handles POST /api/rules/templates/{id}/validate.
func (h *Handlers) ValidateRuleTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := mux.Vars(r)["id"]

	var validateRequest struct {
		TestScenario map[string]interface{} `json:"test_scenario"`
	}

	if err := json.NewDecoder(r.Body).Decode(&validateRequest); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	template, err := h.ruleEngine.GetRuleTemplate(templateID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	// Validate the rule by attempting to compile it.
	compiled, err := h.ruleEngine.CompileRule(template)
	validationResult := &services.ValidationResult{
		IsValid: err == nil,
		Errors:  []string{},
	}
	if err != nil {
		validationResult.Errors = append(validationResult.Errors, err.Error())
	}

	// If valid, try to execute with test scenario.
	var executionResult interface{}
	if validationResult.IsValid && validateRequest.TestScenario != nil {
		// Create a test instance for validation.
		testInstance := &models.RuleInstance{
			ID:              "test-" + templateID,
			TemplateID:      templateID,
			ParameterValues: validateRequest.TestScenario,
		}

		// Create test trigger data.
		testTrigger := services.TriggerData{
			Type:       "test",
			Properties: validateRequest.TestScenario,
		}

		executionResult, err = h.ruleEngine.ExecuteRule(r.Context(), compiled, testInstance, testTrigger)
		if err != nil {
			fmt.Printf("rule execution error: %v\n", err)
		}
	}

	resp := map[string]interface{}{
		"is_valid":         validationResult.IsValid,
		"errors":           validationResult.Errors,
		"warnings":         validationResult.Warnings,
		"execution_result": executionResult,
	}

	response.JSON(w, r, http.StatusOK, resp)
}

// AnalyzeRuleBalance handles POST /api/rules/templates/{id}/analyze.
func (h *Handlers) AnalyzeRuleBalance(w http.ResponseWriter, r *http.Request) {
	templateID := mux.Vars(r)["id"]

	var analysisRequest struct {
		SimulationCount int               `json:"simulation_count"`
		LevelRange      models.LevelRange `json:"level_range"`
		Scenarios       []string          `json:"scenarios"`
	}

	if err := json.NewDecoder(r.Body).Decode(&analysisRequest); err != nil {
		// Use defaults.
		analysisRequest.SimulationCount = 1000
		analysisRequest.LevelRange = models.LevelRange{Min: 1, Max: 20}
		analysisRequest.Scenarios = []string{"pvp", "pve", "exploration", "roleplay"}
	}

	template, err := h.ruleEngine.GetRuleTemplate(templateID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	// Call the balance analyzer with the template.
	analysis, err := h.balanceAnalyzer.AnalyzeRuleBalance(r.Context(), template)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, analysis)
}

// GetNodeTemplates handles GET /api/rules/nodes/templates.
func (h *Handlers) GetNodeTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.ruleEngine.GetNodeTemplates()
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, templates)
}

// Rule Instance Handlers.
// GetActiveRules handles GET /api/rules/active.
func (h *Handlers) GetActiveRules(w http.ResponseWriter, r *http.Request) {
	gameSessionID := r.URL.Query().Get("game_session_id")
	characterID := r.URL.Query().Get("character_id")

	if gameSessionID == "" && characterID == "" {
		response.BadRequest(w, r, "Either game_session_id or character_id is required")
		return
	}

	rules, err := h.ruleEngine.GetActiveRules(gameSessionID, characterID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, rules)
}

// ActivateRule handles POST /api/rules/activate.
func (h *Handlers) ActivateRule(w http.ResponseWriter, r *http.Request) {
	var activateRequest struct {
		TemplateID    string                 `json:"template_id"`
		GameSessionID string                 `json:"game_session_id"`
		CharacterID   string                 `json:"character_id"`
		Parameters    map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&activateRequest); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	activeRule, err := h.ruleEngine.ActivateRule(
		activateRequest.TemplateID,
		activateRequest.GameSessionID,
		activateRequest.CharacterID,
		activateRequest.Parameters,
	)

	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, activeRule)
}

// DeactivateRule handles DELETE /api/rules/active/{id}.
func (h *Handlers) DeactivateRule(w http.ResponseWriter, r *http.Request) {
	ruleID := mux.Vars(r)["id"]

	if err := h.ruleEngine.DeactivateRule(ruleID); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusNoContent, nil)
}

// ExecuteRule handles POST /api/rules/execute.
func (h *Handlers) ExecuteRule(w http.ResponseWriter, r *http.Request) {
	var executeRequest struct {
		RuleID  string                 `json:"rule_id"`
		Context map[string]interface{} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&executeRequest); err != nil {
		response.BadRequest(w, r, "Invalid request body")
		return
	}

	// Get the rule instance and template.
	ruleInstance, err := h.ruleEngine.GetRuleInstance(executeRequest.RuleID)
	if err != nil {
		response.NotFound(w, r, "Rule instance not found")
		return
	}

	template, err := h.ruleEngine.GetRuleTemplate(ruleInstance.TemplateID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	// Compile the rule.
	compiled, err := h.ruleEngine.CompileRule(template)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Create trigger data from context.
	trigger := services.TriggerData{
		Type:       "manual",
		Properties: executeRequest.Context,
	}

	// Execute the rule.
	result, err := h.ruleEngine.ExecuteRule(r.Context(), compiled, ruleInstance, trigger)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, result)
}

// GetConditionalModifiers handles GET /api/rules/conditional/{id}.
func (h *Handlers) GetConditionalModifiers(w http.ResponseWriter, r *http.Request) {
	ruleID := mux.Vars(r)["id"]

	var context struct {
		Plane     string `json:"plane"`
		Weather   string `json:"weather"`
		Emotion   string `json:"emotion"`
		Terrain   string `json:"terrain"`
		Time      string `json:"time"`
		MoonPhase string `json:"moon_phase"`
	}

	// Get context from query parameters.
	context.Plane = r.URL.Query().Get("plane")
	context.Weather = r.URL.Query().Get("weather")
	context.Emotion = r.URL.Query().Get("emotion")
	context.Terrain = r.URL.Query().Get("terrain")
	context.Time = r.URL.Query().Get("time")
	context.MoonPhase = r.URL.Query().Get("moon_phase")

	// Get the rule template.
	template, err := h.ruleEngine.GetRuleTemplate(ruleID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	// Create contexts from the provided parameters.
	contexts := []models.ConditionalContext{}
	if context.Plane != "" {
		contexts = append(contexts, models.ConditionalContext{
			ContextType:  "plane",
			ContextValue: map[string]interface{}{"value": context.Plane},
			IsActive:     true,
		})
	}
	if context.Weather != "" {
		contexts = append(contexts, models.ConditionalContext{
			ContextType:  "weather",
			ContextValue: map[string]interface{}{"value": context.Weather},
			IsActive:     true,
		})
	}
	// Add other contexts similarly...

	// Apply conditional rules to see what modifications would be made.
	testInstance := &models.RuleInstance{
		ID:         "test-" + ruleID,
		TemplateID: ruleID,
	}

	modifiedTemplate, err := h.conditionalReality.ApplyConditionalRules(template, testInstance, contexts)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Return the differences between original and modified template.
	modifiers := map[string]interface{}{
		"original":    template,
		"modified":    modifiedTemplate,
		"contexts":    contexts,
		"description": "Conditional modifications based on provided context",
	}

	response.JSON(w, r, http.StatusOK, modifiers)
}

// ExportRuleTemplate handles GET /api/rules/templates/{id}/export.
func (h *Handlers) ExportRuleTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := mux.Vars(r)["id"]
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	template, err := h.ruleEngine.GetRuleTemplate(templateID)
	if err != nil {
		response.NotFound(w, r, "Rule template not found")
		return
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename="+template.Name+".json")
		if err := json.NewEncoder(w).Encode(template); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	default:
		response.BadRequest(w, r, "Unsupported export format")
	}
}

// ImportRuleTemplate handles POST /api/rules/templates/import.
func (h *Handlers) ImportRuleTemplate(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromContext(r.Context())

	var template models.RuleTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		response.BadRequest(w, r, "Invalid rule template format")
		return
	}

	// Clear ID and set creator.
	template.ID = ""
	template.CreatedByID = userID

	imported, err := h.ruleEngine.CreateRuleTemplate(&template)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusCreated, imported)
}

// GetRuleHistory handles GET /api/rules/history.
func (h *Handlers) GetRuleHistory(w http.ResponseWriter, r *http.Request) {
	gameSessionID := r.URL.Query().Get("game_session_id")
	characterID := r.URL.Query().Get("character_id")
	limit := 50

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	history, err := h.ruleEngine.GetRuleExecutionHistory(gameSessionID, characterID, limit)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, r, http.StatusOK, history)
}

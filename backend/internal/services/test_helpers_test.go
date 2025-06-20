package services_test

import (
	"context"
	"errors"
	"math/rand"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Test constants
const (
	testErrCombatNotFound = "combat not found"
)

// Helper functions for StartCombat
func validateCombatInput(sessionID string, participants []models.Combatant) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	if len(participants) < 2 {
		return errors.New("at least two combatants are required")
	}

	// Validate participants
	for _, p := range participants {
		if p.ID == "" {
			return errors.New("combatant ID is required")
		}
		if p.Name == "" {
			return errors.New("combatant name is required")
		}
		if p.HP <= 0 {
			return errors.New("combatant must have positive HP")
		}
	}
	return nil
}

func rollInitiativeForCombatants(participants []models.Combatant) {
	for i := range participants {
		participants[i].Initiative = rand.Intn(20) + 1 + participants[i].InitiativeRoll
		if participants[i].Initiative == 0 {
			participants[i].Initiative = rand.Intn(20) + 1
		}
	}
}

func createTurnOrder(participants []models.Combatant) []string {
	turnOrder := make([]string, len(participants))
	for i, p := range participants {
		turnOrder[i] = p.ID
	}

	// Sort turn order by initiative
	for i := 0; i < len(turnOrder)-1; i++ {
		for j := i + 1; j < len(turnOrder); j++ {
			init1 := getInitiativeForCombatant(participants, turnOrder[i])
			init2 := getInitiativeForCombatant(participants, turnOrder[j])
			if init2 > init1 {
				turnOrder[i], turnOrder[j] = turnOrder[j], turnOrder[i]
			}
		}
	}
	return turnOrder
}

func getInitiativeForCombatant(participants []models.Combatant, combatantID string) int {
	for _, p := range participants {
		if p.ID == combatantID {
			return p.Initiative
		}
	}
	return 0
}

// Helper functions for ApplyDamage
func calculateFinalDamage(target *models.Combatant, damage int, damageType string) int {
	// Check immunities first
	for _, immunity := range target.DamageImmunities {
		if immunity == damageType {
			return 0
		}
	}

	finalDamage := damage

	// Check resistances
	for _, resistance := range target.DamageResistances {
		if resistance == damageType {
			finalDamage = damage / 2
			break
		}
	}

	// Check vulnerabilities
	for _, vulnerability := range target.DamageVulnerabilities {
		if vulnerability == damageType {
			finalDamage = damage * 2
			break
		}
	}

	return finalDamage
}

func applyDamageToTarget(target *models.Combatant, damage int) {
	target.HP -= damage
	if target.HP < 0 {
		target.HP = 0
	}
}

func updateUnconsciousCondition(target *models.Combatant) {
	if target.HP == 0 {
		hasUnconscious := false
		for _, cond := range target.Conditions {
			if cond == models.ConditionUnconscious {
				hasUnconscious = true
				break
			}
		}
		if !hasUnconscious {
			target.Conditions = append(target.Conditions, models.ConditionUnconscious)
		}
	}
}

// Helper methods for ExecuteAction
func (s *CombatService) processAttackAction(testCombat *TestCombat, action *models.CombatAction) {
	// Simulate attack (hit/miss)
	action.Hit = rand.Intn(20)+1+action.AttackBonus >= 10 // Simple hit calculation
	if action.Hit && action.TargetID != "" {
		// Apply damage if hit
		if target, ok := testCombat.CombatantsMap[action.TargetID]; ok {
			damage := rand.Intn(8) + 1 + action.DamageBonus
			target.HP -= damage
			if target.HP < 0 {
				target.HP = 0
			}
		}
	}
}

func (s *CombatService) processMoveAction(testCombat *TestCombat, action *models.CombatAction) {
	if actor, ok := testCombat.CombatantsMap[action.ActorID]; ok {
		actor.Position = action.NewPosition
	}
}

func (s *CombatService) processDodgeAction(testCombat *TestCombat, action *models.CombatAction) {
	if actor, ok := testCombat.CombatantsMap[action.ActorID]; ok {
		if actor.Conditions == nil {
			actor.Conditions = []models.Condition{}
		}
		actor.Conditions = append(actor.Conditions, models.ConditionDodging)
	}
}

func (s *CombatService) processEndTurnAction(combat *models.Combat) {
	combat.Turn++
	combat.CurrentTurn = combat.Turn
	if combat.Turn >= len(combat.TurnOrder) {
		combat.Turn = 0
		combat.CurrentTurn = 0
		combat.Round++
	}
}

// CombatService adapter for tests
type CombatService struct {
	combats  map[string]*TestCombat
	charRepo interface{}
	diceRepo interface{}
	llm      interface{}
}

// TestCombat wraps models.Combat with test-specific fields
type TestCombat struct {
	*models.Combat
	CombatantsMap map[string]*models.Combatant
}

// NewCombatService creates a new combat service for tests
func NewCombatService(charRepo, diceRepo, llm interface{}) *CombatService {
	return &CombatService{
		combats:  make(map[string]*TestCombat),
		charRepo: charRepo,
		diceRepo: diceRepo,
		llm:      llm,
	}
}

// StartCombat initializes a new combat
func (s *CombatService) StartCombat(ctx context.Context, sessionID string, participants []models.Combatant) (*models.Combat, error) {
	if err := validateCombatInput(sessionID, participants); err != nil {
		return nil, err
	}

	// Roll initiative for each combatant
	rollInitiativeForCombatants(participants)

	// Sort by initiative (descending)
	turnOrder := createTurnOrder(participants)

	combat := &models.Combat{
		ID:            uuid.New().String(),
		GameSessionID: sessionID,
		SessionID:     sessionID, // Alias
		Round:         1,
		CurrentTurn:   0,
		Turn:          0, // Alias
		Combatants:    participants,
		TurnOrder:     turnOrder,
		IsActive:      true,
		Status:        models.CombatStatusActive,
		ActionHistory: []models.CombatAction{},
	}

	// Create combatants map
	combatantsMap := make(map[string]*models.Combatant)
	for i := range combat.Combatants {
		combatantsMap[combat.Combatants[i].ID] = &combat.Combatants[i]
	}

	testCombat := &TestCombat{
		Combat:        combat,
		CombatantsMap: combatantsMap,
	}

	s.combats[combat.ID] = testCombat
	return combat, nil
}

// GetCombatState retrieves the current state of a combat
func (s *CombatService) GetCombatState(_ context.Context, combatID string) (*models.Combat, error) {
	if combatID == "" {
		return nil, errors.New("combat ID is required")
	}

	testCombat, exists := s.combats[combatID]
	if !exists {
		return nil, errors.New(testErrCombatNotFound)
	}

	// Update combatants from map to slice
	combatants := make([]models.Combatant, 0, len(testCombat.CombatantsMap))
	for _, c := range testCombat.CombatantsMap {
		combatants = append(combatants, *c)
	}
	testCombat.Combatants = combatants

	return testCombat.Combat, nil
}

// ExecuteAction processes a combat action
func (s *CombatService) ExecuteAction(ctx context.Context, combatID string, action *models.CombatAction) (*models.Combat, error) {
	testCombat, err := s.validateAction(combatID, action)
	if err != nil {
		return nil, err
	}

	// Process action based on type
	if err := s.processAction(testCombat, action); err != nil {
		return nil, err
	}

	return s.GetCombatState(ctx, combatID)
}

// validateAction validates the combat and action parameters
func (s *CombatService) validateAction(combatID string, action *models.CombatAction) (*TestCombat, error) {
	if combatID == "" {
		return nil, errors.New(testErrCombatNotFound)
	}

	if action == nil {
		return nil, errors.New("action is required")
	}

	testCombat, exists := s.combats[combatID]
	if !exists {
		return nil, errors.New(testErrCombatNotFound)
	}

	combat := testCombat.Combat
	if !combat.IsActive || combat.Status == models.CombatStatusCompleted {
		return nil, errors.New("combat has already ended")
	}

	// Check if it's the actor's turn
	currentTurnID := combat.TurnOrder[combat.Turn]
	if action.ActorID != currentTurnID && action.ActionType != models.ActionTypeReaction {
		return nil, errors.New("not this combatant's turn")
	}

	return testCombat, nil
}

// processAction handles the specific action type
func (s *CombatService) processAction(testCombat *TestCombat, action *models.CombatAction) error {
	switch action.ActionType {
	case models.ActionTypeAttack:
		s.processAttackAction(testCombat, action)
	case models.ActionTypeCastSpell, models.ActionTypeCast:
		// Record spell cast
	case models.ActionTypeMove:
		s.processMoveAction(testCombat, action)
	case models.ActionTypeDodge:
		s.processDodgeAction(testCombat, action)
	case models.ActionTypeEndTurn:
		s.processEndTurnAction(testCombat.Combat)
	case models.ActionTypeDash:
		// Double movement speed for this turn

	case models.ActionTypeHelp:
		// Grant advantage to an ally

	case models.ActionTypeHide:
		// Attempt to hide

	case models.ActionTypeReady:
		// Ready an action for a trigger

	case models.ActionTypeSearch:
		// Search the area

	case models.ActionTypeUseItem:
		// Use an item from inventory

	case models.ActionTypeBonusAction:
		// Bonus action taken

	case models.ActionTypeReaction:
		// Reaction triggered

	case models.ActionTypeDeathSave:
		// Death saving throw

	case models.ActionTypeConcentration:
		// Concentration check

	case models.ActionTypeSavingThrow:
		// Saving throw made
	}

	// Record the action with all modifications
	testCombat.Combat.ActionHistory = append(testCombat.Combat.ActionHistory, *action)

	return nil
}

// EndCombat ends an active combat
func (s *CombatService) EndCombat(_ context.Context, combatID string) error {
	testCombat, exists := s.combats[combatID]
	if !exists {
		return errors.New(testErrCombatNotFound)
	}

	if !testCombat.IsActive || testCombat.Status == models.CombatStatusCompleted {
		return errors.New("combat has already ended")
	}

	testCombat.IsActive = false
	testCombat.Status = models.CombatStatusCompleted
	return nil
}

// SetCombatState sets the combat state (for testing)
func (s *CombatService) SetCombatState(combat *models.Combat) {
	combatantsMap := make(map[string]*models.Combatant)
	for i := range combat.Combatants {
		combatantsMap[combat.Combatants[i].ID] = &combat.Combatants[i]
	}

	// Set aliases
	if combat.GameSessionID != "" {
		combat.SessionID = combat.GameSessionID
	}
	combat.Turn = combat.CurrentTurn

	// Sync IsActive with Status if needed
	switch combat.Status {
	case models.CombatStatusActive:
		combat.IsActive = true
	case models.CombatStatusCompleted:
		combat.IsActive = false
	case models.CombatStatusPaused:
		combat.IsActive = false // Paused combat is not active
	}

	s.combats[combat.ID] = &TestCombat{
		Combat:        combat,
		CombatantsMap: combatantsMap,
	}
}

// ApplyDamage applies damage to a combatant
func (s *CombatService) ApplyDamage(ctx context.Context, combatID, targetID string, damage int, damageType string) (*models.Combat, error) {
	testCombat, exists := s.combats[combatID]
	if !exists {
		return nil, errors.New(testErrCombatNotFound)
	}

	target, ok := testCombat.CombatantsMap[targetID]
	if !ok {
		return nil, errors.New("target not found in combat")
	}

	// Calculate final damage based on resistances/immunities/vulnerabilities
	finalDamage := calculateFinalDamage(target, damage, damageType)

	// Apply damage and update HP
	applyDamageToTarget(target, finalDamage)

	// Handle unconscious condition
	updateUnconsciousCondition(target)

	return s.GetCombatState(ctx, combatID)
}

// ApplyHealing applies healing to a combatant
func (s *CombatService) ApplyHealing(ctx context.Context, combatID, targetID string, healing int) (*models.Combat, error) {
	if healing <= 0 {
		return nil, errors.New("healing must be positive")
	}

	testCombat, exists := s.combats[combatID]
	if !exists {
		return nil, errors.New(testErrCombatNotFound)
	}

	target, ok := testCombat.CombatantsMap[targetID]
	if !ok {
		return nil, errors.New("target not found in combat")
	}

	// Apply healing
	target.HP += healing
	if target.HP > target.MaxHP {
		target.HP = target.MaxHP
	}

	// Remove unconscious condition if healed
	if target.HP > 0 && target.Conditions != nil {
		newConditions := []models.Condition{}
		for _, cond := range target.Conditions {
			if cond != models.ConditionUnconscious {
				newConditions = append(newConditions, cond)
			}
		}
		target.Conditions = newConditions
	}

	return s.GetCombatState(ctx, combatID)
}

// DeathSavingThrow processes a death saving throw
func (s *CombatService) DeathSavingThrow(ctx context.Context, combatID, characterID string) (*models.Combat, *models.DeathSaveResult, error) {
	testCombat, exists := s.combats[combatID]
	if !exists {
		return nil, nil, errors.New(testErrCombatNotFound)
	}

	character, ok := testCombat.CombatantsMap[characterID]
	if !ok {
		return nil, nil, errors.New("character not found in combat")
	}

	if character.HP > 0 {
		return nil, nil, errors.New("character is not unconscious")
	}

	// Roll d20
	roll := rand.Intn(20) + 1
	result := &models.DeathSaveResult{
		Roll: roll,
	}

	// Critical success (20) - regain 1 HP
	if roll == 20 {
		handleCriticalSuccess(character, result)
		combat, _ := s.GetCombatState(ctx, combatID)
		return combat, result, nil
	}

	// Critical failure - 2 failures
	if roll == 1 {
		handleCriticalFailure(character, result)
		combat, _ := s.GetCombatState(ctx, combatID)
		return combat, result, nil
	}

	// Normal success
	if roll >= 10 {
		handleSuccess(character, result)
	} else {
		// Failure
		handleFailure(character, result)
	}

	combat, _ := s.GetCombatState(ctx, combatID)
	return combat, result, nil
}

// handleCriticalSuccess processes a critical success on a death save
func handleCriticalSuccess(combatant *models.Combatant, result *models.DeathSaveResult) {
	result.CritSuccess = true
	result.Success = true
	combatant.HP = 1

	// Remove unconscious condition
	removeUnconsciousCondition(combatant)

	combatant.DeathSaveSuccesses = 0
	combatant.DeathSaveFailures = 0
}

// handleCriticalFailure processes a critical failure on a death save
func handleCriticalFailure(combatant *models.Combatant, result *models.DeathSaveResult) {
	result.CritFailure = true
	combatant.DeathSaveFailures += 2

	if combatant.DeathSaveFailures >= 3 {
		combatant.DeathSaveFailures = 3
		combatant.Conditions = append(combatant.Conditions, models.ConditionDead)
	}
}

// handleSuccess processes a normal success on a death save
func handleSuccess(combatant *models.Combatant, result *models.DeathSaveResult) {
	result.Success = true
	combatant.DeathSaveSuccesses++

	if combatant.DeathSaveSuccesses >= 3 {
		// Stabilized
		combatant.DeathSaveSuccesses = 0
		combatant.DeathSaveFailures = 0
		combatant.Conditions = append(combatant.Conditions, models.ConditionStable)
	}
}

// handleFailure processes a failure on a death save
func handleFailure(combatant *models.Combatant, _ *models.DeathSaveResult) {
	combatant.DeathSaveFailures++

	if combatant.DeathSaveFailures >= 3 {
		combatant.DeathSaveFailures = 3
		combatant.Conditions = append(combatant.Conditions, models.ConditionDead)
	}
}

// removeUnconsciousCondition removes the unconscious condition from a combatant
func removeUnconsciousCondition(combatant *models.Combatant) {
	if combatant.Conditions == nil {
		return
	}

	newConditions := []models.Condition{}
	for _, cond := range combatant.Conditions {
		if cond != models.ConditionUnconscious {
			newConditions = append(newConditions, cond)
		}
	}
	combatant.Conditions = newConditions
}

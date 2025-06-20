package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/game"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// Error messages
const (
	errCombatNotFound    = "combat not found"
	errActorNotFound     = "actor not found"
	errTargetNotFound    = "target not found"
	errCombatantNotFound = "combatant not found"
)

type CombatService struct {
	engine  *game.CombatEngine
	combats map[string]*models.Combat // In-memory storage for active combats
}

func NewCombatService() *CombatService {
	return &CombatService{
		engine:  game.NewCombatEngine(),
		combats: make(map[string]*models.Combat),
	}
}

func (s *CombatService) StartCombat(_ context.Context, gameSessionID string, combatants []models.Combatant) (*models.Combat, error) {
	combat, err := s.engine.StartCombat(gameSessionID, combatants)
	if err != nil {
		return nil, err
	}

	s.combats[combat.ID] = combat
	return combat, nil
}

func (s *CombatService) GetCombat(_ context.Context, combatID string) (*models.Combat, error) {
	combat, exists := s.combats[combatID]
	if !exists {
		return nil, fmt.Errorf(errCombatNotFound)
	}
	return combat, nil
}

func (s *CombatService) GetCombatBySession(_ context.Context, gameSessionID string) (*models.Combat, error) {
	for _, combat := range s.combats {
		if combat.GameSessionID == gameSessionID && combat.IsActive {
			return combat, nil
		}
	}
	return nil, fmt.Errorf("no active combat for session")
}

func (s *CombatService) NextTurn(ctx context.Context, combatID string) (*models.Combatant, error) {
	combat, err := s.GetCombat(ctx, combatID)
	if err != nil {
		return nil, err
	}

	combatant, hasNext := s.engine.NextTurn(combat)
	if !hasNext {
		return nil, fmt.Errorf("no more turns")
	}

	return combatant, nil
}

func (s *CombatService) ProcessAction(ctx context.Context, combatID string, request models.CombatRequest) (*models.CombatAction, error) {
	combat, err := s.GetCombat(ctx, combatID)
	if err != nil {
		return nil, err
	}

	// Find actor
	actor := s.findCombatant(combat, request.ActorID)
	if actor == nil {
		return nil, fmt.Errorf(errActorNotFound)
	}

	// Create action record
	action := s.createCombatAction(combatID, combat.Round, request)

	// Process the action
	err = s.executeAction(combat, actor, request, action)
	if err != nil {
		return nil, err
	}

	// Auto-advance turn after most actions (except reactions and some special cases)
	if s.shouldAdvanceTurn(request.Action) {
		s.engine.NextTurn(combat)
	}

	return action, nil
}

func (s *CombatService) findCombatant(combat *models.Combat, combatantID string) *models.Combatant {
	for i := range combat.Combatants {
		if combat.Combatants[i].ID == combatantID {
			return &combat.Combatants[i]
		}
	}
	return nil
}

func (s *CombatService) createCombatAction(combatID string, round int, request models.CombatRequest) *models.CombatAction {
	return &models.CombatAction{
		ID:          uuid.New().String(),
		CombatID:    combatID,
		Round:       round,
		ActorID:     request.ActorID,
		ActionType:  request.Action,
		TargetID:    request.TargetID,
		Description: request.Description,
	}
}

func (s *CombatService) shouldAdvanceTurn(actionType models.ActionType) bool {
	return actionType != models.ActionTypeReaction && actionType != models.ActionTypeConcentration
}

func (s *CombatService) executeAction(combat *models.Combat, actor *models.Combatant, request models.CombatRequest, action *models.CombatAction) error {
	// Handle implemented actions
	switch request.Action {
	case models.ActionTypeAttack:
		return s.processAttack(combat, actor, request, action)
	case models.ActionTypeMove:
		return s.processMovement(combat, actor, request, action)
	case models.ActionTypeDeathSave:
		return s.processDeathSave(combat, actor, action)
	case models.ActionTypeDash:
		return s.processDash(combat, actor, action)
	case models.ActionTypeDodge:
		return s.processDodge(combat, actor, action)
	case models.ActionTypeEndTurn:
		action.Description = fmt.Sprintf("%s ends their turn", actor.Name)
		return nil
	default:
		return s.handleUnimplementedAction(request.Action)
	}
}

func (s *CombatService) handleUnimplementedAction(actionType models.ActionType) error {
	unimplementedActions := map[models.ActionType]string{
		models.ActionTypeCast:          "spell casting",
		models.ActionTypeCastSpell:     "spell casting",
		models.ActionTypeHelp:          "help action",
		models.ActionTypeHide:          "hide action",
		models.ActionTypeReady:         "ready action",
		models.ActionTypeSearch:        "search action",
		models.ActionTypeUseItem:       "use item action",
		models.ActionTypeBonusAction:   "bonus action",
		models.ActionTypeReaction:      "reaction",
		models.ActionTypeConcentration: "concentration check",
		models.ActionTypeSavingThrow:   "saving throw",
	}

	if actionName, ok := unimplementedActions[actionType]; ok {
		return fmt.Errorf("%s not yet implemented", actionName)
	}
	
	return fmt.Errorf("unsupported action type: %s", actionType)
}

func (s *CombatService) processAttack(combat *models.Combat, actor *models.Combatant, request models.CombatRequest, action *models.CombatAction) error {
	// Check action economy
	if err := s.engine.UseAction(actor, models.ActionTypeAttack); err != nil {
		return err
	}

	// Find target
	target := s.findCombatant(combat, request.TargetID)
	if target == nil {
		return fmt.Errorf(errTargetNotFound)
	}

	// Make attack roll
	attackRoll, err := s.performAttackRoll(actor, target, request)
	if err != nil {
		return err
	}
	action.Rolls = append(action.Rolls, *attackRoll)

	// Process hit or miss
	if s.isHit(attackRoll, target) {
		return s.processHit(actor, target, attackRoll, action)
	}
	
	action.Description = fmt.Sprintf("%s misses %s", actor.Name, target.Name)
	return nil
}

func (s *CombatService) performAttackRoll(actor, target *models.Combatant, request models.CombatRequest) (*models.Roll, error) {
	hasAdvantage := request.Advantage || s.engine.AttacksHaveAdvantage(target)
	hasDisadvantage := request.Disadvantage || s.engine.HasAttackDisadvantage(actor)
	return s.engine.AttackRoll(actor.AttackBonus, hasAdvantage, hasDisadvantage)
}

func (s *CombatService) isHit(attackRoll *models.Roll, target *models.Combatant) bool {
	return attackRoll.Result >= target.AC || attackRoll.Critical
}

func (s *CombatService) processHit(actor, target *models.Combatant, attackRoll *models.Roll, action *models.CombatAction) error {
	// Roll damage (example with 1d8+3 damage)
	damageRoll, damage, err := s.engine.DamageRoll("1d8", 3, models.DamageTypeSlashing, attackRoll.Critical)
	if err != nil {
		return err
	}
	action.Rolls = append(action.Rolls, *damageRoll)
	action.Damage = damage

	// Apply damage
	totalDamage := s.engine.ApplyDamage(target, damage)

	// Check for concentration if applicable
	if target.IsConcentrating && totalDamage > 0 {
		s.checkConcentration(target, totalDamage, action)
	}

	action.Description = fmt.Sprintf("%s hits %s for %d damage", actor.Name, target.Name, totalDamage)
	return nil
}

func (s *CombatService) checkConcentration(target *models.Combatant, damage int, action *models.CombatAction) {
	concRoll, success, err := s.engine.ConcentrationCheck(target, damage)
	if err != nil {
		return
	}
	
	action.Rolls = append(action.Rolls, *concRoll)
	if !success {
		s.engine.BreakConcentration(target)
		action.Effects = append(action.Effects, "Concentration broken")
	}
}

func (s *CombatService) processMovement(_ *models.Combat, actor *models.Combatant, request models.CombatRequest, action *models.CombatAction) error {
	// Calculate distance
	distance := 5 // Example: each square is 5 feet

	err := s.engine.UseMovement(actor, distance)
	if err != nil {
		return err
	}

	action.Description = fmt.Sprintf("%s moves %d feet", actor.Name, distance)
	return nil
}

func (s *CombatService) processDeathSave(_ *models.Combat, actor *models.Combatant, action *models.CombatAction) error {
	roll, err := s.engine.DeathSavingThrow(actor)
	if err != nil {
		return err
	}

	action.Rolls = append(action.Rolls, *roll)

	if roll.Critical {
		action.Description = fmt.Sprintf("%s rolls a natural 20 on death save and regains consciousness with 1 HP!", actor.Name)
	} else if actor.DeathSaves.IsStable {
		action.Description = fmt.Sprintf("%s is now stable", actor.Name)
	} else if actor.DeathSaves.IsDead {
		action.Description = fmt.Sprintf("%s has died", actor.Name)
	} else {
		successes := actor.DeathSaves.Successes
		failures := actor.DeathSaves.Failures
		action.Description = fmt.Sprintf("%s death saves: %d successes, %d failures", actor.Name, successes, failures)
	}

	return nil
}

func (s *CombatService) processDash(_ *models.Combat, actor *models.Combatant, action *models.CombatAction) error {
	if err := s.engine.UseAction(actor, models.ActionTypeDash); err != nil {
		return err
	}

	actor.Movement += actor.Speed
	action.Description = fmt.Sprintf("%s takes the Dash action, gaining %d feet of movement", actor.Name, actor.Speed)
	return nil
}

func (s *CombatService) processDodge(_ *models.Combat, actor *models.Combatant, action *models.CombatAction) error {
	if err := s.engine.UseAction(actor, models.ActionTypeDodge); err != nil {
		return err
	}

	// Apply dodge condition (would give disadvantage to attacks against them)
	s.engine.ApplyCondition(actor, models.Condition("dodging"))
	action.Description = fmt.Sprintf("%s takes the Dodge action", actor.Name)
	action.Effects = append(action.Effects, "Dodging until start of next turn")
	return nil
}

func (s *CombatService) EndCombat(ctx context.Context, combatID string) error {
	combat, err := s.GetCombat(ctx, combatID)
	if err != nil {
		return err
	}

	combat.IsActive = false
	delete(s.combats, combatID)
	return nil
}

func (s *CombatService) MakeSavingThrow(ctx context.Context, combatID, combatantID, ability string, dc int, advantage, disadvantage bool) (*models.Roll, bool, error) {
	combat, err := s.GetCombat(ctx, combatID)
	if err != nil {
		return nil, false, err
	}

	var combatant *models.Combatant
	for i := range combat.Combatants {
		if combat.Combatants[i].ID == combatantID {
			combatant = &combat.Combatants[i]
			break
		}
	}
	if combatant == nil {
		return nil, false, fmt.Errorf(errCombatantNotFound)
	}

	return s.engine.SavingThrow(combatant, ability, dc, advantage, disadvantage)
}

func (s *CombatService) ApplyDamage(ctx context.Context, combatID, combatantID string, damage []models.Damage) (int, error) {
	combat, err := s.GetCombat(ctx, combatID)
	if err != nil {
		return 0, err
	}

	var combatant *models.Combatant
	for i := range combat.Combatants {
		if combat.Combatants[i].ID == combatantID {
			combatant = &combat.Combatants[i]
			break
		}
	}
	if combatant == nil {
		return 0, fmt.Errorf(errCombatantNotFound)
	}

	return s.engine.ApplyDamage(combatant, damage), nil
}

func (s *CombatService) HealCombatant(ctx context.Context, combatID, combatantID string, healing int) error {
	combat, err := s.GetCombat(ctx, combatID)
	if err != nil {
		return err
	}

	var combatant *models.Combatant
	for i := range combat.Combatants {
		if combat.Combatants[i].ID == combatantID {
			combatant = &combat.Combatants[i]
			break
		}
	}
	if combatant == nil {
		return fmt.Errorf(errCombatantNotFound)
	}

	// Heal cannot exceed max HP
	combatant.HP += healing
	if combatant.HP > combatant.MaxHP {
		combatant.HP = combatant.MaxHP
	}

	// Reset death saves if healing from 0
	if combatant.HP > 0 && (combatant.DeathSaves.Successes > 0 || combatant.DeathSaves.Failures > 0) {
		combatant.DeathSaves = models.DeathSaves{}
	}

	return nil
}

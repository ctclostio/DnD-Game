package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/game"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
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
		return nil, fmt.Errorf("combat not found")
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
	var actor *models.Combatant
	for i := range combat.Combatants {
		if combat.Combatants[i].ID == request.ActorID {
			actor = &combat.Combatants[i]
			break
		}
	}
	if actor == nil {
		return nil, fmt.Errorf("actor not found")
	}

	// Create action record
	action := &models.CombatAction{
		ID:          uuid.New().String(),
		CombatID:    combatID,
		Round:       combat.Round,
		ActorID:     request.ActorID,
		ActionType:  request.Action,
		TargetID:    request.TargetID,
		Description: request.Description,
	}

	// Process based on action type
	switch request.Action {
	case models.ActionTypeAttack:
		err = s.processAttack(combat, actor, request, action)
	case models.ActionTypeMove:
		err = s.processMovement(combat, actor, request, action)
	case models.ActionTypeDeathSave:
		err = s.processDeathSave(combat, actor, action)
	case models.ActionTypeDash:
		err = s.processDash(combat, actor, action)
	case models.ActionTypeDodge:
		err = s.processDodge(combat, actor, action)
	case models.ActionTypeEndTurn:
		// End turn just creates the action, turn will advance below
		action.Description = fmt.Sprintf("%s ends their turn", actor.Name)
	case models.ActionTypeCast, models.ActionTypeCastSpell:
		// TODO: Implement spell casting
		err = fmt.Errorf("spell casting not yet implemented")
	case models.ActionTypeHelp:
		// TODO: Implement help action
		err = fmt.Errorf("help action not yet implemented")
	case models.ActionTypeHide:
		// TODO: Implement hide action
		err = fmt.Errorf("hide action not yet implemented")
	case models.ActionTypeReady:
		// TODO: Implement ready action
		err = fmt.Errorf("ready action not yet implemented")
	case models.ActionTypeSearch:
		// TODO: Implement search action
		err = fmt.Errorf("search action not yet implemented")
	case models.ActionTypeUseItem:
		// TODO: Implement use item action
		err = fmt.Errorf("use item action not yet implemented")
	case models.ActionTypeBonusAction:
		// TODO: Implement bonus action
		err = fmt.Errorf("bonus action not yet implemented")
	case models.ActionTypeReaction:
		// TODO: Implement reaction
		err = fmt.Errorf("reaction not yet implemented")
	case models.ActionTypeConcentration:
		// TODO: Implement concentration check
		err = fmt.Errorf("concentration check not yet implemented")
	case models.ActionTypeSavingThrow:
		// TODO: Implement saving throw
		err = fmt.Errorf("saving throw not yet implemented")
	default:
		err = fmt.Errorf("unsupported action type: %s", request.Action)
	}

	if err != nil {
		return nil, err
	}

	// Auto-advance turn after most actions (except reactions and some special cases)
	if request.Action != models.ActionTypeReaction && request.Action != models.ActionTypeConcentration {
		s.engine.NextTurn(combat)
	}

	return action, nil
}

func (s *CombatService) processAttack(combat *models.Combat, actor *models.Combatant, request models.CombatRequest, action *models.CombatAction) error {
	// Check action economy
	if err := s.engine.UseAction(actor, models.ActionTypeAttack); err != nil {
		return err
	}

	// Find target
	var target *models.Combatant
	for i := range combat.Combatants {
		if combat.Combatants[i].ID == request.TargetID {
			target = &combat.Combatants[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("target not found")
	}

	// Determine advantage/disadvantage
	hasAdvantage := request.Advantage || s.engine.AttacksHaveAdvantage(target)
	hasDisadvantage := request.Disadvantage || s.engine.HasAttackDisadvantage(actor)

	// Make attack roll
	attackRoll, err := s.engine.AttackRoll(actor.AttackBonus, hasAdvantage, hasDisadvantage)
	if err != nil {
		return err
	}
	action.Rolls = append(action.Rolls, *attackRoll)

	// Check if hit
	if attackRoll.Result >= target.AC || attackRoll.Critical {
		// Roll damage (example with 1d8+3 damage)
		damageRoll, damage, err := s.engine.DamageRoll("1d8", 3, models.DamageTypeSlashing, attackRoll.Critical)
		if err != nil {
			return err
		}
		action.Rolls = append(action.Rolls, *damageRoll)
		action.Damage = damage

		// Apply damage
		totalDamage := s.engine.ApplyDamage(target, damage)

		// Check for concentration
		if target.IsConcentrating && totalDamage > 0 {
			concRoll, success, err := s.engine.ConcentrationCheck(target, totalDamage)
			if err == nil {
				action.Rolls = append(action.Rolls, *concRoll)
				if !success {
					s.engine.BreakConcentration(target)
					action.Effects = append(action.Effects, "Concentration broken")
				}
			}
		}

		action.Description = fmt.Sprintf("%s hits %s for %d damage", actor.Name, target.Name, totalDamage)
	} else {
		action.Description = fmt.Sprintf("%s misses %s", actor.Name, target.Name)
	}

	return nil
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
		return nil, false, fmt.Errorf("combatant not found")
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
		return 0, fmt.Errorf("combatant not found")
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
		return fmt.Errorf("combatant not found")
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

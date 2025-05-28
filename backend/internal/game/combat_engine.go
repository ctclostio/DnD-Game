package game

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/pkg/dice"
)

type CombatEngine struct {
	roller *dice.Roller
}

func NewCombatEngine() *CombatEngine {
	return &CombatEngine{
		roller: dice.NewRoller(),
	}
}

// Initiative and Turn Order
func (ce *CombatEngine) RollInitiative(dexterityModifier int) (int, int, error) {
	result, err := ce.roller.Roll("1d20")
	if err != nil {
		return 0, 0, err
	}
	roll := result.Dice[0]
	total := roll + dexterityModifier
	return roll, total, nil
}

func (ce *CombatEngine) StartCombat(gameSessionID string, combatants []models.Combatant) (*models.Combat, error) {
	// Roll initiative for each combatant
	for i := range combatants {
		if combatants[i].Initiative == 0 {
			dexMod := (combatants[i].Abilities["dexterity"] - 10) / 2
			roll, total, err := ce.RollInitiative(dexMod)
			if err != nil {
				return nil, err
			}
			combatants[i].InitiativeRoll = roll
			combatants[i].Initiative = total
		}
		combatants[i].ID = uuid.New().String()
		
		// Reset action economy
		combatants[i].Actions = 1
		combatants[i].BonusActions = 1
		combatants[i].Reactions = 1
		combatants[i].Movement = combatants[i].Speed
	}

	// Sort by initiative (descending)
	sort.Slice(combatants, func(i, j int) bool {
		if combatants[i].Initiative == combatants[j].Initiative {
			// Tie-breaker: higher dexterity goes first
			return combatants[i].Abilities["dexterity"] > combatants[j].Abilities["dexterity"]
		}
		return combatants[i].Initiative > combatants[j].Initiative
	})

	// Create turn order
	turnOrder := make([]string, len(combatants))
	for i, c := range combatants {
		turnOrder[i] = c.ID
	}

	combat := &models.Combat{
		ID:            uuid.New().String(),
		GameSessionID: gameSessionID,
		Round:         1,
		CurrentTurn:   0,
		Combatants:    combatants,
		TurnOrder:     turnOrder,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return combat, nil
}

func (ce *CombatEngine) NextTurn(combat *models.Combat) (*models.Combatant, bool) {
	if !combat.IsActive || len(combat.TurnOrder) == 0 {
		return nil, false
	}

	// Increment turn
	combat.CurrentTurn++
	
	// Check if we need to start a new round
	if combat.CurrentTurn >= len(combat.TurnOrder) {
		combat.CurrentTurn = 0
		combat.Round++
		ce.StartNewRound(combat)
	}

	// Get current combatant
	currentID := combat.TurnOrder[combat.CurrentTurn]
	for i := range combat.Combatants {
		if combat.Combatants[i].ID == currentID {
			// Skip unconscious/dead combatants
			if combat.Combatants[i].HP <= 0 && !combat.Combatants[i].DeathSaves.IsStable {
				return ce.NextTurn(combat)
			}
			
			// Reset action economy for the combatant
			combat.Combatants[i].Actions = 1
			combat.Combatants[i].BonusActions = 1
			combat.Combatants[i].Movement = combat.Combatants[i].Speed
			
			return &combat.Combatants[i], true
		}
	}

	return nil, false
}

func (ce *CombatEngine) StartNewRound(combat *models.Combat) {
	// Update effect durations
	for i := range combat.ActiveEffects {
		combat.ActiveEffects[i].RemainingTime--
		if combat.ActiveEffects[i].RemainingTime <= 0 {
			// Remove expired effects
			combat.ActiveEffects = append(combat.ActiveEffects[:i], combat.ActiveEffects[i+1:]...)
			i--
		}
	}

	// Reset reactions for all combatants
	for i := range combat.Combatants {
		combat.Combatants[i].Reactions = 1
	}
}

// Attack System
func (ce *CombatEngine) AttackRoll(attackBonus int, advantage, disadvantage bool) (*models.Roll, error) {
	var result *dice.RollResult
	var err error

	if advantage && !disadvantage {
		result, err = ce.roller.RollAdvantage()
	} else if disadvantage && !advantage {
		result, err = ce.roller.RollDisadvantage()
	} else {
		result, err = ce.roller.Roll("1d20")
	}

	if err != nil {
		return nil, err
	}

	roll := &models.Roll{
		Type:         models.RollTypeAttack,
		Dice:         "1d20",
		Modifier:     attackBonus,
		Result:       result.Total + attackBonus,
		Individual:   result.Dice,
		Advantage:    advantage && !disadvantage,
		Disadvantage: disadvantage && !advantage,
		Critical:     result.Dice[0] == 20,
		CriticalMiss: result.Dice[0] == 1,
	}

	return roll, nil
}

func (ce *CombatEngine) DamageRoll(damageDice string, damageModifier int, damageType models.DamageType, isCritical bool) (*models.Roll, []models.Damage, error) {
	finalDice := damageDice
	if isCritical {
		// Double the number of dice on critical
		var numDice, dieSize int
		fmt.Sscanf(damageDice, "%dd%d", &numDice, &dieSize)
		finalDice = fmt.Sprintf("%dd%d", numDice*2, dieSize)
	}

	result, err := ce.roller.Roll(finalDice)
	if err != nil {
		return nil, nil, err
	}

	roll := &models.Roll{
		Type:       models.RollTypeDamage,
		Dice:       finalDice,
		Modifier:   damageModifier,
		Result:     result.Total + damageModifier,
		Individual: result.Dice,
		Critical:   isCritical,
	}

	damage := []models.Damage{{
		Amount: result.Total + damageModifier,
		Type:   damageType,
	}}

	return roll, damage, nil
}

func (ce *CombatEngine) ApplyDamage(combatant *models.Combatant, damage []models.Damage) int {
	totalDamage := 0

	for _, d := range damage {
		finalDamage := d.Amount

		// Check resistances
		for _, resistance := range combatant.Resistances {
			if resistance == d.Type {
				finalDamage = int(math.Floor(float64(finalDamage) / 2))
				break
			}
		}

		// Check immunities
		for _, immunity := range combatant.Immunities {
			if immunity == d.Type {
				finalDamage = 0
				break
			}
		}

		// Check vulnerabilities
		for _, vulnerability := range combatant.Vulnerabilities {
			if vulnerability == d.Type {
				finalDamage *= 2
				break
			}
		}

		totalDamage += finalDamage
	}

	// Apply damage to temp HP first
	if combatant.TempHP > 0 {
		if totalDamage <= combatant.TempHP {
			combatant.TempHP -= totalDamage
			totalDamage = 0
		} else {
			totalDamage -= combatant.TempHP
			combatant.TempHP = 0
		}
	}

	// Apply remaining damage to HP
	combatant.HP -= totalDamage
	if combatant.HP < 0 {
		combatant.HP = 0
	}

	return totalDamage
}

// Saving Throws
func (ce *CombatEngine) SavingThrow(combatant *models.Combatant, ability string, dc int, advantage, disadvantage bool) (*models.Roll, bool, error) {
	modifier := combatant.SavingThrows[ability]

	var result *dice.RollResult
	var err error

	if advantage && !disadvantage {
		result, err = ce.roller.RollAdvantage()
	} else if disadvantage && !advantage {
		result, err = ce.roller.RollDisadvantage()
	} else {
		result, err = ce.roller.Roll("1d20")
	}

	if err != nil {
		return nil, false, err
	}

	roll := &models.Roll{
		Type:         models.RollTypeSavingThrow,
		Dice:         "1d20",
		Modifier:     modifier,
		Result:       result.Total + modifier,
		Individual:   result.Dice,
		Advantage:    advantage && !disadvantage,
		Disadvantage: disadvantage && !advantage,
		Critical:     result.Dice[0] == 20,
		CriticalMiss: result.Dice[0] == 1,
	}

	success := roll.Result >= dc || roll.Critical

	return roll, success, nil
}

// Concentration
func (ce *CombatEngine) ConcentrationCheck(combatant *models.Combatant, damageTaken int) (*models.Roll, bool, error) {
	// DC is 10 or half the damage taken, whichever is higher
	dc := 10
	if damageTaken/2 > dc {
		dc = damageTaken / 2
	}

	// Constitution saving throw
	return ce.SavingThrow(combatant, "constitution", dc, false, false)
}

func (ce *CombatEngine) BreakConcentration(combatant *models.Combatant) {
	combatant.IsConcentrating = false
	combatant.ConcentrationSpell = ""
}

// Death Saves
func (ce *CombatEngine) DeathSavingThrow(combatant *models.Combatant) (*models.Roll, error) {
	if combatant.HP > 0 || combatant.DeathSaves.IsStable || combatant.DeathSaves.IsDead {
		return nil, fmt.Errorf("character does not need death saves")
	}

	result, err := ce.roller.Roll("1d20")
	if err != nil {
		return nil, err
	}

	roll := &models.Roll{
		Type:       models.RollTypeDeathSave,
		Dice:       "1d20",
		Modifier:   0,
		Result:     result.Total,
		Individual: result.Dice,
		Critical:   result.Dice[0] == 20,
		CriticalMiss: result.Dice[0] == 1,
	}

	// Natural 20: regain 1 HP
	if roll.Critical {
		combatant.HP = 1
		combatant.DeathSaves = models.DeathSaves{}
		return roll, nil
	}

	// Natural 1: two failures
	if roll.CriticalMiss {
		combatant.DeathSaves.Failures += 2
	} else if roll.Result >= 10 {
		combatant.DeathSaves.Successes++
	} else {
		combatant.DeathSaves.Failures++
	}

	// Check for stabilization or death
	if combatant.DeathSaves.Successes >= 3 {
		combatant.DeathSaves.IsStable = true
		combatant.DeathSaves.Successes = 0
		combatant.DeathSaves.Failures = 0
	} else if combatant.DeathSaves.Failures >= 3 {
		combatant.DeathSaves.IsDead = true
	}

	return roll, nil
}

// Action Economy
func (ce *CombatEngine) UseAction(combatant *models.Combatant, actionType models.ActionType) error {
	switch actionType {
	case models.ActionTypeAttack, models.ActionTypeCast, models.ActionTypeDash, 
	     models.ActionTypeDodge, models.ActionTypeHelp, models.ActionTypeHide, 
	     models.ActionTypeReady, models.ActionTypeSearch, models.ActionTypeUseItem:
		if combatant.Actions <= 0 {
			return fmt.Errorf("no actions remaining")
		}
		combatant.Actions--
		
	case models.ActionTypeBonusAction:
		if combatant.BonusActions <= 0 {
			return fmt.Errorf("no bonus actions remaining")
		}
		combatant.BonusActions--
		
	case models.ActionTypeReaction:
		if combatant.Reactions <= 0 {
			return fmt.Errorf("no reactions remaining")
		}
		combatant.Reactions--
		
	case models.ActionTypeMove:
		// Movement is tracked separately
		
	default:
		// Free actions don't consume resources
	}

	return nil
}

func (ce *CombatEngine) UseMovement(combatant *models.Combatant, distance int) error {
	if distance > combatant.Movement {
		return fmt.Errorf("insufficient movement: %d feet remaining", combatant.Movement)
	}
	combatant.Movement -= distance
	return nil
}

// Conditions
func (ce *CombatEngine) ApplyCondition(combatant *models.Combatant, condition models.Condition) {
	// Check if condition already exists
	for _, c := range combatant.Conditions {
		if c == condition {
			return
		}
	}
	combatant.Conditions = append(combatant.Conditions, condition)
}

func (ce *CombatEngine) RemoveCondition(combatant *models.Combatant, condition models.Condition) {
	for i, c := range combatant.Conditions {
		if c == condition {
			combatant.Conditions = append(combatant.Conditions[:i], combatant.Conditions[i+1:]...)
			return
		}
	}
}

func (ce *CombatEngine) HasCondition(combatant *models.Combatant, condition models.Condition) bool {
	for _, c := range combatant.Conditions {
		if c == condition {
			return true
		}
	}
	return false
}

// Helper function to check if combatant has disadvantage on attack rolls
func (ce *CombatEngine) HasAttackDisadvantage(combatant *models.Combatant) bool {
	disadvantageConditions := []models.Condition{
		models.ConditionBlinded,
		models.ConditionFrightened,
		models.ConditionPoisoned,
		models.ConditionProne,
		models.ConditionRestrained,
		models.ConditionExhaustion3,
	}

	for _, condition := range disadvantageConditions {
		if ce.HasCondition(combatant, condition) {
			return true
		}
	}
	return false
}

// Helper function to check if attacks against combatant have advantage
func (ce *CombatEngine) AttacksHaveAdvantage(target *models.Combatant) bool {
	advantageConditions := []models.Condition{
		models.ConditionBlinded,
		models.ConditionParalyzed,
		models.ConditionPetrified,
		models.ConditionProne,
		models.ConditionRestrained,
		models.ConditionStunned,
		models.ConditionUnconscious,
	}

	for _, condition := range advantageConditions {
		if ce.HasCondition(target, condition) {
			return true
		}
	}
	return false
}
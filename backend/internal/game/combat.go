package game

import (
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

func (ce *CombatEngine) RollInitiative(dexterityModifier int) (int, error) {
	result, err := ce.roller.Roll("1d20")
	if err != nil {
		return 0, err
	}
	return result.Total + dexterityModifier, nil
}

func (ce *CombatEngine) AttackRoll(attackBonus int, hasAdvantage, hasDisadvantage bool) (*AttackResult, error) {
	var result *dice.RollResult
	var err error
	
	if hasAdvantage && !hasDisadvantage {
		result, err = ce.roller.RollAdvantage()
	} else if hasDisadvantage && !hasAdvantage {
		result, err = ce.roller.RollDisadvantage()
	} else {
		result, err = ce.roller.Roll("1d20")
	}
	
	if err != nil {
		return nil, err
	}
	
	return &AttackResult{
		Roll:         result.Dice[0],
		Total:        result.Total + attackBonus,
		IsCritical:   result.Dice[0] == 20,
		IsCriticalMiss: result.Dice[0] == 1,
	}, nil
}

func (ce *CombatEngine) DamageRoll(damageDice string, damageModifier int, isCritical bool) (*DamageResult, error) {
	result, err := ce.roller.Roll(damageDice)
	if err != nil {
		return nil, err
	}
	
	total := result.Total + damageModifier
	
	// Double dice damage on critical hit
	if isCritical {
		critResult, err := ce.roller.Roll(damageDice)
		if err != nil {
			return nil, err
		}
		total += critResult.Total
	}
	
	return &DamageResult{
		Rolls:    result.Dice,
		Modifier: damageModifier,
		Total:    total,
		IsCritical: isCritical,
	}, nil
}

func (ce *CombatEngine) SavingThrow(ability int, proficiencyBonus int, hasProficiency bool) (*SavingThrowResult, error) {
	modifier := (ability - 10) / 2
	if hasProficiency {
		modifier += proficiencyBonus
	}
	
	result, err := ce.roller.Roll("1d20")
	if err != nil {
		return nil, err
	}
	
	return &SavingThrowResult{
		Roll:     result.Dice[0],
		Modifier: modifier,
		Total:    result.Total + modifier,
		Natural20: result.Dice[0] == 20,
		Natural1:  result.Dice[0] == 1,
	}, nil
}

type AttackResult struct {
	Roll           int  `json:"roll"`
	Total          int  `json:"total"`
	IsCritical     bool `json:"isCritical"`
	IsCriticalMiss bool `json:"isCriticalMiss"`
}

type DamageResult struct {
	Rolls      []int `json:"rolls"`
	Modifier   int   `json:"modifier"`
	Total      int   `json:"total"`
	IsCritical bool  `json:"isCritical"`
}

type SavingThrowResult struct {
	Roll      int  `json:"roll"`
	Modifier  int  `json:"modifier"`
	Total     int  `json:"total"`
	Natural20 bool `json:"natural20"`
	Natural1  bool `json:"natural1"`
}
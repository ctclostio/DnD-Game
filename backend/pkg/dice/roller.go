package dice

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/ctclostio/DnD-Game/backend/pkg/game"
)

type Roller struct {
	rng *game.Random
}

type RollResult struct {
	Dice     []int
	Modifier int
	Total    int
}

func NewRoller() *Roller {
	return &Roller{
		rng: game.NewRandom(),
	}
}

// Roll parses dice notation like "2d6+3" or "1d20-2"
func (r *Roller) Roll(notation string) (*RollResult, error) {
	// Parse dice notation using regex
	re := regexp.MustCompile(`^(\d+)d(\d+)([+-]\d+)?$`)
	matches := re.FindStringSubmatch(notation)

	if len(matches) == 0 {
		return nil, errors.New("invalid dice notation")
	}

	count, _ := strconv.Atoi(matches[1])
	sides, _ := strconv.Atoi(matches[2])

	modifier := 0
	if len(matches) > 3 && matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	if count < 1 || count > 100 {
		return nil, errors.New("dice count must be between 1 and 100")
	}

	if sides < 2 || (sides != 4 && sides != 6 && sides != 8 && sides != 10 && sides != 12 && sides != 20 && sides != 100) {
		return nil, errors.New("invalid dice type")
	}

	result := &RollResult{
		Dice:     make([]int, count),
		Modifier: modifier,
		Total:    modifier,
	}

	for i := 0; i < count; i++ {
		roll := r.rng.RollDice(sides)
		result.Dice[i] = roll
		result.Total += roll
	}

	return result, nil
}

// RollAdvantage rolls with advantage (roll twice, take higher)
func (r *Roller) RollAdvantage() (*RollResult, error) {
	roll1, _ := r.Roll("1d20")
	roll2, _ := r.Roll("1d20")

	if roll1.Total >= roll2.Total {
		return roll1, nil
	}
	return roll2, nil
}

// RollDisadvantage rolls with disadvantage (roll twice, take lower)
func (r *Roller) RollDisadvantage() (*RollResult, error) {
	roll1, _ := r.Roll("1d20")
	roll2, _ := r.Roll("1d20")

	if roll1.Total <= roll2.Total {
		return roll1, nil
	}
	return roll2, nil
}

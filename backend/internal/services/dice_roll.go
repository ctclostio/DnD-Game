package services

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

type DiceRollService struct {
	repo            database.DiceRollRepository
	gameSessionRepo database.GameSessionRepository
}

func NewDiceRollService(repo database.DiceRollRepository) *DiceRollService {
	return &DiceRollService{
		repo: repo,
	}
}

// SetGameSessionRepo sets the game session repository (to avoid circular dependency)
func (s *DiceRollService) SetGameSessionRepo(repo database.GameSessionRepository) {
	s.gameSessionRepo = repo
}

// RollDice performs a dice roll and saves it to the database
func (s *DiceRollService) RollDice(ctx context.Context, roll *models.DiceRoll) error {
	// Validate input
	if roll.GameSessionID == "" {
		return fmt.Errorf("game session ID is required")
	}
	if roll.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if roll.RollNotation == "" {
		return fmt.Errorf("roll notation is required")
	}

	// Parse roll notation
	count, diceType, modifier, err := parseRollNotation(roll.RollNotation)
	if err != nil {
		return fmt.Errorf("invalid roll notation: %w", err)
	}

	roll.Count = count
	roll.DiceType = diceType
	roll.Modifier = modifier

	// Perform the rolls
	roll.Results = make([]int, count)
	roll.Total = modifier

	// Create a local random generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	diceMax := getDiceMax(diceType)

	for i := 0; i < count; i++ {
		result := rng.Intn(diceMax) + 1
		roll.Results[i] = result
		roll.Total += result
	}

	// Save to database
	return s.repo.Create(ctx, roll)
}

// GetRollByID retrieves a dice roll by ID
func (s *DiceRollService) GetRollByID(ctx context.Context, id string) (*models.DiceRoll, error) {
	return s.repo.GetByID(ctx, id)
}

// GetRollsBySession retrieves dice rolls for a game session
func (s *DiceRollService) GetRollsBySession(ctx context.Context, sessionID string, offset, limit int) ([]*models.DiceRoll, error) {
	return s.repo.GetByGameSession(ctx, sessionID, offset, limit)
}

// GetRollsByUser retrieves dice rolls for a user
func (s *DiceRollService) GetRollsByUser(ctx context.Context, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	return s.repo.GetByUser(ctx, userID, offset, limit)
}

// GetRollsBySessionAndUser retrieves dice rolls for a specific user in a game session
func (s *DiceRollService) GetRollsBySessionAndUser(ctx context.Context, sessionID, userID string, offset, limit int) ([]*models.DiceRoll, error) {
	return s.repo.GetByGameSessionAndUser(ctx, sessionID, userID, offset, limit)
}

// DeleteRoll deletes a dice roll
func (s *DiceRollService) DeleteRoll(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// SimulateRoll simulates a dice roll without saving to the database
func (s *DiceRollService) SimulateRoll(notation string) (*models.DiceRoll, error) {
	// Parse roll notation
	count, diceType, modifier, err := parseRollNotation(notation)
	if err != nil {
		return nil, err
	}

	roll := &models.DiceRoll{
		RollNotation: notation,
		Count:        count,
		DiceType:     diceType,
		Modifier:     modifier,
		Results:      make([]int, count),
		Total:        modifier,
	}

	// Perform the rolls
	// Create a local random generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	diceMax := getDiceMax(diceType)

	for i := 0; i < count; i++ {
		result := rng.Intn(diceMax) + 1
		roll.Results[i] = result
		roll.Total += result
	}

	return roll, nil
}

// parseRollNotation parses a dice roll notation like "2d20+5" or "1d6-2"
func parseRollNotation(notation string) (count int, diceType string, modifier int, err error) {
	// Regular expression to match dice notation
	re := regexp.MustCompile(`^(\d+)?d(\d+|%)([+-]\d+)?$`)
	matches := re.FindStringSubmatch(strings.ToLower(notation))

	if len(matches) == 0 {
		return 0, "", 0, fmt.Errorf("invalid dice notation format")
	}

	// Parse count (default to 1 if not specified)
	if matches[1] == "" {
		count = 1
	} else {
		count, err = strconv.Atoi(matches[1])
		if err != nil || count < 1 {
			return 0, "", 0, fmt.Errorf("invalid dice count")
		}
	}

	// Parse dice type
	if matches[2] == "%" {
		diceType = "d100"
	} else {
		diceNum, err := strconv.Atoi(matches[2])
		if err != nil {
			return 0, "", 0, fmt.Errorf("invalid dice type")
		}

		// Validate dice type
		validDice := map[int]bool{4: true, 6: true, 8: true, 10: true, 12: true, 20: true, 100: true}
		if !validDice[diceNum] {
			return 0, "", 0, fmt.Errorf("invalid dice type: d%d", diceNum)
		}

		diceType = fmt.Sprintf("d%d", diceNum)
	}

	// Parse modifier
	if matches[3] != "" {
		modifier, err = strconv.Atoi(matches[3])
		if err != nil {
			return 0, "", 0, fmt.Errorf("invalid modifier")
		}
	}

	return count, diceType, modifier, nil
}

// getDiceMax returns the maximum value for a dice type
func getDiceMax(diceType string) int {
	switch diceType {
	case "d4":
		return 4
	case "d6":
		return 6
	case "d8":
		return 8
	case "d10":
		return 10
	case "d12":
		return 12
	case "d20":
		return 20
	case "d100":
		return 100
	default:
		return 20 // Default to d20
	}
}

// RollInitiative rolls initiative for multiple participants
func (s *DiceRollService) RollInitiative(ctx context.Context, sessionID string, participants []struct {
	UserID      string
	CharacterID string
	DexModifier int
}) ([]struct {
	UserID      string
	CharacterID string
	Initiative  int
}, error) {
	results := make([]struct {
		UserID      string
		CharacterID string
		Initiative  int
	}, len(participants))

	for i, p := range participants {
		// Roll 1d20 + dexterity modifier
		rollNotation := "1d20"
		if p.DexModifier > 0 {
			rollNotation = fmt.Sprintf("1d20+%d", p.DexModifier)
		} else if p.DexModifier < 0 {
			rollNotation = fmt.Sprintf("1d20%d", p.DexModifier) // negative already has the minus sign
		}

		roll := &models.DiceRoll{
			GameSessionID: sessionID,
			UserID:        p.UserID,
			DiceType:      "d20",
			Count:         1,
			Modifier:      p.DexModifier,
			Purpose:       "initiative",
			RollNotation:  rollNotation,
		}

		if err := s.RollDice(ctx, roll); err != nil {
			return nil, fmt.Errorf("failed to roll initiative for user %s: %w", p.UserID, err)
		}

		results[i] = struct {
			UserID      string
			CharacterID string
			Initiative  int
		}{
			UserID:      p.UserID,
			CharacterID: p.CharacterID,
			Initiative:  roll.Total,
		}
	}

	return results, nil
}

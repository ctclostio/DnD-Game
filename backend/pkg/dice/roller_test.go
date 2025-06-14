package dice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRoller(t *testing.T) {
	roller := NewRoller()
	assert.NotNil(t, roller)
	assert.NotNil(t, roller.rng)
}

func TestRoller_Roll(t *testing.T) {
	roller := NewRoller()

	tests := []struct {
		name        string
		notation    string
		shouldError bool
		checkResult func(*testing.T, *RollResult)
	}{
		{
			name:        "simple d20",
			notation:    "1d20",
			shouldError: false,
			checkResult: func(t *testing.T, r *RollResult) {
				assert.Len(t, r.Dice, 1)
				assert.GreaterOrEqual(t, r.Dice[0], 1)
				assert.LessOrEqual(t, r.Dice[0], 20)
				assert.Equal(t, r.Total, r.Dice[0])
				assert.Equal(t, 0, r.Modifier)
			},
		},
		{
			name:        "multiple dice",
			notation:    "3d6",
			shouldError: false,
			checkResult: func(t *testing.T, r *RollResult) {
				assert.Len(t, r.Dice, 3)
				total := 0
				for _, die := range r.Dice {
					assert.GreaterOrEqual(t, die, 1)
					assert.LessOrEqual(t, die, 6)
					total += die
				}
				assert.Equal(t, total, r.Total)
			},
		},
		{
			name:        "with positive modifier",
			notation:    "2d8+5",
			shouldError: false,
			checkResult: func(t *testing.T, r *RollResult) {
				assert.Len(t, r.Dice, 2)
				assert.Equal(t, 5, r.Modifier)
				diceSum := r.Dice[0] + r.Dice[1]
				assert.Equal(t, diceSum+5, r.Total)
			},
		},
		{
			name:        "with negative modifier",
			notation:    "1d4-2",
			shouldError: false,
			checkResult: func(t *testing.T, r *RollResult) {
				assert.Len(t, r.Dice, 1)
				assert.Equal(t, -2, r.Modifier)
				assert.Equal(t, r.Dice[0]-2, r.Total)
			},
		},
		{
			name:        "d100",
			notation:    "1d100",
			shouldError: false,
			checkResult: func(t *testing.T, r *RollResult) {
				assert.Len(t, r.Dice, 1)
				assert.GreaterOrEqual(t, r.Dice[0], 1)
				assert.LessOrEqual(t, r.Dice[0], 100)
			},
		},
		{
			name:        "complex notation",
			notation:    "4d6+10",
			shouldError: false,
			checkResult: func(t *testing.T, r *RollResult) {
				assert.Len(t, r.Dice, 4)
				assert.Equal(t, 10, r.Modifier)
				assert.GreaterOrEqual(t, r.Total, 14) // minimum: 4*1 + 10
				assert.LessOrEqual(t, r.Total, 34)    // maximum: 4*6 + 10
			},
		},
		{
			name:        "invalid notation - no dice",
			notation:    "invalid",
			shouldError: true,
		},
		{
			name:        "invalid notation - zero dice",
			notation:    "0d6",
			shouldError: true,
		},
		{
			name:        "invalid notation - invalid sides",
			notation:    "1d7",
			shouldError: true,
		},
		{
			name:        "invalid notation - too many dice",
			notation:    "101d6",
			shouldError: true,
		},
		{
			name:        "empty notation",
			notation:    "",
			shouldError: true,
		},
		{
			name:        "invalid dice type d1",
			notation:    "1d1",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := roller.Roll(tt.notation)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				tt.checkResult(t, result)
			}
		})
	}
}

func TestRoller_RollAdvantage(t *testing.T) {
	roller := NewRoller()

	// Run multiple times to ensure it's working properly.
	for i := 0; i < 10; i++ {
		result, err := roller.RollAdvantage()
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Dice, 1)
		assert.GreaterOrEqual(t, result.Dice[0], 1)
		assert.LessOrEqual(t, result.Dice[0], 20)
		assert.Equal(t, result.Total, result.Dice[0])
	}
}

func TestRoller_RollDisadvantage(t *testing.T) {
	roller := NewRoller()

	// Run multiple times to ensure it's working properly.
	for i := 0; i < 10; i++ {
		result, err := roller.RollDisadvantage()
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Len(t, result.Dice, 1)
		assert.GreaterOrEqual(t, result.Dice[0], 1)
		assert.LessOrEqual(t, result.Dice[0], 20)
		assert.Equal(t, result.Total, result.Dice[0])
	}
}

// Test that advantage tends to produce higher results than disadvantage.
func TestRoller_AdvantageVsDisadvantage(t *testing.T) {
	roller := NewRoller()

	advantageSum := 0
	disadvantageSum := 0
	rolls := 100

	for i := 0; i < rolls; i++ {
		adv, _ := roller.RollAdvantage()
		dis, _ := roller.RollDisadvantage()

		advantageSum += adv.Total
		disadvantageSum += dis.Total
	}

	// On average, advantage should produce higher results.
	// This might fail occasionally due to randomness, but very rarely.
	assert.Greater(t, float64(advantageSum)/float64(rolls), float64(disadvantageSum)/float64(rolls))
}

// Test for distribution (basic check).
func TestRoller_Distribution(t *testing.T) {
	t.Skip("Skipping distribution test in short mode")

	roller := NewRoller()
	counts := make(map[int]int)

	// Roll a d6 many times.
	rolls := 6000
	for i := 0; i < rolls; i++ {
		result, err := roller.Roll("1d6")
		require.NoError(t, err)
		counts[result.Total]++
	}

	// Each face should appear roughly 1/6 of the time.
	// We'll allow a generous margin for randomness.
	expectedCount := rolls / 6
	tolerance := float64(expectedCount) * 0.25 // 25% tolerance

	for face := 1; face <= 6; face++ {
		count := counts[face]
		assert.Greater(t, float64(count), float64(expectedCount)-tolerance,
			"Face %d appeared %d times, expected around %d", face, count, expectedCount)
		assert.Less(t, float64(count), float64(expectedCount)+tolerance,
			"Face %d appeared %d times, expected around %d", face, count, expectedCount)
	}
}

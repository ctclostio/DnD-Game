package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/services/mocks"
)

func TestDiceRollService_RollDice(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		roll          *models.DiceRoll
		setupMock     func(*mocks.MockDiceRollRepository)
		expectedError string
		validate      func(*testing.T, *models.DiceRoll)
	}{
		{
			name: "simple d20 roll",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "1d20",
				Purpose:       "attack",
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(roll *models.DiceRoll) bool {
					return roll.RollNotation == "1d20" && 
						roll.Count == 1 && 
						roll.DiceType == "d20" &&
						len(roll.Results) == 1 &&
						roll.Results[0] >= 1 && roll.Results[0] <= 20
				})).Return(nil)
			},
			validate: func(t *testing.T, roll *models.DiceRoll) {
				assert.Equal(t, "1d20", roll.RollNotation)
				assert.Equal(t, 1, roll.Count)
				assert.Equal(t, "d20", roll.DiceType)
				assert.Len(t, roll.Results, 1)
				assert.GreaterOrEqual(t, roll.Results[0], 1)
				assert.LessOrEqual(t, roll.Results[0], 20)
				assert.Equal(t, roll.Total, roll.Results[0])
			},
		},
		{
			name: "multiple dice",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "3d6",
				Purpose:       "ability check",
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(roll *models.DiceRoll) bool {
					return roll.Count == 3 && roll.DiceType == "d6" && len(roll.Results) == 3
				})).Return(nil)
			},
			validate: func(t *testing.T, roll *models.DiceRoll) {
				assert.Equal(t, "3d6", roll.RollNotation)
				assert.Equal(t, 3, roll.Count)
				assert.Equal(t, "d6", roll.DiceType)
				assert.Len(t, roll.Results, 3)
				total := 0
				for _, result := range roll.Results {
					assert.GreaterOrEqual(t, result, 1)
					assert.LessOrEqual(t, result, 6)
					total += result
				}
				assert.Equal(t, total, roll.Total)
			},
		},
		{
			name: "roll with modifier",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "2d8+3",
				Purpose:       "damage",
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(roll *models.DiceRoll) bool {
					return roll.Count == 2 && 
						roll.DiceType == "d8" && 
						roll.Modifier == 3 &&
						len(roll.Results) == 2
				})).Return(nil)
			},
			validate: func(t *testing.T, roll *models.DiceRoll) {
				assert.Equal(t, "2d8+3", roll.RollNotation)
				assert.Equal(t, 2, roll.Count)
				assert.Equal(t, "d8", roll.DiceType)
				assert.Equal(t, 3, roll.Modifier)
				assert.Len(t, roll.Results, 2)
				sum := 0
				for _, result := range roll.Results {
					assert.GreaterOrEqual(t, result, 1)
					assert.LessOrEqual(t, result, 8)
					sum += result
				}
				assert.Equal(t, sum+3, roll.Total)
			},
		},
		{
			name: "percentile dice",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "1d%",
				Purpose:       "percentile check",
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(roll *models.DiceRoll) bool {
					return roll.DiceType == "d100" && roll.Results[0] >= 1 && roll.Results[0] <= 100
				})).Return(nil)
			},
			validate: func(t *testing.T, roll *models.DiceRoll) {
				assert.Equal(t, "d100", roll.DiceType)
				assert.Len(t, roll.Results, 1)
				assert.GreaterOrEqual(t, roll.Results[0], 1)
				assert.LessOrEqual(t, roll.Results[0], 100)
			},
		},
		{
			name: "repository error",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "1d20",
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name: "invalid notation - no dice",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "invalid",
			},
			expectedError: "invalid roll notation",
		},
		{
			name: "missing game session ID",
			roll: &models.DiceRoll{
				GameSessionID: "",
				UserID:        "user-123",
				RollNotation:  "1d20",
			},
			expectedError: "game session ID is required",
		},
		{
			name: "missing user ID",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "",
				RollNotation:  "1d20",
			},
			expectedError: "user ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockDiceRollRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewDiceRollService(mockRepo)
			err := service.RollDice(ctx, tt.roll)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.roll)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDiceRollService_GetRollByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		rollID        string
		setupMock     func(*mocks.MockDiceRollRepository)
		expectedError string
		validate      func(*testing.T, *models.DiceRoll)
	}{
		{
			name:   "successful get",
			rollID: "roll-123",
			setupMock: func(m *mocks.MockDiceRollRepository) {
				roll := &models.DiceRoll{
					ID:            "roll-123",
					GameSessionID: "session-123",
					UserID:        "user-123",
					RollNotation:  "1d20+5",
					DiceType:      "d20",
					Count:         1,
					Modifier:      5,
					Results:       []int{15},
					Total:         20,
					Purpose:       "attack",
					Timestamp:     time.Now(),
				}
				m.On("GetByID", ctx, "roll-123").Return(roll, nil)
			},
			validate: func(t *testing.T, roll *models.DiceRoll) {
				assert.Equal(t, "roll-123", roll.ID)
				assert.Equal(t, "1d20+5", roll.RollNotation)
				assert.Equal(t, 20, roll.Total)
			},
		},
		{
			name:   "roll not found",
			rollID: "nonexistent",
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockDiceRollRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewDiceRollService(mockRepo)
			roll, err := service.GetRollByID(ctx, tt.rollID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, roll)
			} else {
				require.NoError(t, err)
				require.NotNil(t, roll)
				if tt.validate != nil {
					tt.validate(t, roll)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDiceRollService_GetRollsBySession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		sessionID     string
		offset        int
		limit         int
		setupMock     func(*mocks.MockDiceRollRepository)
		expectedError string
		validate      func(*testing.T, []*models.DiceRoll)
	}{
		{
			name:      "successful get multiple rolls",
			sessionID: "session-123",
			offset:    0,
			limit:     10,
			setupMock: func(m *mocks.MockDiceRollRepository) {
				rolls := []*models.DiceRoll{
					{
						ID:            "roll-1",
						GameSessionID: "session-123",
						RollNotation:  "1d20",
						Total:         15,
					},
					{
						ID:            "roll-2",
						GameSessionID: "session-123",
						RollNotation:  "2d6+3",
						Total:         10,
					},
				}
				m.On("GetByGameSession", ctx, "session-123", 0, 10).Return(rolls, nil)
			},
			validate: func(t *testing.T, rolls []*models.DiceRoll) {
				assert.Len(t, rolls, 2)
				assert.Equal(t, "roll-1", rolls[0].ID)
				assert.Equal(t, "roll-2", rolls[1].ID)
			},
		},
		{
			name:      "empty results",
			sessionID: "session-456",
			offset:    0,
			limit:     10,
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("GetByGameSession", ctx, "session-456", 0, 10).Return([]*models.DiceRoll{}, nil)
			},
			validate: func(t *testing.T, rolls []*models.DiceRoll) {
				assert.Empty(t, rolls)
			},
		},
		{
			name:      "repository error",
			sessionID: "session-789",
			offset:    0,
			limit:     10,
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("GetByGameSession", ctx, "session-789", 0, 10).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockDiceRollRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewDiceRollService(mockRepo)
			rolls, err := service.GetRollsBySession(ctx, tt.sessionID, tt.offset, tt.limit)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, rolls)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rolls)
				if tt.validate != nil {
					tt.validate(t, rolls)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDiceRollService_SimulateRoll(t *testing.T) {
	service := services.NewDiceRollService(nil)

	tests := []struct {
		name          string
		notation      string
		expectedError string
		validate      func(*testing.T, *models.DiceRoll)
	}{
		{
			name:     "simple d20",
			notation: "1d20",
			validate: func(t *testing.T, roll *models.DiceRoll) {
				assert.Equal(t, 1, roll.Count)
				assert.Equal(t, "d20", roll.DiceType)
				assert.Equal(t, 0, roll.Modifier)
				assert.Len(t, roll.Results, 1)
				assert.GreaterOrEqual(t, roll.Results[0], 1)
				assert.LessOrEqual(t, roll.Results[0], 20)
			},
		},
		{
			name:     "multiple dice with modifier",
			notation: "3d6+2",
			validate: func(t *testing.T, roll *models.DiceRoll) {
				assert.Equal(t, 3, roll.Count)
				assert.Equal(t, "d6", roll.DiceType)
				assert.Equal(t, 2, roll.Modifier)
				assert.Len(t, roll.Results, 3)
				sum := 0
				for _, r := range roll.Results {
					assert.GreaterOrEqual(t, r, 1)
					assert.LessOrEqual(t, r, 6)
					sum += r
				}
				assert.Equal(t, sum+2, roll.Total)
			},
		},
		{
			name:     "negative modifier",
			notation: "1d10-2",
			validate: func(t *testing.T, roll *models.DiceRoll) {
				assert.Equal(t, -2, roll.Modifier)
				assert.Equal(t, roll.Results[0]-2, roll.Total)
			},
		},
		{
			name:          "invalid notation",
			notation:      "invalid",
			expectedError: "invalid dice notation format",
		},
		{
			name:          "invalid dice type",
			notation:      "1d7",
			expectedError: "invalid dice type: d7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roll, err := service.SimulateRoll(tt.notation)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, roll)
			} else {
				require.NoError(t, err)
				require.NotNil(t, roll)
				if tt.validate != nil {
					tt.validate(t, roll)
				}
			}
		})
	}
}
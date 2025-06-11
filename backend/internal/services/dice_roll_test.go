package services_test

import (
	"context"
	"errors"
	"testing"

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
		validateRoll  func(*testing.T, *models.DiceRoll)
	}{
		{
			name: "successful d20 roll",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "1d20",
				Purpose:       "Attack roll",
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(r *models.DiceRoll) bool {
					return r.Total >= 1 && r.Total <= 20 &&
						len(r.Results) == 1 &&
						r.Results[0] >= 1 && r.Results[0] <= 20 &&
						r.Count == 1 &&
						r.DiceType == "d20" &&
						r.Modifier == 0
				})).Return(nil)
			},
			validateRoll: func(t *testing.T, r *models.DiceRoll) {
				assert.GreaterOrEqual(t, r.Total, 1)
				assert.LessOrEqual(t, r.Total, 20)
				assert.Len(t, r.Results, 1)
				assert.Equal(t, "d20", r.DiceType)
				assert.Equal(t, 1, r.Count)
				assert.Equal(t, 0, r.Modifier)
			},
		},
		{
			name: "multiple dice with modifier",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "2d6+3",
				Purpose:       "Damage roll",
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(r *models.DiceRoll) bool {
					return r.Total >= 5 && r.Total <= 15 && // 2-12 + 3
						len(r.Results) == 2 &&
						r.Count == 2 &&
						r.DiceType == "d6" &&
						r.Modifier == 3
				})).Return(nil)
			},
			validateRoll: func(t *testing.T, r *models.DiceRoll) {
				assert.GreaterOrEqual(t, r.Total, 5)
				assert.LessOrEqual(t, r.Total, 15)
				assert.Len(t, r.Results, 2)
				assert.Equal(t, "d6", r.DiceType)
				assert.Equal(t, 2, r.Count)
				assert.Equal(t, 3, r.Modifier)
			},
		},
		{
			name: "negative modifier",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "1d20-2",
				Purpose:       "Saving throw",
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(r *models.DiceRoll) bool {
					return r.Total >= -1 && r.Total <= 18 && // 1-20 - 2
						r.Modifier == -2
				})).Return(nil)
			},
			validateRoll: func(t *testing.T, r *models.DiceRoll) {
				assert.GreaterOrEqual(t, r.Total, -1)
				assert.LessOrEqual(t, r.Total, 18)
				assert.Equal(t, -2, r.Modifier)
			},
		},
		{
			name:          "missing game session ID",
			roll:          &models.DiceRoll{UserID: "user-123", RollNotation: "1d20"},
			expectedError: "game session ID is required",
		},
		{
			name:          "missing user ID",
			roll:          &models.DiceRoll{GameSessionID: "session-123", RollNotation: "1d20"},
			expectedError: "user ID is required",
		},
		{
			name:          "missing roll notation",
			roll:          &models.DiceRoll{GameSessionID: "session-123", UserID: "user-123"},
			expectedError: "roll notation is required",
		},
		{
			name: "invalid roll notation - no dice",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "not a roll",
			},
			expectedError: "invalid roll notation",
		},
		{
			name: "invalid dice type",
			roll: &models.DiceRoll{
				GameSessionID: "session-123",
				UserID:        "user-123",
				RollNotation:  "1d7",
			},
			expectedError: "invalid dice type",
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
				if tt.validateRoll != nil {
					tt.validateRoll(t, tt.roll)
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
		limit         int
		offset        int
		setupMock     func(*mocks.MockDiceRollRepository)
		expected      []*models.DiceRoll
		expectedError string
	}{
		{
			name:      "successful retrieval",
			sessionID: "session-123",
			limit:     10,
			offset:    0,
			setupMock: func(m *mocks.MockDiceRollRepository) {
				rolls := []*models.DiceRoll{
					{
						ID:            "roll-1",
						GameSessionID: "session-123",
						UserID:        "user-123",
						RollNotation:  "1d20+5",
						Total:         18,
						Purpose:       "Attack roll",
					},
					{
						ID:            "roll-2",
						GameSessionID: "session-123",
						UserID:        "user-456",
						RollNotation:  "2d6+3",
						Total:         10,
						Purpose:       "Damage roll",
					},
				}
				m.On("GetByGameSession", ctx, "session-123", 0, 10).Return(rolls, nil)
			},
			expected: []*models.DiceRoll{
				{
					ID:            "roll-1",
					GameSessionID: "session-123",
					UserID:        "user-123",
					RollNotation:  "1d20+5",
					Total:         18,
					Purpose:       "Attack roll",
				},
				{
					ID:            "roll-2",
					GameSessionID: "session-123",
					UserID:        "user-456",
					RollNotation:  "2d6+3",
					Total:         10,
					Purpose:       "Damage roll",
				},
			},
		},
		{
			name:      "empty results",
			sessionID: "session-999",
			limit:     10,
			offset:    0,
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("GetByGameSession", ctx, "session-999", 0, 10).Return([]*models.DiceRoll{}, nil)
			},
			expected: []*models.DiceRoll{},
		},
		{
			name:      "repository error",
			sessionID: "session-123",
			limit:     10,
			offset:    0,
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("GetByGameSession", ctx, "session-123", 0, 10).Return(nil, errors.New("database error"))
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
			result, err := service.GetRollsBySession(ctx, tt.sessionID, tt.offset, tt.limit)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDiceRollService_GetRollsByUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        string
		limit         int
		offset        int
		setupMock     func(*mocks.MockDiceRollRepository)
		expected      []*models.DiceRoll
		expectedError string
	}{
		{
			name:   "successful retrieval",
			userID: "user-123",
			limit:  20,
			offset: 0,
			setupMock: func(m *mocks.MockDiceRollRepository) {
				rolls := []*models.DiceRoll{
					{
						ID:            "roll-1",
						GameSessionID: "session-123",
						UserID:        "user-123",
						RollNotation:  "1d20",
						Total:         15,
					},
					{
						ID:            "roll-2",
						GameSessionID: "session-456",
						UserID:        "user-123",
						RollNotation:  "3d6",
						Total:         12,
					},
				}
				m.On("GetByUser", ctx, "user-123", 0, 20).Return(rolls, nil)
			},
			expected: []*models.DiceRoll{
				{
					ID:            "roll-1",
					GameSessionID: "session-123",
					UserID:        "user-123",
					RollNotation:  "1d20",
					Total:         15,
				},
				{
					ID:            "roll-2",
					GameSessionID: "session-456",
					UserID:        "user-123",
					RollNotation:  "3d6",
					Total:         12,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockDiceRollRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewDiceRollService(mockRepo)
			result, err := service.GetRollsByUser(ctx, tt.userID, tt.offset, tt.limit)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDiceRollService_SimulateRoll(t *testing.T) {
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
				assert.Equal(t, roll.Results[0], roll.Total)
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

	service := services.NewDiceRollService(nil)

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

func TestDiceRollService_RollInitiative(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		sessionID    string
		participants []struct {
			UserID      string
			CharacterID string
			DexModifier int
		}
		setupMock     func(*mocks.MockDiceRollRepository)
		expectedError string
		validate      func(*testing.T, []struct {
			UserID      string
			CharacterID string
			Initiative  int
		})
	}{
		{
			name:      "successful initiative rolls",
			sessionID: "session-123",
			participants: []struct {
				UserID      string
				CharacterID string
				DexModifier int
			}{
				{UserID: "user-1", CharacterID: "char-1", DexModifier: 2},
				{UserID: "user-2", CharacterID: "char-2", DexModifier: -1},
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(r *models.DiceRoll) bool {
					return r.Purpose == "initiative" &&
						r.DiceType == "d20" &&
						r.Count == 1
				})).Return(nil).Times(2)
			},
			validate: func(t *testing.T, results []struct {
				UserID      string
				CharacterID string
				Initiative  int
			}) {
				assert.Len(t, results, 2)
				// First participant has +2 modifier
				assert.GreaterOrEqual(t, results[0].Initiative, 3) // 1+2
				assert.LessOrEqual(t, results[0].Initiative, 22)   // 20+2
				// Second participant has -1 modifier
				assert.GreaterOrEqual(t, results[1].Initiative, 0) // 1-1
				assert.LessOrEqual(t, results[1].Initiative, 19)   // 20-1
			},
		},
		{
			name:      "repository error",
			sessionID: "session-123",
			participants: []struct {
				UserID      string
				CharacterID string
				DexModifier int
			}{
				{UserID: "user-1", CharacterID: "char-1", DexModifier: 0},
			},
			setupMock: func(m *mocks.MockDiceRollRepository) {
				m.On("Create", ctx, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "failed to roll initiative for user user-1: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockDiceRollRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := services.NewDiceRollService(mockRepo)
			results, err := service.RollInitiative(ctx, tt.sessionID, tt.participants)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, results)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

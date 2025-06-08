package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
)

// MockCharacterService is a mock implementation of the character service
type MockCharacterService struct {
	mock.Mock
}

func (m *MockCharacterService) GetAllCharacters(ctx context.Context, userID string) ([]*models.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

func (m *MockCharacterService) GetCharacterByID(ctx context.Context, id string) (*models.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterService) CreateCharacter(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterService) UpdateCharacter(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterService) DeleteCharacter(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCharacterService) UseSpellSlot(ctx context.Context, characterID string, spellLevel int) error {
	args := m.Called(ctx, characterID, spellLevel)
	return args.Error(0)
}

func (m *MockCharacterService) RestoreSpellSlots(ctx context.Context, characterID string, restType string) error {
	args := m.Called(ctx, characterID, restType)
	return args.Error(0)
}

func (m *MockCharacterService) GenerateCustomClass(ctx context.Context, userID, name, description, role, style, features string) (*models.CustomClass, error) {
	args := m.Called(ctx, userID, name, description, role, style, features)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CustomClass), args.Error(1)
}

func (m *MockCharacterService) GetUserCustomClasses(ctx context.Context, userID string, includeUnapproved bool) ([]*models.CustomClass, error) {
	args := m.Called(ctx, userID, includeUnapproved)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CustomClass), args.Error(1)
}

func (m *MockCharacterService) GetCustomClass(ctx context.Context, classID string) (*models.CustomClass, error) {
	args := m.Called(ctx, classID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CustomClass), args.Error(1)
}

func (m *MockCharacterService) AddExperience(ctx context.Context, characterID string, xp int) error {
	args := m.Called(ctx, characterID, xp)
	return args.Error(0)
}

func (m *MockCharacterService) GetXPForNextLevel(level int) int {
	args := m.Called(level)
	return args.Int(0)
}

// Helper function to create a test context with auth claims
func createAuthContext(userID, username, email, role string) context.Context {
	claims := &auth.Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		Type:     auth.AccessToken,
	}
	return context.WithValue(context.Background(), auth.UserContextKey, claims)
}

// Helper function to create test handlers
func createTestHandlers() (*Handlers, *MockCharacterService) {
	mockCharService := new(MockCharacterService)
	mockServices := &services.Services{
		Characters: mockCharService,
	}
	handlers := NewHandlers(mockServices, nil)
	return handlers, mockCharService
}

func TestHandlers_GetCharacters(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockCharacterService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:   "successful retrieval",
			userID: "user-123",
			setupMock: func(m *MockCharacterService) {
				characters := []*models.Character{
					{
						ID:     "char-1",
						UserID: "user-123",
						Name:   "Aragorn",
						Level:  10,
					},
					{
						ID:     "char-2",
						UserID: "user-123",
						Name:   "Legolas",
						Level:  8,
					},
				}
				m.On("GetAllCharacters", mock.Anything, "user-123").Return(characters, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var characters []*models.Character
				err := json.Unmarshal(body, &characters)
				require.NoError(t, err)
				assert.Len(t, characters, 2)
				assert.Equal(t, "Aragorn", characters[0].Name)
				assert.Equal(t, "Legolas", characters[1].Name)
			},
		},
		{
			name:   "empty character list",
			userID: "user-456",
			setupMock: func(m *MockCharacterService) {
				m.On("GetAllCharacters", mock.Anything, "user-456").Return([]*models.Character{}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var characters []*models.Character
				err := json.Unmarshal(body, &characters)
				require.NoError(t, err)
				assert.Empty(t, characters)
			},
		},
		{
			name:   "service error",
			userID: "user-789",
			setupMock: func(m *MockCharacterService) {
				m.On("GetAllCharacters", mock.Anything, "user-789").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "database error", response["error"])
			},
		},
		{
			name:           "unauthorized - no auth context",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "Unauthorized", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockCharService := createTestHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockCharService)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/characters", nil)
			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handlers.GetCharacters(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockCharService.AssertExpectations(t)
		})
	}
}

func TestHandlers_GetCharacter(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		userID         string
		setupMock      func(*MockCharacterService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful retrieval - owner",
			characterID: "char-123",
			userID:      "user-123",
			setupMock: func(m *MockCharacterService) {
				character := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
					Name:   "Aragorn",
					Level:  10,
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var character models.Character
				err := json.Unmarshal(body, &character)
				require.NoError(t, err)
				assert.Equal(t, "char-123", character.ID)
				assert.Equal(t, "Aragorn", character.Name)
			},
		},
		{
			name:        "forbidden - not owner",
			characterID: "char-123",
			userID:      "user-456",
			setupMock: func(m *MockCharacterService) {
				character := &models.Character{
					ID:     "char-123",
					UserID: "user-123", // Different user
					Name:   "Aragorn",
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil)
			},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "permission")
			},
		},
		{
			name:        "not found",
			characterID: "nonexistent",
			userID:      "user-123",
			setupMock: func(m *MockCharacterService) {
				m.On("GetCharacterByID", mock.Anything, "nonexistent").Return(nil, errors.New("character not found"))
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "character not found", response["error"])
			},
		},
		{
			name:           "unauthorized",
			characterID:    "char-123",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockCharService := createTestHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockCharService)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/characters/"+tt.characterID, nil)
			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			// Set up router vars
			req = mux.SetURLVars(req, map[string]string{
				"id": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handlers.GetCharacter(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockCharService.AssertExpectations(t)
		})
	}
}

func TestHandlers_CreateCharacter(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		setupMock      func(*MockCharacterService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:   "successful creation",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"name":  "Gandalf",
				"race":  "Human",
				"class": "Wizard",
				"level": 20,
			},
			setupMock: func(m *MockCharacterService) {
				m.On("CreateCharacter", mock.Anything, mock.MatchedBy(func(c *models.Character) bool {
					return c.Name == "Gandalf" && c.UserID == "user-123"
				})).Return(nil).Run(func(args mock.Arguments) {
					// Simulate setting ID on creation
					char := args.Get(1).(*models.Character)
					char.ID = "char-new"
				})
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body []byte) {
				var character models.Character
				err := json.Unmarshal(body, &character)
				require.NoError(t, err)
				assert.Equal(t, "char-new", character.ID)
				assert.Equal(t, "Gandalf", character.Name)
				assert.Equal(t, "user-123", character.UserID)
			},
		},
		{
			name:        "invalid request body",
			userID:      "user-123",
			requestBody: "invalid json",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name:   "service error",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"name": "Gandalf",
			},
			setupMock: func(m *MockCharacterService) {
				m.On("CreateCharacter", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "database error", response["error"])
			},
		},
		{
			name:           "unauthorized",
			requestBody:    map[string]interface{}{"name": "Gandalf"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockCharService := createTestHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockCharService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/characters", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handlers.CreateCharacter(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockCharService.AssertExpectations(t)
		})
	}
}

func TestHandlers_UpdateCharacter(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		userID         string
		requestBody    interface{}
		setupMock      func(*MockCharacterService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful update",
			characterID: "char-123",
			userID:      "user-123",
			requestBody: map[string]interface{}{
				"name":  "Gandalf the White",
				"level": 21,
			},
			setupMock: func(m *MockCharacterService) {
				// First call - check ownership
				existingChar := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
					Name:   "Gandalf",
					Level:  20,
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(existingChar, nil).Once()

				// Update character
				m.On("UpdateCharacter", mock.Anything, mock.MatchedBy(func(c *models.Character) bool {
					return c.ID == "char-123" && c.Name == "Gandalf the White" && c.UserID == "user-123"
				})).Return(nil)

				// Final retrieval
				updatedChar := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
					Name:   "Gandalf the White",
					Level:  21,
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(updatedChar, nil).Once()
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var character models.Character
				err := json.Unmarshal(body, &character)
				require.NoError(t, err)
				assert.Equal(t, "Gandalf the White", character.Name)
				assert.Equal(t, 21, character.Level)
			},
		},
		{
			name:        "forbidden - not owner",
			characterID: "char-123",
			userID:      "user-456",
			requestBody: map[string]interface{}{
				"name": "Gandalf the White",
			},
			setupMock: func(m *MockCharacterService) {
				existingChar := &models.Character{
					ID:     "char-123",
					UserID: "user-123", // Different owner
					Name:   "Gandalf",
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(existingChar, nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "character not found",
			characterID: "nonexistent",
			userID:      "user-123",
			requestBody: map[string]interface{}{
				"name": "Updated",
			},
			setupMock: func(m *MockCharacterService) {
				m.On("GetCharacterByID", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid request body",
			characterID:    "char-123",
			userID:         "user-123",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
			setupMock: func(m *MockCharacterService) {
				existingChar := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(existingChar, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockCharService := createTestHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockCharService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/characters/"+tt.characterID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			req = mux.SetURLVars(req, map[string]string{
				"id": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handlers.UpdateCharacter(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockCharService.AssertExpectations(t)
		})
	}
}

func TestHandlers_DeleteCharacter(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		userID         string
		setupMock      func(*MockCharacterService)
		expectedStatus int
	}{
		{
			name:        "successful deletion",
			characterID: "char-123",
			userID:      "user-123",
			setupMock: func(m *MockCharacterService) {
				character := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil)
				m.On("DeleteCharacter", mock.Anything, "char-123").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "forbidden - not owner",
			characterID: "char-123",
			userID:      "user-456",
			setupMock: func(m *MockCharacterService) {
				character := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "character not found",
			characterID: "nonexistent",
			userID:      "user-123",
			setupMock: func(m *MockCharacterService) {
				m.On("GetCharacterByID", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "delete error",
			characterID: "char-123",
			userID:      "user-123",
			setupMock: func(m *MockCharacterService) {
				character := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil)
				m.On("DeleteCharacter", mock.Anything, "char-123").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockCharService := createTestHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockCharService)
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/characters/"+tt.characterID, nil)
			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			req = mux.SetURLVars(req, map[string]string{
				"id": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handlers.DeleteCharacter(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockCharService.AssertExpectations(t)
		})
	}
}

func TestHandlers_AddExperience(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		userID         string
		requestBody    interface{}
		setupMock      func(*MockCharacterService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "successful xp addition",
			characterID: "char-123",
			userID:      "user-123",
			requestBody: map[string]interface{}{
				"experience": 500,
			},
			setupMock: func(m *MockCharacterService) {
				// Check ownership
				character := &models.Character{
					ID:               "char-123",
					UserID:           "user-123",
					Level:            5,
					ExperiencePoints: 6000,
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil).Once()

				// Add experience
				m.On("AddExperience", mock.Anything, "char-123", 500).Return(nil)

				// Get updated character
				updatedChar := &models.Character{
					ID:               "char-123",
					UserID:           "user-123",
					Level:            5,
					ExperiencePoints: 6500,
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(updatedChar, nil).Once()

				// Get XP for next level
				m.On("GetXPForNextLevel", 5).Return(14000)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				character := response["character"].(map[string]interface{})
				assert.Equal(t, float64(6500), character["experiencePoints"])
				assert.Equal(t, float64(14000), response["xpForNext"])
				assert.Equal(t, false, response["leveledUp"])
			},
		},
		{
			name:        "invalid experience amount",
			characterID: "char-123",
			userID:      "user-123",
			requestBody: map[string]interface{}{
				"experience": -100,
			},
			setupMock: func(m *MockCharacterService) {
				character := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil)
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "Experience must be positive", response["error"])
			},
		},
		{
			name:        "forbidden - not owner",
			characterID: "char-123",
			userID:      "user-456",
			requestBody: map[string]interface{}{
				"experience": 500,
			},
			setupMock: func(m *MockCharacterService) {
				character := &models.Character{
					ID:     "char-123",
					UserID: "user-123",
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockCharService := createTestHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockCharService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/characters/"+tt.characterID+"/experience", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := createAuthContext(tt.userID, "testuser", "test@example.com", "player")
				req = req.WithContext(ctx)
			}

			req = mux.SetURLVars(req, map[string]string{
				"id": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handlers.AddExperience(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockCharService.AssertExpectations(t)
		})
	}
}

func TestHandlers_Rest(t *testing.T) {
	tests := []struct {
		name           string
		characterID    string
		requestBody    interface{}
		setupMock      func(*MockCharacterService)
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:        "short rest",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"restType": "short",
			},
			setupMock: func(m *MockCharacterService) {
				m.On("RestoreSpellSlots", mock.Anything, "char-123", "short").Return(nil)

				// Get updated character
				character := &models.Character{
					ID:          "char-123",
					HitPoints:   50,
					MaxHitPoints: 100,
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var character models.Character
				err := json.Unmarshal(body, &character)
				require.NoError(t, err)
				assert.Equal(t, 50, character.HitPoints) // Not restored on short rest
			},
		},
		{
			name:        "long rest",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"restType": "long",
			},
			setupMock: func(m *MockCharacterService) {
				m.On("RestoreSpellSlots", mock.Anything, "char-123", "long").Return(nil)

				// Get character for HP restoration
				character := &models.Character{
					ID:           "char-123",
					HitPoints:    50,
					MaxHitPoints: 100,
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(character, nil).Once()

				// Update character with full HP
				m.On("UpdateCharacter", mock.Anything, mock.MatchedBy(func(c *models.Character) bool {
					return c.ID == "char-123" && c.HitPoints == 100
				})).Return(nil)

				// Final retrieval
				restoredChar := &models.Character{
					ID:           "char-123",
					HitPoints:    100,
					MaxHitPoints: 100,
				}
				m.On("GetCharacterByID", mock.Anything, "char-123").Return(restoredChar, nil).Once()
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var character models.Character
				err := json.Unmarshal(body, &character)
				require.NoError(t, err)
				assert.Equal(t, 100, character.HitPoints) // Fully restored on long rest
			},
		},
		{
			name:        "invalid rest type",
			characterID: "char-123",
			requestBody: map[string]interface{}{
				"restType": "invalid",
			},
			setupMock: func(m *MockCharacterService) {
				m.On("RestoreSpellSlots", mock.Anything, "char-123", "invalid").Return(errors.New("invalid rest type"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			characterID:    "char-123",
			requestBody:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers, mockCharService := createTestHandlers()
			if tt.setupMock != nil {
				tt.setupMock(mockCharService)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/characters/"+tt.characterID+"/rest", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			req = mux.SetURLVars(req, map[string]string{
				"id": tt.characterID,
			})

			rr := httptest.NewRecorder()
			handlers.Rest(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			mockCharService.AssertExpectations(t)
		})
	}
}
package handlers

import (
	"net/http"
	"testing"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCharacterService using test helpers
type MockCharacterService struct {
	testhelpers.MockService
}

func (m *MockCharacterService) CreateCharacter(userID string, char *models.Character) error {
	args := m.Called(userID, char)
	return args.Error(0)
}

func (m *MockCharacterService) GetCharacter(characterID string) (*models.Character, error) {
	args := m.Called(characterID)
	if char := args.Get(0); char != nil {
		return char.(*models.Character), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCharacterService) UpdateCharacter(char *models.Character) error {
	args := m.Called(char)
	return args.Error(0)
}

func (m *MockCharacterService) DeleteCharacter(characterID string) error {
	args := m.Called(characterID)
	return args.Error(0)
}

func (m *MockCharacterService) GetUserCharacters(userID string) ([]*models.Character, error) {
	args := m.Called(userID)
	if chars := args.Get(0); chars != nil {
		return chars.([]*models.Character), args.Error(1)
	}
	return nil, args.Error(1)
}

// TestCharacterHandler_RequestValidation using test helpers
func TestCharacterHandler_RequestValidation_Refactored(t *testing.T) {
	// Setup
	mockService := new(MockCharacterService)
	handler := &Handlers{characterService: mockService}
	
	// Create test user
	testUser := testhelpers.NewTestPlayer()
	
	// Define test cases using HTTPTestCase
	testCases := []testhelpers.HTTPTestCase{
		{
			Name:   "valid character creation request",
			Method: http.MethodPost,
			Path:   "/api/characters",
			Body: map[string]interface{}{
				"name":  "Aragorn",
				"race":  "Human",
				"class": "Ranger",
				"level": 10,
				"abilities": map[string]int{
					"strength":     16,
					"dexterity":    14,
					"constitution": 15,
					"intelligence": 12,
					"wisdom":       14,
					"charisma":     13,
				},
			},
			UserID:         testUser.ID,
			Role:           testUser.Role,
			ExpectedStatus: http.StatusCreated,
			ValidateBody: func(t *testing.T, body map[string]interface{}) {
				assert.NotEmpty(t, body["id"])
				assert.Equal(t, "Aragorn", body["name"])
				assert.Equal(t, "Human", body["race"])
				assert.Equal(t, "Ranger", body["class"])
			},
		},
		{
			Name:   "invalid character creation - missing name",
			Method: http.MethodPost,
			Path:   "/api/characters",
			Body: map[string]interface{}{
				"race":  "Elf",
				"class": "Wizard",
				"level": 1,
			},
			UserID:         testUser.ID,
			Role:           testUser.Role,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "name is required",
		},
		{
			Name:   "invalid character creation - invalid ability scores",
			Method: http.MethodPost,
			Path:   "/api/characters",
			Body: map[string]interface{}{
				"name":  "Invalid Character",
				"race":  "Human",
				"class": "Fighter",
				"level": 1,
				"abilities": map[string]int{
					"strength":     25, // Too high
					"dexterity":    14,
					"constitution": 15,
					"intelligence": 12,
					"wisdom":       14,
					"charisma":     13,
				},
			},
			UserID:         testUser.ID,
			Role:           testUser.Role,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "ability score",
		},
		{
			Name:           "no authentication",
			Method:         http.MethodPost,
			Path:           "/api/characters",
			Body:           map[string]interface{}{"name": "Test"},
			ExpectedStatus: http.StatusUnauthorized,
		},
	}
	
	// Setup mock expectations for valid creation
	mockService.On("CreateCharacter", testUser.ID, mock.AnythingOfType("*models.Character")).
		Return(nil).Maybe()
	
	// Run all test cases
	testhelpers.RunHTTPTestCases(t, testCases, handler.CreateCharacter)
	
	// Verify expectations
	mockService.AssertExpectations(t)
}

// TestCharacterHandler_GetCharacter using test helpers
func TestCharacterHandler_GetCharacter_Refactored(t *testing.T) {
	mockService := new(MockCharacterService)
	handler := &Handlers{characterService: mockService}
	
	// Test data
	testUser := testhelpers.NewTestPlayer()
	testChar := testhelpers.NewCharacterBuilder().
		WithUserID(testUser.ID).
		WithName("Gandalf").
		WithClass("Wizard").
		WithLevel(20).
		Build()
	
	testCases := []testhelpers.HTTPTestCase{
		{
			Name:           "get existing character",
			Method:         http.MethodGet,
			Path:           "/api/characters/" + testChar.ID,
			PathVars:       map[string]string{"id": testChar.ID},
			UserID:         testUser.ID,
			Role:           testUser.Role,
			ExpectedStatus: http.StatusOK,
			ValidateBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, testChar.ID, body["id"])
				assert.Equal(t, "Gandalf", body["name"])
				assert.Equal(t, "Wizard", body["class"])
				assert.Equal(t, float64(20), body["level"])
			},
		},
		{
			Name:           "get non-existent character",
			Method:         http.MethodGet,
			Path:           "/api/characters/" + uuid.New().String(),
			PathVars:       map[string]string{"id": uuid.New().String()},
			UserID:         testUser.ID,
			Role:           testUser.Role,
			ExpectedStatus: http.StatusNotFound,
			ExpectedError:  "character not found",
		},
		{
			Name:           "invalid character ID format",
			Method:         http.MethodGet,
			Path:           "/api/characters/invalid-id",
			PathVars:       map[string]string{"id": "invalid-id"},
			UserID:         testUser.ID,
			Role:           testUser.Role,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError:  "invalid character ID",
		},
	}
	
	// Setup mock expectations
	mockService.On("GetCharacter", testChar.ID).Return(testChar, nil).Once()
	mockService.On("GetCharacter", mock.AnythingOfType("string")).
		Return(nil, models.ErrCharacterNotFound).Maybe()
	
	// Run test cases
	testhelpers.RunHTTPTestCases(t, testCases, handler.GetCharacter)
	
	mockService.AssertExpectations(t)
}

// TestCharacterHandler_UpdateCharacter using test helpers and builders
func TestCharacterHandler_UpdateCharacter_Refactored(t *testing.T) {
	mockService := new(MockCharacterService)
	handler := &Handlers{characterService: mockService}
	
	// Test data using builders
	testUser := testhelpers.NewTestPlayer()
	testChar := testhelpers.NewCharacterBuilder().
		WithUserID(testUser.ID).
		WithName("Frodo").
		WithLevel(5).
		Build()
	
	otherUser := testhelpers.NewTestPlayer()
	otherChar := testhelpers.NewCharacterBuilder().
		WithUserID(otherUser.ID).
		Build()
	
	testCases := []testhelpers.HTTPTestCase{
		{
			Name:   "update own character",
			Method: http.MethodPut,
			Path:   "/api/characters/" + testChar.ID,
			PathVars: map[string]string{"id": testChar.ID},
			Body: map[string]interface{}{
				"id":    testChar.ID,
				"name":  "Frodo Baggins",
				"level": 6,
			},
			UserID:         testUser.ID,
			Role:           testUser.Role,
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "cannot update another user's character",
			Method: http.MethodPut,
			Path:   "/api/characters/" + otherChar.ID,
			PathVars: map[string]string{"id": otherChar.ID},
			Body: map[string]interface{}{
				"id":   otherChar.ID,
				"name": "Should Fail",
			},
			UserID:         testUser.ID,
			Role:           testUser.Role,
			ExpectedStatus: http.StatusForbidden,
			ExpectedError:  "forbidden",
		},
		{
			Name:   "DM can update any character",
			Method: http.MethodPut,
			Path:   "/api/characters/" + otherChar.ID,
			PathVars: map[string]string{"id": otherChar.ID},
			Body: map[string]interface{}{
				"id":   otherChar.ID,
				"name": "DM Update",
			},
			UserID:         uuid.New().String(),
			Role:           "dm",
			ExpectedStatus: http.StatusOK,
		},
	}
	
	// Setup mocks using helpers
	testhelpers.SetupMockCalls(&mockService.Mock, []testhelpers.MockCall{
		testhelpers.DataCall("GetCharacter", testChar, testChar.ID),
		testhelpers.DataCall("GetCharacter", otherChar, otherChar.ID),
		testhelpers.SuccessCall("UpdateCharacter", mock.AnythingOfType("*models.Character")),
	})
	
	testhelpers.RunHTTPTestCases(t, testCases, handler.UpdateCharacter)
	
	mockService.AssertExpectations(t)
}

// Example of testing with database mocks
func TestCharacterRepository_Create_Refactored(t *testing.T) {
	// Create test database
	testDB := testhelpers.NewTestDB(t)
	defer testDB.Close()
	
	// Create test character
	char := testhelpers.NewCharacterBuilder().
		WithName("Legolas").
		WithRace("Elf").
		WithClass("Ranger").
		Build()
	
	// Setup expectations
	testDB.SetupCreateSuccess("characters")
	
	// Create repository (would be your actual repository)
	// repo := database.NewCharacterRepository(testDB.DB)
	// err := repo.Create(char)
	
	// Assert expectations
	testDB.AssertExpectationsMet(t)
}

// Demonstration of assertion helpers
func TestCharacterValidation_Refactored(t *testing.T) {
	// Create various test characters
	validChar := testhelpers.NewCharacterBuilder().Build()
	invalidChar := &models.Character{
		Name:  "", // Invalid - empty name
		Level: 0,  // Invalid - level 0
		Stats: models.CharacterStats{
			Strength: 35, // Invalid - too high
		},
	}
	
	// Use assertion helpers
	testhelpers.AssertValidCharacter(t, validChar)
	
	// Test invalid character (would fail assertions)
	// testhelpers.AssertValidCharacter(t, invalidChar)
	
	// Test specific validations
	testhelpers.AssertValidStats(t, &validChar.Stats)
	
	// UUID validation
	testhelpers.AssertUUID(t, validChar.ID, "Character ID")
}

// Example showing reduction in code
// BEFORE: ~40 lines for request creation and validation
// AFTER: ~5 lines using HTTPTestCase
func ExampleCodeReduction(t *testing.T) {
	handler := &Handlers{}
	
	// BEFORE (40+ lines):
	/*
	body, _ := json.Marshal(map[string]interface{}{"name": "Test"})
	req := httptest.NewRequest(http.MethodPost, "/api/characters", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	claims := &auth.Claims{UserID: "123", Role: "player"}
	ctx := context.WithValue(req.Context(), auth.UserContextKey, claims)
	req = req.WithContext(ctx)
	
	w := httptest.NewRecorder()
	handler.CreateCharacter(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["id"])
	*/
	
	// AFTER (5 lines):
	testhelpers.ExecuteTestCase(t, testhelpers.HTTPTestCase{
		Name:           "create character",
		Method:         http.MethodPost,
		Path:           "/api/characters",
		Body:           map[string]interface{}{"name": "Test"},
		UserID:         "123",
		Role:           "player",
		ExpectedStatus: http.StatusCreated,
	}, handler.CreateCharacter)
}
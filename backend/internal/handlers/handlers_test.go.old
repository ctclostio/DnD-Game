package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/websocket"
)

// Mock services
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ValidatePassword(user *models.User, password string) bool {
	args := m.Called(user, password)
	return args.Bool(0)
}

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

func (m *MockCharacterService) CreateCharacter(ctx context.Context, char *models.Character) error {
	args := m.Called(ctx, char)
	return args.Error(0)
}

func (m *MockCharacterService) GetCharacter(ctx context.Context, id string) (*models.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterService) UpdateCharacter(ctx context.Context, char *models.Character) error {
	args := m.Called(ctx, char)
	return args.Error(0)
}

func (m *MockCharacterService) DeleteCharacter(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Test setup helper
func setupTestHandlers() (*Handlers, *MockUserService, *MockCharacterService) {
	mockUserService := new(MockUserService)
	mockCharService := new(MockCharacterService)
	
	svc := &services.Services{
		Users:      mockUserService,
		Characters: mockCharService,
		JWTManager: auth.NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour),
	}
	
	hub := websocket.GetHub()
	handlers := NewHandlers(svc, hub)
	
	return handlers, mockUserService, mockCharService
}

func TestHandlers_HealthCheck(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "dnd-game-backend", response["service"])
}

func TestHandlers_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		handlers, mockUserService, _ := setupTestHandlers()

		reqBody := map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		mockUserService.On("GetByUsername", mock.Anything, "testuser").Return(nil, nil)
		mockUserService.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
		mockUserService.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
			return u.Username == "testuser" && u.Email == "test@example.com"
		})).Return(nil)

		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["access_token"])
		assert.NotEmpty(t, response["refresh_token"])
		assert.Equal(t, "testuser", response["user"].(map[string]interface{})["username"])

		mockUserService.AssertExpectations(t)
	})

	t.Run("username already exists", func(t *testing.T) {
		handlers, mockUserService, _ := setupTestHandlers()

		reqBody := map[string]string{
			"username": "existinguser",
			"email":    "new@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		existingUser := &models.User{Username: "existinguser"}
		mockUserService.On("GetByUsername", mock.Anything, "existinguser").Return(existingUser, nil)

		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.Register(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]string
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "username already exists")

		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		handlers, _, _ := setupTestHandlers()

		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandlers_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		handlers, mockUserService, _ := setupTestHandlers()

		reqBody := map[string]string{
			"username": "testuser",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		user := &models.User{
			ID:       "user-123",
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
		}
		mockUserService.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
		mockUserService.On("ValidatePassword", user, "password123").Return(true)

		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["access_token"])
		assert.NotEmpty(t, response["refresh_token"])
		assert.Equal(t, "testuser", response["user"].(map[string]interface{})["username"])

		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		handlers, mockUserService, _ := setupTestHandlers()

		reqBody := map[string]string{
			"username": "testuser",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		user := &models.User{
			ID:       "user-123",
			Username: "testuser",
		}
		mockUserService.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
		mockUserService.On("ValidatePassword", user, "wrongpassword").Return(false)

		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid credentials")

		mockUserService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		handlers, mockUserService, _ := setupTestHandlers()

		reqBody := map[string]string{
			"username": "nonexistent",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		mockUserService.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, nil)

		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockUserService.AssertExpectations(t)
	})
}

func TestHandlers_GetCharacters(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		handlers, _, mockCharService := setupTestHandlers()

		userID := "user-123"
		characters := []*models.Character{
			{ID: "char-1", Name: "Thorin", UserID: userID},
			{ID: "char-2", Name: "Gandalf", UserID: userID},
		}

		mockCharService.On("GetAllCharacters", mock.Anything, userID).Return(characters, nil)

		req := httptest.NewRequest("GET", "/api/v1/characters", nil)
		ctx := context.WithValue(req.Context(), "claims", &auth.Claims{UserID: userID})
		req = req.WithContext(ctx)
		
		w := httptest.NewRecorder()

		handlers.GetCharacters(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*models.Character
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, "Thorin", response[0].Name)
		assert.Equal(t, "Gandalf", response[1].Name)

		mockCharService.AssertExpectations(t)
	})

	t.Run("no claims in context", func(t *testing.T) {
		handlers, _, _ := setupTestHandlers()

		req := httptest.NewRequest("GET", "/api/v1/characters", nil)
		w := httptest.NewRecorder()

		// This should panic, so we test for recovery
		assert.Panics(t, func() {
			handlers.GetCharacters(w, req)
		})
	})
}

func TestHandlers_CreateCharacter(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		handlers, _, mockCharService := setupTestHandlers()

		userID := "user-123"
		charData := &models.Character{
			Name:  "Aragorn",
			Race:  "Human",
			Class: "Ranger",
		}
		body, _ := json.Marshal(charData)

		mockCharService.On("CreateCharacter", mock.Anything, mock.MatchedBy(func(c *models.Character) bool {
			return c.Name == "Aragorn" && c.UserID == userID
		})).Return(nil)

		req := httptest.NewRequest("POST", "/api/v1/characters", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), "claims", &auth.Claims{UserID: userID})
		req = req.WithContext(ctx)
		
		w := httptest.NewRecorder()

		handlers.CreateCharacter(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["id"])
		assert.Equal(t, "Character created successfully", response["message"])

		mockCharService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		handlers, _, _ := setupTestHandlers()

		req := httptest.NewRequest("POST", "/api/v1/characters", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), "claims", &auth.Claims{UserID: "user-123"})
		req = req.WithContext(ctx)
		
		w := httptest.NewRecorder()

		handlers.CreateCharacter(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Helper function to add URL vars to request
func setURLVars(r *http.Request, vars map[string]string) *http.Request {
	return mux.SetURLVars(r, vars)
}
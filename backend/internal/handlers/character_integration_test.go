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
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/config"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/testutil"
	"github.com/your-username/dnd-game/backend/internal/websocket"
)

func setupIntegrationTest(t *testing.T) (*Handlers, *mux.Router, func()) {
	// Setup test database
	sqlxDB := testutil.SetupTestDB(t)
	
	// Wrap in database.DB type
	db := &database.DB{
		DB: sqlxDB,
	}

	// Create repositories
	repos := &database.Repositories{
		Users:        database.NewUserRepository(db),
		Characters:   database.NewCharacterRepository(db),
		GameSessions: database.NewGameSessionRepository(db),
		DiceRolls:    database.NewDiceRollRepository(db),
		NPCs:         database.NewNPCRepository(db),
		Inventory:    database.NewInventoryRepository(sqlxDB),
	}

	// Create services
	jwtManager := auth.NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour)
	
	// Create mock LLM provider for character service
	mockLLM := &testutil.MockLLMProvider{}
	
	svc := &services.Services{
		Users:         services.NewUserService(repos.Users, nil),
		Characters:    services.NewCharacterService(repos.Characters, repos.CustomClasses, mockLLM),
		GameSessions:  services.NewGameSessionService(repos.GameSessions, repos.Characters, repos.Users),
		DiceRolls:     services.NewDiceRollService(repos.DiceRolls),
		Combat:        services.NewCombatService(),
		NPCs:          services.NewNPCService(repos.NPCs, nil),
		Inventory:     services.NewInventoryService(repos.Inventory, repos.Characters),
		JWTManager:    jwtManager,
		Config:        &config.Config{},
	}

	// Create handlers
	hub := websocket.GetHub()
	handlers := NewHandlers(svc, hub)

	// Setup router
	router := mux.NewRouter()
	authMiddleware := auth.NewMiddleware(jwtManager)

	// Character routes
	router.HandleFunc("/api/v1/characters", authMiddleware.Authenticate(handlers.GetCharacters)).Methods("GET")
	router.HandleFunc("/api/v1/characters", authMiddleware.Authenticate(handlers.CreateCharacter)).Methods("POST")
	router.HandleFunc("/api/v1/characters/{id}", authMiddleware.Authenticate(handlers.GetCharacter)).Methods("GET")
	router.HandleFunc("/api/v1/characters/{id}", authMiddleware.Authenticate(handlers.UpdateCharacter)).Methods("PUT")
	router.HandleFunc("/api/v1/characters/{id}", authMiddleware.Authenticate(handlers.DeleteCharacter)).Methods("DELETE")

	// Inventory routes
	inventoryHandler := NewInventoryHandler(svc.Inventory)
	router.HandleFunc("/api/v1/characters/{characterId}/inventory", authMiddleware.Authenticate(inventoryHandler.GetCharacterInventory)).Methods("GET")
	router.HandleFunc("/api/v1/characters/{characterId}/inventory", authMiddleware.Authenticate(inventoryHandler.AddItemToInventory)).Methods("POST")
	router.HandleFunc("/api/v1/characters/{characterId}/inventory/{itemId}/equip", authMiddleware.Authenticate(inventoryHandler.EquipItem)).Methods("POST")

	cleanup := func() {
		testutil.CleanupDB(db)
	}

	return handlers, router, cleanup
}

func createAuthenticatedRequest(t *testing.T, method, url string, body interface{}, userID string, jwtManager *auth.JWTManager) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	// Generate token and add to request
	token, _, err := jwtManager.GenerateTokens(userID, "testuser", auth.RolePlayer)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	return req
}

func TestCharacterAPI_Integration(t *testing.T) {
	handlers, router, cleanup := setupIntegrationTest(t)
	defer cleanup()

	userID := "user-test-123"
	
	// Seed test user
	testutil.SeedTestUser(t, handlers.userService.(*services.UserService).repo.(*database.UserRepository).db, 
		userID, "testuser", "test@example.com", "player")

	t.Run("Create Character", func(t *testing.T) {
		charData := models.Character{
			Name:  "Aragorn",
			Race:  "Human",
			Class: "Ranger",
			Level: 1,
			Attributes: models.Attributes{
				Strength:     16,
				Dexterity:    14,
				Constitution: 13,
				Intelligence: 12,
				Wisdom:       15,
				Charisma:     10,
			},
		}

		req := createAuthenticatedRequest(t, "POST", "/api/v1/characters", charData, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["id"])
		assert.Equal(t, "Character created successfully", response["message"])

		// Verify character was created in database
		testutil.AssertRowExists(t, handlers.characterService.(*services.CharacterService).repo.(*database.CharacterRepository).db,
			"characters", "name", "Aragorn")
	})

	t.Run("Get Characters", func(t *testing.T) {
		// Create a second character
		testutil.SeedTestCharacter(t, handlers.characterService.(*services.CharacterService).repo.(*database.CharacterRepository).db,
			"char-2", userID, "Legolas")

		req := createAuthenticatedRequest(t, "GET", "/api/v1/characters", nil, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var characters []models.Character
		err := json.NewDecoder(w.Body).Decode(&characters)
		require.NoError(t, err)
		assert.Len(t, characters, 2)
		
		names := []string{characters[0].Name, characters[1].Name}
		assert.Contains(t, names, "Aragorn")
		assert.Contains(t, names, "Legolas")
	})

	t.Run("Get Single Character", func(t *testing.T) {
		// Get the character ID from database
		var charID string
		db := handlers.characterService.(*services.CharacterService).repo.(*database.CharacterRepository).db
		err := db.Get(&charID, "SELECT id FROM characters WHERE name = 'Aragorn'")
		require.NoError(t, err)

		req := createAuthenticatedRequest(t, "GET", "/api/v1/characters/"+charID, nil, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var character models.Character
		err = json.NewDecoder(w.Body).Decode(&character)
		require.NoError(t, err)
		assert.Equal(t, "Aragorn", character.Name)
		assert.Equal(t, "Human", character.Race)
		assert.Equal(t, "Ranger", character.Class)
	})

	t.Run("Update Character", func(t *testing.T) {
		// Get the character ID
		var charID string
		db := handlers.characterService.(*services.CharacterService).repo.(*database.CharacterRepository).db
		err := db.Get(&charID, "SELECT id FROM characters WHERE name = 'Aragorn'")
		require.NoError(t, err)

		updateData := models.Character{
			ID:    charID,
			Name:  "Strider",
			Level: 2,
		}

		req := createAuthenticatedRequest(t, "PUT", "/api/v1/characters/"+charID, updateData, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedName string
		err = db.Get(&updatedName, "SELECT name FROM characters WHERE id = $1", charID)
		require.NoError(t, err)
		assert.Equal(t, "Strider", updatedName)
	})

	t.Run("Delete Character", func(t *testing.T) {
		// Get Legolas's ID
		var charID string
		db := handlers.characterService.(*services.CharacterService).repo.(*database.CharacterRepository).db
		err := db.Get(&charID, "SELECT id FROM characters WHERE name = 'Legolas'")
		require.NoError(t, err)

		req := createAuthenticatedRequest(t, "DELETE", "/api/v1/characters/"+charID, nil, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		testutil.AssertRowNotExists(t, db, "characters", "id", charID)
	})

	t.Run("Cannot Access Other User's Characters", func(t *testing.T) {
		// Create another user and their character
		otherUserID := "user-other-456"
		testutil.SeedTestUser(t, handlers.userService.(*services.UserService).repo.(*database.UserRepository).db,
			otherUserID, "otheruser", "other@example.com", "player")
		testutil.SeedTestCharacter(t, handlers.characterService.(*services.CharacterService).repo.(*database.CharacterRepository).db,
			"char-other", otherUserID, "Gimli")

		// Try to get the other user's character
		req := createAuthenticatedRequest(t, "GET", "/api/v1/characters", nil, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var characters []models.Character
		err := json.NewDecoder(w.Body).Decode(&characters)
		require.NoError(t, err)
		
		// Should only see own character (Strider), not Gimli
		assert.Len(t, characters, 1)
		assert.Equal(t, "Strider", characters[0].Name)
	})
}

func TestInventoryAPI_Integration(t *testing.T) {
	handlers, router, cleanup := setupIntegrationTest(t)
	defer cleanup()

	userID := "user-test-123"
	charID := "char-test-123"
	itemID := "item-test-123"

	// Seed test data
	db := handlers.characterService.(*services.CharacterService).repo.(*database.CharacterRepository).db
	testutil.SeedTestUser(t, db, userID, "testuser", "test@example.com", "player")
	testutil.SeedTestCharacter(t, db, charID, userID, "Thorin")
	testutil.SeedTestItem(t, db, itemID, "Longsword", "weapon", 1500)

	t.Run("Add Item to Inventory", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"item_id":  itemID,
			"quantity": 1,
		}

		req := createAuthenticatedRequest(t, "POST", "/api/v1/characters/"+charID+"/inventory", reqBody, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify item was added
		testutil.AssertRowExists(t, db, "character_inventory", "character_id", charID)
	})

	t.Run("Get Character Inventory", func(t *testing.T) {
		req := createAuthenticatedRequest(t, "GET", "/api/v1/characters/"+charID+"/inventory", nil, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var inventory []models.InventoryItem
		err := json.NewDecoder(w.Body).Decode(&inventory)
		require.NoError(t, err)
		assert.Len(t, inventory, 1)
		assert.Equal(t, "Longsword", inventory[0].Item.Name)
	})

	t.Run("Equip Item", func(t *testing.T) {
		req := createAuthenticatedRequest(t, "POST", "/api/v1/characters/"+charID+"/inventory/"+itemID+"/equip", nil, userID, handlers.jwtManager)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify item is equipped
		var equipped bool
		err := db.Get(&equipped, "SELECT equipped FROM character_inventory WHERE character_id = $1 AND item_id = $2", charID, itemID)
		require.NoError(t, err)
		assert.True(t, equipped)
	})
}
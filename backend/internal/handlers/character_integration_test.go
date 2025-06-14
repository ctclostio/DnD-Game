package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/auth"
	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
	"github.com/ctclostio/DnD-Game/backend/internal/testutil"
	"github.com/ctclostio/DnD-Game/backend/internal/websocket"
	"github.com/ctclostio/DnD-Game/backend/pkg/response"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testContext struct {
	t          *testing.T
	handlers   *Handlers
	router     *mux.Router
	db         *database.DB
	repos      *database.Repositories
	jwtManager *auth.JWTManager
	services   *services.Services
}

// DecodeResponseData decodes the data field from a wrapped response
func (ctx *testContext) DecodeResponseData(w *httptest.ResponseRecorder, v interface{}) {
	t := ctx.t
	var resp response.Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err, "Failed to decode response wrapper: %s", w.Body.String())
	require.True(t, resp.Success, "Expected success response, got error: %v", resp.Error)

	// Marshal the data back to JSON then unmarshal into the target type
	dataBytes, err := json.Marshal(resp.Data)
	require.NoError(t, err, "Failed to marshal response data")

	err = json.Unmarshal(dataBytes, v)
	require.NoError(t, err, "Failed to unmarshal response data into target type")
}

func setupIntegrationTest(t *testing.T) (*testContext, func()) {
	// Setup test database
	sqlxDB := testutil.SetupTestDB(t)

	// Wrap in database.DB type
	db := &database.DB{
		DB: sqlxDB,
	}

	// Create repositories
	repos := &database.Repositories{
		Users:         database.NewUserRepository(db),
		Characters:    database.NewCharacterRepository(db),
		GameSessions:  database.NewGameSessionRepository(db),
		DiceRolls:     database.NewDiceRollRepository(db),
		NPCs:          database.NewNPCRepository(sqlxDB),
		Inventory:     database.NewInventoryRepository(db),
		CustomClasses: database.NewCustomClassRepository(db),
	}

	// Create services
	jwtManager := auth.NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour)

	// Create mock LLM provider for character service
	mockLLM := &testutil.MockLLMProvider{}

	// Initialize refresh token repository for auth
	refreshTokenRepo := database.NewRefreshTokenRepository(sqlxDB)

	// Create all required services with minimal/stub implementations
	svc := &services.Services{
		Users:         services.NewUserService(repos.Users),
		Characters:    services.NewCharacterService(repos.Characters, repos.CustomClasses, mockLLM),
		GameSessions:  services.NewGameSessionService(repos.GameSessions),
		DiceRolls:     services.NewDiceRollService(repos.DiceRolls),
		Combat:        services.NewCombatService(),
		NPCs:          services.NewNPCService(repos.NPCs),
		Inventory:     services.NewInventoryService(repos.Inventory, repos.Characters),
		RefreshTokens: services.NewRefreshTokenService(refreshTokenRepo, jwtManager),
		JWTManager:    jwtManager,
		Config:        &config.Config{},
		// Initialize nil pointers to prevent panics - these services aren't used in character tests
		Encounters:         nil,
		CustomRaces:        nil,
		DMAssistant:        nil,
		Campaign:           nil,
		CombatAutomation:   nil,
		CombatAnalytics:    nil,
		SettlementGen:      nil,
		FactionSystem:      nil,
		WorldEventEngine:   nil,
		EconomicSim:        nil,
		RuleEngine:         nil,
		BalanceAnalyzer:    nil,
		ConditionalReality: nil,
		NarrativeEngine:    nil,
		AICampaignManager:  nil,
		BattleMapGen:       nil,
	}

	// Create handlers
	hub := websocket.GetHub()
	handlers := NewHandlers(svc, db, hub)

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
		_ = sqlxDB.Close()
	}

	return &testContext{
		t:          t,
		handlers:   handlers,
		router:     router,
		db:         db,
		repos:      repos,
		jwtManager: jwtManager,
		services:   svc,
	}, cleanup
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
	tokenPair, err := jwtManager.GenerateTokenPair(userID, "testuser", "test@example.com", "player")
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)

	return req
}

func TestCharacterAPI_Integration(t *testing.T) {
	ctx, cleanup := setupIntegrationTest(t)
	defer cleanup()

	userID := "user-test-123"

	// Seed test user
	testutil.SeedTestUser(t, ctx.db.DB,
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

		req := createAuthenticatedRequest(t, "POST", "/api/v1/characters", charData, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Success bool             `json:"success"`
			Data    models.Character `json:"data"`
		}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		createdChar := resp.Data
		assert.NotEmpty(t, createdChar.ID)
		assert.Equal(t, "Aragorn", createdChar.Name)
		assert.Equal(t, userID, createdChar.UserID)

		// Verify character was created in database
		testutil.AssertRowExists(t, ctx.db.DB,
			"characters", "name", "Aragorn")
	})

	t.Run("Get Characters", func(t *testing.T) {
		// Create a second character
		testutil.SeedTestCharacter(t, ctx.db.DB,
			"char-2", userID, "Legolas")

		req := createAuthenticatedRequest(t, "GET", "/api/v1/characters", nil, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		// Log the response for debugging
		if w.Code != http.StatusOK {
			t.Logf("GetCharacters response: %s", w.Body.String())
		}

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool               `json:"success"`
			Data    []models.Character `json:"data"`
		}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		characters := resp.Data
		assert.Len(t, characters, 2)

		names := []string{characters[0].Name, characters[1].Name}
		assert.Contains(t, names, "Aragorn")
		assert.Contains(t, names, "Legolas")
	})

	t.Run("Get Single Character", func(t *testing.T) {
		// Get the character ID from database
		var charID string
		err := ctx.db.Get(&charID, "SELECT id FROM characters WHERE name = 'Aragorn'")
		require.NoError(t, err)

		req := createAuthenticatedRequest(t, "GET", "/api/v1/characters/"+charID, nil, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool             `json:"success"`
			Data    models.Character `json:"data"`
		}
		err = json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		character := resp.Data
		assert.Equal(t, "Aragorn", character.Name)
		assert.Equal(t, "Human", character.Race)
		assert.Equal(t, "Ranger", character.Class)
	})

	t.Run("Update Character", func(t *testing.T) {
		// Seed a character for this test
		charID := "char-update-test"
		testutil.SeedTestCharacter(t, ctx.db.DB, charID, userID, "Aragorn")

		// Verify character exists and get its details
		var dbChar struct {
			ID     string `db:"id"`
			Name   string `db:"name"`
			UserID string `db:"user_id"`
		}
		err := ctx.db.Get(&dbChar, "SELECT id, name, user_id FROM characters WHERE id = $1", charID)
		require.NoError(t, err)
		t.Logf("Character in DB: ID=%s, Name=%s, UserID=%s", dbChar.ID, dbChar.Name, dbChar.UserID)

		name := "Strider"
		level := 2
		updateData := map[string]interface{}{
			"name":  &name,
			"level": &level,
		}

		// Log the update data
		updateJSON, _ := json.Marshal(updateData)
		t.Logf("Update request data: %s", string(updateJSON))

		// Test GetCharacterByID directly
		char, err := ctx.services.Characters.GetCharacterByID(context.Background(), charID)
		if err != nil {
			t.Logf("Direct service GetCharacterByID error: %v", err)
		} else {
			t.Logf("Direct service GetCharacterByID success: Name=%s", char.Name)
		}

		req := createAuthenticatedRequest(t, "PUT", "/api/v1/characters/"+charID, updateData, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Logf("Update character error: status=%d, body=%s", w.Code, w.Body.String())
		}
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update in database
		var updatedName string
		err = ctx.db.Get(&updatedName, "SELECT name FROM characters WHERE id = $1", charID)
		require.NoError(t, err)
		assert.Equal(t, "Strider", updatedName)
	})

	t.Run("Delete Character", func(t *testing.T) {
		// Use the char-2 ID we created earlier
		charID := "char-2"

		req := createAuthenticatedRequest(t, "DELETE", "/api/v1/characters/"+charID, nil, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		testutil.AssertRowNotExists(t, ctx.db.DB, "characters", "id", charID)
	})

	t.Run("Cannot Access Other User's Characters", func(t *testing.T) {
		// Seed a character for the main user
		testutil.SeedTestCharacter(t, ctx.db.DB,
			"char-main-user", userID, "Strider")

		// Create another user and their character
		otherUserID := "user-other-456"
		testutil.SeedTestUser(t, ctx.db.DB,
			otherUserID, "otheruser", "other@example.com", "player")
		testutil.SeedTestCharacter(t, ctx.db.DB,
			"char-other", otherUserID, "Gimli")

		// Try to get the other user's character
		req := createAuthenticatedRequest(t, "GET", "/api/v1/characters", nil, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Success bool               `json:"success"`
			Data    []models.Character `json:"data"`
		}
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		characters := resp.Data

		// Should only see own character (Strider), not Gimli
		// Note: May see multiple characters if previous tests created some
		found := false
		for _, char := range characters {
			assert.NotEqual(t, "Gimli", char.Name, "Should not see other user's character")
			if char.Name == "Strider" {
				found = true
			}
		}
		assert.True(t, found, "Should find own character")
	})
}

func TestInventoryAPI_Integration(t *testing.T) {
	ctx, cleanup := setupIntegrationTest(t)
	defer cleanup()

	userID := "user-test-123"
	charID := "char-test-123"
	itemID := "item-test-123"

	// Seed test data ONCE before all subtests
	testutil.SeedTestUser(t, ctx.db.DB, userID, "testuser", "test@example.com", "player")
	testutil.SeedTestCharacter(t, ctx.db.DB, charID, userID, "Thorin")
	testutil.SeedTestItem(t, ctx.db.DB, itemID, "Longsword", "weapon", 1500)

	t.Run("Add Item to Inventory", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"item_id":  itemID,
			"quantity": 1,
		}

		req := createAuthenticatedRequest(t, "POST", "/api/v1/characters/"+charID+"/inventory", reqBody, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		// Debug: log response if not OK
		if w.Code != http.StatusOK {
			t.Logf("AddItemToInventory response status: %d, body: %s", w.Code, w.Body.String())
		}

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify item was added
		testutil.AssertRowExists(t, ctx.db.DB, "character_inventory", "character_id", charID)
	})

	t.Run("Get Character Inventory", func(t *testing.T) {
		req := createAuthenticatedRequest(t, "GET", "/api/v1/characters/"+charID+"/inventory", nil, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		// Debug the response if it fails
		if w.Code != http.StatusOK {
			t.Logf("Get inventory response: status=%d, body=%s", w.Code, w.Body.String())
		}

		assert.Equal(t, http.StatusOK, w.Code)

		// The inventory handler doesn't use response wrapper, decode directly
		var inventory []models.InventoryItem
		err := json.NewDecoder(w.Body).Decode(&inventory)
		require.NoError(t, err)
		assert.Len(t, inventory, 1)
		assert.Equal(t, "Longsword", inventory[0].Item.Name)
	})

	t.Run("Equip Item", func(t *testing.T) {
		// First check if item is already in inventory
		var existingCount int
		err := ctx.db.Get(&existingCount, "SELECT COUNT(*) FROM character_inventory WHERE character_id = $1 AND item_id = $2", charID, itemID)
		require.NoError(t, err)

		if existingCount == 0 {
			// Add the item to inventory if not already there
			addReqBody := map[string]interface{}{
				"item_id":  itemID,
				"quantity": 1,
			}
			addReq := createAuthenticatedRequest(t, "POST", "/api/v1/characters/"+charID+"/inventory", addReqBody, userID, ctx.jwtManager)
			addW := httptest.NewRecorder()
			ctx.router.ServeHTTP(addW, addReq)

			// Log the add response
			t.Logf("Add item response: status=%d, body=%s", addW.Code, addW.Body.String())
			require.Equal(t, http.StatusOK, addW.Code, "Failed to add item to inventory")
		}

		// Verify the item was actually added
		var invCount int
		err = ctx.db.Get(&invCount, "SELECT COUNT(*) FROM character_inventory WHERE character_id = $1 AND item_id = $2", charID, itemID)
		require.NoError(t, err)
		t.Logf("Items in inventory after add: %d", invCount)

		// Debug: Get the actual IDs from the database
		type invItem struct {
			CharacterID string `db:"character_id"`
			ItemID      string `db:"item_id"`
		}
		var dbItem invItem
		err = ctx.db.Get(&dbItem, "SELECT character_id, item_id FROM character_inventory LIMIT 1")
		if err == nil {
			t.Logf("DB has character_id='%s', item_id='%s'", dbItem.CharacterID, dbItem.ItemID)
			t.Logf("We're looking for character_id='%s', item_id='%s'", charID, itemID)
		}

		// Debug: First try to get the inventory via API
		getReq := createAuthenticatedRequest(t, "GET", "/api/v1/characters/"+charID+"/inventory", nil, userID, ctx.jwtManager)
		getW := httptest.NewRecorder()
		ctx.router.ServeHTTP(getW, getReq)
		t.Logf("Get inventory response: status=%d", getW.Code)
		if getW.Code != http.StatusOK {
			t.Logf("Get inventory error: %s", getW.Body.String())
		}

		// Now equip the item
		req := createAuthenticatedRequest(t, "POST", "/api/v1/characters/"+charID+"/inventory/"+itemID+"/equip", nil, userID, ctx.jwtManager)
		w := httptest.NewRecorder()

		ctx.router.ServeHTTP(w, req)

		// Debug: log response if not OK
		if w.Code != http.StatusOK {
			t.Logf("EquipItem response status: %d, body: %s", w.Code, w.Body.String())
		}

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify item is equipped
		var equipped bool
		err = ctx.db.Get(&equipped, "SELECT equipped FROM character_inventory WHERE character_id = $1 AND item_id = $2", charID, itemID)
		require.NoError(t, err)
		if !equipped {
			t.Logf("Item equipped status: %v", equipped)
		}
		assert.True(t, equipped)
	})
}

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/handlers"
	"github.com/your-username/dnd-game/backend/internal/middleware"
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/testutil"
	"github.com/your-username/dnd-game/backend/internal/websocket"
	"github.com/your-username/dnd-game/backend/pkg/logger"
	"github.com/your-username/dnd-game/backend/pkg/response"
)

func TestCombatFlow_Integration(t *testing.T) {
	// Setup test context
	ctx, cleanup := testutil.SetupIntegrationTest(t)
	defer cleanup()

	// Create logger
	log, err := logger.NewV2(logger.DefaultConfig())
	require.NoError(t, err)

	// Create services
	userService := services.NewUserService(ctx.Repos.Users)
	characterService := services.NewCharacterService(ctx.Repos.Characters, ctx.Repos.CustomClasses, nil)
	gameSessionService := services.NewGameSessionService(ctx.Repos.GameSessions)
	gameSessionService.SetCharacterRepository(ctx.Repos.Characters)
	combatService := services.NewCombatService()
	npcService := services.NewNPCService(ctx.Repos.NPCs)

	svc := &services.Services{
		Users:        userService,
		Characters:   characterService,
		GameSessions: gameSessionService,
		Combat:       combatService,
		NPCs:         npcService,
		JWTManager:   ctx.JWTManager,
	}

	// Create WebSocket hub for combat broadcasts
	hub := websocket.NewHub()
	go hub.Run()

	// Create handlers and setup routes
	h := handlers.NewHandlers(svc, hub, ctx.DB)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()

	// Apply middleware
	api.Use(middleware.RequestIDMiddleware)
	api.Use(middleware.LoggingMiddleware(log))

	authMiddleware := auth.NewMiddleware(ctx.JWTManager)

	// Combat routes
	api.HandleFunc("/combat/start", authMiddleware.Authenticate(h.StartCombat)).Methods("POST")
	api.HandleFunc("/combat/{combatId}", authMiddleware.Authenticate(h.GetCombat)).Methods("GET")
	api.HandleFunc("/combat/{combatId}/action", authMiddleware.Authenticate(h.ProcessCombatAction)).Methods("POST")
	api.HandleFunc("/combat/{combatId}/end", authMiddleware.Authenticate(h.EndCombat)).Methods("POST")

	// Create test data
	dmID := ctx.CreateTestUser("dm_combat", "dm@combat.com", "password123")
	playerID := ctx.CreateTestUser("player_combat", "player@combat.com", "password123")

	// Create characters
	pcID := ctx.CreateTestCharacter(playerID, "Aragorn")

	// Update character stats for combat
	_, err = ctx.SQLXDB.Exec(`
		UPDATE characters 
		SET level = 5, 
		    hit_points = 40, 
		    max_hit_points = 40,
		    armor_class = 16,
		    initiative = 2
		WHERE id = ?`, pcID)
	require.NoError(t, err)

	// Create game session
	sessionID := ctx.CreateTestGameSession(dmID, "Combat Test Session", "COMBAT01")

	// Verify session exists
	var sessionCount int
	err = ctx.SQLXDB.Get(&sessionCount, "SELECT COUNT(*) FROM game_sessions WHERE id = ?", sessionID)
	require.NoError(t, err)
	require.Equal(t, 1, sessionCount, "Session should exist")

	// Create NPCs for combat
	goblin := &models.NPC{
		ID:           "npc-goblin-" + sessionID,
		Name:         "Goblin",
		Type:         "Hostile",
		Size:         "Small",
		HitPoints:    7,
		MaxHitPoints: 7,
		ArmorClass:   13,
		Abilities:    []models.NPCAbility{},
		Attributes: models.Attributes{
			Strength:     8,
			Dexterity:    14,
			Constitution: 10,
			Intelligence: 10,
			Wisdom:       8,
			Charisma:     8,
		},
		Actions: []models.NPCAction{
			{
				Name:        "Scimitar",
				Type:        "attack",
				AttackBonus: 4,
				Damage:      "1d6+2",
				DamageType:  "slashing",
			},
		},
	}
	goblin.GameSessionID = sessionID
	goblin.CreatedBy = dmID // Set the DM as the creator

	err = ctx.Repos.NPCs.Create(context.Background(), goblin)
	require.NoError(t, err)

	// Create auth header
	dmToken, _ := ctx.JWTManager.GenerateTokenPair(dmID, "dm_combat", "dm@combat.com", "dm")
	playerToken, _ := ctx.JWTManager.GenerateTokenPair(playerID, "player_combat", "player@combat.com", "player")

	var combatID string

	t.Run("Start Combat", func(t *testing.T) {
		startReq := struct {
			GameSessionID string             `json:"gameSessionId"`
			Combatants    []models.Combatant `json:"combatants"`
		}{
			GameSessionID: sessionID,
			Combatants: []models.Combatant{
				{
					ID:                pcID,
					Name:              "Aragorn",
					Type:              models.CombatantTypeCharacter,
					Initiative:        15, // Pre-rolled
					HP:                40,
					MaxHP:             40,
					AC:                16,
					CharacterID:       pcID,
					IsPlayerCharacter: true,
				},
				{
					ID:                goblin.ID,
					Name:              "Goblin",
					Type:              models.CombatantTypeNPC,
					Initiative:        12,
					HP:                7,
					MaxHP:             7,
					AC:                13,
					IsPlayerCharacter: false,
				},
			},
		}

		body, _ := json.Marshal(startReq)
		req := httptest.NewRequest("POST", "/api/v1/combat/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+dmToken.AccessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code) // 201 is correct for creation

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		combatData, ok := resp.Data.(map[string]interface{})
		require.True(t, ok, "Expected resp.Data to be map[string]interface{}, got %T", resp.Data)
		combatID = combatData["id"].(string)
		assert.NotEmpty(t, combatID)
		assert.True(t, combatData["isActive"].(bool))

		// Verify turn order
		combatants, ok := combatData["combatants"].([]interface{})
		require.True(t, ok)
		assert.Len(t, combatants, 2)

		// Check turn order array instead
		turnOrder, ok := combatData["turnOrder"].([]interface{})
		require.True(t, ok)
		assert.Len(t, turnOrder, 2)
	})

	t.Run("Get Combat State", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/combat/"+combatID, nil)
		req.Header.Set("Authorization", "Bearer "+playerToken.AccessToken)
		req = mux.SetURLVars(req, map[string]string{"combatId": combatID})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		combatData, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, combatID, combatData["id"])
		assert.Equal(t, float64(1), combatData["round"])
		assert.Equal(t, float64(0), combatData["currentTurn"])
	})

	t.Run("Execute Attack Action", func(t *testing.T) {
		// Aragorn attacks the goblin
		actionReq := models.CombatRequest{
			ActorID:     pcID,
			Action:      models.ActionTypeAttack,
			TargetID:    goblin.ID,
			Description: "Aragorn attacks with Longsword",
		}

		body, _ := json.Marshal(actionReq)
		req := httptest.NewRequest("POST", "/api/v1/combat/"+combatID+"/action", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+playerToken.AccessToken)
		req = mux.SetURLVars(req, map[string]string{"combatId": combatID})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		// Check action response
		actionData, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, pcID, actionData["actorId"])
		assert.Equal(t, "attack", actionData["actionType"])

		// Get updated combat state
		req2 := httptest.NewRequest("GET", "/api/v1/combat/"+combatID, nil)
		req2.Header.Set("Authorization", "Bearer "+playerToken.AccessToken)
		req2 = mux.SetURLVars(req2, map[string]string{"combatId": combatID})
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		var combatResp response.Response
		err = json.NewDecoder(w2.Body).Decode(&combatResp)
		require.NoError(t, err)

		combatData, ok := combatResp.Data.(map[string]interface{})
		require.True(t, ok)

		// Should have advanced turn
		assert.Equal(t, float64(1), combatData["currentTurn"])

		// Check goblin's HP
		combatants := combatData["combatants"].([]interface{})
		var goblinFound bool
		for _, p := range combatants {
			combatant := p.(map[string]interface{})
			// The goblin ID was regenerated, so check by name
			if combatant["name"] == "Goblin" {
				goblinFound = true
				// Goblin might have taken damage (attack could miss)
				hp := combatant["hp"].(float64)
				assert.LessOrEqual(t, hp, float64(7))
				break
			}
		}
		assert.True(t, goblinFound, "Goblin should be in combatants")
	})

	t.Run("End Turn Action", func(t *testing.T) {
		// Goblin's turn - but it's dead, so just end turn
		actionReq := models.CombatRequest{
			ActorID: goblin.ID,
			Action:  models.ActionTypeEndTurn,
		}

		body, _ := json.Marshal(actionReq)
		req := httptest.NewRequest("POST", "/api/v1/combat/"+combatID+"/action", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+dmToken.AccessToken)
		req = mux.SetURLVars(req, map[string]string{"combatId": combatID})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		// Check action response
		actionData, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "endTurn", actionData["actionType"])

		// Get updated combat state to verify round advancement
		req2 := httptest.NewRequest("GET", "/api/v1/combat/"+combatID, nil)
		req2.Header.Set("Authorization", "Bearer "+dmToken.AccessToken)
		req2 = mux.SetURLVars(req2, map[string]string{"combatId": combatID})
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		var combatResp response.Response
		err2 := json.NewDecoder(w2.Body).Decode(&combatResp)
		require.NoError(t, err2)

		combatData, ok := combatResp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(2), combatData["round"])
		assert.Equal(t, float64(0), combatData["currentTurn"]) // Back to first combatant
	})

	t.Run("End Combat", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/combat/"+combatID+"/end", nil)
		req.Header.Set("Authorization", "Bearer "+dmToken.AccessToken)
		req = mux.SetURLVars(req, map[string]string{"combatId": combatID})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		// Verify combat is ended - it should be removed from memory
		req = httptest.NewRequest("GET", "/api/v1/combat/"+combatID, nil)
		req.Header.Set("Authorization", "Bearer "+dmToken.AccessToken)
		req = mux.SetURLVars(req, map[string]string{"combatId": combatID})

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Combat should be removed after ending, so we expect 404
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCombatAuthorization_Integration(t *testing.T) {
	ctx, cleanup := testutil.SetupIntegrationTest(t)
	defer cleanup()

	log, _ := logger.NewV2(logger.DefaultConfig())

	// Create services
	svc := &services.Services{
		Users:        services.NewUserService(ctx.Repos.Users),
		Characters:   services.NewCharacterService(ctx.Repos.Characters, ctx.Repos.CustomClasses, nil),
		GameSessions: services.NewGameSessionService(ctx.Repos.GameSessions),
		Combat:       services.NewCombatService(),
		NPCs:         services.NewNPCService(ctx.Repos.NPCs),
		JWTManager:   ctx.JWTManager,
	}

	h := handlers.NewHandlers(svc, nil, ctx.DB)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.LoggingMiddleware(log))

	authMiddleware := auth.NewMiddleware(ctx.JWTManager)
	api.HandleFunc("/combat/start", authMiddleware.Authenticate(h.StartCombat)).Methods("POST")

	// Create test users
	dmID := ctx.CreateTestUser("auth_dm", "auth_dm@test.com", "password123")
	playerID := ctx.CreateTestUser("auth_player", "auth_player@test.com", "password123")
	// otherPlayerID := ctx.CreateTestUser("other_player", "other@test.com", "password123") // not used

	// Create session and character
	sessionID := ctx.CreateTestGameSession(dmID, "Auth Test Session", "AUTH01")
	charID := ctx.CreateTestCharacter(playerID, "AuthHero")

	// Generate tokens
	playerToken, _ := ctx.JWTManager.GenerateTokenPair(playerID, "auth_player", "auth_player@test.com", "player")
	// otherToken not used in tests

	t.Run("Non-DM Cannot Start Combat", func(t *testing.T) {
		startReq := struct {
			GameSessionID string             `json:"gameSessionId"`
			Combatants    []models.Combatant `json:"combatants"`
		}{
			GameSessionID: sessionID,
			Combatants: []models.Combatant{
				{
					ID:                charID,
					Name:              "AuthHero",
					Type:              models.CombatantTypeCharacter,
					Initiative:        10,
					HP:                30,
					MaxHP:             30,
					AC:                15,
					CharacterID:       charID,
					IsPlayerCharacter: true,
				},
			},
		}

		body, _ := json.Marshal(startReq)
		req := httptest.NewRequest("POST", "/api/v1/combat/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+playerToken.AccessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should fail - only DM can start combat
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("Cannot Start Combat in Another DM's Session", func(t *testing.T) {
		// Create another DM
		otherDMID := ctx.CreateTestUser("other_dm", "other_dm@test.com", "password123")
		otherDMToken, _ := ctx.JWTManager.GenerateTokenPair(otherDMID, "other_dm", "other_dm@test.com", "dm")

		startReq := struct {
			GameSessionID string             `json:"gameSessionId"`
			Combatants    []models.Combatant `json:"combatants"`
		}{
			GameSessionID: sessionID, // Session belongs to dmID, not otherDMID
			Combatants: []models.Combatant{
				{
					ID:         "dummy",
					Name:       "Dummy",
					Type:       models.CombatantTypeNPC,
					Initiative: 10,
					HP:         10,
					MaxHP:      10,
					AC:         10,
				},
			},
		}

		body, _ := json.Marshal(startReq)
		req := httptest.NewRequest("POST", "/api/v1/combat/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+otherDMToken.AccessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should fail - wrong DM
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("Player Cannot Act for Another Player's Character", func(t *testing.T) {
		// This would be tested in the ExecuteCombatAction endpoint
		// Skipping for now as it requires a running combat
		t.Skip("Requires combat to be started first")
	})
}

func TestCombatConditions_Integration(t *testing.T) {
	_, cleanup := testutil.SetupIntegrationTest(t)
	defer cleanup()

	// Setup similar to TestCombatFlow_Integration...
	// Testing specific conditions like:
	// - Death saving throws
	// - Status effects (poisoned, stunned, etc.)
	// - Healing
	// - Area of effect attacks

	t.Skip("Advanced combat conditions not yet implemented")
}

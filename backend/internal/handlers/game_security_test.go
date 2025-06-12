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
	"github.com/your-username/dnd-game/backend/internal/models"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/testutil"
)

func TestGameSessionSecurity(t *testing.T) {
	// Setup test context
	ctx := context.Background()
	testCtx, cleanup := testutil.SetupIntegrationTest(t)
	defer cleanup()

	// Create logger (not used in these tests)
	// log, err := logger.NewV2(logger.DefaultConfig())
	// require.NoError(t, err)

	// Create services with repositories
	gameService := services.NewGameSessionService(testCtx.Repos.GameSessions)
	gameService.SetCharacterRepository(testCtx.Repos.Characters)
	gameService.SetUserRepository(testCtx.Repos.Users)

	svc := &services.Services{
		Users:        services.NewUserService(testCtx.Repos.Users),
		Characters:   services.NewCharacterService(testCtx.Repos.Characters, nil, nil),
		GameSessions: gameService,
		JWTManager:   testCtx.JWTManager,
	}

	// Create handlers
	h := handlers.NewHandlers(svc, nil)

	// Create test users
	dm := createTestUser(t, testCtx, "dm_user", "dm@example.com", "dm")
	player1 := createTestUser(t, testCtx, "player1", "player1@example.com", "player")
	player2 := createTestUser(t, testCtx, "player2", "player2@example.com", "player")

	// Create test characters
	char1 := createTestCharacter(t, testCtx, player1.ID, "Fighter", 5)
	char2 := createTestCharacter(t, testCtx, player2.ID, "Wizard", 3)
	charHighLevel := createTestCharacter(t, testCtx, player1.ID, "Paladin", 10)

	// Create test session
	session := &models.GameSession{
		DMID:                  dm.ID,
		Name:                  "Security Test Session",
		Description:           "Testing security features",
		MaxPlayers:            4,
		IsPublic:              false,
		RequiresInvite:        true,
		AllowedCharacterLevel: 5,
	}
	err := svc.GameSessions.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Run("JoinSession_Security", func(t *testing.T) {
		tests := []struct {
			name           string
			userID         string
			characterID    string
			setupFunc      func()
			expectedStatus int
			expectedError  string
		}{
			{
				name:           "Valid join with character",
				userID:         player1.ID,
				characterID:    char1.ID,
				expectedStatus: http.StatusOK,
			},
			{
				name:           "Cannot join twice",
				userID:         player1.ID,
				characterID:    char1.ID,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "already in this session",
			},
			{
				name:           "Cannot join with another user's character",
				userID:         player2.ID,
				characterID:    char1.ID, // Belongs to player1
				expectedStatus: http.StatusBadRequest,
				expectedError:  "don't own this character",
			},
			{
				name:        "Cannot join with high level character",
				userID:      player1.ID,
				characterID: charHighLevel.ID,
				setupFunc: func() {
					_ = svc.GameSessions.LeaveSession(ctx, session.ID, player1.ID)
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "exceeds session limit",
			},
			{
				name:           "Can join without character",
				userID:         player2.ID,
				characterID:    "",
				expectedStatus: http.StatusOK,
			},
			{
				name:   "Cannot join full session",
				userID: "another_user",
				setupFunc: func() {
					// Fill the session to capacity
					session.MaxPlayers = 2 // DM + 1 player
					testCtx.Repos.GameSessions.Update(ctx, session)
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "session is full",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.setupFunc != nil {
					tt.setupFunc()
				}

				// Create request
				body := map[string]interface{}{}
				if tt.characterID != "" {
					body["character_id"] = tt.characterID
				}
				jsonBody, _ := json.Marshal(body)

				req := httptest.NewRequest("POST", "/api/v1/game/sessions/"+session.ID+"/join", bytes.NewBuffer(jsonBody))
				// Create user claims and add to context
				claims := &auth.Claims{
					UserID:   tt.userID,
					Username: "testuser",
					Email:    "test@example.com",
					Role:     "player",
				}
				req = req.WithContext(context.WithValue(req.Context(), auth.UserContextKey, claims))
				req = mux.SetURLVars(req, map[string]string{"id": session.ID})

				// Execute request
				rr := httptest.NewRecorder()
				h.JoinGameSession(rr, req)

				// Check response
				assert.Equal(t, tt.expectedStatus, rr.Code)
				if tt.expectedError != "" {
					var response map[string]interface{}
					err := json.NewDecoder(rr.Body).Decode(&response)
					require.NoError(t, err)

					errMap, ok := response["error"].(map[string]interface{})
					require.True(t, ok)
					assert.Contains(t, errMap["message"], tt.expectedError)
				}
			})
		}
	})

	t.Run("GetGameSession_Authorization", func(t *testing.T) {
		// Ensure player1 can join the session
		_ = svc.GameSessions.LeaveSession(ctx, session.ID, player1.ID)
		session.MaxPlayers = 4
		testCtx.Repos.GameSessions.Update(ctx, session)
		_ = svc.GameSessions.JoinSession(ctx, session.ID, player1.ID, &char1.ID)
		// Create private session
		privateSession := &models.GameSession{
			DMID:        dm.ID,
			Name:        "Private Session",
			Description: "Should not be visible to non-participants",
			IsPublic:    false,
		}
		err := svc.GameSessions.CreateSession(ctx, privateSession)
		require.NoError(t, err)

		tests := []struct {
			name           string
			userID         string
			sessionID      string
			expectedStatus int
		}{
			{
				name:           "DM can view their session",
				userID:         dm.ID,
				sessionID:      privateSession.ID,
				expectedStatus: http.StatusOK,
			},
			{
				name:           "Non-participant cannot view private session",
				userID:         player1.ID,
				sessionID:      privateSession.ID,
				expectedStatus: http.StatusForbidden,
			},
			{
				name:           "Participant can view session",
				userID:         player1.ID,
				sessionID:      session.ID, // player1 joined this session
				expectedStatus: http.StatusOK,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/api/v1/game/sessions/"+tt.sessionID, nil)
				// Create user claims and add to context
				claims := &auth.Claims{
					UserID:   tt.userID,
					Username: "testuser",
					Email:    "test@example.com",
					Role:     "player",
				}
				req = req.WithContext(context.WithValue(req.Context(), auth.UserContextKey, claims))
				req = mux.SetURLVars(req, map[string]string{"id": tt.sessionID})

				rr := httptest.NewRecorder()
				h.GetGameSession(rr, req)

				assert.Equal(t, tt.expectedStatus, rr.Code)
			})
		}
	})

	t.Run("KickPlayer_Security", func(t *testing.T) {
		// Ensure player2 is in the session for kick testing
		_ = svc.GameSessions.LeaveSession(ctx, session.ID, player2.ID)
		err := svc.GameSessions.JoinSession(ctx, session.ID, player2.ID, &char2.ID)
		require.NoError(t, err)

		tests := []struct {
			name           string
			dmUserID       string
			playerToKick   string
			expectedStatus int
			expectedError  string
		}{
			{
				name:           "DM can kick player",
				dmUserID:       dm.ID,
				playerToKick:   player2.ID,
				expectedStatus: http.StatusOK,
			},
			{
				name:           "Non-DM cannot kick player",
				dmUserID:       player1.ID,
				playerToKick:   player2.ID,
				expectedStatus: http.StatusForbidden,
				expectedError:  "Only the DM can kick",
			},
			{
				name:           "DM cannot kick themselves",
				dmUserID:       dm.ID,
				playerToKick:   dm.ID,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "cannot kick themselves",
			},
			{
				name:           "Cannot kick non-existent player",
				dmUserID:       dm.ID,
				playerToKick:   "non_existent_id",
				expectedStatus: http.StatusBadRequest,
				expectedError:  "not in this session",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/api/v1/game/sessions/"+session.ID+"/kick/"+tt.playerToKick, nil)
				claims := &auth.Claims{
					UserID:   tt.dmUserID,
					Username: "test",
					Email:    "test@example.com",
					Role:     "dm",
				}
				req = req.WithContext(context.WithValue(req.Context(), auth.UserContextKey, claims))
				req = mux.SetURLVars(req, map[string]string{
					"id":       session.ID,
					"playerId": tt.playerToKick,
				})

				rr := httptest.NewRecorder()
				h.KickPlayer(rr, req)

				assert.Equal(t, tt.expectedStatus, rr.Code)
				if tt.expectedError != "" {
					var response map[string]interface{}
					err := json.NewDecoder(rr.Body).Decode(&response)
					require.NoError(t, err)

					errMap, ok := response["error"].(map[string]interface{})
					require.True(t, ok)
					assert.Contains(t, errMap["message"], tt.expectedError)
				}
			})
		}
	})

	t.Run("SessionState_Security", func(t *testing.T) {
		// Test operations on inactive session
		inactiveSession := &models.GameSession{
			DMID:     dm.ID,
			Name:     "Inactive Session",
			Status:   models.GameStatusActive,
			IsActive: false,
		}
		err := svc.GameSessions.CreateSession(ctx, inactiveSession)
		require.NoError(t, err)

		// Try to join inactive session
		req := httptest.NewRequest("POST", "/api/v1/game/sessions/"+inactiveSession.ID+"/join", bytes.NewBufferString("{}"))
		// Create user claims and add to context
		claims := &auth.Claims{
			UserID:   player1.ID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
		}
		req = req.WithContext(context.WithValue(req.Context(), auth.UserContextKey, claims))
		req = mux.SetURLVars(req, map[string]string{"id": inactiveSession.ID})

		rr := httptest.NewRecorder()
		h.JoinGameSession(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var response map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&response)
		errMap, ok := response["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, errMap["message"], "not active")

		// Test operations on completed session
		completedSession := &models.GameSession{
			DMID:   dm.ID,
			Name:   "Completed Session",
			Status: models.GameStatusCompleted,
		}
		err = svc.GameSessions.CreateSession(ctx, completedSession)
		require.NoError(t, err)

		// Try to join completed session
		req = httptest.NewRequest("POST", "/api/v1/game/sessions/"+completedSession.ID+"/join", bytes.NewBufferString("{}"))
		// Create user claims and add to context
		claims2 := &auth.Claims{
			UserID:   player1.ID,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "player",
		}
		req = req.WithContext(context.WithValue(req.Context(), auth.UserContextKey, claims2))
		req = mux.SetURLVars(req, map[string]string{"id": completedSession.ID})

		rr = httptest.NewRecorder()
		h.JoinGameSession(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		json.NewDecoder(rr.Body).Decode(&response)
		errMap, ok = response["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, errMap["message"], "completed session")
	})
}

// Helper functions
func createTestUser(t *testing.T, ctx *testutil.IntegrationTestContext, username, email, role string) *models.User {
	user := &models.User{
		Username: username,
		Email:    email,
		Role:     role,
	}
	err := ctx.Repos.Users.Create(context.Background(), user)
	require.NoError(t, err)
	return user
}

func createTestCharacter(t *testing.T, ctx *testutil.IntegrationTestContext, userID, name string, level int) *models.Character {
	char := &models.Character{
		UserID: userID,
		Name:   name,
		Level:  level,
		Class:  "Fighter",
		Race:   "Human",
	}
	err := ctx.Repos.Characters.Create(context.Background(), char)
	require.NoError(t, err)
	return char
}

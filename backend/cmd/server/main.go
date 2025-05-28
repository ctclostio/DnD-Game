package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/your-username/dnd-game/backend/internal/auth"
	"github.com/your-username/dnd-game/backend/internal/config"
	"github.com/your-username/dnd-game/backend/internal/database"
	"github.com/your-username/dnd-game/backend/internal/handlers"
	"github.com/your-username/dnd-game/backend/internal/services"
	"github.com/your-username/dnd-game/backend/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize database
	db, repos, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTokenDuration,
		cfg.Auth.RefreshTokenDuration,
	)

	// Create refresh token repository
	refreshTokenRepo := database.NewRefreshTokenRepository(db.DB)

	// Create services
	refreshTokenService := services.NewRefreshTokenService(refreshTokenRepo, jwtManager)
	
	// Start refresh token cleanup task
	refreshTokenService.StartCleanupTask(1 * time.Hour)

	svc := &services.Services{
		Users:         services.NewUserService(repos.Users),
		Characters:    services.NewCharacterService(repos.Characters),
		GameSessions:  services.NewGameSessionService(repos.GameSessions),
		DiceRolls:     services.NewDiceRollService(repos.DiceRolls),
		Combat:        services.NewCombatService(),
		NPCs:          services.NewNPCService(repos.NPCs),
		Inventory:     services.NewInventoryService(repos.Inventory, repos.Characters),
		JWTManager:    jwtManager,
		RefreshTokens: refreshTokenService,
		Config:        cfg,
	}

	// Get websocket hub
	wsHub := websocket.GetHub()
	
	// Create handlers with services
	h := handlers.NewHandlers(svc, wsHub)
	
	// Create character creation handler
	charCreationHandler := handlers.NewCharacterCreationHandler(svc.Characters)
	
	// Create inventory handler
	inventoryHandler := handlers.NewInventoryHandler(svc.Inventory)

	// Create authentication middleware
	authMiddleware := auth.NewMiddleware(jwtManager)

	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", h.HealthCheck).Methods("GET")
	
	// Auth routes (public)
	api.HandleFunc("/auth/register", h.Register).Methods("POST")
	api.HandleFunc("/auth/login", h.Login).Methods("POST")
	api.HandleFunc("/auth/refresh", h.RefreshToken).Methods("POST")
	api.HandleFunc("/auth/logout", authMiddleware.Authenticate(h.Logout)).Methods("POST")
	api.HandleFunc("/auth/me", authMiddleware.Authenticate(h.GetCurrentUser)).Methods("GET")
	
	// Character creation routes (protected)
	api.HandleFunc("/characters/options", authMiddleware.Authenticate(charCreationHandler.GetCharacterOptions)).Methods("GET")
	api.HandleFunc("/characters/create", authMiddleware.Authenticate(charCreationHandler.CreateCharacter)).Methods("POST")
	api.HandleFunc("/characters/create-custom", authMiddleware.Authenticate(charCreationHandler.CreateCustomCharacter)).Methods("POST")
	api.HandleFunc("/characters/validate", authMiddleware.Authenticate(charCreationHandler.ValidateCharacter)).Methods("POST")
	api.HandleFunc("/characters/roll-abilities", authMiddleware.Authenticate(charCreationHandler.RollAbilityScores)).Methods("POST")
	
	// Character routes (protected)
	api.HandleFunc("/characters", authMiddleware.Authenticate(h.GetCharacters)).Methods("GET")
	api.HandleFunc("/characters", authMiddleware.Authenticate(h.CreateCharacter)).Methods("POST")
	api.HandleFunc("/characters/{id}", authMiddleware.Authenticate(h.GetCharacter)).Methods("GET")
	api.HandleFunc("/characters/{id}", authMiddleware.Authenticate(h.UpdateCharacter)).Methods("PUT")
	api.HandleFunc("/characters/{id}", authMiddleware.Authenticate(h.DeleteCharacter)).Methods("DELETE")
	api.HandleFunc("/characters/{id}/cast-spell", authMiddleware.Authenticate(h.CastSpell)).Methods("POST")
	api.HandleFunc("/characters/{id}/rest", authMiddleware.Authenticate(h.Rest)).Methods("POST")
	api.HandleFunc("/characters/{id}/add-experience", authMiddleware.Authenticate(h.AddExperience)).Methods("POST")
	
	// Dice roll routes (protected)
	api.HandleFunc("/dice/roll", authMiddleware.Authenticate(h.RollDice)).Methods("POST")
	
	// Game session routes (protected)
	api.HandleFunc("/game/sessions", authMiddleware.RequireDM()(h.CreateGameSession)).Methods("POST")
	api.HandleFunc("/game/sessions/{id}", authMiddleware.Authenticate(h.GetGameSession)).Methods("GET")
	api.HandleFunc("/game/sessions/{id}", authMiddleware.RequireDM()(h.UpdateGameSession)).Methods("PUT")
	api.HandleFunc("/game/sessions/{id}/join", authMiddleware.Authenticate(h.JoinGameSession)).Methods("POST")
	api.HandleFunc("/game/sessions/{id}/leave", authMiddleware.Authenticate(h.LeaveGameSession)).Methods("POST")

	// Combat routes (protected)
	api.HandleFunc("/combat/start", authMiddleware.Authenticate(h.StartCombat)).Methods("POST")
	api.HandleFunc("/combat/{id}", authMiddleware.Authenticate(h.GetCombat)).Methods("GET")
	api.HandleFunc("/combat/session/{sessionId}", authMiddleware.Authenticate(h.GetCombatBySession)).Methods("GET")
	api.HandleFunc("/combat/{id}/next-turn", authMiddleware.Authenticate(h.NextTurn)).Methods("POST")
	api.HandleFunc("/combat/{id}/action", authMiddleware.Authenticate(h.ProcessCombatAction)).Methods("POST")
	api.HandleFunc("/combat/{id}/end", authMiddleware.Authenticate(h.EndCombat)).Methods("POST")
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/save", authMiddleware.Authenticate(h.MakeSavingThrow)).Methods("POST")
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/damage", authMiddleware.Authenticate(h.ApplyDamage)).Methods("POST")
	api.HandleFunc("/combat/{id}/combatants/{combatantId}/heal", authMiddleware.Authenticate(h.HealCombatant)).Methods("POST")

	// NPC routes (protected)
	api.HandleFunc("/npcs", authMiddleware.RequireDM()(h.CreateNPC)).Methods("POST")
	api.HandleFunc("/npcs/{id}", authMiddleware.Authenticate(h.GetNPC)).Methods("GET")
	api.HandleFunc("/npcs/{id}", authMiddleware.RequireDM()(h.UpdateNPC)).Methods("PUT")
	api.HandleFunc("/npcs/{id}", authMiddleware.RequireDM()(h.DeleteNPC)).Methods("DELETE")
	api.HandleFunc("/npcs/session/{sessionId}", authMiddleware.Authenticate(h.GetNPCsBySession)).Methods("GET")
	api.HandleFunc("/npcs/search", authMiddleware.Authenticate(h.SearchNPCs)).Methods("GET")
	api.HandleFunc("/npcs/templates", authMiddleware.Authenticate(h.GetNPCTemplates)).Methods("GET")
	api.HandleFunc("/npcs/create-from-template", authMiddleware.RequireDM()(h.CreateNPCFromTemplate)).Methods("POST")
	api.HandleFunc("/npcs/{id}/action/{action}", authMiddleware.RequireDM()(h.NPCQuickActions)).Methods("POST")
	
	// Skill check routes (protected)
	api.HandleFunc("/skill-check", authMiddleware.Authenticate(h.PerformSkillCheck)).Methods("POST")
	api.HandleFunc("/characters/{id}/checks", authMiddleware.Authenticate(h.GetCharacterChecks)).Methods("GET")
	
	// Inventory routes (protected)
	api.HandleFunc("/characters/{characterId}/inventory", authMiddleware.Authenticate(inventoryHandler.GetCharacterInventory)).Methods("GET")
	api.HandleFunc("/characters/{characterId}/inventory", authMiddleware.Authenticate(inventoryHandler.AddItemToInventory)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/remove", authMiddleware.Authenticate(inventoryHandler.RemoveItemFromInventory)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/equip", authMiddleware.Authenticate(inventoryHandler.EquipItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/unequip", authMiddleware.Authenticate(inventoryHandler.UnequipItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/attune", authMiddleware.Authenticate(inventoryHandler.AttuneItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/{itemId}/unattune", authMiddleware.Authenticate(inventoryHandler.UnattuneItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/currency", authMiddleware.Authenticate(inventoryHandler.GetCharacterCurrency)).Methods("GET")
	api.HandleFunc("/characters/{characterId}/currency", authMiddleware.Authenticate(inventoryHandler.UpdateCharacterCurrency)).Methods("PUT")
	api.HandleFunc("/characters/{characterId}/inventory/purchase", authMiddleware.Authenticate(inventoryHandler.PurchaseItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/inventory/sell", authMiddleware.Authenticate(inventoryHandler.SellItem)).Methods("POST")
	api.HandleFunc("/characters/{characterId}/weight", authMiddleware.Authenticate(inventoryHandler.GetCharacterWeight)).Methods("GET")
	api.HandleFunc("/items", authMiddleware.RequireDM()(inventoryHandler.CreateItem)).Methods("POST")
	api.HandleFunc("/items", authMiddleware.Authenticate(inventoryHandler.GetItemsByType)).Methods("GET")

	// Initialize WebSocket with JWT manager
	websocket.SetJWTManager(jwtManager)
	
	// WebSocket endpoint
	router.HandleFunc("/ws", websocket.HandleWebSocket)

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/build/")))

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Authorization"},
	})

	handler := c.Handler(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
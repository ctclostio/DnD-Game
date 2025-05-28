package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/your-username/dnd-game/backend/internal/handlers"
	"github.com/your-username/dnd-game/backend/internal/websocket"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	api.HandleFunc("/characters", handlers.GetCharacters).Methods("GET")
	api.HandleFunc("/characters", handlers.CreateCharacter).Methods("POST")
	api.HandleFunc("/characters/{id}", handlers.GetCharacter).Methods("GET")
	api.HandleFunc("/characters/{id}", handlers.UpdateCharacter).Methods("PUT")
	api.HandleFunc("/dice/roll", handlers.RollDice).Methods("POST")
	api.HandleFunc("/game/session", handlers.CreateGameSession).Methods("POST")
	api.HandleFunc("/game/session/{id}", handlers.GetGameSession).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws", websocket.HandleWebSocket)

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend/build/")))

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
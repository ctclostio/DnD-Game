package routes

import (
	"github.com/gorilla/mux"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(api *mux.Router, cfg *Config) {
	// Auth routes with rate limiting
	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.Use(cfg.AuthRateLimiter.Middleware())

	// Public auth endpoints
	authRouter.HandleFunc("/register", cfg.Handlers.Register).Methods("POST")
	authRouter.HandleFunc("/login", cfg.Handlers.Login).Methods("POST")
	authRouter.HandleFunc("/refresh", cfg.Handlers.RefreshToken).Methods("POST")

	// Protected auth endpoints
	authRouter.HandleFunc("/logout",
		cfg.AuthMiddleware.Authenticate(cfg.Handlers.Logout)).Methods("POST")
	authRouter.HandleFunc("/me",
		cfg.AuthMiddleware.Authenticate(cfg.Handlers.GetCurrentUser)).Methods("GET")
}

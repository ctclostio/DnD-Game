package routes

import (
	"github.com/gorilla/mux"
)

// RegisterRuleBuilderRoutes registers all rule builder-related routes.
func RegisterRuleBuilderRoutes(api *mux.Router, cfg *Config) {
	auth := cfg.AuthMiddleware.Authenticate
	dmOnly := cfg.AuthMiddleware.RequireDM()

	// Rule management.
	api.HandleFunc("/rules", auth(cfg.Handlers.GetRules)).Methods("GET")
	api.HandleFunc("/rules", dmOnly(cfg.Handlers.CreateRule)).Methods("POST")
	api.HandleFunc("/rules/{id}", auth(cfg.Handlers.GetRule)).Methods("GET")
	api.HandleFunc("/rules/{id}", dmOnly(cfg.Handlers.UpdateRule)).Methods("PUT")
	api.HandleFunc("/rules/{id}", dmOnly(cfg.Handlers.DeleteRule)).Methods("DELETE")
	api.HandleFunc("/rules/{id}/activate", dmOnly(cfg.Handlers.ActivateRule)).Methods("POST")
	api.HandleFunc("/rules/{id}/deactivate", dmOnly(cfg.Handlers.DeactivateRule)).Methods("POST")

	// Rule validation and testing.
	api.HandleFunc("/rules/validate", dmOnly(cfg.Handlers.ValidateRule)).Methods("POST")
	api.HandleFunc("/rules/test", dmOnly(cfg.Handlers.TestRule)).Methods("POST")
	api.HandleFunc("/rules/simulate", dmOnly(cfg.Handlers.SimulateRule)).Methods("POST")

	// Rule templates and library.
	api.HandleFunc("/rules/templates", auth(cfg.Handlers.GetRuleTemplates)).Methods("GET")
	api.HandleFunc("/rules/templates/{id}", auth(cfg.Handlers.GetRuleTemplate)).Methods("GET")
	api.HandleFunc("/rules/library", auth(cfg.Handlers.GetRuleLibrary)).Methods("GET")

	// Balance analysis.
	api.HandleFunc("/rules/analyze/balance", dmOnly(cfg.Handlers.AnalyzeBalance)).Methods("POST")
	api.HandleFunc("/rules/analyze/impact", dmOnly(cfg.Handlers.AnalyzeRuleImpact)).Methods("POST")
}

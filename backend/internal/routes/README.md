# Routes Package

This package organizes all HTTP route definitions for the D&D Game backend API.

## Structure

Routes are organized by domain/feature:
- `auth.go` - Authentication routes (login, register, etc.)
- `character.go` - Character management and actions
- `combat.go` - Combat system routes
- `game_session.go` - Game session management
- `npc.go` - NPC management
- `inventory.go` - Inventory and item management
- `dm_assistant.go` - DM Assistant AI features
- `world_building.go` - World building tools
- `rule_builder.go` - Custom rule creation
- `narrative.go` - Narrative engine features

## Adding New Routes

1. Create a new file for your feature domain if needed
2. Create a `Register<Domain>Routes` function
3. Add the registration call to `RegisterRoutes` in `routes.go`
4. Update the `Config` struct if new handlers are needed

## Route Patterns

- All routes are prefixed with `/api/v1`
- Protected routes use `cfg.AuthMiddleware.Authenticate`
- DM-only routes use `cfg.AuthMiddleware.RequireDM()`
- Auth routes have additional rate limiting

## Example

```go
func RegisterMyFeatureRoutes(api *mux.Router, cfg *Config) {
    auth := cfg.AuthMiddleware.Authenticate
    dmOnly := cfg.AuthMiddleware.RequireDM()
    
    // Public route
    api.HandleFunc("/myfeature/public", cfg.Handlers.PublicHandler).Methods("GET")
    
    // Protected route
    api.HandleFunc("/myfeature/data", auth(cfg.Handlers.GetData)).Methods("GET")
    
    // DM-only route
    api.HandleFunc("/myfeature/admin", dmOnly(cfg.Handlers.AdminAction)).Methods("POST")
}
```
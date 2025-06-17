package constants

// API Version and Prefixes
const (
	APIv1Prefix       = "/api/v1"
	APIPrefix         = "/api"
	APIv1AuthPrefix   = "/api/v1/auth"
	APIv1GamePrefix   = "/api/v1/game"
	APIv1CombatPrefix = "/api/v1/combat"
)

// Authentication Routes
const (
	AuthRegisterPath = "/api/v1/auth/register"
	AuthLoginPath    = "/api/v1/auth/login"
	AuthLogoutPath   = "/api/v1/auth/logout"
	AuthMePath       = "/api/v1/auth/me"
)

// Game Session Routes
const (
	GameSessionsPath       = "/api/v1/game/sessions"
	GameSessionByIDPath    = "/api/v1/game/sessions/"
	GameSessionJoinPath    = "/sessions/{id}/join"
	GameSessionLeavePath   = "/sessions/{id}/leave"
	SessionsPath           = "/sessions"
	SessionByIDPath        = "/sessions/{id}"
	APISessionsPath        = "/api/sessions/"
)

// Character Routes
const (
	CharactersPath      = "/api/characters"
	CharacterByIDPath   = "/api/characters/"
	CharacterInventory  = "/inventory"
)

// Combat Routes
const (
	CombatStartPath    = "/api/v1/combat/start"
	CombatByIDPath     = "/api/v1/combat/"
)

// Health Check Routes
const (
	HealthPath         = "/health"
	HealthLivePath     = "/health/live"
	HealthReadyPath    = "/health/ready"
	HealthDetailedPath = "/health/detailed"
)

// Other Routes
const (
	CSRFTokenPath = "/csrf-token"
	SwaggerPath   = "/swagger"
	WebSocketPath = "/ws"
)
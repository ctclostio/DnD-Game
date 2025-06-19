package constants

// HTTP Headers
const (
	ContentType     = "Content-Type"
	Authorization   = "Authorization"
	Bearer          = "Bearer "
	CorrelationID   = "X-Correlation-ID"
	CacheControl    = "Cache-Control"
)

// Content Types
const (
	ApplicationJSON = "application/json"
)

// Common HTTP Error Messages
const (
	// Moved to errors.go to avoid duplication
	ErrNotFound       = "Not found"
	ErrUnauthorized   = "Unauthorized"
	ErrForbidden      = "Forbidden"
	ErrInternalServer = "Internal server error"
)

// Database Error Messages
const (
	ErrDatabaseNotInitialized = "database not initialized"
	ErrIteratingRows          = "error iterating rows: %w"
)

// Session Error Messages
const (
	// Moved to errors.go to avoid duplication
	ErrSessionNameRequired = "session name is required"
)

// Character Error Messages
const (
	// Moved to errors.go to avoid duplication
)

// Common Parameter Names
const (
	ParamSessionID   = "sessionId"
	ParamCharacterID = "characterId"
	ParamItemID      = "itemId"
	ParamUserID      = "userId"
	ParamCombatID    = "combatId"
	ParamNPCID       = "npcId"
	ParamArcID       = "arcId"
	ParamElementID   = "elementId"
)

// Common Field Names
const (
	FieldDescription   = "description"
	FieldMetadata      = "metadata"
	FieldCreatedAt     = "created_at"
	FieldUpdatedAt     = "updated_at"
	FieldGameSessionID = "game_session_id"
)

// Other Common Errors
const (
	ErrDatabaseError      = "Database error"
	// ErrInvalidItemID moved to errors.go
	ErrPlayerNotInSession = "Player not in session"
	// ErrCombatNotActive moved to errors.go
	ErrUnauthorizedAction = "Unauthorized action"
)

// WebSocket Messages
const (
	WebSocketCloseError = "Failed to close WebSocket connection: %v"
)

// URL Patterns
const (
	PaginationURLFormat = "%s?page=%d&limit=%d"
	WebSocketURLFormat  = "%s?room=%s"
)

// Test-related constants that are also used in non-test code
const (
	LocalhostURL = "http://localhost:3000"
)

// SQL Query Patterns
const (
	SQLOrderByFormat   = " ORDER BY %s %s"
	SQLLimitOffsetFormat = " LIMIT ? OFFSET ?"
)
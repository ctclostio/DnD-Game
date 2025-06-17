package testutil

// Test ID constants
const (
	TestSessionID      = "session-123"
	TestUserID         = "user-123"
	TestDMID           = "dm-123"
	TestCharacterID    = "char-123"
	TestParticipantID  = "participant-123"
	TestCombatID       = "combat-123"
	TestEncounterID    = "encounter-123"
	TestItemID         = "item-123"
	TestNarrativeID    = "narrative-123"
	TestWorldID        = "world-123"
	TestFactionID      = "faction-123"
	TestEventID        = "event-123"
)

// Test names and descriptions
const (
	TestSessionName        = "Test Session"
	TestSessionDescription = "A test session"
	TestCharacterName      = "Test Character"
	TestEmail              = "test@example.com"
)

// Test error messages
const (
	TestDatabaseError    = "database error"
	TestValidationError  = "validation error"
	TestNotFoundError    = "not found"
	TestUnauthorizedError = "unauthorized"
)

// API endpoint constants
const (
	APISessionsPath   = "/api/v1/sessions/"
	APICharactersPath = "/api/v1/characters/"
	APICombatPath     = "/api/v1/combat/"
	APIAuthPath       = "/api/v1/auth/"
	APIAuthMePath     = "/api/v1/auth/me"
	APIInventoryPath  = "/api/v1/inventory/"
)

// HTTP constants
const (
	ContentTypeHeader = "Content-Type"
	ContentTypeJSON   = "application/json"
	AuthHeader        = "Authorization"
	BearerPrefix      = "Bearer "
)
package constants

// Error message templates
const (
	// Generic error templates
	ErrMsgFailedTo        = "failed to %s: %w"
	ErrMsgNotFound        = "%s not found"
	ErrMsgRequired        = "%s is required"
	ErrMsgInvalid         = "invalid %s"
	ErrMsgUnauthorized    = "unauthorized"
	ErrMsgForbidden       = "forbidden"
	ErrMsgAlreadyExists   = "%s already exists"
	ErrMsgCannotProcess   = "cannot process %s"
	ErrMsgDatabaseError   = "database error"
	ErrMsgInternalError   = "internal server error"
	ErrMsgValidationError = "validation error"

	// Entity-specific not found errors
	ErrCharacterNotFound   = "character not found"
	ErrSessionNotFound     = "session not found"
	ErrGameSessionNotFound = "game session not found"
	ErrUserNotFound        = "user not found"
	ErrItemNotFound        = "item not found"
	ErrCampaignNotFound    = "campaign not found"
	ErrRaceNotFound        = "race not found"
	ErrClassNotFound       = "class not found"
	ErrSubclassNotFound    = "subclass not found"
	ErrBackgroundNotFound  = "background not found"
	ErrInventoryNotFound   = "inventory not found"
	ErrEncounterNotFound   = "encounter not found"
	ErrRuleNotFound        = "rule not found"
	ErrEventNotFound       = "event not found"
	ErrProfileNotFound     = "profile not found"
	ErrAgendaNotFound      = "agenda not found"
	ErrFactionNotFound     = "faction not found"
	ErrSettlementNotFound  = "settlement not found"
	ErrWorldStateNotFound  = "world state not found"
	ErrRelationshipNotFound = "relationship not found"

	// Validation errors
	ErrInvalidInput        = "invalid input"
	ErrInvalidID           = "invalid ID"
	ErrInvalidSessionID    = "invalid session ID"
	ErrInvalidCharacterID  = "invalid character ID"
	ErrInvalidUserID       = "invalid user ID"
	ErrInvalidCampaignID   = "invalid campaign ID"
	ErrInvalidItemID       = "invalid item ID"
	ErrInvalidJSON         = "invalid JSON"
	ErrInvalidRequest      = "invalid request"
	ErrInvalidCredentials  = "invalid username or password"
	ErrInvalidToken        = "invalid token"
	ErrTokenExpired        = "token expired"

	// Required field errors
	ErrUserIDRequired      = "user ID is required"
	ErrSessionIDRequired   = "session ID is required"
	ErrCharacterIDRequired = "character ID is required"
	ErrNameRequired        = "name is required"
	ErrEmailRequired       = "email is required"
	ErrPasswordRequired    = "password is required"
	ErrUsernameRequired    = "username is required"

	// Operation errors
	ErrFailedToCreate      = "failed to create %s"
	ErrFailedToUpdate      = "failed to update %s"
	ErrFailedToDelete      = "failed to delete %s"
	ErrFailedToGet         = "failed to get %s"
	ErrFailedToList        = "failed to list %s"
	ErrFailedToParse       = "failed to parse %s"
	ErrFailedToValidate    = "failed to validate %s"
	ErrFailedToMarshal     = "failed to marshal %s"
	ErrFailedToUnmarshal   = "failed to unmarshal %s"
	ErrFailedToEncode      = "failed to encode %s"
	ErrFailedToDecode      = "failed to decode %s"
	ErrFailedToHash        = "failed to hash password"
	ErrFailedToGenerate    = "failed to generate %s"
	ErrFailedToProcess     = "failed to process %s"
	ErrFailedToSave        = "failed to save %s"
	ErrFailedToLoad        = "failed to load %s"

	// Permission errors
	ErrNotOwner            = "not the owner of this %s"
	ErrNoPermission        = "no permission to %s"
	ErrAccessDenied        = "access denied"

	// State errors
	ErrAlreadyInCombat     = "already in combat"
	ErrNotInCombat         = "not in combat"
	ErrCombatNotStarted    = "combat not started"
	ErrCombatAlreadyStarted = "combat already started"
	ErrNotYourTurn         = "not your turn"
	ErrInvalidState        = "invalid state"

	// Limit errors
	ErrLimitExceeded       = "%s limit exceeded"
	ErrMaxRetriesExceeded  = "maximum retries exceeded"
	ErrRateLimitExceeded   = "rate limit exceeded"
	
	// Database operation format strings
	ErrFailedToMarshalParameters = "failed to marshal parameters: %w"
	ErrFailedToMarshalMetadata   = "failed to marshal metadata: %w"
	ErrFailedToUnmarshalMetadata = "failed to unmarshal metadata: %w"
	ErrFailedToGetRowsAffected   = "failed to get rows affected: %w"
	
	// Migration errors
	ErrFailedToCreateMigrationSource   = "failed to create migration source: %w"
	ErrFailedToCreateMigrationDriver   = "failed to create migration driver: %w"
	ErrFailedToCreateMigrateInstance   = "failed to create migrate instance: %w"
	
	// API errors
	ErrInvalidRequestBody = "Invalid request body"
	ErrCharacterNotFoundCap = "Character not found"
)
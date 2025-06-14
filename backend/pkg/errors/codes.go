package errors

// ErrorCode represents specific error codes for better debugging.
type ErrorCode string

const (
	// Authentication & Authorization.
	ErrCodeInvalidCredentials    ErrorCode = "AUTH001"
	ErrCodeTokenExpired          ErrorCode = "AUTH002"
	ErrCodeTokenInvalid          ErrorCode = "AUTH003"
	ErrCodeInsufficientPrivilege ErrorCode = "AUTH004"
	ErrCodeSessionExpired        ErrorCode = "AUTH005"
	ErrCodeCSRFTokenMismatch     ErrorCode = "AUTH006"

	// User Management.
	ErrCodeUserNotFound     ErrorCode = "USER001"
	ErrCodeUserExists       ErrorCode = "USER002"
	ErrCodeInvalidPassword  ErrorCode = "USER003"
	ErrCodeEmailNotVerified ErrorCode = "USER004"

	// Character Management.
	ErrCodeCharacterNotFound     ErrorCode = "CHAR001"
	ErrCodeCharacterLimitReached ErrorCode = "CHAR002"
	ErrCodeInvalidCharacterData  ErrorCode = "CHAR003"
	ErrCodeCharacterNotOwned     ErrorCode = "CHAR004"

	// Game Session.
	ErrCodeSessionNotFound   ErrorCode = "GAME001"
	ErrCodeSessionFull       ErrorCode = "GAME002"
	ErrCodeSessionInProgress ErrorCode = "GAME003"
	ErrCodeNotInSession      ErrorCode = "GAME004"
	ErrCodeNotDM             ErrorCode = "GAME005"

	// Combat.
	ErrCodeCombatNotActive    ErrorCode = "COMBAT001"
	ErrCodeNotYourTurn        ErrorCode = "COMBAT002"
	ErrCodeInvalidTarget      ErrorCode = "COMBAT003"
	ErrCodeInsufficientRange  ErrorCode = "COMBAT004"
	ErrCodeNoActionsRemaining ErrorCode = "COMBAT005"

	// Inventory.
	ErrCodeItemNotFound        ErrorCode = "INV001"
	ErrCodeInventoryFull       ErrorCode = "INV002"
	ErrCodeInsufficientFunds   ErrorCode = "INV003"
	ErrCodeItemNotEquippable   ErrorCode = "INV004"
	ErrCodeItemAlreadyEquipped ErrorCode = "INV005"

	// AI Services.
	ErrCodeAIServiceUnavailable ErrorCode = "AI001"
	ErrCodeAIGenerationFailed   ErrorCode = "AI002"
	ErrCodeAIRateLimitExceeded  ErrorCode = "AI003"
	ErrCodeAIInvalidRequest     ErrorCode = "AI004"

	// Validation.
	ErrCodeValidationFailed ErrorCode = "VAL001"
	ErrCodeInvalidInput     ErrorCode = "VAL002"
	ErrCodeMissingRequired  ErrorCode = "VAL003"
	ErrCodeInvalidFormat    ErrorCode = "VAL004"
	ErrCodeOutOfRange       ErrorCode = "VAL005"

	// Database.
	ErrCodeDatabaseError       ErrorCode = "DB001"
	ErrCodeDuplicateEntry      ErrorCode = "DB002"
	ErrCodeForeignKeyViolation ErrorCode = "DB003"
	ErrCodeDeadlock            ErrorCode = "DB004"

	// General.
	ErrCodeInternalError      ErrorCode = "INT001"
	ErrCodeServiceUnavailable ErrorCode = "INT002"
	ErrCodeTimeout            ErrorCode = "INT003"
	ErrCodeRateLimitExceeded  ErrorCode = "INT004"
)

// ErrorCodeMessages provides human-readable descriptions for error codes.
var ErrorCodeMessages = map[ErrorCode]string{
	// Authentication & Authorization.
	ErrCodeInvalidCredentials:    "Invalid username or password",
	ErrCodeTokenExpired:          "Authentication token has expired",
	ErrCodeTokenInvalid:          "Invalid authentication token",
	ErrCodeInsufficientPrivilege: "Insufficient privileges to perform this action",
	ErrCodeSessionExpired:        "Session has expired",
	ErrCodeCSRFTokenMismatch:     "CSRF token mismatch",

	// User Management.
	ErrCodeUserNotFound:     "User not found",
	ErrCodeUserExists:       "User already exists",
	ErrCodeInvalidPassword:  "Password does not meet requirements",
	ErrCodeEmailNotVerified: "Email address not verified",

	// Character Management.
	ErrCodeCharacterNotFound:     "Character not found",
	ErrCodeCharacterLimitReached: "Character limit reached",
	ErrCodeInvalidCharacterData:  "Invalid character data",
	ErrCodeCharacterNotOwned:     "Character not owned by user",

	// Game Session.
	ErrCodeSessionNotFound:   "Game session not found",
	ErrCodeSessionFull:       "Game session is full",
	ErrCodeSessionInProgress: "Game session already in progress",
	ErrCodeNotInSession:      "Not a participant in this session",
	ErrCodeNotDM:             "Only the DM can perform this action",

	// Combat.
	ErrCodeCombatNotActive:    "No active combat",
	ErrCodeNotYourTurn:        "Not your turn",
	ErrCodeInvalidTarget:      "Invalid target",
	ErrCodeInsufficientRange:  "Target out of range",
	ErrCodeNoActionsRemaining: "No actions remaining",

	// Inventory.
	ErrCodeItemNotFound:        "Item not found",
	ErrCodeInventoryFull:       "Inventory is full",
	ErrCodeInsufficientFunds:   "Insufficient funds",
	ErrCodeItemNotEquippable:   "Item cannot be equipped",
	ErrCodeItemAlreadyEquipped: "Item is already equipped",

	// AI Services.
	ErrCodeAIServiceUnavailable: "AI service is temporarily unavailable",
	ErrCodeAIGenerationFailed:   "AI content generation failed",
	ErrCodeAIRateLimitExceeded:  "AI service rate limit exceeded",
	ErrCodeAIInvalidRequest:     "Invalid request to AI service",

	// Validation.
	ErrCodeValidationFailed: "Validation failed",
	ErrCodeInvalidInput:     "Invalid input provided",
	ErrCodeMissingRequired:  "Missing required field",
	ErrCodeInvalidFormat:    "Invalid format",
	ErrCodeOutOfRange:       "Value out of allowed range",

	// Database.
	ErrCodeDatabaseError:       "Database operation failed",
	ErrCodeDuplicateEntry:      "Duplicate entry",
	ErrCodeForeignKeyViolation: "Foreign key constraint violation",
	ErrCodeDeadlock:            "Database deadlock detected",

	// General.
	ErrCodeInternalError:      "Internal server error",
	ErrCodeServiceUnavailable: "Service temporarily unavailable",
	ErrCodeTimeout:            "Request timeout",
	ErrCodeRateLimitExceeded:  "Rate limit exceeded",
}

// GetErrorMessage returns the message for an error code.
func GetErrorMessage(code ErrorCode) string {
	if msg, ok := ErrorCodeMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}

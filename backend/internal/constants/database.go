package constants

// Database query fragments
const (
	// Common JSON values
	EmptyJSON = "{}"
	
	// Query fragments
	AndGameSessionIDClause = " AND game_session_id = ?"
	AndCharacterIDClause   = " AND character_id = ?"
	
	// Status values
	StatusActive = "active"
)
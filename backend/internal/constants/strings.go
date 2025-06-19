package constants

// Common string constants used throughout the application
const (
	// HTTP Headers
	HeaderContentType     = "Content-Type"
	HeaderAuthorization   = "Authorization"
	HeaderCacheControl    = "Cache-Control"
	HeaderIfNoneMatch     = "If-None-Match"
	HeaderETag            = "ETag"
	HeaderLastModified    = "Last-Modified"
	HeaderIfModifiedSince = "If-Modified-Since"
	HeaderXRequestID      = "X-Request-ID"
	HeaderXUserID         = "X-User-ID"
	HeaderXSessionID      = "X-Session-ID"

	// Content Types
	ContentTypeJSON        = "application/json"
	ContentTypeHTML        = "text/html"
	ContentTypePlainText   = "text/plain"
	ContentTypeFormEncoded = "application/x-www-form-urlencoded"
	ContentTypeMultipart   = "multipart/form-data"

	// Cache Control Values
	CacheControlNoCache      = "no-cache"
	CacheControlNoStore      = "no-store"
	CacheControlPublic       = "public, max-age=3600"
	CacheControlPrivate      = "private, max-age=3600"
	CacheControlMustRevalidate = "must-revalidate"

	// Database Column Names
	ColumnID          = "id"
	ColumnCreatedAt   = "created_at"
	ColumnUpdatedAt   = "updated_at"
	ColumnDeletedAt   = "deleted_at"
	ColumnUserID      = "user_id"
	ColumnSessionID   = "session_id"
	ColumnCharacterID = "character_id"
	ColumnCampaignID  = "campaign_id"
	ColumnName        = "name"
	ColumnDescription = "description"
	ColumnEmail       = "email"
	ColumnUsername    = "username"
	ColumnPassword    = "password_hash"
	ColumnStatus      = "status"
	ColumnType        = "type"
	ColumnData        = "data"
	ColumnMetadata    = "metadata"
	ColumnIsActive    = "is_active"
	ColumnIsDeleted   = "is_deleted"

	// SQL Query Fragments
	SQLSelectAll      = "SELECT * FROM"
	SQLSelectCount    = "SELECT COUNT(*) FROM"
	SQLWhere          = "WHERE"
	SQLAnd            = "AND"
	SQLOr             = "OR"
	SQLOrderBy        = "ORDER BY"
	SQLLimit          = "LIMIT"
	SQLOffset         = "OFFSET"
	SQLInsertInto     = "INSERT INTO"
	SQLValues         = "VALUES"
	SQLUpdate         = "UPDATE"
	SQLSet            = "SET"
	SQLDeleteFrom     = "DELETE FROM"
	SQLJoin           = "JOIN"
	SQLLeftJoin       = "LEFT JOIN"
	SQLInnerJoin      = "INNER JOIN"
	SQLOn             = "ON"
	SQLGroupBy        = "GROUP BY"
	SQLHaving         = "HAVING"
	SQLReturning      = "RETURNING"

	// Pagination
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100

	// Status Values
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusDeleted   = "deleted"
	StatusDraft     = "draft"
	StatusPublished = "published"

	// Common Formats
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
	RFC3339Format  = "2006-01-02T15:04:05Z07:00"

	// Validation Messages
	ValidationMinLength = "must be at least %d characters"
	ValidationMaxLength = "must be at most %d characters"
	ValidationRequired  = "is required"
	ValidationEmail     = "must be a valid email"
	ValidationNumeric   = "must be numeric"
	ValidationAlpha     = "must contain only letters"
	ValidationAlphaNum  = "must contain only letters and numbers"

	// Log Messages
	LogRequestReceived = "request received"
	LogRequestCompleted = "request completed"
	LogDatabaseQuery   = "executing database query"
	LogCacheHit        = "cache hit"
	LogCacheMiss       = "cache miss"
	LogError           = "error occurred"
	LogWarning         = "warning"
	LogInfo            = "info"
	LogDebug           = "debug"
)
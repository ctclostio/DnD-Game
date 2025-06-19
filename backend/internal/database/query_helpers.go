package database

import (
	"fmt"
	"strings"
	
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
)

// Common format strings
const (
	threePartFormat = "%s %s %s"
)

// QueryBuilder provides helper functions to construct SQL queries
type QueryBuilder struct{}

// NewQueryBuilder creates a new query builder instance
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

// SelectAll constructs a SELECT * FROM query
func (qb *QueryBuilder) SelectAll(table string) string {
	return fmt.Sprintf("%s %s", constants.SQLSelectAll, table)
}

// SelectColumns constructs a SELECT columns FROM query
func (qb *QueryBuilder) SelectColumns(columns []string, table string) string {
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table)
}

// SelectCount constructs a SELECT COUNT(*) FROM query
func (qb *QueryBuilder) SelectCount(table string) string {
	return fmt.Sprintf("%s %s", constants.SQLSelectCount, table)
}

// Where adds a WHERE clause
func (qb *QueryBuilder) Where(query, condition string) string {
	return fmt.Sprintf(threePartFormat, query, constants.SQLWhere, condition)
}

// And adds an AND condition
func (qb *QueryBuilder) And(query, condition string) string {
	return fmt.Sprintf(threePartFormat, query, constants.SQLAnd, condition)
}

// Or adds an OR condition
func (qb *QueryBuilder) Or(query, condition string) string {
	return fmt.Sprintf(threePartFormat, query, constants.SQLOr, condition)
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(query, column, direction string) string {
	return fmt.Sprintf("%s %s %s %s", query, constants.SQLOrderBy, column, direction)
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(query string, limit int) string {
	return fmt.Sprintf("%s %s %d", query, constants.SQLLimit, limit)
}

// Offset adds an OFFSET clause
func (qb *QueryBuilder) Offset(query string, offset int) string {
	return fmt.Sprintf("%s %s %d", query, constants.SQLOffset, offset)
}

// LimitOffset adds both LIMIT and OFFSET clauses
func (qb *QueryBuilder) LimitOffset(query string, limit, offset int) string {
	return fmt.Sprintf("%s %s %d %s %d", query, constants.SQLLimit, limit, constants.SQLOffset, offset)
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(query, table, condition string) string {
	return fmt.Sprintf("%s %s %s %s %s", query, constants.SQLJoin, table, constants.SQLOn, condition)
}

// LeftJoin adds a LEFT JOIN clause
func (qb *QueryBuilder) LeftJoin(query, table, condition string) string {
	return fmt.Sprintf("%s %s %s %s %s", query, constants.SQLLeftJoin, table, constants.SQLOn, condition)
}

// Common query fragments
const (
	// Common ORDER BY fragments
	OrderByCreatedAtDesc = " ORDER BY created_at DESC"
	OrderByCreatedAtAsc  = " ORDER BY created_at ASC"
	OrderByUpdatedAtDesc = " ORDER BY updated_at DESC"
	OrderByUpdatedAtAsc  = " ORDER BY updated_at ASC"
	OrderByNameAsc       = " ORDER BY name ASC"
	OrderByNameDesc      = " ORDER BY name DESC"
	
	// Common LIMIT OFFSET fragments
	LimitOffsetPlaceholder = " LIMIT ? OFFSET ?"
	
	// Common WHERE conditions
	WhereID              = " WHERE id = ?"
	WhereUserID          = " WHERE user_id = ?"
	WhereSessionID       = " WHERE session_id = ?"
	WhereCharacterID     = " WHERE character_id = ?"
	WhereCampaignID      = " WHERE campaign_id = ?"
	WhereDeletedAtIsNull = " WHERE deleted_at IS NULL"
	WhereIsActive        = " WHERE is_active = true"
	
	// Common column lists
	CommonColumns = "id, created_at, updated_at"
	AuditColumns  = "created_at, updated_at, deleted_at"
)
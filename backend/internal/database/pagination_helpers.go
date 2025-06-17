package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/pagination"
	"github.com/jmoiron/sqlx"
)

// Common pagination constants
const (
	DefaultSortColumn = "created_at"
	LimitOffsetClause = " LIMIT ? OFFSET ?"
	
	// CharacterSelectQuery is the base query for selecting characters
	CharacterSelectQuery = `
		SELECT id, user_id, name, race, class, level, experience, 
		       hit_points, max_hit_points, armor_class, alignment,
		       background, attributes, skills, proficiencies,
		       equipment, spells, features, created_at, updated_at
		FROM characters
		WHERE user_id = ?`
)

// PaginatedRepository provides pagination helpers for repositories
type PaginatedRepository struct {
	db *DB
}

// NewPaginatedRepository creates a new paginated repository
func NewPaginatedRepository(db *DB) *PaginatedRepository {
	return &PaginatedRepository{db: db}
}

// GetCharactersPaginated returns paginated characters
func (pr *PaginatedRepository) GetCharactersPaginated(ctx context.Context, userID string, params *pagination.PaginationParams) (*pagination.PageResult, error) {
	// Base query
	baseQuery := CharacterSelectQuery

	// Count query
	countQuery := `SELECT COUNT(*) FROM characters WHERE user_id = ?`

	// Apply filters
	var whereClauses []string
	var args []interface{}
	args = append(args, userID)

	if class, ok := params.Filters["class"].(string); ok && class != "" {
		whereClauses = append(whereClauses, "class = ?")
		args = append(args, class)
	}

	if race, ok := params.Filters["race"].(string); ok && race != "" {
		whereClauses = append(whereClauses, "race = ?")
		args = append(args, race)
	}

	if minLevel, ok := params.Filters["min_level"].(int); ok && minLevel > 0 {
		whereClauses = append(whereClauses, "level >= ?")
		args = append(args, minLevel)
	}

	// Add WHERE clauses
	if len(whereClauses) > 0 {
		whereClause := " AND " + strings.Join(whereClauses, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Add sorting
	sortColumn := DefaultSortColumn
	if params.SortBy != "" {
		// Validate sort column
		validColumns := map[string]bool{
			"name":       true,
			"level":      true,
			"class":      true,
			"race":       true,
			"created_at": true,
			"updated_at": true,
		}
		if validColumns[params.SortBy] {
			sortColumn = params.SortBy
		}
	}
	baseQuery += fmt.Sprintf(constants.SQLOrderByFormat, sortColumn, params.SortDir)

	// Add pagination
	baseQuery += LimitOffsetClause
	args = append(args, params.Limit, params.GetOffset())

	// Execute count query
	var total int64
	countArgs := args[:len(args)-2] // Exclude LIMIT and OFFSET
	err := pr.db.QueryRowContext(ctx, pr.db.Rebind(countQuery), countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count characters: %w", err)
	}

	// Execute main query
	var characters []*models.Character
	reboundQuery := pr.db.Rebind(baseQuery)
	err = pr.db.SelectContext(ctx, &characters, reboundQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query characters: %w", err)
	}

	return pagination.NewPageResult(characters, params, total), nil
}

// GetGameSessionsPaginated returns paginated game sessions
func (pr *PaginatedRepository) GetGameSessionsPaginated(ctx context.Context, params *pagination.PaginationParams) (*pagination.PageResult, error) {
	// Base query
	baseQuery := `
		SELECT id, dm_user_id, name, description, status, 
		       max_players, current_players, settings, state,
		       created_at, updated_at, started_at, ended_at
		FROM game_sessions
		WHERE 1=1`

	// Count query
	countQuery := `SELECT COUNT(*) FROM game_sessions WHERE 1=1`

	// Apply filters
	var whereClauses []string
	var args []interface{}

	if status, ok := params.Filters["status"].(string); ok && status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}

	if dmUserID, ok := params.Filters["dm_user_id"].(string); ok && dmUserID != "" {
		whereClauses = append(whereClauses, "dm_user_id = ?")
		args = append(args, dmUserID)
	}

	// Add WHERE clauses
	if len(whereClauses) > 0 {
		whereClause := " AND " + strings.Join(whereClauses, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Add sorting
	sortColumn := DefaultSortColumn
	if params.SortBy != "" {
		validColumns := map[string]bool{
			"name":            true,
			"status":          true,
			"current_players": true,
			"created_at":      true,
			"updated_at":      true,
			"started_at":      true,
		}
		if validColumns[params.SortBy] {
			sortColumn = params.SortBy
		}
	}
	baseQuery += fmt.Sprintf(constants.SQLOrderByFormat, sortColumn, params.SortDir)

	// Add pagination
	baseQuery += LimitOffsetClause
	args = append(args, params.Limit, params.GetOffset())

	// Execute count query
	var total int64
	countArgs := args[:len(args)-2]
	err := pr.db.QueryRowContext(ctx, pr.db.Rebind(countQuery), countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count game sessions: %w", err)
	}

	// Execute main query
	var sessions []*models.GameSession
	reboundQuery := pr.db.Rebind(baseQuery)
	err = pr.db.SelectContext(ctx, &sessions, reboundQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query game sessions: %w", err)
	}

	return pagination.NewPageResult(sessions, params, total), nil
}

// GetCampaignsPaginated returns paginated campaigns
// TODO: This function needs to be updated to match the current schema
// Currently commented out to fix build errors
/*
func (pr *PaginatedRepository) GetCampaignsPaginated(ctx context.Context, userID string, params *pagination.PaginationParams) (*pagination.PageResult, error) {
	baseQuery := `
		SELECT c.*, COUNT(DISTINCT cp.user_id) as player_count
		FROM campaigns c
		LEFT JOIN campaign_players cp ON c.id = cp.campaign_id
		WHERE c.owner_id = ? OR cp.user_id = ?
		GROUP BY c.id`

	countQuery := `
		SELECT COUNT(DISTINCT c.id)
		FROM campaigns c
		LEFT JOIN campaign_players cp ON c.id = cp.campaign_id
		WHERE c.owner_id = ? OR cp.user_id = ?`

	args := []interface{}{userID, userID}

	// Apply filters
	if status, ok := params.Filters["status"].(string); ok && status != "" {
		baseQuery = strings.Replace(baseQuery, "WHERE", "WHERE c.status = ? AND", 1)
		countQuery = strings.Replace(countQuery, "WHERE", "WHERE c.status = ? AND", 1)
		args = append([]interface{}{status}, args...)
	}

	// Add sorting
	sortColumn := "c.created_at"
	if params.SortBy == "name" {
		sortColumn = "c.name"
	} else if params.SortBy == "player_count" {
		sortColumn = "player_count"
	}
	baseQuery += fmt.Sprintf(constants.SQLOrderByFormat, sortColumn, params.SortDir)

	// Add pagination
	baseQuery += LimitOffsetClause
	args = append(args, params.Limit, params.GetOffset())

	// Execute count query
	var total int64
	countArgs := args[:len(args)-2]
	err := pr.db.QueryRowContext(ctx, pr.db.Rebind(countQuery), countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Execute main query
	rows, err := pr.db.QueryContext(ctx, pr.db.Rebind(baseQuery), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*models.GameSession
	for rows.Next() {
		var campaign models.GameSession
		var playerCount int
		err := rows.Scan(
			&campaign.ID, &campaign.OwnerID, &campaign.Name, &campaign.Description,
			&campaign.Status, &campaign.Settings, &campaign.CreatedAt, &campaign.UpdatedAt,
			&playerCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		campaign.PlayerCount = playerCount
		campaigns = append(campaigns, &campaign)
	}

	return pagination.NewPageResult(campaigns, params, total), nil
}
*/

// CursorPaginationHelper helps with cursor-based pagination
type CursorPaginationHelper struct {
	db *DB
}

// NewCursorPaginationHelper creates a new cursor pagination helper
func NewCursorPaginationHelper(db *DB) *CursorPaginationHelper {
	return &CursorPaginationHelper{db: db}
}

// GetCharactersCursor returns cursor-paginated characters
func (cph *CursorPaginationHelper) GetCharactersCursor(ctx context.Context, userID string, params *pagination.PaginationParams) (*pagination.CursorResult, error) {
	var characters []*models.Character
	query := CharacterSelectQuery

	args := []interface{}{userID}

	// Handle cursor
	if params.Cursor != "" {
		cursor, err := pagination.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}

		// Add cursor condition
		if params.SortDir == "desc" {
			query += " AND created_at < ?"
		} else {
			query += " AND created_at > ?"
		}
		args = append(args, cursor.Timestamp)
	}

	// Add sorting and limit
	query += fmt.Sprintf(" ORDER BY created_at %s, id %s LIMIT ?", params.SortDir, params.SortDir)
	args = append(args, params.Limit+1) // Get one extra to check if there's more

	// Execute query
	reboundQuery := cph.db.Rebind(query)
	err := cph.db.SelectContext(ctx, &characters, reboundQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query characters: %w", err)
	}

	// Check if there are more results
	hasMore := len(characters) > params.Limit
	if hasMore {
		characters = characters[:params.Limit]
	}

	// Create cursors
	var nextCursor, prevCursor *pagination.Cursor

	if hasMore && len(characters) > 0 {
		last := characters[len(characters)-1]
		nextCursor = &pagination.Cursor{
			ID:        last.ID,
			Timestamp: last.CreatedAt,
		}
	}

	// For previous cursor, we'd need to query in reverse direction
	// This is simplified - in production, you'd implement proper bidirectional cursors

	return pagination.NewCursorResult(characters, params, nextCursor, prevCursor), nil
}

// BatchPaginator helps paginate through large datasets for batch processing
type BatchPaginator struct {
	db        *DB
	query     string
	args      []interface{}
	batchSize int
	lastID    string
}

// NewBatchPaginator creates a new batch paginator
func NewBatchPaginator(db *DB, query string, args []interface{}, batchSize int) *BatchPaginator {
	return &BatchPaginator{
		db:        db,
		query:     query,
		args:      args,
		batchSize: batchSize,
	}
}

// NextBatch retrieves the next batch of results
func (bp *BatchPaginator) NextBatch(ctx context.Context, scanFunc func(*sqlx.Rows) error) (hasMore bool, err error) {
	batchQuery := bp.query
	args := bp.args

	// Add cursor condition if not first batch
	if bp.lastID != "" {
		batchQuery += " AND id > ?"
		args = append(args, bp.lastID)
	}

	// Add ordering and limit
	batchQuery += " ORDER BY id ASC LIMIT ?"
	args = append(args, bp.batchSize+1)

	// Execute query
	rows, err := bp.db.QueryxContext(ctx, bp.db.Rebind(batchQuery), args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	count := 0
	var lastID string

	for rows.Next() {
		if count >= bp.batchSize {
			hasMore = true
			break
		}

		if err := scanFunc(rows); err != nil {
			return false, err
		}

		// Get the ID for cursor (assumes first column is ID)
		var id string
		if err := rows.Scan(&id); err != nil {
			return false, fmt.Errorf("failed to scan id: %w", err)
		}
		lastID = id
		count++
	}

	bp.lastID = lastID
	return hasMore, nil
}

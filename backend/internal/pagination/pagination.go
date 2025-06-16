package pagination

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Sort direction constants
const (
	SortDirectionAsc  = "asc"
	SortDirectionDesc = "desc"
)

// Paginator interface for different pagination strategies
type Paginator interface {
	GetLimit() int
	GetOffset() int
	GetCursor() string
	HasMore() bool
	GetTotal() int64
}

// PaginationParams contains common pagination parameters
type PaginationParams struct {
	// Offset-based pagination
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`

	// Cursor-based pagination
	Cursor string `json:"cursor,omitempty"`

	// Sorting
	SortBy  string `json:"sort_by,omitempty"`
	SortDir string `json:"sort_dir,omitempty"`

	// Filtering
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// DefaultPaginationParams returns default pagination parameters
func DefaultPaginationParams() *PaginationParams {
	return &PaginationParams{
		Page:    1,
		Limit:   20,
		SortDir: "asc",
	}
}

// FromRequest extracts pagination parameters from HTTP request
func FromRequest(r *http.Request) *PaginationParams {
	params := DefaultPaginationParams()
	query := r.URL.Query()

	// Page-based pagination
	if page := query.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	// Limit
	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			params.Limit = l
			// Cap limit to prevent abuse
			if params.Limit > 100 {
				params.Limit = 100
			}
		}
	}

	// Cursor-based pagination
	if cursor := query.Get("cursor"); cursor != "" {
		params.Cursor = cursor
	}

	// Sorting
	if sortBy := query.Get("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}

	if sortDir := query.Get("sort_dir"); sortDir != "" {
		if sortDir == SortDirectionDesc || sortDir == SortDirectionAsc {
			params.SortDir = sortDir
		}
	}

	// Filters (simple implementation - enhance as needed)
	params.Filters = make(map[string]interface{})
	for key, values := range query {
		if strings.HasPrefix(key, "filter_") {
			filterKey := strings.TrimPrefix(key, "filter_")
			if len(values) > 0 {
				params.Filters[filterKey] = values[0]
			}
		}
	}

	return params
}

// Validate validates pagination parameters
func (p *PaginationParams) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page must be >= 1")
	}
	if p.Limit < 1 || p.Limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100")
	}
	if p.SortDir != "asc" && p.SortDir != "desc" {
		return fmt.Errorf("sort_dir must be 'asc' or 'desc'")
	}
	return nil
}

// GetOffset calculates offset from page and limit
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// PageResult represents a paginated result set
type PageResult struct {
	Data       interface{} `json:"data"`
	Pagination PageInfo    `json:"pagination"`
}

// PageInfo contains pagination metadata
type PageInfo struct {
	Page       int    `json:"page,omitempty"`
	Limit      int    `json:"limit"`
	Total      int64  `json:"total"`
	TotalPages int    `json:"total_pages,omitempty"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// NewPageResult creates a new paginated result
func NewPageResult(data interface{}, params *PaginationParams, total int64) *PageResult {
	totalPages := int(total / int64(params.Limit))
	if total%int64(params.Limit) > 0 {
		totalPages++
	}

	hasMore := params.Page < totalPages

	return &PageResult{
		Data: data,
		Pagination: PageInfo{
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    hasMore,
		},
	}
}

// Cursor represents a cursor for cursor-based pagination
type Cursor struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value,omitempty"`
}

// EncodeCursor encodes a cursor to a string
func EncodeCursor(cursor *Cursor) string {
	data, _ := json.Marshal(cursor)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes a cursor from a string
func DecodeCursor(encoded string) (*Cursor, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format")
	}

	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor data")
	}

	return &cursor, nil
}

// CursorResult represents a cursor-based paginated result
type CursorResult struct {
	Data       interface{} `json:"data"`
	Pagination CursorInfo  `json:"pagination"`
}

// CursorInfo contains cursor pagination metadata
type CursorInfo struct {
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
}

// NewCursorResult creates a new cursor-based result
func NewCursorResult(data interface{}, params *PaginationParams, nextCursor, prevCursor *Cursor) *CursorResult {
	result := &CursorResult{
		Data: data,
		Pagination: CursorInfo{
			Limit:   params.Limit,
			HasMore: nextCursor != nil,
		},
	}

	if nextCursor != nil {
		result.Pagination.NextCursor = EncodeCursor(nextCursor)
	}

	if prevCursor != nil {
		result.Pagination.PrevCursor = EncodeCursor(prevCursor)
	}

	return result
}

// SQLBuilder helps build paginated SQL queries
type SQLBuilder struct {
	baseQuery  string
	countQuery string
	params     *PaginationParams
	bindings   []interface{}
	paramIndex int
}

// NewSQLBuilder creates a new SQL query builder
func NewSQLBuilder(baseQuery, countQuery string, params *PaginationParams) *SQLBuilder {
	return &SQLBuilder{
		baseQuery:  baseQuery,
		countQuery: countQuery,
		params:     params,
		bindings:   []interface{}{},
		paramIndex: 1,
	}
}

// AddFilter adds a filter condition
func (b *SQLBuilder) AddFilter(column string, value interface{}) *SQLBuilder {
	if value != nil && value != "" {
		if !strings.Contains(b.baseQuery, "WHERE") {
			b.baseQuery += " WHERE"
		} else {
			b.baseQuery += " AND"
		}
		b.baseQuery += fmt.Sprintf(" %s = ?", column)
		b.bindings = append(b.bindings, value)
		b.paramIndex++
	}
	return b
}

// AddSort adds sorting
func (b *SQLBuilder) AddSort(defaultColumn string) *SQLBuilder {
	sortColumn := b.params.SortBy
	if sortColumn == "" {
		sortColumn = defaultColumn
	}

	// Validate sort column to prevent SQL injection
	allowedColumns := map[string]bool{
		"id":         true,
		"created_at": true,
		"updated_at": true,
		"name":       true,
		"level":      true,
		"status":     true,
	}

	if !allowedColumns[sortColumn] {
		sortColumn = defaultColumn
	}

	b.baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortColumn, b.params.SortDir)
	return b
}

// AddPagination adds limit and offset
func (b *SQLBuilder) AddPagination() *SQLBuilder {
	b.baseQuery += " LIMIT ? OFFSET ?"
	b.bindings = append(b.bindings, b.params.Limit, b.params.GetOffset())
	return b
}

// Build returns the final query and bindings
func (b *SQLBuilder) Build() (string, []interface{}) {
	return b.baseQuery, b.bindings
}

// GetCountQuery returns the count query
func (b *SQLBuilder) GetCountQuery() string {
	return b.countQuery
}

// PaginatedQuery executes a paginated database query
type PaginatedQuery struct {
	DB         QueryExecutor
	BaseQuery  string
	CountQuery string
	Params     *PaginationParams
	ScanFunc   func(rows Scanner) (interface{}, error)
}

// QueryExecutor interface for database operations
type QueryExecutor interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

// Rows interface for database rows
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
}

// Row interface for single database row
type Row interface {
	Scan(dest ...interface{}) error
}

// Scanner interface for row scanning
type Scanner interface {
	Scan(dest ...interface{}) error
}

// Execute runs the paginated query
func (pq *PaginatedQuery) Execute(ctx context.Context) (*PageResult, error) {
	// Build query
	builder := NewSQLBuilder(pq.BaseQuery, pq.CountQuery, pq.Params)

	// Add filters from params
	for key, value := range pq.Params.Filters {
		builder.AddFilter(key, value)
	}

	// Add sorting and pagination
	builder.AddSort("created_at").AddPagination()

	// Execute count query
	var total int64
	err := pq.DB.QueryRowContext(ctx, builder.GetCountQuery()).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count results: %w", err)
	}

	// Execute main query
	query, bindings := builder.Build()
	rows, err := pq.DB.QueryContext(ctx, query, bindings...)
	if err != nil {
		return nil, fmt.Errorf("failed to query results: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []interface{}
	for rows.Next() {
		item, err := pq.ScanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, item)
	}

	return NewPageResult(results, pq.Params, total), nil
}

// Links generates pagination links for API responses
type Links struct {
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Self  string `json:"self"`
}

// GenerateLinks creates pagination links
func GenerateLinks(baseURL string, params *PaginationParams, totalPages int) Links {
	links := Links{
		Self: fmt.Sprintf("%s?page=%d&limit=%d", baseURL, params.Page, params.Limit),
	}

	// First page
	if params.Page > 1 {
		links.First = fmt.Sprintf("%s?page=1&limit=%d", baseURL, params.Limit)
	}

	// Last page
	if totalPages > 0 && params.Page < totalPages {
		links.Last = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, totalPages, params.Limit)
	}

	// Previous page
	if params.Page > 1 {
		links.Prev = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, params.Page-1, params.Limit)
	}

	// Next page
	if params.Page < totalPages {
		links.Next = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, params.Page+1, params.Limit)
	}

	return links
}

// WritePaginationHeaders writes pagination info to HTTP headers
func WritePaginationHeaders(w http.ResponseWriter, info PageInfo) {
	w.Header().Set("X-Pagination-Page", strconv.Itoa(info.Page))
	w.Header().Set("X-Pagination-Limit", strconv.Itoa(info.Limit))
	w.Header().Set("X-Pagination-Total", strconv.FormatInt(info.Total, 10))
	w.Header().Set("X-Pagination-Total-Pages", strconv.Itoa(info.TotalPages))

	if info.HasMore {
		w.Header().Set("X-Pagination-Has-More", "true")
	} else {
		w.Header().Set("X-Pagination-Has-More", "false")
	}
}

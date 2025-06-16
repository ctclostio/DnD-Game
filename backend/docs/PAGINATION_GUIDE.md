# Pagination Implementation Guide

## Overview

This guide covers the pagination implementation for the D&D Game API, including offset-based pagination, cursor-based pagination, and best practices for handling large datasets.

## Pagination Strategies

### 1. Offset-Based Pagination

Traditional pagination using page numbers and limits.

**Pros:**
- Simple to implement and understand
- Allows jumping to specific pages
- Easy to display total pages

**Cons:**
- Performance degrades with large offsets
- Inconsistent results if data changes
- Not suitable for real-time data

**When to use:**
- Static or slowly changing data
- When users need to jump to specific pages
- Total count is important

### 2. Cursor-Based Pagination

Uses a pointer (cursor) to track position in the dataset.

**Pros:**
- Consistent results even with data changes
- Better performance for large datasets
- Ideal for infinite scrolling

**Cons:**
- Can't jump to specific pages
- More complex to implement
- Harder to show progress

**When to use:**
- Real-time or frequently changing data
- Mobile apps with infinite scrolling
- Large datasets where performance matters

## Implementation

### Basic Usage

#### Offset-Based Pagination

```go
// Handler implementation
func (h *Handler) GetCharacters(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := auth.GetUserID(ctx)

    // Parse pagination parameters from query string
    params := pagination.FromRequest(r)
    
    // Validate parameters
    if err := params.Validate(); err != nil {
        response.Error(w, r, errors.NewValidationError(err.Error()))
        return
    }

    // Get paginated results
    result, err := h.characterService.GetCharactersPaginated(ctx, userID, params)
    if err != nil {
        response.Error(w, r, errors.Wrap(err, "failed to get characters"))
        return
    }

    // Write pagination headers
    pagination.WritePaginationHeaders(w, result.Pagination)

    // Return paginated response
    response.Success(w, r, result)
}
```

#### Service Implementation

```go
func (s *CharacterService) GetCharactersPaginated(ctx context.Context, userID string, params *pagination.PaginationParams) (*pagination.PageResult, error) {
    // Build base query
    baseQuery := `
        SELECT id, name, class, level, race, created_at
        FROM characters
        WHERE user_id = ?`
    
    countQuery := `
        SELECT COUNT(*)
        FROM characters
        WHERE user_id = ?`

    args := []interface{}{userID}

    // Apply filters
    if class := params.Filters["class"]; class != nil {
        baseQuery += " AND class = ?"
        countQuery += " AND class = ?"
        args = append(args, class)
    }

    // Get total count
    var total int64
    err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
    if err != nil {
        return nil, err
    }

    // Apply sorting
    sortColumn := "created_at"
    if params.SortBy != "" && isValidSortColumn(params.SortBy) {
        sortColumn = params.SortBy
    }
    baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortColumn, params.SortDir)

    // Apply pagination
    baseQuery += " LIMIT ? OFFSET ?"
    args = append(args, params.Limit, params.GetOffset())

    // Execute query
    var characters []*Character
    err = s.db.SelectContext(ctx, &characters, baseQuery, args...)
    if err != nil {
        return nil, err
    }

    return pagination.NewPageResult(characters, params, total), nil
}
```

### Cursor-Based Pagination

```go
// Handler for cursor pagination
func (h *Handler) GetCharactersCursor(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := auth.GetUserID(ctx)

    params := pagination.FromRequest(r)
    
    // Get cursor-paginated results
    result, err := h.characterService.GetCharactersCursor(ctx, userID, params)
    if err != nil {
        response.Error(w, r, errors.Wrap(err, "failed to get characters"))
        return
    }

    response.Success(w, r, result)
}

// Service implementation
func (s *CharacterService) GetCharactersCursor(ctx context.Context, userID string, params *pagination.PaginationParams) (*pagination.CursorResult, error) {
    query := `
        SELECT id, name, class, level, created_at
        FROM characters
        WHERE user_id = ?`
    
    args := []interface{}{userID}

    // Handle cursor
    if params.Cursor != "" {
        cursor, err := pagination.DecodeCursor(params.Cursor)
        if err != nil {
            return nil, err
        }

        // Add cursor condition
        query += " AND (created_at, id) > (?, ?)"
        args = append(args, cursor.Timestamp, cursor.ID)
    }

    // Order by cursor fields
    query += " ORDER BY created_at ASC, id ASC LIMIT ?"
    args = append(args, params.Limit+1) // Get one extra to check hasMore

    // Execute query
    var characters []*Character
    err := s.db.SelectContext(ctx, &characters, query, args...)
    if err != nil {
        return nil, err
    }

    // Check if there are more results
    hasMore := len(characters) > params.Limit
    if hasMore {
        characters = characters[:params.Limit]
    }

    // Create next cursor
    var nextCursor *pagination.Cursor
    if hasMore && len(characters) > 0 {
        last := characters[len(characters)-1]
        nextCursor = &pagination.Cursor{
            ID:        last.ID,
            Timestamp: last.CreatedAt,
        }
    }

    return pagination.NewCursorResult(characters, params, nextCursor, nil), nil
}
```

## API Design

### Request Parameters

#### Offset-Based
```
GET /api/characters?page=2&limit=20&sort_by=level&sort_dir=desc&filter_class=wizard
```

Parameters:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)
- `sort_by`: Field to sort by
- `sort_dir`: Sort direction (asc/desc)
- `filter_*`: Filter parameters

#### Cursor-Based
```
GET /api/characters/feed?cursor=eyJpZCI6IjEyMyIsInRpbWVzdGFtcCI6IjIwMjQtMDEtMDEifQ&limit=20
```

Parameters:
- `cursor`: Opaque cursor string
- `limit`: Items per page
- `sort_dir`: Sort direction (affects cursor direction)

### Response Format

#### Offset-Based Response
```json
{
  "data": [
    {
      "id": "123",
      "name": "Gandalf",
      "class": "Wizard",
      "level": 20
    }
  ],
  "pagination": {
    "page": 2,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "has_more": true
  }
}
```

#### Cursor-Based Response
```json
{
  "data": [
    {
      "id": "123",
      "name": "Gandalf",
      "class": "Wizard",
      "level": 20
    }
  ],
  "pagination": {
    "limit": 20,
    "has_more": true,
    "next_cursor": "eyJpZCI6IjQ1NiIsInRpbWVzdGFtcCI6IjIwMjQtMDEtMDIifQ",
    "prev_cursor": "eyJpZCI6IjEwMCIsInRpbWVzdGFtcCI6IjIwMjMtMTItMzEifQ"
  }
}
```

### HTTP Headers

The API also returns pagination info in headers:

```
X-Pagination-Page: 2
X-Pagination-Limit: 20
X-Pagination-Total: 150
X-Pagination-Total-Pages: 8
X-Pagination-Has-More: true
Link: <https://api.example.com/characters?page=1>; rel="first",
      <https://api.example.com/characters?page=8>; rel="last",
      <https://api.example.com/characters?page=1>; rel="prev",
      <https://api.example.com/characters?page=3>; rel="next"
```

## Performance Optimization

### 1. Database Indexes

Ensure proper indexes for pagination queries:

```sql
-- For offset pagination on characters
CREATE INDEX idx_characters_user_created ON characters(user_id, created_at);
CREATE INDEX idx_characters_user_level ON characters(user_id, level);
CREATE INDEX idx_characters_user_class ON characters(user_id, class);

-- For cursor pagination
CREATE INDEX idx_characters_created_id ON characters(created_at, id);

-- For filtering
CREATE INDEX idx_characters_class_level ON characters(class, level);
```

### 2. Query Optimization

#### Avoid COUNT(*) for Large Tables

```go
// Instead of exact count, use estimate for large tables
func getEstimatedCount(ctx context.Context, db *sql.DB, table string) (int64, error) {
    var count int64
    query := `
        SELECT reltuples::BIGINT AS estimate
        FROM pg_class
        WHERE relname = ?`
    err := db.QueryRowContext(ctx, query, table).Scan(&count)
    return count, err
}
```

#### Use Covering Indexes

```sql
-- Covering index includes all needed columns
CREATE INDEX idx_characters_list ON characters(user_id, created_at) 
INCLUDE (id, name, class, level);
```

### 3. Caching Strategy

```go
// Cache paginated results
func (s *Service) GetCharactersCached(ctx context.Context, userID string, params *pagination.PaginationParams) (*pagination.PageResult, error) {
    // Generate cache key including pagination params
    cacheKey := fmt.Sprintf("characters:user:%s:page:%d:limit:%d:sort:%s:%s",
        userID, params.Page, params.Limit, params.SortBy, params.SortDir)

    // Try cache first
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached.(*pagination.PageResult), nil
    }

    // Get from database
    result, err := s.GetCharactersPaginated(ctx, userID, params)
    if err != nil {
        return nil, err
    }

    // Cache for 5 minutes
    s.cache.Set(ctx, cacheKey, result, 5*time.Minute)
    return result, nil
}
```

## Client Implementation

### JavaScript/React

```javascript
// Pagination hook
function usePagination(endpoint, options = {}) {
  const [data, setData] = useState([]);
  const [pagination, setPagination] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchPage = useCallback(async (page = 1) => {
    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams({
        page,
        limit: options.limit || 20,
        ...options.filters
      });

      const response = await fetch(`${endpoint}?${params}`, {
        headers: {
          'Authorization': `Bearer ${getToken()}`
        }
      });

      if (!response.ok) throw new Error('Failed to fetch');

      const result = await response.json();
      setData(result.data);
      setPagination(result.pagination);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  }, [endpoint, options]);

  return { data, pagination, loading, error, fetchPage };
}

// Usage in component
function CharacterList() {
  const { data, pagination, loading, fetchPage } = usePagination('/api/characters', {
    limit: 20,
    filters: { class: 'wizard' }
  });

  useEffect(() => {
    fetchPage(1);
  }, []);

  return (
    <div>
      {loading && <Spinner />}
      {error && <ErrorMessage error={error} />}
      
      <ul>
        {data.map(character => (
          <li key={character.id}>{character.name}</li>
        ))}
      </ul>

      <Pagination
        current={pagination.page}
        total={pagination.total_pages}
        onChange={fetchPage}
      />
    </div>
  );
}
```

### Infinite Scroll with Cursor

```javascript
function CharacterFeed() {
  const [characters, setCharacters] = useState([]);
  const [cursor, setCursor] = useState(null);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);

  const loadMore = useCallback(async () => {
    if (loading || !hasMore) return;

    setLoading(true);
    try {
      const params = new URLSearchParams({ limit: 20 });
      if (cursor) params.append('cursor', cursor);

      const response = await fetch(`/api/characters/feed?${params}`);
      const result = await response.json();

      setCharacters(prev => [...prev, ...result.data]);
      setCursor(result.pagination.next_cursor);
      setHasMore(result.pagination.has_more);
    } finally {
      setLoading(false);
    }
  }, [cursor, hasMore, loading]);

  return (
    <InfiniteScroll
      dataLength={characters.length}
      next={loadMore}
      hasMore={hasMore}
      loader={<Spinner />}
    >
      {characters.map(char => (
        <CharacterCard key={char.id} character={char} />
      ))}
    </InfiniteScroll>
  );
}
```

## Best Practices

### 1. Consistent Parameter Names

Always use the same parameter names across all endpoints:
- `page`, `limit` for offset pagination
- `cursor`, `limit` for cursor pagination
- `sort_by`, `sort_dir` for sorting
- `filter_*` prefix for filters

### 2. Reasonable Limits

```go
const (
    DefaultLimit = 20
    MaxLimit     = 100
    MinLimit     = 1
)

func validateLimit(limit int) int {
    if limit < MinLimit {
        return DefaultLimit
    }
    if limit > MaxLimit {
        return MaxLimit
    }
    return limit
}
```

### 3. Error Handling

```go
// Handle common pagination errors
func handlePaginationError(err error) error {
    if err == sql.ErrNoRows {
        return errors.NewNotFoundError("no results found")
    }
    if strings.Contains(err.Error(), "invalid page") {
        return errors.NewValidationError("invalid page number")
    }
    return errors.Wrap(err, "pagination error")
}
```

### 4. Security Considerations

```go
// Prevent SQL injection in sort columns
func isValidSortColumn(column string) bool {
    validColumns := map[string]bool{
        "id":         true,
        "name":       true,
        "created_at": true,
        "updated_at": true,
        "level":      true,
        "class":      true,
    }
    return validColumns[column]
}

// Sanitize filter values
func sanitizeFilters(filters map[string]interface{}) map[string]interface{} {
    cleaned := make(map[string]interface{})
    for key, value := range filters {
        // Only allow alphanumeric keys
        if isAlphanumeric(key) {
            cleaned[key] = sanitizeValue(value)
        }
    }
    return cleaned
}
```

### 5. Documentation

Always document pagination in your API:

```yaml
# OpenAPI/Swagger example
/api/characters:
  get:
    parameters:
      - name: page
        in: query
        description: Page number (starting from 1)
        schema:
          type: integer
          minimum: 1
          default: 1
      - name: limit
        in: query
        description: Number of items per page
        schema:
          type: integer
          minimum: 1
          maximum: 100
          default: 20
    responses:
      200:
        description: Paginated list of characters
        headers:
          X-Pagination-Total:
            description: Total number of items
            schema:
              type: integer
        content:
          application/json:
            schema:
              type: object
              properties:
                data:
                  type: array
                  items:
                    $ref: '#/components/schemas/Character'
                pagination:
                  $ref: '#/components/schemas/PaginationInfo'
```

## Testing

### Unit Tests

```go
func TestPagination(t *testing.T) {
    t.Run("ValidateParams", func(t *testing.T) {
        params := &PaginationParams{
            Page:  0,
            Limit: 200,
        }
        
        err := params.Validate()
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "page must be >= 1")
    })

    t.Run("GetOffset", func(t *testing.T) {
        params := &PaginationParams{
            Page:  3,
            Limit: 20,
        }
        
        assert.Equal(t, 40, params.GetOffset())
    })

    t.Run("CursorEncodeDecode", func(t *testing.T) {
        cursor := &Cursor{
            ID:        "123",
            Timestamp: time.Now(),
        }
        
        encoded := EncodeCursor(cursor)
        decoded, err := DecodeCursor(encoded)
        
        assert.NoError(t, err)
        assert.Equal(t, cursor.ID, decoded.ID)
    })
}
```

### Integration Tests

```go
func TestCharactersPagination(t *testing.T) {
    // Setup test data
    db := setupTestDB(t)
    defer db.Close()
    
    userID := createTestUser(t, db)
    for i := 0; i < 50; i++ {
        createTestCharacter(t, db, userID, fmt.Sprintf("Character %d", i))
    }

    // Test first page
    params := &PaginationParams{Page: 1, Limit: 20}
    result, err := GetCharactersPaginated(context.Background(), userID, params)
    
    assert.NoError(t, err)
    assert.Len(t, result.Data, 20)
    assert.Equal(t, int64(50), result.Pagination.Total)
    assert.True(t, result.Pagination.HasMore)

    // Test last page
    params.Page = 3
    result, err = GetCharactersPaginated(context.Background(), userID, params)
    
    assert.NoError(t, err)
    assert.Len(t, result.Data, 10)
    assert.False(t, result.Pagination.HasMore)
}
```

## Summary

The pagination implementation provides:

1. **Flexible pagination options** - Both offset and cursor-based
2. **Consistent API design** - Standardized parameters and responses
3. **Performance optimization** - Proper indexes and caching
4. **Security** - Input validation and SQL injection prevention
5. **Developer-friendly** - Clear documentation and examples

Choose offset-based pagination for traditional interfaces with page numbers, and cursor-based pagination for modern infinite-scroll interfaces or real-time data.
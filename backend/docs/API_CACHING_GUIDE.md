# API Response Caching Strategy Guide

## Overview

This guide explains the comprehensive API response caching strategy implemented for the D&D Game backend, including HTTP response caching, data-level caching, and cache invalidation patterns.

## Architecture

### Three-Tier Caching Strategy

1. **HTTP Response Cache**: Full response caching at the middleware level
2. **Service-Level Cache**: Data object caching in services
3. **Database Query Cache**: Query result caching (via Redis)

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
┌──────▼──────┐
│  Nginx CDN  │ ← Static assets, long-term cache
└──────┬──────┘
       │
┌──────▼──────┐
│HTTP Response│ ← Full response cache (Redis)
│   Cache     │
└──────┬──────┘
       │
┌──────▼──────┐
│   Service   │ ← Object-level cache (Redis)
│    Cache    │
└──────┬──────┘
       │
┌──────▼──────┐
│  Database   │ ← Query cache, connection pool
└─────────────┘
```

## HTTP Response Caching

### Middleware Configuration

```go
// In your router setup
func setupRoutes(r *mux.Router, deps *Dependencies) {
    // Create cache middleware
    cacheMiddleware := middleware.NewCacheMiddleware(
        deps.Cache,
        &middleware.CacheConfig{
            DefaultTTL:           5 * time.Minute,
            MaxBodySize:          1 * 1024 * 1024, // 1MB
            IncludeQuery:         true,
            IncludeHeaders:       []string{"Accept", "Accept-Language"},
            ExcludeRoutes:        []string{"/api/auth/", "/api/ws/", "/health"},
            CacheableStatusCodes: []int{http.StatusOK, http.StatusNoContent},
            StaleWhileRevalidate: 1 * time.Minute,
        },
        deps.Logger,
    )

    // Create invalidation middleware
    invalidationMiddleware := middleware.NewInvalidationMiddleware(
        deps.Cache,
        deps.Logger,
    )

    // Apply middlewares
    r.Use(cacheMiddleware.Handler)
    r.Use(invalidationMiddleware.Handler)
}
```

### Cache-Control Headers

Set appropriate cache headers in your handlers:

```go
// Long cache for static data
func (h *Handler) GetRaces(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour
    // ... handler logic
}

// Short cache for dynamic data
func (h *Handler) GetCharacter(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "private, max-age=300") // 5 minutes
    // ... handler logic
}

// No cache for sensitive data
func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    // ... handler logic
}
```

### ETags for Conditional Requests

```go
func (h *Handler) GetCharacter(w http.ResponseWriter, r *http.Request) {
    character, err := h.service.GetCharacter(ctx, characterID)
    if err != nil {
        resp.Error(w, r, err)
        return
    }

    // Generate ETag based on content
    etag := generateETag(character)
    w.Header().Set("ETag", etag)

    // Check if client has current version
    if clientETag := r.Header.Get("If-None-Match"); clientETag == etag {
        w.WriteHeader(http.StatusNotModified)
        return
    }

    resp.Success(w, r, character)
}

func generateETag(data interface{}) string {
    bytes, _ := json.Marshal(data)
    hash := sha256.Sum256(bytes)
    return fmt.Sprintf(`"%x"`, hash[:8])
}
```

## Service-Level Caching

### Cache Service Integration

```go
type CharacterService struct {
    repo  CharacterRepository
    cache *cache.CacheService
    log   *logger.LoggerV2
}

func (s *CharacterService) GetCharacterByID(ctx context.Context, id string) (*models.Character, error) {
    // Try cache first
    if s.cache != nil {
        character, err := s.cache.GetCharacter(ctx, id)
        if err == nil && character != nil {
            return character, nil
        }
    }

    // Cache miss - get from database
    character, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Update cache asynchronously
    if s.cache != nil && character != nil {
        go func() {
            ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
            defer cancel()
            s.cache.SetCharacter(ctx, character)
        }()
    }

    return character, nil
}

func (s *CharacterService) UpdateCharacter(ctx context.Context, character *models.Character) error {
    // Update database
    if err := s.repo.Update(ctx, character); err != nil {
        return err
    }

    // Invalidate cache
    if s.cache != nil {
        s.cache.InvalidateUser(ctx, character.UserID)
    }

    return nil
}
```

### Cache Warming

Preload frequently accessed data:

```go
func warmupCache(ctx context.Context, services *Services) {
    warmer := cache.NewCacheWarmer(
        services.Cache,
        logger,
        30 * time.Minute, // Warm every 30 minutes
    )

    warmupFunc := func(ctx context.Context) (map[string][]interface{}, error) {
        data := make(map[string][]interface{})

        // Get active sessions
        sessions, err := services.GameSession.GetActiveSessions(ctx)
        if err == nil {
            items := make([]interface{}, len(sessions))
            for i, s := range sessions {
                items[i] = s
            }
            data["sessions"] = items
        }

        // Get recently active characters
        characters, err := services.Character.GetRecentlyActive(ctx, 24*time.Hour)
        if err == nil {
            items := make([]interface{}, len(characters))
            for i, c := range characters {
                items[i] = c
            }
            data["characters"] = items
        }

        return data, nil
    }

    go warmer.Start(ctx, warmupFunc)
}
```

## Cache Invalidation Patterns

### Automatic Invalidation

The invalidation middleware automatically clears related caches on mutations:

```go
// POST /api/characters/{id}
// Automatically invalidates:
// - response:*characters/{id}*
// - response:*characters*
// - character:{id}*
// - characters:list:*
```

### Manual Invalidation

For complex scenarios, manually invalidate caches:

```go
func (s *GameSessionService) EndSession(ctx context.Context, sessionID string) error {
    // End the session
    if err := s.repo.UpdateStatus(ctx, sessionID, "ended"); err != nil {
        return err
    }

    // Invalidate all related caches
    patterns := []string{
        fmt.Sprintf("session:%s*", sessionID),
        fmt.Sprintf("characters:session:%s*", sessionID),
        "sessions:active:*",
        "response:*sessions*",
    }

    for _, pattern := range patterns {
        s.cache.Invalidate(ctx, pattern)
    }

    return nil
}
```

## Cache Configuration by Endpoint

### Recommended TTLs

| Endpoint | TTL | Cache Type | Invalidation |
|----------|-----|------------|--------------|
| `/api/races` | 1 hour | Public | Manual |
| `/api/classes` | 1 hour | Public | Manual |
| `/api/characters` | 5 min | Private | On mutation |
| `/api/characters/{id}` | 10 min | Private | On mutation |
| `/api/game-sessions` | 2 min | Private | On mutation |
| `/api/game-sessions/{id}/state` | 30 sec | Private | On update |
| `/api/users/profile` | No cache | - | - |
| `/api/auth/*` | No cache | - | - |

### Implementation Examples

```go
// Static game data - long cache
func (h *Handler) GetClasses(w http.ResponseWriter, r *http.Request) {
    // Cache for 1 hour with public cache
    w.Header().Set("Cache-Control", "public, max-age=3600, s-maxage=7200")
    w.Header().Set("Vary", "Accept-Language") // Vary by language

    classes := h.gameDataService.GetClasses(r.Header.Get("Accept-Language"))
    resp.Success(w, r, classes)
}

// User-specific data - private cache
func (h *Handler) GetMyCharacters(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r.Context())
    
    // Cache for 5 minutes, private to user
    w.Header().Set("Cache-Control", "private, max-age=300")
    w.Header().Set("Vary", "Authorization")

    characters, err := h.characterService.GetUserCharacters(r.Context(), userID)
    if err != nil {
        resp.Error(w, r, err)
        return
    }

    resp.Success(w, r, characters)
}

// Real-time data - minimal cache
func (h *Handler) GetGameState(w http.ResponseWriter, r *http.Request) {
    sessionID := mux.Vars(r)["id"]
    
    // Very short cache for real-time data
    w.Header().Set("Cache-Control", "private, max-age=5, must-revalidate")
    
    state, err := h.gameService.GetCurrentState(r.Context(), sessionID)
    if err != nil {
        resp.Error(w, r, err)
        return
    }

    // Add timestamp for freshness
    response := map[string]interface{}{
        "state":     state,
        "timestamp": time.Now().Unix(),
    }

    resp.Success(w, r, response)
}
```

## Client-Side Caching

### JavaScript/React Implementation

```javascript
// API client with cache support
class APIClient {
    constructor() {
        this.cache = new Map();
    }

    async get(url, options = {}) {
        const cacheKey = this.getCacheKey(url, options);
        
        // Check if we have a fresh cache entry
        const cached = this.cache.get(cacheKey);
        if (cached && !this.isExpired(cached)) {
            return cached.data;
        }

        // Make request with conditional headers
        const headers = { ...options.headers };
        if (cached?.etag) {
            headers['If-None-Match'] = cached.etag;
        }

        const response = await fetch(url, { ...options, headers });

        // Handle 304 Not Modified
        if (response.status === 304 && cached) {
            cached.timestamp = Date.now();
            return cached.data;
        }

        // Parse and cache response
        const data = await response.json();
        const cacheControl = response.headers.get('Cache-Control');
        const maxAge = this.parseMaxAge(cacheControl);
        const etag = response.headers.get('ETag');

        if (maxAge > 0) {
            this.cache.set(cacheKey, {
                data,
                timestamp: Date.now(),
                maxAge: maxAge * 1000,
                etag,
            });
        }

        return data;
    }

    isExpired(entry) {
        return Date.now() - entry.timestamp > entry.maxAge;
    }

    parseMaxAge(cacheControl) {
        if (!cacheControl) return 0;
        const match = cacheControl.match(/max-age=(\d+)/);
        return match ? parseInt(match[1], 10) : 0;
    }

    getCacheKey(url, options) {
        return `${options.method || 'GET'}:${url}`;
    }

    invalidate(pattern) {
        for (const key of this.cache.keys()) {
            if (key.includes(pattern)) {
                this.cache.delete(key);
            }
        }
    }
}

// Usage in React components
function CharacterList() {
    const [characters, setCharacters] = useState([]);
    const api = useAPIClient();

    useEffect(() => {
        api.get('/api/characters')
            .then(setCharacters)
            .catch(console.error);
    }, []);

    const handleCharacterUpdate = async (character) => {
        await api.put(`/api/characters/${character.id}`, character);
        
        // Invalidate related caches
        api.invalidate('/api/characters');
        
        // Refresh list
        const updated = await api.get('/api/characters');
        setCharacters(updated);
    };

    return (
        // Component JSX
    );
}
```

## Monitoring and Metrics

### Cache Performance Metrics

```go
// Prometheus metrics
var (
    cacheHits = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_cache_hits_total",
            Help: "Total number of cache hits",
        },
        []string{"cache_type", "endpoint"},
    )
    
    cacheMisses = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_cache_misses_total",
            Help: "Total number of cache misses",
        },
        []string{"cache_type", "endpoint"},
    )
    
    cacheLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_cache_latency_seconds",
            Help:    "Cache operation latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation", "cache_type"},
    )
)
```

### Health Check Endpoint

```go
func (h *Handler) CacheHealth(w http.ResponseWriter, r *http.Request) {
    stats, err := h.cacheService.GetCacheStats(r.Context())
    if err != nil {
        resp.Error(w, r, err)
        return
    }

    // Calculate hit rate
    hits := cacheHits.Sum()
    misses := cacheMisses.Sum()
    hitRate := 0.0
    if total := hits + misses; total > 0 {
        hitRate = hits / total * 100
    }

    resp.Success(w, r, map[string]interface{}{
        "status":   "healthy",
        "hit_rate": fmt.Sprintf("%.2f%%", hitRate),
        "stats":    stats,
    })
}
```

## Best Practices

### 1. Cache Key Design

- Include all parameters that affect the response
- Use consistent naming patterns
- Include version in key for easy invalidation

```go
// Good cache key examples
"user:123:profile:v1"
"characters:user:123:page:1:limit:20:v1"
"session:456:state:v2"

// Bad cache key examples
"user-profile"  // No user ID
"characters"    // No pagination info
"session"       // No session ID
```

### 2. Cache Stampede Prevention

Prevent multiple requests from hitting the database when cache expires:

```go
func (s *Service) GetWithSingleflight(ctx context.Context, key string) (*Data, error) {
    // Use singleflight to deduplicate concurrent requests
    result, err, _ := s.group.Do(key, func() (interface{}, error) {
        // Check cache
        if data, err := s.cache.Get(ctx, key); err == nil {
            return data, nil
        }

        // Get from database
        data, err := s.repo.Get(ctx, key)
        if err != nil {
            return nil, err
        }

        // Update cache
        s.cache.Set(ctx, key, data, 5*time.Minute)
        return data, nil
    })

    if err != nil {
        return nil, err
    }

    return result.(*Data), nil
}
```

### 3. Graceful Degradation

Handle cache failures gracefully:

```go
func (s *Service) GetCharacter(ctx context.Context, id string) (*Character, error) {
    // Try cache, but don't fail if cache is down
    if s.cache != nil {
        if char, err := s.cache.GetCharacter(ctx, id); err == nil && char != nil {
            return char, nil
        } else if err != nil {
            s.log.Error().Err(err).Msg("Cache error, falling back to database")
        }
    }

    // Always fall back to database
    return s.repo.GetByID(ctx, id)
}
```

### 4. Cache Warming Strategy

```go
// Warm cache during off-peak hours
func scheduleCache/warming(cacheService *cache.CacheService) {
    // Run at 3 AM daily
    c := cron.New()
    c.AddFunc("0 3 * * *", func() {
        ctx := context.Background()
        
        // Warm popular characters
        popularChars, _ := getPopularCharacters(ctx)
        for _, char := range popularChars {
            cacheService.SetCharacter(ctx, char)
        }

        // Warm active sessions
        activeSessions, _ := getActiveSessions(ctx)
        for _, session := range activeSessions {
            cacheService.SetGameSession(ctx, session)
        }
    })
    c.Start()
}
```

## Testing Cache Behavior

### Unit Tests

```go
func TestCacheMiddleware(t *testing.T) {
    // Create mock cache
    mockCache := &MockCache{}
    middleware := NewCacheMiddleware(mockCache, nil, nil)

    // Test cache hit
    t.Run("CacheHit", func(t *testing.T) {
        // Set up cached response
        mockCache.On("Get", mock.Anything).Return(&CachedResponse{
            StatusCode: 200,
            Body:       []byte(`{"data":"cached"}`),
            CachedAt:   time.Now(),
            TTL:        5 * time.Minute,
        }, nil)

        // Make request
        req := httptest.NewRequest("GET", "/api/test", nil)
        rec := httptest.NewRecorder()

        handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            t.Error("Handler should not be called on cache hit")
        }))

        handler.ServeHTTP(rec, req)

        assert.Equal(t, 200, rec.Code)
        assert.Equal(t, "HIT", rec.Header().Get("X-Cache"))
    })
}
```

### Load Testing

```bash
# Test cache performance with hey
hey -n 10000 -c 100 -H "Authorization: Bearer $TOKEN" \
    https://api.example.com/api/characters

# Monitor cache hit rate during load test
watch -n 1 'curl -s localhost:8080/metrics | grep cache_hit'
```

## Summary

The comprehensive caching strategy provides:

1. **Multi-level caching** for optimal performance
2. **Automatic invalidation** on data mutations
3. **Flexible configuration** per endpoint
4. **Graceful degradation** on cache failures
5. **Performance monitoring** and metrics
6. **Client-side integration** support

This approach can reduce database load by 70-90% for read-heavy workloads while maintaining data consistency.
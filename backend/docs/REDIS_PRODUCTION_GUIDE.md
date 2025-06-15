# Redis Production Configuration Guide

## Overview

This guide covers Redis integration for caching and async job processing in the D&D Game application, including setup, best practices, and monitoring.

## Redis Configuration

### Environment Variables

```bash
# Redis Connection
REDIS_HOST=your-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=your-secure-password
REDIS_DB=0

# Optional: Redis Sentinel for HA
REDIS_SENTINEL_HOSTS=sentinel1:26379,sentinel2:26379,sentinel3:26379
REDIS_SENTINEL_MASTER=mymaster
```

### Docker Compose Setup

For local development and testing:

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass yourpassword
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "yourpassword", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis-commander:
    image: rediscommander/redis-commander:latest
    environment:
      - REDIS_HOSTS=local:redis:6379:0:yourpassword
    ports:
      - "8081:8081"
    depends_on:
      - redis

volumes:
  redis_data:
```

## Caching Implementation

### 1. Initialize Redis Cache

```go
import (
    "github.com/ctclostio/DnD-Game/backend/internal/cache"
    "github.com/ctclostio/DnD-Game/backend/internal/config"
)

func initializeCache(cfg *config.Config, logger *logger.LoggerV2) (*cache.RedisClient, error) {
    // Create Redis client
    redisClient, err := cache.NewRedisClient(&cfg.Redis, logger)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    // Verify connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := redisClient.Ping(ctx); err != nil {
        return nil, fmt.Errorf("Redis ping failed: %w", err)
    }

    return redisClient, nil
}
```

### 2. Create Service-Specific Caches

```go
// Character cache example
type CharacterCache struct {
    cache *cache.Cache
}

func NewCharacterCache(redisClient *cache.RedisClient) *CharacterCache {
    return &CharacterCache{
        cache: cache.NewCache(redisClient, "character", 1*time.Hour),
    }
}

func (cc *CharacterCache) Get(ctx context.Context, characterID string) (*models.Character, error) {
    var character models.Character
    err := cc.cache.GetJSON(ctx, characterID, &character)
    if err == redis.Nil {
        return nil, nil // Cache miss
    }
    return &character, err
}

func (cc *CharacterCache) Set(ctx context.Context, character *models.Character) error {
    return cc.cache.SetJSON(ctx, character.ID, character)
}

func (cc *CharacterCache) Delete(ctx context.Context, characterID string) error {
    return cc.cache.Delete(ctx, characterID)
}

func (cc *CharacterCache) InvalidateUser(ctx context.Context, userID string) error {
    // Invalidate all characters for a user
    return cc.cache.Invalidate(ctx, fmt.Sprintf("user:%s:*", userID))
}
```

### 3. Cache-Aside Pattern Implementation

```go
func (s *CharacterService) GetCharacterByID(ctx context.Context, id string) (*models.Character, error) {
    // Try cache first
    if s.cache != nil {
        character, err := s.cache.Get(ctx, id)
        if err != nil {
            s.logger.Error().Err(err).Msg("Cache read failed")
            // Continue to database on cache error
        } else if character != nil {
            s.logger.Debug().Str("character_id", id).Msg("Cache hit")
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
            
            if err := s.cache.Set(ctx, character); err != nil {
                s.logger.Error().Err(err).Msg("Failed to update cache")
            }
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
        if err := s.cache.Delete(ctx, character.ID); err != nil {
            s.logger.Error().Err(err).Msg("Failed to invalidate cache")
        }
    }

    return nil
}
```

## Job Queue Implementation

### 1. Initialize Job Queue

```go
import (
    "github.com/ctclostio/DnD-Game/backend/internal/jobs"
)

func initializeJobQueue(cfg *config.Config, logger *logger.LoggerV2, services *Services) (*jobs.JobQueue, error) {
    // Create job queue
    jobQueue, err := jobs.NewJobQueue(&cfg.Redis, logger)
    if err != nil {
        return nil, fmt.Errorf("failed to create job queue: %w", err)
    }

    // Create and register job handlers
    handlers := jobs.NewJobHandlers(
        logger,
        services.AI,
        services.Email,
        services.Character,
        services.Campaign,
        services.Export,
        services.Cleanup,
    )
    
    handlers.RegisterAll(jobQueue)

    // Start processing jobs
    go func() {
        if err := jobQueue.Start(); err != nil {
            logger.Fatal().Err(err).Msg("Job queue failed to start")
        }
    }()

    return jobQueue, nil
}
```

### 2. Enqueue Jobs

```go
// Example: Enqueue AI content generation
func (h *Handler) GenerateNPC(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := auth.GetUserID(ctx)

    var req struct {
        Prompt  string `json:"prompt"`
        Context string `json:"context"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        resp.Error(w, r, errors.Wrap(err, "invalid request"))
        return
    }

    // Enqueue job for async processing
    payload := jobs.AIGenerationPayload{
        UserID:  userID,
        Type:    "npc",
        Prompt:  req.Prompt,
        Context: req.Context,
    }

    opts := jobs.JobOptions{
        Queue:     jobs.QueueDefault,
        MaxRetry:  3,
        Retention: 24 * time.Hour,
    }

    info, err := h.jobQueue.Enqueue(ctx, jobs.JobTypeAIContentGeneration, payload, opts)
    if err != nil {
        resp.Error(w, r, errors.Wrap(err, "failed to enqueue job"))
        return
    }

    resp.Success(w, r, map[string]interface{}{
        "job_id": info.ID,
        "status": "queued",
        "message": "NPC generation started",
    })
}
```

### 3. Schedule Recurring Jobs

```go
func scheduleRecurringJobs(jobQueue *jobs.JobQueue, logger *logger.LoggerV2) {
    // Schedule daily cleanup
    go func() {
        ticker := time.NewTicker(24 * time.Hour)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                ctx := context.Background()
                
                // Cleanup expired tokens
                payload := jobs.CleanupPayload{
                    Type:      "expired_tokens",
                    OlderThan: time.Now().Add(-7 * 24 * time.Hour),
                }
                
                if _, err := jobQueue.Enqueue(ctx, jobs.JobTypeCleanupExpired, payload); err != nil {
                    logger.Error().Err(err).Msg("Failed to enqueue cleanup job")
                }
            }
        }
    }()
}
```

## Monitoring and Health Checks

### 1. Redis Health Check

```go
func (h *Handler) RedisHealth(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Check Redis connectivity
    if err := h.redisClient.HealthCheck(ctx); err != nil {
        resp.Error(w, r, errors.Wrap(err, "Redis health check failed"))
        return
    }

    // Get pool stats
    stats := h.redisClient.Stats()
    
    // Get job queue stats
    queueStats, err := h.jobQueue.GetQueueStats()
    if err != nil {
        resp.Error(w, r, errors.Wrap(err, "Failed to get queue stats"))
        return
    }

    resp.Success(w, r, map[string]interface{}{
        "status": "healthy",
        "redis": map[string]interface{}{
            "connected":      true,
            "hits":          stats.Hits,
            "misses":        stats.Misses,
            "timeouts":      stats.Timeouts,
            "total_conns":   stats.TotalConns,
            "idle_conns":    stats.IdleConns,
            "stale_conns":   stats.StaleConns,
        },
        "queues": queueStats,
    })
}
```

### 2. Metrics Collection

```go
// Prometheus metrics example
var (
    cacheHits = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
        []string{"cache_name"},
    )
    
    cacheMisses = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total number of cache misses",
        },
        []string{"cache_name"},
    )
    
    jobsQueued = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "jobs_queued_total",
            Help: "Total number of jobs queued",
        },
        []string{"job_type", "queue"},
    )
    
    jobsProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "jobs_processed_total",
            Help: "Total number of jobs processed",
        },
        []string{"job_type", "status"},
    )
)
```

## Best Practices

### 1. Cache Key Naming

```go
// Use consistent, hierarchical key naming
const (
    // User-related keys
    KeyUserProfile      = "user:%s:profile"
    KeyUserCharacters   = "user:%s:characters"
    KeyUserSessions     = "user:%s:sessions"
    
    // Character-related keys
    KeyCharacter        = "character:%s"
    KeyCharacterSheet   = "character:%s:sheet"
    KeyCharacterInventory = "character:%s:inventory"
    
    // Session-related keys
    KeyGameSession      = "session:%s"
    KeySessionPlayers   = "session:%s:players"
    KeySessionState     = "session:%s:state"
)
```

### 2. Cache Invalidation Strategies

```go
// Invalidate related caches when data changes
func (s *CharacterService) UpdateCharacter(ctx context.Context, char *models.Character) error {
    // Update database
    if err := s.repo.Update(ctx, char); err != nil {
        return err
    }

    // Invalidate all related caches
    keys := []string{
        fmt.Sprintf(KeyCharacter, char.ID),
        fmt.Sprintf(KeyCharacterSheet, char.ID),
        fmt.Sprintf(KeyCharacterInventory, char.ID),
        fmt.Sprintf(KeyUserCharacters, char.UserID),
    }
    
    return s.cache.Delete(ctx, keys...)
}
```

### 3. Circuit Breaker for Redis

```go
import "github.com/sony/gobreaker"

type CacheWithBreaker struct {
    cache   *cache.Cache
    breaker *gobreaker.CircuitBreaker
}

func NewCacheWithBreaker(cache *cache.Cache) *CacheWithBreaker {
    st := gobreaker.Settings{
        Name:        "RedisCache",
        MaxRequests: 10,
        Interval:    10 * time.Second,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
    }
    
    return &CacheWithBreaker{
        cache:   cache,
        breaker: gobreaker.NewCircuitBreaker(st),
    }
}

func (c *CacheWithBreaker) Get(ctx context.Context, key string) (interface{}, error) {
    result, err := c.breaker.Execute(func() (interface{}, error) {
        return c.cache.Get(ctx, key)
    })
    
    if err == gobreaker.ErrOpenState {
        // Circuit is open, return nil to fallback to database
        return nil, nil
    }
    
    return result, err
}
```

## Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Set maxmemory limit in Redis config
   - Configure eviction policy: `maxmemory-policy allkeys-lru`
   - Monitor key expiration

2. **Connection Pool Exhaustion**
   - Increase pool size in Redis client config
   - Check for connection leaks
   - Monitor slow commands: `SLOWLOG GET`

3. **Job Processing Delays**
   - Check queue backlogs
   - Increase worker concurrency
   - Monitor job execution times

### Debug Commands

```bash
# Monitor Redis commands in real-time
redis-cli -a yourpassword MONITOR

# Check slow queries
redis-cli -a yourpassword SLOWLOG GET 10

# Get memory usage info
redis-cli -a yourpassword INFO memory

# Check connected clients
redis-cli -a yourpassword CLIENT LIST

# Get queue statistics (using asynq CLI)
asynq stats
asynq queue ls
asynq task ls --queue=default --state=pending
```

## Security Considerations

1. **Always use password authentication**
2. **Enable TLS for production Redis**
3. **Use firewall rules to restrict access**
4. **Regularly rotate Redis passwords**
5. **Monitor for suspicious patterns**
6. **Use separate Redis instances for cache vs jobs**

## Performance Tuning

### Redis Configuration

```conf
# /etc/redis/redis.conf

# Persistence (adjust based on needs)
save 900 1
save 300 10
save 60 10000

# Memory management
maxmemory 2gb
maxmemory-policy allkeys-lru

# Connection handling
tcp-keepalive 300
timeout 300

# Performance
tcp-backlog 511
databases 16

# Slow log
slowlog-log-slower-than 10000
slowlog-max-len 128
```

### Application-Level Optimizations

1. **Batch operations when possible**
2. **Use pipelining for multiple commands**
3. **Set appropriate TTLs for all keys**
4. **Compress large values before storing**
5. **Use Redis data structures efficiently**

## Summary

Redis integration provides:
- High-performance caching to reduce database load
- Async job processing for long-running tasks
- Distributed locking for coordination
- Session storage for horizontal scaling

Monitor closely in production and adjust configuration based on actual usage patterns.
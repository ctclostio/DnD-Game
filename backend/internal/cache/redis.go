package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/ctclostio/DnD-Game/backend/internal/config"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client *redis.Client
	logger *logger.LoggerV2
	config *config.RedisConfig
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.RedisConfig, log *logger.LoggerV2) (*RedisClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is required")
	}

	// Create Redis options
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		
		// Connection pool settings
		PoolSize:     50,                      // Maximum number of socket connections
		MinIdleConns: 10,                      // Minimum idle connections
		MaxRetries:   3,                        // Maximum retries before giving up
		DialTimeout:  5 * time.Second,          // Dial timeout
		ReadTimeout:  3 * time.Second,          // Read timeout
		WriteTimeout: 3 * time.Second,          // Write timeout
		PoolTimeout:  4 * time.Second,          // Amount of time to wait for a connection
		IdleTimeout:  5 * time.Minute,          // Close connections after remaining idle for this duration
		MaxConnAge:   30 * time.Minute,         // Close connections older than this duration
	}

	// Create client
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	rc := &RedisClient{
		client: client,
		logger: log,
		config: cfg,
	}

	if log != nil {
		log.Info().
			Str("host", cfg.Host).
			Int("port", cfg.Port).
			Int("db", cfg.DB).
			Msg("Successfully connected to Redis")
	}

	return rc, nil
}

// Close closes the Redis connection
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// Ping checks if Redis is accessible
func (rc *RedisClient) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Get retrieves a value from cache
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	val, err := rc.client.Get(ctx, key).Result()
	
	if rc.logger != nil {
		rc.logOperation(ctx, "GET", key, time.Since(start), err)
	}

	if err == redis.Nil {
		return "", nil // Key doesn't exist
	}
	return val, err
}

// Set stores a value in cache with expiration
func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	
	// Convert value to string if needed
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		// Serialize to JSON for complex types
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		val = string(jsonData)
	}

	err := rc.client.Set(ctx, key, val, expiration).Err()
	
	if rc.logger != nil {
		rc.logOperation(ctx, "SET", key, time.Since(start), err)
	}

	return err
}

// Delete removes a key from cache
func (rc *RedisClient) Delete(ctx context.Context, keys ...string) error {
	start := time.Now()
	err := rc.client.Del(ctx, keys...).Err()
	
	if rc.logger != nil {
		rc.logOperation(ctx, "DEL", fmt.Sprintf("%d keys", len(keys)), time.Since(start), err)
	}

	return err
}

// Exists checks if keys exist
func (rc *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return rc.client.Exists(ctx, keys...).Result()
}

// TTL returns the remaining time to live of a key
func (rc *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return rc.client.TTL(ctx, key).Result()
}

// GetJSON retrieves and unmarshals a JSON value
func (rc *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.Get(ctx, key)
	if err != nil {
		return err
	}
	
	if val == "" {
		return redis.Nil
	}

	return json.Unmarshal([]byte(val), dest)
}

// SetJSON marshals and stores a JSON value
func (rc *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.Set(ctx, key, value, expiration)
}

// Increment atomically increments a counter
func (rc *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return rc.client.Incr(ctx, key).Result()
}

// IncrementBy atomically increments a counter by a specific amount
func (rc *RedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return rc.client.IncrBy(ctx, key, value).Result()
}

// Expire sets a timeout on a key
func (rc *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rc.client.Expire(ctx, key, expiration).Err()
}

// FlushDB removes all keys from the current database
func (rc *RedisClient) FlushDB(ctx context.Context) error {
	return rc.client.FlushDB(ctx).Err()
}

// logOperation logs Redis operations
func (rc *RedisClient) logOperation(ctx context.Context, operation, key string, duration time.Duration, err error) {
	log := rc.logger.WithContext(ctx)
	
	event := log.Debug().
		Str("operation", operation).
		Str("key", key).
		Dur("duration", duration).
		Int64("duration_ms", duration.Milliseconds())

	if err != nil && err != redis.Nil {
		event.Err(err).Msg("Redis operation failed")
	} else if err == redis.Nil {
		event.Msg("Redis key not found")
	} else {
		event.Msg("Redis operation completed")
	}
}

// GetClient returns the underlying Redis client for advanced operations
func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}

// Stats returns connection pool statistics
func (rc *RedisClient) Stats() *redis.PoolStats {
	return rc.client.PoolStats()
}

// HealthCheck performs a health check on the Redis connection
func (rc *RedisClient) HealthCheck(ctx context.Context) error {
	// Set a test key with short expiration
	testKey := "health:check"
	testValue := time.Now().Unix()
	
	if err := rc.Set(ctx, testKey, testValue, 10*time.Second); err != nil {
		return fmt.Errorf("health check write failed: %w", err)
	}

	// Read it back
	val, err := rc.Get(ctx, testKey)
	if err != nil {
		return fmt.Errorf("health check read failed: %w", err)
	}

	if val == "" {
		return fmt.Errorf("health check read returned empty value")
	}

	return nil
}

// Lock implements a simple distributed lock using SET NX
type Lock struct {
	client     *RedisClient
	key        string
	value      string
	expiration time.Duration
}

// NewLock creates a new distributed lock
func (rc *RedisClient) NewLock(key string, expiration time.Duration) *Lock {
	return &Lock{
		client:     rc,
		key:        fmt.Sprintf("lock:%s", key),
		value:      fmt.Sprintf("%d", time.Now().UnixNano()),
		expiration: expiration,
	}
}

// Acquire attempts to acquire the lock
func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	return l.client.client.SetNX(ctx, l.key, l.value, l.expiration).Result()
}

// Release releases the lock if we own it
func (l *Lock) Release(ctx context.Context) error {
	// Use Lua script to ensure we only delete if we own the lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	
	return l.client.client.Eval(ctx, script, []string{l.key}, l.value).Err()
}

// Cache provides high-level caching operations
type Cache struct {
	client        *RedisClient
	defaultExpiry time.Duration
	keyPrefix     string
}

// NewCache creates a new cache instance
func NewCache(client *RedisClient, keyPrefix string, defaultExpiry time.Duration) *Cache {
	return &Cache{
		client:        client,
		keyPrefix:     keyPrefix,
		defaultExpiry: defaultExpiry,
	}
}

// makeKey creates a namespaced cache key
func (c *Cache) makeKey(key string) string {
	if c.keyPrefix != "" {
		return fmt.Sprintf("%s:%s", c.keyPrefix, key)
	}
	return key
}

// Get retrieves a cached value
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, c.makeKey(key))
}

// Set stores a value in cache
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	exp := c.defaultExpiry
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return c.client.Set(ctx, c.makeKey(key), value, exp)
}

// Delete removes a value from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Delete(ctx, c.makeKey(key))
}

// GetJSON retrieves and unmarshals a JSON value
func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	return c.client.GetJSON(ctx, c.makeKey(key), dest)
}

// SetJSON marshals and stores a JSON value
func (c *Cache) SetJSON(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	exp := c.defaultExpiry
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return c.client.SetJSON(ctx, c.makeKey(key), value, exp)
}

// Invalidate removes multiple keys by pattern
func (c *Cache) Invalidate(ctx context.Context, pattern string) error {
	// Scan for keys matching pattern
	keys := []string{}
	iter := c.client.client.Scan(ctx, 0, c.makeKey(pattern), 100).Iterator()
	
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	
	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Delete(ctx, keys...)
	}

	return nil
}
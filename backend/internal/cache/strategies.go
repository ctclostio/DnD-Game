package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// CacheStrategy defines cache behavior for different data types
type CacheStrategy interface {
	GetKey(id string, params ...string) string
	GetTTL() time.Duration
	GetInvalidationPatterns(id string) []string
}

// UserCacheStrategy handles user-related caching
type UserCacheStrategy struct{}

func (s *UserCacheStrategy) GetKey(id string, params ...string) string {
	if len(params) > 0 {
		return fmt.Sprintf("user:%s:%s", id, params[0])
	}
	return fmt.Sprintf("user:%s", id)
}

func (s *UserCacheStrategy) GetTTL() time.Duration {
	return 15 * time.Minute
}

func (s *UserCacheStrategy) GetInvalidationPatterns(id string) []string {
	return []string{
		fmt.Sprintf("user:%s*", id),
		fmt.Sprintf("characters:user:%s*", id),
		fmt.Sprintf("sessions:user:%s*", id),
	}
}

// CharacterCacheStrategy handles character caching
type CharacterCacheStrategy struct{}

func (s *CharacterCacheStrategy) GetKey(id string, params ...string) string {
	if len(params) > 0 {
		return fmt.Sprintf("character:%s:%s", id, params[0])
	}
	return fmt.Sprintf("character:%s", id)
}

func (s *CharacterCacheStrategy) GetTTL() time.Duration {
	return 10 * time.Minute
}

func (s *CharacterCacheStrategy) GetInvalidationPatterns(id string) []string {
	// Get character to find user ID for broader invalidation
	return []string{
		fmt.Sprintf("character:%s*", id),
		"characters:list:*",      // Invalidate all character lists
		"response:*characters*",  // Invalidate HTTP response cache
	}
}

// GameSessionCacheStrategy handles game session caching
type GameSessionCacheStrategy struct{}

func (s *GameSessionCacheStrategy) GetKey(id string, params ...string) string {
	if len(params) > 0 {
		return fmt.Sprintf("session:%s:%s", id, params[0])
	}
	return fmt.Sprintf("session:%s", id)
}

func (s *GameSessionCacheStrategy) GetTTL() time.Duration {
	return 5 * time.Minute // Shorter TTL for active game data
}

func (s *GameSessionCacheStrategy) GetInvalidationPatterns(id string) []string {
	return []string{
		fmt.Sprintf("session:%s*", id),
		"sessions:active:*",
		"sessions:list:*",
		"response:*sessions*",
	}
}

// CacheService provides high-level caching operations
type CacheService struct {
	client     *RedisClient
	logger     *logger.LoggerV2
	strategies map[string]CacheStrategy
}

// NewCacheService creates a new cache service
func NewCacheService(client *RedisClient, logger *logger.LoggerV2) *CacheService {
	return &CacheService{
		client: client,
		logger: logger,
		strategies: map[string]CacheStrategy{
			"user":      &UserCacheStrategy{},
			"character": &CharacterCacheStrategy{},
			"session":   &GameSessionCacheStrategy{},
		},
	}
}

// GetUser retrieves a cached user
func (cs *CacheService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	strategy := cs.strategies["user"]
	key := strategy.GetKey(userID)

	var user models.User
	err := cs.client.GetJSON(ctx, key, &user)
	if err != nil {
		return nil, err
	}

	cs.logCacheHit("user", userID)
	return &user, nil
}

// SetUser caches a user
func (cs *CacheService) SetUser(ctx context.Context, user *models.User) error {
	strategy := cs.strategies["user"]
	key := strategy.GetKey(user.ID)
	ttl := strategy.GetTTL()

	return cs.client.SetJSON(ctx, key, user, ttl)
}

// InvalidateUser removes user from cache
func (cs *CacheService) InvalidateUser(ctx context.Context, userID string) error {
	strategy := cs.strategies["user"]
	patterns := strategy.GetInvalidationPatterns(userID)
	
	for _, pattern := range patterns {
		cs.invalidatePattern(ctx, pattern)
	}
	
	return nil
}

// invalidatePattern deletes all keys matching the given pattern
func (cs *CacheService) invalidatePattern(ctx context.Context, pattern string) {
	keys, err := cs.getKeysByPattern(ctx, pattern)
	if err != nil {
		cs.logger.Error().Err(err).Str("pattern", pattern).Msg("Failed to get keys for invalidation")
		return
	}
	
	if len(keys) == 0 {
		return
	}
	
	if err := cs.client.Delete(ctx, keys...); err != nil {
		cs.logger.Error().Err(err).Str("pattern", pattern).Msg("Failed to delete keys")
		return
	}
	
	cs.logger.Debug().
		Str("pattern", pattern).
		Int("keys_deleted", len(keys)).
		Msg("Cache invalidated")
}

// GetCharacter retrieves a cached character
func (cs *CacheService) GetCharacter(ctx context.Context, characterID string) (*models.Character, error) {
	strategy := cs.strategies["character"]
	key := strategy.GetKey(characterID)

	var character models.Character
	err := cs.client.GetJSON(ctx, key, &character)
	if err != nil {
		return nil, err
	}

	cs.logCacheHit("character", characterID)
	return &character, nil
}

// SetCharacter caches a character
func (cs *CacheService) SetCharacter(ctx context.Context, character *models.Character) error {
	strategy := cs.strategies["character"]
	key := strategy.GetKey(character.ID)
	ttl := strategy.GetTTL()

	return cs.client.SetJSON(ctx, key, character, ttl)
}

// GetCharacterList retrieves a cached character list
func (cs *CacheService) GetCharacterList(ctx context.Context, userID string, filters ...string) ([]*models.Character, error) {
	filterKey := "all"
	if len(filters) > 0 {
		filterKey = filters[0]
	}

	key := fmt.Sprintf("characters:list:user:%s:filter:%s", userID, filterKey)

	var characters []*models.Character
	err := cs.client.GetJSON(ctx, key, &characters)
	if err != nil {
		return nil, err
	}

	cs.logCacheHit("character_list", userID)
	return characters, nil
}

// SetCharacterList caches a character list
func (cs *CacheService) SetCharacterList(ctx context.Context, userID string, characters []*models.Character, filters ...string) error {
	filterKey := "all"
	if len(filters) > 0 {
		filterKey = filters[0]
	}

	key := fmt.Sprintf("characters:list:user:%s:filter:%s", userID, filterKey)
	ttl := 5 * time.Minute // Shorter TTL for lists

	return cs.client.SetJSON(ctx, key, characters, ttl)
}

// GetGameSession retrieves a cached game session
func (cs *CacheService) GetGameSession(ctx context.Context, sessionID string) (*models.GameSession, error) {
	strategy := cs.strategies["session"]
	key := strategy.GetKey(sessionID)

	var session models.GameSession
	err := cs.client.GetJSON(ctx, key, &session)
	if err != nil {
		return nil, err
	}

	cs.logCacheHit("session", sessionID)
	return &session, nil
}

// SetGameSession caches a game session
func (cs *CacheService) SetGameSession(ctx context.Context, session *models.GameSession) error {
	strategy := cs.strategies["session"]
	key := strategy.GetKey(session.ID)
	ttl := strategy.GetTTL()

	err := cs.client.SetJSON(ctx, key, session, ttl)
	if err != nil {
		return err
	}

	// Also update active sessions cache if applicable
	if session.Status == "active" {
		activeKey := fmt.Sprintf("sessions:active:%s", session.ID)
		if err := cs.client.Set(ctx, activeKey, "1", ttl); err != nil {
			return fmt.Errorf("failed to cache active session: %w", err)
		}
	}

	return nil
}

// GetActiveSessionIDs retrieves cached active session IDs
func (cs *CacheService) GetActiveSessionIDs(ctx context.Context) ([]string, error) {
	pattern := "sessions:active:*"
	keys, err := cs.getKeysByPattern(ctx, pattern)
	if err != nil {
		return nil, err
	}
	
	return cs.extractSessionIDsFromKeys(keys), nil
}

// extractSessionIDsFromKeys extracts session IDs from cache keys
func (cs *CacheService) extractSessionIDsFromKeys(keys []string) []string {
	sessionIDs := make([]string, 0, len(keys))
	const sessionKeyPrefixParts = 3 // sessions:active:{id}
	
	for _, key := range keys {
		if id := extractIDFromKey(key, sessionKeyPrefixParts); id != "" {
			sessionIDs = append(sessionIDs, id)
		}
	}
	
	return sessionIDs
}

// extractIDFromKey extracts the ID from a cache key
func extractIDFromKey(key string, requiredParts int) string {
	parts := splitKey(key, ":")
	if len(parts) >= requiredParts {
		return parts[requiredParts-1]
	}
	return ""
}

// warmers maps data types to their warming functions
var warmers = map[string]string{
	"characters": "warmCharacters",
	"sessions":   "warmSessions",
}

// WarmCache pre-loads frequently accessed data
func (cs *CacheService) WarmCache(ctx context.Context, dataType string, items []interface{}) error {
	if _, supported := warmers[dataType]; !supported {
		return fmt.Errorf("unsupported data type: %s", dataType)
	}
	
	switch dataType {
	case "characters":
		cs.warmCharacters(ctx, items)
	case "sessions":
		cs.warmSessions(ctx, items)
	}
	
	cs.logger.Info().
		Str("data_type", dataType).
		Int("items_count", len(items)).
		Msg("Cache warmed")
	
	return nil
}

// warmCharacters warms the cache with character data
func (cs *CacheService) warmCharacters(ctx context.Context, items []interface{}) {
	cs.warmItems(ctx, items, cs.warmSingleCharacter)
}

// warmSessions warms the cache with session data
func (cs *CacheService) warmSessions(ctx context.Context, items []interface{}) {
	cs.warmItems(ctx, items, cs.warmSingleSession)
}

// warmItems is a generic function to warm cache with items
func (cs *CacheService) warmItems(ctx context.Context, items []interface{}, warmer func(context.Context, interface{}) error) {
	for _, item := range items {
		if err := warmer(ctx, item); err != nil {
			cs.logger.Error().Err(err).Msg("Failed to warm cache item")
		}
	}
}

// warmSingleCharacter warms cache with a single character
func (cs *CacheService) warmSingleCharacter(ctx context.Context, item interface{}) error {
	char, ok := item.(*models.Character)
	if !ok {
		return nil // Skip non-character items
	}
	
	if err := cs.SetCharacter(ctx, char); err != nil {
		return fmt.Errorf("failed to warm character %s: %w", char.ID, err)
	}
	
	return nil
}

// warmSingleSession warms cache with a single session
func (cs *CacheService) warmSingleSession(ctx context.Context, item interface{}) error {
	session, ok := item.(*models.GameSession)
	if !ok {
		return nil // Skip non-session items
	}
	
	if err := cs.SetGameSession(ctx, session); err != nil {
		return fmt.Errorf("failed to warm session %s: %w", session.ID, err)
	}
	
	return nil
}

// GetCacheStats returns cache statistics
func (cs *CacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := cs.client.GetClient().Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	// Parse relevant stats
	stats := map[string]interface{}{
		"raw_info": info,
		// Add parsed stats here
	}

	// Get memory info
	memInfo, err := cs.client.GetClient().Info(ctx, "memory").Result()
	if err == nil {
		stats["memory"] = memInfo
	}

	return stats, nil
}

// Helper methods

func (cs *CacheService) logCacheHit(dataType, id string) {
	if cs.logger != nil {
		cs.logger.Debug().
			Str("type", dataType).
			Str("id", id).
			Msg("Cache hit")
	}
}

func (cs *CacheService) getKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	iter := cs.client.GetClient().Scan(ctx, 0, pattern, 100).Iterator()
	
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	
	return keys, iter.Err()
}

func splitKey(key, delimiter string) []string {
	return strings.Split(key, delimiter)
}

// CacheWarmer runs periodic cache warming
type CacheWarmer struct {
	service  *CacheService
	logger   *logger.LoggerV2
	interval time.Duration
	stopCh   chan struct{}
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(service *CacheService, logger *logger.LoggerV2, interval time.Duration) *CacheWarmer {
	return &CacheWarmer{
		service:  service,
		logger:   logger,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the cache warming process
func (cw *CacheWarmer) Start(ctx context.Context, warmupFunc func(context.Context) (map[string][]interface{}, error)) {
	ticker := time.NewTicker(cw.interval)
	defer ticker.Stop()

	// Initial warmup
	cw.performWarmup(ctx, warmupFunc)

	for {
		select {
		case <-ticker.C:
			cw.performWarmup(ctx, warmupFunc)
		case <-cw.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop stops the cache warmer
func (cw *CacheWarmer) Stop() {
	close(cw.stopCh)
}

func (cw *CacheWarmer) performWarmup(ctx context.Context, warmupFunc func(context.Context) (map[string][]interface{}, error)) {
	start := time.Now()
	
	data, err := warmupFunc(ctx)
	if err != nil {
		cw.logger.Error().Err(err).Msg("Failed to get data for cache warming")
		return
	}

	for dataType, items := range data {
		if err := cw.service.WarmCache(ctx, dataType, items); err != nil {
			cw.logger.Error().
				Err(err).
				Str("data_type", dataType).
				Msg("Failed to warm cache")
		}
	}

	cw.logger.Info().
		Dur("duration", time.Since(start)).
		Msg("Cache warming completed")
}
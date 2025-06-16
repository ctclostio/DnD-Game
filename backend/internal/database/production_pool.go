package database

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/ctclostio/DnD-Game/backend/pkg/logger"
)

// PoolMetrics tracks database connection pool statistics
type PoolMetrics struct {
	mu sync.RWMutex

	// Connection pool stats
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxIdleTimeClosed int64
	MaxLifetimeClosed int64

	// Query stats
	TotalQueries         int64
	FailedQueries        int64
	SlowQueries          int64
	AverageQueryDuration time.Duration

	// Transaction stats
	TotalTransactions      int64
	FailedTransactions     int64
	RolledBackTransactions int64

	LastUpdate time.Time
}

// PoolMonitor monitors database connection pool health
type PoolMonitor struct {
	db          *DB
	metrics     *PoolMetrics
	logger      *logger.LoggerV2
	ticker      *time.Ticker
	stopCh      chan struct{}
	slowQueryMs int64
}

// ProductionPoolConfig contains production-optimized connection pool settings
type ProductionPoolConfig struct {
	// Connection limits
	MaxOpenConns int           // Maximum number of open connections
	MaxIdleConns int           // Maximum number of idle connections
	MaxLifetime  time.Duration // Maximum lifetime of a connection

	// Performance tuning
	MaxIdleTime time.Duration // Maximum time a connection can be idle
	SlowQueryMs int64         // Threshold for slow query logging (milliseconds)

	// Monitoring
	EnableMonitoring bool          // Enable pool monitoring
	MonitorInterval  time.Duration // How often to collect metrics
}

// DefaultProductionPoolConfig returns recommended production settings
func DefaultProductionPoolConfig() *ProductionPoolConfig {
	return &ProductionPoolConfig{
		// Connection limits based on typical production needs
		MaxOpenConns: 100,              // Support concurrent requests
		MaxIdleConns: 25,               // Keep some connections ready
		MaxLifetime:  30 * time.Minute, // Rotate connections periodically

		// Performance tuning
		MaxIdleTime: 5 * time.Minute, // Close idle connections after 5 minutes
		SlowQueryMs: 100,             // Log queries taking > 100ms

		// Monitoring
		EnableMonitoring: true,
		MonitorInterval:  30 * time.Second,
	}
}

// ConfigureProductionPool applies production-optimized settings to a database connection
func ConfigureProductionPool(db *DB, cfg *ProductionPoolConfig) (*PoolMonitor, error) {
	if cfg == nil {
		cfg = DefaultProductionPoolConfig()
	}

	// Apply connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Set max idle time if supported (Go 1.15+)
	if cfg.MaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.MaxIdleTime)
	}

	// Create pool monitor if enabled
	if cfg.EnableMonitoring {
		monitor := &PoolMonitor{
			db:          db,
			metrics:     &PoolMetrics{},
			logger:      db.logger,
			slowQueryMs: cfg.SlowQueryMs,
			stopCh:      make(chan struct{}),
		}

		// Start monitoring
		monitor.Start(cfg.MonitorInterval)
		return monitor, nil
	}

	return nil, nil
}

// Start begins monitoring the connection pool
func (pm *PoolMonitor) Start(interval time.Duration) {
	pm.ticker = time.NewTicker(interval)

	go func() {
		// Collect initial metrics
		pm.collectMetrics()

		for {
			select {
			case <-pm.ticker.C:
				pm.collectMetrics()
				pm.logMetrics()
			case <-pm.stopCh:
				pm.ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the pool monitor
func (pm *PoolMonitor) Stop() {
	if pm.ticker != nil {
		close(pm.stopCh)
	}
}

// collectMetrics gathers current pool statistics
func (pm *PoolMonitor) collectMetrics() {
	stats := pm.db.Stats()

	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	pm.metrics.OpenConnections = stats.OpenConnections
	pm.metrics.InUse = stats.InUse
	pm.metrics.Idle = stats.Idle
	pm.metrics.WaitCount = stats.WaitCount
	pm.metrics.WaitDuration = stats.WaitDuration
	pm.metrics.MaxIdleClosed = stats.MaxIdleClosed
	pm.metrics.MaxIdleTimeClosed = stats.MaxIdleTimeClosed
	pm.metrics.MaxLifetimeClosed = stats.MaxLifetimeClosed
	pm.metrics.LastUpdate = time.Now()
}

// logMetrics logs current pool metrics
func (pm *PoolMonitor) logMetrics() {
	if pm.logger == nil {
		return
	}

	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()

	// Log pool health
	pm.logger.Info().
		Int("open_connections", pm.metrics.OpenConnections).
		Int("in_use", pm.metrics.InUse).
		Int("idle", pm.metrics.Idle).
		Int64("wait_count", pm.metrics.WaitCount).
		Dur("wait_duration", pm.metrics.WaitDuration).
		Int64("max_idle_closed", pm.metrics.MaxIdleClosed).
		Int64("max_lifetime_closed", pm.metrics.MaxLifetimeClosed).
		Msg("Database connection pool metrics")

	// Warn if pool is under pressure
	utilizationPercent := 0
	if pm.metrics.OpenConnections > 0 {
		utilizationPercent = (pm.metrics.InUse * 100) / pm.metrics.OpenConnections
	}

	if utilizationPercent > 80 {
		pm.logger.Warn().
			Int("utilization_percent", utilizationPercent).
			Int("in_use", pm.metrics.InUse).
			Int("open", pm.metrics.OpenConnections).
			Msg("High database connection pool utilization")
	}

	if pm.metrics.WaitCount > 0 {
		pm.logger.Warn().
			Int64("wait_count", pm.metrics.WaitCount).
			Dur("total_wait_duration", pm.metrics.WaitDuration).
			Msg("Database connections are waiting for available slots")
	}
}

// GetMetrics returns a copy of current metrics
func (pm *PoolMonitor) GetMetrics() PoolMetrics {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()
	return *pm.metrics
}

// RecordQuery updates query statistics
func (pm *PoolMonitor) RecordQuery(duration time.Duration, err error) {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	pm.metrics.TotalQueries++

	if err != nil && err != sql.ErrNoRows {
		pm.metrics.FailedQueries++
	}

	if duration.Milliseconds() > pm.slowQueryMs {
		pm.metrics.SlowQueries++
	}

	// Update average query duration (simple moving average)
	if pm.metrics.TotalQueries == 1 {
		pm.metrics.AverageQueryDuration = duration
	} else {
		// Weighted average to prevent overflow
		weight := 0.1 // Give 10% weight to new sample
		pm.metrics.AverageQueryDuration = time.Duration(
			float64(pm.metrics.AverageQueryDuration)*(1-weight) + float64(duration)*weight,
		)
	}
}

// HealthCheck performs a database health check
func HealthCheck(ctx context.Context, db *DB) error {
	// Set a timeout for the health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Perform a simple query
	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return err
	}

	// Check connection pool health
	stats := db.Stats()

	// Warn if too many connections are waiting
	if stats.WaitCount > 10 {
		if db.logger != nil {
			db.logger.Warn().
				Int64("wait_count", stats.WaitCount).
				Msg("Database health check: high connection wait count")
		}
	}

	return nil
}

// OptimizeForReadHeavyWorkload adjusts pool settings for read-heavy applications
func OptimizeForReadHeavyWorkload(cfg *ProductionPoolConfig) {
	cfg.MaxOpenConns = 150      // More connections for concurrent reads
	cfg.MaxIdleConns = 50       // Keep more idle connections
	cfg.MaxLifetime = time.Hour // Longer lifetime is OK for reads
	cfg.MaxIdleTime = 10 * time.Minute
}

// OptimizeForWriteHeavyWorkload adjusts pool settings for write-heavy applications
func OptimizeForWriteHeavyWorkload(cfg *ProductionPoolConfig) {
	cfg.MaxOpenConns = 50              // Fewer connections to reduce contention
	cfg.MaxIdleConns = 10              // Less idle connections
	cfg.MaxLifetime = 15 * time.Minute // Shorter lifetime to handle locks
	cfg.MaxIdleTime = 2 * time.Minute
}

// OptimizeForMixedWorkload adjusts pool settings for balanced read/write
func OptimizeForMixedWorkload(cfg *ProductionPoolConfig) {
	cfg.MaxOpenConns = 100
	cfg.MaxIdleConns = 25
	cfg.MaxLifetime = 30 * time.Minute
	cfg.MaxIdleTime = 5 * time.Minute
}

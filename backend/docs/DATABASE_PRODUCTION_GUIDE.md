# Database Production Configuration Guide

## Overview

This guide provides recommendations for configuring the D&D Game database for production environments, including connection pooling, monitoring, and performance optimization.

## Connection Pool Configuration

### Environment Variables

Configure these environment variables for production:

```bash
# Connection Pool Settings
DB_MAX_OPEN_CONNS=100      # Maximum open connections (default: 25)
DB_MAX_IDLE_CONNS=25       # Maximum idle connections (default: 25)
DB_MAX_LIFETIME=30m        # Maximum connection lifetime (default: 5m)

# Database Connection
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=dndgame_prod
DB_PASSWORD=secure-password
DB_NAME=dndgame_production
DB_SSLMODE=require         # Use 'require' for production
```

### Production Pool Configuration

For production environments, use the `ProductionPoolConfig`:

```go
import "github.com/ctclostio/DnD-Game/backend/internal/database"

// In your main.go or initialization code
func initializeDatabase(cfg *config.Config) (*database.DB, error) {
    // Initialize database connection
    db, repos, err := database.Initialize(cfg)
    if err != nil {
        return nil, err
    }

    // Apply production pool configuration
    poolCfg := database.DefaultProductionPoolConfig()
    
    // Adjust based on your workload
    // database.OptimizeForReadHeavyWorkload(poolCfg)  // For read-heavy apps
    // database.OptimizeForWriteHeavyWorkload(poolCfg) // For write-heavy apps
    database.OptimizeForMixedWorkload(poolCfg)        // For balanced workload

    // Enable monitoring
    monitor, err := database.ConfigureProductionPool(db, poolCfg)
    if err != nil {
        return nil, err
    }

    // Store monitor for graceful shutdown
    app.poolMonitor = monitor

    return db, nil
}
```

## Workload-Specific Optimizations

### Read-Heavy Applications

For applications with mostly SELECT queries:

```bash
DB_MAX_OPEN_CONNS=150
DB_MAX_IDLE_CONNS=50
DB_MAX_LIFETIME=1h
```

Characteristics:
- Higher connection count for concurrent reads
- More idle connections to reduce connection overhead
- Longer connection lifetime acceptable

### Write-Heavy Applications

For applications with many INSERT/UPDATE/DELETE operations:

```bash
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_MAX_LIFETIME=15m
```

Characteristics:
- Lower connection count to reduce lock contention
- Fewer idle connections
- Shorter lifetime to handle transaction locks

### Mixed Workload (Recommended for D&D Game)

For balanced read/write operations:

```bash
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=25
DB_MAX_LIFETIME=30m
```

## Monitoring and Health Checks

### Health Check Endpoint

Implement a health check endpoint:

```go
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Check database connectivity
    if err := database.HealthCheck(ctx, h.db); err != nil {
        resp.Error(w, r, errors.Wrap(err, "database health check failed"))
        return
    }

    // Get pool metrics if monitor is available
    var poolStats map[string]interface{}
    if h.poolMonitor != nil {
        metrics := h.poolMonitor.GetMetrics()
        poolStats = map[string]interface{}{
            "open_connections": metrics.OpenConnections,
            "in_use":          metrics.InUse,
            "idle":            metrics.Idle,
            "wait_count":      metrics.WaitCount,
        }
    }

    resp.Success(w, r, map[string]interface{}{
        "status":      "healthy",
        "database":    "connected",
        "pool_stats":  poolStats,
        "timestamp":   time.Now().UTC(),
    })
}
```

### Monitoring Alerts

Set up alerts for these conditions:

1. **High Connection Usage** (> 80%)
   ```
   Alert: (in_use / max_open_conns) > 0.8
   Action: Consider increasing DB_MAX_OPEN_CONNS
   ```

2. **Connection Wait Time**
   ```
   Alert: wait_count > 10 OR wait_duration > 1s
   Action: Increase connection pool size or optimize queries
   ```

3. **Excessive Connection Churn**
   ```
   Alert: max_lifetime_closed > 100/minute
   Action: Increase DB_MAX_LIFETIME
   ```

4. **Slow Queries**
   ```
   Alert: slow_queries > 10% of total_queries
   Action: Add indexes or optimize query patterns
   ```

## Performance Best Practices

### 1. Use Prepared Statements

The Rebind methods automatically handle prepared statements:

```go
query := `SELECT * FROM characters WHERE user_id = ? AND level >= ?`
rows, err := db.QueryContextRebind(ctx, query, userID, minLevel)
```

### 2. Connection Lifecycle

- **Don't hold connections unnecessarily**: Use contexts with timeouts
- **Close rows immediately**: Use `defer rows.Close()`
- **Use transactions wisely**: Keep them short

### 3. Query Optimization

- Use the production indexes (migration 022)
- Monitor slow queries in logs
- Use `EXPLAIN ANALYZE` for query planning

### 4. Graceful Shutdown

Implement graceful shutdown to close connections properly:

```go
func gracefulShutdown(db *database.DB, monitor *database.PoolMonitor) {
    // Stop monitoring
    if monitor != nil {
        monitor.Stop()
    }

    // Close database connections
    if err := db.Close(); err != nil {
        log.Error().Err(err).Msg("Error closing database")
    }
}
```

## Troubleshooting

### Common Issues

1. **"too many connections" errors**
   - Reduce `DB_MAX_OPEN_CONNS`
   - Check for connection leaks (unclosed rows/transactions)

2. **High latency**
   - Check `wait_count` in metrics
   - Increase connection pool size
   - Add database read replicas

3. **Connection timeouts**
   - Increase `DB_MAX_LIFETIME`
   - Check network connectivity
   - Verify database server resources

### Debug Logging

Enable debug logging for connection pool issues:

```go
// In development/staging only
db.SetLogger(logger.NewWithLevel(logger.DebugLevel))
```

## PostgreSQL Specific Settings

For PostgreSQL, also configure these server-side settings:

```sql
-- Maximum connections (should be > your app's max_open_conns)
ALTER SYSTEM SET max_connections = 200;

-- Statement timeout (prevent long-running queries)
ALTER SYSTEM SET statement_timeout = '30s';

-- Log slow queries
ALTER SYSTEM SET log_min_duration_statement = 100; -- 100ms

-- Connection limits per database
ALTER DATABASE dndgame_production CONNECTION LIMIT 150;
```

## Monitoring Tools Integration

### Prometheus Metrics

Export pool metrics for Prometheus:

```go
// Expose metrics at /metrics endpoint
connectionPoolGauge.WithLabelValues("open").Set(float64(metrics.OpenConnections))
connectionPoolGauge.WithLabelValues("in_use").Set(float64(metrics.InUse))
connectionPoolGauge.WithLabelValues("idle").Set(float64(metrics.Idle))
```

### Datadog Integration

Send metrics to Datadog:

```go
statsd.Gauge("db.pool.connections.open", float64(metrics.OpenConnections), nil, 1)
statsd.Gauge("db.pool.connections.in_use", float64(metrics.InUse), nil, 1)
statsd.Count("db.pool.wait_count", metrics.WaitCount, nil, 1)
```

## Testing Pool Configuration

Test your pool configuration under load:

```bash
# Use a load testing tool
hey -n 10000 -c 100 https://your-api.com/api/characters

# Monitor database connections during the test
watch -n 1 'psql -c "SELECT count(*) FROM pg_stat_activity WHERE datname = '\''dndgame_production'\'';"'
```

## Summary

For the D&D Game in production:

1. Start with mixed workload configuration
2. Enable connection pool monitoring
3. Set up health checks and alerts
4. Monitor and adjust based on actual usage patterns
5. Use the production indexes from migration 022
6. Implement graceful shutdown

Remember to test these settings in a staging environment first!
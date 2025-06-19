package health

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/database"
)

// Health status constants
const (
	StatusHealthy  = "healthy"
	StatusWarning  = "warning"
	StatusCritical = "critical"
	
	// Error messages
	ErrDatabaseNotInitialized = "database not initialized"
)

// DBPoolChecker provides comprehensive database health checks including pool statistics
type DBPoolChecker struct {
	DB                 *database.DB
	WarnThresholds     PoolThresholds
	CriticalThresholds PoolThresholds
}

// PoolThresholds defines warning and critical thresholds for pool metrics
type PoolThresholds struct {
	ConnectionUsagePercent int           // Percentage of connections in use
	WaitCount              int64         // Number of connections waiting
	WaitDuration           time.Duration // Total wait duration
	QueryTimeout           time.Duration // Maximum query execution time
}

// DefaultWarnThresholds returns default warning thresholds
func DefaultWarnThresholds() PoolThresholds {
	return PoolThresholds{
		ConnectionUsagePercent: 70,
		WaitCount:              5,
		WaitDuration:           100 * time.Millisecond,
		QueryTimeout:           1 * time.Second,
	}
}

// DefaultCriticalThresholds returns default critical thresholds
func DefaultCriticalThresholds() PoolThresholds {
	return PoolThresholds{
		ConnectionUsagePercent: 90,
		WaitCount:              20,
		WaitDuration:           1 * time.Second,
		QueryTimeout:           5 * time.Second,
	}
}

// PoolHealthResult contains detailed pool health information
type PoolHealthResult struct {
	Status         string                 `json:"status"` // healthy, warning, critical
	Message        string                 `json:"message,omitempty"`
	Stats          sql.DBStats            `json:"stats"`
	Metrics        map[string]interface{} `json:"metrics"`
	Warnings       []string               `json:"warnings,omitempty"`
	CriticalIssues []string               `json:"critical_issues,omitempty"`
}

func (d *DBPoolChecker) Name() string { return "database_pool" }

// Check performs comprehensive database and pool health checks
func (d *DBPoolChecker) Check(ctx context.Context) error {
	if d.DB == nil {
		return fmt.Errorf(ErrDatabaseNotInitialized)
	}

	// Set default thresholds if not configured
	if d.WarnThresholds.ConnectionUsagePercent == 0 {
		d.WarnThresholds = DefaultWarnThresholds()
	}
	if d.CriticalThresholds.ConnectionUsagePercent == 0 {
		d.CriticalThresholds = DefaultCriticalThresholds()
	}

	// Test database connectivity with timeout
	queryCtx, cancel := context.WithTimeout(ctx, d.CriticalThresholds.QueryTimeout)
	defer cancel()

	start := time.Now()
	var result int
	err := d.DB.QueryRowContext(queryCtx, "SELECT 1").Scan(&result)
	queryDuration := time.Since(start)

	if err != nil {
		return fmt.Errorf("database connectivity check failed: %w", err)
	}

	// Check query performance
	if queryDuration > d.CriticalThresholds.QueryTimeout {
		return fmt.Errorf("database query exceeded critical timeout: %v", queryDuration)
	}

	// Get pool statistics
	stats := d.DB.Stats()

	// Calculate usage percentage
	usagePercent := 0
	if stats.MaxOpenConnections > 0 {
		usagePercent = int((float64(stats.InUse) / float64(stats.MaxOpenConnections)) * 100)
	}

	// Check critical conditions
	if usagePercent >= d.CriticalThresholds.ConnectionUsagePercent {
		return fmt.Errorf("critical: connection pool usage at %d%%", usagePercent)
	}

	if stats.WaitCount >= d.CriticalThresholds.WaitCount {
		return fmt.Errorf("critical: %d connections waiting for available slot", stats.WaitCount)
	}

	if stats.WaitDuration >= d.CriticalThresholds.WaitDuration {
		return fmt.Errorf("critical: connections waited %v total", stats.WaitDuration)
	}

	// Check warning conditions (these don't fail the health check)
	if usagePercent >= d.WarnThresholds.ConnectionUsagePercent {
		// Log warning but don't fail
		// TODO: Add logging for warning conditions
	}

	return nil
}

// GetDetailedHealth returns comprehensive pool health information
func (d *DBPoolChecker) GetDetailedHealth(ctx context.Context) (*PoolHealthResult, error) {
	result := &PoolHealthResult{
		Status:         StatusHealthy,
		Metrics:        make(map[string]interface{}),
		Warnings:       []string{},
		CriticalIssues: []string{},
	}

	if d.DB == nil {
		result.Status = StatusCritical
		result.Message = "Database not initialized"
		return result, fmt.Errorf(ErrDatabaseNotInitialized)
	}

	// Test connectivity
	queryDuration, err := d.testConnectivity(ctx)
	if err != nil {
		result.Status = StatusCritical
		result.Message = fmt.Sprintf("Database connectivity failed: %v", err)
		result.CriticalIssues = append(result.CriticalIssues, result.Message)
		return result, err
	}

	// Get pool statistics
	stats := d.DB.Stats()
	result.Stats = stats

	// Calculate and populate metrics
	usagePercent, avgWaitTime := d.calculateMetrics(stats)
	d.populateMetrics(result.Metrics, stats, usagePercent, avgWaitTime, queryDuration)

	// Check for issues
	d.checkCriticalIssues(result, stats, usagePercent, queryDuration)
	if result.Status != StatusCritical {
		d.checkWarnings(result, stats, usagePercent, queryDuration)
	}

	// Set appropriate message
	result.Message = d.getStatusMessage(result.Status, stats)

	return result, nil
}

func (d *DBPoolChecker) testConnectivity(ctx context.Context) (time.Duration, error) {
	queryCtx, cancel := context.WithTimeout(ctx, d.CriticalThresholds.QueryTimeout)
	defer cancel()

	start := time.Now()
	var testResult int
	err := d.DB.QueryRowContext(queryCtx, "SELECT 1").Scan(&testResult)
	return time.Since(start), err
}

func (d *DBPoolChecker) calculateMetrics(stats sql.DBStats) (int, time.Duration) {
	usagePercent := 0
	if stats.MaxOpenConnections > 0 {
		usagePercent = int((float64(stats.InUse) / float64(stats.MaxOpenConnections)) * 100)
	}

	avgWaitTime := time.Duration(0)
	if stats.WaitCount > 0 {
		avgWaitTime = time.Duration(int64(stats.WaitDuration) / stats.WaitCount)
	}

	return usagePercent, avgWaitTime
}

func (d *DBPoolChecker) populateMetrics(metrics map[string]interface{}, stats sql.DBStats, usagePercent int, avgWaitTime, queryDuration time.Duration) {
	metrics["connection_usage_percent"] = usagePercent
	metrics["query_duration_ms"] = queryDuration.Milliseconds()
	metrics["avg_wait_time_ms"] = avgWaitTime.Milliseconds()
	metrics["connections_recycled"] = stats.MaxIdleClosed + stats.MaxIdleTimeClosed + stats.MaxLifetimeClosed
}

func (d *DBPoolChecker) checkCriticalIssues(result *PoolHealthResult, stats sql.DBStats, usagePercent int, queryDuration time.Duration) {
	if usagePercent >= d.CriticalThresholds.ConnectionUsagePercent {
		issue := fmt.Sprintf("Connection pool usage critical: %d%% (threshold: %d%%)",
			usagePercent, d.CriticalThresholds.ConnectionUsagePercent)
		result.CriticalIssues = append(result.CriticalIssues, issue)
		result.Status = StatusCritical
	}

	if stats.WaitCount >= d.CriticalThresholds.WaitCount {
		issue := fmt.Sprintf("Too many connections waiting: %d (threshold: %d)",
			stats.WaitCount, d.CriticalThresholds.WaitCount)
		result.CriticalIssues = append(result.CriticalIssues, issue)
		result.Status = StatusCritical
	}

	if queryDuration > d.CriticalThresholds.QueryTimeout {
		issue := fmt.Sprintf("Query timeout critical: %v (threshold: %v)",
			queryDuration, d.CriticalThresholds.QueryTimeout)
		result.CriticalIssues = append(result.CriticalIssues, issue)
		result.Status = StatusCritical
	}
}

func (d *DBPoolChecker) checkWarnings(result *PoolHealthResult, stats sql.DBStats, usagePercent int, queryDuration time.Duration) {
	if usagePercent >= d.WarnThresholds.ConnectionUsagePercent {
		warning := fmt.Sprintf("Connection pool usage high: %d%% (threshold: %d%%)",
			usagePercent, d.WarnThresholds.ConnectionUsagePercent)
		result.Warnings = append(result.Warnings, warning)
		result.Status = StatusWarning
	}

	if stats.WaitCount >= d.WarnThresholds.WaitCount {
		warning := fmt.Sprintf("Connections waiting: %d (threshold: %d)",
			stats.WaitCount, d.WarnThresholds.WaitCount)
		result.Warnings = append(result.Warnings, warning)
		result.Status = StatusWarning
	}

	if queryDuration > d.WarnThresholds.QueryTimeout {
		warning := fmt.Sprintf("Query slow: %v (threshold: %v)",
			queryDuration, d.WarnThresholds.QueryTimeout)
		result.Warnings = append(result.Warnings, warning)
		result.Status = StatusWarning
	}
}

func (d *DBPoolChecker) getStatusMessage(status string, stats sql.DBStats) string {
	connInfo := fmt.Sprintf("%d/%d connections in use", stats.InUse, stats.MaxOpenConnections)
	
	switch status {
	case StatusHealthy:
		return fmt.Sprintf("Database healthy - %s", connInfo)
	case StatusWarning:
		return fmt.Sprintf("Database operational with warnings - %s", connInfo)
	default:
		return fmt.Sprintf("Database critical - %s", connInfo)
	}
}

// GetPoolMetrics returns just the pool metrics for monitoring
func (d *DBPoolChecker) GetPoolMetrics() map[string]interface{} {
	if d.DB == nil {
		return map[string]interface{}{
			"error": ErrDatabaseNotInitialized,
		}
	}

	stats := d.DB.Stats()

	usagePercent := 0
	if stats.MaxOpenConnections > 0 {
		usagePercent = int((float64(stats.InUse) / float64(stats.MaxOpenConnections)) * 100)
	}

	avgWaitTime := time.Duration(0)
	if stats.WaitCount > 0 {
		avgWaitTime = time.Duration(int64(stats.WaitDuration) / stats.WaitCount)
	}

	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"max_open_connections": stats.MaxOpenConnections,
		"wait_count":           stats.WaitCount,
		"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
		"avg_wait_time_ms":     avgWaitTime.Milliseconds(),
		"usage_percent":        usagePercent,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

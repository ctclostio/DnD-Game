package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Service   string                 `json:"service"`
	Checks    map[string]HealthCheckResult `json:"checks,omitempty"`
}

// HealthCheckResult represents an individual component health check
type HealthCheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// DetailedHealthResponse includes system metrics
type DetailedHealthResponse struct {
	HealthResponse
	System SystemInfo `json:"system"`
}

// SystemInfo contains system metrics
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	Memory       Memory `json:"memory"`
}

// Memory contains memory usage information
type Memory struct {
	Alloc      uint64 `json:"alloc_mb"`
	TotalAlloc uint64 `json:"total_alloc_mb"`
	Sys        uint64 `json:"sys_mb"`
	NumGC      uint32 `json:"num_gc"`
}

var startTime = time.Now()

// Health is a simple health check endpoint
// @Summary Basic health check
// @Description Check if the service is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Router /health [get]
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // You might want to inject this from build
		Uptime:    time.Since(startTime).String(),
		Service:   "dnd-game-api",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LivenessProbe is used by orchestrators to check if the service should be restarted
// @Summary Liveness probe
// @Description Check if the service is alive and should not be restarted
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Service is alive"
// @Failure 503 {object} map[string]string "Service is not alive"
// @Router /health/live [get]
func (h *Handlers) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	// Basic liveness check - if we can respond, we're alive
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessProbe checks if the service is ready to accept traffic
// @Summary Readiness probe
// @Description Check if the service is ready to accept traffic
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is ready"
// @Failure 503 {object} HealthResponse "Service is not ready"
// @Router /health/ready [get]
func (h *Handlers) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]HealthCheckResult)
	allHealthy := true

	// TODO: Check database connection if available
	// This would require adding a GetDB() method to repositories or passing DB separately
	/*
	if db, ok := h.repositories.(*sql.DB); ok && db != nil {
		if err := db.Ping(); err != nil {
			checks["database"] = HealthCheckResult{
				Status:  "unhealthy",
				Message: "Database connection failed",
			}
			allHealthy = false
		} else {
			checks["database"] = HealthCheckResult{
				Status: "healthy",
			}
		}
	}
	*/

	// Check if we have required services
	if h.services.Users == nil || h.services.Characters == nil {
		checks["services"] = HealthCheckResult{
			Status:  "unhealthy",
			Message: "Required services not initialized",
		}
		allHealthy = false
	} else {
		checks["services"] = HealthCheckResult{
			Status: "healthy",
		}
	}

	// Check JWT manager
	if h.jwtManager == nil {
		checks["auth"] = HealthCheckResult{
			Status:  "unhealthy",
			Message: "JWT manager not initialized",
		}
		allHealthy = false
	} else {
		checks["auth"] = HealthCheckResult{
			Status: "healthy",
		}
	}

	response := HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		Service:   "dnd-game-api",
		Checks:    checks,
	}

	if !allHealthy {
		response.Status = "not_ready"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DetailedHealth provides detailed health information including system metrics
// @Summary Detailed health check
// @Description Get detailed health information including system metrics
// @Tags health
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} DetailedHealthResponse "Detailed health information"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /health/detailed [get]
func (h *Handlers) DetailedHealth(w http.ResponseWriter, r *http.Request) {
	// Get basic health info
	checks := make(map[string]HealthCheckResult)

	// Check database if available
	if db, ok := h.repositories.(*sql.DB); ok && db != nil {
		if err := db.Ping(); err != nil {
			checks["database"] = HealthCheckResult{
				Status:  "unhealthy",
				Message: err.Error(),
			}
		} else {
			checks["database"] = HealthCheckResult{
				Status: "healthy",
			}
		}
	} else if h.repositories.Users != nil {
		checks["database"] = HealthCheckResult{
			Status: "healthy",
		}
	}

	// Check services
	if h.services.Users != nil && h.services.Characters != nil {
		checks["services"] = HealthCheckResult{
			Status: "healthy",
		}
	}

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := DetailedHealthResponse{
		HealthResponse: HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   "1.0.0",
			Uptime:    time.Since(startTime).String(),
			Service:   "dnd-game-api",
			Checks:    checks,
		},
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			Memory: Memory{
				Alloc:      m.Alloc / 1024 / 1024,      // MB
				TotalAlloc: m.TotalAlloc / 1024 / 1024, // MB
				Sys:        m.Sys / 1024 / 1024,        // MB
				NumGC:      m.NumGC,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Standalone HealthCheckV2 function for backward compatibility
func HealthCheckV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "dnd-game-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
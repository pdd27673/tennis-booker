package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// SystemHandlers handles system-level endpoints
type SystemHandlers struct {
	db        *mongo.Database
	startTime time.Time
}

// NewSystemHandlers creates a new system handlers instance
func NewSystemHandlers(db *mongo.Database) *SystemHandlers {
	return &SystemHandlers{
		db:        db,
		startTime: time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Database  string    `json:"database"`
	Uptime    string    `json:"uptime"`
}

// MetricsResponse represents the metrics response
type MetricsResponse struct {
	Status      string            `json:"status"`
	Service     string            `json:"service"`
	Timestamp   time.Time         `json:"timestamp"`
	Uptime      string            `json:"uptime"`
	Memory      MemoryStats       `json:"memory"`
	Runtime     RuntimeStats      `json:"runtime"`
	Database    DatabaseStats     `json:"database"`
	System      SystemStats       `json:"system"`
}

// MemoryStats represents memory statistics
type MemoryStats struct {
	Allocated      uint64  `json:"allocated_mb"`
	TotalAllocated uint64  `json:"total_allocated_mb"`
	SystemMemory   uint64  `json:"system_mb"`
	GCRuns         uint32  `json:"gc_runs"`
	MemoryUsage    float64 `json:"memory_usage_percent"`
}

// RuntimeStats represents runtime statistics
type RuntimeStats struct {
	GoVersion    string `json:"go_version"`
	Goroutines   int    `json:"goroutines"`
	CPUs         int    `json:"cpus"`
	GoMaxProcs   int    `json:"gomaxprocs"`
}

// DatabaseStats represents database connection statistics
type DatabaseStats struct {
	Status     string `json:"status"`
	Name       string `json:"name"`
	Connected  bool   `json:"connected"`
}

// SystemStats represents system-level statistics
type SystemStats struct {
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
}

// HealthCheck handles GET /api/health
func (h *SystemHandlers) HealthCheck(c *gin.Context) {
	// Test database connection
	dbStatus := "connected"
	ctx := c.Request.Context()
	
	if err := h.db.Client().Ping(ctx, nil); err != nil {
		dbStatus = "disconnected"
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:    "unhealthy",
			Service:   "tennis-booking-api",
			Timestamp: time.Now().UTC(),
			Version:   "1.0.0",
			Database:  dbStatus,
			Uptime:    time.Since(h.startTime).String(),
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Service:   "tennis-booking-api",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
		Database:  dbStatus,
		Uptime:    time.Since(h.startTime).String(),
	})
}

// Metrics handles GET /api/metrics
func (h *SystemHandlers) Metrics(c *gin.Context) {
	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Test database connection
	dbConnected := true
	ctx := c.Request.Context()
	if err := h.db.Client().Ping(ctx, nil); err != nil {
		dbConnected = false
	}

	// Calculate uptime
	uptime := time.Since(h.startTime)

	response := MetricsResponse{
		Status:    "ok",
		Service:   "tennis-booking-api",
		Timestamp: time.Now().UTC(),
		Uptime:    uptime.String(),
		Memory: MemoryStats{
			Allocated:      bToMb(memStats.Alloc),
			TotalAllocated: bToMb(memStats.TotalAlloc),
			SystemMemory:   bToMb(memStats.Sys),
			GCRuns:         memStats.NumGC,
			MemoryUsage:    float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		},
		Runtime: RuntimeStats{
			GoVersion:    runtime.Version(),
			Goroutines:   runtime.NumGoroutine(),
			CPUs:         runtime.NumCPU(),
			GoMaxProcs:   runtime.GOMAXPROCS(0),
		},
		Database: DatabaseStats{
			Status:    getDBStatus(dbConnected),
			Name:      h.db.Name(),
			Connected: dbConnected,
		},
		System: SystemStats{
			Platform:     runtime.GOOS,
			Architecture: runtime.GOARCH,
		},
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to convert bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Helper function to get database status string
func getDBStatus(connected bool) string {
	if connected {
		return "connected"
	}
	return "disconnected"
} 
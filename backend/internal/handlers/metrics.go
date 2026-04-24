package handlers

import (
	"net/http"
	"runtime"
	"time"

	"radimaster/internal/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var startTime = time.Now()

// MetricsHandler exposes system health and performance metrics.
type MetricsHandler struct {
	db  *gorm.DB
	hub *ws.Hub
}

// NewMetricsHandler creates a new MetricsHandler.
func NewMetricsHandler(db *gorm.DB, hub *ws.Hub) *MetricsHandler {
	return &MetricsHandler{db: db, hub: hub}
}

// GetMetrics returns comprehensive system metrics for the admin dashboard.
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Database pool stats
	sqlDB, _ := h.db.DB()
	dbStats := sqlDB.Stats()

	// WebSocket stats from hub
	wsStats := h.hub.Stats()

	// Database health
	dbStatus := "connected"
	if err := sqlDB.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	// Count records
	var userCount, caseCount, classroomCount int64
	h.db.Table("users").Count(&userCount)
	h.db.Table("medical_cases").Count(&caseCount)
	h.db.Table("classrooms").Count(&classroomCount)

	c.JSON(http.StatusOK, gin.H{
		"system": gin.H{
			"uptime_seconds": int(time.Since(startTime).Seconds()),
			"uptime_human":   time.Since(startTime).Round(time.Second).String(),
			"go_version":     runtime.Version(),
			"goroutines":     runtime.NumGoroutine(),
			"cpu_cores":      runtime.NumCPU(),
		},
		"memory": gin.H{
			"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
			"sys_mb":         float64(m.Sys) / 1024 / 1024,
			"gc_cycles":      m.NumGC,
			"gc_pause_total": time.Duration(m.PauseTotalNs).String(),
		},
		"database": gin.H{
			"status":           dbStatus,
			"open_connections": dbStats.OpenConnections,
			"in_use":           dbStats.InUse,
			"idle":             dbStats.Idle,
			"max_open":         dbStats.MaxOpenConnections,
			"wait_count":       dbStats.WaitCount,
			"wait_duration":    dbStats.WaitDuration.String(),
		},
		"websocket": wsStats,
		"data": gin.H{
			"users":      userCount,
			"cases":      caseCount,
			"classrooms": classroomCount,
		},
	})
}

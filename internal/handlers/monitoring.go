package handlers

import (
	"net/http"
	"time"

	"ccany/internal/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// MonitoringHandler monitoring handler
type MonitoringHandler struct {
	systemMonitor *monitoring.SystemMonitor
	logger        *logrus.Logger
}

// NewMonitoringHandler creates new monitoring handler
func NewMonitoringHandler(systemMonitor *monitoring.SystemMonitor, logger *logrus.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		systemMonitor: systemMonitor,
		logger:        logger,
	}
}

// GetHealthStatus gets system health status
func (h *MonitoringHandler) GetHealthStatus(c *gin.Context) {
	healthStatus := h.systemMonitor.GetHealthStatus()

	// set HTTP status code based on health status
	var statusCode int
	switch healthStatus.Status {
	case "healthy":
		statusCode = http.StatusOK
	case "degraded":
		statusCode = http.StatusOK // degraded status still returns 200, but marked in response
	case "unhealthy":
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, healthStatus)
}

// GetSystemMetrics gets system metrics
func (h *MonitoringHandler) GetSystemMetrics(c *gin.Context) {
	metrics := h.systemMonitor.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"metrics":   metrics,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetDetailedHealth gets detailed health check information
func (h *MonitoringHandler) GetDetailedHealth(c *gin.Context) {
	healthStatus := h.systemMonitor.GetHealthStatus()

	// add additional detailed information
	response := gin.H{
		"status":    healthStatus.Status,
		"timestamp": healthStatus.Timestamp.Format(time.RFC3339),
		"checks":    healthStatus.Checks,
		"metrics":   healthStatus.Metrics,
		"summary": gin.H{
			"total_checks": len(healthStatus.Checks),
			"healthy_checks": func() int {
				count := 0
				for _, check := range healthStatus.Checks {
					if check.Status == "healthy" {
						count++
					}
				}
				return count
			}(),
			"degraded_checks": func() int {
				count := 0
				for _, check := range healthStatus.Checks {
					if check.Status == "degraded" {
						count++
					}
				}
				return count
			}(),
			"unhealthy_checks": func() int {
				count := 0
				for _, check := range healthStatus.Checks {
					if check.Status == "unhealthy" {
						count++
					}
				}
				return count
			}(),
		},
	}

	// set HTTP status code based on health status
	var statusCode int
	switch healthStatus.Status {
	case "healthy":
		statusCode = http.StatusOK
	case "degraded":
		statusCode = http.StatusOK
	case "unhealthy":
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, response)
}

// GetLivenessProbe liveness probe (Kubernetes style)
func (h *MonitoringHandler) GetLivenessProbe(c *gin.Context) {
	// simple liveness check - considered alive as long as program is running
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetReadinessProbe readiness probe (Kubernetes style)
func (h *MonitoringHandler) GetReadinessProbe(c *gin.Context) {
	healthStatus := h.systemMonitor.GetHealthStatus()

	// readiness check - check if critical services are available
	isReady := true
	readinessChecks := make(map[string]bool)

	// check database connection
	if dbCheck, exists := healthStatus.Checks["database"]; exists {
		dbReady := dbCheck.Status == "healthy"
		readinessChecks["database"] = dbReady
		if !dbReady {
			isReady = false
		}
	}

	// check memory usage
	if memoryCheck, exists := healthStatus.Checks["memory"]; exists {
		memoryReady := memoryCheck.Status != "unhealthy"
		readinessChecks["memory"] = memoryReady
		if !memoryReady {
			isReady = false
		}
	}

	response := gin.H{
		"status": func() string {
			if isReady {
				return "ready"
			}
			return "not_ready"
		}(),
		"timestamp": time.Now().Format(time.RFC3339),
		"checks":    readinessChecks,
	}

	if isReady {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// GetStartupProbe startup probe (Kubernetes style)
func (h *MonitoringHandler) GetStartupProbe(c *gin.Context) {
	metrics := h.systemMonitor.GetMetrics()

	// startup probe - check if application has finished starting
	// simply check if uptime exceeds 30 seconds
	isStarted := metrics.Uptime > 30*time.Second

	response := gin.H{
		"status": func() string {
			if isStarted {
				return "started"
			}
			return "starting"
		}(),
		"timestamp":      time.Now().Format(time.RFC3339),
		"uptime_seconds": metrics.Uptime.Seconds(),
	}

	if isStarted {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// GetSystemInfo gets system information
func (h *MonitoringHandler) GetSystemInfo(c *gin.Context) {
	metrics := h.systemMonitor.GetMetrics()

	response := gin.H{
		"application": gin.H{
			"name":    "Claude Code Proxy",
			"version": "1.0.0",
			"uptime":  metrics.Uptime.String(),
		},
		"system": gin.H{
			"goroutines":     metrics.GoroutineCount,
			"memory_usage":   metrics.MemoryUsage,
			"memory_total":   metrics.MemoryTotal,
			"memory_percent": metrics.MemoryPercent,
		},
		"database": gin.H{
			"healthy": metrics.DBHealthy,
		},
		"performance": gin.H{
			"request_count":     metrics.RequestCount,
			"error_count":       metrics.ErrorCount,
			"avg_response_time": metrics.AvgResponseTime,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

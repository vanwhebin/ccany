package monitoring

import (
	"context"
	"runtime"
	"sync"
	"time"

	"ccany/ent"
	"ccany/internal/logging"

	"github.com/sirupsen/logrus"
)

// SystemMonitor system monitoring service
type SystemMonitor struct {
	db            *ent.Client
	requestLogger *logging.RequestLogger
	logger        *logrus.Logger

	// monitoring metrics
	metrics      *SystemMetrics
	metricsMutex sync.RWMutex

	// monitoring state
	isRunning   bool
	stopChannel chan struct{}

	// configuration
	config *MonitorConfig
}

// SystemMetrics system metrics
type SystemMetrics struct {
	// system resources
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    uint64  `json:"memory_usage"`
	MemoryTotal    uint64  `json:"memory_total"`
	MemoryPercent  float64 `json:"memory_percent"`
	GoroutineCount int     `json:"goroutine_count"`

	// application metrics
	RequestCount    int64   `json:"request_count"`
	ErrorCount      int64   `json:"error_count"`
	AvgResponseTime float64 `json:"avg_response_time"`

	// database metrics
	DBConnections int  `json:"db_connections"`
	DBHealthy     bool `json:"db_healthy"`

	// timestamp
	Timestamp time.Time     `json:"timestamp"`
	Uptime    time.Duration `json:"uptime"`
}

// MonitorConfig monitoring configuration
type MonitorConfig struct {
	// monitoring interval
	CollectionInterval time.Duration `json:"collection_interval"`

	// threshold settings
	MemoryThreshold       float64 `json:"memory_threshold"`
	ResponseTimeThreshold float64 `json:"response_time_threshold"`
	ErrorRateThreshold    float64 `json:"error_rate_threshold"`

	// alert settings
	AlertEnabled  bool          `json:"alert_enabled"`
	AlertCooldown time.Duration `json:"alert_cooldown"`

	// data retention
	MetricsRetentionDays int `json:"metrics_retention_days"`
}

// NewSystemMonitor creates a new system monitoring service
func NewSystemMonitor(db *ent.Client, requestLogger *logging.RequestLogger, logger *logrus.Logger, config *MonitorConfig) *SystemMonitor {
	if config == nil {
		config = DefaultMonitorConfig()
	}

	return &SystemMonitor{
		db:            db,
		requestLogger: requestLogger,
		logger:        logger,
		metrics:       &SystemMetrics{},
		config:        config,
		stopChannel:   make(chan struct{}),
	}
}

// DefaultMonitorConfig default monitoring configuration
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		CollectionInterval:    30 * time.Second,
		MemoryThreshold:       80.0,   // 80%
		ResponseTimeThreshold: 5000.0, // 5 seconds
		ErrorRateThreshold:    5.0,    // 5%
		AlertEnabled:          true,
		AlertCooldown:         5 * time.Minute,
		MetricsRetentionDays:  30,
	}
}

// Start starts the monitoring service
func (sm *SystemMonitor) Start(ctx context.Context) error {
	if sm.isRunning {
		return nil
	}

	sm.isRunning = true
	startTime := time.Now()

	go func() {
		ticker := time.NewTicker(sm.config.CollectionInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				sm.logger.Info("System monitor stopped due to context cancellation")
				return
			case <-sm.stopChannel:
				sm.logger.Info("System monitor stopped")
				return
			case <-ticker.C:
				sm.collectMetrics(ctx, startTime)
			}
		}
	}()

	sm.logger.Info("System monitor started")
	return nil
}

// Stop stops the monitoring service
func (sm *SystemMonitor) Stop() {
	if !sm.isRunning {
		return
	}

	sm.isRunning = false
	close(sm.stopChannel)
	sm.logger.Info("System monitor stop requested")
}

// collectMetrics collects system metrics
func (sm *SystemMonitor) collectMetrics(ctx context.Context, startTime time.Time) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// collect basic system metrics
	metrics := &SystemMetrics{
		MemoryUsage:    m.Alloc,
		MemoryTotal:    m.Sys,
		MemoryPercent:  float64(m.Alloc) / float64(m.Sys) * 100,
		GoroutineCount: runtime.NumGoroutine(),
		Timestamp:      time.Now(),
		Uptime:         time.Since(startTime),
	}

	// collect database metrics
	metrics.DBHealthy = sm.checkDatabaseHealth(ctx)

	// collect request statistics
	if sm.requestLogger != nil {
		sm.collectRequestMetrics(ctx, metrics)
	}

	// update metrics
	sm.metricsMutex.Lock()
	sm.metrics = metrics
	sm.metricsMutex.Unlock()

	// check alert conditions
	if sm.config.AlertEnabled {
		sm.checkAlerts(metrics)
	}

	sm.logger.WithFields(logrus.Fields{
		"memory_usage":      metrics.MemoryUsage,
		"memory_percent":    metrics.MemoryPercent,
		"goroutines":        metrics.GoroutineCount,
		"db_healthy":        metrics.DBHealthy,
		"request_count":     metrics.RequestCount,
		"error_count":       metrics.ErrorCount,
		"avg_response_time": metrics.AvgResponseTime,
	}).Debug("System metrics collected")
}

// checkDatabaseHealth checks database health status
func (sm *SystemMonitor) checkDatabaseHealth(ctx context.Context) bool {
	// use timeout context
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// execute simple database query
	_, err := sm.db.AppConfig.Query().Count(healthCtx)
	return err == nil
}

// collectRequestMetrics collects request metrics
func (sm *SystemMonitor) collectRequestMetrics(ctx context.Context, metrics *SystemMetrics) {
	// get request statistics for the recent collection period
	endTime := time.Now()
	startTime := endTime.Add(-sm.config.CollectionInterval)

	opts := &logging.RequestLogQueryOptions{
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     1000, // limit query count
	}

	stats, err := sm.requestLogger.GetRequestLogStats(ctx, opts)
	if err != nil {
		sm.logger.WithError(err).Error("Failed to get request statistics")
		return
	}

	metrics.RequestCount = int64(stats.TotalRequests)
	metrics.ErrorCount = int64(stats.ErrorRequests)
	metrics.AvgResponseTime = stats.AvgDurationMs
}

// checkAlerts checks alert conditions
func (sm *SystemMonitor) checkAlerts(metrics *SystemMetrics) {
	alerts := []string{}

	// memory usage alert
	if metrics.MemoryPercent > sm.config.MemoryThreshold {
		alerts = append(alerts, "High memory usage")
	}

	// response time alert
	if metrics.AvgResponseTime > sm.config.ResponseTimeThreshold {
		alerts = append(alerts, "High response time")
	}

	// error rate alert
	if metrics.RequestCount > 0 {
		errorRate := float64(metrics.ErrorCount) / float64(metrics.RequestCount) * 100
		if errorRate > sm.config.ErrorRateThreshold {
			alerts = append(alerts, "High error rate")
		}
	}

	// database health alert
	if !metrics.DBHealthy {
		alerts = append(alerts, "Database unhealthy")
	}

	// send alerts
	if len(alerts) > 0 {
		sm.logger.WithFields(logrus.Fields{
			"alerts":  alerts,
			"metrics": metrics,
		}).Warn("System health alerts triggered")
	}
}

// GetMetrics retrieves current system metrics
func (sm *SystemMonitor) GetMetrics() *SystemMetrics {
	sm.metricsMutex.RLock()
	defer sm.metricsMutex.RUnlock()

	// return a copy of the metrics
	metrics := *sm.metrics
	return &metrics
}

// GetHealthStatus retrieves system health status
func (sm *SystemMonitor) GetHealthStatus() *HealthStatus {
	metrics := sm.GetMetrics()

	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Checks: map[string]CheckResult{
			"memory": {
				Status:  "healthy",
				Message: "Memory usage within normal range",
				Details: map[string]interface{}{
					"usage_percent": metrics.MemoryPercent,
					"threshold":     sm.config.MemoryThreshold,
				},
			},
			"database": {
				Status:  "healthy",
				Message: "Database connection healthy",
				Details: map[string]interface{}{
					"healthy": metrics.DBHealthy,
				},
			},
			"response_time": {
				Status:  "healthy",
				Message: "Response time within acceptable range",
				Details: map[string]interface{}{
					"avg_ms":    metrics.AvgResponseTime,
					"threshold": sm.config.ResponseTimeThreshold,
				},
			},
		},
		Metrics: metrics,
	}

	// check each health status
	if metrics.MemoryPercent > sm.config.MemoryThreshold {
		status.Status = "unhealthy"
		memoryCheck := status.Checks["memory"]
		memoryCheck.Status = "unhealthy"
		memoryCheck.Message = "High memory usage detected"
		status.Checks["memory"] = memoryCheck
	}

	if !metrics.DBHealthy {
		status.Status = "unhealthy"
		dbCheck := status.Checks["database"]
		dbCheck.Status = "unhealthy"
		dbCheck.Message = "Database connection failed"
		status.Checks["database"] = dbCheck
	}

	if metrics.AvgResponseTime > sm.config.ResponseTimeThreshold {
		status.Status = "degraded"
		responseCheck := status.Checks["response_time"]
		responseCheck.Status = "degraded"
		responseCheck.Message = "High response time detected"
		status.Checks["response_time"] = responseCheck
	}

	return status
}

// HealthStatus health status
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
	Metrics   *SystemMetrics         `json:"metrics"`
}

// CheckResult check result
type CheckResult struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

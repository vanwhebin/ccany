package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ccany/ent"
	"ccany/internal/logging"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogsHandler handles request logs related APIs
type RequestLogsHandler struct {
	requestLogger *logging.RequestLogger
	logger        *logrus.Logger
}

// NewRequestLogsHandler creates new request logs handler
func NewRequestLogsHandler(requestLogger *logging.RequestLogger, logger *logrus.Logger) *RequestLogsHandler {
	return &RequestLogsHandler{
		requestLogger: requestLogger,
		logger:        logger,
	}
}

// GetRequestLogs gets request logs list
func (h *RequestLogsHandler) GetRequestLogs(c *gin.Context) {
	opts := &logging.RequestLogQueryOptions{}

	// parse query parameters
	if model := c.Query("model"); model != "" {
		opts.ClaudeModel = model
	}

	if statusCode := c.Query("status_code"); statusCode != "" {
		if code, err := strconv.Atoi(statusCode); err == nil {
			opts.StatusCode = code
		}
	}

	if streaming := c.Query("streaming"); streaming != "" {
		if isStreaming, err := strconv.ParseBool(streaming); err == nil {
			opts.IsStreaming = &isStreaming
		}
	}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			opts.StartTime = t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			opts.EndTime = t
		}
	}

	// pagination parameters
	opts.Limit = 50 // default limit
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 1000 {
			opts.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			opts.Offset = o
		}
	}

	// get log records
	logs, err := h.requestLogger.GetRequestLogs(c.Request.Context(), opts)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get request logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get request logs"})
		return
	}

	// convert to response format, hiding sensitive information
	var responseLogs []gin.H
	for _, log := range logs {
		responseLog := gin.H{
			"id":            log.ID,
			"claude_model":  log.ClaudeModel,
			"openai_model":  log.OpenaiModel,
			"status_code":   log.StatusCode,
			"is_streaming":  log.IsStreaming,
			"input_tokens":  log.InputTokens,
			"output_tokens": log.OutputTokens,
			"duration_ms":   log.DurationMs,
			"created_at":    log.CreatedAt,
		}

		if log.ErrorMessage != "" {
			responseLog["error_message"] = log.ErrorMessage
		}

		// only include request and response body when needed (to avoid data being too large)
		if c.Query("include_body") == "true" {
			responseLog["request_body"] = log.RequestBody
			responseLog["response_body"] = log.ResponseBody
		}

		responseLogs = append(responseLogs, responseLog)
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  responseLogs,
		"count": len(responseLogs),
		"query": opts,
	})
}

// GetRequestLogStats gets request log statistics
func (h *RequestLogsHandler) GetRequestLogStats(c *gin.Context) {
	opts := &logging.RequestLogQueryOptions{}

	// parse query parameters
	if model := c.Query("model"); model != "" {
		opts.ClaudeModel = model
	}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			opts.StartTime = t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			opts.EndTime = t
		}
	}

	// if no time range specified, default to last 24 hours
	if opts.StartTime.IsZero() && opts.EndTime.IsZero() {
		opts.EndTime = time.Now()
		opts.StartTime = opts.EndTime.Add(-24 * time.Hour)
	}

	// get statistics
	stats, err := h.requestLogger.GetRequestLogStats(c.Request.Context(), opts)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get request log stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get request log stats"})
		return
	}

	// add additional statistics
	successRate := 0.0
	if stats.TotalRequests > 0 {
		successRate = float64(stats.SuccessRequests) / float64(stats.TotalRequests) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
		"metrics": gin.H{
			"success_rate": successRate,
			"error_rate":   100 - successRate,
			"avg_input_tokens": func() float64 {
				if stats.TotalRequests > 0 {
					return float64(stats.TotalInputTokens) / float64(stats.TotalRequests)
				}
				return 0
			}(),
			"avg_output_tokens": func() float64 {
				if stats.TotalRequests > 0 {
					return float64(stats.TotalOutputTokens) / float64(stats.TotalRequests)
				}
				return 0
			}(),
		},
		"query": opts,
	})
}

// GetRequestLogDetails gets single request log details
func (h *RequestLogsHandler) GetRequestLogDetails(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request ID is required"})
		return
	}

	// query single log by ID
	log, err := h.requestLogger.GetRequestLogByID(c.Request.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request log not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get request log details")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get request log details"})
		return
	}

	// prepare detailed response
	response := gin.H{
		"id":            log.ID,
		"claude_model":  log.ClaudeModel,
		"openai_model":  log.OpenaiModel,
		"status_code":   log.StatusCode,
		"is_streaming":  log.IsStreaming,
		"input_tokens":  log.InputTokens,
		"output_tokens": log.OutputTokens,
		"duration_ms":   log.DurationMs,
		"created_at":    log.CreatedAt,
		"request_body":  log.RequestBody,
		"response_body": log.ResponseBody,
	}

	if log.ErrorMessage != "" {
		response["error_message"] = log.ErrorMessage
	}

	c.JSON(http.StatusOK, gin.H{
		"log": response,
	})
}

// DeleteOldLogs deletes expired request logs
func (h *RequestLogsHandler) DeleteOldLogs(c *gin.Context) {
	// default to delete logs older than 30 days
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter"})
		return
	}

	beforeTime := time.Now().AddDate(0, 0, -days)

	err = h.requestLogger.DeleteOldLogs(c.Request.Context(), beforeTime)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete old logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Old logs deleted successfully",
		"before_time": beforeTime.Format(time.RFC3339),
		"days":        days,
	})
}

// GetDashboardData gets dashboard data
func (h *RequestLogsHandler) GetDashboardData(c *gin.Context) {
	now := time.Now()

	// get last 24 hours data
	last24h := &logging.RequestLogQueryOptions{
		StartTime: now.Add(-24 * time.Hour),
		EndTime:   now,
	}

	// get last 7 days data
	last7d := &logging.RequestLogQueryOptions{
		StartTime: now.AddDate(0, 0, -7),
		EndTime:   now,
	}

	// get last 30 days data
	last30d := &logging.RequestLogQueryOptions{
		StartTime: now.AddDate(0, 0, -30),
		EndTime:   now,
	}

	// concurrently get statistics
	stats24h, err := h.requestLogger.GetRequestLogStats(c.Request.Context(), last24h)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get 24h stats")
		stats24h = &logging.RequestLogStats{}
	}

	stats7d, err := h.requestLogger.GetRequestLogStats(c.Request.Context(), last7d)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get 7d stats")
		stats7d = &logging.RequestLogStats{}
	}

	stats30d, err := h.requestLogger.GetRequestLogStats(c.Request.Context(), last30d)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get 30d stats")
		stats30d = &logging.RequestLogStats{}
	}

	c.JSON(http.StatusOK, gin.H{
		"dashboard": gin.H{
			"last_24h": gin.H{
				"stats": stats24h,
				"success_rate": func() float64 {
					if stats24h.TotalRequests > 0 {
						return float64(stats24h.SuccessRequests) / float64(stats24h.TotalRequests) * 100
					}
					return 0
				}(),
			},
			"last_7d": gin.H{
				"stats": stats7d,
				"success_rate": func() float64 {
					if stats7d.TotalRequests > 0 {
						return float64(stats7d.SuccessRequests) / float64(stats7d.TotalRequests) * 100
					}
					return 0
				}(),
			},
			"last_30d": gin.H{
				"stats": stats30d,
				"success_rate": func() float64 {
					if stats30d.TotalRequests > 0 {
						return float64(stats30d.SuccessRequests) / float64(stats30d.TotalRequests) * 100
					}
					return 0
				}(),
			},
		},
		"timestamp": now.Format(time.RFC3339),
	})
}

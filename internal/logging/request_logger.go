package logging

import (
	"context"
	"encoding/json"
	"time"

	"ccany/ent"
	"ccany/ent/requestlog"

	"github.com/sirupsen/logrus"
)

// RequestLogger handles request logging
type RequestLogger struct {
	db     *ent.Client
	logger *logrus.Logger
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(db *ent.Client, logger *logrus.Logger) *RequestLogger {
	return &RequestLogger{
		db:     db,
		logger: logger,
	}
}

// LogRequest logs request information
func (rl *RequestLogger) LogRequest(ctx context.Context, req *RequestLogData) error {
	// serialize request and response body
	requestBody, err := json.Marshal(req.RequestBody)
	if err != nil {
		rl.logger.WithError(err).Error("Failed to marshal request body")
		requestBody = []byte("{}")
	}

	var responseBody []byte
	if req.ResponseBody != nil {
		responseBody, err = json.Marshal(req.ResponseBody)
		if err != nil {
			rl.logger.WithError(err).Error("Failed to marshal response body")
			responseBody = []byte("{}")
		}
	}

	// create request log record
	create := rl.db.RequestLog.Create().
		SetID(req.ID).
		SetClaudeModel(req.ClaudeModel).
		SetOpenaiModel(req.OpenAIModel).
		SetRequestBody(string(requestBody)).
		SetStatusCode(req.StatusCode).
		SetIsStreaming(req.IsStreaming).
		SetInputTokens(req.InputTokens).
		SetOutputTokens(req.OutputTokens).
		SetDurationMs(req.DurationMs).
		SetCreatedAt(req.CreatedAt)

	if responseBody != nil {
		create = create.SetResponseBody(string(responseBody))
	}

	if req.ErrorMessage != "" {
		create = create.SetErrorMessage(req.ErrorMessage)
	}

	_, err = create.Save(ctx)
	if err != nil {
		rl.logger.WithError(err).Error("Failed to save request log")
		return err
	}

	return nil
}

// GetRequestLogs retrieves request log list
func (rl *RequestLogger) GetRequestLogs(ctx context.Context, opts *RequestLogQueryOptions) ([]*ent.RequestLog, error) {
	query := rl.db.RequestLog.Query()

	// apply filters
	if opts.ClaudeModel != "" {
		query = query.Where(requestlog.ClaudeModelEQ(opts.ClaudeModel))
	}

	if opts.StatusCode != 0 {
		query = query.Where(requestlog.StatusCodeEQ(opts.StatusCode))
	}

	if opts.IsStreaming != nil {
		query = query.Where(requestlog.IsStreamingEQ(*opts.IsStreaming))
	}

	if !opts.StartTime.IsZero() {
		query = query.Where(requestlog.CreatedAtGTE(opts.StartTime))
	}

	if !opts.EndTime.IsZero() {
		query = query.Where(requestlog.CreatedAtLTE(opts.EndTime))
	}

	// sorting
	query = query.Order(ent.Desc(requestlog.FieldCreatedAt))

	// pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}

	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	return query.All(ctx)
}

// GetRequestLogStats retrieves request log statistics
func (rl *RequestLogger) GetRequestLogStats(ctx context.Context, opts *RequestLogQueryOptions) (*RequestLogStats, error) {
	query := rl.db.RequestLog.Query()

	// apply filters
	if opts.ClaudeModel != "" {
		query = query.Where(requestlog.ClaudeModelEQ(opts.ClaudeModel))
	}

	if !opts.StartTime.IsZero() {
		query = query.Where(requestlog.CreatedAtGTE(opts.StartTime))
	}

	if !opts.EndTime.IsZero() {
		query = query.Where(requestlog.CreatedAtLTE(opts.EndTime))
	}

	// get all records for statistics
	logs, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	stats := &RequestLogStats{
		TotalRequests:     len(logs),
		SuccessRequests:   0,
		ErrorRequests:     0,
		StreamingRequests: 0,
		TotalInputTokens:  0,
		TotalOutputTokens: 0,
		AvgDurationMs:     0,
		TotalDurationMs:   0,
	}

	var totalDuration float64
	for _, log := range logs {
		if log.StatusCode >= 200 && log.StatusCode < 300 {
			stats.SuccessRequests++
		} else {
			stats.ErrorRequests++
		}

		if log.IsStreaming {
			stats.StreamingRequests++
		}

		stats.TotalInputTokens += log.InputTokens
		stats.TotalOutputTokens += log.OutputTokens
		totalDuration += log.DurationMs
	}

	stats.TotalDurationMs = totalDuration
	if len(logs) > 0 {
		stats.AvgDurationMs = totalDuration / float64(len(logs))
	}

	return stats, nil
}

// GetRequestLogByID retrieves single request log by ID
func (rl *RequestLogger) GetRequestLogByID(ctx context.Context, id string) (*ent.RequestLog, error) {
	return rl.db.RequestLog.Query().
		Where(requestlog.IDEQ(id)).
		Only(ctx)
}

// DeleteOldLogs deletes expired log records
func (rl *RequestLogger) DeleteOldLogs(ctx context.Context, beforeTime time.Time) error {
	deleted, err := rl.db.RequestLog.Delete().
		Where(requestlog.CreatedAtLT(beforeTime)).
		Exec(ctx)
	if err != nil {
		return err
	}

	rl.logger.WithFields(logrus.Fields{
		"deleted_count": deleted,
		"before_time":   beforeTime.Format(time.RFC3339),
	}).Info("Deleted old request logs")

	return nil
}

// RequestLogData request log data structure
type RequestLogData struct {
	ID           string
	ClaudeModel  string
	OpenAIModel  string
	RequestBody  interface{}
	ResponseBody interface{}
	StatusCode   int
	IsStreaming  bool
	InputTokens  int
	OutputTokens int
	DurationMs   float64
	ErrorMessage string
	CreatedAt    time.Time
}

// RequestLogQueryOptions request log query options
type RequestLogQueryOptions struct {
	ClaudeModel string
	StatusCode  int
	IsStreaming *bool
	StartTime   time.Time
	EndTime     time.Time
	Limit       int
	Offset      int
}

// RequestLogStats request log statistics
type RequestLogStats struct {
	TotalRequests     int     `json:"total_requests"`
	SuccessRequests   int     `json:"success_requests"`
	ErrorRequests     int     `json:"error_requests"`
	StreamingRequests int     `json:"streaming_requests"`
	TotalInputTokens  int     `json:"total_input_tokens"`
	TotalOutputTokens int     `json:"total_output_tokens"`
	AvgDurationMs     float64 `json:"avg_duration_ms"`
	TotalDurationMs   float64 `json:"total_duration_ms"`
}

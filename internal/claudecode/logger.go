package claudecode

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// RouterLogger provides enhanced logging for Claude Code router
type RouterLogger struct {
	mu          sync.RWMutex
	logger      *logrus.Logger
	logFile     *os.File
	logPath     string
	enabled     bool
	fileLogging bool
}

// LoggerConfig contains logger configuration
type LoggerConfig struct {
	Enabled     bool
	FileLogging bool
	LogPath     string
	LogLevel    LogLevel
}

// NewRouterLogger creates a new router logger
func NewRouterLogger(config *LoggerConfig) (*RouterLogger, error) {
	rl := &RouterLogger{
		logger:      logrus.New(),
		enabled:     config.Enabled,
		fileLogging: config.FileLogging,
		logPath:     config.LogPath,
	}

	// Set log level
	switch config.LogLevel {
	case LogLevelDebug:
		rl.logger.SetLevel(logrus.DebugLevel)
	case LogLevelInfo:
		rl.logger.SetLevel(logrus.InfoLevel)
	case LogLevelWarn:
		rl.logger.SetLevel(logrus.WarnLevel)
	case LogLevelError:
		rl.logger.SetLevel(logrus.ErrorLevel)
	default:
		rl.logger.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	rl.logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     false,
	})

	// Setup file logging if enabled
	if config.FileLogging && config.LogPath != "" {
		if err := rl.setupFileLogging(); err != nil {
			return nil, fmt.Errorf("failed to setup file logging: %w", err)
		}
	}

	return rl, nil
}

// setupFileLogging sets up logging to file
func (rl *RouterLogger) setupFileLogging() error {
	// Create log directory if needed
	logDir := filepath.Dir(rl.logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(rl.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	rl.logFile = file

	// Set multi-writer output (console + file)
	multiWriter := io.MultiWriter(os.Stdout, file)
	rl.logger.SetOutput(multiWriter)

	return nil
}

// Log logs a message if logging is enabled
func (rl *RouterLogger) Log(level LogLevel, message string, fields map[string]interface{}) {
	if !rl.enabled {
		return
	}

	entry := rl.logger.WithFields(logrus.Fields(fields))

	switch level {
	case LogLevelDebug:
		entry.Debug(message)
	case LogLevelInfo:
		entry.Info(message)
	case LogLevelWarn:
		entry.Warn(message)
	case LogLevelError:
		entry.Error(message)
	}
}

// LogRoutingDecision logs a routing decision with detailed information
func (rl *RouterLogger) LogRoutingDecision(decision *RoutingDecision) {
	if !rl.enabled {
		return
	}

	fields := map[string]interface{}{
		"original_model":   decision.OriginalModel,
		"routed_model":     decision.RoutedModel,
		"reason":           decision.Reason,
		"token_count":      decision.TokenCount,
		"has_tools":        decision.HasTools,
		"has_thinking":     decision.HasThinking,
		"message_count":    decision.MessageCount,
		"routing_duration": decision.Duration.Milliseconds(),
	}

	// Add custom metadata
	for k, v := range decision.Metadata {
		fields[k] = v
	}

	rl.Log(LogLevelInfo, "Routing decision made", fields)
}

// LogRequestMetrics logs request metrics
func (rl *RouterLogger) LogRequestMetrics(metrics *RequestMetrics) {
	if !rl.enabled {
		return
	}

	fields := map[string]interface{}{
		"request_id":    metrics.RequestID,
		"model":         metrics.Model,
		"input_tokens":  metrics.InputTokens,
		"output_tokens": metrics.OutputTokens,
		"total_tokens":  metrics.TotalTokens,
		"latency_ms":    metrics.Latency.Milliseconds(),
		"status":        metrics.Status,
		"error":         metrics.Error,
		"cache_hit":     metrics.CacheHit,
		"stream":        metrics.Stream,
	}

	level := LogLevelInfo
	if metrics.Error != "" {
		level = LogLevelError
	}

	rl.Log(level, "Request processed", fields)
}

// LogToolExecution logs tool execution details
func (rl *RouterLogger) LogToolExecution(toolLog *ToolExecutionLog) {
	if !rl.enabled {
		return
	}

	fields := map[string]interface{}{
		"tool_name":    toolLog.ToolName,
		"tool_type":    toolLog.ToolType,
		"execution_ms": toolLog.ExecutionTime.Milliseconds(),
		"success":      toolLog.Success,
		"error":        toolLog.Error,
		"input_size":   toolLog.InputSize,
		"output_size":  toolLog.OutputSize,
	}

	level := LogLevelInfo
	if !toolLog.Success {
		level = LogLevelError
	}

	rl.Log(level, "Tool executed", fields)
}

// Close closes the logger and releases resources
func (rl *RouterLogger) Close() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.logFile != nil {
		return rl.logFile.Close()
	}
	return nil
}

// SetEnabled enables or disables logging
func (rl *RouterLogger) SetEnabled(enabled bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = enabled
}

// IsEnabled returns whether logging is enabled
func (rl *RouterLogger) IsEnabled() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.enabled
}

// RoutingDecision represents a routing decision
type RoutingDecision struct {
	OriginalModel string
	RoutedModel   string
	Reason        string
	TokenCount    int
	HasTools      bool
	HasThinking   bool
	MessageCount  int
	Duration      time.Duration
	Metadata      map[string]interface{}
}

// RequestMetrics represents request processing metrics
type RequestMetrics struct {
	RequestID    string
	Model        string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Latency      time.Duration
	Status       string
	Error        string
	CacheHit     bool
	Stream       bool
}

// ToolExecutionLog represents tool execution information
type ToolExecutionLog struct {
	ToolName      string
	ToolType      string
	ExecutionTime time.Duration
	Success       bool
	Error         string
	InputSize     int
	OutputSize    int
}

// NewLoggerFromEnv creates a logger based on environment variables
func NewLoggerFromEnv() (*RouterLogger, error) {
	config := &LoggerConfig{
		Enabled:     os.Getenv("CLAUDE_CODE_LOG") == "true",
		FileLogging: os.Getenv("CLAUDE_CODE_LOG_FILE") != "",
		LogPath:     os.Getenv("CLAUDE_CODE_LOG_FILE"),
		LogLevel:    LogLevel(os.Getenv("CLAUDE_CODE_LOG_LEVEL")),
	}

	if config.LogPath == "" && config.FileLogging {
		// Default log path
		homeDir, _ := os.UserHomeDir()
		config.LogPath = filepath.Join(homeDir, ".ccany", "claude-code-router.log")
	}

	if config.LogLevel == "" {
		config.LogLevel = LogLevelInfo
	}

	return NewRouterLogger(config)
}

// LogMiddleware provides request/response logging middleware
func LogMiddleware(logger *RouterLogger) func(next func(interface{}) interface{}) func(interface{}) interface{} {
	return func(next func(interface{}) interface{}) func(interface{}) interface{} {
		return func(req interface{}) interface{} {
			start := time.Now()

			// Execute the next handler
			resp := next(req)

			// Log the request/response
			duration := time.Since(start)

			logger.Log(LogLevelInfo, "Request processed", map[string]interface{}{
				"duration_ms": duration.Milliseconds(),
				"timestamp":   start.Format(time.RFC3339),
			})

			return resp
		}
	}
}

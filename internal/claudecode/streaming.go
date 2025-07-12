package claudecode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ccany/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// StreamingService handles Claude Code compatible streaming responses
type StreamingService struct {
	logger *logrus.Logger
}

// NewStreamingService creates a new streaming service
func NewStreamingService(logger *logrus.Logger) *StreamingService {
	return &StreamingService{
		logger: logger,
	}
}

// StreamingContext holds the context for a streaming request
type StreamingContext struct {
	RequestID      string
	Model          string
	StartTime      time.Time
	InputTokens    int
	OutputTokens   int
	HasError       bool
	ErrorMessage   string
	ContentBuffer  strings.Builder
	IsFirstChunk   bool
	ToolCallBuffer map[string]interface{}
}

// InitializeStreaming sets up the streaming response headers and context
func (s *StreamingService) InitializeStreaming(c *gin.Context, requestID, model string) *StreamingContext {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	ctx := &StreamingContext{
		RequestID:      requestID,
		Model:          model,
		StartTime:      time.Now(),
		IsFirstChunk:   true,
		ToolCallBuffer: make(map[string]interface{}),
	}

	// Send message_start event
	s.sendMessageStartEvent(c, ctx)

	// Send content_block_start event
	s.sendContentBlockStartEvent(c, ctx)

	// Send initial ping event
	s.sendPingEvent(c)

	return ctx
}

// sendMessageStartEvent sends the message_start SSE event
func (s *StreamingService) sendMessageStartEvent(c *gin.Context, ctx *StreamingContext) {
	startEvent := models.ClaudeStreamStartEvent{
		Type: "message_start",
		Message: models.ClaudeResponse{
			ID:           ctx.RequestID,
			Type:         "message",
			Role:         "assistant",
			Content:      []models.ClaudeContentBlock{},
			Model:        ctx.Model,
			StopReason:   "",
			StopSequence: nil,
			Usage: models.ClaudeUsage{
				InputTokens:  ctx.InputTokens,
				OutputTokens: ctx.OutputTokens,
			},
		},
	}

	s.writeSSEEvent(c, "message_start", startEvent)
}

// sendContentBlockStartEvent sends the content_block_start SSE event
func (s *StreamingService) sendContentBlockStartEvent(c *gin.Context, ctx *StreamingContext) {
	contentBlockStartEvent := models.ClaudeStreamContentBlockStartEvent{
		Type:  "content_block_start",
		Index: 0,
		ContentBlock: models.ClaudeContentBlock{
			Type: "text",
			Text: "",
		},
	}

	s.writeSSEEvent(c, "content_block_start", contentBlockStartEvent)
}

// sendPingEvent sends a ping event for keep-alive
func (s *StreamingService) sendPingEvent(c *gin.Context) {
	pingEvent := models.ClaudeStreamPingEvent{
		Type: "ping",
	}

	s.writeSSEEvent(c, "ping", pingEvent)
}

// ProcessTextChunk processes a text chunk from the OpenAI stream
func (s *StreamingService) ProcessTextChunk(c *gin.Context, ctx *StreamingContext, text string) {
	if text == "" {
		return
	}

	// Add to content buffer
	ctx.ContentBuffer.WriteString(text)

	// Send content_block_delta event
	deltaEvent := models.ClaudeStreamEvent{
		Type:  "content_block_delta",
		Index: 0,
		Delta: &models.ClaudeContentBlock{
			Type: "text_delta",
			Text: text,
		},
	}

	s.writeSSEEvent(c, "content_block_delta", deltaEvent)
}

// ProcessToolCall processes a tool call from the OpenAI stream
func (s *StreamingService) ProcessToolCall(c *gin.Context, ctx *StreamingContext, toolCall interface{}) {
	// Handle tool call streaming with incremental JSON parsing
	if toolCallMap, ok := toolCall.(map[string]interface{}); ok {
		if function, exists := toolCallMap["function"]; exists {
			if functionMap, ok := function.(map[string]interface{}); ok {
				if args, exists := functionMap["arguments"]; exists {
					if argsStr, ok := args.(string); ok {
						// Send input_json_delta event for incremental tool arguments
						deltaEvent := models.ClaudeStreamEvent{
							Type:  "content_block_delta",
							Index: 0,
							Delta: &models.ClaudeContentBlock{
								Type: "input_json_delta",
								Input: map[string]interface{}{
									"partial_json": argsStr,
								},
							},
						}

						s.writeSSEEvent(c, "content_block_delta", deltaEvent)
					}
				}
			}
		}
	}
}

// CheckClientDisconnect checks if the client has disconnected
func (s *StreamingService) CheckClientDisconnect(c *gin.Context) bool {
	// Check if the client context is done
	select {
	case <-c.Request.Context().Done():
		s.logger.Debug("Client disconnected")
		return true
	default:
		return false
	}
}

// SendPeriodicPing sends periodic ping events during streaming
func (s *StreamingService) SendPeriodicPing(c *gin.Context, ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendPingEvent(c)
		}
	}
}

// FinalizeStreaming sends the final events and completes the stream
func (s *StreamingService) FinalizeStreaming(c *gin.Context, ctx *StreamingContext, stopReason string) {
	// Send content_block_stop event
	s.sendContentBlockStopEvent(c, ctx)

	// Send message_delta event with stop reason and usage
	s.sendMessageDeltaEvent(c, ctx, stopReason)

	// Send message_stop event
	s.sendMessageStopEvent(c, ctx)

	// Log completion
	duration := time.Since(ctx.StartTime)
	s.logger.WithFields(logrus.Fields{
		"request_id":    ctx.RequestID,
		"duration":      duration,
		"input_tokens":  ctx.InputTokens,
		"output_tokens": ctx.OutputTokens,
		"stop_reason":   stopReason,
		"has_error":     ctx.HasError,
	}).Info("Claude Code streaming completed")
}

// sendContentBlockStopEvent sends the content_block_stop SSE event
func (s *StreamingService) sendContentBlockStopEvent(c *gin.Context, ctx *StreamingContext) {
	contentBlockStopEvent := models.ClaudeStreamContentBlockStopEvent{
		Type:  "content_block_stop",
		Index: 0,
	}

	s.writeSSEEvent(c, "content_block_stop", contentBlockStopEvent)
}

// sendMessageDeltaEvent sends the message_delta SSE event
func (s *StreamingService) sendMessageDeltaEvent(c *gin.Context, ctx *StreamingContext, stopReason string) {
	messageDeltaEvent := models.ClaudeStreamMessageDeltaEvent{
		Type: "message_delta",
		Delta: models.ClaudeMessageDelta{
			StopReason:   stopReason,
			StopSequence: nil,
		},
		Usage: &models.ClaudeUsage{
			InputTokens:  ctx.InputTokens,
			OutputTokens: ctx.OutputTokens,
		},
	}

	s.writeSSEEvent(c, "message_delta", messageDeltaEvent)
}

// sendMessageStopEvent sends the message_stop SSE event
func (s *StreamingService) sendMessageStopEvent(c *gin.Context, ctx *StreamingContext) {
	messageStopEvent := models.ClaudeStreamMessageStopEvent{
		Type: "message_stop",
	}

	s.writeSSEEvent(c, "message_stop", messageStopEvent)
}

// HandleStreamingError handles errors during streaming
func (s *StreamingService) HandleStreamingError(c *gin.Context, ctx *StreamingContext, err error) {
	ctx.HasError = true
	ctx.ErrorMessage = err.Error()

	// Send error event
	errorEvent := models.ClaudeErrorResponse{
		Type: "error",
		Error: models.ClaudeError{
			Type:    "api_error",
			Message: err.Error(),
		},
	}

	s.writeSSEEvent(c, "error", errorEvent)

	s.logger.WithFields(logrus.Fields{
		"request_id": ctx.RequestID,
		"error":      err.Error(),
	}).Error("Streaming error occurred")
}

// writeSSEEvent writes an SSE event to the response
func (s *StreamingService) writeSSEEvent(c *gin.Context, eventType string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal SSE event")
		return
	}

	// Write event type
	if _, err := fmt.Fprintf(c.Writer, "event: %s\n", eventType); err != nil {
		s.logger.WithError(err).Error("Failed to write SSE event type")
		return
	}

	// Write data
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", string(jsonData)); err != nil {
		s.logger.WithError(err).Error("Failed to write SSE data")
		return
	}

	// Flush immediately
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// UpdateUsageTokens updates the token counts in the streaming context
func (s *StreamingService) UpdateUsageTokens(ctx *StreamingContext, inputTokens, outputTokens int) {
	ctx.InputTokens = inputTokens
	ctx.OutputTokens = outputTokens
}

// GetStreamingStats returns statistics about the streaming request
func (s *StreamingService) GetStreamingStats(ctx *StreamingContext) map[string]interface{} {
	duration := time.Since(ctx.StartTime)

	return map[string]interface{}{
		"request_id":     ctx.RequestID,
		"duration_ms":    duration.Milliseconds(),
		"input_tokens":   ctx.InputTokens,
		"output_tokens":  ctx.OutputTokens,
		"content_length": ctx.ContentBuffer.Len(),
		"has_error":      ctx.HasError,
		"error_message":  ctx.ErrorMessage,
	}
}

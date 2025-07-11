package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ccany/internal/client"
	"ccany/internal/config"
	"ccany/internal/converter"
	"ccany/internal/logging"
	"ccany/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MessagesHandler handles Claude messages API requests
type MessagesHandler struct {
	config        *config.Config
	openaiClient  *client.OpenAIClient
	requestLogger *logging.RequestLogger
	logger        *logrus.Logger
}

// NewMessagesHandler creates a new messages handler
func NewMessagesHandler(cfg *config.Config, openaiClient *client.OpenAIClient, requestLogger *logging.RequestLogger, logger *logrus.Logger) *MessagesHandler {
	return &MessagesHandler{
		config:        cfg,
		openaiClient:  openaiClient,
		requestLogger: requestLogger,
		logger:        logger,
	}
}

// CreateMessage handles POST /v1/messages
func (h *MessagesHandler) CreateMessage(c *gin.Context) {
	var claudeReq models.ClaudeMessagesRequest
	if err := c.ShouldBindJSON(&claudeReq); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, converter.CreateClaudeErrorResponse("invalid_request_error", "Invalid request format"))
		return
	}

	requestID := uuid.New().String()

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"model":      claudeReq.Model,
		"stream":     claudeReq.Stream,
		"max_tokens": claudeReq.MaxTokens,
	}).Info("Processing Claude messages request")

	// Convert Claude request to OpenAI format
	openaiReq, err := converter.ConvertClaudeToOpenAI(&claudeReq, h.config.BigModel, h.config.SmallModel)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert request")
		c.JSON(http.StatusBadRequest, converter.CreateClaudeErrorResponse("invalid_request_error", "Failed to convert request"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.RequestTimeout)*time.Second)
	defer cancel()

	if claudeReq.Stream {
		h.handleStreamingRequest(c, ctx, requestID, &claudeReq, openaiReq)
	} else {
		h.handleNonStreamingRequest(c, ctx, requestID, &claudeReq, openaiReq)
	}
}

// handleNonStreamingRequest handles non-streaming requests
func (h *MessagesHandler) handleNonStreamingRequest(c *gin.Context, ctx context.Context, requestID string, claudeReq *models.ClaudeMessagesRequest, openaiReq *models.OpenAIChatCompletionRequest) {
	startTime := time.Now()

	// Send request to OpenAI
	openaiResp, err := h.openaiClient.CreateChatCompletion(ctx, openaiReq)
	duration := time.Since(startTime)

	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"duration":   duration,
			"error":      err.Error(),
		}).Error("OpenAI request failed")

		// log failed request
		if h.requestLogger != nil {
			logData := &logging.RequestLogData{
				ID:           requestID,
				ClaudeModel:  claudeReq.Model,
				OpenAIModel:  openaiReq.Model,
				RequestBody:  claudeReq,
				ResponseBody: nil,
				StatusCode:   http.StatusInternalServerError,
				IsStreaming:  false,
				InputTokens:  0,
				OutputTokens: 0,
				DurationMs:   float64(duration.Milliseconds()),
				ErrorMessage: err.Error(),
				CreatedAt:    startTime,
			}
			go func() {
				if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
					h.logger.WithError(err).Error("Failed to log request")
				}
			}()
		}

		errorType := h.openaiClient.ClassifyError(err)
		c.JSON(http.StatusInternalServerError, converter.CreateClaudeErrorResponse(errorType, err.Error()))
		return
	}

	// Convert OpenAI response to Claude format
	claudeResp, err := converter.ConvertOpenAIToClaudeResponse(openaiResp, claudeReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert response")

		// log conversion failed request
		if h.requestLogger != nil {
			logData := &logging.RequestLogData{
				ID:           requestID,
				ClaudeModel:  claudeReq.Model,
				OpenAIModel:  openaiReq.Model,
				RequestBody:  claudeReq,
				ResponseBody: openaiResp,
				StatusCode:   http.StatusInternalServerError,
				IsStreaming:  false,
				InputTokens:  0,
				OutputTokens: 0,
				DurationMs:   float64(duration.Milliseconds()),
				ErrorMessage: "Failed to convert response",
				CreatedAt:    startTime,
			}
			go func() {
				if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
					h.logger.WithError(err).Error("Failed to log request")
				}
			}()
		}

		c.JSON(http.StatusInternalServerError, converter.CreateClaudeErrorResponse("api_error", "Failed to convert response"))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":    requestID,
		"duration":      duration,
		"input_tokens":  claudeResp.Usage.InputTokens,
		"output_tokens": claudeResp.Usage.OutputTokens,
		"stop_reason":   claudeResp.StopReason,
	}).Info("Request completed successfully")

	// log successful request
	if h.requestLogger != nil {
		logData := &logging.RequestLogData{
			ID:           requestID,
			ClaudeModel:  claudeReq.Model,
			OpenAIModel:  openaiReq.Model,
			RequestBody:  claudeReq,
			ResponseBody: claudeResp,
			StatusCode:   http.StatusOK,
			IsStreaming:  false,
			InputTokens:  claudeResp.Usage.InputTokens,
			OutputTokens: claudeResp.Usage.OutputTokens,
			DurationMs:   float64(duration.Milliseconds()),
			ErrorMessage: "",
			CreatedAt:    startTime,
		}
		go func() {
			if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
				h.logger.WithError(err).Error("Failed to log request")
			}
		}()
	}

	c.JSON(http.StatusOK, claudeResp)
}

// handleStreamingRequest handles streaming requests
func (h *MessagesHandler) handleStreamingRequest(c *gin.Context, ctx context.Context, requestID string, claudeReq *models.ClaudeMessagesRequest, openaiReq *models.OpenAIChatCompletionRequest) {
	startTime := time.Now()

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")

	// Get streaming response from OpenAI
	streamChan, err := h.openaiClient.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		duration := time.Since(startTime)
		h.logger.WithError(err).Error("Failed to create stream")

		// log streaming request failure
		if h.requestLogger != nil {
			logData := &logging.RequestLogData{
				ID:           requestID,
				ClaudeModel:  claudeReq.Model,
				OpenAIModel:  openaiReq.Model,
				RequestBody:  claudeReq,
				ResponseBody: nil,
				StatusCode:   http.StatusInternalServerError,
				IsStreaming:  true,
				InputTokens:  0,
				OutputTokens: 0,
				DurationMs:   float64(duration.Milliseconds()),
				ErrorMessage: err.Error(),
				CreatedAt:    startTime,
			}
			go func() {
				if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
					h.logger.WithError(err).Error("Failed to log request")
				}
			}()
		}

		h.writeSSEError(c, converter.CreateClaudeErrorResponse(h.openaiClient.ClassifyError(err), err.Error()))
		return
	}

	// Send stream start event
	startEvent := converter.CreateClaudeStreamStartEvent(requestID, claudeReq.Model)
	h.writeSSEEvent(c, "message_start", startEvent)

	var totalInputTokens, totalOutputTokens int
	hasError := false
	var streamError error

	// Process stream chunks
	for chunk := range streamChan {
		select {
		case <-ctx.Done():
			h.logger.WithField("request_id", requestID).Warn("Request context cancelled")
			hasError = true
			streamError = ctx.Err()
			goto exitLoop
		default:
		}

		if chunk.Error != nil {
			h.logger.WithError(chunk.Error).Error("Stream error")
			h.writeSSEError(c, converter.CreateClaudeErrorResponse("api_error", chunk.Error.Error()))
			hasError = true
			streamError = chunk.Error
			goto exitLoop
		}

		if chunk.Done {
			goto exitLoop
		}

		if chunk.Data != nil {
			// Convert OpenAI stream chunk to Claude events
			events, err := converter.ConvertOpenAIStreamToClaudeStream(chunk.Data, claudeReq)
			if err != nil {
				h.logger.WithError(err).Warn("Failed to convert stream chunk")
				continue
			}

			// Send each event
			for _, event := range events {
				h.writeSSEEvent(c, event.Type, event)
			}
		}
	}

exitLoop:
	// Send final usage event
	usage := models.ClaudeUsage{
		InputTokens:  totalInputTokens,
		OutputTokens: totalOutputTokens,
	}
	stopEvent := converter.CreateClaudeStreamStopEvent(usage)
	h.writeSSEEvent(c, "message_delta", stopEvent)

	duration := time.Since(startTime)
	h.logger.WithFields(logrus.Fields{
		"request_id":    requestID,
		"duration":      duration,
		"input_tokens":  totalInputTokens,
		"output_tokens": totalOutputTokens,
	}).Info("Streaming request completed")

	// log streaming request result
	if h.requestLogger != nil {
		statusCode := http.StatusOK
		errorMessage := ""

		if hasError {
			statusCode = http.StatusInternalServerError
			if streamError != nil {
				errorMessage = streamError.Error()
			}
		}

		logData := &logging.RequestLogData{
			ID:          requestID,
			ClaudeModel: claudeReq.Model,
			OpenAIModel: openaiReq.Model,
			RequestBody: claudeReq,
			ResponseBody: gin.H{
				"usage":     usage,
				"streaming": true,
			},
			StatusCode:   statusCode,
			IsStreaming:  true,
			InputTokens:  totalInputTokens,
			OutputTokens: totalOutputTokens,
			DurationMs:   float64(duration.Milliseconds()),
			ErrorMessage: errorMessage,
			CreatedAt:    startTime,
		}
		go func() {
			if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
				h.logger.WithError(err).Error("Failed to log request")
			}
		}()
	}
}

// writeSSEEvent writes an SSE event to the response
func (h *MessagesHandler) writeSSEEvent(c *gin.Context, eventType string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal SSE event")
		return
	}

	if _, err := fmt.Fprintf(c.Writer, "event: %s\n", eventType); err != nil {
		h.logger.WithError(err).Error("Failed to write SSE event type")
		return
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", string(jsonData)); err != nil {
		h.logger.WithError(err).Error("Failed to write SSE data")
		return
	}
	c.Writer.Flush()
}

// writeSSEError writes an SSE error event
func (h *MessagesHandler) writeSSEError(c *gin.Context, errorResp *models.ClaudeErrorResponse) {
	h.writeSSEEvent(c, "error", errorResp)
}

// CountTokens handles POST /v1/messages/count_tokens
func (h *MessagesHandler) CountTokens(c *gin.Context) {
	var req models.ClaudeTokenCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, converter.CreateClaudeErrorResponse("invalid_request_error", "Invalid request format"))
		return
	}

	// Simple token estimation (4 characters per token)
	totalChars := 0

	// Count system message characters
	if req.System != nil {
		if systemStr, ok := req.System.(string); ok {
			totalChars += len(systemStr)
		}
	}

	// Count message characters
	for _, msg := range req.Messages {
		if msg.Content != nil {
			if contentStr, ok := msg.Content.(string); ok {
				totalChars += len(contentStr)
			} else if contentBlocks, ok := msg.Content.([]interface{}); ok {
				for _, block := range contentBlocks {
					if blockMap, ok := block.(map[string]interface{}); ok {
						if text, exists := blockMap["text"]; exists {
							if textStr, ok := text.(string); ok {
								totalChars += len(textStr)
							}
						}
					}
				}
			}
		}
	}

	// Rough estimation: 4 characters per token
	estimatedTokens := max(1, totalChars/4)

	c.JSON(http.StatusOK, gin.H{
		"input_tokens": estimatedTokens,
	})
}

// CreateChatCompletion handles POST /v1/chat/completions (OpenAI API compatible)
func (h *MessagesHandler) CreateChatCompletion(c *gin.Context) {
	var openaiReq models.OpenAIChatCompletionRequest
	if err := c.ShouldBindJSON(&openaiReq); err != nil {
		h.logger.WithError(err).Error("Failed to bind OpenAI request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request format",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	requestID := uuid.New().String()

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"model":      openaiReq.Model,
		"stream":     openaiReq.Stream,
		"max_tokens": openaiReq.MaxTokens,
	}).Info("Processing OpenAI chat completion request")

	// use model mapping from configuration
	switch openaiReq.Model {
	case "claude-3-5-sonnet-20241022", "claude-3-haiku-20240307":
		openaiReq.Model = h.config.BigModel
	default:
		openaiReq.Model = h.config.SmallModel
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.RequestTimeout)*time.Second)
	defer cancel()

	if openaiReq.Stream {
		h.handleOpenAIStreamingRequest(c, ctx, requestID, &openaiReq)
	} else {
		h.handleOpenAINonStreamingRequest(c, ctx, requestID, &openaiReq)
	}
}

// handleOpenAINonStreamingRequest handles non-streaming OpenAI requests
func (h *MessagesHandler) handleOpenAINonStreamingRequest(c *gin.Context, ctx context.Context, requestID string, openaiReq *models.OpenAIChatCompletionRequest) {
	startTime := time.Now()

	// Send request to OpenAI
	openaiResp, err := h.openaiClient.CreateChatCompletion(ctx, openaiReq)
	duration := time.Since(startTime)

	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"duration":   duration,
			"error":      err.Error(),
		}).Error("OpenAI request failed")

		// log failed request
		if h.requestLogger != nil {
			logData := &logging.RequestLogData{
				ID:           requestID,
				ClaudeModel:  "openai-direct",
				OpenAIModel:  openaiReq.Model,
				RequestBody:  openaiReq,
				ResponseBody: nil,
				StatusCode:   http.StatusInternalServerError,
				IsStreaming:  false,
				InputTokens:  0,
				OutputTokens: 0,
				DurationMs:   float64(duration.Milliseconds()),
				ErrorMessage: err.Error(),
				CreatedAt:    startTime,
			}
			go func() {
				if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
					h.logger.WithError(err).Error("Failed to log request")
				}
			}()
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "api_error",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"duration":   duration,
	}).Info("OpenAI request completed successfully")

	// log successful request
	if h.requestLogger != nil {
		logData := &logging.RequestLogData{
			ID:           requestID,
			ClaudeModel:  "openai-direct",
			OpenAIModel:  openaiReq.Model,
			RequestBody:  openaiReq,
			ResponseBody: openaiResp,
			StatusCode:   http.StatusOK,
			IsStreaming:  false,
			InputTokens:  0,
			OutputTokens: 0,
			DurationMs:   float64(duration.Milliseconds()),
			ErrorMessage: "",
			CreatedAt:    startTime,
		}
		go func() {
			if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
				h.logger.WithError(err).Error("Failed to log request")
			}
		}()
	}

	c.JSON(http.StatusOK, openaiResp)
}

// handleOpenAIStreamingRequest handles streaming OpenAI requests
func (h *MessagesHandler) handleOpenAIStreamingRequest(c *gin.Context, ctx context.Context, requestID string, openaiReq *models.OpenAIChatCompletionRequest) {
	startTime := time.Now()

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")

	// Get streaming response from OpenAI
	streamChan, err := h.openaiClient.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		duration := time.Since(startTime)
		h.logger.WithError(err).Error("Failed to create OpenAI stream")

		// log streaming request failure
		if h.requestLogger != nil {
			logData := &logging.RequestLogData{
				ID:           requestID,
				ClaudeModel:  "openai-direct",
				OpenAIModel:  openaiReq.Model,
				RequestBody:  openaiReq,
				ResponseBody: nil,
				StatusCode:   http.StatusInternalServerError,
				IsStreaming:  true,
				InputTokens:  0,
				OutputTokens: 0,
				DurationMs:   float64(duration.Milliseconds()),
				ErrorMessage: err.Error(),
				CreatedAt:    startTime,
			}
			go func() {
				if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
					h.logger.WithError(err).Error("Failed to log request")
				}
			}()
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": err.Error(),
				"type":    "api_error",
			},
		})
		return
	}

	var totalInputTokens, totalOutputTokens int
	hasError := false
	var streamError error

	// Process stream chunks
	for chunk := range streamChan {
		select {
		case <-ctx.Done():
			h.logger.WithField("request_id", requestID).Warn("Request context cancelled")
			hasError = true
			streamError = ctx.Err()
			goto exitOpenAILoop
		default:
		}

		if chunk.Error != nil {
			h.logger.WithError(chunk.Error).Error("Stream error")
			hasError = true
			streamError = chunk.Error
			goto exitOpenAILoop
		}

		if chunk.Done {
			goto exitOpenAILoop
		}

		if chunk.Data != nil {
			// directly output OpenAI format streaming response
			jsonData, err := json.Marshal(chunk.Data)
			if err != nil {
				h.logger.WithError(err).Error("Failed to marshal stream chunk")
				continue
			}

			if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", string(jsonData)); err != nil {
				h.logger.WithError(err).Error("Failed to write stream data")
				continue
			}
			c.Writer.Flush()
		}
	}

exitOpenAILoop:
	// Send [DONE] signal
	if _, err := fmt.Fprintf(c.Writer, "data: [DONE]\n\n"); err != nil {
		h.logger.WithError(err).Error("Failed to write [DONE] signal")
	}
	c.Writer.Flush()

	duration := time.Since(startTime)
	h.logger.WithFields(logrus.Fields{
		"request_id":    requestID,
		"duration":      duration,
		"input_tokens":  totalInputTokens,
		"output_tokens": totalOutputTokens,
	}).Info("OpenAI streaming request completed")

	// log streaming request result
	if h.requestLogger != nil {
		statusCode := http.StatusOK
		errorMessage := ""

		if hasError {
			statusCode = http.StatusInternalServerError
			if streamError != nil {
				errorMessage = streamError.Error()
			}
		}

		logData := &logging.RequestLogData{
			ID:          requestID,
			ClaudeModel: "openai-direct",
			OpenAIModel: openaiReq.Model,
			RequestBody: openaiReq,
			ResponseBody: gin.H{
				"streaming": true,
			},
			StatusCode:   statusCode,
			IsStreaming:  true,
			InputTokens:  totalInputTokens,
			OutputTokens: totalOutputTokens,
			DurationMs:   float64(duration.Milliseconds()),
			ErrorMessage: errorMessage,
			CreatedAt:    startTime,
		}
		go func() {
			if err := h.requestLogger.LogRequest(context.Background(), logData); err != nil {
				h.logger.WithError(err).Error("Failed to log request")
			}
		}()
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

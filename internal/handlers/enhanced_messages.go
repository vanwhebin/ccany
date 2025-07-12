package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ccany/internal/app"
	"ccany/internal/claudecode"
	"ccany/internal/client"
	"ccany/internal/config"
	"ccany/internal/converter"
	"ccany/internal/logging"
	"ccany/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EnhancedMessagesHandler handles Claude messages API requests with Claude Code compatibility
type EnhancedMessagesHandler struct {
	config           *config.Config
	configManager    *app.ConfigManager
	openaiClient     *client.OpenAIClient
	requestLogger    *logging.RequestLogger
	logger           *logrus.Logger
	streamingService *claudecode.StreamingService
	modelRouter      *claudecode.ModelRouter
	configService    *claudecode.ConfigService
}

// NewEnhancedMessagesHandler creates a new enhanced messages handler
func NewEnhancedMessagesHandler(cfg *config.Config, configManager *app.ConfigManager, openaiClient *client.OpenAIClient, requestLogger *logging.RequestLogger, logger *logrus.Logger) *EnhancedMessagesHandler {
	return &EnhancedMessagesHandler{
		config:           cfg,
		configManager:    configManager,
		openaiClient:     openaiClient,
		requestLogger:    requestLogger,
		logger:           logger,
		streamingService: claudecode.NewStreamingService(logger),
		modelRouter:      claudecode.NewModelRouter(logger, cfg.BigModel, cfg.SmallModel),
		configService:    claudecode.NewConfigService(logger),
	}
}

// getCurrentOpenAIClient creates a fresh OpenAI client using current configuration
func (h *EnhancedMessagesHandler) getCurrentOpenAIClient() (*client.OpenAIClient, *config.Config, error) {
	cfg, err := h.configManager.GetConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current config: %w", err)
	}

	if cfg.OpenAIAPIKey == "" {
		return nil, cfg, fmt.Errorf("OpenAI API key not configured")
	}

	openaiClient := client.NewOpenAIClient(
		cfg.OpenAIAPIKey,
		cfg.OpenAIBaseURL,
		cfg.RequestTimeout,
		h.logger,
	)

	return openaiClient, cfg, nil
}

// CreateMessage handles POST /v1/messages with Claude Code compatibility
func (h *EnhancedMessagesHandler) CreateMessage(c *gin.Context) {
	// Get current OpenAI client with fresh configuration
	openaiClient, cfg, err := h.getCurrentOpenAIClient()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get current OpenAI client")
		c.JSON(http.StatusServiceUnavailable, converter.CreateClaudeErrorResponse("service_unavailable", "OpenAI API is not configured. Please configure OpenAI API key in settings."))
		return
	}

	var claudeReq models.ClaudeMessagesRequest
	if err := c.ShouldBindJSON(&claudeReq); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, converter.CreateClaudeErrorResponse("invalid_request_error", "Invalid request format"))
		return
	}

	requestID := uuid.New().String()

	// Log enhanced request info including Claude Code specific fields
	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"model":      claudeReq.Model,
		"stream":     claudeReq.Stream,
		"max_tokens": claudeReq.MaxTokens,
		"thinking":   claudeReq.Thinking,
		"has_tools":  len(claudeReq.Tools) > 0,
		"has_system": claudeReq.System != nil,
	}).Info("Processing Claude Code compatible request")

	// Check for model commands in message content
	if len(claudeReq.Messages) > 0 {
		if content, ok := claudeReq.Messages[0].Content.(string); ok {
			if provider, model, hasCommand := h.modelRouter.ParseModelCommand(content); hasCommand {
				h.logger.WithFields(logrus.Fields{
					"request_id": requestID,
					"provider":   provider,
					"model":      model,
				}).Info("Detected model command")

				// Update request model based on command
				claudeReq.Model = model
			}
		}
	}

	// Apply intelligent model routing
	// Update model router with current configuration
	h.modelRouter.UpdateModelConfiguration(cfg.BigModel, cfg.SmallModel)

	routedModel := h.modelRouter.RouteModel(&claudeReq)
	if routedModel != claudeReq.Model {
		h.logger.WithFields(logrus.Fields{
			"request_id":     requestID,
			"original_model": claudeReq.Model,
			"routed_model":   routedModel,
		}).Info("Applied model routing")
		claudeReq.Model = routedModel
	}

	// Convert Claude request to OpenAI format
	openaiReq, err := converter.ConvertClaudeToOpenAI(&claudeReq, cfg.BigModel, cfg.SmallModel)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert request")
		c.JSON(http.StatusBadRequest, converter.CreateClaudeErrorResponse("invalid_request_error", "Failed to convert request"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(cfg.RequestTimeout)*time.Second)
	defer cancel()

	if claudeReq.Stream {
		h.handleClaudeCodeStreamingRequest(c, ctx, requestID, &claudeReq, openaiReq, openaiClient)
	} else {
		h.handleNonStreamingRequest(c, ctx, requestID, &claudeReq, openaiReq, openaiClient)
	}
}

// handleClaudeCodeStreamingRequest handles streaming requests with Claude Code compatibility
func (h *EnhancedMessagesHandler) handleClaudeCodeStreamingRequest(c *gin.Context, ctx context.Context, requestID string, claudeReq *models.ClaudeMessagesRequest, openaiReq *models.OpenAIChatCompletionRequest, openaiClient *client.OpenAIClient) {
	startTime := time.Now()

	// Initialize Claude Code compatible streaming
	streamCtx := h.streamingService.InitializeStreaming(c, requestID, claudeReq.Model)

	// Check for client disconnect before starting stream
	if h.streamingService.CheckClientDisconnect(c) {
		h.logger.WithField("request_id", requestID).Info("Client disconnected before stream start")
		return
	}

	// Get streaming response from OpenAI
	streamChan, err := openaiClient.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		duration := time.Since(startTime)
		h.logger.WithError(err).Error("Failed to create stream")

		// Handle streaming error with Claude Code format
		h.streamingService.HandleStreamingError(c, streamCtx, err)

		// Log failed request
		h.logStreamingRequest(requestID, claudeReq, openaiReq, nil, http.StatusInternalServerError, true, 0, 0, duration, err.Error(), startTime)
		return
	}

	// Start periodic ping in background
	pingCtx, pingCancel := context.WithCancel(ctx)
	defer pingCancel()
	go h.streamingService.SendPeriodicPing(c, pingCtx)

	var totalInputTokens, totalOutputTokens int
	hasError := false
	var streamError error

	// Process stream chunks with Claude Code compatibility
	for chunk := range streamChan {
		// Check for client disconnect
		if h.streamingService.CheckClientDisconnect(c) {
			h.logger.WithField("request_id", requestID).Info("Client disconnected during streaming")
			hasError = true
			streamError = fmt.Errorf("client disconnected")
			break
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			h.logger.WithField("request_id", requestID).Warn("Request context cancelled")
			hasError = true
			streamError = ctx.Err()
			break
		default:
		}

		if chunk.Error != nil {
			h.logger.WithError(chunk.Error).Error("Stream error")
			h.streamingService.HandleStreamingError(c, streamCtx, chunk.Error)
			hasError = true
			streamError = chunk.Error
			break
		}

		if chunk.Done {
			break
		}

		if chunk.Data != nil {
			// Process chunk data with Claude Code events
			h.processStreamChunk(c, streamCtx, chunk.Data, &totalInputTokens, &totalOutputTokens)
		}
	}

	// Update usage tokens
	h.streamingService.UpdateUsageTokens(streamCtx, totalInputTokens, totalOutputTokens)

	// Determine stop reason
	stopReason := "end_turn"
	if hasError {
		stopReason = "error"
	}

	// Finalize streaming with proper Claude Code events
	h.streamingService.FinalizeStreaming(c, streamCtx, stopReason)

	duration := time.Since(startTime)

	// Log streaming request result
	statusCode := http.StatusOK
	errorMessage := ""
	if hasError {
		statusCode = http.StatusInternalServerError
		if streamError != nil {
			errorMessage = streamError.Error()
		}
	}

	h.logStreamingRequest(requestID, claudeReq, openaiReq, h.streamingService.GetStreamingStats(streamCtx), statusCode, true, totalInputTokens, totalOutputTokens, duration, errorMessage, startTime)
}

// processStreamChunk processes individual stream chunks and converts them to Claude Code events
func (h *EnhancedMessagesHandler) processStreamChunk(c *gin.Context, streamCtx *claudecode.StreamingContext, data interface{}, inputTokens, outputTokens *int) {
	if dataMap, ok := data.(map[string]interface{}); ok {
		// Handle choices
		if choices, exists := dataMap["choices"]; exists {
			if choicesArray, ok := choices.([]interface{}); ok && len(choicesArray) > 0 {
				if choice, ok := choicesArray[0].(map[string]interface{}); ok {
					// Handle delta content
					if delta, exists := choice["delta"]; exists {
						if deltaMap, ok := delta.(map[string]interface{}); ok {
							// Handle text content
							if content, exists := deltaMap["content"]; exists {
								if contentStr, ok := content.(string); ok {
									h.streamingService.ProcessTextChunk(c, streamCtx, contentStr)
								}
							}

							// Handle tool calls
							if toolCalls, exists := deltaMap["tool_calls"]; exists {
								h.streamingService.ProcessToolCall(c, streamCtx, toolCalls)
							}
						}
					}
				}
			}
		}

		// Handle usage information
		if usage, exists := dataMap["usage"]; exists {
			if usageMap, ok := usage.(map[string]interface{}); ok {
				if prompt, exists := usageMap["prompt_tokens"]; exists {
					if promptInt, ok := prompt.(int); ok {
						*inputTokens = promptInt
					}
				}
				if completion, exists := usageMap["completion_tokens"]; exists {
					if completionInt, ok := completion.(int); ok {
						*outputTokens = completionInt
					}
				}
			}
		}
	}
}

// handleNonStreamingRequest handles non-streaming requests (unchanged from original)
func (h *EnhancedMessagesHandler) handleNonStreamingRequest(c *gin.Context, ctx context.Context, requestID string, claudeReq *models.ClaudeMessagesRequest, openaiReq *models.OpenAIChatCompletionRequest, openaiClient *client.OpenAIClient) {
	startTime := time.Now()

	// Send request to OpenAI
	openaiResp, err := openaiClient.CreateChatCompletion(ctx, openaiReq)
	duration := time.Since(startTime)

	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"duration":   duration,
			"error":      err.Error(),
		}).Error("OpenAI request failed")

		// Log failed request
		h.logNonStreamingRequest(requestID, claudeReq, openaiReq, nil, http.StatusInternalServerError, false, 0, 0, duration, err.Error(), startTime)

		errorType := openaiClient.ClassifyError(err)
		c.JSON(http.StatusInternalServerError, converter.CreateClaudeErrorResponse(errorType, err.Error()))
		return
	}

	// Convert OpenAI response to Claude format
	claudeResp, err := converter.ConvertOpenAIToClaudeResponse(openaiResp, claudeReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert response")

		// Log conversion failed request
		h.logNonStreamingRequest(requestID, claudeReq, openaiReq, openaiResp, http.StatusInternalServerError, false, 0, 0, duration, "Failed to convert response", startTime)

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

	// Log successful request
	h.logNonStreamingRequest(requestID, claudeReq, openaiReq, claudeResp, http.StatusOK, false, claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens, duration, "", startTime)

	c.JSON(http.StatusOK, claudeResp)
}

// logStreamingRequest logs streaming request data
func (h *EnhancedMessagesHandler) logStreamingRequest(requestID string, claudeReq *models.ClaudeMessagesRequest, openaiReq *models.OpenAIChatCompletionRequest, responseData interface{}, statusCode int, isStreaming bool, inputTokens, outputTokens int, duration time.Duration, errorMessage string, startTime time.Time) {
	if h.requestLogger != nil {
		var claudeModel string
		if claudeReq != nil {
			claudeModel = claudeReq.Model
		}

		logData := &logging.RequestLogData{
			ID:           requestID,
			ClaudeModel:  claudeModel,
			OpenAIModel:  openaiReq.Model,
			RequestBody:  claudeReq,
			ResponseBody: responseData,
			StatusCode:   statusCode,
			IsStreaming:  isStreaming,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
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

// logNonStreamingRequest logs non-streaming request data
func (h *EnhancedMessagesHandler) logNonStreamingRequest(requestID string, claudeReq *models.ClaudeMessagesRequest, openaiReq *models.OpenAIChatCompletionRequest, responseData interface{}, statusCode int, isStreaming bool, inputTokens, outputTokens int, duration time.Duration, errorMessage string, startTime time.Time) {
	if h.requestLogger != nil {
		var claudeModel string
		if claudeReq != nil {
			claudeModel = claudeReq.Model
		}

		logData := &logging.RequestLogData{
			ID:           requestID,
			ClaudeModel:  claudeModel,
			OpenAIModel:  openaiReq.Model,
			RequestBody:  claudeReq,
			ResponseBody: responseData,
			StatusCode:   statusCode,
			IsStreaming:  isStreaming,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
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

// CountTokens handles POST /v1/messages/count_tokens with enhanced estimation
func (h *EnhancedMessagesHandler) CountTokens(c *gin.Context) {
	var req models.ClaudeTokenCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, converter.CreateClaudeErrorResponse("invalid_request_error", "Invalid request format"))
		return
	}

	// Use model router for better token estimation
	claudeReq := &models.ClaudeMessagesRequest{
		Model:    req.Model,
		Messages: req.Messages,
		System:   req.System,
	}

	// Get estimated token count from model router
	estimatedTokens := h.modelRouter.EstimateTokenCount(claudeReq)

	h.logger.WithFields(logrus.Fields{
		"model":            req.Model,
		"estimated_tokens": estimatedTokens,
		"message_count":    len(req.Messages),
	}).Info("Token count estimation completed")

	c.JSON(http.StatusOK, gin.H{
		"input_tokens": estimatedTokens,
	})
}

// GetModelCapabilities returns model capabilities for Claude Code
func (h *EnhancedMessagesHandler) GetModelCapabilities(c *gin.Context) {
	capabilities := h.modelRouter.GetModelCapabilities()
	c.JSON(http.StatusOK, capabilities)
}

// InitializeClaudeCodeConfig initializes Claude Code configuration
func (h *EnhancedMessagesHandler) InitializeClaudeCodeConfig() error {
	if err := h.configService.InitializeConfig(); err != nil {
		h.logger.WithError(err).Error("Failed to initialize Claude Code configuration")
		return err
	}

	// Increment startup count
	if err := h.configService.IncrementStartupCount(); err != nil {
		h.logger.WithError(err).Warn("Failed to increment startup count")
	}

	return nil
}

// CreateChatCompletion handles POST /v1/chat/completions (OpenAI API compatible)
func (h *EnhancedMessagesHandler) CreateChatCompletion(c *gin.Context) {
	// Get current OpenAI client with fresh configuration
	openaiClient, cfg, err := h.getCurrentOpenAIClient()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get current OpenAI client")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{
				"message": "OpenAI API is not configured. Please configure OpenAI API key in settings.",
				"type":    "service_unavailable",
			},
		})
		return
	}

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

	// Use model mapping from configuration
	switch openaiReq.Model {
	case "claude-3-5-sonnet-20241022", "claude-3-haiku-20240307":
		openaiReq.Model = cfg.BigModel
	default:
		openaiReq.Model = cfg.SmallModel
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(cfg.RequestTimeout)*time.Second)
	defer cancel()

	if openaiReq.Stream {
		h.handleOpenAIStreamingRequest(c, ctx, requestID, &openaiReq, openaiClient)
	} else {
		h.handleOpenAINonStreamingRequest(c, ctx, requestID, &openaiReq, openaiClient)
	}
}

// handleOpenAINonStreamingRequest handles non-streaming OpenAI requests
func (h *EnhancedMessagesHandler) handleOpenAINonStreamingRequest(c *gin.Context, ctx context.Context, requestID string, openaiReq *models.OpenAIChatCompletionRequest, openaiClient *client.OpenAIClient) {
	startTime := time.Now()

	// Send request to OpenAI
	openaiResp, err := openaiClient.CreateChatCompletion(ctx, openaiReq)
	duration := time.Since(startTime)

	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"duration":   duration,
			"error":      err.Error(),
		}).Error("OpenAI request failed")

		// Log failed request
		h.logNonStreamingRequest(requestID, nil, openaiReq, nil, http.StatusInternalServerError, false, 0, 0, duration, err.Error(), startTime)

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

	// Log successful request
	h.logNonStreamingRequest(requestID, nil, openaiReq, openaiResp, http.StatusOK, false, 0, 0, duration, "", startTime)

	c.JSON(http.StatusOK, openaiResp)
}

// handleOpenAIStreamingRequest handles streaming OpenAI requests
func (h *EnhancedMessagesHandler) handleOpenAIStreamingRequest(c *gin.Context, ctx context.Context, requestID string, openaiReq *models.OpenAIChatCompletionRequest, openaiClient *client.OpenAIClient) {
	startTime := time.Now()

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "*")

	// Get streaming response from OpenAI
	streamChan, err := openaiClient.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		duration := time.Since(startTime)
		h.logger.WithError(err).Error("Failed to create OpenAI stream")

		// Log streaming request failure
		h.logStreamingRequest(requestID, nil, openaiReq, nil, http.StatusInternalServerError, true, 0, 0, duration, err.Error(), startTime)

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
			// Directly output OpenAI format streaming response
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

	// Log streaming request result
	statusCode := http.StatusOK
	errorMessage := ""

	if hasError {
		statusCode = http.StatusInternalServerError
		if streamError != nil {
			errorMessage = streamError.Error()
		}
	}

	h.logStreamingRequest(requestID, nil, openaiReq, gin.H{"streaming": true}, statusCode, true, totalInputTokens, totalOutputTokens, duration, errorMessage, startTime)
}

// EstimateTokenCount is a public method that wraps the model router's token estimation
func (h *EnhancedMessagesHandler) EstimateTokenCount(req *models.ClaudeMessagesRequest) int {
	return h.modelRouter.EstimateTokenCount(req)
}

// TestOpenAIModel tests a specific OpenAI model directly
func (h *EnhancedMessagesHandler) TestOpenAIModel(c *gin.Context) {
	// Get current configuration to create a fresh client
	cfg, err := h.configManager.GetConfig()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get current config for model test")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Failed to get current configuration",
				"type":    "config_error",
			},
		})
		return
	}

	// Check if OpenAI API key is configured
	if cfg.OpenAIAPIKey == "" {
		h.logger.Error("OpenAI API key not configured")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{
				"message": "OpenAI API is not configured. Please configure OpenAI API key in settings.",
				"type":    "service_unavailable",
			},
		})
		return
	}

	var testReq struct {
		Model string `json:"model" binding:"required"`
	}

	if err := c.ShouldBindJSON(&testReq); err != nil {
		h.logger.WithError(err).Error("Failed to bind test model request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Invalid request format",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// Create a fresh OpenAI client with current configuration
	openaiClient := client.NewOpenAIClient(
		cfg.OpenAIAPIKey,
		cfg.OpenAIBaseURL,
		cfg.RequestTimeout,
		h.logger,
	)

	requestID := uuid.New().String()
	ctx := context.Background()

	// Create a simple test request
	openaiReq := &models.OpenAIChatCompletionRequest{
		Model: testReq.Model,
		Messages: []models.Message{
			{
				Role:    "user",
				Content: "Hello, this is a test message. Please respond briefly with 'Test successful'.",
			},
		},
		MaxTokens:   &[]int{50}[0],
		Temperature: &[]float64{0.7}[0],
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"model":      testReq.Model,
		"test_type":  "model_test",
	}).Info("Testing OpenAI model with fresh client")

	startTime := time.Now()

	// Send request directly to OpenAI
	openaiResp, err := openaiClient.CreateChatCompletion(ctx, openaiReq)
	duration := time.Since(startTime)

	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"model":      testReq.Model,
			"duration":   duration,
			"error":      err.Error(),
		}).Error("OpenAI model test failed")

		errorType := openaiClient.ClassifyError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
			"type":    errorType,
			"model":   testReq.Model,
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":    requestID,
		"model":         testReq.Model,
		"duration":      duration,
		"input_tokens":  openaiResp.Usage.PromptTokens,
		"output_tokens": openaiResp.Usage.CompletionTokens,
		"response_id":   openaiResp.ID,
	}).Info("OpenAI model test successful")

	// Return successful test result
	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"message":       "Model test successful",
		"model":         testReq.Model,
		"duration":      duration.String(),
		"input_tokens":  openaiResp.Usage.PromptTokens,
		"output_tokens": openaiResp.Usage.CompletionTokens,
		"response_id":   openaiResp.ID,
		"response_text": openaiResp.Choices[0].Message.Content,
	})
}

package handlers

import (
	"context"
	"net/http"
	"time"

	"ccany/internal/app"
	"ccany/internal/client"
	"ccany/internal/config"
	"ccany/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	config        *config.Config
	configManager *app.ConfigManager
	logger        *logrus.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config, configManager *app.ConfigManager, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		config:        cfg,
		configManager: configManager,
		logger:        logger,
	}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	// Get current configuration for accurate status
	cfg, err := h.configManager.GetConfig()
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get current config for health check, using static config")
		cfg = h.config // fallback to static config
	}

	c.JSON(http.StatusOK, gin.H{
		"status":                "healthy",
		"timestamp":             time.Now().Format(time.RFC3339),
		"openai_api_configured": cfg.OpenAIAPIKey != "",
		"api_key_valid":         cfg.ValidateAPIKey(),
	})
}

// TestConnection handles GET /test-connection
func (h *HealthHandler) TestConnection(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get current configuration
	cfg, err := h.configManager.GetConfig()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get current config for connection test")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     "failed",
			"error_type": "Configuration Error",
			"message":    "Failed to get current configuration",
			"timestamp":  time.Now().Format(time.RFC3339),
			"suggestions": []string{
				"Check server configuration",
				"Verify database connectivity",
			},
		})
		return
	}

	// Check if OpenAI API key is configured
	if cfg.OpenAIAPIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     "failed",
			"error_type": "Configuration Error",
			"message":    "OpenAI API key not configured",
			"timestamp":  time.Now().Format(time.RFC3339),
			"suggestions": []string{
				"Configure your OpenAI API key in the settings",
				"Ensure the API key is valid and has the necessary permissions",
			},
		})
		return
	}

	// Create OpenAI client dynamically with current configuration
	openaiClient := client.NewOpenAIClient(
		cfg.OpenAIAPIKey,
		cfg.OpenAIBaseURL,
		cfg.RequestTimeout,
		h.logger,
	)

	// Create a simple test request
	testReq := &models.OpenAIChatCompletionRequest{
		Model: cfg.SmallModel,
		Messages: []models.OpenAIMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		MaxTokens: func() *int { i := 5; return &i }(),
	}

	startTime := time.Now()
	resp, err := openaiClient.CreateChatCompletion(ctx, testReq)
	duration := time.Since(startTime)

	if err != nil {
		h.logger.WithError(err).Error("API connectivity test failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":     "failed",
			"error_type": "API Error",
			"message":    err.Error(),
			"timestamp":  time.Now().Format(time.RFC3339),
			"duration":   duration.String(),
			"suggestions": []string{
				"Check your OPENAI_API_KEY is valid",
				"Verify your API key has the necessary permissions",
				"Check if you have reached rate limits",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Successfully connected to OpenAI API",
		"model_used":  cfg.SmallModel,
		"timestamp":   time.Now().Format(time.RFC3339),
		"duration":    duration.String(),
		"response_id": resp.ID,
	})
}

// Root handles GET /
func (h *HealthHandler) Root(c *gin.Context) {
	// Get current configuration for accurate status
	cfg, err := h.configManager.GetConfig()
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get current config for root endpoint, using static config")
		cfg = h.config // fallback to static config
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Claude-to-OpenAI API Proxy v1.0.0",
		"status":  "running",
		"config": gin.H{
			"openai_base_url":    cfg.OpenAIBaseURL,
			"max_tokens_limit":   cfg.MaxTokensLimit,
			"api_key_configured": cfg.OpenAIAPIKey != "",
			"big_model":          cfg.BigModel,
			"small_model":        cfg.SmallModel,
		},
		"endpoints": gin.H{
			"messages":        "/v1/messages",
			"count_tokens":    "/v1/messages/count_tokens",
			"health":          "/health",
			"test_connection": "/test-connection",
		},
	})
}

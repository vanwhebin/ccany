package handlers

import (
	"context"
	"net/http"
	"time"

	"ccany/internal/client"
	"ccany/internal/config"
	"ccany/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	config       *config.Config
	openaiClient *client.OpenAIClient
	logger       *logrus.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config, openaiClient *client.OpenAIClient, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		config:       cfg,
		openaiClient: openaiClient,
		logger:       logger,
	}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":                "healthy",
		"timestamp":             time.Now().Format(time.RFC3339),
		"openai_api_configured": h.config.OpenAIAPIKey != "",
		"api_key_valid":         h.config.ValidateAPIKey(),
	})
}

// TestConnection handles GET /test-connection
func (h *HealthHandler) TestConnection(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Create a simple test request
	testReq := &models.OpenAIChatCompletionRequest{
		Model: h.config.SmallModel,
		Messages: []models.OpenAIMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		MaxTokens: func() *int { i := 5; return &i }(),
	}

	startTime := time.Now()
	resp, err := h.openaiClient.CreateChatCompletion(ctx, testReq)
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
		"model_used":  h.config.SmallModel,
		"timestamp":   time.Now().Format(time.RFC3339),
		"duration":    duration.String(),
		"response_id": resp.ID,
	})
}

// Root handles GET /
func (h *HealthHandler) Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Claude-to-OpenAI API Proxy v1.0.0",
		"status":  "running",
		"config": gin.H{
			"openai_base_url":    h.config.OpenAIBaseURL,
			"max_tokens_limit":   h.config.MaxTokensLimit,
			"api_key_configured": h.config.OpenAIAPIKey != "",
			"big_model":          h.config.BigModel,
			"small_model":        h.config.SmallModel,
		},
		"endpoints": gin.H{
			"messages":        "/v1/messages",
			"count_tokens":    "/v1/messages/count_tokens",
			"health":          "/health",
			"test_connection": "/test-connection",
		},
	})
}

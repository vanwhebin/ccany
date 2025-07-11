package tests

import (
	"ccany/internal/claudecode"
	"ccany/internal/models"
	"github.com/sirupsen/logrus"
)

func TestClaudeCodeFixes(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise during tests

	t.Run("ConfigService", func(t *testing.T) {
		configService := claudecode.NewConfigService(logger)

		// Test config path
		if configService.GetConfigPath() == "" {
			t.Error("Config path should not be empty")
		}

		// Test config existence check
		exists := configService.ConfigExists()
		t.Logf("Config exists: %v", exists)
	})

	t.Run("ModelRouter", func(t *testing.T) {
		modelRouter := claudecode.NewModelRouter(logger, "gpt-4", "gpt-3.5-turbo")

		// Test thinking mode routing
		req := &models.ClaudeMessagesRequest{
			Model:    "claude-3-5-sonnet-20241022",
			Thinking: true,
			Messages: []models.ClaudeMessage{
				{Role: "user", Content: "Hello"},
			},
		}

		routedModel := modelRouter.RouteModel(req)
		t.Logf("Routed model for thinking mode: %s", routedModel)

		// Test token estimation
		tokenCount := modelRouter.EstimateTokenCount(req)
		if tokenCount <= 0 {
			t.Error("Token count should be positive")
		}
		t.Logf("Estimated token count: %d", tokenCount)

		// Test model command parsing
		provider, model, hasCommand := modelRouter.ParseModelCommand("/model openai,gpt-4")
		if !hasCommand {
			t.Error("Should detect model command")
		}
		if provider != "openai" || model != "gpt-4" {
			t.Errorf("Expected openai,gpt-4 but got %s,%s", provider, model)
		}
	})

	t.Run("StreamingService", func(t *testing.T) {
		streamingService := claudecode.NewStreamingService(logger)

		// Test streaming context creation
		ctx := &claudecode.StreamingContext{
			RequestID: "test-123",
			Model:     "claude-3-5-sonnet-20241022",
		}

		// Test stats generation
		stats := streamingService.GetStreamingStats(ctx)
		if stats["request_id"] != "test-123" {
			t.Error("Request ID should match")
		}

		t.Logf("Streaming stats: %+v", stats)
	})

	t.Run("EnhancedModels", func(t *testing.T) {
		// Test enhanced Claude request with thinking mode
		req := models.ClaudeMessagesRequest{
			Model:    "claude-3-5-sonnet-20241022",
			Thinking: true,
			Metadata: map[string]interface{}{
				"source": "claude-code",
			},
			Messages: []models.ClaudeMessage{
				{Role: "user", Content: "Test message"},
			},
		}

		if !req.Thinking {
			t.Error("Thinking mode should be enabled")
		}

		if req.Metadata == nil {
			t.Error("Metadata should not be nil")
		}

		// Test enhanced usage with cache tokens
		usage := models.ClaudeUsage{
			InputTokens:          100,
			OutputTokens:         50,
			CacheReadInputTokens: 25,
		}

		if usage.CacheReadInputTokens != 25 {
			t.Error("Cache read input tokens should be 25")
		}

		t.Logf("Enhanced usage: %+v", usage)
	})
}

func TestClaudeCodeIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("EndToEndFlow", func(t *testing.T) {
		// Test the complete flow of a Claude Code request
		modelRouter := claudecode.NewModelRouter(logger, "gpt-4", "gpt-3.5-turbo")

		// Create a complex request
		req := &models.ClaudeMessagesRequest{
			Model:     "claude-3-5-sonnet-20241022",
			Thinking:  true,
			MaxTokens: 4000,
			Messages: []models.ClaudeMessage{
				{Role: "user", Content: "Please help me write a complex algorithm"},
			},
			Tools: []models.ClaudeTool{
				{Name: "code_executor", Description: "Execute code"},
			},
		}

		// Test model routing
		routedModel := modelRouter.RouteModel(req)
		t.Logf("Original model: %s, Routed model: %s", req.Model, routedModel)

		// Test token estimation
		tokenCount := modelRouter.EstimateTokenCount(req)
		t.Logf("Estimated tokens: %d", tokenCount)

		// Test complexity detection
		if tokenCount <= 0 {
			t.Error("Token count should be positive for complex request")
		}

		// Test capabilities
		capabilities := modelRouter.GetModelCapabilities()
		if capabilities["models"] == nil {
			t.Error("Model capabilities should include models")
		}

		t.Logf("Model capabilities: %+v", capabilities)
	})
}

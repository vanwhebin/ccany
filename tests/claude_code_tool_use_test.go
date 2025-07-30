package tests

import (
	"ccany/internal/claudecode"
	"ccany/internal/models"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestToolUseRouting(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Test cases for tool use routing
	tests := []struct {
		name           string
		bigModel       string
		smallModel     string
		request        *models.ClaudeMessagesRequest
		expectedModel  string
		expectedReason string
	}{
		{
			name:       "Request with tools should use big model",
			bigModel:   "gpt-4o",
			smallModel: "gpt-4o-mini",
			request: &models.ClaudeMessagesRequest{
				Model: "claude-3-5-sonnet-20241022",
				Messages: []models.ClaudeMessage{
					{
						Role:    "user",
						Content: "Help me with a task",
					},
				},
				Tools: []models.ClaudeTool{
					{
						Name:        "read_file",
						Description: "Read file contents",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"path": map[string]interface{}{
									"type":        "string",
									"description": "File path",
								},
							},
							"required": []string{"path"},
						},
					},
				},
			},
			expectedModel:  "gpt-4o",
			expectedReason: "tool_use_detected",
		},
		{
			name:       "Request without tools should use default routing",
			bigModel:   "gpt-4o",
			smallModel: "gpt-4o-mini",
			request: &models.ClaudeMessagesRequest{
				Model: "claude-3-5-sonnet-20241022",
				Messages: []models.ClaudeMessage{
					{
						Role:    "user",
						Content: "Hello, how are you?",
					},
				},
			},
			expectedModel:  "gpt-4o", // Default is big model
			expectedReason: "default",
		},
		{
			name:       "Tool use routing should not apply to Kimi model",
			bigModel:   "moonshotai/Kimi-K2-Instruct",
			smallModel: "deepseek-ai/DeepSeek-V3",
			request: &models.ClaudeMessagesRequest{
				Model: "claude-3-5-sonnet-20241022",
				Messages: []models.ClaudeMessage{
					{
						Role:    "user",
						Content: "Help me with a task",
					},
				},
				Tools: []models.ClaudeTool{
					{
						Name:        "read_file",
						Description: "Read file contents",
					},
				},
			},
			expectedModel:  "moonshotai/Kimi-K2-Instruct", // Even with tools, still uses configured model
			expectedReason: "tool_use_detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with test models
			router := claudecode.NewModelRouter(logger, tt.bigModel, tt.smallModel)

			// Route the model
			routedModel := router.RouteModel(tt.request)

			// Assert the routed model
			assert.Equal(t, tt.expectedModel, routedModel, "Routed model should match expected")

			// Note: We can't easily test the reason without exposing internal state,
			// but we can verify the model is routed correctly
		})
	}
}

func TestToolUseStrategy(t *testing.T) {
	logger := logrus.New()

	tests := []struct {
		name        string
		model       string
		enabled     bool
		request     *models.ClaudeMessagesRequest
		shouldApply bool
	}{
		{
			name:    "Should apply when tools present and enabled",
			model:   "gpt-4o",
			enabled: true,
			request: &models.ClaudeMessagesRequest{
				Tools: []models.ClaudeTool{
					{Name: "test_tool"},
				},
			},
			shouldApply: true,
		},
		{
			name:    "Should not apply when no tools",
			model:   "gpt-4o",
			enabled: true,
			request: &models.ClaudeMessagesRequest{
				Tools: []models.ClaudeTool{},
			},
			shouldApply: false,
		},
		{
			name:    "Should not apply when disabled",
			model:   "gpt-4o",
			enabled: false,
			request: &models.ClaudeMessagesRequest{
				Tools: []models.ClaudeTool{
					{Name: "test_tool"},
				},
			},
			shouldApply: false,
		},
		{
			name:    "Should not apply when model empty",
			model:   "",
			enabled: true,
			request: &models.ClaudeMessagesRequest{
				Tools: []models.ClaudeTool{
					{Name: "test_tool"},
				},
			},
			shouldApply: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := claudecode.NewToolUseStrategy(tt.model, tt.enabled, logger)
			result := strategy.ShouldApply(tt.request, 0)
			assert.Equal(t, tt.shouldApply, result)

			if tt.shouldApply {
				assert.Equal(t, tt.model, strategy.GetModel())
				assert.Equal(t, "tool_use_detected", strategy.GetReason())
			}
		})
	}
}

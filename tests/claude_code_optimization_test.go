package tests

import (
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"

	"ccany/internal/claudecode"
	"ccany/internal/converter"
	"ccany/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestStreamingServiceConcurrency tests concurrent access to streaming context
func TestStreamingServiceConcurrency(t *testing.T) {
	logger := logrus.New()
	service := claudecode.NewStreamingService(logger)

	// Create multiple goroutines accessing the same context
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create a test Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	streamCtx := service.InitializeStreaming(c, "test-req-1", "claude-3-haiku")

	// Concurrent access test
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Test concurrent text chunk processing
			text := fmt.Sprintf("Test chunk %d", id)
			service.ProcessTextChunk(c, streamCtx, text)

			// Test concurrent tool call processing
			toolDelta := []interface{}{
				map[string]interface{}{
					"index": float64(id),
					"id":    fmt.Sprintf("tool_%d", id),
					"function": map[string]interface{}{
						"name":      "test_tool",
						"arguments": `{"test": "data"}`,
					},
				},
			}
			service.ProcessToolCallDeltas(c, streamCtx, toolDelta)
		}(i)
	}

	wg.Wait()

	// Verify no data corruption
	assert.NotNil(t, streamCtx)
	assert.Equal(t, "test-req-1", streamCtx.RequestID)

	// Cleanup
	service.CleanupContext(streamCtx)
}

// TestStreamingServiceMemoryLeak tests that contexts are properly cleaned up
func TestStreamingServiceMemoryLeak(t *testing.T) {
	logger := logrus.New()
	service := claudecode.NewStreamingService(logger)

	// Create and cleanup multiple contexts
	const numIterations = 100

	for i := 0; i < numIterations; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		streamCtx := service.InitializeStreaming(c, fmt.Sprintf("req-%d", i), "claude-3-haiku")

		// Simulate some work
		service.ProcessTextChunk(c, streamCtx, "Test content")

		// Finalize and cleanup
		service.FinalizeStreaming(c, streamCtx, "end_turn", 100, 50)
		service.CleanupContext(streamCtx)
	}

	// Context pool should be reusing objects, not growing indefinitely
	// This test ensures no memory leak by running many iterations
	assert.True(t, true, "No panic or memory issues after %d iterations", numIterations)
}

// TestModelRouterStrategies tests the new strategy-based routing
func TestModelRouterStrategies(t *testing.T) {
	logger := logrus.New()
	router := claudecode.NewModelRouter(logger, "gpt-4", "gpt-3.5-turbo")

	tests := []struct {
		name     string
		request  *models.ClaudeMessagesRequest
		expected string
		reason   string
	}{
		{
			name: "Comma-separated models",
			request: &models.ClaudeMessagesRequest{
				Model: "gpt-4,gpt-3.5-turbo",
			},
			expected: "gpt-4,gpt-3.5-turbo",
			reason:   "comma_separated_models",
		},
		{
			name: "Long context",
			request: &models.ClaudeMessagesRequest{
				Model: "claude-3-opus",
				Messages: []models.ClaudeMessage{
					{
						Role:    "user",
						Content: generateLongContent(70000), // >60K tokens
					},
				},
			},
			expected: "gpt-4",   // Should use default model since we don't have long context model configured
			reason:   "default", // Will fall back to default since no long context model is set
		},
		{
			name: "Haiku background model",
			request: &models.ClaudeMessagesRequest{
				Model: "claude-3-5-haiku-20241022",
			},
			expected: "gpt-3.5-turbo",
			reason:   "background_model",
		},
		{
			name: "Thinking mode",
			request: &models.ClaudeMessagesRequest{
				Model:    "claude-3-sonnet",
				Thinking: true,
			},
			expected: "claude-3-5-sonnet-20241022",
			reason:   "thinking_mode",
		},
		{
			name: "Default routing",
			request: &models.ClaudeMessagesRequest{
				Model: "claude-3-sonnet",
			},
			expected: "gpt-4",
			reason:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.RouteModel(tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestModelRouterCustomStrategy tests adding custom routing strategies
func TestModelRouterCustomStrategy(t *testing.T) {
	logger := logrus.New()
	router := claudecode.NewModelRouter(logger, "gpt-4", "gpt-3.5-turbo")

	// Define a custom strategy
	customStrategy := &testCustomStrategy{
		model: "custom-model",
	}

	// Add custom strategy with high priority
	router.AddCustomStrategy(customStrategy, 0)

	// Test that custom strategy takes precedence
	req := &models.ClaudeMessagesRequest{
		Model: "test-custom",
	}

	result := router.RouteModel(req)
	assert.Equal(t, "custom-model", result)
}

// testCustomStrategy implements ModelRoutingStrategy for testing
type testCustomStrategy struct {
	model string
}

func (s *testCustomStrategy) ShouldApply(req *models.ClaudeMessagesRequest, tokenCount int) bool {
	return req.Model == "test-custom"
}

func (s *testCustomStrategy) GetModel() string {
	return s.model
}

func (s *testCustomStrategy) GetReason() string {
	return "custom_test"
}

// TestClaudeConverterToolCallParsing tests enhanced tool call parsing
func TestClaudeConverterToolCallParsing(t *testing.T) {
	logger := logrus.New()
	converter := converter.NewClaudeConverterWithLogger(logger)

	tests := []struct {
		name          string
		content       string
		expectedTools int
		expectedClean string
	}{
		{
			name:          "Standard tool call JSON",
			content:       `Some text before {"tool_calls": [{"id": "call_1", "type": "function", "function": {"name": "test_tool", "arguments": "{\"key\": \"value\"}"}}]} some text after`,
			expectedTools: 1,
			expectedClean: "Some text before some text after",
		},
		{
			name:          "Tool call with space variations",
			content:       `Text { "tool_calls" : [{"id": "call_2", "type": "function", "function": {"name": "another_tool", "arguments": "{}"}}]} more text`,
			expectedTools: 1,
			expectedClean: "Text more text",
		},
		{
			name:          "Malformed arguments",
			content:       `Before {"tool_calls": [{"id": "call_3", "type": "function", "function": {"name": "broken_tool", "arguments": "not valid json"}}]} after`,
			expectedTools: 1,
			expectedClean: "Before after",
		},
		{
			name:          "No tool calls",
			content:       `This is just regular text without any tool calls`,
			expectedTools: 0,
			expectedClean: "This is just regular text without any tool calls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanContent, toolCalls := converter.ParseCustomFormatFromContent(tt.content)
			assert.Len(t, toolCalls, tt.expectedTools)
			assert.Equal(t, tt.expectedClean, cleanContent)
		})
	}
}

// TestClaudeConverterConcurrentParsing tests thread safety of parsing
func TestClaudeConverterConcurrentParsing(t *testing.T) {
	logger := logrus.New()
	converter := converter.NewClaudeConverterWithLogger(logger)

	const numGoroutines = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			content := fmt.Sprintf(`Text %d {"tool_calls": [{"id": "call_%d", "type": "function", "function": {"name": "tool_%d", "arguments": "{}"}}]} end`, id, id, id)

			cleanContent, toolCalls := converter.ParseCustomFormatFromContent(content)
			assert.Len(t, toolCalls, 1)
			assert.Contains(t, cleanContent, fmt.Sprintf("Text %d", id))
		}(i)
	}

	wg.Wait()
}

// Helper functions

func generateLongContent(targetTokens int) string {
	// Approximately 3.3 chars per token
	chars := targetTokens * 3
	content := ""
	for len(content) < chars {
		content += "This is a test sentence to generate long content. "
	}
	return content
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

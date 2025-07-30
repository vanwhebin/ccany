package tokenizer

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTokenCounter_CountTokens(t *testing.T) {
	logger := logrus.New()
	tc := NewTokenCounter(logger)
	defer tc.Close()

	tests := []struct {
		name     string
		text     string
		model    string
		minCount int
		maxCount int
	}{
		{
			name:     "Simple English text",
			text:     "Hello, world! This is a test.",
			model:    "gpt-4",
			minCount: 5,
			maxCount: 10,
		},
		{
			name:     "Chinese text",
			text:     "你好，世界！这是一个测试。",
			model:    "claude-3-5-sonnet-20241022",
			minCount: 8,
			maxCount: 15,
		},
		{
			name:     "Code snippet",
			text:     "function hello() { return 'Hello, world!'; }",
			model:    "gpt-3.5-turbo",
			minCount: 10,
			maxCount: 15,
		},
		{
			name:     "Empty text",
			text:     "",
			model:    "gpt-4",
			minCount: 0,
			maxCount: 0,
		},
		{
			name:     "Long text",
			text:     strings.Repeat("The quick brown fox jumps over the lazy dog. ", 10),
			model:    "gpt-4",
			minCount: 80,
			maxCount: 120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tc.CountTokens(tt.text, tt.model)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, count, tt.minCount, "Token count should be at least %d, got %d", tt.minCount, count)
			assert.LessOrEqual(t, count, tt.maxCount, "Token count should be at most %d, got %d", tt.maxCount, count)
			t.Logf("Text: %q, Model: %s, Token count: %d", tt.text, tt.model, count)
		})
	}
}

func TestTokenCounter_CountMessagesTokens(t *testing.T) {
	logger := logrus.New()
	tc := NewTokenCounter(logger)
	defer tc.Close()

	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": "You are a helpful assistant.",
		},
		{
			"role":    "user",
			"content": "Hello!",
		},
		{
			"role":    "assistant",
			"content": "Hi there! How can I help you today?",
		},
	}

	count, err := tc.CountMessagesTokens(messages, "gpt-4")
	assert.NoError(t, err)
	assert.Greater(t, count, 20, "Message tokens should be greater than 20")
	assert.Less(t, count, 50, "Message tokens should be less than 50")
	t.Logf("Messages token count: %d", count)
}

func TestTokenCounter_CountClaudeMessagesTokens(t *testing.T) {
	logger := logrus.New()
	tc := NewTokenCounter(logger)
	defer tc.Close()

	system := "You are Claude, a helpful AI assistant."
	messages := []interface{}{
		map[string]interface{}{
			"role":    "user",
			"content": "What's the weather like?",
		},
		map[string]interface{}{
			"role": "assistant",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "I don't have access to real-time weather data.",
				},
			},
		},
	}

	count, err := tc.CountClaudeMessagesTokens(messages, system, "claude-3-5-sonnet-20241022")
	assert.NoError(t, err)
	assert.Greater(t, count, 20, "Claude messages tokens should be greater than 20")
	t.Logf("Claude messages token count: %d", count)
}

func TestTokenCounter_EstimateTokensFromLength(t *testing.T) {
	logger := logrus.New()
	tc := NewTokenCounter(logger)
	defer tc.Close()

	tests := []struct {
		name     string
		text     string
		minRatio float64 // min chars per token
		maxRatio float64 // max chars per token
	}{
		{
			name:     "English text",
			text:     "The quick brown fox jumps over the lazy dog.",
			minRatio: 3.0,
			maxRatio: 5.0,
		},
		{
			name:     "Code snippet",
			text:     "function test() { console.log('Hello'); }",
			minRatio: 2.5,
			maxRatio: 4.0,
		},
		{
			name:     "Chinese text",
			text:     "这是一段中文测试文本。",
			minRatio: 1.5,
			maxRatio: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tc.EstimateTokensFromLength(tt.text)
			charCount := len([]rune(tt.text))
			ratio := float64(charCount) / float64(tokens)

			assert.GreaterOrEqual(t, ratio, tt.minRatio, "Chars per token ratio should be at least %.1f, got %.1f", tt.minRatio, ratio)
			assert.LessOrEqual(t, ratio, tt.maxRatio, "Chars per token ratio should be at most %.1f, got %.1f", tt.maxRatio, ratio)
			t.Logf("Text: %q, Tokens: %d, Ratio: %.2f chars/token", tt.text, tokens, ratio)
		})
	}
}

func TestTokenCounter_ContentWithImages(t *testing.T) {
	logger := logrus.New()
	tc := NewTokenCounter(logger)
	defer tc.Close()

	content := []interface{}{
		map[string]interface{}{
			"type": "text",
			"text": "Here's an image:",
		},
		map[string]interface{}{
			"type": "image",
			"source": map[string]interface{}{
				"media_type": "image/jpeg",
				"data":       "base64data...",
			},
		},
		map[string]interface{}{
			"type": "text",
			"text": "What do you see?",
		},
	}

	tokens, err := tc.CountContentTokens(content, "gpt-4-vision-preview")
	assert.NoError(t, err)
	assert.Greater(t, tokens, 100, "Content with image should have more than 100 tokens due to image cost")
	t.Logf("Content with image token count: %d", tokens)
}

func TestTokenCounter_ToolUseContent(t *testing.T) {
	logger := logrus.New()
	tc := NewTokenCounter(logger)
	defer tc.Close()

	content := []interface{}{
		map[string]interface{}{
			"type": "tool_use",
			"id":   "tool_1",
			"name": "get_weather",
			"input": map[string]interface{}{
				"location": "San Francisco",
				"unit":     "celsius",
			},
		},
		map[string]interface{}{
			"type":        "tool_result",
			"tool_use_id": "tool_1",
			"content":     "Temperature: 20°C, Sunny",
		},
	}

	tokens, err := tc.CountContentTokens(content, "claude-3-5-sonnet-20241022")
	assert.NoError(t, err)
	assert.Greater(t, tokens, 50, "Tool use content should have substantial tokens")
	t.Logf("Tool use content token count: %d", tokens)
}

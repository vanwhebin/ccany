package converter

import (
	"strings"

	"ccany/internal/models"
)

// Model mapping functions
func mapClaudeModelToOpenAI(claudeModel, bigModel, smallModel string) string {
	// Handle comma-separated models (take the first one for OpenAI)
	if strings.Contains(claudeModel, ",") {
		parts := strings.Split(claudeModel, ",")
		if len(parts) > 0 {
			claudeModel = strings.TrimSpace(parts[0])
		}
	}

	claudeModelLower := strings.ToLower(claudeModel)

	// Check for haiku models (small/background)
	if strings.Contains(claudeModelLower, "haiku") {
		return smallModel
	}

	// Check for sonnet or opus models (big)
	if strings.Contains(claudeModelLower, "sonnet") || strings.Contains(claudeModelLower, "opus") {
		return bigModel
	}

	// Check for specific provider models (e.g., anthropic/claude-sonnet-4)
	if strings.Contains(claudeModelLower, "anthropic/") {
		// Extract model name after provider
		parts := strings.Split(claudeModelLower, "/")
		if len(parts) > 1 {
			modelName := parts[1]
			if strings.Contains(modelName, "haiku") {
				return smallModel
			}
			if strings.Contains(modelName, "sonnet") || strings.Contains(modelName, "opus") {
				return bigModel
			}
		}
	}

	// Default to big model for unknown Claude models
	return bigModel
}

// Role mapping functions
func mapOpenAIRoleToGemini(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "model"
	default:
		return "user"
	}
}

func mapGeminiRoleToOpenAI(role string) string {
	switch role {
	case "user":
		return "user"
	case "model":
		return "assistant"
	default:
		return "user"
	}
}

func mapClaudeRoleToGemini(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "model"
	default:
		return "user"
	}
}

// ConvertClaudeToOpenAI converts a Claude request to OpenAI format (deprecated, use OpenAIConverter)
func ConvertClaudeToOpenAI(claudeReq *models.ClaudeMessagesRequest, bigModel, smallModel string) (*models.OpenAIChatCompletionRequest, error) {
	converter := NewOpenAIConverter()
	return converter.ConvertFromClaude(claudeReq, bigModel, smallModel)
}

// ConvertClaudeToOpenAIWithConverter converts a Claude request to OpenAI format using the new converter
func ConvertClaudeToOpenAIWithConverter(claudeReq *models.ClaudeMessagesRequest, bigModel, smallModel string) (*models.OpenAIChatCompletionRequest, error) {
	converter := NewOpenAIConverter()
	return converter.ConvertFromClaude(claudeReq, bigModel, smallModel)
}

// ConvertOpenAIToGemini converts OpenAI request to Gemini format
func ConvertOpenAIToGemini(openaiReq *models.OpenAIChatCompletionRequest) (*models.GeminiRequest, error) {
	converter := NewGeminiConverter()
	return converter.ConvertFromOpenAI(openaiReq)
}

// ConvertGeminiToOpenAI converts Gemini request to OpenAI format
func ConvertGeminiToOpenAI(geminiReq *models.GeminiRequest) (*models.OpenAIChatCompletionRequest, error) {
	converter := NewGeminiConverter()
	return converter.ConvertToOpenAI(geminiReq)
}

// ConvertClaudeToGemini converts Claude request to Gemini format
func ConvertClaudeToGemini(claudeReq *models.ClaudeMessagesRequest) (*models.GeminiRequest, error) {
	converter := NewGeminiConverter()
	return converter.ConvertFromClaude(claudeReq)
}

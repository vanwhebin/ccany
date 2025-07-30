package converter

import (
	"fmt"
	"sync"

	"ccany/internal/models"
)

// ConverterRegistry provides a centralized way to access all converters
type ConverterRegistry struct {
	openaiConverter *OpenAIConverter
	claudeConverter *ClaudeConverter
	geminiConverter *GeminiConverter
	mu              sync.RWMutex
}

// NewConverterRegistry creates a new converter registry
func NewConverterRegistry() *ConverterRegistry {
	return &ConverterRegistry{
		openaiConverter: NewOpenAIConverter(),
		claudeConverter: NewClaudeConverter(),
		geminiConverter: NewGeminiConverter(),
	}
}

// GetOpenAIConverter returns the OpenAI converter
func (cr *ConverterRegistry) GetOpenAIConverter() *OpenAIConverter {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.openaiConverter
}

// GetClaudeConverter returns the Claude converter
func (cr *ConverterRegistry) GetClaudeConverter() *ClaudeConverter {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.claudeConverter
}

// GetGeminiConverter returns the Gemini converter
func (cr *ConverterRegistry) GetGeminiConverter() *GeminiConverter {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.geminiConverter
}

// ConvertRequest provides unified request conversion
func (cr *ConverterRegistry) ConvertRequest(sourceFormat, targetFormat string, data interface{}) (interface{}, error) {
	switch sourceFormat {
	case "claude", "anthropic":
		claudeReq, ok := data.(*models.ClaudeMessagesRequest)
		if !ok {
			return nil, fmt.Errorf("invalid Claude request format")
		}

		switch targetFormat {
		case "openai":
			return cr.openaiConverter.ConvertFromClaude(claudeReq, "gpt-4", "gpt-3.5-turbo")
		case "gemini":
			return cr.geminiConverter.ConvertFromClaude(claudeReq)
		default:
			return nil, fmt.Errorf("unsupported target format: %s", targetFormat)
		}

	case "openai":
		openaiReq, ok := data.(*models.OpenAIChatCompletionRequest)
		if !ok {
			return nil, fmt.Errorf("invalid OpenAI request format")
		}

		switch targetFormat {
		case "gemini":
			return cr.geminiConverter.ConvertFromOpenAI(openaiReq)
		default:
			return nil, fmt.Errorf("unsupported target format: %s", targetFormat)
		}

	case "gemini":
		geminiReq, ok := data.(*models.GeminiRequest)
		if !ok {
			return nil, fmt.Errorf("invalid Gemini request format")
		}

		switch targetFormat {
		case "openai":
			return cr.geminiConverter.ConvertToOpenAI(geminiReq)
		default:
			return nil, fmt.Errorf("unsupported target format: %s", targetFormat)
		}

	default:
		return nil, fmt.Errorf("unsupported source format: %s", sourceFormat)
	}
}

// ConvertResponse provides unified response conversion
func (cr *ConverterRegistry) ConvertResponse(sourceFormat, targetFormat string, data interface{}, originalReq interface{}) (interface{}, error) {
	switch sourceFormat {
	case "openai":
		openaiResp, ok := data.(*models.OpenAIChatCompletionResponse)
		if !ok {
			return nil, fmt.Errorf("invalid OpenAI response format")
		}

		switch targetFormat {
		case "claude", "anthropic":
			claudeReq, ok := originalReq.(*models.ClaudeMessagesRequest)
			if !ok {
				return nil, fmt.Errorf("invalid original Claude request")
			}
			return cr.claudeConverter.ConvertFromOpenAI(openaiResp, claudeReq)
		default:
			return nil, fmt.Errorf("unsupported target format: %s", targetFormat)
		}

	case "gemini":
		geminiResp, ok := data.(*models.GeminiResponse)
		if !ok {
			return nil, fmt.Errorf("invalid Gemini response format")
		}

		switch targetFormat {
		case "claude", "anthropic":
			claudeReq, ok := originalReq.(*models.ClaudeMessagesRequest)
			if !ok {
				return nil, fmt.Errorf("invalid original Claude request")
			}
			return cr.geminiConverter.ConvertToClaude(geminiResp, claudeReq)
		default:
			return nil, fmt.Errorf("unsupported target format: %s", targetFormat)
		}

	default:
		return nil, fmt.Errorf("unsupported source format: %s", sourceFormat)
	}
}

// GetSupportedConversions returns a map of supported conversions
func (cr *ConverterRegistry) GetSupportedConversions() map[string][]string {
	return map[string][]string{
		"claude":    {"openai", "gemini"},
		"anthropic": {"openai", "gemini"},
		"openai":    {"claude", "anthropic", "gemini"},
		"gemini":    {"openai", "claude", "anthropic"},
	}
}

// Global converter registry instance
var DefaultConverterRegistry = NewConverterRegistry()

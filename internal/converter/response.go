package converter

import (
	"ccany/internal/models"
)

// StreamingContext holds streaming state for proper Claude format conversion
type StreamingContext struct {
	MessageID       string
	Model           string
	InputTokens     int
	OutputTokens    int
	ContentStarted  bool
	ToolCallStarted bool
	CurrentToolCall map[string]interface{}
	ContentBuffer   string
}

// ConvertOpenAIToClaudeResponse converts OpenAI response to Claude format (deprecated, use ClaudeConverter)
func ConvertOpenAIToClaudeResponse(openaiResp *models.OpenAIChatCompletionResponse, originalReq *models.ClaudeMessagesRequest) (*models.ClaudeResponse, error) {
	converter := NewClaudeConverter()
	return converter.ConvertFromOpenAI(openaiResp, originalReq)
}

// ConvertOpenAIStreamToClaudeStream converts OpenAI streaming response to Claude format (deprecated, use ClaudeConverter)
func ConvertOpenAIStreamToClaudeStream(openaiChunk *models.OpenAIStreamResponse, originalReq *models.ClaudeMessagesRequest, ctx *StreamingContext) ([]models.ClaudeStreamEvent, error) {
	converter := NewClaudeConverter()
	return converter.ConvertStreamFromOpenAI(openaiChunk, originalReq, ctx)
}

// Legacy support functions (deprecated - use ClaudeConverter methods instead)

// CreateClaudeErrorResponse creates a Claude-formatted error response
func CreateClaudeErrorResponse(errorType, message string) *models.ClaudeErrorResponse {
	converter := NewClaudeConverter()
	return converter.CreateErrorResponse(errorType, message)
}

// CreateClaudeStreamStartEvent creates a stream start event
func CreateClaudeStreamStartEvent(messageID, model string) models.ClaudeStreamEvent {
	converter := NewClaudeConverter()
	return converter.CreateStreamStartEvent(messageID, model)
}

// CreateClaudeStreamPingEvent creates a ping event for keep-alive
func CreateClaudeStreamPingEvent() models.ClaudeStreamEvent {
	converter := NewClaudeConverter()
	return converter.CreateStreamPingEvent()
}

// CreateStreamingContext creates a new streaming context
func CreateStreamingContext(messageID, model string, inputTokens int) *StreamingContext {
	converter := NewClaudeConverter()
	return converter.CreateStreamingContext(messageID, model, inputTokens)
}

// CreateClaudeStreamStopEvent creates a stream stop event
func CreateClaudeStreamStopEvent(usage models.ClaudeUsage) models.ClaudeStreamEvent {
	converter := NewClaudeConverter()
	return converter.CreateStreamStopEvent(usage)
}

// ConvertGeminiToClaudeResponse converts Gemini response to Claude format
func ConvertGeminiToClaudeResponse(geminiResp *models.GeminiResponse, originalReq *models.ClaudeMessagesRequest) (*models.ClaudeResponse, error) {
	converter := NewGeminiConverter()
	return converter.ConvertToClaude(geminiResp, originalReq)
}

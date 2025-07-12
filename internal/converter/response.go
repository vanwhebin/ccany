package converter

import (
	"fmt"

	"ccany/internal/models"
)

// ConvertOpenAIToClaudeResponse converts OpenAI response to Claude format
func ConvertOpenAIToClaudeResponse(openaiResp *models.OpenAIChatCompletionResponse, originalReq *models.ClaudeMessagesRequest) (*models.ClaudeResponse, error) {
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice := openaiResp.Choices[0]

	// Convert content
	content, err := convertMessageToClaudeContent(choice.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message content: %w", err)
	}

	// Map finish reason
	stopReason := mapFinishReasonToClaudeStopReason(choice.FinishReason)

	claudeResp := &models.ClaudeResponse{
		ID:         openaiResp.ID,
		Type:       "message",
		Role:       "assistant",
		Content:    content,
		Model:      originalReq.Model, // Use original Claude model name
		StopReason: stopReason,
		Usage: models.ClaudeUsage{
			InputTokens:  openaiResp.Usage.PromptTokens,
			OutputTokens: openaiResp.Usage.CompletionTokens,
		},
	}

	return claudeResp, nil
}

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

// ConvertOpenAIStreamToClaudeStream converts OpenAI streaming response to Claude format
func ConvertOpenAIStreamToClaudeStream(openaiChunk *models.OpenAIStreamResponse, originalReq *models.ClaudeMessagesRequest, ctx *StreamingContext) ([]models.ClaudeStreamEvent, error) {
	var events []models.ClaudeStreamEvent

	if len(openaiChunk.Choices) == 0 {
		return events, nil
	}

	choice := openaiChunk.Choices[0]

	// Handle content block start if needed
	if choice.Delta.Content != "" && !ctx.ContentStarted {
		events = append(events, models.ClaudeStreamEvent{
			Type:  "content_block_start",
			Index: 0,
			ContentBlock: &models.ClaudeContentBlock{
				Type: "text",
				Text: "",
			},
		})
		ctx.ContentStarted = true
	}

	// Handle text content delta
	if choice.Delta.Content != "" {
		ctx.ContentBuffer += choice.Delta.Content
		events = append(events, models.ClaudeStreamEvent{
			Type:  "content_block_delta",
			Index: 0,
			Delta: &models.ClaudeContentBlock{
				Type: "text_delta",
				Text: choice.Delta.Content,
			},
		})
	}

	// Handle tool calls (if present) - Note: basic StreamDelta doesn't support tool calls
	// This would need to be implemented if streaming tool calls are required

	// Handle finish reason
	if choice.FinishReason != "" {
		// Send content block stop if content was started
		if ctx.ContentStarted {
			events = append(events, models.ClaudeStreamEvent{
				Type:  "content_block_stop",
				Index: 0,
			})
		}

		stopReason := mapFinishReasonToClaudeStopReason(choice.FinishReason)

		// Send message delta with stop reason and usage
		events = append(events, models.ClaudeStreamEvent{
			Type: "message_delta",
			Delta: &models.ClaudeContentBlock{
				Type: "stop_reason",
				Text: stopReason,
			},
			Usage: &models.ClaudeUsage{
				InputTokens:  ctx.InputTokens,
				OutputTokens: ctx.OutputTokens,
			},
		})

		// Send message stop
		events = append(events, models.ClaudeStreamEvent{
			Type: "message_stop",
		})
	}

	return events, nil
}

// convertMessageToClaudeContent converts simple Message to Claude content blocks
func convertMessageToClaudeContent(msg models.Message) ([]models.ClaudeContentBlock, error) {
	var content []models.ClaudeContentBlock

	// Handle text content
	if msg.Content != "" {
		content = append(content, models.ClaudeContentBlock{
			Type: "text",
			Text: msg.Content,
		})
	}

	// If no content, add empty text block
	if len(content) == 0 {
		content = append(content, models.ClaudeContentBlock{
			Type: "text",
			Text: "",
		})
	}

	return content, nil
}

// mapFinishReasonToClaudeStopReason maps finish reasons to Claude stop reasons
func mapFinishReasonToClaudeStopReason(finishReason string) string {
	if finishReason == "" {
		return "end_turn"
	}

	switch finishReason {
	case "stop":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "tool_calls":
		return "tool_use"
	case "content_filter":
		return "stop_sequence"
	default:
		return "end_turn"
	}
}

// CreateClaudeErrorResponse creates a Claude-formatted error response
func CreateClaudeErrorResponse(errorType, message string) *models.ClaudeErrorResponse {
	return &models.ClaudeErrorResponse{
		Type: "error",
		Error: models.ClaudeError{
			Type:    errorType,
			Message: message,
		},
	}
}

// CreateClaudeStreamStartEvent creates a stream start event
func CreateClaudeStreamStartEvent(messageID, model string) models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type: "message_start",
		Message: &models.ClaudeResponse{
			ID:           messageID,
			Type:         "message",
			Role:         "assistant",
			Model:        model,
			Content:      []models.ClaudeContentBlock{},
			StopReason:   "",
			StopSequence: nil,
			Usage: models.ClaudeUsage{
				InputTokens:  0,
				OutputTokens: 0,
			},
		},
	}
}

// CreateClaudeStreamPingEvent creates a ping event for keep-alive
func CreateClaudeStreamPingEvent() models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type: "ping",
	}
}

// CreateStreamingContext creates a new streaming context
func CreateStreamingContext(messageID, model string, inputTokens int) *StreamingContext {
	return &StreamingContext{
		MessageID:       messageID,
		Model:           model,
		InputTokens:     inputTokens,
		OutputTokens:    0,
		ContentStarted:  false,
		ToolCallStarted: false,
		CurrentToolCall: make(map[string]interface{}),
		ContentBuffer:   "",
	}
}

// CreateClaudeStreamStopEvent creates a stream stop event
func CreateClaudeStreamStopEvent(usage models.ClaudeUsage) models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type:  "message_delta",
		Usage: &usage,
	}
}

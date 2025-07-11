package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"ccany/internal/models"
)

// ConvertOpenAIToClaudeResponse converts OpenAI response to Claude format
func ConvertOpenAIToClaudeResponse(openaiResp *models.OpenAIChatCompletionResponse, originalReq *models.ClaudeMessagesRequest) (*models.ClaudeResponse, error) {
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice := openaiResp.Choices[0]

	// Convert content
	content, err := convertOpenAIMessageToClaudeContent(choice.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message content: %w", err)
	}

	// Map finish reason
	stopReason := mapOpenAIFinishReasonToClaudeStopReason(choice.FinishReason)

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

// ConvertOpenAIStreamToClaudeStream converts OpenAI streaming response to Claude format
func ConvertOpenAIStreamToClaudeStream(openaiChunk *models.OpenAIStreamResponse, originalReq *models.ClaudeMessagesRequest) ([]models.ClaudeStreamEvent, error) {
	var events []models.ClaudeStreamEvent

	if len(openaiChunk.Choices) == 0 {
		return events, nil
	}

	choice := openaiChunk.Choices[0]

	// Handle different types of streaming events
	if choice.Delta != nil {
		if choice.Delta.Content != nil {
			// Text content delta
			if contentStr, ok := choice.Delta.Content.(string); ok && contentStr != "" {
				events = append(events, models.ClaudeStreamEvent{
					Type:  "content_block_delta",
					Index: 0,
					Delta: &models.ClaudeContentBlock{
						Type: "text_delta",
						Text: contentStr,
					},
				})
			}
		}

		// Handle tool calls in streaming
		if len(choice.Delta.ToolCalls) > 0 {
			for _, toolCall := range choice.Delta.ToolCalls {
				if toolCall.Function.Name != "" {
					// Tool use start
					events = append(events, models.ClaudeStreamEvent{
						Type:  "content_block_start",
						Index: 0,
						Delta: &models.ClaudeContentBlock{
							Type: "tool_use",
							ID:   strings.TrimPrefix(toolCall.ID, "call_"),
							Name: toolCall.Function.Name,
						},
					})
				}

				if toolCall.Function.Arguments != "" {
					// Tool arguments delta
					var input interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err == nil {
						events = append(events, models.ClaudeStreamEvent{
							Type:  "content_block_delta",
							Index: 0,
							Delta: &models.ClaudeContentBlock{
								Type:  "input_json_delta",
								Input: input,
							},
						})
					}
				}
			}
		}
	}

	// Handle finish reason
	if choice.FinishReason != nil {
		stopReason := mapOpenAIFinishReasonToClaudeStopReason(choice.FinishReason)

		events = append(events, models.ClaudeStreamEvent{
			Type: "message_delta",
			Delta: &models.ClaudeContentBlock{
				Type: "stop_reason",
				Text: stopReason,
			},
		})
	}

	return events, nil
}

// convertOpenAIMessageToClaudeContent converts OpenAI message to Claude content blocks
func convertOpenAIMessageToClaudeContent(msg models.OpenAIMessage) ([]models.ClaudeContentBlock, error) {
	var content []models.ClaudeContentBlock

	// Handle regular text content
	if msg.Content != nil {
		switch v := msg.Content.(type) {
		case string:
			if v != "" {
				content = append(content, models.ClaudeContentBlock{
					Type: "text",
					Text: v,
				})
			}
		case []interface{}:
			// Handle multi-part content
			for _, part := range v {
				if partMap, ok := part.(map[string]interface{}); ok {
					if partType, exists := partMap["type"]; exists {
						switch partType {
						case "text":
							if text, exists := partMap["text"]; exists {
								content = append(content, models.ClaudeContentBlock{
									Type: "text",
									Text: fmt.Sprintf("%v", text),
								})
							}
						}
					}
				}
			}
		}
	}

	// Handle tool calls
	if len(msg.ToolCalls) > 0 {
		for _, toolCall := range msg.ToolCalls {
			var input interface{}
			if toolCall.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					// Log error but continue processing
					input = toolCall.Function.Arguments
				}
			}

			content = append(content, models.ClaudeContentBlock{
				Type:  "tool_use",
				ID:    strings.TrimPrefix(toolCall.ID, "call_"),
				Name:  toolCall.Function.Name,
				Input: input,
			})
		}
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

// mapOpenAIFinishReasonToClaudeStopReason maps OpenAI finish reasons to Claude stop reasons
func mapOpenAIFinishReasonToClaudeStopReason(finishReason *string) string {
	if finishReason == nil {
		return "end_turn"
	}

	switch *finishReason {
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
			ID:    messageID,
			Type:  "message",
			Role:  "assistant",
			Model: model,
			Usage: models.ClaudeUsage{
				InputTokens:  0,
				OutputTokens: 0,
			},
		},
	}
}

// CreateClaudeStreamStopEvent creates a stream stop event
func CreateClaudeStreamStopEvent(usage models.ClaudeUsage) models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type:  "message_delta",
		Usage: &usage,
	}
}

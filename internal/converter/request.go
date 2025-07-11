package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"ccany/internal/models"
)

// ConvertClaudeToOpenAI converts a Claude request to OpenAI format
func ConvertClaudeToOpenAI(claudeReq *models.ClaudeMessagesRequest, bigModel, smallModel string) (*models.OpenAIChatCompletionRequest, error) {
	// Map Claude model to OpenAI model
	openaiModel := mapClaudeModelToOpenAI(claudeReq.Model, bigModel, smallModel)

	// Convert messages
	openaiMessages, err := convertMessages(claudeReq.Messages, claudeReq.System)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Create OpenAI request
	openaiReq := &models.OpenAIChatCompletionRequest{
		Model:       openaiModel,
		Messages:    openaiMessages,
		MaxTokens:   &claudeReq.MaxTokens,
		Temperature: claudeReq.Temperature,
		TopP:        claudeReq.TopP,
		Stream:      claudeReq.Stream,
	}

	// Convert stop sequences
	if len(claudeReq.StopSequences) > 0 {
		if len(claudeReq.StopSequences) == 1 {
			openaiReq.Stop = claudeReq.StopSequences[0]
		} else {
			openaiReq.Stop = claudeReq.StopSequences
		}
	}

	// Convert tools
	if len(claudeReq.Tools) > 0 {
		openaiTools, err := convertTools(claudeReq.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		openaiReq.Tools = openaiTools
		openaiReq.ToolChoice = convertToolChoice(claudeReq.ToolChoice)
	}

	return openaiReq, nil
}

// mapClaudeModelToOpenAI maps Claude model names to OpenAI model names
func mapClaudeModelToOpenAI(claudeModel, bigModel, smallModel string) string {
	claudeModelLower := strings.ToLower(claudeModel)

	// Check for haiku models (small)
	if strings.Contains(claudeModelLower, "haiku") {
		return smallModel
	}

	// Check for sonnet or opus models (big)
	if strings.Contains(claudeModelLower, "sonnet") || strings.Contains(claudeModelLower, "opus") {
		return bigModel
	}

	// Default to big model for unknown Claude models
	return bigModel
}

// convertMessages converts Claude messages to OpenAI format
func convertMessages(claudeMessages []models.ClaudeMessage, system interface{}) ([]models.OpenAIMessage, error) {
	var openaiMessages []models.OpenAIMessage

	// Add system message if present
	if system != nil {
		systemContent, err := convertContent(system)
		if err != nil {
			return nil, fmt.Errorf("failed to convert system message: %w", err)
		}
		openaiMessages = append(openaiMessages, models.OpenAIMessage{
			Role:    "system",
			Content: systemContent,
		})
	}

	// Convert regular messages
	for _, msg := range claudeMessages {
		content, err := convertContent(msg.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message content: %w", err)
		}

		openaiMsg := models.OpenAIMessage{
			Role:    msg.Role,
			Content: content,
		}

		openaiMessages = append(openaiMessages, openaiMsg)
	}

	return openaiMessages, nil
}

// convertContent converts Claude content to OpenAI format
func convertContent(content interface{}) (interface{}, error) {
	switch v := content.(type) {
	case string:
		return v, nil
	case []interface{}:
		// Handle content blocks
		var parts []map[string]interface{}
		for _, block := range v {
			if blockMap, ok := block.(map[string]interface{}); ok {
				part := make(map[string]interface{})

				if blockType, exists := blockMap["type"]; exists {
					switch blockType {
					case "text":
						part["type"] = "text"
						if text, exists := blockMap["text"]; exists {
							part["text"] = text
						}
					case "image":
						part["type"] = "image_url"
						if source, exists := blockMap["source"]; exists {
							if sourceMap, ok := source.(map[string]interface{}); ok {
								if mediaType, exists := sourceMap["media_type"]; exists {
									if data, exists := sourceMap["data"]; exists {
										imageURL := fmt.Sprintf("data:%s;base64,%s", mediaType, data)
										part["image_url"] = map[string]interface{}{
											"url": imageURL,
										}
									}
								}
							}
						}
					case "tool_use":
						// Handle tool use - convert to OpenAI tool call format
						if name, exists := blockMap["name"]; exists {
							if input, exists := blockMap["input"]; exists {
								inputJSON, _ := json.Marshal(input)
								return map[string]interface{}{
									"role": "assistant",
									"tool_calls": []models.OpenAIToolCall{
										{
											ID:   fmt.Sprintf("call_%s", blockMap["id"]),
											Type: "function",
											Function: models.OpenAIFunctionCall{
												Name:      name.(string),
												Arguments: string(inputJSON),
											},
										},
									},
								}, nil
							}
						}
					case "tool_result":
						// Handle tool result - convert to OpenAI tool response format
						if toolUseID, exists := blockMap["tool_use_id"]; exists {
							if content, exists := blockMap["content"]; exists {
								return map[string]interface{}{
									"role":         "tool",
									"tool_call_id": fmt.Sprintf("call_%s", toolUseID),
									"content":      content,
								}, nil
							}
						}
					}
				}

				if len(part) > 0 {
					parts = append(parts, part)
				}
			}
		}

		if len(parts) == 1 && parts[0]["type"] == "text" {
			return parts[0]["text"], nil
		}
		return parts, nil
	default:
		return v, nil
	}
}

// convertTools converts Claude tools to OpenAI format
func convertTools(claudeTools []models.ClaudeTool) ([]models.OpenAITool, error) {
	var openaiTools []models.OpenAITool

	for _, tool := range claudeTools {
		openaiTool := models.OpenAITool{
			Type: "function",
			Function: models.OpenAIFunctionDef{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		}
		openaiTools = append(openaiTools, openaiTool)
	}

	return openaiTools, nil
}

// convertToolChoice converts Claude tool choice to OpenAI format
func convertToolChoice(claudeToolChoice interface{}) interface{} {
	if claudeToolChoice == nil {
		return nil
	}

	switch v := claudeToolChoice.(type) {
	case string:
		switch v {
		case "auto":
			return "auto"
		case "required":
			return "required"
		default:
			return "auto"
		}
	case map[string]interface{}:
		if toolType, exists := v["type"]; exists && toolType == "tool" {
			if name, exists := v["name"]; exists {
				return map[string]interface{}{
					"type": "function",
					"function": map[string]interface{}{
						"name": name,
					},
				}
			}
		}
	}

	return "auto"
}

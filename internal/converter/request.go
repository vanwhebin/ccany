package converter

import (
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
func convertMessages(claudeMessages []models.ClaudeMessage, system interface{}) ([]models.Message, error) {
	var openaiMessages []models.Message

	// Add system message if present
	if system != nil {
		systemContent, err := convertContentToString(system)
		if err != nil {
			return nil, fmt.Errorf("failed to convert system message: %w", err)
		}
		openaiMessages = append(openaiMessages, models.Message{
			Role:    "system",
			Content: systemContent,
		})
	}

	// Convert regular messages
	for _, msg := range claudeMessages {
		content, err := convertContentToString(msg.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message content: %w", err)
		}

		openaiMsg := models.Message{
			Role:    msg.Role,
			Content: content,
		}

		openaiMessages = append(openaiMessages, openaiMsg)
	}

	return openaiMessages, nil
}

// convertContentToString converts Claude content to a simple string format
func convertContentToString(content interface{}) (string, error) {
	switch v := content.(type) {
	case string:
		return v, nil
	case []interface{}:
		// Handle content blocks - extract text content
		var textParts []string
		for _, block := range v {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockType, exists := blockMap["type"]; exists {
					switch blockType {
					case "text":
						if text, exists := blockMap["text"]; exists {
							if textStr, ok := text.(string); ok {
								textParts = append(textParts, textStr)
							}
						}
						// For now, skip image and tool blocks in simple string conversion
					}
				}
			}
		}
		return strings.Join(textParts, " "), nil
	default:
		// Try to convert to string
		if str, ok := v.(string); ok {
			return str, nil
		}
		return fmt.Sprintf("%v", v), nil
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

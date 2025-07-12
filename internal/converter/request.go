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
	hasTools := len(claudeReq.Tools) > 0
	openaiMessages, err := convertMessagesWithToolPrompt(claudeReq.Messages, claudeReq.System, hasTools)
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

		// Enhanced tool choice handling for better tool usage
		toolChoice := convertToolChoice(claudeReq.ToolChoice)

		// 根据工具类型智能设置 tool_choice
		if containsFileOperationTools(claudeReq.Tools) {
			// 对于文件操作工具，强制使用 required 确保调用
			toolChoice = "required"
		} else if toolChoice == nil || toolChoice == "auto" {
			// 对于其他工具，优先使用 required 提高调用率
			toolChoice = "required"
		}

		openaiReq.ToolChoice = toolChoice
	}

	return openaiReq, nil
}

// mapClaudeModelToOpenAI maps Claude model names to OpenAI model names
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

// convertMessagesWithToolPrompt converts Claude messages and adds tool usage instruction
func convertMessagesWithToolPrompt(claudeMessages []models.ClaudeMessage, system interface{}, hasTools bool) ([]models.Message, error) {
	var openaiMessages []models.Message

	// Add system message with tool instruction if present
	if system != nil {
		systemContent, err := convertContentToString(system)
		if err != nil {
			return nil, fmt.Errorf("failed to convert system message: %w", err)
		}

		// Add tool usage instruction for better compliance
		if hasTools {
			systemContent += "\n\n=== MANDATORY TOOL CALLING REQUIREMENTS ===\nYou MUST call tools using OpenAI function calling format:\n1. Use tool_calls array format, NEVER use <antsArtifact> or any XML format\n2. Each tool call must include: id, type: \"function\", function: {name, arguments}\n3. Format example: {\"tool_calls\": [{\"id\": \"call_xxx\", \"type\": \"function\", \"function\": {\"name\": \"tool_name\", \"arguments\": \"{JSON_params}\"}}]}\n4. When file operations are needed, immediately call the appropriate tool without description\n5. For code creation, file editing, command execution tasks, you MUST use tools to complete them\n\nIMPORTANT: You are interacting with an OpenAI API compatible system. Strictly follow OpenAI function calling specifications."
		}

		openaiMessages = append(openaiMessages, models.Message{
			Role:    "system",
			Content: systemContent,
		})
	} else if hasTools {
		// Add tool instruction even without existing system message
		openaiMessages = append(openaiMessages, models.Message{
			Role:    "system",
			Content: "=== MANDATORY TOOL CALLING REQUIREMENTS ===\nYou MUST call tools using OpenAI function calling format:\n1. Use tool_calls array format, NEVER use <antsArtifact> or any XML format\n2. Each tool call must include: id, type: \"function\", function: {name, arguments}\n3. Format example: {\"tool_calls\": [{\"id\": \"call_xxx\", \"type\": \"function\", \"function\": {\"name\": \"tool_name\", \"arguments\": \"{JSON_params}\"}}]}\n4. When file operations are needed, immediately call the appropriate tool without description\n5. For code creation, file editing, command execution tasks, you MUST use tools to complete them\n\nIMPORTANT: You are interacting with an OpenAI API compatible system. Strictly follow OpenAI function calling specifications.",
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

// convertTools converts Claude tools to OpenAI format with enhanced descriptions
func convertTools(claudeTools []models.ClaudeTool) ([]models.OpenAITool, error) {
	var openaiTools []models.OpenAITool

	for _, tool := range claudeTools {
		// 增强工具描述，明确指出何时使用
		enhancedDescription := tool.Description
		if tool.Name == "str_replace_editor" || tool.Name == "str_replace_based_edit_tool" {
			enhancedDescription += " MUST be used when creating, editing or modifying files. Required for all file operations."
		} else if tool.Name == "bash" {
			enhancedDescription += " MUST be used when executing system commands, running scripts or performing system operations."
		} else if strings.Contains(strings.ToLower(tool.Name), "file") {
			enhancedDescription += " MUST be used for file operations."
		}

		openaiTool := models.OpenAITool{
			Type: "function",
			Function: models.OpenAIFunctionDef{
				Name:        tool.Name,
				Description: enhancedDescription,
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

// containsFileOperationTools checks if the tools contain file operation capabilities
func containsFileOperationTools(tools []models.ClaudeTool) bool {
	fileOperationKeywords := []string{"file", "write", "create", "edit", "bash", "str_replace"}

	for _, tool := range tools {
		toolNameLower := strings.ToLower(tool.Name)
		toolDescLower := strings.ToLower(tool.Description)

		for _, keyword := range fileOperationKeywords {
			if strings.Contains(toolNameLower, keyword) || strings.Contains(toolDescLower, keyword) {
				return true
			}
		}
	}
	return false
}

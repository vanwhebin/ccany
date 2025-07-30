package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"ccany/internal/models"
)

// GeminiConverter handles conversions involving Gemini format
type GeminiConverter struct{}

// NewGeminiConverter creates a new Gemini converter
func NewGeminiConverter() *GeminiConverter {
	return &GeminiConverter{}
}

// ConvertFromOpenAI converts OpenAI format request to Gemini format
func (c *GeminiConverter) ConvertFromOpenAI(openaiReq *models.OpenAIChatCompletionRequest) (*models.GeminiRequest, error) {
	geminiReq := &models.GeminiRequest{
		Contents: []models.GeminiContent{},
	}

	// Convert messages to contents
	var systemInstruction *models.GeminiContent
	for _, msg := range openaiReq.Messages {
		if msg.Role == "system" {
			// Handle system message as system instruction
			systemInstruction = &models.GeminiContent{
				Parts: []models.GeminiPart{
					{Text: msg.Content},
				},
			}
		} else {
			// Convert user/assistant messages
			role := c.mapOpenAIRoleToGemini(msg.Role)
			var parts []models.GeminiPart

			// Handle tool calls in assistant messages
			if len(msg.ToolCalls) > 0 {
				for _, toolCall := range msg.ToolCalls {
					parts = append(parts, models.GeminiPart{
						FunctionCall: &models.GeminiFunctionCall{
							Name: toolCall.Function.Name,
							Args: c.parseToolCallArguments(toolCall.Function.Arguments),
						},
					})
				}
			} else {
				// Regular text content
				parts = append(parts, models.GeminiPart{
					Text: msg.Content,
				})
			}

			geminiReq.Contents = append(geminiReq.Contents, models.GeminiContent{
				Role:  role,
				Parts: parts,
			})
		}
	}

	// Set system instruction
	if systemInstruction != nil {
		geminiReq.SystemInstruction = systemInstruction
	}

	// Convert generation config
	if openaiReq.Temperature != nil || openaiReq.MaxTokens != nil || openaiReq.TopP != nil {
		config := &models.GeminiGenerationConfig{}
		if openaiReq.Temperature != nil {
			config.Temperature = openaiReq.Temperature
		}
		if openaiReq.MaxTokens != nil {
			config.MaxOutputTokens = openaiReq.MaxTokens
		}
		if openaiReq.TopP != nil {
			config.TopP = openaiReq.TopP
		}
		geminiReq.GenerationConfig = config
	}

	// Convert tools
	if len(openaiReq.Tools) > 0 {
		geminiTools := []models.GeminiTool{}
		for _, tool := range openaiReq.Tools {
			if tool.Type == "function" {
				funcDecl := models.GeminiFunctionDeclaration{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
				}

				// Convert parameters schema
				if tool.Function.Parameters != nil {
					params := &models.GeminiFunctionParameters{
						Type:       "object",
						Properties: tool.Function.Parameters,
					}
					if required, ok := tool.Function.Parameters["required"].([]interface{}); ok {
						reqStrings := make([]string, len(required))
						for i, r := range required {
							if str, ok := r.(string); ok {
								reqStrings[i] = str
							}
						}
						params.Required = reqStrings
					}
					funcDecl.Parameters = params
				}

				geminiTools = append(geminiTools, models.GeminiTool{
					FunctionDeclarations: []models.GeminiFunctionDeclaration{funcDecl},
				})
			}
		}
		geminiReq.Tools = geminiTools

		// Set tool config based on tool choice
		if openaiReq.ToolChoice != nil {
			config := &models.GeminiToolConfig{
				FunctionCallingConfig: &models.GeminiFunctionCallingConfig{},
			}

			switch tc := openaiReq.ToolChoice.(type) {
			case string:
				switch tc {
				case "auto":
					config.FunctionCallingConfig.Mode = "AUTO"
				case "required":
					config.FunctionCallingConfig.Mode = "ANY"
				case "none":
					config.FunctionCallingConfig.Mode = "NONE"
				}
			}
			geminiReq.ToolConfig = config
		}
	}

	return geminiReq, nil
}

// ConvertToOpenAI converts Gemini format request to OpenAI format
func (c *GeminiConverter) ConvertToOpenAI(geminiReq *models.GeminiRequest) (*models.OpenAIChatCompletionRequest, error) {
	openaiReq := &models.OpenAIChatCompletionRequest{
		Messages: []models.Message{},
	}

	// Handle system instruction
	if geminiReq.SystemInstruction != nil {
		systemContent := c.extractTextFromParts(geminiReq.SystemInstruction.Parts)
		if systemContent != "" {
			openaiReq.Messages = append(openaiReq.Messages, models.Message{
				Role:    "system",
				Content: systemContent,
			})
		}
	}

	// Convert contents to messages
	for _, content := range geminiReq.Contents {
		role := c.mapGeminiRoleToOpenAI(content.Role)

		// Check for function calls
		var toolCalls []models.OpenAIToolCall
		var textContent string

		for _, part := range content.Parts {
			if part.FunctionCall != nil {
				toolCalls = append(toolCalls, models.OpenAIToolCall{
					ID:   fmt.Sprintf("call_%d", len(toolCalls)+1),
					Type: "function",
					Function: models.OpenAIFunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: c.convertArgsToJSON(part.FunctionCall.Args),
					},
				})
			} else if part.Text != "" {
				textContent += part.Text
			}
		}

		msg := models.Message{
			Role:    role,
			Content: textContent,
		}

		if len(toolCalls) > 0 {
			msg.ToolCalls = toolCalls
		}

		openaiReq.Messages = append(openaiReq.Messages, msg)
	}

	// Convert generation config
	if geminiReq.GenerationConfig != nil {
		config := geminiReq.GenerationConfig
		if config.Temperature != nil {
			openaiReq.Temperature = config.Temperature
		}
		if config.MaxOutputTokens != nil {
			openaiReq.MaxTokens = config.MaxOutputTokens
		}
		if config.TopP != nil {
			openaiReq.TopP = config.TopP
		}
	}

	// Convert tools
	if len(geminiReq.Tools) > 0 {
		var openaiTools []models.OpenAITool
		for _, tool := range geminiReq.Tools {
			for _, funcDecl := range tool.FunctionDeclarations {
				openaiTool := models.OpenAITool{
					Type: "function",
					Function: models.OpenAIFunctionDef{
						Name:        funcDecl.Name,
						Description: funcDecl.Description,
					},
				}

				if funcDecl.Parameters != nil {
					openaiTool.Function.Parameters = funcDecl.Parameters.Properties
				}

				openaiTools = append(openaiTools, openaiTool)
			}
		}
		openaiReq.Tools = openaiTools
	}

	return openaiReq, nil
}

// ConvertFromClaude converts Claude format request to Gemini format
func (c *GeminiConverter) ConvertFromClaude(claudeReq *models.ClaudeMessagesRequest) (*models.GeminiRequest, error) {
	geminiReq := &models.GeminiRequest{
		Contents: []models.GeminiContent{},
	}

	// Handle system message
	if claudeReq.System != nil {
		systemContent, err := c.convertContentToString(claudeReq.System)
		if err == nil && systemContent != "" {
			geminiReq.SystemInstruction = &models.GeminiContent{
				Parts: []models.GeminiPart{
					{Text: systemContent},
				},
			}
		}
	}

	// Convert messages to contents
	for _, msg := range claudeReq.Messages {
		role := c.mapClaudeRoleToGemini(msg.Role)
		parts := []models.GeminiPart{}

		// Handle different content types
		switch content := msg.Content.(type) {
		case string:
			parts = append(parts, models.GeminiPart{Text: content})
		case []interface{}:
			// Handle content blocks
			for _, block := range content {
				if blockMap, ok := block.(map[string]interface{}); ok {
					if blockType, exists := blockMap["type"]; exists {
						switch blockType {
						case "text":
							if text, ok := blockMap["text"].(string); ok {
								parts = append(parts, models.GeminiPart{Text: text})
							}
						case "tool_use":
							if name, ok := blockMap["name"].(string); ok {
								args := make(map[string]interface{})
								if input, ok := blockMap["input"].(map[string]interface{}); ok {
									args = input
								}
								parts = append(parts, models.GeminiPart{
									FunctionCall: &models.GeminiFunctionCall{
										Name: name,
										Args: args,
									},
								})
							}
						}
					}
				}
			}
		}

		if len(parts) > 0 {
			geminiReq.Contents = append(geminiReq.Contents, models.GeminiContent{
				Role:  role,
				Parts: parts,
			})
		}
	}

	// Convert generation config
	config := &models.GeminiGenerationConfig{}
	if claudeReq.Temperature != nil {
		config.Temperature = claudeReq.Temperature
	}
	if claudeReq.MaxTokens > 0 {
		config.MaxOutputTokens = &claudeReq.MaxTokens
	}
	if claudeReq.TopP != nil {
		config.TopP = claudeReq.TopP
	}
	if claudeReq.TopK != nil {
		config.TopK = claudeReq.TopK
	}
	if len(claudeReq.StopSequences) > 0 {
		config.StopSequences = claudeReq.StopSequences
	}
	geminiReq.GenerationConfig = config

	// Convert tools
	if len(claudeReq.Tools) > 0 {
		var geminiTools []models.GeminiTool
		for _, tool := range claudeReq.Tools {
			funcDecl := models.GeminiFunctionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
			}

			if tool.InputSchema != nil {
				params := &models.GeminiFunctionParameters{
					Type:       "object",
					Properties: tool.InputSchema,
				}
				if required, ok := tool.InputSchema["required"].([]interface{}); ok {
					reqStrings := make([]string, len(required))
					for i, r := range required {
						if str, ok := r.(string); ok {
							reqStrings[i] = str
						}
					}
					params.Required = reqStrings
				}
				funcDecl.Parameters = params
			}

			geminiTools = append(geminiTools, models.GeminiTool{
				FunctionDeclarations: []models.GeminiFunctionDeclaration{funcDecl},
			})
		}
		geminiReq.Tools = geminiTools
	}

	return geminiReq, nil
}

// ConvertToClaude converts Gemini response to Claude format
func (c *GeminiConverter) ConvertToClaude(geminiResp *models.GeminiResponse, originalReq *models.ClaudeMessagesRequest) (*models.ClaudeResponse, error) {
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Gemini response")
	}

	candidate := geminiResp.Candidates[0]
	var content []models.ClaudeContentBlock

	if candidate.Content != nil {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				content = append(content, models.ClaudeContentBlock{
					Type: "text",
					Text: part.Text,
				})
			}
			if part.FunctionCall != nil {
				content = append(content, models.ClaudeContentBlock{
					Type:  "tool_use",
					ID:    fmt.Sprintf("tool_%d", len(content)+1),
					Name:  part.FunctionCall.Name,
					Input: part.FunctionCall.Args,
				})
			}
		}
	}

	// Map finish reason
	stopReason := c.mapGeminiFinishReasonToClaudeStopReason(candidate.FinishReason)

	// Create usage information
	usage := models.ClaudeUsage{}
	if geminiResp.UsageMetadata != nil {
		usage.InputTokens = geminiResp.UsageMetadata.PromptTokenCount
		usage.OutputTokens = geminiResp.UsageMetadata.CandidatesTokenCount
	}

	claudeResp := &models.ClaudeResponse{
		ID:           fmt.Sprintf("msg_%d", len(content)),
		Type:         "message",
		Role:         "assistant",
		Content:      content,
		Model:        originalReq.Model,
		StopReason:   stopReason,
		StopSequence: nil,
		Usage:        usage,
	}

	return claudeResp, nil
}

// Helper functions

func (c *GeminiConverter) mapOpenAIRoleToGemini(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "model"
	default:
		return "user"
	}
}

func (c *GeminiConverter) mapGeminiRoleToOpenAI(role string) string {
	switch role {
	case "user":
		return "user"
	case "model":
		return "assistant"
	default:
		return "user"
	}
}

func (c *GeminiConverter) mapClaudeRoleToGemini(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "model"
	default:
		return "user"
	}
}

func (c *GeminiConverter) mapGeminiFinishReasonToClaudeStopReason(finishReason string) string {
	switch finishReason {
	case "STOP":
		return "end_turn"
	case "MAX_TOKENS":
		return "max_tokens"
	case "SAFETY":
		return "stop_sequence"
	case "RECITATION":
		return "stop_sequence"
	default:
		return "end_turn"
	}
}

func (c *GeminiConverter) extractTextFromParts(parts []models.GeminiPart) string {
	var texts []string
	for _, part := range parts {
		if part.Text != "" {
			texts = append(texts, part.Text)
		}
	}
	return strings.Join(texts, " ")
}

func (c *GeminiConverter) parseToolCallArguments(args string) map[string]interface{} {
	result := make(map[string]interface{})
	if args == "" {
		return result
	}
	if err := json.Unmarshal([]byte(args), &result); err != nil {
		return make(map[string]interface{})
	}
	return result
}

func (c *GeminiConverter) convertArgsToJSON(args map[string]interface{}) string {
	if len(args) == 0 {
		return "{}"
	}
	data, err := json.Marshal(args)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (c *GeminiConverter) convertContentToString(content interface{}) (string, error) {
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
					}
				}
			}
		}
		return strings.Join(textParts, " "), nil
	default:
		if str, ok := v.(string); ok {
			return str, nil
		}
		return fmt.Sprintf("%v", v), nil
	}
}

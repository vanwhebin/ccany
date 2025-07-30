package converter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ccany/internal/models"
	"ccany/internal/toolmapping"
	"github.com/sirupsen/logrus"
)

// ClaudeConverter handles conversions involving Claude/Anthropic format
type ClaudeConverter struct {
	logger *logrus.Logger
}

// NewClaudeConverter creates a new Claude converter
func NewClaudeConverter() *ClaudeConverter {
	return &ClaudeConverter{}
}

// NewClaudeConverterWithLogger creates a new Claude converter with logger
func NewClaudeConverterWithLogger(logger *logrus.Logger) *ClaudeConverter {
	return &ClaudeConverter{
		logger: logger,
	}
}

// ConvertFromOpenAI converts OpenAI response to Claude format
func (c *ClaudeConverter) ConvertFromOpenAI(openaiResp *models.OpenAIChatCompletionResponse, originalReq *models.ClaudeMessagesRequest) (*models.ClaudeResponse, error) {
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice := openaiResp.Choices[0]

	// Convert content
	content, err := c.convertMessageToClaudeContent(choice.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message content: %w", err)
	}

	// Map finish reason - improved mapping based on tool use
	stopReason := c.mapFinishReasonToClaudeStopReason(choice.FinishReason)

	// If content contains tool_use, override stop_reason regardless of finish_reason
	hasTools := c.hasToolUseContent(content)
	if hasTools {
		stopReason = "tool_use"
	}

	// Generate proper message ID if not provided
	messageID := openaiResp.ID
	if messageID == "" {
		messageID = fmt.Sprintf("msg_%d", time.Now().Unix())
	}

	claudeResp := &models.ClaudeResponse{
		ID:           messageID,
		Type:         "message",
		Role:         "assistant",
		Content:      content,
		Model:        originalReq.Model, // Use original Claude model name
		StopReason:   stopReason,
		StopSequence: nil, // Always nil for Claude responses
		Usage: models.ClaudeUsage{
			InputTokens:  openaiResp.Usage.PromptTokens,
			OutputTokens: openaiResp.Usage.CompletionTokens,
		},
	}

	return claudeResp, nil
}

// ConvertStreamFromOpenAI converts OpenAI streaming response to Claude format
func (c *ClaudeConverter) ConvertStreamFromOpenAI(openaiChunk *models.OpenAIStreamResponse, originalReq *models.ClaudeMessagesRequest, ctx *StreamingContext) ([]models.ClaudeStreamEvent, error) {
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

	// Handle finish reason
	if choice.FinishReason != "" {
		// Send content block stop if content was started
		if ctx.ContentStarted {
			events = append(events, models.ClaudeStreamEvent{
				Type:  "content_block_stop",
				Index: 0,
			})
		}

		stopReason := c.mapFinishReasonToClaudeStopReason(choice.FinishReason)

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
func (c *ClaudeConverter) convertMessageToClaudeContent(msg models.Message) ([]models.ClaudeContentBlock, error) {
	var content []models.ClaudeContentBlock

	// 1. First handle standard OpenAI tool calls (highest priority)
	for _, toolCall := range msg.ToolCalls {
		// Map OpenAI tool name to Claude tool name
		claudeToolName := toolmapping.MapOpenAIToClaudeName(toolCall.Function.Name)

		// Parse arguments from string to interface{}
		var args interface{}
		if toolCall.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return nil, fmt.Errorf("failed to parse tool call arguments: %w", err)
			}
		}

		content = append(content, models.ClaudeContentBlock{
			Type:  "tool_use",
			ID:    toolCall.ID,
			Name:  claudeToolName,
			Input: args,
		})
	}

	// 2. If no standard tool calls, try to parse custom formats from content
	cleanedContent := msg.Content
	if len(msg.ToolCalls) == 0 {
		var customToolCalls []models.ClaudeContentBlock
		cleanedContent, customToolCalls = c.ParseCustomFormatFromContent(msg.Content)
		content = append(content, customToolCalls...)
	}

	// 3. Handle remaining text content
	if cleanedContent != "" {
		content = append(content, models.ClaudeContentBlock{
			Type: "text",
			Text: cleanedContent,
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
func (c *ClaudeConverter) mapFinishReasonToClaudeStopReason(finishReason string) string {
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
	case "function_call": // Legacy OpenAI function calling
		return "tool_use"
	default:
		return "end_turn"
	}
}

// ParseCustomFormatFromContent parses custom tool call format from content using robust parsing
func (c *ClaudeConverter) ParseCustomFormatFromContent(content string) (string, []models.ClaudeContentBlock) {
	parser := NewToolCallParser()

	// First try the robust parser for Unicode formats
	cleanContent, toolCalls := parser.ParseContent(content)

	// If no tool calls found, try to parse embedded JSON tool calls
	if len(toolCalls) == 0 {
		jsonToolCalls, jsonCleanContent := c.parseEmbeddedJSONToolCalls(content)
		if len(jsonToolCalls) > 0 {
			return jsonCleanContent, jsonToolCalls
		}
	}

	return cleanContent, toolCalls
}

// parseEmbeddedJSONToolCalls parses tool calls embedded as JSON in text content
func (c *ClaudeConverter) parseEmbeddedJSONToolCalls(content string) ([]models.ClaudeContentBlock, string) {
	var toolCalls []models.ClaudeContentBlock

	// Enhanced pattern matching with more variations
	patterns := []struct {
		pattern string
		desc    string
	}{
		{`{"tool_calls":`, "Direct JSON"},
		{`"tool_calls": [`, "Spaced JSON"},
		{`"tool_calls":[`, "No space JSON"},
		{`tool_calls": [`, "Within larger JSON"},
		{`{"tool_calls" :`, "Space before colon"},
		{`{ "tool_calls":`, "Space after brace"},
		{`{ "tool_calls" : [`, "Space everywhere"},
		{`{"tool_calls" : [`, "Space around colon"},
		{`"tool_calls" : [`, "Quoted with spaces"},
	}

	jsonStart := -1
	var matchedPattern string

	for _, p := range patterns {
		if idx := strings.Index(content, p.pattern); idx != -1 {
			// Find the start of the JSON object by looking backwards for '{'
			jsonStart = idx
			for jsonStart > 0 && content[jsonStart] != '{' {
				jsonStart--
			}
			matchedPattern = p.desc
			break
		}
	}

	if jsonStart == -1 {
		return toolCalls, content
	}

	// Find the matching closing brace using improved bracket counting
	jsonEnd := c.findMatchingBrace(content, jsonStart)
	if jsonEnd == -1 {
		// Log warning but don't fail
		if c.logger != nil {
			c.logger.Warnf("Failed to find matching brace for JSON starting at position %d", jsonStart)
		}
		return toolCalls, content
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	// Fix malformed JSON where arguments is not properly escaped
	jsonStr = c.fixArgumentsJSON(jsonStr)

	// Try multiple parsing strategies
	toolCalls = c.tryParseToolCalls(jsonStr, matchedPattern)

	// Remove JSON tool calls from content only if we successfully parsed them
	if len(toolCalls) > 0 {
		beforeJSON := strings.TrimSpace(content[:jsonStart])
		afterJSON := ""
		if jsonEnd+1 < len(content) {
			afterJSON = strings.TrimSpace(content[jsonEnd+1:])
		}

		// Combine the parts with proper spacing
		cleanContent := beforeJSON
		if beforeJSON != "" && afterJSON != "" {
			cleanContent += " " + afterJSON
		} else if afterJSON != "" {
			cleanContent = afterJSON
		}

		return toolCalls, cleanContent
	}

	return toolCalls, content
}

// tryParseToolCalls attempts multiple strategies to parse tool calls
func (c *ClaudeConverter) tryParseToolCalls(jsonStr string, pattern string) []models.ClaudeContentBlock {
	var toolCalls []models.ClaudeContentBlock

	// Strategy 1: Standard parsing
	var toolCallContainer struct {
		ToolCalls []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &toolCallContainer); err == nil {
		for _, tc := range toolCallContainer.ToolCalls {
			// Parse function arguments with error handling
			args, err := c.parseToolArguments(tc.Function.Arguments)
			if err != nil {
				// Log but continue with empty args
				if c.logger != nil {
					c.logger.Warnf("Failed to parse tool arguments for %s: %v", tc.Function.Name, err)
				}
				args = make(map[string]interface{})
			}

			// Map tool name to Claude tool name
			claudeToolName := c.mapToolName(tc.Function.Name)

			toolCalls = append(toolCalls, models.ClaudeContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  claudeToolName,
				Input: args,
			})
		}
		return toolCalls
	}

	// Strategy 2: Try alternative structure where arguments might be an object
	var altContainer struct {
		ToolCalls []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &altContainer); err == nil {
		for _, tc := range altContainer.ToolCalls {
			claudeToolName := c.mapToolName(tc.Function.Name)
			toolCalls = append(toolCalls, models.ClaudeContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  claudeToolName,
				Input: tc.Function.Arguments,
			})
		}
		return toolCalls
	}

	// Log failure if debug logging is available
	if c.logger != nil {
		c.logger.Debugf("Failed to parse tool calls from JSON (pattern: %s): %s", pattern, jsonStr)
	}

	return toolCalls
}

// parseToolArguments safely parses tool arguments from string
func (c *ClaudeConverter) parseToolArguments(argsStr string) (map[string]interface{}, error) {
	if argsStr == "" {
		return make(map[string]interface{}), nil
	}

	var args map[string]interface{}

	// Try direct parsing first
	if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
		return args, nil
	}

	// Try fixing common issues
	// 1. Unescape if double-escaped
	if strings.Contains(argsStr, "\\\"") {
		unescaped := strings.ReplaceAll(argsStr, "\\\"", "\"")
		unescaped = strings.ReplaceAll(unescaped, "\\\\", "\\")
		if err := json.Unmarshal([]byte(unescaped), &args); err == nil {
			return args, nil
		}
	}

	// 2. Try wrapping in quotes if it looks like it's missing them
	if !strings.HasPrefix(argsStr, "{") && !strings.HasPrefix(argsStr, "[") {
		if err := json.Unmarshal([]byte("\""+argsStr+"\""), &args); err == nil {
			return args, nil
		}
	}

	return nil, fmt.Errorf("unable to parse arguments: %s", argsStr)
}

// findMatchingBrace finds the matching closing brace for JSON
func (c *ClaudeConverter) findMatchingBrace(content string, start int) int {
	braceCount := 0
	inString := false
	escapeNext := false

	for i := start; i < len(content); i++ {
		char := content[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if !inString {
			if char == '{' {
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 {
					return i
				}
			}
		}
	}

	return -1
}

// mapToolName maps various tool name formats to Claude tool names using the centralized mapper
func (c *ClaudeConverter) mapToolName(toolName string) string {
	// Try custom mapping first
	claudeToolName := toolmapping.MapCustomToClaudeName(toolName)
	if claudeToolName != toolName {
		return claudeToolName
	}

	// Try OpenAI mapping
	claudeToolName = toolmapping.MapOpenAIToClaudeName(toolName)
	if claudeToolName != toolName {
		return claudeToolName
	}

	// Add dynamic mappings for common patterns if not already in the mapper
	switch strings.ToLower(toolName) {
	case "create_file", "createfile", "file_create":
		toolmapping.AddCustomMapping(toolName, "Write")
		return "Write"
	case "read_file", "readfile", "file_read":
		toolmapping.AddCustomMapping(toolName, "Read")
		return "Read"
	case "edit_file", "editfile", "file_edit":
		toolmapping.AddCustomMapping(toolName, "Edit")
		return "Edit"
	case "run_command", "runcommand", "command", "exec":
		toolmapping.AddCustomMapping(toolName, "Bash")
		return "Bash"
	default:
		return toolName
	}
}

// fixArgumentsJSON fixes malformed JSON where arguments contains unescaped nested JSON
func (c *ClaudeConverter) fixArgumentsJSON(jsonStr string) string {
	// Try multiple patterns for arguments field
	patterns := []string{
		`"arguments": "`,
		`"arguments":"`,
		`"arguments" : "`,
		`"arguments" :"`,
	}

	for _, pattern := range patterns {
		start := strings.Index(jsonStr, pattern)
		if start != -1 {
			fixedStr := c.fixArgumentsAtPosition(jsonStr, start, len(pattern))
			if fixedStr != jsonStr {
				return fixedStr
			}
		}
	}

	return jsonStr
}

// fixArgumentsAtPosition fixes arguments JSON at a specific position
func (c *ClaudeConverter) fixArgumentsAtPosition(jsonStr string, start int, patternLen int) string {
	contentStart := start + patternLen
	braceCount := 0
	bracketCount := 0
	i := contentStart
	var contentEnd int = -1
	escaped := false
	inString := false

	for i < len(jsonStr) {
		char := jsonStr[i]

		if escaped {
			escaped = false
			i++
			continue
		}

		if char == '\\' {
			escaped = true
			i++
			continue
		}

		// Track string boundaries to avoid counting braces inside strings
		if char == '"' && (i == contentStart || jsonStr[i-1] != '\\') {
			if !inString {
				// Check if this quote ends the arguments value
				if braceCount == 0 && bracketCount == 0 && i > contentStart {
					contentEnd = i
					break
				}
			}
			inString = !inString
		}

		if !inString {
			switch char {
			case '{':
				braceCount++
			case '}':
				braceCount--
			case '[':
				bracketCount++
			case ']':
				bracketCount--
			}
		}

		i++
	}

	if contentEnd == -1 {
		return jsonStr
	}

	unescapedContent := jsonStr[contentStart:contentEnd]

	// More sophisticated escaping
	escapedContent := c.escapeJSONString(unescapedContent)

	result := jsonStr[:contentStart] + escapedContent + jsonStr[contentEnd:]
	return result
}

// escapeJSONString properly escapes a string for JSON
func (c *ClaudeConverter) escapeJSONString(s string) string {
	// First preserve already escaped quotes
	s = strings.ReplaceAll(s, `\"`, "\x00")

	// Escape unescaped quotes
	s = strings.ReplaceAll(s, `"`, `\"`)

	// Restore already escaped quotes
	s = strings.ReplaceAll(s, "\x00", `\"`)

	// Escape other special characters
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)

	// Fix double escaping
	s = strings.ReplaceAll(s, `\\n`, `\n`)
	s = strings.ReplaceAll(s, `\\r`, `\r`)
	s = strings.ReplaceAll(s, `\\t`, `\t`)

	return s
}

// hasToolUseContent checks if content contains tool_use blocks
func (c *ClaudeConverter) hasToolUseContent(content []models.ClaudeContentBlock) bool {
	for _, block := range content {
		if block.Type == "tool_use" {
			return true
		}
	}
	return false
}

// CreateErrorResponse creates a Claude-formatted error response
func (c *ClaudeConverter) CreateErrorResponse(errorType, message string) *models.ClaudeErrorResponse {
	return &models.ClaudeErrorResponse{
		Type: "error",
		Error: models.ClaudeError{
			Type:    errorType,
			Message: message,
		},
	}
}

// CreateStreamStartEvent creates a stream start event
func (c *ClaudeConverter) CreateStreamStartEvent(messageID, model string) models.ClaudeStreamEvent {
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

// CreateStreamPingEvent creates a ping event for keep-alive
func (c *ClaudeConverter) CreateStreamPingEvent() models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type: "ping",
	}
}

// CreateStreamingContext creates a new streaming context
func (c *ClaudeConverter) CreateStreamingContext(messageID, model string, inputTokens int) *StreamingContext {
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

// CreateStreamStopEvent creates a stream stop event
func (c *ClaudeConverter) CreateStreamStopEvent(usage models.ClaudeUsage) models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type:  "message_delta",
		Usage: &usage,
	}
}

// ToolCallParser represents a robust tool call parser
type ToolCallParser struct {
	patterns []ToolCallPattern
}

// ToolCallPattern defines a pattern for parsing tool calls
type ToolCallPattern struct {
	StartMarker string
	EndMarker   string
	ToolSep     string
	Name        string
}

// NewToolCallParser creates a new parser with predefined patterns
func NewToolCallParser() *ToolCallParser {
	return &ToolCallParser{
		patterns: []ToolCallPattern{
			// Unicode pattern (current backend)
			{
				StartMarker: "<｜tool▁calls▁begin｜><｜tool▁call▁begin｜>function<｜tool▁sep｜>",
				EndMarker:   "<｜tool▁call▁end｜><｜tool▁calls▁end｜>",
				ToolSep:     "<｜tool▁sep｜>",
				Name:        "unicode",
			},
			// Standard pattern (fallback)
			{
				StartMarker: "<|tool_calls_begin|><|tool_call_begin|>function<|tool_sep|>",
				EndMarker:   "<|tool_call_end|><|tool_calls_end|>",
				ToolSep:     "<|tool_sep|>",
				Name:        "standard",
			},
			// Alternative patterns for different backends
			{
				StartMarker: "```tool_call",
				EndMarker:   "```",
				ToolSep:     "\n",
				Name:        "markdown",
			},
		},
	}
}

// ParseContent parses content and extracts tool calls using multiple patterns
func (p *ToolCallParser) ParseContent(content string) (string, []models.ClaudeContentBlock) {
	var allToolCalls []models.ClaudeContentBlock
	cleanContent := content

	// Try each pattern
	for _, pattern := range p.patterns {
		toolCalls, updatedContent := p.parseWithPattern(cleanContent, pattern)
		allToolCalls = append(allToolCalls, toolCalls...)
		cleanContent = updatedContent
	}

	return cleanContent, allToolCalls
}

// parseWithPattern parses content using a specific pattern
func (p *ToolCallParser) parseWithPattern(content string, pattern ToolCallPattern) ([]models.ClaudeContentBlock, string) {
	var toolCalls []models.ClaudeContentBlock
	cleanContent := content
	callCounter := 1

	// Look for multiple tool calls in the same content
	searchContent := content
	offset := 0

	for {
		startPos := strings.Index(searchContent, pattern.StartMarker)
		if startPos == -1 {
			break
		}

		actualStartPos := offset + startPos

		// Extract tool call
		toolCall, endPos := p.extractSingleToolCall(content[actualStartPos:], pattern, callCounter)
		if toolCall != nil {
			toolCalls = append(toolCalls, *toolCall)
			callCounter++

			// Remove this tool call from content
			if endPos > 0 {
				actualEndPos := actualStartPos + endPos
				beforeTool := content[:actualStartPos]
				afterTool := ""
				if actualEndPos < len(content) {
					afterTool = content[actualEndPos:]
				}
				cleanContent = strings.TrimSpace(beforeTool + afterTool)
				content = cleanContent // Update for next iteration

				// Reset search
				searchContent = cleanContent
				offset = 0
			} else {
				// If we couldn't find the end, move past this start marker
				searchContent = searchContent[startPos+len(pattern.StartMarker):]
				offset = actualStartPos + len(pattern.StartMarker)
			}
		} else {
			// If extraction failed, move past this start marker
			searchContent = searchContent[startPos+len(pattern.StartMarker):]
			offset = actualStartPos + len(pattern.StartMarker)
		}
	}

	return toolCalls, cleanContent
}

// extractSingleToolCall extracts a single tool call from content
func (p *ToolCallParser) extractSingleToolCall(content string, pattern ToolCallPattern, callID int) (*models.ClaudeContentBlock, int) {
	if !strings.HasPrefix(content, pattern.StartMarker) {
		return nil, 0
	}

	// Extract everything after the start marker
	afterStart := content[len(pattern.StartMarker):]

	// Find the end marker
	endPos := strings.Index(afterStart, pattern.EndMarker)
	var toolCallContent string
	var totalLength int

	if endPos != -1 {
		toolCallContent = afterStart[:endPos]
		totalLength = len(pattern.StartMarker) + endPos + len(pattern.EndMarker)
	} else {
		// No end marker found, take everything
		toolCallContent = afterStart
		totalLength = len(content)
	}

	// Parse tool name and arguments
	toolName, args := p.parseToolCallContent(toolCallContent, pattern)
	if toolName == "" {
		return nil, 0
	}

	// Map tool name to Claude tool name
	claudeToolName := toolmapping.MapCustomToClaudeName(toolName)

	toolCall := &models.ClaudeContentBlock{
		Type:  "tool_use",
		ID:    fmt.Sprintf("call_%d", callID),
		Name:  claudeToolName,
		Input: args,
	}

	return toolCall, totalLength
}

// parseToolCallContent parses the content inside a tool call
func (p *ToolCallParser) parseToolCallContent(content string, pattern ToolCallPattern) (string, map[string]interface{}) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "", nil
	}

	// First line should contain the tool name
	toolName := strings.TrimSpace(lines[0])

	// Rest should be JSON
	var jsonLines []string
	if len(lines) > 1 {
		jsonLines = lines[1:]
	}

	// Join and clean JSON content
	jsonContent := strings.Join(jsonLines, "\n")
	jsonContent = strings.ReplaceAll(jsonContent, "```", "")
	jsonContent = strings.TrimSpace(jsonContent)

	// Try multiple JSON extraction strategies
	args := p.extractJSON(jsonContent)

	return toolName, args
}

// extractJSON attempts to extract JSON using multiple strategies
func (p *ToolCallParser) extractJSON(content string) map[string]interface{} {
	if content == "" {
		return make(map[string]interface{})
	}

	// Strategy 1: Direct JSON parsing
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(content), &args); err == nil {
		return args
	}

	// Strategy 2: Find JSON boundaries
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonStr := content[jsonStart : jsonEnd+1]

		// Clean up common issues
		jsonStr = strings.ReplaceAll(jsonStr, "\n", " ")
		jsonStr = strings.ReplaceAll(jsonStr, "\t", " ")

		// Normalize multiple spaces
		for strings.Contains(jsonStr, "  ") {
			jsonStr = strings.ReplaceAll(jsonStr, "  ", " ")
		}

		if err := json.Unmarshal([]byte(jsonStr), &args); err == nil {
			return args
		}
	}

	// Strategy 3: Manual key-value extraction for malformed JSON
	return p.extractKeyValuePairs(content)
}

// extractKeyValuePairs manually extracts key-value pairs from malformed JSON
func (p *ToolCallParser) extractKeyValuePairs(content string) map[string]interface{} {
	args := make(map[string]interface{})

	// Remove braces
	content = strings.Trim(content, "{}")
	content = strings.TrimSpace(content)

	if content == "" {
		return args
	}

	// Try to split by commas, but be careful about commas inside strings
	pairs := p.smartSplit(content, ',')

	for _, pair := range pairs {
		if key, value := p.parseKeyValue(strings.TrimSpace(pair)); key != "" {
			args[key] = value
		}
	}

	return args
}

// smartSplit splits by delimiter but respects quoted strings
func (p *ToolCallParser) smartSplit(content string, delimiter rune) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for _, r := range content {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
			continue
		}

		if r == '\\' {
			escapeNext = true
			current.WriteRune(r)
			continue
		}

		if r == '"' {
			inQuotes = !inQuotes
			current.WriteRune(r)
			continue
		}

		if r == delimiter && !inQuotes {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseKeyValue parses a key-value pair
func (p *ToolCallParser) parseKeyValue(pair string) (string, interface{}) {
	colonPos := strings.Index(pair, ":")
	if colonPos == -1 {
		return "", nil
	}

	key := strings.TrimSpace(pair[:colonPos])
	value := strings.TrimSpace(pair[colonPos+1:])

	// Remove quotes from key
	key = strings.Trim(key, "\"'")

	// Parse value
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		// String value
		value = strings.Trim(value, "\"")
		// Unescape common escape sequences
		value = strings.ReplaceAll(value, "\\n", "\n")
		value = strings.ReplaceAll(value, "\\t", "\t")
		value = strings.ReplaceAll(value, "\\\"", "\"")
		return key, value
	}

	// Try to parse as number or boolean
	if value == "true" {
		return key, true
	}
	if value == "false" {
		return key, false
	}
	if value == "null" {
		return key, nil
	}

	// Try as number
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		if float64(int(num)) == num {
			return key, int(num)
		}
		return key, num
	}

	// Default to string
	return key, value
}

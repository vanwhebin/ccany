package models

import (
	"encoding/json"
)

// ClaudeMessage represents a message in Claude format
type ClaudeMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// ClaudeMessagesRequest represents a Claude messages API request
type ClaudeMessagesRequest struct {
	Model         string          `json:"model"`
	MaxTokens     int             `json:"max_tokens"`
	Messages      []ClaudeMessage `json:"messages"`
	System        interface{}     `json:"system,omitempty"`
	Temperature   *float64        `json:"temperature,omitempty"`
	TopP          *float64        `json:"top_p,omitempty"`
	TopK          *int            `json:"top_k,omitempty"`
	Stream        bool            `json:"stream,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
	Tools         []ClaudeTool    `json:"tools,omitempty"`
	ToolChoice    interface{}     `json:"tool_choice,omitempty"`
	Thinking      bool            `json:"thinking,omitempty"` // Support for Claude Code thinking mode
	Metadata      interface{}     `json:"metadata,omitempty"` // Additional metadata for Claude Code
}

// ClaudeTool represents a tool in Claude format
type ClaudeTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ClaudeTokenCountRequest represents a token count request
type ClaudeTokenCountRequest struct {
	Model    string          `json:"model"`
	Messages []ClaudeMessage `json:"messages"`
	System   interface{}     `json:"system,omitempty"`
}

// ClaudeResponse represents a Claude API response
type ClaudeResponse struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Role         string               `json:"role"`
	Content      []ClaudeContentBlock `json:"content"`
	Model        string               `json:"model"`
	StopReason   string               `json:"stop_reason"`
	StopSequence *string              `json:"stop_sequence"`
	Usage        ClaudeUsage          `json:"usage"`
}

// ClaudeContentBlock represents a content block in Claude response
type ClaudeContentBlock struct {
	Type  string      `json:"type"`
	Text  string      `json:"text,omitempty"`
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Input interface{} `json:"input,omitempty"`
}

// ClaudeUsage represents usage information
type ClaudeUsage struct {
	InputTokens          int `json:"input_tokens"`
	OutputTokens         int `json:"output_tokens"`
	CacheReadInputTokens int `json:"cache_read_input_tokens,omitempty"` // Support for cache token usage
}

// ClaudeStreamEvent represents a streaming event
type ClaudeStreamEvent struct {
	Type         string              `json:"type"`
	Message      *ClaudeResponse     `json:"message,omitempty"`
	Index        int                 `json:"index,omitempty"`
	Delta        *ClaudeContentBlock `json:"delta,omitempty"`
	Usage        *ClaudeUsage        `json:"usage,omitempty"`
	ContentBlock *ClaudeContentBlock `json:"content_block,omitempty"` // For content_block_start event
}

// ClaudeStreamStartEvent represents the message_start event
type ClaudeStreamStartEvent struct {
	Type    string         `json:"type"`
	Message ClaudeResponse `json:"message"`
}

// ClaudeStreamContentBlockStartEvent represents the content_block_start event
type ClaudeStreamContentBlockStartEvent struct {
	Type         string             `json:"type"`
	Index        int                `json:"index"`
	ContentBlock ClaudeContentBlock `json:"content_block"`
}

// ClaudeStreamContentBlockStopEvent represents the content_block_stop event
type ClaudeStreamContentBlockStopEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// ClaudeStreamPingEvent represents the ping event
type ClaudeStreamPingEvent struct {
	Type string `json:"type"`
}

// ClaudeStreamMessageDeltaEvent represents the message_delta event
type ClaudeStreamMessageDeltaEvent struct {
	Type  string             `json:"type"`
	Delta ClaudeMessageDelta `json:"delta"`
	Usage *ClaudeUsage       `json:"usage,omitempty"`
}

// ClaudeStreamMessageStopEvent represents the message_stop event
type ClaudeStreamMessageStopEvent struct {
	Type string `json:"type"`
}

// ClaudeMessageDelta represents delta information for message updates
type ClaudeMessageDelta struct {
	StopReason   string  `json:"stop_reason,omitempty"`
	StopSequence *string `json:"stop_sequence,omitempty"`
}

// ClaudeContentBlockDelta represents delta information for content updates
type ClaudeContentBlockDelta struct {
	Type        string      `json:"type"`
	Text        string      `json:"text,omitempty"`
	PartialJSON string      `json:"partial_json,omitempty"` // For tool call streaming
	InputJSON   interface{} `json:"input_json,omitempty"`   // For tool call completion
}

// ClaudeErrorResponse represents an error response
type ClaudeErrorResponse struct {
	Type  string      `json:"type"`
	Error ClaudeError `json:"error"`
}

// ClaudeError represents an error
type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ToJSON converts the request to JSON
func (r *ClaudeMessagesRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON creates a request from JSON
func (r *ClaudeMessagesRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

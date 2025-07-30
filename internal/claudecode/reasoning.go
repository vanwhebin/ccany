package claudecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ReasoningProcessor handles reasoning/thinking content in Claude responses
type ReasoningProcessor struct {
	logger *logrus.Logger
}

// NewReasoningProcessor creates a new reasoning processor
func NewReasoningProcessor(logger *logrus.Logger) *ReasoningProcessor {
	return &ReasoningProcessor{
		logger: logger,
	}
}

// ReasoningContext tracks reasoning content during streaming
type ReasoningContext struct {
	ReasoningContent    *bytes.Buffer
	IsReasoningComplete bool
	ThinkingBlockSent   bool
	ThinkingSignature   string
}

// NewReasoningContext creates a new reasoning context
func NewReasoningContext() *ReasoningContext {
	return &ReasoningContext{
		ReasoningContent:    &bytes.Buffer{},
		IsReasoningComplete: false,
		ThinkingBlockSent:   false,
	}
}

// ProcessReasoningDelta processes reasoning content from OpenAI streaming
func (p *ReasoningProcessor) ProcessReasoningDelta(c *gin.Context, streamCtx *StreamingContext, reasoningCtx *ReasoningContext, delta map[string]interface{}) bool {
	// Check for reasoning_content in delta
	if reasoningContent, exists := delta["reasoning_content"]; exists && reasoningContent != nil {
		if content, ok := reasoningContent.(string); ok && content != "" {
			// Accumulate reasoning content
			reasoningCtx.ReasoningContent.WriteString(content)

			p.logger.WithFields(logrus.Fields{
				"content_length": len(content),
				"total_length":   reasoningCtx.ReasoningContent.Len(),
			}).Debug("Accumulating reasoning content")

			// Send thinking delta event
			thinkingDelta := map[string]interface{}{
				"type":  "content_block_delta",
				"index": 0, // Thinking block is always at index 0
				"delta": map[string]interface{}{
					"type": "thinking_delta",
					"thinking": map[string]interface{}{
						"content": content,
					},
				},
			}
			p.writeSSEEvent(c, "content_block_delta", thinkingDelta)

			// Remove reasoning_content from original delta to prevent duplication
			delete(delta, "reasoning_content")
			return true
		}
	}

	// Check if we need to complete reasoning block
	if !reasoningCtx.IsReasoningComplete && reasoningCtx.ReasoningContent.Len() > 0 {
		// Check if we have actual content or tool calls starting
		hasContent := false
		hasToolCalls := false

		if content, exists := delta["content"]; exists && content != nil {
			if str, ok := content.(string); ok && str != "" {
				hasContent = true
			}
		}

		if toolCalls, exists := delta["tool_calls"]; exists && toolCalls != nil {
			if tc, ok := toolCalls.([]interface{}); ok && len(tc) > 0 {
				hasToolCalls = true
			}
		}

		// Complete reasoning block when actual content starts
		if hasContent || hasToolCalls {
			reasoningCtx.IsReasoningComplete = true
			reasoningCtx.ThinkingSignature = fmt.Sprintf("%d", time.Now().UnixNano())

			p.logger.WithFields(logrus.Fields{
				"reasoning_length": reasoningCtx.ReasoningContent.Len(),
				"signature":        reasoningCtx.ThinkingSignature,
			}).Info("Completing reasoning block")

			// Send complete thinking block
			thinkingComplete := map[string]interface{}{
				"type":  "content_block_stop",
				"index": 0,
			}
			p.writeSSEEvent(c, "content_block_stop", thinkingComplete)

			// Adjust indices for subsequent content
			streamCtx.mu.Lock()
			streamCtx.TextBlockIndex = 1 // Move text to index 1 after thinking
			streamCtx.mu.Unlock()

			// Send new content_block_start for text content
			if hasContent {
				contentBlockStart := map[string]interface{}{
					"type":  "content_block_start",
					"index": 1,
					"content_block": map[string]interface{}{
						"type": "text",
						"text": "",
					},
				}
				p.writeSSEEvent(c, "content_block_start", contentBlockStart)
			}
		}
	}

	return false
}

// InitializeThinkingBlock initializes a thinking block at the start of streaming
func (p *ReasoningProcessor) InitializeThinkingBlock(c *gin.Context, streamCtx *StreamingContext, enableThinking bool) *ReasoningContext {
	reasoningCtx := NewReasoningContext()

	if !enableThinking {
		return reasoningCtx
	}

	p.logger.Info("Initializing thinking block for Claude response")

	// Send content_block_start for thinking block
	thinkingBlockStart := map[string]interface{}{
		"type":  "content_block_start",
		"index": 0,
		"content_block": map[string]interface{}{
			"type": "thinking",
			"thinking": map[string]interface{}{
				"content": "",
			},
		},
	}
	p.writeSSEEvent(c, "content_block_start", thinkingBlockStart)

	return reasoningCtx
}

// FinalizeThinkingBlock finalizes any pending thinking content
func (p *ReasoningProcessor) FinalizeThinkingBlock(c *gin.Context, reasoningCtx *ReasoningContext) {
	if reasoningCtx.ReasoningContent.Len() > 0 && !reasoningCtx.IsReasoningComplete {
		// Send final thinking content
		p.logger.WithField("content_length", reasoningCtx.ReasoningContent.Len()).Info("Finalizing thinking block")

		thinkingComplete := map[string]interface{}{
			"type":  "content_block_stop",
			"index": 0,
		}
		p.writeSSEEvent(c, "content_block_stop", thinkingComplete)
	}
}

// writeSSEEvent writes a Server-Sent Event
func (p *ReasoningProcessor) writeSSEEvent(c *gin.Context, event string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		p.logger.WithError(err).Error("Failed to marshal SSE event data")
		return
	}

	if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, string(jsonData)); err != nil {
		p.logger.WithError(err).Debug("Failed to write SSE event - client may have disconnected")
	}
}

// ExtractReasoningFromOpenAI extracts reasoning content from OpenAI response
func (p *ReasoningProcessor) ExtractReasoningFromOpenAI(openAIResponse interface{}) (string, interface{}) {
	// Handle different response types
	switch resp := openAIResponse.(type) {
	case map[string]interface{}:
		return p.extractReasoningFromMap(resp)
	case *map[string]interface{}:
		return p.extractReasoningFromMap(*resp)
	default:
		return "", openAIResponse
	}
}

// extractReasoningFromMap extracts reasoning content from a map response
func (p *ReasoningProcessor) extractReasoningFromMap(respMap map[string]interface{}) (string, interface{}) {
	reasoningContent := ""

	// Check for reasoning in choices
	if choices, ok := respMap["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			// Extract reasoning from message
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if reasoning, exists := message["reasoning"]; exists && reasoning != nil {
					if reasoningStr, ok := reasoning.(string); ok {
						reasoningContent = reasoningStr
						// Remove reasoning from message
						delete(message, "reasoning")
					}
				}
			}

			// Extract reasoning from delta (streaming)
			if delta, ok := choice["delta"].(map[string]interface{}); ok {
				if reasoning, exists := delta["reasoning_content"]; exists && reasoning != nil {
					if reasoningStr, ok := reasoning.(string); ok {
						reasoningContent = reasoningStr
						// Remove reasoning from delta
						delete(delta, "reasoning_content")
					}
				}
			}
		}
	}

	return reasoningContent, respMap
}

// CreateThinkingResponse creates a thinking-aware response structure
func (p *ReasoningProcessor) CreateThinkingResponse(originalResponse interface{}, thinkingContent string) interface{} {
	if thinkingContent == "" {
		return originalResponse
	}

	// Create enhanced response with thinking
	signature := uuid.New().String()

	switch resp := originalResponse.(type) {
	case map[string]interface{}:
		// Add thinking to content array
		if content, ok := resp["content"].([]interface{}); ok {
			// Prepend thinking block
			thinkingBlock := map[string]interface{}{
				"type": "thinking",
				"thinking": map[string]interface{}{
					"content":   thinkingContent,
					"signature": signature,
				},
			}
			resp["content"] = append([]interface{}{thinkingBlock}, content...)
		}
		return resp
	default:
		return originalResponse
	}
}

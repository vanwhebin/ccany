package claudecode

import (
	"strings"

	"ccany/internal/models"

	"github.com/sirupsen/logrus"
)

// ModelRouter handles intelligent model routing based on request characteristics
type ModelRouter struct {
	logger           *logrus.Logger
	bigModel         string
	smallModel       string
	reasoningModel   string
	longContextModel string
}

// NewModelRouter creates a new model router
func NewModelRouter(logger *logrus.Logger, bigModel, smallModel string) *ModelRouter {
	return &ModelRouter{
		logger:           logger,
		bigModel:         bigModel,
		smallModel:       smallModel,
		reasoningModel:   "claude-3-5-sonnet-20241022", // Default reasoning model
		longContextModel: "claude-3-5-sonnet-20241022", // Default long context model
	}
}

// RouteModel determines the appropriate model based on request characteristics
func (r *ModelRouter) RouteModel(req *models.ClaudeMessagesRequest) string {
	// Check for thinking mode
	if req.Thinking {
		r.logger.WithFields(logrus.Fields{
			"original_model": req.Model,
			"routed_model":   r.reasoningModel,
			"reason":         "thinking_mode",
		}).Info("Routing to reasoning model for thinking mode")
		return r.reasoningModel
	}

	// Check for model-specific routing
	switch req.Model {
	case "claude-3-5-haiku-20241022":
		// Route haiku to small model for background tasks
		r.logger.WithFields(logrus.Fields{
			"original_model": req.Model,
			"routed_model":   r.smallModel,
			"reason":         "background_model",
		}).Info("Routing haiku to background model")
		return r.smallModel

	case "claude-3-5-sonnet-20241022":
		// Check token count for long context routing
		tokenCount := r.estimateTokenCount(req)
		if tokenCount > 60000 {
			r.logger.WithFields(logrus.Fields{
				"original_model":   req.Model,
				"routed_model":     r.longContextModel,
				"estimated_tokens": tokenCount,
				"reason":           "long_context",
			}).Info("Routing to long context model")
			return r.longContextModel
		}
		return r.bigModel

	default:
		// Default routing based on complexity
		if r.isComplexRequest(req) {
			r.logger.WithFields(logrus.Fields{
				"original_model": req.Model,
				"routed_model":   r.bigModel,
				"reason":         "complex_request",
			}).Info("Routing to big model for complex request")
			return r.bigModel
		}
		return r.smallModel
	}
}

// ParseModelCommand parses model command from message content
func (r *ModelRouter) ParseModelCommand(content string) (provider, model string, hasCommand bool) {
	// Check for /model provider,model command
	if strings.HasPrefix(content, "/model ") {
		commandPart := strings.TrimPrefix(content, "/model ")
		parts := strings.Split(commandPart, ",")
		if len(parts) == 2 {
			provider = strings.TrimSpace(parts[0])
			model = strings.TrimSpace(parts[1])
			hasCommand = true

			r.logger.WithFields(logrus.Fields{
				"provider": provider,
				"model":    model,
			}).Info("Parsed model command")
		}
	}

	return provider, model, hasCommand
}

// EstimateTokenCount estimates the token count for a request (public method)
func (r *ModelRouter) EstimateTokenCount(req *models.ClaudeMessagesRequest) int {
	return r.estimateTokenCount(req)
}

// estimateTokenCount estimates the token count for a request (private method)
func (r *ModelRouter) estimateTokenCount(req *models.ClaudeMessagesRequest) int {
	totalChars := 0

	// Count system message characters
	if req.System != nil {
		if systemStr, ok := req.System.(string); ok {
			totalChars += len(systemStr)
		}
	}

	// Count message characters
	for _, msg := range req.Messages {
		if msg.Content != nil {
			totalChars += r.countContentChars(msg.Content)
		}
	}

	// Count tool definitions
	for _, tool := range req.Tools {
		totalChars += len(tool.Name) + len(tool.Description)
		// Estimate schema size
		totalChars += 200 // Average schema size
	}

	// Rough estimation: 4 characters per token
	return totalChars / 4
}

// countContentChars counts characters in content (handles both string and block formats)
func (r *ModelRouter) countContentChars(content interface{}) int {
	switch c := content.(type) {
	case string:
		return len(c)
	case []interface{}:
		totalChars := 0
		for _, block := range c {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if text, exists := blockMap["text"]; exists {
					if textStr, ok := text.(string); ok {
						totalChars += len(textStr)
					}
				}
				// Handle image blocks (estimate size)
				if blockType, exists := blockMap["type"]; exists {
					if blockType == "image" {
						totalChars += 1000 // Estimated image token cost
					}
				}
			}
		}
		return totalChars
	default:
		return 0
	}
}

// isComplexRequest determines if a request is complex based on various factors
func (r *ModelRouter) isComplexRequest(req *models.ClaudeMessagesRequest) bool {
	// Check for tools
	if len(req.Tools) > 0 {
		return true
	}

	// Check for high max tokens
	if req.MaxTokens > 4000 {
		return true
	}

	// Check for multiple messages (conversation)
	if len(req.Messages) > 5 {
		return true
	}

	// Check for complex content blocks
	for _, msg := range req.Messages {
		if r.hasComplexContent(msg.Content) {
			return true
		}
	}

	return false
}

// hasComplexContent checks if content contains complex elements
func (r *ModelRouter) hasComplexContent(content interface{}) bool {
	switch c := content.(type) {
	case []interface{}:
		for _, block := range c {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockType, exists := blockMap["type"]; exists {
					// Complex content types
					if blockType == "image" || blockType == "tool_use" || blockType == "tool_result" {
						return true
					}
				}
				// Check for long text blocks
				if text, exists := blockMap["text"]; exists {
					if textStr, ok := text.(string); ok {
						if len(textStr) > 2000 {
							return true
						}
					}
				}
			}
		}
	case string:
		// Check for long text content
		if len(c) > 2000 {
			return true
		}
	}
	return false
}

// GetModelCapabilities returns capabilities for different models
func (r *ModelRouter) GetModelCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"models": map[string]interface{}{
			r.bigModel: map[string]interface{}{
				"max_tokens":     8192,
				"supports_tools": true,
				"supports_image": true,
				"context_window": 200000,
			},
			r.smallModel: map[string]interface{}{
				"max_tokens":     4096,
				"supports_tools": true,
				"supports_image": true,
				"context_window": 100000,
			},
			r.reasoningModel: map[string]interface{}{
				"max_tokens":        8192,
				"supports_tools":    true,
				"supports_image":    true,
				"context_window":    200000,
				"supports_thinking": true,
			},
		},
		"routing_rules": []string{
			"thinking_mode -> reasoning_model",
			"token_count > 60K -> long_context_model",
			"haiku_model -> background_model",
			"complex_request -> big_model",
			"default -> small_model",
		},
	}
}

// SetReasoningModel sets the reasoning model for thinking mode
func (r *ModelRouter) SetReasoningModel(model string) {
	r.reasoningModel = model
	r.logger.WithField("model", model).Info("Updated reasoning model")
}

// SetLongContextModel sets the long context model for high token count requests
func (r *ModelRouter) SetLongContextModel(model string) {
	r.longContextModel = model
	r.logger.WithField("model", model).Info("Updated long context model")
}

// LogRoutingDecision logs the model routing decision
func (r *ModelRouter) LogRoutingDecision(originalModel, routedModel, reason string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"original_model": originalModel,
		"routed_model":   routedModel,
		"reason":         reason,
	}

	for k, v := range metadata {
		fields[k] = v
	}

	r.logger.WithFields(fields).Info("Model routing decision")
}

// UpdateModelConfiguration updates the model configuration dynamically
func (r *ModelRouter) UpdateModelConfiguration(bigModel, smallModel string) {
	r.bigModel = bigModel
	r.smallModel = smallModel
	r.logger.WithFields(logrus.Fields{
		"big_model":   bigModel,
		"small_model": smallModel,
	}).Info("Updated model router configuration")
}

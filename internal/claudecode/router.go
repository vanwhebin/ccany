package claudecode

import (
	"strings"
	"sync"

	"ccany/internal/models"

	"github.com/sirupsen/logrus"
)

// ModelRouter handles intelligent model routing based on request characteristics
type ModelRouter struct {
	mu               sync.RWMutex // Protects model configuration fields
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
// Follows claude-code-router getUseModel logic exactly
func (r *ModelRouter) RouteModel(req *models.ClaudeMessagesRequest) string {
	r.mu.RLock()
	bigModel := r.bigModel
	smallModel := r.smallModel
	reasoningModel := r.reasoningModel
	longContextModel := r.longContextModel
	r.mu.RUnlock()

	// 1. Handle comma-separated models (direct passthrough) - First priority
	if strings.Contains(req.Model, ",") {
		r.logger.WithFields(logrus.Fields{
			"original_model": req.Model,
			"reason":         "comma_separated_models",
		}).Info("Using comma-separated model specification")
		return req.Model
	}

	// Calculate token count for routing decisions
	tokenCount := r.estimateTokenCount(req)

	// 2. if tokenCount is greater than 60K, use the long context model - Second priority
	if tokenCount > 1000*60 && longContextModel != "" {
		r.logger.WithFields(logrus.Fields{
			"original_model":   req.Model,
			"routed_model":     longContextModel,
			"estimated_tokens": tokenCount,
			"reason":           "long_context",
		}).Info("Using long context model due to token count")
		return longContextModel
	}

	// 3. If the model is claude-3-5-haiku, use the background model - Third priority
	if strings.HasPrefix(req.Model, "claude-3-5-haiku") && smallModel != "" {
		r.logger.WithFields(logrus.Fields{
			"original_model": req.Model,
			"routed_model":   smallModel,
			"reason":         "background_model",
		}).Info("Using background model for haiku")
		return smallModel
	}

	// 4. if exits thinking, use the think model - Fourth priority
	if req.Thinking && reasoningModel != "" {
		r.logger.WithFields(logrus.Fields{
			"original_model": req.Model,
			"routed_model":   reasoningModel,
			"reason":         "thinking_mode",
		}).Info("Using think model for thinking mode")
		return reasoningModel
	}

	// 5. Default model - Final fallback (matches claude-code-router)
	r.logger.WithFields(logrus.Fields{
		"original_model":   req.Model,
		"routed_model":     bigModel,
		"estimated_tokens": tokenCount,
		"reason":           "default",
	}).Info("Using default model")
	return bigModel
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

// estimateTokenCount estimates the token count for a request using tiktoken-compatible logic
func (r *ModelRouter) estimateTokenCount(req *models.ClaudeMessagesRequest) int {
	totalTokens := 0

	// Count system message tokens
	if req.System != nil {
		if systemStr, ok := req.System.(string); ok {
			totalTokens += r.estimateTextTokens(systemStr)
		} else if systemArray, ok := req.System.([]interface{}); ok {
			for _, item := range systemArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if itemType, exists := itemMap["type"]; exists && itemType == "text" {
						if text, exists := itemMap["text"]; exists {
							if textStr, ok := text.(string); ok {
								totalTokens += r.estimateTextTokens(textStr)
							} else if textArray, ok := text.([]interface{}); ok {
								for _, textPart := range textArray {
									if textPartStr, ok := textPart.(string); ok {
										totalTokens += r.estimateTextTokens(textPartStr)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Count message tokens (similar to claude-code-router logic)
	for _, msg := range req.Messages {
		if msg.Content != nil {
			totalTokens += r.countContentTokens(msg.Content)
		}
	}

	// Count tool definition tokens (similar to claude-code-router)
	for _, tool := range req.Tools {
		totalTokens += r.estimateTextTokens(tool.Name + tool.Description)
		// Estimate input_schema tokens
		if tool.InputSchema != nil {
			// Serialize to JSON and estimate tokens
			totalTokens += len(tool.Name) / 4 // Rough schema size estimation
			totalTokens += 50                 // Base overhead for schema structure
		}
	}

	return totalTokens
}

// countContentTokens counts tokens in content using tiktoken-compatible logic
func (r *ModelRouter) countContentTokens(content interface{}) int {
	switch c := content.(type) {
	case string:
		return r.estimateTextTokens(c)
	case []interface{}:
		totalTokens := 0
		for _, block := range c {
			if blockMap, ok := block.(map[string]interface{}); ok {
				if blockType, exists := blockMap["type"]; exists {
					switch blockType {
					case "text":
						if text, exists := blockMap["text"]; exists {
							if textStr, ok := text.(string); ok {
								totalTokens += r.estimateTextTokens(textStr)
							}
						}
					case "tool_use":
						// Similar to claude-code-router: count tool use input
						if input, exists := blockMap["input"]; exists {
							// Serialize and count tokens
							totalTokens += 50 // Base overhead
							if inputMap, ok := input.(map[string]interface{}); ok {
								for k, v := range inputMap {
									totalTokens += r.estimateTextTokens(k)
									if vStr, ok := v.(string); ok {
										totalTokens += r.estimateTextTokens(vStr)
									} else {
										totalTokens += 10 // Estimate for non-string values
									}
								}
							}
						}
					case "tool_result":
						// Similar to claude-code-router: count tool result content
						if content, exists := blockMap["content"]; exists {
							if contentStr, ok := content.(string); ok {
								totalTokens += r.estimateTextTokens(contentStr)
							} else {
								// Handle complex content structures
								totalTokens += 100 // Base estimate for complex content
							}
						}
					case "image":
						// Estimated image token cost (like claude-code-router)
						totalTokens += 1000
					}
				}
			}
		}
		return totalTokens
	default:
		return 0
	}
}

// estimateTextTokens provides better token estimation than simple character count
func (r *ModelRouter) estimateTextTokens(text string) int {
	if text == "" {
		return 0
	}

	// More accurate token estimation based on tiktoken patterns
	// Account for punctuation, spaces, and word boundaries
	chars := len(text)
	words := len(strings.Fields(text))

	// Improved estimation formula based on tiktoken behavior:
	// - Average 3.3 chars per token for English text
	// - Punctuation and special characters affect token count
	// - Word boundaries and spaces are important

	baseTokens := chars / 3 // More aggressive than 4 chars per token
	wordBonus := words / 4  // Account for word boundary effects

	totalTokens := baseTokens + wordBonus
	if totalTokens < 1 && chars > 0 {
		totalTokens = 1 // Minimum one token for non-empty text
	}

	return totalTokens
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
	r.mu.RLock()
	bigModel := r.bigModel
	smallModel := r.smallModel
	reasoningModel := r.reasoningModel
	r.mu.RUnlock()

	return map[string]interface{}{
		"models": map[string]interface{}{
			bigModel: map[string]interface{}{
				"max_tokens":     8192,
				"supports_tools": true,
				"supports_image": true,
				"context_window": 200000,
			},
			smallModel: map[string]interface{}{
				"max_tokens":     4096,
				"supports_tools": true,
				"supports_image": true,
				"context_window": 100000,
			},
			reasoningModel: map[string]interface{}{
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
	r.mu.Lock()
	r.reasoningModel = model
	r.mu.Unlock()
	r.logger.WithField("model", model).Info("Updated reasoning model")
}

// SetLongContextModel sets the long context model for high token count requests
func (r *ModelRouter) SetLongContextModel(model string) {
	r.mu.Lock()
	r.longContextModel = model
	r.mu.Unlock()
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
	r.mu.Lock()
	r.bigModel = bigModel
	r.smallModel = smallModel
	r.mu.Unlock()
	r.logger.WithFields(logrus.Fields{
		"big_model":   bigModel,
		"small_model": smallModel,
	}).Info("Updated model router configuration")
}

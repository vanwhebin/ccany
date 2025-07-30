package claudecode

import (
	"strings"
	"sync"

	"ccany/internal/models"
	"ccany/internal/tokenizer"

	"github.com/sirupsen/logrus"
)

// ModelRoutingStrategy defines the interface for model routing strategies
type ModelRoutingStrategy interface {
	ShouldApply(req *models.ClaudeMessagesRequest, tokenCount int) bool
	GetModel() string
	GetReason() string
}

// CommaSeperatedStrategy handles comma-separated model lists
type CommaSeperatedStrategy struct {
	originalModel string
}

func (s *CommaSeperatedStrategy) ShouldApply(req *models.ClaudeMessagesRequest, tokenCount int) bool {
	return strings.Contains(req.Model, ",")
}

func (s *CommaSeperatedStrategy) GetModel() string {
	return s.originalModel
}

func (s *CommaSeperatedStrategy) GetReason() string {
	return "comma_separated_models"
}

// LongContextStrategy handles requests with high token count
type LongContextStrategy struct {
	model     string
	threshold int
}

func (s *LongContextStrategy) ShouldApply(req *models.ClaudeMessagesRequest, tokenCount int) bool {
	return tokenCount > s.threshold && s.model != ""
}

func (s *LongContextStrategy) GetModel() string {
	return s.model
}

func (s *LongContextStrategy) GetReason() string {
	return "long_context"
}

// BackgroundModelStrategy handles haiku model requests
type BackgroundModelStrategy struct {
	model string
}

func (s *BackgroundModelStrategy) ShouldApply(req *models.ClaudeMessagesRequest, tokenCount int) bool {
	return strings.HasPrefix(req.Model, "claude-3-5-haiku") && s.model != ""
}

func (s *BackgroundModelStrategy) GetModel() string {
	return s.model
}

func (s *BackgroundModelStrategy) GetReason() string {
	return "background_model"
}

// ThinkingModeStrategy handles thinking mode requests
type ThinkingModeStrategy struct {
	model string
}

func (s *ThinkingModeStrategy) ShouldApply(req *models.ClaudeMessagesRequest, tokenCount int) bool {
	return req.Thinking && s.model != ""
}

func (s *ThinkingModeStrategy) GetModel() string {
	return s.model
}

func (s *ThinkingModeStrategy) GetReason() string {
	return "thinking_mode"
}

// WebSearchStrategy handles requests with web search tools
type WebSearchStrategy struct {
	model   string
	enabled bool
}

func (s *WebSearchStrategy) ShouldApply(req *models.ClaudeMessagesRequest, tokenCount int) bool {
	if !s.enabled || s.model == "" {
		return false
	}

	// Check if any tool is web_search type
	for _, tool := range req.Tools {
		if tool.Name == "web_search" {
			return true
		}
		// Check in input schema for type field
		if typeVal, exists := tool.InputSchema["type"]; exists {
			if typeStr, ok := typeVal.(string); ok && typeStr == "web_search" {
				return true
			}
		}
	}
	return false
}

func (s *WebSearchStrategy) ShouldApplyToRequest(req *RouteRequest) bool {
	if !s.enabled || s.model == "" {
		return false
	}

	// Check if any tool has web_search type or name
	for _, tool := range req.Tools {
		if toolMap, ok := tool.(map[string]interface{}); ok {
			// Check name
			if name, exists := toolMap["name"]; exists {
				if nameStr, ok := name.(string); ok && nameStr == "web_search" {
					return true
				}
			}
			// Check type field if exists
			if toolType, exists := toolMap["type"]; exists {
				if typeStr, ok := toolType.(string); ok && typeStr == "web_search" {
					return true
				}
			}
		}
	}
	return false
}

func (s *WebSearchStrategy) GetModel() string {
	return s.model
}

func (s *WebSearchStrategy) GetReason() string {
	return "web_search_tool"
}

// ModelRouter handles intelligent model routing based on request characteristics
type ModelRouter struct {
	mu               sync.RWMutex // Protects model configuration fields
	logger           *logrus.Logger
	bigModel         string
	smallModel       string
	reasoningModel   string
	longContextModel string
	strategies       []ModelRoutingStrategy  // Ordered list of strategies
	tokenCounter     *tokenizer.TokenCounter // Token counter for accurate counting
	routerLogger     *RouterLogger           // Enhanced router logger
}

// NewModelRouter creates a new model router
func NewModelRouter(logger *logrus.Logger, bigModel, smallModel string) *ModelRouter {
	// Create enhanced logger
	routerLogger, err := NewLoggerFromEnv()
	if err != nil {
		logger.WithError(err).Warn("Failed to create router logger, using default")
	}

	router := &ModelRouter{
		logger:           logger,
		bigModel:         bigModel,
		smallModel:       smallModel,
		reasoningModel:   "claude-3-5-sonnet-20241022", // Default reasoning model
		longContextModel: "claude-3-5-sonnet-20241022", // Default long context model
		tokenCounter:     tokenizer.NewTokenCounter(logger),
		routerLogger:     routerLogger,
	}

	// Initialize strategies in priority order
	router.initializeStrategies()
	return router
}

// initializeStrategies sets up the routing strategies in priority order
func (r *ModelRouter) initializeStrategies() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Use default configuration
	config := DefaultRouterConfig()

	r.strategies = []ModelRoutingStrategy{
		// Priority 1: Comma-separated models (direct passthrough)
		&CommaSeperatedStrategy{},

		// Priority 2: Tool use detection (highest priority for functional requirements)
		NewToolUseStrategy(r.bigModel, config.EnableToolUseDetection, r.logger),

		// Priority 3: Long context handling
		&LongContextStrategy{
			model:     config.LongContext,
			threshold: config.LongContextThreshold,
		},

		// Priority 4: Web search detection
		&WebSearchStrategy{
			model:   config.WebSearch,
			enabled: config.EnableWebSearchDetection,
		},

		// Priority 5: Background model for haiku
		&BackgroundModelStrategy{
			model: r.smallModel, // Use smallModel for background tasks
		},

		// Priority 6: Thinking mode
		&ThinkingModeStrategy{
			model: config.Think,
		},
	}

	// Set default models if not already set
	if r.reasoningModel == "" {
		r.reasoningModel = config.Think
	}
	if r.longContextModel == "" {
		r.longContextModel = config.LongContext
	}
}

// RouteModel determines the appropriate model based on request characteristics
// Follows claude-code-router getUseModel logic exactly
func (r *ModelRouter) RouteModel(req *models.ClaudeMessagesRequest) string {
	// Calculate token count for routing decisions
	tokenCount := r.estimateTokenCount(req)

	// Apply strategies in order
	r.mu.RLock()
	strategies := r.strategies
	defaultModel := r.bigModel // Use bigModel as default
	r.mu.RUnlock()

	// Special handling for comma-separated models
	if strings.Contains(req.Model, ",") {
		r.logger.WithFields(logrus.Fields{
			"original_model": req.Model,
			"reason":         "comma_separated_models",
		}).Info("Using comma-separated model specification")
		return req.Model
	}

	// Convert request for strategy evaluation
	routeReq := &RouteRequest{
		Model:      req.Model,
		TokenCount: tokenCount,
		Tools:      make([]interface{}, len(req.Tools)),
	}
	for i, tool := range req.Tools {
		routeReq.Tools[i] = map[string]interface{}{
			"name":         tool.Name,
			"description":  tool.Description,
			"input_schema": tool.InputSchema,
		}
	}

	// Apply each strategy in priority order
	for _, strategy := range strategies {
		shouldApply := false

		// Handle different strategy types
		switch s := strategy.(type) {
		case *WebSearchStrategy:
			shouldApply = s.ShouldApplyToRequest(routeReq)
		default:
			shouldApply = strategy.ShouldApply(req, tokenCount)
		}

		if shouldApply {
			model := strategy.GetModel()
			reason := strategy.GetReason()

			// Log routing decision
			if r.routerLogger != nil && r.routerLogger.IsEnabled() {
				decision := &RoutingDecision{
					OriginalModel: req.Model,
					RoutedModel:   model,
					Reason:        reason,
					TokenCount:    tokenCount,
					HasTools:      len(req.Tools) > 0,
					HasThinking:   req.Thinking,
					MessageCount:  len(req.Messages),
					Metadata: map[string]interface{}{
						"max_tokens": req.MaxTokens,
					},
				}
				r.routerLogger.LogRoutingDecision(decision)
			} else {
				r.logger.WithFields(logrus.Fields{
					"original_model":   req.Model,
					"routed_model":     model,
					"estimated_tokens": tokenCount,
					"reason":           reason,
				}).Info("Model routing decision")
			}

			return model
		}
	}

	// Default model - Final fallback
	r.logger.WithFields(logrus.Fields{
		"original_model":   req.Model,
		"routed_model":     defaultModel,
		"estimated_tokens": tokenCount,
		"reason":           "default",
	}).Info("Using default model")
	return defaultModel
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
	// Use tokenizer if available, fallback to simple estimation
	if r.tokenCounter != nil {
		// Convert messages to interface slice for token counter
		messages := make([]interface{}, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			}
		}

		tokens, err := r.tokenCounter.CountClaudeMessagesTokens(messages, req.System, req.Model)
		if err == nil {
			// Add tool tokens if present
			for _, tool := range req.Tools {
				toolTokens, _ := r.tokenCounter.CountTokens(tool.Name+" "+tool.Description, req.Model)
				tokens += toolTokens
				if tool.InputSchema != nil {
					// Add schema overhead
					tokens += 50
				}
			}
			return tokens
		}
		// If error, fall back to simple estimation
		r.logger.WithError(err).Warn("Failed to count tokens with tokenizer, using fallback")
	}

	// Fallback: simple estimation
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
	// Try to use tokenizer if available
	if r.tokenCounter != nil {
		model := r.bigModel // Default model for counting
		tokens, err := r.tokenCounter.CountContentTokens(content, model)
		if err == nil {
			return tokens
		}
	}

	// Fallback to simple estimation
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

	// Use strings.Builder for efficient analysis
	var wordCount, punctCount int
	inWord := false

	for _, char := range text {
		switch {
		case char == ' ' || char == '\t' || char == '\n' || char == '\r':
			inWord = false
		case char == '.' || char == ',' || char == '!' || char == '?' ||
			char == ';' || char == ':' || char == '"' || char == '\'' ||
			char == '(' || char == ')' || char == '[' || char == ']' ||
			char == '{' || char == '}':
			punctCount++
			inWord = false
		default:
			if !inWord {
				wordCount++
				inWord = true
			}
		}
	}

	// Improved estimation formula based on tiktoken behavior:
	// - Average 3.3 chars per token for English text
	// - Punctuation affects token count
	// - Word boundaries are important
	chars := len(text)
	baseTokens := chars / 3
	wordBonus := wordCount / 4
	punctBonus := punctCount / 10 // Punctuation also contributes to tokens

	totalTokens := baseTokens + wordBonus + punctBonus
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

	// Reinitialize strategies with new model
	r.initializeStrategies()
	r.logger.WithField("model", model).Info("Updated reasoning model")
}

// SetLongContextModel sets the long context model for high token count requests
func (r *ModelRouter) SetLongContextModel(model string) {
	r.mu.Lock()
	r.longContextModel = model
	r.mu.Unlock()

	// Reinitialize strategies with new model
	r.initializeStrategies()
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

	// Reinitialize strategies with new models
	r.initializeStrategies()
	r.logger.WithFields(logrus.Fields{
		"big_model":   bigModel,
		"small_model": smallModel,
	}).Info("Updated model router configuration")
}

// AddCustomStrategy adds a custom routing strategy
func (r *ModelRouter) AddCustomStrategy(strategy ModelRoutingStrategy, priority int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Insert strategy at specified priority position
	if priority < 0 || priority > len(r.strategies) {
		priority = len(r.strategies) // Append at end if invalid priority
	}

	// Create new slice with capacity for one more element
	newStrategies := make([]ModelRoutingStrategy, len(r.strategies)+1)
	copy(newStrategies[:priority], r.strategies[:priority])
	newStrategies[priority] = strategy
	copy(newStrategies[priority+1:], r.strategies[priority:])

	r.strategies = newStrategies
	r.logger.WithField("priority", priority).Info("Added custom routing strategy")
}

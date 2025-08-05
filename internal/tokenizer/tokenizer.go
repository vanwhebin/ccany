package tokenizer

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/tiktoken-go/tokenizer"
)

// TokenCounter provides token counting functionality for different models
type TokenCounter struct {
	logger     *logrus.Logger
	encoders   map[string]tokenizer.Codec
	mu         sync.RWMutex
	modelCache map[string]string // model name to encoding name mapping
}

// NewTokenCounter creates a new token counter
func NewTokenCounter(logger *logrus.Logger) *TokenCounter {
	tc := &TokenCounter{
		logger:     logger,
		encoders:   make(map[string]tokenizer.Codec),
		modelCache: make(map[string]string),
	}

	// Initialize default model mappings
	tc.initializeModelMappings()

	return tc
}

// initializeModelMappings sets up the model to encoding mappings
func (tc *TokenCounter) initializeModelMappings() {
	// OpenAI models
	tc.modelCache["gpt-4"] = "cl100k_base"
	tc.modelCache["gpt-4-0125-preview"] = "cl100k_base"
	tc.modelCache["gpt-4-turbo-preview"] = "cl100k_base"
	tc.modelCache["gpt-4-vision-preview"] = "cl100k_base"
	tc.modelCache["gpt-4-32k"] = "cl100k_base"
	tc.modelCache["gpt-4-0613"] = "cl100k_base"
	tc.modelCache["gpt-4-32k-0613"] = "cl100k_base"

	tc.modelCache["gpt-3.5-turbo"] = "cl100k_base"
	tc.modelCache["gpt-3.5-turbo-0125"] = "cl100k_base"
	tc.modelCache["gpt-3.5-turbo-1106"] = "cl100k_base"
	tc.modelCache["gpt-3.5-turbo-0613"] = "cl100k_base"
	tc.modelCache["gpt-3.5-turbo-16k"] = "cl100k_base"
	tc.modelCache["gpt-3.5-turbo-16k-0613"] = "cl100k_base"

	// Claude models - use cl100k_base as approximation
	tc.modelCache["claude-3-opus-20240229"] = "cl100k_base"
	tc.modelCache["claude-3-sonnet-20240229"] = "cl100k_base"
	tc.modelCache["claude-3-haiku-20240307"] = "cl100k_base"
	tc.modelCache["claude-3-5-sonnet-20241022"] = "cl100k_base"
	tc.modelCache["claude-3-5-haiku-20241022"] = "cl100k_base"

	// Default for unknown models
	tc.modelCache["default"] = "cl100k_base"
}

// getEncoder gets or creates an encoder for the given encoding
func (tc *TokenCounter) getEncoder(encoding string) (tokenizer.Codec, error) {
	tc.mu.RLock()
	if enc, exists := tc.encoders[encoding]; exists {
		tc.mu.RUnlock()
		return enc, nil
	}
	tc.mu.RUnlock()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Double-check after acquiring write lock
	if enc, exists := tc.encoders[encoding]; exists {
		return enc, nil
	}

	// Create new encoder - use Get method with predefined encoding
	var enc tokenizer.Codec
	var err error

	switch encoding {
	case "cl100k_base":
		enc, err = tokenizer.Get(tokenizer.Cl100kBase)
	case "p50k_base":
		enc, err = tokenizer.Get(tokenizer.P50kBase)
	case "p50k_edit":
		enc, err = tokenizer.Get(tokenizer.P50kEdit)
	case "r50k_base":
		enc, err = tokenizer.Get(tokenizer.R50kBase)
	default:
		// Default to cl100k_base for unknown encodings
		tc.logger.Warnf("Unknown encoding %s, using cl100k_base", encoding)
		enc, err = tokenizer.Get(tokenizer.Cl100kBase)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get encoder for %s: %w", encoding, err)
	}

	tc.encoders[encoding] = enc
	return enc, nil
}

// getEncodingForModel returns the appropriate encoding for a model
func (tc *TokenCounter) getEncodingForModel(model string) string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Direct match
	if encoding, exists := tc.modelCache[model]; exists {
		return encoding
	}

	// Check for partial matches (e.g., "gpt-4" prefix)
	lowerModel := strings.ToLower(model)
	for modelPrefix, encoding := range tc.modelCache {
		if strings.HasPrefix(lowerModel, strings.ToLower(modelPrefix)) {
			return encoding
		}
	}

	// Default encoding
	return tc.modelCache["default"]
}

// CountTokens counts tokens in text for a specific model
func (tc *TokenCounter) CountTokens(text string, model string) (int, error) {
	if text == "" {
		return 0, nil
	}

	encoding := tc.getEncodingForModel(model)
	encoder, err := tc.getEncoder(encoding)
	if err != nil {
		return 0, fmt.Errorf("failed to get encoder: %w", err)
	}

	tokens, _, err := encoder.Encode(text)
	if err != nil {
		return 0, fmt.Errorf("failed to encode text: %w", err)
	}

	return len(tokens), nil
}

// CountMessagesTokens counts tokens for a conversation (OpenAI format)
func (tc *TokenCounter) CountMessagesTokens(messages []map[string]interface{}, model string) (int, error) {
	totalTokens := 0

	// Message formatting overhead per message
	// Ref: https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	tokensPerMessage := 3 // Default for cl100k_base
	tokensPerName := 1

	for _, message := range messages {
		totalTokens += tokensPerMessage

		// Count role tokens
		if role, ok := message["role"].(string); ok {
			roleTokens, err := tc.CountTokens(role, model)
			if err != nil {
				return 0, err
			}
			totalTokens += roleTokens
		}

		// Count content tokens
		if content, ok := message["content"].(string); ok {
			contentTokens, err := tc.CountTokens(content, model)
			if err != nil {
				return 0, err
			}
			totalTokens += contentTokens
		}

		// Count name tokens if present
		if name, ok := message["name"].(string); ok && name != "" {
			nameTokens, err := tc.CountTokens(name, model)
			if err != nil {
				return 0, err
			}
			totalTokens += nameTokens + tokensPerName
		}
	}

	// Every reply is primed with <|start|>assistant<|message|>
	totalTokens += 3

	return totalTokens, nil
}

// CountClaudeMessagesTokens counts tokens for Claude messages format
func (tc *TokenCounter) CountClaudeMessagesTokens(messages []interface{}, system interface{}, model string) (int, error) {
	totalTokens := 0

	// Count system message tokens
	if system != nil {
		systemTokens, err := tc.CountContentTokens(system, model)
		if err != nil {
			return 0, err
		}
		totalTokens += systemTokens
	}

	// Count message tokens
	for _, msg := range messages {
		if msgMap, ok := msg.(map[string]interface{}); ok {
			// Count role
			if role, ok := msgMap["role"].(string); ok {
				roleTokens, err := tc.CountTokens(role, model)
				if err != nil {
					return 0, err
				}
				totalTokens += roleTokens
			}

			// Count content
			if content, exists := msgMap["content"]; exists {
				contentTokens, err := tc.CountContentTokens(content, model)
				if err != nil {
					return 0, err
				}
				totalTokens += contentTokens
			}
		}
	}

	return totalTokens, nil
}

// CountContentTokens counts tokens in content which can be string or array
func (tc *TokenCounter) CountContentTokens(content interface{}, model string) (int, error) {
	switch c := content.(type) {
	case string:
		return tc.CountTokens(c, model)

	case []interface{}:
		totalTokens := 0
		for _, block := range c {
			if blockMap, ok := block.(map[string]interface{}); ok {
				blockType, _ := blockMap["type"].(string)

				switch blockType {
				case "text":
					if text, ok := blockMap["text"].(string); ok {
						tokens, err := tc.CountTokens(text, model)
						if err != nil {
							return 0, err
						}
						totalTokens += tokens
					}

				case "image":
					// Image tokens estimation
					// Based on OpenAI's image token calculation
					totalTokens += 85 // Base cost
					if source, ok := blockMap["source"].(map[string]interface{}); ok {
						if mediaType, ok := source["media_type"].(string); ok {
							// High detail images cost more
							if strings.Contains(mediaType, "high") {
								totalTokens += 1105 // Additional cost for high detail
							} else {
								totalTokens += 85 // Low detail
							}
						}
					}

				case "tool_use":
					// Tool use token estimation
					totalTokens += 50 // Base overhead
					if name, ok := blockMap["name"].(string); ok {
						nameTokens, _ := tc.CountTokens(name, model)
						totalTokens += nameTokens
					}
					if input, ok := blockMap["input"].(map[string]interface{}); ok {
						// Estimate based on JSON serialization
						inputTokens := tc.estimateJSONTokens(input, model)
						totalTokens += inputTokens
					}

				case "tool_result":
					// Tool result token estimation
					totalTokens += 30 // Base overhead
					if content, ok := blockMap["content"]; ok {
						if contentStr, ok := content.(string); ok {
							tokens, _ := tc.CountTokens(contentStr, model)
							totalTokens += tokens
						} else {
							// Complex content
							totalTokens += 100
						}
					}
				}
			}
		}
		return totalTokens, nil

	default:
		// Fallback estimation
		return 10, nil
	}
}

// estimateJSONTokens estimates tokens for JSON data
func (tc *TokenCounter) estimateJSONTokens(data interface{}, model string) int {
	// Simple estimation based on JSON string representation
	switch v := data.(type) {
	case string:
		tokens, _ := tc.CountTokens(v, model)
		return tokens
	case map[string]interface{}:
		totalTokens := 2 // {} brackets
		for key, value := range v {
			keyTokens, _ := tc.CountTokens(key, model)
			totalTokens += keyTokens + 2 // key + : and ,
			totalTokens += tc.estimateJSONTokens(value, model)
		}
		return totalTokens
	case []interface{}:
		totalTokens := 2 // [] brackets
		for _, item := range v {
			totalTokens += tc.estimateJSONTokens(item, model)
			totalTokens += 1 // comma
		}
		return totalTokens
	default:
		// Numbers, booleans, null
		return 1
	}
}

// EstimateTokensFromLength provides a quick estimation based on text length
// This is used as a fallback when precise counting is not needed
func (tc *TokenCounter) EstimateTokensFromLength(text string) int {
	if text == "" {
		return 0
	}

	// Use different ratios based on content type
	// English text: ~4 chars per token
	// Code: ~3 chars per token
	// Mixed content: ~3.5 chars per token

	// Simple heuristic: check for code indicators
	hasCode := strings.Contains(text, "```") ||
		strings.Contains(text, "function") ||
		strings.Contains(text, "class") ||
		strings.Contains(text, "import") ||
		strings.Contains(text, "const ") ||
		strings.Contains(text, "var ") ||
		strings.Contains(text, "let ")

	charsPerToken := 4.0
	if hasCode {
		charsPerToken = 3.0
	}

	// Account for Unicode characters (they typically use more tokens)
	runeCount := len([]rune(text))
	byteCount := len(text)
	if float64(byteCount)/float64(runeCount) > 1.5 {
		// Has multi-byte characters
		charsPerToken = 2.5
	}

	tokens := int(float64(runeCount) / charsPerToken)
	if tokens < 1 && runeCount > 0 {
		tokens = 1
	}

	return tokens
}

// Close cleans up resources
func (tc *TokenCounter) Close() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Clear encoders
	tc.encoders = make(map[string]tokenizer.Codec)
}

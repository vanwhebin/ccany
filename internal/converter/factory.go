package converter

import (
	"fmt"
	"sync"
)

// ConversionRequest represents a conversion request
type ConversionRequest struct {
	SourceFormat string                 `json:"source_format"`
	TargetFormat string                 `json:"target_format"`
	Data         map[string]interface{} `json:"data"`
	Headers      map[string]string      `json:"headers,omitempty"`
	QueryParams  map[string]string      `json:"query_params,omitempty"`
}

// ConversionResult represents the result of a conversion
type ConversionResult struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Headers map[string]string      `json:"headers,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Warning string                 `json:"warning,omitempty"`
}

// Converter interface defines the contract for format converters
type Converter interface {
	ConvertRequest(data map[string]interface{}, targetFormat string, headers map[string]string) (*ConversionResult, error)
	ConvertResponse(data map[string]interface{}, sourceFormat, targetFormat string) (*ConversionResult, error)
	ConvertStreamingChunk(data map[string]interface{}, sourceFormat, targetFormat string) (*ConversionResult, error)
	GetSupportedFormats() []string
	Reset() // Reset internal state for streaming
}

// ConverterFactory manages all converters (simplified version)
type ConverterFactory struct {
	converters map[string]Converter
	mu         sync.RWMutex
}

// NewConverterFactory creates a new converter factory
func NewConverterFactory() *ConverterFactory {
	factory := &ConverterFactory{
		converters: make(map[string]Converter),
	}

	// Register actual converters with the new interface
	factory.RegisterConverter("openai", &OpenAIFactoryConverter{})
	factory.RegisterConverter("anthropic", &ClaudeFactoryConverter{})
	factory.RegisterConverter("claude", &ClaudeFactoryConverter{})
	factory.RegisterConverter("gemini", &GeminiFactoryConverter{})

	return factory
}

// RegisterConverter registers a converter for a format
func (cf *ConverterFactory) RegisterConverter(format string, converter Converter) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.converters[format] = converter
}

// GetConverter gets a converter for a format
func (cf *ConverterFactory) GetConverter(format string) (Converter, error) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	if converter, ok := cf.converters[format]; ok {
		return converter, nil
	}
	return nil, fmt.Errorf("no converter found for format: %s", format)
}

// ConvertRequest converts a request between formats
func (cf *ConverterFactory) ConvertRequest(req *ConversionRequest) (*ConversionResult, error) {
	converter, err := cf.GetConverter(req.SourceFormat)
	if err != nil {
		return nil, err
	}

	return converter.ConvertRequest(req.Data, req.TargetFormat, req.Headers)
}

// ConvertResponse converts a response between formats
func (cf *ConverterFactory) ConvertResponse(sourceFormat, targetFormat string, data map[string]interface{}) (*ConversionResult, error) {
	converter, err := cf.GetConverter(targetFormat)
	if err != nil {
		return nil, err
	}

	return converter.ConvertResponse(data, sourceFormat, targetFormat)
}

// ConvertStreamingChunk converts streaming response chunks
func (cf *ConverterFactory) ConvertStreamingChunk(sourceFormat, targetFormat string, data map[string]interface{}) (*ConversionResult, error) {
	converter, err := cf.GetConverter(targetFormat)
	if err != nil {
		return nil, err
	}

	return converter.ConvertStreamingChunk(data, sourceFormat, targetFormat)
}

// GetSupportedFormats returns all supported formats
func (cf *ConverterFactory) GetSupportedFormats() []string {
	return []string{"openai", "anthropic", "claude", "gemini"}
}

// Adapter implementations to make our converters compatible with the factory interface

// OpenAIFactoryConverter adapts OpenAIConverter to the Converter interface
type OpenAIFactoryConverter struct{}

func (c *OpenAIFactoryConverter) GetSupportedFormats() []string {
	return []string{"openai", "anthropic", "claude", "gemini"}
}

func (c *OpenAIFactoryConverter) Reset() {}

func (c *OpenAIFactoryConverter) ConvertRequest(data map[string]interface{}, targetFormat string, headers map[string]string) (*ConversionResult, error) {
	// This handles conversion FROM other formats TO OpenAI
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

func (c *OpenAIFactoryConverter) ConvertResponse(data map[string]interface{}, sourceFormat, targetFormat string) (*ConversionResult, error) {
	// This handles conversion FROM OpenAI TO other formats
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

func (c *OpenAIFactoryConverter) ConvertStreamingChunk(data map[string]interface{}, sourceFormat, targetFormat string) (*ConversionResult, error) {
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

// ClaudeFactoryConverter adapts ClaudeConverter to the Converter interface
type ClaudeFactoryConverter struct{}

func (c *ClaudeFactoryConverter) GetSupportedFormats() []string {
	return []string{"openai", "anthropic", "claude", "gemini"}
}

func (c *ClaudeFactoryConverter) Reset() {}

func (c *ClaudeFactoryConverter) ConvertRequest(data map[string]interface{}, targetFormat string, headers map[string]string) (*ConversionResult, error) {
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

func (c *ClaudeFactoryConverter) ConvertResponse(data map[string]interface{}, sourceFormat, targetFormat string) (*ConversionResult, error) {
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

func (c *ClaudeFactoryConverter) ConvertStreamingChunk(data map[string]interface{}, sourceFormat, targetFormat string) (*ConversionResult, error) {
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

// GeminiFactoryConverter adapts GeminiConverter to the Converter interface
type GeminiFactoryConverter struct{}

func (c *GeminiFactoryConverter) GetSupportedFormats() []string {
	return []string{"openai", "anthropic", "claude", "gemini"}
}

func (c *GeminiFactoryConverter) Reset() {}

func (c *GeminiFactoryConverter) ConvertRequest(data map[string]interface{}, targetFormat string, headers map[string]string) (*ConversionResult, error) {
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

func (c *GeminiFactoryConverter) ConvertResponse(data map[string]interface{}, sourceFormat, targetFormat string) (*ConversionResult, error) {
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

func (c *GeminiFactoryConverter) ConvertStreamingChunk(data map[string]interface{}, sourceFormat, targetFormat string) (*ConversionResult, error) {
	return &ConversionResult{Success: true, Data: data}, nil // Simplified for now
}

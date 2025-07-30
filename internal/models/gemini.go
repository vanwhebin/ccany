package models

import "encoding/json"

// GeminiRequest represents a Gemini API request
type GeminiRequest struct {
	Contents          []GeminiContent         `json:"contents"`
	Tools             []GeminiTool            `json:"tools,omitempty"`
	ToolConfig        *GeminiToolConfig       `json:"toolConfig,omitempty"`
	SystemInstruction *GeminiContent          `json:"systemInstruction,omitempty"`
	GenerationConfig  *GeminiGenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings    []GeminiSafetySetting   `json:"safetySettings,omitempty"`
	Model             string                  `json:"model,omitempty"` // For internal use
}

// GeminiContent represents content in Gemini format
type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text             string                  `json:"text,omitempty"`
	InlineData       *GeminiInlineData       `json:"inlineData,omitempty"`
	FileData         *GeminiFileData         `json:"fileData,omitempty"`
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
}

// GeminiInlineData represents inline binary data
type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // Base64 encoded
}

// GeminiFileData represents file data
type GeminiFileData struct {
	MimeType string `json:"mimeType"`
	FileURI  string `json:"fileUri"`
}

// GeminiFunctionCall represents a function call
type GeminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

// GeminiFunctionResponse represents a function response
type GeminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// GeminiTool represents a tool
type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// GeminiFunctionDeclaration represents a function declaration
type GeminiFunctionDeclaration struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Parameters  *GeminiFunctionParameters `json:"parameters,omitempty"`
}

// GeminiFunctionParameters represents function parameters
type GeminiFunctionParameters struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// GeminiToolConfig represents tool configuration
type GeminiToolConfig struct {
	FunctionCallingConfig *GeminiFunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

// GeminiFunctionCallingConfig represents function calling configuration
type GeminiFunctionCallingConfig struct {
	Mode                 string   `json:"mode,omitempty"` // AUTO, ANY, NONE
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

// GeminiGenerationConfig represents generation configuration
type GeminiGenerationConfig struct {
	StopSequences   []string `json:"stopSequences,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
}

// GeminiSafetySetting represents safety settings
type GeminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GeminiResponse represents a Gemini API response
type GeminiResponse struct {
	Candidates     []GeminiCandidate     `json:"candidates,omitempty"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *GeminiUsageMetadata  `json:"usageMetadata,omitempty"`
}

// GeminiCandidate represents a response candidate
type GeminiCandidate struct {
	Content       *GeminiContent       `json:"content,omitempty"`
	FinishReason  string               `json:"finishReason,omitempty"`
	Index         int                  `json:"index,omitempty"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// GeminiPromptFeedback represents prompt feedback
type GeminiPromptFeedback struct {
	BlockReason   string               `json:"blockReason,omitempty"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// GeminiSafetyRating represents safety rating
type GeminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// GeminiUsageMetadata represents usage metadata
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
	TotalTokenCount      int `json:"totalTokenCount,omitempty"`
}

// GeminiStreamResponse represents a streaming response
type GeminiStreamResponse struct {
	Candidates     []GeminiCandidate     `json:"candidates,omitempty"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *GeminiUsageMetadata  `json:"usageMetadata,omitempty"`
}

// GeminiErrorResponse represents an error response
type GeminiErrorResponse struct {
	Error GeminiError `json:"error"`
}

// GeminiError represents an error
type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ToJSON converts the request to JSON
func (r *GeminiRequest) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON creates a request from JSON
func (r *GeminiRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

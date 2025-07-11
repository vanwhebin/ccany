package config

// ClaudeConfig Claude API configuration
type ClaudeConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// OpenAIConfig OpenAI API configuration
type OpenAIConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// ServerConfig server configuration
type ServerConfig struct {
	Port     string `json:"port"`
	LogLevel string `json:"log_level"`
}

// ModelConfig model configuration
type ModelConfig struct {
	BigModel   string `json:"big_model"`
	SmallModel string `json:"small_model"`
}

// PerformanceConfig performance configuration
type PerformanceConfig struct {
	MaxTokens      int `json:"max_tokens"`
	RequestTimeout int `json:"request_timeout"`
	MaxRetries     int `json:"max_retries"`
}

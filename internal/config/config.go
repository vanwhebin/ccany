package config

// Config holds all configuration for the application
type Config struct {
	// OpenAI API Configuration
	OpenAIAPIKey    string
	OpenAIBaseURL   string
	AzureAPIVersion string

	// Claude API Configuration
	ClaudeAPIKey  string
	ClaudeBaseURL string

	// Model Configuration
	BigModel   string
	SmallModel string

	// Server Configuration
	Host     string
	Port     int
	LogLevel string

	// Performance Configuration
	MaxTokensLimit int
	MinTokensLimit int
	RequestTimeout int
	MaxRetries     int
	Temperature    float64
	StreamEnabled  bool

	// Database Configuration
	DatabaseURL string

	// Proxy Configuration
	ProxyEnabled          bool
	ProxyType             string // "http" or "socks5"
	HTTPProxy             string
	SOCKS5Proxy           string
	SOCKS5ProxyUser       string
	SOCKS5ProxyPassword   string
	IgnoreSSLVerification bool
}

// ValidateAPIKey performs basic validation on the API key
func (c *Config) ValidateAPIKey() bool {
	if c.OpenAIAPIKey == "" {
		return false
	}
	// Skip format check - allow any non-empty API key
	return true
}

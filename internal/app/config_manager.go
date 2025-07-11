package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"ccany/internal/config"
	"ccany/internal/database"
)

// ConfigManager application layer configuration manager
type ConfigManager struct {
	db            *database.Database
	configService *config.Service
	ctx           context.Context
}

// NewConfigManager creates a configuration manager
func NewConfigManager(db *database.Database, ctx context.Context) *ConfigManager {
	// Create configuration service
	configService := config.NewService(db.Client, db.CryptoService)

	return &ConfigManager{
		db:            db,
		configService: configService,
		ctx:           ctx,
	}
}

// Initialize initializes the configuration manager
func (m *ConfigManager) Initialize() error {
	// Initialize default configurations to database
	if err := m.configService.InitializeDefaultConfigs(m.ctx); err != nil {
		return fmt.Errorf("failed to initialize default configs: %w", err)
	}

	return nil
}

// GetConfig retrieves complete configuration (from database)
func (m *ConfigManager) GetConfig() (*config.Config, error) {
	// Retrieve configuration from database
	return m.getConfigFromDB()
}

// getConfigFromDB retrieves configuration from database
func (m *ConfigManager) getConfigFromDB() (*config.Config, error) {
	// Get configuration, use default value if failed or empty
	getConfigOrDefault := func(key, defaultValue string) string {
		value, err := m.configService.GetConfig(m.ctx, key)
		if err != nil || value == "" {
			return defaultValue
		}
		return value
	}

	// Get OpenAI configuration
	openaiKey := getConfigOrDefault(config.KeyOpenAIAPIKey, "")
	openaiBaseURL := getConfigOrDefault(config.KeyOpenAIBaseURL, "https://api.openai.com/v1")

	// Get Claude configuration
	claudeKey := getConfigOrDefault(config.KeyClaudeAPIKey, "")
	claudeBaseURL := getConfigOrDefault(config.KeyClaudeBaseURL, "https://api.anthropic.com")

	// Get Azure configuration
	azureAPIVersion := getConfigOrDefault(config.KeyAzureAPIVersion, "")

	// Get model configuration
	bigModel := getConfigOrDefault(config.KeyBigModel, "gpt-4o")
	smallModel := getConfigOrDefault(config.KeySmallModel, "gpt-4o-mini")

	// Get server configuration
	serverHost := getConfigOrDefault(config.KeyServerHost, "0.0.0.0")
	serverPort := getConfigOrDefault(config.KeyServerPort, "8082")
	logLevel := getConfigOrDefault(config.KeyLogLevel, "info")

	// Get performance configuration
	maxTokensStr := getConfigOrDefault(config.KeyMaxTokens, "4096")
	minTokensStr := getConfigOrDefault(config.KeyMinTokens, "100")
	requestTimeoutStr := getConfigOrDefault(config.KeyRequestTimeout, "90")
	maxRetriesStr := getConfigOrDefault(config.KeyMaxRetries, "2")
	temperatureStr := getConfigOrDefault(config.KeyTemperature, "0.7")
	streamEnabledStr := getConfigOrDefault(config.KeyStreamEnabled, "true")

	// Convert data types
	port, _ := strconv.Atoi(serverPort)
	maxTokens, _ := strconv.Atoi(maxTokensStr)
	minTokens, _ := strconv.Atoi(minTokensStr)
	requestTimeout, _ := strconv.Atoi(requestTimeoutStr)
	maxRetries, _ := strconv.Atoi(maxRetriesStr)
	temperature, _ := strconv.ParseFloat(temperatureStr, 64)
	streamEnabled, _ := strconv.ParseBool(streamEnabledStr)

	return &config.Config{
		OpenAIAPIKey:    openaiKey,
		OpenAIBaseURL:   openaiBaseURL,
		AzureAPIVersion: azureAPIVersion,
		ClaudeAPIKey:    claudeKey,
		ClaudeBaseURL:   claudeBaseURL,
		BigModel:        bigModel,
		SmallModel:      smallModel,
		Host:            serverHost,
		Port:            port,
		LogLevel:        logLevel,
		MaxTokensLimit:  maxTokens,
		MinTokensLimit:  minTokens,
		RequestTimeout:  requestTimeout,
		MaxRetries:      maxRetries,
		Temperature:     temperature,
		StreamEnabled:   streamEnabled,
		DatabaseURL:     "sqlite3://./data/proxy.db", // Can be set through environment variables or configuration files
	}, nil
}

// UpdateConfig updates configuration to database
func (m *ConfigManager) UpdateConfig(key, value string, encrypted bool) error {
	return m.configService.SetConfig(m.ctx, key, value, encrypted)
}

// GetConfigValue retrieves a single configuration value
func (m *ConfigManager) GetConfigValue(key string) (string, error) {
	return m.configService.GetConfig(m.ctx, key)
}

// GetAllConfigs retrieves all configurations (for management interface)
func (m *ConfigManager) GetAllConfigs() (map[string]interface{}, error) {
	return m.configService.GetAllConfigs(m.ctx)
}

// IsConfigured checks if necessary API keys are configured
func (m *ConfigManager) IsConfigured() (bool, error) {
	return m.configService.IsConfigured(m.ctx)
}

// ValidateConfig validates the configuration
func (m *ConfigManager) ValidateConfig() error {
	cfg, err := m.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Validate at least one API key
	if cfg.OpenAIAPIKey == "" && cfg.ClaudeAPIKey == "" {
		return fmt.Errorf("at least one API key (OpenAI or Claude) is required")
	}

	// Validate OpenAI API key format
	if cfg.OpenAIAPIKey != "" && !strings.HasPrefix(cfg.OpenAIAPIKey, "sk-") {
		return fmt.Errorf("invalid OpenAI API key format")
	}

	// Validate Claude API key format
	if cfg.ClaudeAPIKey != "" && !strings.HasPrefix(cfg.ClaudeAPIKey, "sk-ant-") {
		return fmt.Errorf("invalid Claude API key format")
	}

	// Validate port range
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", cfg.Port)
	}

	// Validate token limits
	if cfg.MaxTokensLimit < cfg.MinTokensLimit {
		return fmt.Errorf("max tokens limit (%d) must be greater than min tokens limit (%d)",
			cfg.MaxTokensLimit, cfg.MinTokensLimit)
	}

	return nil
}

// GetConfigService retrieves configuration service (for other modules)
func (m *ConfigManager) GetConfigService() *config.Service {
	return m.configService
}

// GetDatabase retrieves database instance
func (m *ConfigManager) GetDatabase() *database.Database {
	return m.db
}

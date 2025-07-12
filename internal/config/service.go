package config

import (
	"context"
	"fmt"
	"log"

	"ccany/ent"
	"ccany/ent/appconfig"
	"ccany/internal/crypto"
)

// Service configuration service
type Service struct {
	db     *ent.Client
	crypto *crypto.CryptoService
}

// NewService creates configuration service
func NewService(db *ent.Client, cryptoService *crypto.CryptoService) *Service {
	return &Service{
		db:     db,
		crypto: cryptoService,
	}
}

// ConfigKey configuration key constants
const (
	KeyClaudeAPIKey    = "claude_api_key"
	KeyClaudeBaseURL   = "claude_base_url"
	KeyOpenAIAPIKey    = "openai_api_key"
	KeyOpenAIBaseURL   = "openai_base_url"
	KeyAzureAPIVersion = "azure_api_version"
	KeyBigModel        = "big_model"
	KeySmallModel      = "small_model"
	KeyServerHost      = "server_host"
	KeyServerPort      = "server_port"
	KeyLogLevel        = "log_level"
	KeyMaxTokens       = "max_tokens"
	KeyMinTokens       = "min_tokens"
	KeyRequestTimeout  = "request_timeout"
	KeyMaxRetries      = "max_retries"
	KeyTemperature     = "temperature"
	KeyStreamEnabled   = "stream_enabled"
	KeyRateLimitRPM    = "rate_limit_rpm"
	KeyRateLimitTPM    = "rate_limit_tpm"
	// 代理配置键
	KeyProxyEnabled          = "proxy_enabled"
	KeyProxyType             = "proxy_type"
	KeyHTTPProxy             = "http_proxy"
	KeySOCKS5Proxy           = "socks5_proxy"
	KeySOCKS5ProxyUser       = "socks5_proxy_user"
	KeySOCKS5ProxyPassword   = "socks5_proxy_password"
	KeyIgnoreSSLVerification = "ignore_ssl_verification"
)

// GetConfig gets configuration value
func (s *Service) GetConfig(ctx context.Context, key string) (string, error) {
	cfg, err := s.db.AppConfig.Query().
		Where(appconfig.Key(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil // return empty string if config doesn't exist
		}
		return "", fmt.Errorf("failed to get config %s: %w", key, err)
	}

	// if sensitive config, need to decrypt
	if cfg.IsEncrypted {
		decrypted, err := s.crypto.Decrypt(cfg.Value)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt config %s: %w", key, err)
		}
		return decrypted, nil
	}

	return cfg.Value, nil
}

// SetConfig sets configuration value
func (s *Service) SetConfig(ctx context.Context, key, value string, encrypted bool) error {
	finalValue := value

	// if encryption is needed
	if encrypted {
		encryptedValue, err := s.crypto.Encrypt(value)
		if err != nil {
			return fmt.Errorf("failed to encrypt config %s: %w", key, err)
		}
		finalValue = encryptedValue
	}

	// Check if config exists
	_, err := s.db.AppConfig.Query().
		Where(appconfig.Key(key)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// Config doesn't exist, create new one
			configID := fmt.Sprintf("config_%s", key)

			_, err = s.db.AppConfig.Create().
				SetID(configID).
				SetKey(key).
				SetValue(finalValue).
				SetIsEncrypted(encrypted).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create config %s: %w", key, err)
			}
		} else {
			return fmt.Errorf("failed to query config %s: %w", key, err)
		}
	} else {
		// Config exists, update it
		err = s.db.AppConfig.Update().
			Where(appconfig.Key(key)).
			SetValue(finalValue).
			SetIsEncrypted(encrypted).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to update config %s: %w", key, err)
		}
	}

	return nil
}

// GetAllConfigs gets all configurations (excluding sensitive values)
func (s *Service) GetAllConfigs(ctx context.Context) (map[string]interface{}, error) {
	configs, err := s.db.AppConfig.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all configs: %w", err)
	}

	result := make(map[string]interface{})
	for _, cfg := range configs {
		if cfg.IsEncrypted {
			// sensitive config only returns whether it's set
			result[cfg.Key] = map[string]interface{}{
				"is_set":       cfg.Value != "",
				"is_encrypted": true,
			}
		} else {
			result[cfg.Key] = map[string]interface{}{
				"value":        cfg.Value,
				"is_encrypted": false,
			}
		}
	}

	return result, nil
}

// DeleteConfig deletes configuration
func (s *Service) DeleteConfig(ctx context.Context, key string) error {
	_, err := s.db.AppConfig.Delete().
		Where(appconfig.Key(key)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete config %s: %w", key, err)
	}
	return nil
}

// InitializeDefaultConfigs initializes default configurations
func (s *Service) InitializeDefaultConfigs(ctx context.Context) error {
	defaults := map[string]struct {
		value     string
		encrypted bool
	}{
		KeyClaudeBaseURL:   {"https://api.anthropic.com", false},
		KeyOpenAIBaseURL:   {"https://api.openai.com/v1", false},
		KeyAzureAPIVersion: {"", false},
		KeyBigModel:        {"gpt-4o", false},
		KeySmallModel:      {"gpt-4o-mini", false},
		KeyServerHost:      {"0.0.0.0", false},
		KeyServerPort:      {"8082", false},
		KeyLogLevel:        {"info", false},
		KeyMaxTokens:       {"4096", false},
		KeyMinTokens:       {"100", false},
		KeyRequestTimeout:  {"90", false},
		KeyMaxRetries:      {"2", false},
		KeyTemperature:     {"0.7", false},
		KeyStreamEnabled:   {"true", false},
		KeyRateLimitRPM:    {"60", false},
		KeyRateLimitTPM:    {"100000", false},
	}

	for key, defaultConfig := range defaults {
		// check if config already exists
		existing, err := s.GetConfig(ctx, key)
		if err != nil {
			log.Printf("Error checking existing config %s: %v", key, err)
			continue
		}

		// if config doesn't exist, set default value
		if existing == "" {
			err = s.SetConfig(ctx, key, defaultConfig.value, defaultConfig.encrypted)
			if err != nil {
				log.Printf("Error setting default config %s: %v", key, err)
			}
		}
	}

	return nil
}

// GetClaudeConfig gets Claude configuration
func (s *Service) GetClaudeConfig(ctx context.Context) (*ClaudeConfig, error) {
	apiKey, err := s.GetConfig(ctx, KeyClaudeAPIKey)
	if err != nil {
		return nil, err
	}

	baseURL, err := s.GetConfig(ctx, KeyClaudeBaseURL)
	if err != nil {
		return nil, err
	}

	return &ClaudeConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}, nil
}

// GetOpenAIConfig gets OpenAI configuration
func (s *Service) GetOpenAIConfig(ctx context.Context) (*OpenAIConfig, error) {
	apiKey, err := s.GetConfig(ctx, KeyOpenAIAPIKey)
	if err != nil {
		return nil, err
	}

	baseURL, err := s.GetConfig(ctx, KeyOpenAIBaseURL)
	if err != nil {
		return nil, err
	}

	return &OpenAIConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}, nil
}

// GetServerConfig gets server configuration
func (s *Service) GetServerConfig(ctx context.Context) (*ServerConfig, error) {
	port, err := s.GetConfig(ctx, KeyServerPort)
	if err != nil {
		return nil, err
	}

	logLevel, err := s.GetConfig(ctx, KeyLogLevel)
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		Port:     port,
		LogLevel: logLevel,
	}, nil
}

// IsConfigured checks if necessary API keys are configured
func (s *Service) IsConfigured(ctx context.Context) (bool, error) {
	claudeKey, err := s.GetConfig(ctx, KeyClaudeAPIKey)
	if err != nil {
		return false, err
	}

	openaiKey, err := s.GetConfig(ctx, KeyOpenAIAPIKey)
	if err != nil {
		return false, err
	}

	// at least one API key needs to be configured
	return claudeKey != "" || openaiKey != "", nil
}

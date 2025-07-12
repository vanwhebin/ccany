package claudecode

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// ClaudeConfig represents the Claude Code configuration structure
type ClaudeConfig struct {
	NumStartups            int                    `json:"numStartups"`
	AutoUpdaterStatus      string                 `json:"autoUpdaterStatus"`
	UserID                 string                 `json:"userID"`
	HasCompletedOnboarding bool                   `json:"hasCompletedOnboarding"`
	LastOnboardingVersion  string                 `json:"lastOnboardingVersion"`
	Projects               map[string]interface{} `json:"projects"`
	InstallationID         string                 `json:"installationID,omitempty"`
	LastUpdateCheckTime    *time.Time             `json:"lastUpdateCheckTime,omitempty"`
	TelemetryEnabled       bool                   `json:"telemetryEnabled"`
	AnalyticsEnabled       bool                   `json:"analyticsEnabled"`
	CrashReportingEnabled  bool                   `json:"crashReportingEnabled"`
}

// ConfigService handles Claude Code configuration management
type ConfigService struct {
	logger     *logrus.Logger
	configPath string
}

// NewConfigService creates a new configuration service
func NewConfigService(logger *logrus.Logger) *ConfigService {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".claude.json")

	return &ConfigService{
		logger:     logger,
		configPath: configPath,
	}
}

// InitializeConfig creates the Claude Code configuration file if it doesn't exist
func (s *ConfigService) InitializeConfig() error {
	// Check if config file already exists
	if _, err := os.Stat(s.configPath); err == nil {
		s.logger.Debug("Claude Code configuration file already exists")
		return nil
	}

	// Generate random user ID (64 characters)
	userID, err := s.generateRandomString(64)
	if err != nil {
		return fmt.Errorf("failed to generate user ID: %w", err)
	}

	// Generate installation ID (32 characters)
	installationID, err := s.generateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate installation ID: %w", err)
	}

	// Create default configuration
	config := ClaudeConfig{
		NumStartups:            184,
		AutoUpdaterStatus:      "enabled",
		UserID:                 userID,
		HasCompletedOnboarding: true,
		LastOnboardingVersion:  "1.0.17",
		Projects:               make(map[string]interface{}),
		InstallationID:         installationID,
		TelemetryEnabled:       true,
		AnalyticsEnabled:       true,
		CrashReportingEnabled:  true,
	}

	// Marshal to JSON with proper formatting
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	s.logger.WithField("path", s.configPath).Info("Created Claude Code configuration file")
	return nil
}

// GetConfig loads the Claude Code configuration
func (s *ConfigService) GetConfig() (*ClaudeConfig, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// UpdateConfig updates the Claude Code configuration
func (s *ConfigService) UpdateConfig(config *ClaudeConfig) error {
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(s.configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// IncrementStartupCount increments the startup count in the configuration
func (s *ConfigService) IncrementStartupCount() error {
	config, err := s.GetConfig()
	if err != nil {
		// If config doesn't exist, initialize it first
		if os.IsNotExist(err) {
			if err := s.InitializeConfig(); err != nil {
				return err
			}
			config, err = s.GetConfig()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	config.NumStartups++
	now := time.Now()
	config.LastUpdateCheckTime = &now

	return s.UpdateConfig(config)
}

// generateRandomString generates a random hex string of specified length
func (s *ConfigService) generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetConfigPath returns the path to the configuration file
func (s *ConfigService) GetConfigPath() string {
	return s.configPath
}

// ConfigExists checks if the configuration file exists
func (s *ConfigService) ConfigExists() bool {
	_, err := os.Stat(s.configPath)
	return err == nil
}

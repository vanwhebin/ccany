package claudecode

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"ccany/internal/config"

	"github.com/sirupsen/logrus"
)

// ClaudeConfig represents the Claude Code configuration structure
type ClaudeConfig struct {
	NumStartups            int                    `json:"numStartups" validate:"min=0"`
	AutoUpdaterStatus      string                 `json:"autoUpdaterStatus" validate:"oneof=enabled disabled"`
	UserID                 string                 `json:"userID" validate:"required,len=64"`
	HasCompletedOnboarding bool                   `json:"hasCompletedOnboarding"`
	LastOnboardingVersion  string                 `json:"lastOnboardingVersion"`
	Projects               map[string]interface{} `json:"projects"`
	InstallationID         string                 `json:"installationID,omitempty" validate:"omitempty,len=32"`
	LastUpdateCheckTime    *time.Time             `json:"lastUpdateCheckTime,omitempty"`
	TelemetryEnabled       bool                   `json:"telemetryEnabled"`
	AnalyticsEnabled       bool                   `json:"analyticsEnabled"`
	CrashReportingEnabled  bool                   `json:"crashReportingEnabled"`
}

// ConfigService handles Claude Code configuration management using database
type ConfigService struct {
	logger        *logrus.Logger
	configService *config.Service
	ctx           context.Context
}

// ConfigOption allows customizing the ConfigService
type ConfigOption func(*ConfigService)

// NewConfigService creates a new configuration service
func NewConfigService(logger *logrus.Logger, configService *config.Service, ctx context.Context, opts ...ConfigOption) *ConfigService {
	service := &ConfigService{
		logger:        logger,
		configService: configService,
		ctx:           ctx,
	}

	// Apply options
	for _, opt := range opts {
		opt(service)
	}

	return service
}

// InitializeConfig creates the Claude Code configuration if it doesn't exist
func (s *ConfigService) InitializeConfig() error {
	// Check if user ID already exists
	userID, err := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeUserID)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	// If user ID doesn't exist, generate a new one
	if userID == "" {
		// Generate random user ID (64 characters)
		userID, err = s.generateRandomString(64)
		if err != nil {
			return fmt.Errorf("failed to generate user ID: %w", err)
		}

		// Save user ID
		if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeUserID, userID, false); err != nil {
			return fmt.Errorf("failed to save user ID: %w", err)
		}
	}

	// Check if installation ID exists
	installationID, err := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeInstallationID)
	if err != nil {
		return fmt.Errorf("failed to get installation ID: %w", err)
	}

	// If installation ID doesn't exist, generate a new one
	if installationID == "" {
		// Generate installation ID (32 characters)
		installationID, err = s.generateRandomString(32)
		if err != nil {
			return fmt.Errorf("failed to generate installation ID: %w", err)
		}

		// Save installation ID
		if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeInstallationID, installationID, false); err != nil {
			return fmt.Errorf("failed to save installation ID: %w", err)
		}
	}

	s.logger.Info("Initialized Claude Code configuration")
	return nil
}

// GetConfig loads the Claude Code configuration from database
func (s *ConfigService) GetConfig() (*ClaudeConfig, error) {
	// Get all configuration values from database
	numStartupsStr, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeNumStartups)
	autoUpdaterStatus, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeAutoUpdaterStatus)
	userID, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeUserID)
	hasCompletedOnboardingStr, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeHasCompletedOnboarding)
	lastOnboardingVersion, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeLastOnboardingVersion)
	installationID, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeInstallationID)
	telemetryEnabledStr, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeTelemetryEnabled)
	analyticsEnabledStr, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeAnalyticsEnabled)
	crashReportingEnabledStr, _ := s.configService.GetConfig(s.ctx, config.KeyClaudeCodeCrashReportingEnabled)

	// Convert string values to appropriate types
	numStartups, _ := strconv.Atoi(numStartupsStr)
	hasCompletedOnboarding, _ := strconv.ParseBool(hasCompletedOnboardingStr)
	telemetryEnabled, _ := strconv.ParseBool(telemetryEnabledStr)
	analyticsEnabled, _ := strconv.ParseBool(analyticsEnabledStr)
	crashReportingEnabled, _ := strconv.ParseBool(crashReportingEnabledStr)

	config := &ClaudeConfig{
		NumStartups:            numStartups,
		AutoUpdaterStatus:      autoUpdaterStatus,
		UserID:                 userID,
		HasCompletedOnboarding: hasCompletedOnboarding,
		LastOnboardingVersion:  lastOnboardingVersion,
		Projects:               make(map[string]interface{}),
		InstallationID:         installationID,
		TelemetryEnabled:       telemetryEnabled,
		AnalyticsEnabled:       analyticsEnabled,
		CrashReportingEnabled:  crashReportingEnabled,
	}

	// Validate configuration
	if err := s.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// UpdateConfig updates the Claude Code configuration in database
func (s *ConfigService) UpdateConfig(cfg *ClaudeConfig) error {
	// Validate before updating
	if err := s.validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save each field to database
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeNumStartups, strconv.Itoa(cfg.NumStartups), false); err != nil {
		return err
	}
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeAutoUpdaterStatus, cfg.AutoUpdaterStatus, false); err != nil {
		return err
	}
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeUserID, cfg.UserID, false); err != nil {
		return err
	}
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeHasCompletedOnboarding, strconv.FormatBool(cfg.HasCompletedOnboarding), false); err != nil {
		return err
	}
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeLastOnboardingVersion, cfg.LastOnboardingVersion, false); err != nil {
		return err
	}
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeInstallationID, cfg.InstallationID, false); err != nil {
		return err
	}
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeTelemetryEnabled, strconv.FormatBool(cfg.TelemetryEnabled), false); err != nil {
		return err
	}
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeAnalyticsEnabled, strconv.FormatBool(cfg.AnalyticsEnabled), false); err != nil {
		return err
	}
	if err := s.configService.SetConfig(s.ctx, config.KeyClaudeCodeCrashReportingEnabled, strconv.FormatBool(cfg.CrashReportingEnabled), false); err != nil {
		return err
	}

	return nil
}

// IncrementStartupCount increments the startup count in the configuration
func (s *ConfigService) IncrementStartupCount() error {
	config, err := s.GetConfig()
	if err != nil {
		// If config doesn't exist, initialize it first
		if err := s.InitializeConfig(); err != nil {
			return err
		}
		config, err = s.GetConfig()
		if err != nil {
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

// validateConfig validates the configuration
func (s *ConfigService) validateConfig(config *ClaudeConfig) error {
	// Basic validation
	if config.UserID == "" {
		return fmt.Errorf("userID is required")
	}
	if len(config.UserID) != 64 {
		return fmt.Errorf("userID must be 64 characters")
	}
	if config.InstallationID != "" && len(config.InstallationID) != 32 {
		return fmt.Errorf("installationID must be 32 characters")
	}
	if config.AutoUpdaterStatus != "" &&
		config.AutoUpdaterStatus != "enabled" &&
		config.AutoUpdaterStatus != "disabled" {
		return fmt.Errorf("autoUpdaterStatus must be 'enabled' or 'disabled'")
	}
	if config.NumStartups < 0 {
		return fmt.Errorf("numStartups cannot be negative")
	}

	return nil
}

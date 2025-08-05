package tests

import (
	"context"
	"testing"

	"ccany/ent"
	"ccany/internal/claudecode"
	"ccany/internal/config"
	"ccany/internal/crypto"
	"ccany/internal/database"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database
func setupTestDB(t *testing.T) (*ent.Client, *crypto.CryptoService) {
	// Use in-memory SQLite for testing with proper parameters
	client, err := ent.Open("sqlite3", "file:test.db?mode=memory&cache=shared&_fk=1&_pragma=foreign_keys(1)")
	require.NoError(t, err)

	// Run migrations
	err = client.Schema.Create(context.Background())
	require.NoError(t, err)

	// Create crypto service
	cryptoService := crypto.NewCryptoService("test-master-key-1234567890123456")

	return client, cryptoService
}

func TestClaudeCodeConfigService(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("InitializeConfig", func(t *testing.T) {
		client, cryptoService := setupTestDB(t)
		defer client.Close()

		configService := config.NewService(client, cryptoService)
		claudeConfigService := claudecode.NewConfigService(logger, configService, ctx)

		// Initialize configuration
		err := claudeConfigService.InitializeConfig()
		assert.NoError(t, err)

		// Check if user ID was created
		userID, err := configService.GetConfig(ctx, config.KeyClaudeCodeUserID)
		assert.NoError(t, err)
		assert.Len(t, userID, 64, "User ID should be 64 characters")

		// Check if installation ID was created
		installationID, err := configService.GetConfig(ctx, config.KeyClaudeCodeInstallationID)
		assert.NoError(t, err)
		assert.Len(t, installationID, 32, "Installation ID should be 32 characters")
	})

	t.Run("GetConfig", func(t *testing.T) {
		client, cryptoService := setupTestDB(t)
		defer client.Close()

		configService := config.NewService(client, cryptoService)

		// Initialize default configs
		err := configService.InitializeDefaultConfigs(ctx)
		require.NoError(t, err)

		// Set some Claude Code specific configs (userID must be 64 characters)
		testUserID := generateTestString(64)
		err = configService.SetConfig(ctx, config.KeyClaudeCodeUserID, testUserID, false)
		require.NoError(t, err)

		err = configService.SetConfig(ctx, config.KeyClaudeCodeNumStartups, "10", false)
		require.NoError(t, err)

		claudeConfigService := claudecode.NewConfigService(logger, configService, ctx)

		// Get configuration
		cfg, err := claudeConfigService.GetConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, testUserID, cfg.UserID)
		assert.Equal(t, 10, cfg.NumStartups)
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		client, cryptoService := setupTestDB(t)
		defer client.Close()

		configService := config.NewService(client, cryptoService)
		claudeConfigService := claudecode.NewConfigService(logger, configService, ctx)

		// Initialize first
		err := claudeConfigService.InitializeConfig()
		require.NoError(t, err)

		// Create a config to update (with proper length IDs)
		cfg := &claudecode.ClaudeConfig{
			NumStartups:            100,
			AutoUpdaterStatus:      "disabled",
			UserID:                 generateTestString(64),
			HasCompletedOnboarding: true,
			LastOnboardingVersion:  "2.0.0",
			InstallationID:         generateTestString(32),
			TelemetryEnabled:       false,
			AnalyticsEnabled:       false,
			CrashReportingEnabled:  false,
		}

		// Update configuration
		err = claudeConfigService.UpdateConfig(cfg)
		assert.NoError(t, err)

		// Verify update
		updatedCfg, err := claudeConfigService.GetConfig()
		assert.NoError(t, err)
		assert.Equal(t, 100, updatedCfg.NumStartups)
		assert.Equal(t, "disabled", updatedCfg.AutoUpdaterStatus)
		assert.False(t, updatedCfg.TelemetryEnabled)
	})

	t.Run("IncrementStartupCount", func(t *testing.T) {
		client, cryptoService := setupTestDB(t)
		defer client.Close()

		configService := config.NewService(client, cryptoService)
		claudeConfigService := claudecode.NewConfigService(logger, configService, ctx)

		// Initialize configuration
		err := claudeConfigService.InitializeConfig()
		require.NoError(t, err)

		// Get initial count
		cfg1, err := claudeConfigService.GetConfig()
		require.NoError(t, err)
		initialCount := cfg1.NumStartups

		// Increment startup count
		err = claudeConfigService.IncrementStartupCount()
		assert.NoError(t, err)

		// Verify increment
		cfg2, err := claudeConfigService.GetConfig()
		assert.NoError(t, err)
		assert.Equal(t, initialCount+1, cfg2.NumStartups)
	})
}

func TestRouterConfigManager(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("LoadConfig", func(t *testing.T) {
		client, cryptoService := setupTestDB(t)
		defer client.Close()

		configService := config.NewService(client, cryptoService)

		// Initialize default configs
		err := configService.InitializeDefaultConfigs(ctx)
		require.NoError(t, err)

		// Create router config manager
		manager := claudecode.NewRouterConfigManager(logger, configService, ctx)

		// Get configuration
		cfg := manager.GetConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.Default)
		assert.Equal(t, "claude-3-5-haiku-20241022", cfg.Background)
		assert.Equal(t, 60000, cfg.LongContextThreshold)
		assert.True(t, cfg.EnableWebSearchDetection)
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		client, cryptoService := setupTestDB(t)
		defer client.Close()

		configService := config.NewService(client, cryptoService)
		manager := claudecode.NewRouterConfigManager(logger, configService, ctx)

		// Create new config
		newConfig := &claudecode.RouterConfig{
			Default:                  "gpt-4",
			Background:               "gpt-3.5-turbo",
			Think:                    "gpt-4",
			LongContext:              "gpt-4-32k",
			WebSearch:                "gpt-4",
			LongContextThreshold:     80000,
			EnableWebSearchDetection: false,
			EnableToolUseDetection:   true,
			EnableDynamicRouting:     true,
		}

		// Update configuration
		err := manager.UpdateConfig(newConfig)
		assert.NoError(t, err)

		// Reload and verify
		err = manager.LoadConfig()
		assert.NoError(t, err)

		cfg := manager.GetConfig()
		assert.Equal(t, "gpt-4", cfg.Default)
		assert.Equal(t, "gpt-3.5-turbo", cfg.Background)
		assert.Equal(t, 80000, cfg.LongContextThreshold)
		assert.False(t, cfg.EnableWebSearchDetection)
	})
}

func TestDatabaseIntegration(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("FullIntegrationFlow", func(t *testing.T) {
		// Setup database
		db, err := database.NewTestDB()
		require.NoError(t, err)
		defer db.Close()

		// Create config service
		configService := config.NewService(db.Client, db.CryptoService)

		// Initialize default configs
		err = configService.InitializeDefaultConfigs(ctx)
		require.NoError(t, err)

		// Create Claude config service
		claudeConfigService := claudecode.NewConfigService(logger, configService, ctx)

		// Initialize Claude Code config
		err = claudeConfigService.InitializeConfig()
		assert.NoError(t, err)

		// Create router config manager
		routerManager := claudecode.NewRouterConfigManager(logger, configService, ctx)

		// Get router config
		routerConfig := routerManager.GetConfig()
		assert.NotNil(t, routerConfig)

		// Test configuration persistence
		testValue := "test-model-xyz"
		err = configService.SetConfig(ctx, config.KeyRouterDefault, testValue, false)
		assert.NoError(t, err)

		// Reload and verify
		err = routerManager.LoadConfig()
		assert.NoError(t, err)

		reloadedConfig := routerManager.GetConfig()
		assert.Equal(t, testValue, reloadedConfig.Default)
	})
}

// Helper function to generate random string for testing
func generateTestString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

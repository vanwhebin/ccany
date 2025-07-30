package claudecode

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"ccany/internal/config"

	"github.com/sirupsen/logrus"
)

// RouterConfig represents the router configuration
type RouterConfig struct {
	// Model configurations for different scenarios
	Default     string `json:"default" yaml:"default"`
	Background  string `json:"background" yaml:"background"`
	Think       string `json:"think" yaml:"think"`
	LongContext string `json:"longContext" yaml:"longContext"`
	WebSearch   string `json:"webSearch" yaml:"webSearch"`

	// Token thresholds
	LongContextThreshold int `json:"longContextThreshold" yaml:"longContextThreshold"`

	// Feature flags
	EnableWebSearchDetection bool `json:"enableWebSearchDetection" yaml:"enableWebSearchDetection"`
	EnableToolUseDetection   bool `json:"enableToolUseDetection" yaml:"enableToolUseDetection"`
	EnableDynamicRouting     bool `json:"enableDynamicRouting" yaml:"enableDynamicRouting"`
}

// RouteRequest contains request information for routing decisions
type RouteRequest struct {
	Model      string                 `json:"model"`
	Messages   []interface{}          `json:"messages"`
	System     interface{}            `json:"system"`
	Tools      []interface{}          `json:"tools"`
	TokenCount int                    `json:"tokenCount"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// RouterConfigManager manages router configurations using database
type RouterConfigManager struct {
	mu            sync.RWMutex
	logger        *logrus.Logger
	configService *config.Service
	ctx           context.Context
	config        *RouterConfig
}

// NewRouterConfigManager creates a new router configuration manager
func NewRouterConfigManager(logger *logrus.Logger, configService *config.Service, ctx context.Context) *RouterConfigManager {
	manager := &RouterConfigManager{
		logger:        logger,
		configService: configService,
		ctx:           ctx,
	}

	// Load initial configuration
	if err := manager.LoadConfig(); err != nil {
		logger.WithError(err).Warn("Failed to load router configuration, using defaults")
		manager.config = DefaultRouterConfig()
	}

	return manager
}

// LoadConfig loads router configuration from database
func (m *RouterConfigManager) LoadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load configuration values from database
	defaultModel, _ := m.configService.GetConfig(m.ctx, config.KeyRouterDefault)
	backgroundModel, _ := m.configService.GetConfig(m.ctx, config.KeyRouterBackground)
	thinkModel, _ := m.configService.GetConfig(m.ctx, config.KeyRouterThink)
	longContextModel, _ := m.configService.GetConfig(m.ctx, config.KeyRouterLongContext)
	webSearchModel, _ := m.configService.GetConfig(m.ctx, config.KeyRouterWebSearch)

	longContextThresholdStr, _ := m.configService.GetConfig(m.ctx, config.KeyRouterLongContextThreshold)
	longContextThreshold, _ := strconv.Atoi(longContextThresholdStr)

	enableWebSearchStr, _ := m.configService.GetConfig(m.ctx, config.KeyRouterEnableWebSearchDetection)
	enableWebSearch, _ := strconv.ParseBool(enableWebSearchStr)

	enableToolUseStr, _ := m.configService.GetConfig(m.ctx, config.KeyRouterEnableToolUseDetection)
	enableToolUse, _ := strconv.ParseBool(enableToolUseStr)

	enableDynamicStr, _ := m.configService.GetConfig(m.ctx, config.KeyRouterEnableDynamicRouting)
	enableDynamic, _ := strconv.ParseBool(enableDynamicStr)

	// Create config with values from database
	m.config = &RouterConfig{
		Default:                  defaultModel,
		Background:               backgroundModel,
		Think:                    thinkModel,
		LongContext:              longContextModel,
		WebSearch:                webSearchModel,
		LongContextThreshold:     longContextThreshold,
		EnableWebSearchDetection: enableWebSearch,
		EnableToolUseDetection:   enableToolUse,
		EnableDynamicRouting:     enableDynamic,
	}

	m.logger.WithField("config", m.config).Info("Loaded router configuration from database")
	return nil
}

// GetConfig returns the current router configuration
func (m *RouterConfigManager) GetConfig() RouterConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.config
}

// UpdateConfig updates the router configuration in database
func (m *RouterConfigManager) UpdateConfig(cfg *RouterConfig) error {
	// Save each field to database
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterDefault, cfg.Default, false); err != nil {
		return fmt.Errorf("failed to save default model: %w", err)
	}
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterBackground, cfg.Background, false); err != nil {
		return fmt.Errorf("failed to save background model: %w", err)
	}
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterThink, cfg.Think, false); err != nil {
		return fmt.Errorf("failed to save think model: %w", err)
	}
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterLongContext, cfg.LongContext, false); err != nil {
		return fmt.Errorf("failed to save long context model: %w", err)
	}
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterWebSearch, cfg.WebSearch, false); err != nil {
		return fmt.Errorf("failed to save web search model: %w", err)
	}
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterLongContextThreshold, strconv.Itoa(cfg.LongContextThreshold), false); err != nil {
		return fmt.Errorf("failed to save long context threshold: %w", err)
	}
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterEnableWebSearchDetection, strconv.FormatBool(cfg.EnableWebSearchDetection), false); err != nil {
		return fmt.Errorf("failed to save web search detection flag: %w", err)
	}
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterEnableToolUseDetection, strconv.FormatBool(cfg.EnableToolUseDetection), false); err != nil {
		return fmt.Errorf("failed to save tool use detection flag: %w", err)
	}
	if err := m.configService.SetConfig(m.ctx, config.KeyRouterEnableDynamicRouting, strconv.FormatBool(cfg.EnableDynamicRouting), false); err != nil {
		return fmt.Errorf("failed to save dynamic routing flag: %w", err)
	}

	m.mu.Lock()
	m.config = cfg
	m.mu.Unlock()

	return nil
}

// WebSearchRouterStrategy handles web search tool detection
type WebSearchRouterStrategy struct {
	model   string
	enabled bool
}

func (s *WebSearchRouterStrategy) ShouldApply(req *RouteRequest) bool {
	if !s.enabled || s.model == "" {
		return false
	}

	// Check if any tool has web_search type
	for _, tool := range req.Tools {
		if toolMap, ok := tool.(map[string]interface{}); ok {
			if toolType, exists := toolMap["type"]; exists {
				if typeStr, ok := toolType.(string); ok && typeStr == "web_search" {
					return true
				}
			}
			// Also check for name containing web_search
			if toolName, exists := toolMap["name"]; exists {
				if nameStr, ok := toolName.(string); ok && nameStr == "web_search" {
					return true
				}
			}
		}
	}
	return false
}

func (s *WebSearchRouterStrategy) GetModel() string {
	return s.model
}

func (s *WebSearchRouterStrategy) GetReason() string {
	return "web_search_tool"
}

// DefaultRouterConfig returns a default router configuration
func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		Default:                  "claude-3-5-sonnet-20241022",
		Background:               "claude-3-5-haiku-20241022",
		Think:                    "claude-3-5-sonnet-20241022",
		LongContext:              "claude-3-5-sonnet-20241022",
		WebSearch:                "claude-3-5-sonnet-20241022",
		LongContextThreshold:     60000,
		EnableWebSearchDetection: true,
		EnableToolUseDetection:   true,
		EnableDynamicRouting:     true,
	}
}

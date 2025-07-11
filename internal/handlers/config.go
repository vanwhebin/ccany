package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"ccany/internal/app"
	"ccany/internal/config"
	"ccany/internal/i18n"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ConfigHandler configuration management handler
type ConfigHandler struct {
	configManager *app.ConfigManager
	logger        *logrus.Logger
}

// NewConfigHandler creates configuration management handler
func NewConfigHandler(configManager *app.ConfigManager, logger *logrus.Logger) *ConfigHandler {
	return &ConfigHandler{
		configManager: configManager,
		logger:        logger,
	}
}

// ConfigItem configuration item response structure
type ConfigItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Category    string      `json:"category"`
	Type        string      `json:"type"`
	IsEncrypted bool        `json:"is_encrypted"`
	IsRequired  bool        `json:"is_required"`
	Description string      `json:"description"`
}

// ConfigUpdateRequest configuration update request
type ConfigUpdateRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// GetAllConfigs gets all configurations - GET /admin/configs
func (h *ConfigHandler) GetAllConfigs(c *gin.Context) {
	h.logger.Info("Getting all configs")
	configs, err := h.configManager.GetAllConfigs()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get all configs")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": i18n.T(c, "errors.config_error"),
		})
		return
	}

	h.logger.WithField("config_count", len(configs)).Info("Retrieved configs")

	// convert to response format
	var configItems []ConfigItem
	for key, valueData := range configs {
		var displayValue interface{}

		// handle complex structure returned from service
		if valueMap, ok := valueData.(map[string]interface{}); ok {
			if isEncrypted, exists := valueMap["is_encrypted"]; exists {
				if encrypted, ok := isEncrypted.(bool); ok && encrypted {
					// encrypted configuration, check if already set
					if isSet, exists := valueMap["is_set"]; exists {
						if set, ok := isSet.(bool); ok && set {
							displayValue = "***masked***"
						} else {
							displayValue = ""
						}
					} else {
						displayValue = ""
					}
				} else {
					// non-encrypted configuration, get value directly
					if value, exists := valueMap["value"]; exists {
						displayValue = value
					} else {
						displayValue = ""
					}
				}
			} else {
				// non-encrypted configuration, get value directly
				if value, exists := valueMap["value"]; exists {
					displayValue = value
				} else {
					displayValue = ""
				}
			}
		} else {
			// if not expected structure, use original value directly
			displayValue = valueData
		}

		configItems = append(configItems, ConfigItem{
			Key:         key,
			Value:       displayValue,
			Category:    h.getConfigCategory(key),
			Type:        h.getConfigType(key),
			IsEncrypted: h.isEncryptedConfig(key),
			IsRequired:  h.isRequiredConfig(key),
			Description: h.getConfigDescription(key),
		})
	}

	h.logger.WithField("config_items_count", len(configItems)).Info("Converted config items")

	c.JSON(http.StatusOK, gin.H{
		"configs": configItems,
	})
}

// GetConfig gets single configuration - GET /admin/configs/:key
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": i18n.T(c, "errors.missing_parameter"),
		})
		return
	}

	value, err := h.configManager.GetConfigValue(key)
	if err != nil {
		h.logger.WithError(err).WithField("key", key).Error("Failed to get config")
		c.JSON(http.StatusNotFound, gin.H{
			"error": i18n.T(c, "errors.config_not_found"),
		})
		return
	}

	configItem := ConfigItem{
		Key:         key,
		Value:       h.maskSensitiveValue(key, value),
		Category:    h.getConfigCategory(key),
		Type:        h.getConfigType(key),
		IsEncrypted: h.isEncryptedConfig(key),
		IsRequired:  h.isRequiredConfig(key),
		Description: h.getConfigDescription(key),
	}

	c.JSON(http.StatusOK, gin.H{
		"config": configItem,
	})
}

// UpdateConfig updates configuration - PUT /admin/configs/:key
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Configuration key is required",
		})
		return
	}

	var req ConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// verify configuration key consistency
	if req.Key != key {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Configuration key mismatch",
		})
		return
	}

	// validate configuration value
	if err := h.validateConfigValue(key, req.Value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// update configuration
	encrypted := h.isEncryptedConfig(key)
	if err := h.configManager.UpdateConfig(key, req.Value, encrypted); err != nil {
		h.logger.WithError(err).WithField("key", key).Error("Failed to update config")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update configuration",
		})
		return
	}

	h.logger.WithField("key", key).Info("Configuration updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"key":     key,
	})
}

// TestConfig tests configuration - POST /admin/configs/test
func (h *ConfigHandler) TestConfig(c *gin.Context) {
	// validate current configuration
	if err := h.configManager.ValidateConfig(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":  false,
			"error":  err.Error(),
			"status": "Configuration validation failed",
		})
		return
	}

	// check if configured
	configured, err := h.configManager.IsConfigured()
	if err != nil {
		h.logger.WithError(err).Error("Failed to check configuration status")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check configuration status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"configured": configured,
		"status":     "Configuration is valid",
	})
}

// TestAPIKey tests API key - POST /admin/config/test-api-key
func (h *ConfigHandler) TestAPIKey(c *gin.Context) {
	h.logger.Info("Testing API key")

	// get current configuration
	cfg, err := h.configManager.GetConfig()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get current config")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get current configuration",
		})
		return
	}

	// check if any API key is configured
	if cfg.OpenAIAPIKey == "" && cfg.ClaudeAPIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No API keys configured. Please configure at least one API key (OpenAI or Claude)",
		})
		return
	}

	// test API key
	results := h.testAPIConnectivity(cfg)

	// check if any test was successful
	hasSuccess := false
	for _, result := range results {
		if result.Success {
			hasSuccess = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": hasSuccess,
		"results": results,
		"message": func() string {
			if hasSuccess {
				return "API key test successful"
			}
			return "API key test failed"
		}(),
	})
}

// APITestResult API test result structure
type APITestResult struct {
	Service      string `json:"service"`
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	Model        string `json:"model,omitempty"`
	ResponseTime string `json:"response_time,omitempty"`
	Error        string `json:"error,omitempty"`
}

// testAPIConnectivity tests API connectivity
func (h *ConfigHandler) testAPIConnectivity(cfg *config.Config) []APITestResult {
	var results []APITestResult

	// test OpenAI API
	if cfg.OpenAIAPIKey != "" {
		result := h.testOpenAIAPI(cfg)
		results = append(results, result)
	}

	// also test Claude API if configured
	if cfg.ClaudeAPIKey != "" {
		result := h.testClaudeAPI(cfg)
		results = append(results, result)
	}

	return results
}

// testOpenAIAPI tests OpenAI API
func (h *ConfigHandler) testOpenAIAPI(cfg *config.Config) APITestResult {
	startTime := time.Now()

	// create simple test request
	testPayload := map[string]interface{}{
		"model":      cfg.BigModel,
		"messages":   []map[string]string{{"role": "user", "content": "Hello"}},
		"max_tokens": 1,
	}

	// create HTTP client
	client := &http.Client{
		Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
	}

	// create request
	jsonData, _ := json.Marshal(testPayload)
	req, err := http.NewRequest("POST", cfg.OpenAIBaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return APITestResult{
			Service: "OpenAI API",
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	// set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return APITestResult{
			Service: "OpenAI API",
			Success: false,
			Error:   fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.logger.WithError(err).Error("Failed to close response body")
		}
	}()

	responseTime := time.Since(startTime)

	// check response status
	if resp.StatusCode == http.StatusOK {
		return APITestResult{
			Service:      "OpenAI API",
			Success:      true,
			Message:      "API key valid and responsive",
			Model:        cfg.BigModel,
			ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1000000),
		}
	} else {
		// read error response
		body, _ := io.ReadAll(resp.Body)
		return APITestResult{
			Service: "OpenAI API",
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}
}

// testClaudeAPI tests Claude API
func (h *ConfigHandler) testClaudeAPI(cfg *config.Config) APITestResult {
	startTime := time.Now()

	// create simple test request
	testPayload := map[string]interface{}{
		"model":      cfg.BigModel,
		"max_tokens": 1,
		"messages":   []map[string]string{{"role": "user", "content": "Hello"}},
	}

	// create HTTP client
	client := &http.Client{
		Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
	}

	// create request
	jsonData, _ := json.Marshal(testPayload)
	req, err := http.NewRequest("POST", cfg.ClaudeBaseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return APITestResult{
			Service: "Claude API",
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	// set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.ClaudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return APITestResult{
			Service: "Claude API",
			Success: false,
			Error:   fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.logger.WithError(err).Error("Failed to close response body")
		}
	}()

	responseTime := time.Since(startTime)

	// check response status
	if resp.StatusCode == http.StatusOK {
		return APITestResult{
			Service:      "Claude API",
			Success:      true,
			Message:      "API key valid and responsive",
			Model:        cfg.BigModel,
			ResponseTime: fmt.Sprintf("%.2fms", float64(responseTime.Nanoseconds())/1000000),
		}
	} else {
		// read error response
		body, _ := io.ReadAll(resp.Body)
		return APITestResult{
			Service: "Claude API",
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}
}

// GetConfigStats gets configuration statistics - GET /admin/configs/stats
func (h *ConfigHandler) GetConfigStats(c *gin.Context) {
	configs, err := h.configManager.GetAllConfigs()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get config stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get configuration statistics",
		})
		return
	}

	var stats = struct {
		Total      int `json:"total"`
		Configured int `json:"configured"`
		Encrypted  int `json:"encrypted"`
		Required   int `json:"required"`
	}{}

	for key, value := range configs {
		stats.Total++
		if value != nil && value != "" {
			stats.Configured++
		}
		if h.isEncryptedConfig(key) {
			stats.Encrypted++
		}
		if h.isRequiredConfig(key) {
			stats.Required++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// maskSensitiveValue masks sensitive values
func (h *ConfigHandler) maskSensitiveValue(key string, value interface{}) interface{} {
	if h.isEncryptedConfig(key) {
		if strValue, ok := value.(string); ok && strValue != "" {
			return "***masked***"
		}
	}
	return value
}

// getConfigCategory gets configuration category
func (h *ConfigHandler) getConfigCategory(key string) string {
	switch key {
	case config.KeyOpenAIAPIKey, config.KeyOpenAIBaseURL, config.KeyClaudeAPIKey, config.KeyClaudeBaseURL:
		return "api"
	case config.KeyServerPort, config.KeyLogLevel:
		return "server"
	case config.KeyMaxTokens, config.KeyTemperature, config.KeyStreamEnabled:
		return "model"
	default:
		return "other"
	}
}

// getConfigType gets configuration type
func (h *ConfigHandler) getConfigType(key string) string {
	switch key {
	case config.KeyServerPort, config.KeyMaxTokens:
		return "int"
	case config.KeyTemperature:
		return "float"
	case config.KeyStreamEnabled:
		return "bool"
	default:
		return "string"
	}
}

// isEncryptedConfig checks if configuration is encrypted
func (h *ConfigHandler) isEncryptedConfig(key string) bool {
	switch key {
	case config.KeyOpenAIAPIKey, config.KeyClaudeAPIKey:
		return true
	default:
		return false
	}
}

// isRequiredConfig checks if configuration is required
func (h *ConfigHandler) isRequiredConfig(key string) bool {
	switch key {
	case config.KeyOpenAIAPIKey, config.KeyClaudeAPIKey:
		return true
	default:
		return false
	}
}

// getConfigDescription gets configuration description
func (h *ConfigHandler) getConfigDescription(key string) string {
	descriptions := map[string]string{
		config.KeyOpenAIAPIKey:  "OpenAI API key for accessing OpenAI services",
		config.KeyOpenAIBaseURL: "OpenAI API base URL",
		config.KeyClaudeAPIKey:  "Claude API key for accessing Anthropic services",
		config.KeyClaudeBaseURL: "Claude API base URL",
		config.KeyServerPort:    "Server listening port",
		config.KeyLogLevel:      "Log level (DEBUG/INFO/WARN/ERROR)",
		config.KeyMaxTokens:     "Maximum token limit",
		config.KeyTemperature:   "Model temperature parameter (0.0-1.0)",
		config.KeyStreamEnabled: "Whether to enable streaming response",
	}

	if desc, exists := descriptions[key]; exists {
		return desc
	}
	return "Configuration item"
}

// validateConfigValue validates configuration value
func (h *ConfigHandler) validateConfigValue(key, value string) error {
	switch key {
	case config.KeyServerPort:
		if port, err := strconv.Atoi(value); err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("invalid port number: must be between 1 and 65535")
		}
	case config.KeyMaxTokens:
		if tokens, err := strconv.Atoi(value); err != nil || tokens < 1 {
			return fmt.Errorf("invalid token limit: must be a positive integer")
		}
	case config.KeyTemperature:
		if temp, err := strconv.ParseFloat(value, 64); err != nil || temp < 0 || temp > 1 {
			return fmt.Errorf("temperature must be between 0 and 1")
		}
	case config.KeyStreamEnabled:
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("invalid boolean value: must be true or false")
		}
	}
	return nil
}

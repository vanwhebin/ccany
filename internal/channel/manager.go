package channel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ccany/ent"
)

// Channel represents a API channel configuration
type Channel struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Provider        string                 `json:"provider"` // openai, anthropic, gemini
	BaseURL         string                 `json:"base_url"`
	APIKey          string                 `json:"api_key"`
	CustomKey       string                 `json:"custom_key"` // User-facing key
	Timeout         int                    `json:"timeout"`
	MaxRetries      int                    `json:"max_retries"`
	Enabled         bool                   `json:"enabled"`
	Weight          int                    `json:"weight"`   // For load balancing
	Priority        int                    `json:"priority"` // Higher priority channels are preferred
	ModelsMapping   map[string]string      `json:"models_mapping,omitempty"`
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastUsedAt      *time.Time             `json:"last_used_at,omitempty"`
	RequestCount    int64                  `json:"request_count"`
	ErrorCount      int64                  `json:"error_count"`
	SuccessRate     float64                `json:"success_rate"`
	TotalTokens     int64                  `json:"total_tokens"`
	AvgResponseTime float64                `json:"avg_response_time"`
}

// ChannelManager manages API channels
type ChannelManager struct {
	db     *ent.Client
	mu     sync.RWMutex
	cache  map[string]*Channel // Cache for fast lookups
	keyMap map[string]string   // Custom key to channel ID mapping
}

// NewChannelManager creates a new channel manager
func NewChannelManager(db *ent.Client) *ChannelManager {
	cm := &ChannelManager{
		db:     db,
		cache:  make(map[string]*Channel),
		keyMap: make(map[string]string),
	}

	// Load existing channels into cache
	cm.loadChannels()

	return cm
}

// loadChannels loads all channels from database into cache
func (cm *ChannelManager) loadChannels() error {
	channels, err := cm.db.Channel.Query().All(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load channels: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Clear existing cache
	cm.cache = make(map[string]*Channel)
	cm.keyMap = make(map[string]string)

	// Load channels into cache
	for _, ch := range channels {
		channel := &Channel{
			ID:              ch.ID,
			Name:            ch.Name,
			Provider:        ch.Provider,
			BaseURL:         ch.BaseURL,
			APIKey:          ch.APIKey,
			CustomKey:       ch.CustomKey,
			Timeout:         ch.Timeout,
			MaxRetries:      ch.MaxRetries,
			Enabled:         ch.Enabled,
			Weight:          ch.Weight,
			Priority:        ch.Priority,
			ModelsMapping:   ch.ModelsMapping,
			Capabilities:    ch.Capabilities,
			CreatedAt:       ch.CreatedAt,
			UpdatedAt:       ch.UpdatedAt,
			LastUsedAt:      ch.LastUsedAt,
			RequestCount:    ch.RequestCount,
			ErrorCount:      ch.ErrorCount,
			SuccessRate:     ch.SuccessRate,
			TotalTokens:     ch.TotalTokens,
			AvgResponseTime: ch.AvgResponseTime,
		}

		cm.cache[ch.ID] = channel
		if ch.CustomKey != "" {
			cm.keyMap[ch.CustomKey] = ch.ID
		}
	}

	return nil
}

// CreateChannel creates a new channel
func (cm *ChannelManager) CreateChannel(ctx context.Context, req *CreateChannelRequest) (*Channel, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if custom key already exists
	if req.CustomKey != "" {
		if _, exists := cm.GetChannelByCustomKey(req.CustomKey); exists {
			return nil, fmt.Errorf("custom key already exists: %s", req.CustomKey)
		}
	}

	// Create in database
	ch, err := cm.db.Channel.Create().
		SetName(req.Name).
		SetProvider(req.Provider).
		SetBaseURL(req.BaseURL).
		SetAPIKey(req.APIKey).
		SetCustomKey(req.CustomKey).
		SetTimeout(req.Timeout).
		SetMaxRetries(req.MaxRetries).
		SetEnabled(true).
		SetWeight(req.Weight).
		SetPriority(req.Priority).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Create channel object
	channel := &Channel{
		ID:           ch.ID,
		Name:         ch.Name,
		Provider:     ch.Provider,
		BaseURL:      ch.BaseURL,
		APIKey:       ch.APIKey,
		CustomKey:    ch.CustomKey,
		Timeout:      ch.Timeout,
		MaxRetries:   ch.MaxRetries,
		Enabled:      ch.Enabled,
		Weight:       ch.Weight,
		Priority:     ch.Priority,
		CreatedAt:    ch.CreatedAt,
		UpdatedAt:    ch.UpdatedAt,
		RequestCount: 0,
		ErrorCount:   0,
	}

	// Update cache
	cm.mu.Lock()
	cm.cache[ch.ID] = channel
	if ch.CustomKey != "" {
		cm.keyMap[ch.CustomKey] = ch.ID
	}
	cm.mu.Unlock()

	return channel, nil
}

// GetChannel gets a channel by ID
func (cm *ChannelManager) GetChannel(id string) (*Channel, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channel, exists := cm.cache[id]
	return channel, exists
}

// GetChannelByCustomKey gets a channel by custom key
func (cm *ChannelManager) GetChannelByCustomKey(customKey string) (*Channel, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channelID, exists := cm.keyMap[customKey]
	if !exists {
		return nil, false
	}

	channel, exists := cm.cache[channelID]
	return channel, exists
}

// GetChannelsByProvider gets all enabled channels for a provider
func (cm *ChannelManager) GetChannelsByProvider(provider string) []*Channel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var channels []*Channel
	for _, ch := range cm.cache {
		if ch.Provider == provider && ch.Enabled {
			channels = append(channels, ch)
		}
	}

	// Sort by priority (higher first) then by weight, then by success rate
	for i := 0; i < len(channels)-1; i++ {
		for j := i + 1; j < len(channels); j++ {
			if channels[i].Priority < channels[j].Priority ||
				(channels[i].Priority == channels[j].Priority && channels[i].Weight < channels[j].Weight) ||
				(channels[i].Priority == channels[j].Priority && channels[i].Weight == channels[j].Weight && channels[i].SuccessRate < channels[j].SuccessRate) {
				channels[i], channels[j] = channels[j], channels[i]
			}
		}
	}

	return channels
}

// RouteRequest selects the best channel for a request based on load balancing and health
func (cm *ChannelManager) RouteRequest(provider string, preferredChannelID string) (*Channel, error) {
	// If a specific channel is preferred and available, use it
	if preferredChannelID != "" {
		if channel, exists := cm.GetChannel(preferredChannelID); exists && channel.Enabled {
			return channel, nil
		}
	}

	// Get all available channels for the provider
	channels := cm.GetChannelsByProvider(provider)
	if len(channels) == 0 {
		return nil, fmt.Errorf("no enabled channels found for provider: %s", provider)
	}

	// Simple weighted round-robin with health consideration
	bestChannel := cm.selectBestChannel(channels)
	if bestChannel == nil {
		return nil, fmt.Errorf("no healthy channels available for provider: %s", provider)
	}

	return bestChannel, nil
}

// selectBestChannel implements intelligent channel selection
func (cm *ChannelManager) selectBestChannel(channels []*Channel) *Channel {
	if len(channels) == 0 {
		return nil
	}

	// Filter out channels with very low success rate (< 0.5)
	healthyChannels := make([]*Channel, 0, len(channels))
	for _, ch := range channels {
		if ch.SuccessRate >= 0.5 || ch.RequestCount < 10 { // Give new channels a chance
			healthyChannels = append(healthyChannels, ch)
		}
	}

	if len(healthyChannels) == 0 {
		// If no healthy channels, return the best of unhealthy ones
		return channels[0]
	}

	// Calculate scores based on priority, weight, and success rate
	bestChannel := healthyChannels[0]
	bestScore := cm.calculateChannelScore(bestChannel)

	for _, ch := range healthyChannels[1:] {
		score := cm.calculateChannelScore(ch)
		if score > bestScore {
			bestScore = score
			bestChannel = ch
		}
	}

	return bestChannel
}

// calculateChannelScore calculates a score for channel selection
func (cm *ChannelManager) calculateChannelScore(ch *Channel) float64 {
	// Base score from priority and weight
	baseScore := float64(ch.Priority*10 + ch.Weight)

	// Boost score based on success rate
	successBoost := ch.SuccessRate * 50

	// Penalize channels with recent errors
	errorPenalty := 0.0
	if ch.RequestCount > 0 {
		errorRate := float64(ch.ErrorCount) / float64(ch.RequestCount)
		errorPenalty = errorRate * 20
	}

	// Prefer channels with reasonable response times
	responsePenalty := 0.0
	if ch.AvgResponseTime > 5.0 { // Penalize if avg response > 5 seconds
		responsePenalty = (ch.AvgResponseTime - 5.0) * 2
	}

	return baseScore + successBoost - errorPenalty - responsePenalty
}

// UpdateChannelMetrics updates channel performance metrics
func (cm *ChannelManager) UpdateChannelMetrics(ctx context.Context, channelID string, responseTime float64, tokenCount int64, success bool) error {
	ch, exists := cm.GetChannel(channelID)
	if !exists {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Update request count
	newRequestCount := ch.RequestCount + 1
	newErrorCount := ch.ErrorCount
	if !success {
		newErrorCount++
	}

	// Calculate new success rate
	newSuccessRate := float64(newRequestCount-newErrorCount) / float64(newRequestCount)

	// Update average response time (exponential moving average)
	alpha := 0.1 // Smoothing factor
	newAvgResponseTime := ch.AvgResponseTime
	if ch.RequestCount == 0 {
		newAvgResponseTime = responseTime
	} else {
		newAvgResponseTime = alpha*responseTime + (1-alpha)*ch.AvgResponseTime
	}

	// Update database
	_, err := cm.db.Channel.UpdateOneID(channelID).
		SetLastUsedAt(time.Now()).
		SetRequestCount(newRequestCount).
		SetErrorCount(newErrorCount).
		SetSuccessRate(newSuccessRate).
		SetTotalTokens(ch.TotalTokens + tokenCount).
		SetAvgResponseTime(newAvgResponseTime).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update channel metrics: %w", err)
	}

	// Update cache
	if cachedCh, exists := cm.cache[channelID]; exists {
		cachedCh.LastUsedAt = &[]time.Time{time.Now()}[0]
		cachedCh.RequestCount = newRequestCount
		cachedCh.ErrorCount = newErrorCount
		cachedCh.SuccessRate = newSuccessRate
		cachedCh.TotalTokens = ch.TotalTokens + tokenCount
		cachedCh.AvgResponseTime = newAvgResponseTime
	}

	return nil
}

// GetAllChannels gets all channels
func (cm *ChannelManager) GetAllChannels() []*Channel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var channels []*Channel
	for _, ch := range cm.cache {
		channels = append(channels, ch)
	}

	return channels
}

// UpdateChannel updates a channel
func (cm *ChannelManager) UpdateChannel(ctx context.Context, id string, req *UpdateChannelRequest) (*Channel, error) {
	// Get existing channel
	existingCh, exists := cm.GetChannel(id)
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", id)
	}

	// Build update query
	update := cm.db.Channel.UpdateOneID(id)

	if req.Name != "" {
		update = update.SetName(req.Name)
	}
	if req.BaseURL != "" {
		update = update.SetBaseURL(req.BaseURL)
	}
	if req.APIKey != "" {
		update = update.SetAPIKey(req.APIKey)
	}
	if req.CustomKey != "" {
		// Check if new custom key already exists (but not for this channel)
		if otherCh, exists := cm.GetChannelByCustomKey(req.CustomKey); exists && otherCh.ID != id {
			return nil, fmt.Errorf("custom key already exists: %s", req.CustomKey)
		}
		update = update.SetCustomKey(req.CustomKey)
	}
	if req.Timeout > 0 {
		update = update.SetTimeout(req.Timeout)
	}
	if req.MaxRetries > 0 {
		update = update.SetMaxRetries(req.MaxRetries)
	}
	if req.Enabled != nil {
		update = update.SetEnabled(*req.Enabled)
	}
	if req.Weight > 0 {
		update = update.SetWeight(req.Weight)
	}
	if req.Priority > 0 {
		update = update.SetPriority(req.Priority)
	}

	// Execute update
	ch, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	// Update cache
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Remove old custom key mapping if changed
	if existingCh.CustomKey != "" && existingCh.CustomKey != ch.CustomKey {
		delete(cm.keyMap, existingCh.CustomKey)
	}

	// Update channel in cache
	channel := &Channel{
		ID:           ch.ID,
		Name:         ch.Name,
		Provider:     ch.Provider,
		BaseURL:      ch.BaseURL,
		APIKey:       ch.APIKey,
		CustomKey:    ch.CustomKey,
		Timeout:      ch.Timeout,
		MaxRetries:   ch.MaxRetries,
		Enabled:      ch.Enabled,
		Weight:       ch.Weight,
		Priority:     ch.Priority,
		CreatedAt:    ch.CreatedAt,
		UpdatedAt:    ch.UpdatedAt,
		LastUsedAt:   ch.LastUsedAt,
		RequestCount: ch.RequestCount,
		ErrorCount:   ch.ErrorCount,
	}

	cm.cache[ch.ID] = channel
	if ch.CustomKey != "" {
		cm.keyMap[ch.CustomKey] = ch.ID
	}

	return channel, nil
}

// DeleteChannel deletes a channel
func (cm *ChannelManager) DeleteChannel(ctx context.Context, id string) error {
	// Get existing channel for cleanup
	existingCh, exists := cm.GetChannel(id)
	if !exists {
		return fmt.Errorf("channel not found: %s", id)
	}

	// Delete from database
	err := cm.db.Channel.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	// Update cache
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.cache, id)
	if existingCh.CustomKey != "" {
		delete(cm.keyMap, existingCh.CustomKey)
	}

	return nil
}

// RecordUsage records channel usage statistics
func (cm *ChannelManager) RecordUsage(ctx context.Context, channelID string, success bool) error {
	ch, exists := cm.GetChannel(channelID)
	if !exists {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	// Update database
	update := cm.db.Channel.UpdateOneID(channelID).
		SetLastUsedAt(time.Now()).
		SetRequestCount(ch.RequestCount + 1)

	if !success {
		update = update.SetErrorCount(ch.ErrorCount + 1)
	}

	updatedCh, err := update.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	// Update cache
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cachedCh, exists := cm.cache[channelID]; exists {
		cachedCh.LastUsedAt = updatedCh.LastUsedAt
		cachedCh.RequestCount = updatedCh.RequestCount
		cachedCh.ErrorCount = updatedCh.ErrorCount
	}

	return nil
}

// GetChannelStats returns channel statistics
func (cm *ChannelManager) GetChannelStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := make(map[string]interface{})
	providerCounts := make(map[string]int)
	totalChannels := 0
	enabledChannels := 0

	for _, ch := range cm.cache {
		totalChannels++
		providerCounts[ch.Provider]++
		if ch.Enabled {
			enabledChannels++
		}
	}

	stats["total_channels"] = totalChannels
	stats["enabled_channels"] = enabledChannels
	stats["provider_counts"] = providerCounts

	return stats
}

// RefreshCache refreshes the channel cache from database
func (cm *ChannelManager) RefreshCache() error {
	return cm.loadChannels()
}

// CreateChannelRequest represents a create channel request
type CreateChannelRequest struct {
	Name       string `json:"name" validate:"required"`
	Provider   string `json:"provider" validate:"required,oneof=openai anthropic gemini"`
	BaseURL    string `json:"base_url" validate:"required,url"`
	APIKey     string `json:"api_key" validate:"required"`
	CustomKey  string `json:"custom_key" validate:"required"`
	Timeout    int    `json:"timeout" validate:"min=1,max=300"`
	MaxRetries int    `json:"max_retries" validate:"min=0,max=10"`
	Weight     int    `json:"weight" validate:"min=1,max=100"`
	Priority   int    `json:"priority" validate:"min=1,max=10"`
}

// Validate validates the create request
func (req *CreateChannelRequest) Validate() error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if req.Provider != "openai" && req.Provider != "anthropic" && req.Provider != "gemini" {
		return fmt.Errorf("provider must be one of: openai, anthropic, gemini")
	}
	if req.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if req.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if req.CustomKey == "" {
		return fmt.Errorf("custom_key is required")
	}
	if req.Timeout <= 0 {
		req.Timeout = 30 // Default timeout
	}
	if req.MaxRetries < 0 {
		req.MaxRetries = 3 // Default retries
	}
	if req.Weight <= 0 {
		req.Weight = 1 // Default weight
	}
	if req.Priority <= 0 {
		req.Priority = 1 // Default priority
	}

	return nil
}

// UpdateChannelRequest represents an update channel request
type UpdateChannelRequest struct {
	Name       string `json:"name,omitempty"`
	BaseURL    string `json:"base_url,omitempty"`
	APIKey     string `json:"api_key,omitempty"`
	CustomKey  string `json:"custom_key,omitempty"`
	Timeout    int    `json:"timeout,omitempty"`
	MaxRetries int    `json:"max_retries,omitempty"`
	Enabled    *bool  `json:"enabled,omitempty"`
	Weight     int    `json:"weight,omitempty"`
	Priority   int    `json:"priority,omitempty"`
}

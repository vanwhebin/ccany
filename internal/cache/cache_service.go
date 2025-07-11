package cache

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CacheService cache service
type CacheService struct {
	cache  map[string]*CacheItem
	mutex  sync.RWMutex
	logger *logrus.Logger
	config *CacheConfig

	// cleanup goroutine control
	stopChan chan struct{}
	running  bool
}

// CacheItem cache item
type CacheItem struct {
	Key         string        `json:"key"`
	Value       interface{}   `json:"value"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpiresAt   time.Time     `json:"expires_at"`
	TTL         time.Duration `json:"ttl"`
	AccessCount int64         `json:"access_count"`
	LastAccess  time.Time     `json:"last_access"`
}

// CacheConfig cache configuration
type CacheConfig struct {
	// basic configuration
	DefaultTTL      time.Duration `json:"default_ttl"`
	MaxSize         int           `json:"max_size"`
	CleanupInterval time.Duration `json:"cleanup_interval"`

	// performance configuration
	EnableCompression bool `json:"enable_compression"`
	EnableMetrics     bool `json:"enable_metrics"`

	// policy configuration
	EvictionPolicy string `json:"eviction_policy"` // LRU, LFU, TTL
}

// CacheStats cache statistics
type CacheStats struct {
	TotalItems    int64   `json:"total_items"`
	HitCount      int64   `json:"hit_count"`
	MissCount     int64   `json:"miss_count"`
	HitRate       float64 `json:"hit_rate"`
	EvictionCount int64   `json:"eviction_count"`

	// memory usage
	MemoryUsage int64 `json:"memory_usage"`

	// time statistics
	CreatedAt   time.Time `json:"created_at"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// NewCacheService creates a new cache service
func NewCacheService(config *CacheConfig, logger *logrus.Logger) *CacheService {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cs := &CacheService{
		cache:    make(map[string]*CacheItem),
		logger:   logger,
		config:   config,
		stopChan: make(chan struct{}),
	}

	return cs
}

// DefaultCacheConfig default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:        5 * time.Minute,
		MaxSize:           1000,
		CleanupInterval:   1 * time.Minute,
		EnableCompression: false,
		EnableMetrics:     true,
		EvictionPolicy:    "LRU",
	}
}

// Start starts the cache service
func (cs *CacheService) Start(ctx context.Context) error {
	if cs.running {
		return nil
	}

	cs.running = true

	// start cleanup goroutine
	go cs.cleanupLoop(ctx)

	cs.logger.Info("Cache service started")
	return nil
}

// Stop stops the cache service
func (cs *CacheService) Stop() {
	if !cs.running {
		return
	}

	cs.running = false
	close(cs.stopChan)
	cs.logger.Info("Cache service stopped")
}

// Set sets cache item
func (cs *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if ttl == 0 {
		ttl = cs.config.DefaultTTL
	}

	// check cache size limit
	if len(cs.cache) >= cs.config.MaxSize {
		cs.evictItems(1)
	}

	now := time.Now()
	item := &CacheItem{
		Key:         key,
		Value:       value,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
		TTL:         ttl,
		AccessCount: 0,
		LastAccess:  now,
	}

	cs.cache[key] = item

	cs.logger.WithFields(logrus.Fields{
		"key": key,
		"ttl": ttl,
	}).Debug("Cache item set")

	return nil
}

// Get retrieves cache item
func (cs *CacheService) Get(key string) (interface{}, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	item, exists := cs.cache[key]
	if !exists {
		cs.logger.WithField("key", key).Debug("Cache miss")
		return nil, false
	}

	// check if expired
	if time.Now().After(item.ExpiresAt) {
		cs.logger.WithField("key", key).Debug("Cache item expired")
		delete(cs.cache, key)
		return nil, false
	}

	// update access statistics
	item.AccessCount++
	item.LastAccess = time.Now()

	cs.logger.WithField("key", key).Debug("Cache hit")
	return item.Value, true
}

// Delete deletes cache item
func (cs *CacheService) Delete(key string) bool {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if _, exists := cs.cache[key]; exists {
		delete(cs.cache, key)
		cs.logger.WithField("key", key).Debug("Cache item deleted")
		return true
	}

	return false
}

// Exists checks if cache item exists
func (cs *CacheService) Exists(key string) bool {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	item, exists := cs.cache[key]
	if !exists {
		return false
	}

	// check if expired
	if time.Now().After(item.ExpiresAt) {
		return false
	}

	return true
}

// Clear clears all cache
func (cs *CacheService) Clear() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.cache = make(map[string]*CacheItem)
	cs.logger.Info("Cache cleared")
}

// GetStats retrieves cache statistics
func (cs *CacheService) GetStats() *CacheStats {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	var hitCount, missCount int64
	totalItems := int64(len(cs.cache))

	// calculate hit rate (simplified handling, should maintain global statistics)
	for _, item := range cs.cache {
		hitCount += item.AccessCount
	}

	var hitRate float64
	if hitCount+missCount > 0 {
		hitRate = float64(hitCount) / float64(hitCount+missCount) * 100
	}

	return &CacheStats{
		TotalItems:    totalItems,
		HitCount:      hitCount,
		MissCount:     missCount,
		HitRate:       hitRate,
		EvictionCount: 0, // simplified handling
		CreatedAt:     time.Now(),
		LastCleanup:   time.Now(),
	}
}

// GetAllKeys retrieves all cache keys
func (cs *CacheService) GetAllKeys() []string {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	keys := make([]string, 0, len(cs.cache))
	for key := range cs.cache {
		keys = append(keys, key)
	}

	return keys
}

// cleanupLoop cleanup expired cache loop
func (cs *CacheService) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(cs.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cs.stopChan:
			return
		case <-ticker.C:
			cs.cleanup()
		}
	}
}

// cleanup cleans up expired cache items
func (cs *CacheService) cleanup() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, item := range cs.cache {
		if now.After(item.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(cs.cache, key)
	}

	if len(expiredKeys) > 0 {
		cs.logger.WithField("expired_count", len(expiredKeys)).Debug("Cleaned up expired cache items")
	}
}

// evictItems evicts cache items based on policy
func (cs *CacheService) evictItems(count int) {
	if len(cs.cache) == 0 {
		return
	}

	switch cs.config.EvictionPolicy {
	case "LRU":
		cs.evictLRU(count)
	case "LFU":
		cs.evictLFU(count)
	case "TTL":
		cs.evictTTL(count)
	default:
		cs.evictLRU(count)
	}
}

// evictLRU evicts based on least recently used policy
func (cs *CacheService) evictLRU(count int) {
	items := make([]*CacheItem, 0, len(cs.cache))
	for _, item := range cs.cache {
		items = append(items, item)
	}

	// sort by last access time
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].LastAccess.After(items[j].LastAccess) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// delete the oldest items
	evicted := 0
	for i := 0; i < len(items) && evicted < count; i++ {
		delete(cs.cache, items[i].Key)
		evicted++
	}

	cs.logger.WithField("evicted_count", evicted).Debug("Evicted cache items using LRU")
}

// evictLFU evicts based on least frequently used policy
func (cs *CacheService) evictLFU(count int) {
	items := make([]*CacheItem, 0, len(cs.cache))
	for _, item := range cs.cache {
		items = append(items, item)
	}

	// sort by access count
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].AccessCount > items[j].AccessCount {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// delete the least frequently accessed items
	evicted := 0
	for i := 0; i < len(items) && evicted < count; i++ {
		delete(cs.cache, items[i].Key)
		evicted++
	}

	cs.logger.WithField("evicted_count", evicted).Debug("Evicted cache items using LFU")
}

// evictTTL evicts based on TTL policy
func (cs *CacheService) evictTTL(count int) {
	items := make([]*CacheItem, 0, len(cs.cache))
	for _, item := range cs.cache {
		items = append(items, item)
	}

	// sort by expiration time
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].ExpiresAt.After(items[j].ExpiresAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// delete items that will expire soon
	evicted := 0
	for i := 0; i < len(items) && evicted < count; i++ {
		delete(cs.cache, items[i].Key)
		evicted++
	}

	cs.logger.WithField("evicted_count", evicted).Debug("Evicted cache items using TTL")
}

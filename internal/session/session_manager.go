package session

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ccany/internal/models"

	"github.com/sirupsen/logrus"
)

// SessionManager manages conversation sessions for maintaining context
type SessionManager struct {
	sessions      map[string]*ConversationSession
	mutex         sync.RWMutex
	logger        *logrus.Logger
	config        *SessionConfig
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// ConversationSession represents a conversation session
type ConversationSession struct {
	ID           string                 `json:"id"`
	Messages     []models.ClaudeMessage `json:"messages"`
	CreatedAt    time.Time              `json:"created_at"`
	LastAccess   time.Time              `json:"last_access"`
	SystemPrompt interface{}            `json:"system_prompt,omitempty"`
	TotalTokens  int                    `json:"total_tokens"`
	MessageCount int                    `json:"message_count"`
	UserID       string                 `json:"user_id,omitempty"`
	ProjectPath  string                 `json:"project_path,omitempty"`
}

// SessionConfig configuration for session management
type SessionConfig struct {
	MaxSessions           int           `json:"max_sessions"`
	SessionTTL            time.Duration `json:"session_ttl"`
	CleanupInterval       time.Duration `json:"cleanup_interval"`
	MaxMessagesPerSession int           `json:"max_messages_per_session"`
	MaxTokensPerSession   int           `json:"max_tokens_per_session"`
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *SessionConfig, logger *logrus.Logger) *SessionManager {
	if config == nil {
		config = DefaultSessionConfig()
	}

	sm := &SessionManager{
		sessions: make(map[string]*ConversationSession),
		logger:   logger,
		config:   config,
		stopChan: make(chan struct{}),
	}

	return sm
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		MaxSessions:           1000,
		SessionTTL:            24 * time.Hour, // 24 hours
		CleanupInterval:       1 * time.Hour,  // cleanup every hour
		MaxMessagesPerSession: 100,
		MaxTokensPerSession:   200000, // 200k tokens max per session
	}
}

// Start starts the session manager
func (sm *SessionManager) Start(ctx context.Context) {
	sm.cleanupTicker = time.NewTicker(sm.config.CleanupInterval)

	go func() {
		defer sm.cleanupTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-sm.stopChan:
				return
			case <-sm.cleanupTicker.C:
				sm.cleanup()
			}
		}
	}()

	sm.logger.Info("Session manager started")
}

// Stop stops the session manager
func (sm *SessionManager) Stop() {
	close(sm.stopChan)
	if sm.cleanupTicker != nil {
		sm.cleanupTicker.Stop()
	}
	sm.logger.Info("Session manager stopped")
}

// GetOrCreateSession gets an existing session or creates a new one
func (sm *SessionManager) GetOrCreateSession(projectPath, userID string, systemPrompt interface{}) (*ConversationSession, error) {
	sessionID := sm.generateSessionID(projectPath, userID)

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if exists {
		// Update last access time
		session.LastAccess = time.Now()
		sm.logger.WithFields(logrus.Fields{
			"session_id":    sessionID,
			"message_count": session.MessageCount,
		}).Debug("Retrieved existing session")
		return session, nil
	}

	// Check session limit
	if len(sm.sessions) >= sm.config.MaxSessions {
		sm.evictOldestSession()
	}

	// Create new session
	now := time.Now()
	session = &ConversationSession{
		ID:           sessionID,
		Messages:     make([]models.ClaudeMessage, 0),
		CreatedAt:    now,
		LastAccess:   now,
		SystemPrompt: systemPrompt,
		TotalTokens:  0,
		MessageCount: 0,
		UserID:       userID,
		ProjectPath:  projectPath,
	}

	sm.sessions[sessionID] = session

	sm.logger.WithFields(logrus.Fields{
		"session_id":   sessionID,
		"project_path": projectPath,
		"user_id":      userID,
	}).Info("Created new conversation session")

	return session, nil
}

// AddMessage adds a message to the session
func (sm *SessionManager) AddMessage(sessionID string, message models.ClaudeMessage, tokens int) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Check limits
	if session.MessageCount >= sm.config.MaxMessagesPerSession {
		// Remove oldest message pair (user + assistant)
		if len(session.Messages) >= 2 {
			session.Messages = session.Messages[2:]
			session.MessageCount -= 2
		}
	}

	if session.TotalTokens+tokens > sm.config.MaxTokensPerSession {
		// Remove oldest messages until under limit
		for len(session.Messages) > 0 && session.TotalTokens+tokens > sm.config.MaxTokensPerSession {
			removed := session.Messages[0]
			session.Messages = session.Messages[1:]
			session.MessageCount--
			// Estimate tokens for removed message (simplified)
			removedTokens := len(fmt.Sprintf("%v", removed.Content)) / 4
			session.TotalTokens -= removedTokens
		}
	}

	// Add new message
	session.Messages = append(session.Messages, message)
	session.MessageCount++
	session.TotalTokens += tokens
	session.LastAccess = time.Now()

	sm.logger.WithFields(logrus.Fields{
		"session_id":    sessionID,
		"message_count": session.MessageCount,
		"total_tokens":  session.TotalTokens,
		"role":          message.Role,
	}).Debug("Added message to session")

	return nil
}

// GetSessionMessages gets all messages from a session with optional system prompt
func (sm *SessionManager) GetSessionMessages(sessionID string) ([]models.ClaudeMessage, interface{}, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Update last access
	session.LastAccess = time.Now()

	// Return copy of messages to avoid concurrent access issues
	messages := make([]models.ClaudeMessage, len(session.Messages))
	copy(messages, session.Messages)

	return messages, session.SystemPrompt, nil
}

// GetSession gets session info
func (sm *SessionManager) GetSession(sessionID string) (*ConversationSession, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Return copy to avoid concurrent access issues
	sessionCopy := *session
	return &sessionCopy, nil
}

// ClearSession clears all messages from a session
func (sm *SessionManager) ClearSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.Messages = make([]models.ClaudeMessage, 0)
	session.MessageCount = 0
	session.TotalTokens = 0
	session.LastAccess = time.Now()

	sm.logger.WithField("session_id", sessionID).Info("Cleared session messages")
	return nil
}

// DeleteSession deletes a session
func (sm *SessionManager) DeleteSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.sessions[sessionID]; !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	delete(sm.sessions, sessionID)
	sm.logger.WithField("session_id", sessionID).Info("Deleted session")
	return nil
}

// GetSessionStats returns session statistics
func (sm *SessionManager) GetSessionStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_sessions": len(sm.sessions),
		"max_sessions":   sm.config.MaxSessions,
	}

	if len(sm.sessions) > 0 {
		var totalMessages, totalTokens int
		oldestAccess := time.Now()
		newestAccess := time.Time{}

		for _, session := range sm.sessions {
			totalMessages += session.MessageCount
			totalTokens += session.TotalTokens

			if session.LastAccess.Before(oldestAccess) {
				oldestAccess = session.LastAccess
			}
			if session.LastAccess.After(newestAccess) {
				newestAccess = session.LastAccess
			}
		}

		stats["avg_messages_per_session"] = float64(totalMessages) / float64(len(sm.sessions))
		stats["avg_tokens_per_session"] = float64(totalTokens) / float64(len(sm.sessions))
		stats["oldest_access"] = oldestAccess
		stats["newest_access"] = newestAccess
	}

	return stats
}

// generateSessionID generates a session ID based on project path and user ID
func (sm *SessionManager) generateSessionID(projectPath, userID string) string {
	data := fmt.Sprintf("%s:%s", projectPath, userID)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("session_%x", hash)
}

// cleanup removes expired sessions
func (sm *SessionManager) cleanup() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	expiredSessions := make([]string, 0)

	for sessionID, session := range sm.sessions {
		if now.Sub(session.LastAccess) > sm.config.SessionTTL {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	for _, sessionID := range expiredSessions {
		delete(sm.sessions, sessionID)
	}

	if len(expiredSessions) > 0 {
		sm.logger.WithField("expired_count", len(expiredSessions)).Info("Cleaned up expired sessions")
	}
}

// evictOldestSession evicts the oldest session when limit is reached
func (sm *SessionManager) evictOldestSession() {
	if len(sm.sessions) == 0 {
		return
	}

	oldestID := ""
	oldestTime := time.Now()

	for sessionID, session := range sm.sessions {
		if session.LastAccess.Before(oldestTime) {
			oldestTime = session.LastAccess
			oldestID = sessionID
		}
	}

	if oldestID != "" {
		delete(sm.sessions, oldestID)
		sm.logger.WithField("evicted_session", oldestID).Info("Evicted oldest session due to limit")
	}
}

// ExportSession exports session data for backup/restore
func (sm *SessionManager) ExportSession(sessionID string) ([]byte, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(session)
}

// ImportSession imports session data from backup
func (sm *SessionManager) ImportSession(data []byte) error {
	var session ConversationSession
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if session already exists
	if _, exists := sm.sessions[session.ID]; exists {
		return fmt.Errorf("session %s already exists", session.ID)
	}

	// Check session limit
	if len(sm.sessions) >= sm.config.MaxSessions {
		sm.evictOldestSession()
	}

	sm.sessions[session.ID] = &session
	sm.logger.WithField("session_id", session.ID).Info("Imported session")

	return nil
}

package claudecode

import (
	"ccany/internal/models"

	"github.com/sirupsen/logrus"
)

// ToolUseStrategy handles requests with tool definitions
type ToolUseStrategy struct {
	model   string
	enabled bool
	logger  *logrus.Logger
}

// NewToolUseStrategy creates a new tool use strategy
func NewToolUseStrategy(model string, enabled bool, logger *logrus.Logger) *ToolUseStrategy {
	return &ToolUseStrategy{
		model:   model,
		enabled: enabled,
		logger:  logger,
	}
}

// ShouldApply checks if the request contains tools
func (s *ToolUseStrategy) ShouldApply(req *models.ClaudeMessagesRequest, tokenCount int) bool {
	if !s.enabled || s.model == "" {
		return false
	}

	// Check if request has tools
	hasTools := len(req.Tools) > 0

	if hasTools && s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"tools_count": len(req.Tools),
			"tool_choice": req.ToolChoice,
			"model":       s.model,
		}).Debug("Tool use strategy detected tools in request")
	}

	return hasTools
}

// GetModel returns the model for tool use
func (s *ToolUseStrategy) GetModel() string {
	return s.model
}

// GetReason returns the reason for routing
func (s *ToolUseStrategy) GetReason() string {
	return "tool_use_detected"
}

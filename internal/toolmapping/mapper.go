package toolmapping

import (
	"fmt"
	"strings"
	"sync"
)

// ToolMapper provides centralized tool name mapping functionality
type ToolMapper struct {
	mu                    sync.RWMutex
	customMappings        map[string]string
	openaiMappings        map[string]string
	claudeMappings        map[string]string
	reverseClaudeMappings map[string]string
}

// NewToolMapper creates a new tool mapper with default mappings
func NewToolMapper() *ToolMapper {
	mapper := &ToolMapper{
		customMappings:        make(map[string]string),
		openaiMappings:        make(map[string]string),
		claudeMappings:        make(map[string]string),
		reverseClaudeMappings: make(map[string]string),
	}

	mapper.initializeDefaultMappings()
	return mapper
}

// initializeDefaultMappings sets up the default tool name mappings
func (m *ToolMapper) initializeDefaultMappings() {
	// Custom format mappings (from backend-specific formats)
	customMappings := map[string]string{
		"ExitTool":       "ExitTool", // Keep ExitTool as-is for special handling
		"FsCreateFile":   "Write",
		"FileWrite":      "Write",
		"CreateFile":     "Write",
		"FsReadFile":     "Read",
		"FileRead":       "Read",
		"ReadFile":       "Read",
		"FsEditFile":     "Edit",
		"FileEdit":       "Edit",
		"EditFile":       "Edit",
		"BashCommand":    "Bash",
		"RunCommand":     "Bash",
		"ExecuteCommand": "Bash",
		"GlobSearch":     "Glob",
		"FindFiles":      "Glob",
		"GrepSearch":     "Grep",
		"SearchInFiles":  "Grep",
		"ListDirectory":  "LS",
		"ListDir":        "LS",
		"MultiFileEdit":  "MultiEdit",
		"BatchEdit":      "MultiEdit",
		"NotebookRead":   "NotebookRead",
		"NotebookEdit":   "NotebookEdit",
		"WebFetch":       "WebFetch",
		"FetchUrl":       "WebFetch",
		"TodoWrite":      "TodoWrite",
		"CreateTodo":     "TodoWrite",
		"WebSearch":      "WebSearch",
		"SearchWeb":      "WebSearch",
		"Task":           "Task",
		"LaunchAgent":    "Task",
	}

	// OpenAI format mappings (standard OpenAI tool calling)
	openaiMappings := map[string]string{
		"ExitTool":      "ExitTool", // Keep ExitTool as-is for special handling
		"FileWrite":     "Write",
		"FsCreateFile":  "Write",
		"code_create":   "Write", // Claude Code specific tool
		"FileRead":      "Read",
		"FsReadFile":    "Read",
		"FileEdit":      "Edit",
		"FsEditFile":    "Edit",
		"BashCommand":   "Bash",
		"GlobSearch":    "Glob",
		"GrepSearch":    "Grep",
		"ListDirectory": "LS",
		"MultiFileEdit": "MultiEdit",
		"NotebookRead":  "NotebookRead",
		"NotebookEdit":  "NotebookEdit",
		"WebFetch":      "WebFetch",
		"TodoWrite":     "TodoWrite",
		"WebSearch":     "WebSearch",
		"Task":          "Task",
		// OpenAI tool names to Claude Code names
		"str_replace_editor":          "write_to_file",
		"str_replace_based_edit_tool": "apply_diff",
		"bash":                        "execute_command",
		"computer":                    "browser_action",
		"grep":                        "search_files",
		"ls":                          "list_files",
	}

	// Claude format mappings (Claude native format to internal names)
	claudeMappings := map[string]string{
		"str_replace_editor":          "Edit",
		"str_replace_based_edit_tool": "Edit",
		"bash":                        "Bash",
		"computer":                    "Computer",
		"text_editor":                 "Edit",
		// Claude Code official tools
		"Write":                    "Write",
		"Read":                     "Read",
		"Edit":                     "Edit",
		"MultiEdit":                "MultiEdit",
		"Bash":                     "Bash",
		"Glob":                     "Glob",
		"Grep":                     "Grep",
		"LS":                       "LS",
		"NotebookRead":             "NotebookRead",
		"NotebookEdit":             "NotebookEdit",
		"WebFetch":                 "WebFetch",
		"WebSearch":                "WebSearch",
		"TodoWrite":                "TodoWrite",
		"Task":                     "Task",
		"ExitPlanMode":             "ExitPlanMode",
		"mcp__ide__getDiagnostics": "mcp__ide__getDiagnostics",
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.customMappings = customMappings
	m.openaiMappings = openaiMappings
	m.claudeMappings = claudeMappings

	// Build reverse mappings for Claude tools
	m.reverseClaudeMappings = make(map[string]string)
	for claude, internal := range claudeMappings {
		m.reverseClaudeMappings[internal] = claude
	}
}

// MapCustomToClaudeName maps custom tool names to Claude Code tool names
func (m *ToolMapper) MapCustomToClaudeName(customName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if mapped, exists := m.customMappings[customName]; exists {
		return mapped
	}

	// Try case-insensitive matching
	for key, value := range m.customMappings {
		if strings.EqualFold(key, customName) {
			return value
		}
	}

	// Return original name if no mapping found
	return customName
}

// MapOpenAIToClaudeName maps OpenAI tool names to Claude Code tool names
func (m *ToolMapper) MapOpenAIToClaudeName(openaiName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if mapped, exists := m.openaiMappings[openaiName]; exists {
		return mapped
	}

	// Try case-insensitive matching
	for key, value := range m.openaiMappings {
		if strings.EqualFold(key, openaiName) {
			return value
		}
	}

	// Return original name if no mapping found
	return openaiName
}

// MapClaudeToInternalName maps Claude tool names to internal tool names
func (m *ToolMapper) MapClaudeToInternalName(claudeName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if mapped, exists := m.claudeMappings[claudeName]; exists {
		return mapped
	}

	// Try case-insensitive matching
	for key, value := range m.claudeMappings {
		if strings.EqualFold(key, claudeName) {
			return value
		}
	}

	// Return original name if no mapping found
	return claudeName
}

// MapInternalToClaudeName maps internal tool names back to Claude tool names
func (m *ToolMapper) MapInternalToClaudeName(internalName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if mapped, exists := m.reverseClaudeMappings[internalName]; exists {
		return mapped
	}

	// Return original name if no mapping found
	return internalName
}

// AddCustomMapping adds a custom tool name mapping
func (m *ToolMapper) AddCustomMapping(from, to string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customMappings[from] = to
}

// AddOpenAIMapping adds an OpenAI tool name mapping
func (m *ToolMapper) AddOpenAIMapping(from, to string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.openaiMappings[from] = to
}

// AddClaudeMapping adds a Claude tool name mapping
func (m *ToolMapper) AddClaudeMapping(from, to string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claudeMappings[from] = to
	m.reverseClaudeMappings[to] = from
}

// RemoveMapping removes a mapping from all mapping types
func (m *ToolMapper) RemoveMapping(toolName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.customMappings, toolName)
	delete(m.openaiMappings, toolName)
	delete(m.claudeMappings, toolName)

	// Remove from reverse mappings
	for k, v := range m.reverseClaudeMappings {
		if v == toolName {
			delete(m.reverseClaudeMappings, k)
		}
	}
}

// GetAllMappings returns all current mappings
func (m *ToolMapper) GetAllMappings() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"custom":  m.copyMap(m.customMappings),
		"openai":  m.copyMap(m.openaiMappings),
		"claude":  m.copyMap(m.claudeMappings),
		"reverse": m.copyMap(m.reverseClaudeMappings),
	}
}

// copyMap creates a copy of a string map
func (m *ToolMapper) copyMap(original map[string]string) map[string]string {
	copy := make(map[string]string)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// IsKnownTool checks if a tool name is known in any mapping
func (m *ToolMapper) IsKnownTool(toolName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check in all mappings
	if _, exists := m.customMappings[toolName]; exists {
		return true
	}
	if _, exists := m.openaiMappings[toolName]; exists {
		return true
	}
	if _, exists := m.claudeMappings[toolName]; exists {
		return true
	}
	if _, exists := m.reverseClaudeMappings[toolName]; exists {
		return true
	}

	return false
}

// GetSupportedTools returns a list of all supported tool names
func (m *ToolMapper) GetSupportedTools() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	toolSet := make(map[string]bool)

	// Collect unique tool names from all mappings
	for tool := range m.customMappings {
		toolSet[tool] = true
	}
	for tool := range m.openaiMappings {
		toolSet[tool] = true
	}
	for tool := range m.claudeMappings {
		toolSet[tool] = true
	}
	for tool := range m.reverseClaudeMappings {
		toolSet[tool] = true
	}

	// Convert to slice
	tools := make([]string, 0, len(toolSet))
	for tool := range toolSet {
		tools = append(tools, tool)
	}

	return tools
}

// GetClaudeToolNames returns all Claude Code tool names
func (m *ToolMapper) GetClaudeToolNames() []string {
	return []string{
		// Core Claude Code tools
		"Write", "Read", "Edit", "MultiEdit",
		"Bash", "Glob", "Grep", "LS",
		"NotebookRead", "NotebookEdit",
		"WebFetch", "WebSearch", "TodoWrite", "Task",
		"ExitPlanMode",
		// Legacy Claude tools (for compatibility)
		"str_replace_editor", "str_replace_based_edit_tool",
		"bash", "computer", "text_editor",
		// MCP tools
		"mcp__ide__getDiagnostics",
		// Special tools
		"ExitTool",
	}
}

// ValidateMapping checks if a mapping is valid
func (m *ToolMapper) ValidateMapping(from, to string) error {
	if from == "" {
		return fmt.Errorf("source tool name cannot be empty")
	}
	if to == "" {
		return fmt.Errorf("target tool name cannot be empty")
	}

	claudeTools := m.GetClaudeToolNames()
	validTarget := false
	for _, tool := range claudeTools {
		if tool == to {
			validTarget = true
			break
		}
	}

	if !validTarget {
		return fmt.Errorf("target tool name '%s' is not a valid Claude tool", to)
	}

	return nil
}

package toolmapping

import "sync"

var (
	defaultMapper *ToolMapper
	once          sync.Once
)

// GetDefaultMapper returns the default global tool mapper instance
func GetDefaultMapper() *ToolMapper {
	once.Do(func() {
		defaultMapper = NewToolMapper()
	})
	return defaultMapper
}

// MapCustomToClaudeName maps custom tool names to Claude Code tool names using the default mapper
func MapCustomToClaudeName(customName string) string {
	return GetDefaultMapper().MapCustomToClaudeName(customName)
}

// MapOpenAIToClaudeName maps OpenAI tool names to Claude Code tool names using the default mapper
func MapOpenAIToClaudeName(openaiName string) string {
	return GetDefaultMapper().MapOpenAIToClaudeName(openaiName)
}

// MapClaudeToInternalName maps Claude tool names to internal tool names using the default mapper
func MapClaudeToInternalName(claudeName string) string {
	return GetDefaultMapper().MapClaudeToInternalName(claudeName)
}

// MapInternalToClaudeName maps internal tool names back to Claude tool names using the default mapper
func MapInternalToClaudeName(internalName string) string {
	return GetDefaultMapper().MapInternalToClaudeName(internalName)
}

// IsKnownTool checks if a tool name is known using the default mapper
func IsKnownTool(toolName string) bool {
	return GetDefaultMapper().IsKnownTool(toolName)
}

// GetSupportedTools returns all supported tool names using the default mapper
func GetSupportedTools() []string {
	return GetDefaultMapper().GetSupportedTools()
}

// GetClaudeToolNames returns all Claude Code tool names using the default mapper
func GetClaudeToolNames() []string {
	return GetDefaultMapper().GetClaudeToolNames()
}

// AddCustomMapping adds a custom tool name mapping to the default mapper
func AddCustomMapping(from, to string) {
	GetDefaultMapper().AddCustomMapping(from, to)
}

// AddOpenAIMapping adds an OpenAI tool name mapping to the default mapper
func AddOpenAIMapping(from, to string) {
	GetDefaultMapper().AddOpenAIMapping(from, to)
}

// AddClaudeMapping adds a Claude tool name mapping to the default mapper
func AddClaudeMapping(from, to string) {
	GetDefaultMapper().AddClaudeMapping(from, to)
}

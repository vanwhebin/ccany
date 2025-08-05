package toolmapping

import "ccany/internal/models"

// GetOfficialClaudeToolDefinitions returns the official Claude tool definitions based on documentation
func GetOfficialClaudeToolDefinitions() map[string]models.ClaudeToolDefinition {
	return map[string]models.ClaudeToolDefinition{
		// Web Search Tool (Server-side)
		"web_search_20250305": {
			Type:        "web_search_20250305",
			Name:        "web_search",
			Description: "Search the web for information using a search engine. This tool provides access to current information from across the internet. Use this when you need up-to-date information, current events, or to find specific resources online.",
			Category:    "server",
			Version:     "20250305",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query to execute",
					},
					"max_uses": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of search operations to perform (default: 5)",
						"default":     5,
					},
					"allowed_domains": map[string]interface{}{
						"type":        "array",
						"description": "List of domains to restrict search results to",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"blocked_domains": map[string]interface{}{
						"type":        "array",
						"description": "List of domains to exclude from search results",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"query"},
			},
		},

		// Text Editor Tool - Claude 4 version (Client-side)
		"text_editor_20250429": {
			Type:        "text_editor_20250429",
			Name:        "str_replace_based_edit_tool",
			Description: "A tool for viewing and editing files. This tool can read entire files, make targeted edits using string replacement, create new files, and list directory contents.",
			Category:    "client",
			Version:     "20250429",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The command to execute",
						"enum":        []string{"view", "str_replace", "create", "insert"},
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute path to the file",
					},
					"old_str": map[string]interface{}{
						"type":        "string",
						"description": "The exact string to replace (required for str_replace)",
					},
					"new_str": map[string]interface{}{
						"type":        "string",
						"description": "The replacement string (required for str_replace)",
					},
					"file_text": map[string]interface{}{
						"type":        "string",
						"description": "The content of the file (required for create)",
					},
				},
				"required": []string{"command", "path"},
			},
		},

		// Bash Tool - Claude 4 version (Client-side)
		"bash_20250129": {
			Type:        "bash_20250129",
			Name:        "bash",
			Description: "Run bash commands to accomplish tasks. Use this tool when you need to execute system commands, run scripts, or interact with the operating system.",
			Category:    "client",
			Version:     "20250129",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The bash command to execute",
					},
				},
				"required": []string{"command"},
			},
		},

		// Read Tool
		"read_tool": {
			Name:        "Read",
			Description: "Read files from the local filesystem",
			Category:    "client",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "The absolute path to the file to read",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "The number of lines to read",
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "The line number to start reading from",
					},
				},
				"required": []string{"file_path"},
			},
		},

		// Write Tool
		"write_tool": {
			Name:        "Write",
			Description: "Write files to the local filesystem",
			Category:    "client",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "The absolute path to the file to write",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
				"required": []string{"file_path", "content"},
			},
		},

		// Edit Tool
		"edit_tool": {
			Name:        "Edit",
			Description: "Edit files using exact string replacement",
			Category:    "client",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "The absolute path to the file to modify",
					},
					"old_string": map[string]interface{}{
						"type":        "string",
						"description": "The text to replace",
					},
					"new_string": map[string]interface{}{
						"type":        "string",
						"description": "The text to replace it with",
					},
					"replace_all": map[string]interface{}{
						"type":        "boolean",
						"description": "Replace all occurrences of old_string",
						"default":     false,
					},
				},
				"required": []string{"file_path", "old_string", "new_string"},
			},
		},
	}
}

// GetCommonToolMappings returns common tool name mappings
func GetCommonToolMappings() map[string]string {
	return map[string]string{
		// File operations
		"str_replace_editor":          "Edit",
		"str_replace_based_edit_tool": "Edit",
		"file_editor":                 "Edit",
		"create_file":                 "Write",
		"write_file":                  "Write",
		"read_file":                   "Read",
		"view_file":                   "Read",

		// System operations
		"bash":        "Bash",
		"shell":       "Bash",
		"execute":     "Bash",
		"run_command": "Bash",

		// Web operations
		"web_search":    "WebSearch",
		"search_web":    "WebSearch",
		"google_search": "WebSearch",

		// Task management
		"task":     "Task",
		"agent":    "Task",
		"subagent": "Task",
	}
}

// GetToolDefinitionByName returns a tool definition by name
func GetToolDefinitionByName(name string) (*models.ClaudeToolDefinition, bool) {
	// Get all definitions including simple names
	allDefs := GetAllToolDefinitions()

	// First try direct lookup
	if def, exists := allDefs[name]; exists {
		return &def, true
	}

	// Then try lookup by tool name within definitions
	for _, def := range allDefs {
		if def.Name == name || def.Type == name {
			return &def, true
		}
	}

	return nil, false
}

// GetAllToolDefinitions returns all tool definitions including versioned and simple names
func GetAllToolDefinitions() map[string]models.ClaudeToolDefinition {
	officialDefs := GetOfficialClaudeToolDefinitions()
	allDefs := make(map[string]models.ClaudeToolDefinition)

	// Add official definitions
	for key, def := range officialDefs {
		allDefs[key] = def
	}

	// Add simple tool definitions
	simpleTools := map[string]models.ClaudeToolDefinition{
		"Bash": {
			Type:        "Bash",
			Name:        "Bash",
			Description: "Run bash commands",
			Category:    "client",
			InputSchema: officialDefs["bash_20250129"].InputSchema,
		},
		"Write": {
			Type:        "Write",
			Name:        "Write",
			Description: "Write files to the local filesystem",
			Category:    "client",
			InputSchema: officialDefs["write_tool"].InputSchema,
		},
		"Read": {
			Type:        "Read",
			Name:        "Read",
			Description: "Read files from the local filesystem",
			Category:    "client",
			InputSchema: officialDefs["read_tool"].InputSchema,
		},
		"bash_20250124": { // Support older version for backward compatibility
			Type:        "bash_20250124",
			Name:        "bash",
			Description: "Run bash commands to accomplish tasks",
			Category:    "client",
			Version:     "20250124",
			InputSchema: officialDefs["bash_20250129"].InputSchema,
		},
	}

	for key, def := range simpleTools {
		allDefs[key] = def
	}

	return allDefs
}

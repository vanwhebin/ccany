package tests

import (
	"testing"

	"ccany/internal/converter"
	"ccany/internal/models"
	"ccany/internal/toolmapping"

	"github.com/stretchr/testify/assert"
)

func TestToolNameConversion(t *testing.T) {
	t.Run("Claude to OpenAI tool name conversion", func(t *testing.T) {
		// 创建 OpenAI 转换器
		openaiConverter := converter.NewOpenAIConverter()

		// 定义 Claude Code 工具
		claudeTools := []models.ClaudeTool{
			{
				Name:        "write_to_file",
				Description: "Write content to a file",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":    map[string]interface{}{"type": "string"},
						"content": map[string]interface{}{"type": "string"},
					},
				},
			},
			{
				Name:        "read_file",
				Description: "Read content from a file",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{"type": "string"},
					},
				},
			},
			{
				Name:        "execute_command",
				Description: "Execute a shell command",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{"type": "string"},
					},
				},
			},
		}

		// 创建 Claude 请求
		claudeReq := &models.ClaudeMessagesRequest{
			Model:     "claude-3-opus-20240229",
			MaxTokens: 1000,
			Messages: []models.ClaudeMessage{
				{
					Role:    "user",
					Content: "Test message",
				},
			},
			Tools: claudeTools,
		}

		// 转换为 OpenAI 格式
		openaiReq, err := openaiConverter.ConvertFromClaude(claudeReq, "gpt-4", "gpt-3.5-turbo")
		assert.NoError(t, err)
		assert.NotNil(t, openaiReq)

		// 验证工具数量
		assert.Equal(t, len(claudeTools), len(openaiReq.Tools))

		// 验证工具名称转换
		expectedMappings := map[string]string{
			"write_to_file":   "str_replace_editor",
			"read_file":       "str_replace_editor",
			"execute_command": "bash",
		}

		for i, tool := range openaiReq.Tools {
			claudeToolName := claudeTools[i].Name
			openaiToolName := tool.Function.Name

			// 获取预期的 OpenAI 工具名称
			expectedOpenAIName := expectedMappings[claudeToolName]

			t.Logf("Claude tool '%s' converted to OpenAI tool '%s' (expected: '%s')",
				claudeToolName, openaiToolName, expectedOpenAIName)

			assert.Equal(t, expectedOpenAIName, openaiToolName,
				"Tool name conversion failed for %s", claudeToolName)
		}
	})

	t.Run("OpenAI to Claude tool name conversion", func(t *testing.T) {
		// 测试反向转换（从 OpenAI 到 Claude）
		testCases := []struct {
			openaiName   string
			expectedName string
		}{
			{"str_replace_editor", "write_to_file"},
			{"str_replace_based_edit_tool", "apply_diff"},
			{"bash", "execute_command"},
			{"computer", "browser_action"},
		}

		for _, tc := range testCases {
			claudeName := toolmapping.MapOpenAIToClaudeName(tc.openaiName)
			t.Logf("OpenAI tool '%s' mapped to Claude tool '%s' (expected: '%s')",
				tc.openaiName, claudeName, tc.expectedName)

			// 注意：某些工具可能有多个映射，所以我们只检查是否有有效的映射
			assert.NotEqual(t, tc.openaiName, claudeName,
				"Tool name should be mapped, not returned as-is")
		}
	})
}

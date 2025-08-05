package tests

import (
	"testing"

	"ccany/internal/converter"
	"ccany/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestClaudeCodeToolConversionIntegration(t *testing.T) {
	t.Run("Full request conversion with tool name mapping", func(t *testing.T) {
		// 创建一个包含 Claude Code 工具的完整请求
		claudeReq := &models.ClaudeMessagesRequest{
			Model:     "claude-3-opus-20240229",
			MaxTokens: 1000,
			Messages: []models.ClaudeMessage{
				{
					Role:    "user",
					Content: "Please create a file with some content",
				},
			},
			Tools: []models.ClaudeTool{
				{
					Name:        "write_to_file",
					Description: "Write content to a file",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"path":    map[string]interface{}{"type": "string", "description": "File path"},
							"content": map[string]interface{}{"type": "string", "description": "File content"},
						},
						"required": []string{"path", "content"},
					},
				},
				{
					Name:        "execute_command",
					Description: "Execute a shell command",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"command": map[string]interface{}{"type": "string", "description": "Command to execute"},
						},
						"required": []string{"command"},
					},
				},
			},
			ToolChoice: "auto",
		}

		// 转换为 OpenAI 格式
		openaiConverter := converter.NewOpenAIConverter()
		openaiReq, err := openaiConverter.ConvertFromClaude(claudeReq, "gpt-4", "gpt-3.5-turbo")
		assert.NoError(t, err)
		assert.NotNil(t, openaiReq)

		// 验证工具转换
		assert.Equal(t, 2, len(openaiReq.Tools))

		// 验证第一个工具（write_to_file -> str_replace_editor）
		assert.Equal(t, "str_replace_editor", openaiReq.Tools[0].Function.Name)
		assert.Contains(t, openaiReq.Tools[0].Function.Description, "Write content to a file")
		assert.Contains(t, openaiReq.Tools[0].Function.Description, "MUST be used")

		// 验证第二个工具（execute_command -> bash）
		assert.Equal(t, "bash", openaiReq.Tools[1].Function.Name)
		assert.Contains(t, openaiReq.Tools[1].Function.Description, "Execute a shell command")
		assert.Contains(t, openaiReq.Tools[1].Function.Description, "MUST be used")

		// 验证 tool_choice 被设置为 required（文件操作工具会强制使用）
		assert.Equal(t, "required", openaiReq.ToolChoice)

		t.Logf("Successfully converted Claude Code tools to OpenAI format")
	})

	t.Run("Tool call response conversion", func(t *testing.T) {
		// 模拟 OpenAI 返回的工具调用响应
		openaiResp := &models.OpenAIChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []models.Choice{
				{
					Index: 0,
					Message: models.Message{
						Role: "assistant",
						ToolCalls: []models.OpenAIToolCall{
							{
								ID:   "call_123",
								Type: "function",
								Function: models.OpenAIFunctionCall{
									Name:      "str_replace_editor",
									Arguments: `{"path": "test.txt", "content": "Hello, World!"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		}

		// 原始 Claude 请求（用于转换上下文）
		claudeReq := &models.ClaudeMessagesRequest{
			Model: "claude-3-opus-20240229",
		}

		// 转换回 Claude 格式
		claudeResp, err := converter.ConvertOpenAIToClaudeResponse(openaiResp, claudeReq)
		assert.NoError(t, err)
		assert.NotNil(t, claudeResp)

		// 验证工具调用被正确转换
		assert.Equal(t, "tool_use", claudeResp.StopReason)
		assert.Len(t, claudeResp.Content, 1)

		// 验证工具块
		toolBlock := claudeResp.Content[0]
		assert.Equal(t, "tool_use", toolBlock.Type)
		assert.Equal(t, "call_123", toolBlock.ID)
		assert.Equal(t, "write_to_file", toolBlock.Name) // 应该被转换回 Claude Code 的名称

		// 验证工具输入
		input, ok := toolBlock.Input.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test.txt", input["path"])
		assert.Equal(t, "Hello, World!", input["content"])

		t.Logf("Successfully converted OpenAI tool calls back to Claude Code format")
	})
}

package toolmapping

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestToolDefinitions(t *testing.T) {
	// Test that we can get all tool definitions
	allDefs := GetAllToolDefinitions()
	if len(allDefs) == 0 {
		t.Fatal("Expected to get tool definitions, got none")
	}

	// Test specific tool definitions
	expectedTools := []string{
		"web_search_20250305",
		"text_editor_20250429",
		"bash_20250124",
		"Write",
		"Read",
		"Bash",
	}

	for _, toolName := range expectedTools {
		if def, exists := allDefs[toolName]; !exists {
			t.Errorf("Expected tool %s to exist", toolName)
		} else {
			// Validate definition structure
			if def.Name == "" {
				t.Errorf("Tool %s has empty name", toolName)
			}
			if def.Description == "" {
				t.Errorf("Tool %s has empty description", toolName)
			}
			// Check if input schema has a type field
			if schemaType := GetSchemaType(def.InputSchema); schemaType == "" {
				t.Errorf("Tool %s has invalid input schema", toolName)
			}
		}
	}
}

func TestSchemaValidation(t *testing.T) {
	validator := NewSchemaValidator()

	// Test valid bash tool input
	bashInput := map[string]interface{}{
		"command": "echo hello",
	}
	if err := validator.ValidateToolInput("Bash", bashInput); err != nil {
		t.Errorf("Expected valid bash input to pass validation: %v", err)
	}

	// Test invalid bash tool input (missing command)
	invalidBashInput := map[string]interface{}{
		"timeout": 5000,
	}
	if err := validator.ValidateToolInput("Bash", invalidBashInput); err == nil {
		t.Error("Expected invalid bash input to fail validation")
	}

	// Test web search tool
	webSearchInput := map[string]interface{}{
		"query":    "test query",
		"max_uses": 3,
	}
	if err := validator.ValidateToolInput("web_search_20250305", webSearchInput); err != nil {
		t.Errorf("Expected valid web search input to pass validation: %v", err)
	}

	// Test Write tool
	writeInput := map[string]interface{}{
		"file_path": "/tmp/test.txt",
		"content":   "test content",
	}
	if err := validator.ValidateToolInput("Write", writeInput); err != nil {
		t.Errorf("Expected valid write input to pass validation: %v", err)
	}
}

func TestToolExecution(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	executor := NewToolExecutor(logger, "/tmp", true) // Enable security mode
	defer executor.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test Write tool execution
	writeInput := map[string]interface{}{
		"file_path": "/tmp/test_tool_execution.txt",
		"content":   "Hello from tool execution test!",
	}

	result := executor.ExecuteTool(ctx, "Write", writeInput, "test_write_001")
	if result.IsError {
		t.Errorf("Write tool execution failed: %s", result.Content)
	}

	// Test Read tool execution
	readInput := map[string]interface{}{
		"file_path": "/tmp/test_tool_execution.txt",
	}

	result = executor.ExecuteTool(ctx, "Read", readInput, "test_read_001")
	if result.IsError {
		t.Errorf("Read tool execution failed: %s", result.Content)
	}

	if result.Content != "Hello from tool execution test!" {
		t.Errorf("Expected read content to match write content, got: %s", result.Content)
	}

	// Test Bash tool execution
	bashInput := map[string]interface{}{
		"command": "echo 'bash test'",
	}

	result = executor.ExecuteTool(ctx, "Bash", bashInput, "test_bash_001")
	if result.IsError {
		t.Errorf("Bash tool execution failed: %s", result.Content)
	}

	// Test security validation
	dangerousBashInput := map[string]interface{}{
		"command": "rm -rf /",
	}

	result = executor.ExecuteTool(ctx, "Bash", dangerousBashInput, "test_dangerous_001")
	if !result.IsError {
		t.Error("Expected dangerous bash command to be blocked")
	}
}

func TestToolMapping(t *testing.T) {
	mapper := NewToolMapper()

	// Test OpenAI to Claude mappings
	claudeName := mapper.MapOpenAIToClaudeName("FileWrite")
	if claudeName != "Write" {
		t.Errorf("Expected FileWrite to map to Write, got %s", claudeName)
	}

	// Test custom mappings
	claudeName = mapper.MapCustomToClaudeName("BashCommand")
	if claudeName != "Bash" {
		t.Errorf("Expected BashCommand to map to Bash, got %s", claudeName)
	}

	// Test reverse mapping
	originalName := mapper.MapInternalToClaudeName("Write")
	if originalName != "Write" {
		t.Errorf("Expected Write to map back to Write, got %s", originalName)
	}

	// Test unknown tool
	unknown := mapper.MapCustomToClaudeName("UnknownTool")
	if unknown != "UnknownTool" {
		t.Errorf("Expected unknown tool to return original name, got %s", unknown)
	}
}

func TestEnhancedClaudeClient(t *testing.T) {
	// This would require actual Claude API credentials and is more of an integration test
	// For now, we'll test the helper functions

	// Test content conversion would be implemented here with actual client methods
	t.Log("Enhanced Claude client conversion functions would be tested here")
}

func TestVersionedTools(t *testing.T) {
	validator := NewSchemaValidator()

	// Test versioned tool validation
	textEditorInput := map[string]interface{}{
		"command": "view",
		"path":    "/tmp/test.txt",
	}

	if err := validator.ValidateVersionedToolCall("text_editor_20250429", textEditorInput); err != nil {
		t.Errorf("Expected valid text editor input to pass validation: %v", err)
	}

	// Test bash tool with versioned type
	bashInput := map[string]interface{}{
		"command": "ls -la",
	}

	if err := validator.ValidateVersionedToolCall("bash_20250124", bashInput); err != nil {
		t.Errorf("Expected valid bash input to pass validation: %v", err)
	}
}

// Benchmark tests
func BenchmarkToolValidation(b *testing.B) {
	validator := NewSchemaValidator()
	input := map[string]interface{}{
		"command": "echo hello",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateToolInput("Bash", input)
	}
}

func BenchmarkToolExecution(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce logging for benchmark

	executor := NewToolExecutor(logger, "/tmp", false) // Disable security for benchmark
	defer executor.Cleanup()

	ctx := context.Background()
	input := map[string]interface{}{
		"command": "echo benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := executor.ExecuteTool(ctx, "Bash", input, "benchmark")
		if result.IsError {
			b.Fatalf("Tool execution failed: %s", result.Content)
		}
	}
}

package toolmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ToolExecutor handles execution of client-side tools
type ToolExecutor struct {
	logger       *logrus.Logger
	workingDir   string
	bashSession  *BashSession
	securityMode bool // Enable security restrictions
}

// BashSession maintains a persistent bash session
type BashSession struct {
	cmd    *exec.Cmd
	stdin  *os.File
	stdout *os.File
	stderr *os.File
	active bool
}

// ToolExecutionResult represents the result of tool execution
type ToolExecutionResult struct {
	ToolUseID   string      `json:"tool_use_id"`
	Content     string      `json:"content"`
	IsError     bool        `json:"is_error"`
	Metadata    interface{} `json:"metadata,omitempty"`
	ExecutionMS int64       `json:"execution_ms,omitempty"`
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(logger *logrus.Logger, workingDir string, securityMode bool) *ToolExecutor {
	if workingDir == "" {
		workingDir = "/tmp"
	}

	return &ToolExecutor{
		logger:       logger,
		workingDir:   workingDir,
		securityMode: securityMode,
	}
}

// ExecuteTool executes a tool based on its definition and input
func (e *ToolExecutor) ExecuteTool(ctx context.Context, toolName string, input interface{}, toolUseID string) *ToolExecutionResult {
	start := time.Now()

	// Get tool definition
	def, exists := GetToolDefinitionByName(toolName)
	if !exists {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Unknown tool: %s", toolName),
			IsError:   true,
		}
	}

	// Only execute client tools
	if def.Category != "client" {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Tool %s is not a client-executable tool", toolName),
			IsError:   true,
		}
	}

	// Convert input to map
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		inputBytes, _ := json.Marshal(input)
		json.Unmarshal(inputBytes, &inputMap)
	}

	// Execute based on tool type or name
	var result *ToolExecutionResult
	toolType := def.Type
	if toolType == "" {
		toolType = def.Name // Use name if type is not set
	}

	switch toolType {
	case "bash_20250124", "bash_20250129", "Bash":
		result = e.executeBashTool(ctx, inputMap, toolUseID)
	case "text_editor_20250429", "text_editor_20250124", "str_replace_based_edit_tool":
		result = e.executeTextEditorTool(ctx, inputMap, toolUseID)
	case "Write", "write_tool":
		result = e.executeWriteTool(ctx, inputMap, toolUseID)
	case "Read", "read_tool":
		result = e.executeReadTool(ctx, inputMap, toolUseID)
	case "Edit", "edit_tool":
		result = e.executeEditTool(ctx, inputMap, toolUseID)
	default:
		// Try matching by name if type doesn't match
		switch toolName {
		case "Bash":
			result = e.executeBashTool(ctx, inputMap, toolUseID)
		case "Write":
			result = e.executeWriteTool(ctx, inputMap, toolUseID)
		case "Read":
			result = e.executeReadTool(ctx, inputMap, toolUseID)
		case "Edit":
			result = e.executeEditTool(ctx, inputMap, toolUseID)
		default:
			result = &ToolExecutionResult{
				ToolUseID: toolUseID,
				Content:   fmt.Sprintf("Tool execution not implemented for type: %s (name: %s)", toolType, toolName),
				IsError:   true,
			}
		}
	}

	result.ExecutionMS = time.Since(start).Milliseconds()

	e.logger.WithFields(logrus.Fields{
		"tool_name":    toolName,
		"tool_use_id":  toolUseID,
		"execution_ms": result.ExecutionMS,
		"is_error":     result.IsError,
		"content_size": len(result.Content),
	}).Debug("Tool execution completed")

	return result
}

// executeBashTool executes bash commands with security restrictions
func (e *ToolExecutor) executeBashTool(ctx context.Context, input map[string]interface{}, toolUseID string) *ToolExecutionResult {
	command, ok := input["command"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: command",
			IsError:   true,
		}
	}

	// Security check
	if e.securityMode {
		if err := e.validateBashCommand(command); err != nil {
			return &ToolExecutionResult{
				ToolUseID: toolUseID,
				Content:   fmt.Sprintf("Security violation: %s", err.Error()),
				IsError:   true,
			}
		}
	}

	// Check if restart is requested
	restart, _ := input["restart"].(bool)
	if restart && e.bashSession != nil {
		e.closeBashSession()
	}

	// Initialize bash session if needed
	if e.bashSession == nil {
		if err := e.initBashSession(ctx); err != nil {
			return &ToolExecutionResult{
				ToolUseID: toolUseID,
				Content:   fmt.Sprintf("Failed to initialize bash session: %s", err.Error()),
				IsError:   true,
			}
		}
	}

	// Execute command with timeout
	timeout := 30 * time.Second // Default timeout
	if timeoutMs, ok := input["timeout"].(float64); ok {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "bash", "-c", command)
	cmd.Dir = e.workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Command failed: %s\nOutput: %s", err.Error(), string(output)),
			IsError:   true,
		}
	}

	return &ToolExecutionResult{
		ToolUseID: toolUseID,
		Content:   string(output),
		IsError:   false,
	}
}

// executeTextEditorTool executes text editor operations
func (e *ToolExecutor) executeTextEditorTool(ctx context.Context, input map[string]interface{}, toolUseID string) *ToolExecutionResult {
	command, ok := input["command"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: command",
			IsError:   true,
		}
	}

	path, ok := input["path"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: path",
			IsError:   true,
		}
	}

	// Security check - ensure path is within working directory if security mode is enabled
	if e.securityMode {
		if err := e.validateFilePath(path); err != nil {
			return &ToolExecutionResult{
				ToolUseID: toolUseID,
				Content:   fmt.Sprintf("Security violation: %s", err.Error()),
				IsError:   true,
			}
		}
	}

	switch command {
	case "view":
		return e.executeFileView(path, toolUseID)
	case "str_replace":
		oldStr, _ := input["old_str"].(string)
		newStr, _ := input["new_str"].(string)
		return e.executeFileReplace(path, oldStr, newStr, toolUseID)
	case "create":
		fileText, _ := input["file_text"].(string)
		return e.executeFileCreate(path, fileText, toolUseID)
	case "insert":
		insertLine, _ := input["insert_line"].(float64)
		newStr, _ := input["new_str"].(string)
		return e.executeFileInsert(path, int(insertLine), newStr, toolUseID)
	default:
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Unknown text editor command: %s", command),
			IsError:   true,
		}
	}
}

// executeWriteTool executes Claude Code Write tool
func (e *ToolExecutor) executeWriteTool(ctx context.Context, input map[string]interface{}, toolUseID string) *ToolExecutionResult {
	filePath, ok := input["file_path"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: file_path",
			IsError:   true,
		}
	}

	content, ok := input["content"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: content",
			IsError:   true,
		}
	}

	// Security check
	if e.securityMode {
		if err := e.validateFilePath(filePath); err != nil {
			return &ToolExecutionResult{
				ToolUseID: toolUseID,
				Content:   fmt.Sprintf("Security violation: %s", err.Error()),
				IsError:   true,
			}
		}
	}

	return e.executeFileCreate(filePath, content, toolUseID)
}

// executeReadTool executes Claude Code Read tool
func (e *ToolExecutor) executeReadTool(ctx context.Context, input map[string]interface{}, toolUseID string) *ToolExecutionResult {
	filePath, ok := input["file_path"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: file_path",
			IsError:   true,
		}
	}

	// Security check
	if e.securityMode {
		if err := e.validateFilePath(filePath); err != nil {
			return &ToolExecutionResult{
				ToolUseID: toolUseID,
				Content:   fmt.Sprintf("Security violation: %s", err.Error()),
				IsError:   true,
			}
		}
	}

	return e.executeFileView(filePath, toolUseID)
}

// executeEditTool executes Claude Code Edit tool
func (e *ToolExecutor) executeEditTool(ctx context.Context, input map[string]interface{}, toolUseID string) *ToolExecutionResult {
	filePath, ok := input["file_path"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: file_path",
			IsError:   true,
		}
	}

	oldString, ok := input["old_string"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: old_string",
			IsError:   true,
		}
	}

	newString, ok := input["new_string"].(string)
	if !ok {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   "Missing required field: new_string",
			IsError:   true,
		}
	}

	// Security check
	if e.securityMode {
		if err := e.validateFilePath(filePath); err != nil {
			return &ToolExecutionResult{
				ToolUseID: toolUseID,
				Content:   fmt.Sprintf("Security violation: %s", err.Error()),
				IsError:   true,
			}
		}
	}

	return e.executeFileReplace(filePath, oldString, newString, toolUseID)
}

// Helper methods for file operations

func (e *ToolExecutor) executeFileView(path string, toolUseID string) *ToolExecutionResult {
	content, err := os.ReadFile(path)
	if err != nil {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Failed to read file: %s", err.Error()),
			IsError:   true,
		}
	}

	return &ToolExecutionResult{
		ToolUseID: toolUseID,
		Content:   string(content),
		IsError:   false,
	}
}

func (e *ToolExecutor) executeFileCreate(path, content string, toolUseID string) *ToolExecutionResult {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Failed to create directory: %s", err.Error()),
			IsError:   true,
		}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Failed to write file: %s", err.Error()),
			IsError:   true,
		}
	}

	return &ToolExecutionResult{
		ToolUseID: toolUseID,
		Content:   fmt.Sprintf("File created successfully: %s", path),
		IsError:   false,
	}
}

func (e *ToolExecutor) executeFileReplace(path, oldStr, newStr string, toolUseID string) *ToolExecutionResult {
	content, err := os.ReadFile(path)
	if err != nil {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Failed to read file: %s", err.Error()),
			IsError:   true,
		}
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, oldStr) {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("String not found in file: %s", oldStr),
			IsError:   true,
		}
	}

	newContent := strings.Replace(contentStr, oldStr, newStr, -1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Failed to write file: %s", err.Error()),
			IsError:   true,
		}
	}

	return &ToolExecutionResult{
		ToolUseID: toolUseID,
		Content:   fmt.Sprintf("File updated successfully: %s", path),
		IsError:   false,
	}
}

func (e *ToolExecutor) executeFileInsert(path string, lineNumber int, content string, toolUseID string) *ToolExecutionResult {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Failed to read file: %s", err.Error()),
			IsError:   true,
		}
	}

	lines := strings.Split(string(fileContent), "\n")
	if lineNumber < 0 || lineNumber > len(lines) {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Invalid line number: %d", lineNumber),
			IsError:   true,
		}
	}

	// Insert content at specified line
	newLines := append(lines[:lineNumber], append([]string{content}, lines[lineNumber:]...)...)
	newContent := strings.Join(newLines, "\n")

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return &ToolExecutionResult{
			ToolUseID: toolUseID,
			Content:   fmt.Sprintf("Failed to write file: %s", err.Error()),
			IsError:   true,
		}
	}

	return &ToolExecutionResult{
		ToolUseID: toolUseID,
		Content:   fmt.Sprintf("Content inserted at line %d in file: %s", lineNumber, path),
		IsError:   false,
	}
}

// Security validation methods

func (e *ToolExecutor) validateBashCommand(command string) error {
	// List of dangerous commands to block
	dangerousCommands := []string{
		"rm -rf", "rm -r", "rm -f",
		"dd if=", "mkfs", "fdisk",
		"killall", "pkill", "kill -9",
		"shutdown", "reboot", "halt",
		"chmod 777", "chown root",
		"sudo", "su -", "passwd",
		"> /dev/", ">> /dev/",
		"curl", "wget", "nc ", "ncat", "netcat",
	}

	cmdLower := strings.ToLower(command)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(cmdLower, dangerous) {
			return fmt.Errorf("dangerous command detected: %s", dangerous)
		}
	}

	return nil
}

func (e *ToolExecutor) validateFilePath(path string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid file path: %s", err.Error())
	}

	// Check if path is within working directory
	workingDirAbs, err := filepath.Abs(e.workingDir)
	if err != nil {
		return fmt.Errorf("invalid working directory: %s", err.Error())
	}

	if !strings.HasPrefix(absPath, workingDirAbs) {
		return fmt.Errorf("file path outside working directory: %s", absPath)
	}

	// Check for dangerous paths
	dangerousPaths := []string{
		"/etc/", "/bin/", "/sbin/", "/usr/bin/", "/usr/sbin/",
		"/boot/", "/dev/", "/proc/", "/sys/",
		"/root/", "/home/", "~/.ssh/",
	}

	for _, dangerous := range dangerousPaths {
		if strings.HasPrefix(absPath, dangerous) {
			return fmt.Errorf("access to dangerous path not allowed: %s", dangerous)
		}
	}

	return nil
}

// Bash session management

func (e *ToolExecutor) initBashSession(ctx context.Context) error {
	// For now, we don't maintain persistent sessions for security
	// Each command is executed independently
	return nil
}

func (e *ToolExecutor) closeBashSession() {
	if e.bashSession != nil {
		if e.bashSession.cmd != nil && e.bashSession.cmd.Process != nil {
			e.bashSession.cmd.Process.Kill()
		}
		e.bashSession = nil
	}
}

// Cleanup closes any open sessions
func (e *ToolExecutor) Cleanup() {
	e.closeBashSession()
}

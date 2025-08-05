package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ParseToolArguments parses tool call arguments with multiple fallback strategies
// First tries standard JSON parsing, then attempts to fix common issues
func ParseToolArguments(argsString string) (string, error) {
	// Handle empty or null input
	argsString = strings.TrimSpace(argsString)
	if argsString == "" || argsString == "{}" || argsString == "null" {
		return "{}", nil
	}

	// First attempt: Standard JSON parsing
	var result interface{}
	if err := json.Unmarshal([]byte(argsString), &result); err == nil {
		// Valid JSON, return as-is
		return argsString, nil
	}

	// Second attempt: Try to fix common JSON issues
	fixed := fixCommonJSONIssues(argsString)
	if err := json.Unmarshal([]byte(fixed), &result); err == nil {
		return fixed, nil
	}

	// Third attempt: Use gjson for more lenient parsing
	if gjson.Valid(argsString) {
		parsed := gjson.Parse(argsString)
		if parsed.Type != gjson.Null {
			return parsed.Raw, nil
		}
	}

	// Final attempt: Try to build valid JSON from the string
	repaired := repairJSON(argsString)
	if err := json.Unmarshal([]byte(repaired), &result); err == nil {
		return repaired, nil
	}

	// All attempts failed, return safe empty object
	return "{}", fmt.Errorf("failed to parse tool arguments: %s", argsString)
}

// fixCommonJSONIssues attempts to fix common JSON formatting issues
func fixCommonJSONIssues(input string) string {
	// Remove trailing commas
	input = removeTrailingCommas(input)

	// Fix single quotes to double quotes
	input = strings.ReplaceAll(input, "'", "\"")

	// Fix unquoted keys
	input = fixUnquotedKeys(input)

	// Handle undefined/null values
	input = strings.ReplaceAll(input, "undefined", "null")

	return input
}

// removeTrailingCommas removes trailing commas from JSON
func removeTrailingCommas(input string) string {
	// Remove commas before closing braces/brackets
	for {
		original := input
		input = strings.ReplaceAll(input, ",}", "}")
		input = strings.ReplaceAll(input, ",]", "]")
		input = strings.ReplaceAll(input, ", }", "}")
		input = strings.ReplaceAll(input, ", ]", "]")
		if original == input {
			break
		}
	}
	return input
}

// fixUnquotedKeys attempts to fix unquoted JSON keys
func fixUnquotedKeys(input string) string {
	// This is a simplified version - a full implementation would need a proper parser
	// For now, we'll handle common cases

	// Look for patterns like: key: value
	// and convert to: "key": value
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if idx := strings.Index(trimmed, ":"); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			if !strings.HasPrefix(key, "\"") && !strings.HasSuffix(key, "\"") {
				// Check if it's likely a key (starts with letter or underscore)
				if len(key) > 0 && (isLetter(rune(key[0])) || key[0] == '_') {
					lines[i] = strings.Replace(line, key+":", "\""+key+"\":", 1)
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

// repairJSON attempts to repair malformed JSON
func repairJSON(input string) string {
	// Try to build a valid JSON structure
	result := "{}"

	// Extract key-value pairs using a simple pattern
	pairs := extractKeyValuePairs(input)

	for key, value := range pairs {
		// Use sjson to safely build JSON
		newResult, err := sjson.Set(result, key, value)
		if err == nil {
			result = newResult
		}
	}

	return result
}

// extractKeyValuePairs extracts potential key-value pairs from a string
func extractKeyValuePairs(input string) map[string]interface{} {
	pairs := make(map[string]interface{})

	// This is a simplified extraction - in production, use a proper parser
	// Look for patterns like "key": "value" or key: value
	lines := strings.Split(input, ",")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			// Clean up the key
			key = strings.Trim(key, "\"'`{[( ")

			// Clean up the value
			value = strings.Trim(value, " ")
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
				pairs[key] = value
			} else if value == "true" || value == "false" {
				pairs[key] = value == "true"
			} else if value == "null" {
				pairs[key] = nil
			} else {
				// Try to parse as number
				var num float64
				if _, err := fmt.Sscanf(value, "%f", &num); err == nil {
					pairs[key] = num
				} else {
					pairs[key] = value
				}
			}
		}
	}

	return pairs
}

// isLetter checks if a rune is a letter
func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// GenerateToolCallID generates a unique ID for tool calls
func GenerateToolCallID() string {
	// Use a simple format similar to OpenAI's tool call IDs
	return fmt.Sprintf("call_%s", GenerateUUID())
}

// GenerateUUID generates a simple UUID-like string
func GenerateUUID() string {
	// Generate 16 random bytes
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if random fails
		return fmt.Sprintf("%d", timeNow().UnixNano())
	}

	// Format as hex string
	return hex.EncodeToString(b)
}

// timeNow is a wrapper for time.Now() to make testing easier
func timeNow() time.Time {
	return time.Now()
}

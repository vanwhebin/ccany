package client

import (
	"testing"
)

func TestNewURLLogicWithOpenRouter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		finalURL string
	}{
		{
			name:     "OpenRouter without trailing slash should not add /v1",
			input:    "https://kilocode.ai/api/openrouter",
			expected: "https://kilocode.ai/api/openrouter",
			finalURL: "https://kilocode.ai/api/openrouter/chat/completions",
		},
		{
			name:     "OpenRouter with trailing slash should remove slash",
			input:    "https://kilocode.ai/api/openrouter/",
			expected: "https://kilocode.ai/api/openrouter",
			finalURL: "https://kilocode.ai/api/openrouter/chat/completions",
		},
		{
			name:     "Regular API should append /v1",
			input:    "https://api.example.com",
			expected: "https://api.example.com/v1",
			finalURL: "https://api.example.com/v1/chat/completions",
		},
		{
			name:     "Default OpenAI should be unchanged",
			input:    "https://api.openai.com/v1",
			expected: "https://api.openai.com/v1",
			finalURL: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "Together.ai should not add /v1",
			input:    "https://api.together.xyz",
			expected: "https://api.together.xyz",
			finalURL: "https://api.together.xyz/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructBaseURL(tt.input)
			if result != tt.expected {
				t.Errorf("constructBaseURL(%q) = %q, expected %q", tt.input, result, tt.expected)
			}

			finalURL := result + "/chat/completions"
			if finalURL != tt.finalURL {
				t.Errorf("Final URL for input %q would be %q, expected %q", tt.input, finalURL, tt.finalURL)
			}

			t.Logf("✅ Input: %q -> BaseURL: %q -> Final: %q", tt.input, result, finalURL)
		})
	}
}

func TestOpenRouterSpecific(t *testing.T) {
	// Test the exact case from the user
	input := "https://kilocode.ai/api/openrouter"
	result := constructBaseURL(input)
	finalURL := result + "/chat/completions"

	t.Logf("User's OpenRouter test:")
	t.Logf("  Input: %s", input)
	t.Logf("  Constructed BaseURL: %s", result)
	t.Logf("  Final URL: %s", finalURL)

	expected := "https://kilocode.ai/api/openrouter/chat/completions"
	if finalURL != expected {
		t.Errorf("OpenRouter URL construction failed. Got %q, expected %q", finalURL, expected)
	} else {
		t.Logf("✅ OpenRouter URL construction SUCCESS!")
	}
}

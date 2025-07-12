package client

import (
	"testing"
)

func TestConstructBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		finalURL string // What the final URL should be after SDK appends /chat/completions
	}{
		{
			name:     "Empty URL should return default OpenAI",
			input:    "",
			expected: "https://api.openai.com/v1",
			finalURL: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "Default OpenAI URL should be unchanged",
			input:    "https://api.openai.com/v1",
			expected: "https://api.openai.com/v1",
			finalURL: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "URL with trailing slash should remove slash",
			input:    "https://kilocode.ai/api/openrouter/",
			expected: "https://kilocode.ai/api/openrouter",
			finalURL: "https://kilocode.ai/api/openrouter/chat/completions",
		},
		{
			name:     "URL without trailing slash should append /v1",
			input:    "https://kilocode.ai/api/openrouter",
			expected: "https://kilocode.ai/api/openrouter/v1",
			finalURL: "https://kilocode.ai/api/openrouter/v1/chat/completions",
		},
		{
			name:     "OpenRouter with trailing slash (user case)",
			input:    "https://kilocode.ai/api/openrouter/",
			expected: "https://kilocode.ai/api/openrouter",
			finalURL: "https://kilocode.ai/api/openrouter/chat/completions",
		},
		{
			name:     "Another service with trailing slash",
			input:    "https://api.example.com/v1/",
			expected: "https://api.example.com/v1",
			finalURL: "https://api.example.com/v1/chat/completions",
		},
		{
			name:     "Another service without trailing slash",
			input:    "https://api.example.com",
			expected: "https://api.example.com/v1",
			finalURL: "https://api.example.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructBaseURL(tt.input)
			if result != tt.expected {
				t.Errorf("constructBaseURL(%q) = %q, expected %q", tt.input, result, tt.expected)
			}

			// Also verify what the final URL would be
			finalURL := result + "/chat/completions"
			if finalURL != tt.finalURL {
				t.Errorf("Final URL for input %q would be %q, expected %q", tt.input, finalURL, tt.finalURL)
			}

			t.Logf("Input: %q -> BaseURL: %q -> Final: %q", tt.input, result, finalURL)
		})
	}
}

// Test specifically for the user's OpenRouter issue
func TestOpenRouterURLConstruction(t *testing.T) {
	// The user has https://kilocode.ai/api/openrouter/ and expects it to call
	// https://kilocode.ai/api/openrouter/chat/completions

	input := "https://kilocode.ai/api/openrouter/"
	result := constructBaseURL(input)
	finalURL := result + "/chat/completions"
	expectedFinal := "https://kilocode.ai/api/openrouter/chat/completions"

	t.Logf("OpenRouter test:")
	t.Logf("  Input: %s", input)
	t.Logf("  Constructed BaseURL: %s", result)
	t.Logf("  Final URL (after SDK appends /chat/completions): %s", finalURL)
	t.Logf("  Expected: %s", expectedFinal)

	if finalURL != expectedFinal {
		t.Errorf("OpenRouter URL construction failed. Got %q, expected %q", finalURL, expectedFinal)
	}
}

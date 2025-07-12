package client

import (
	"testing"
)

func TestUniversalURLConstruction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		finalURL string
		reason   string
	}{
		{
			name:     "Empty URL",
			input:    "",
			expected: "https://api.openai.com/v1",
			finalURL: "https://api.openai.com/v1/chat/completions",
			reason:   "Default OpenAI endpoint",
		},
		{
			name:     "OpenAI API - already has v1",
			input:    "https://api.openai.com/v1",
			expected: "https://api.openai.com/v1",
			finalURL: "https://api.openai.com/v1/chat/completions",
			reason:   "Already contains /v1, don't add another",
		},
		{
			name:     "OpenAI API with trailing slash",
			input:    "https://api.openai.com/v1/",
			expected: "https://api.openai.com/v1",
			finalURL: "https://api.openai.com/v1/chat/completions",
			reason:   "Remove trailing slash, already has /v1",
		},
		{
			name:     "Simple domain - should add v1",
			input:    "https://api.example.com",
			expected: "https://api.example.com/v1",
			finalURL: "https://api.example.com/v1/chat/completions",
			reason:   "Simple domain needs /v1",
		},
		{
			name:     "Single path segment - should add v1",
			input:    "https://example.com/api",
			expected: "https://example.com/api/v1",
			finalURL: "https://example.com/api/v1/chat/completions",
			reason:   "Single path segment needs /v1",
		},
		{
			name:     "Two path segments - no v1 (OpenRouter case)",
			input:    "https://kilocode.ai/api/openrouter",
			expected: "https://kilocode.ai/api/openrouter",
			finalURL: "https://kilocode.ai/api/openrouter/chat/completions",
			reason:   "Two path segments indicates proxy service, no /v1 needed",
		},
		{
			name:     "Two path segments with trailing slash",
			input:    "https://kilocode.ai/api/openrouter/",
			expected: "https://kilocode.ai/api/openrouter",
			finalURL: "https://kilocode.ai/api/openrouter/chat/completions",
			reason:   "Remove trailing slash, two path segments no /v1",
		},
		{
			name:     "Three path segments - no v1",
			input:    "https://api.service.com/proxy/v2/openai",
			expected: "https://api.service.com/proxy/v2/openai",
			finalURL: "https://api.service.com/proxy/v2/openai/chat/completions",
			reason:   "Multiple path segments indicate complex routing",
		},
		{
			name:     "URL already has v1 in middle",
			input:    "https://proxy.com/v1/openai",
			expected: "https://proxy.com/v1/openai",
			finalURL: "https://proxy.com/v1/openai/chat/completions",
			reason:   "Already contains /v1, don't add another",
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

			t.Logf("✅ %s: %q -> %q -> %q (%s)", tt.name, tt.input, result, finalURL, tt.reason)
		})
	}
}

func TestUserOpenRouterCase(t *testing.T) {
	// The user's exact configuration from logs
	input := "https://kilocode.ai/api/openrouter"
	result := constructBaseURL(input)
	finalURL := result + "/chat/completions"
	expected := "https://kilocode.ai/api/openrouter/chat/completions"

	t.Logf("🎯 User's OpenRouter test:")
	t.Logf("   Input: %s", input)
	t.Logf("   Result: %s", result)
	t.Logf("   Final: %s", finalURL)
	t.Logf("   Expected: %s", expected)

	if finalURL != expected {
		t.Errorf("FAILED: Got %q, expected %q", finalURL, expected)
	} else {
		t.Logf("✅ SUCCESS: OpenRouter URL construction fixed!")
	}
}

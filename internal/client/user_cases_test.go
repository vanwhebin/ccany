package client

import (
	"testing"
)

func TestUserProvidedCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		finalURL string
	}{
		{
			name:     "Volces Ark API with /v3/ path",
			input:    "https://ark.cn-beijing.volces.com/api/v3/",
			expected: "https://ark.cn-beijing.volces.com/api/v3",
			finalURL: "https://ark.cn-beijing.volces.com/api/v3/chat/completions",
		},
		{
			name:     "X.AI API simple domain",
			input:    "https://api.x.ai",
			expected: "https://api.x.ai/v1",
			finalURL: "https://api.x.ai/v1/chat/completions",
		},
		{
			name:     "OpenRouter with existing /v1/ path",
			input:    "https://openrouter.ai/api/v1/",
			expected: "https://openrouter.ai/api/v1",
			finalURL: "https://openrouter.ai/api/v1/chat/completions",
		},
		{
			name:     "User's kilocode OpenRouter case",
			input:    "https://kilocode.ai/api/openrouter",
			expected: "https://kilocode.ai/api/openrouter",
			finalURL: "https://kilocode.ai/api/openrouter/chat/completions",
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

			t.Logf("✅ %s:", tt.name)
			t.Logf("   Input: %s", tt.input)
			t.Logf("   BaseURL: %s", result)
			t.Logf("   Final: %s", finalURL)
			t.Logf("   Expected: %s", tt.finalURL)
		})
	}
}

func TestCurrentLogicAnalysis(t *testing.T) {
	// Let's see what our current logic produces for these cases
	cases := map[string]string{
		"https://ark.cn-beijing.volces.com/api/v3/": "https://ark.cn-beijing.volces.com/api/v3/chat/completions",
		"https://api.x.ai":                          "https://api.x.ai/v1/chat/completions",
		"https://openrouter.ai/api/v1/":             "https://openrouter.ai/api/v1/chat/completions",
		"https://kilocode.ai/api/openrouter":        "https://kilocode.ai/api/openrouter/chat/completions",
	}

	t.Logf("🔍 Analyzing current logic against user test cases:")

	for input, expected := range cases {
		result := constructBaseURL(input)
		finalURL := result + "/chat/completions"

		status := "✅ PASS"
		if finalURL != expected {
			status = "❌ FAIL"
		}

		t.Logf("%s %s -> %s (expected: %s)", status, input, finalURL, expected)

		if finalURL != expected {
			t.Errorf("Logic failed for %s: got %s, expected %s", input, finalURL, expected)
		}
	}
}

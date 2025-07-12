package client

import (
	"testing"
)

func TestActualStoredURL(t *testing.T) {
	// Test what happens with the actual stored URL from the logs: https://kilocode.ai/api/openrouter
	input := "https://kilocode.ai/api/openrouter"
	result := constructBaseURL(input)
	finalURL := result + "/chat/completions"

	t.Logf("Actual stored URL test:")
	t.Logf("  Input: %s", input)
	t.Logf("  Constructed BaseURL: %s", result)
	t.Logf("  Final URL (after SDK appends /chat/completions): %s", finalURL)

	// According to our current logic, this should produce:
	// https://kilocode.ai/api/openrouter/v1/chat/completions
	// But the error suggests this is wrong and it should be:
	// https://kilocode.ai/api/openrouter/chat/completions

	if finalURL == "https://kilocode.ai/api/openrouter/v1/chat/completions" {
		t.Logf("Current logic produces: %s", finalURL)
		t.Logf("But OpenRouter expects: https://kilocode.ai/api/openrouter/chat/completions")
		t.Logf("This explains the 'Invalid path' error!")
	}
}

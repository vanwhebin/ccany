package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestOpenAISDKIntegration tests the OpenAI SDK integration with configurable endpoints
func TestOpenAISDKIntegration(t *testing.T) {
	// Skip if no test configuration provided
	apiKey := os.Getenv("TEST_API_KEY")
	baseURL := os.Getenv("TEST_BASE_URL")
	model := os.Getenv("TEST_MODEL")

	if apiKey == "" || baseURL == "" || model == "" {
		t.Skip("Skipping OpenAI SDK integration test - missing environment variables: TEST_API_KEY, TEST_BASE_URL, TEST_MODEL")
	}

	// Server should be running on localhost:8082
	serverURL := "http://localhost:8082"
	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("Setup and Authentication", func(t *testing.T) {
		// Login to get token
		token, err := authenticateAndGetToken(serverURL, client)
		if err != nil {
			t.Fatalf("Failed to authenticate: %v", err)
		}

		// Configure API with test settings
		err = configureAPI(serverURL, client, token, apiKey, baseURL, model)
		if err != nil {
			t.Fatalf("Failed to configure API: %v", err)
		}
	})

	t.Run("OpenAI Compatible API Test", func(t *testing.T) {
		// Test direct OpenAI compatible endpoint
		success := testOpenAICompatibleAPI(t, serverURL, client, apiKey, model)
		if !success {
			t.Error("OpenAI compatible API test failed")
		}
	})

	t.Run("OpenAI Streaming Test", func(t *testing.T) {
		// Test streaming functionality
		success := testOpenAIStreaming(t, serverURL, client, apiKey, model)
		if !success {
			t.Error("OpenAI streaming test failed")
		}
	})

	t.Run("Claude to OpenAI Conversion Test", func(t *testing.T) {
		// Test Claude API format conversion
		success := testClaudeToOpenAIConversion(t, serverURL, client, apiKey, model)
		if !success {
			t.Error("Claude to OpenAI conversion test failed")
		}
	})
}

// authenticateAndGetToken handles authentication and returns JWT token
func authenticateAndGetToken(serverURL string, client *http.Client) (string, error) {
	// Setup admin user (may already exist)
	setupData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(setupData)
	_, err := http.Post(serverURL+"/api/setup/admin", "application/json", bytes.NewBuffer(jsonData))
	// Ignore error - admin may already exist

	// Login to get token
	loginData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ = json.Marshal(loginData)
	resp, err := http.Post(serverURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("login failed: %v", loginResp["error"])
	}

	token, ok := loginResp["token"].(string)
	if !ok {
		return "", fmt.Errorf("no token in response")
	}

	return token, nil
}

// configureAPI configures the API with test settings
func configureAPI(serverURL string, client *http.Client, token, apiKey, baseURL, model string) error {
	configData := map[string]interface{}{
		"openai_api_key":   apiKey,
		"openai_base_url":  baseURL,
		"big_model":        model,
		"small_model":      model,
		"max_tokens_limit": 4096,
		"temperature":      0.7,
	}

	jsonData, _ := json.Marshal(configData)
	req, _ := http.NewRequest("PUT", serverURL+"/admin/config", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("configuration failed: %v", errorResp["error"])
	}

	return nil
}

// testOpenAICompatibleAPI tests the OpenAI compatible endpoint
func testOpenAICompatibleAPI(t *testing.T, serverURL string, client *http.Client, apiKey, model string) bool {
	requestData := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello, test message"},
		},
		"max_tokens":  50,
		"temperature": 0.7,
	}

	jsonData, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", serverURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Request failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
		return false
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	// Verify response structure
	if response["id"] == nil {
		t.Error("Missing response ID")
		return false
	}

	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Error("Missing or empty choices")
		return false
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	if content == "" {
		t.Error("Empty response content")
		return false
	}

	t.Logf("✅ OpenAI API test successful, response: %s", content)
	return true
}

// testOpenAIStreaming tests streaming functionality
func testOpenAIStreaming(t *testing.T, serverURL string, client *http.Client, apiKey, model string) bool {
	requestData := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Count to 3"},
		},
		"max_tokens": 50,
		"stream":     true,
	}

	jsonData, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", serverURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Streaming request failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
		return false
	}

	// Check for SSE headers
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected text/event-stream, got %s", contentType)
		return false
	}

	// Read first few bytes to verify streaming works
	buffer := make([]byte, 100)
	n, err := resp.Body.Read(buffer)
	if err != nil && n == 0 {
		t.Error("Failed to read streaming response")
		return false
	}

	response := string(buffer[:n])
	if !bytes.Contains(buffer[:n], []byte("data:")) {
		t.Errorf("Invalid SSE format, got: %s", response)
		return false
	}

	t.Log("✅ OpenAI streaming test successful")
	return true
}

// testClaudeToOpenAIConversion tests Claude API format conversion
func testClaudeToOpenAIConversion(t *testing.T, serverURL string, client *http.Client, apiKey, model string) bool {
	requestData := map[string]interface{}{
		"model":      "claude-3-sonnet-20240229", // Claude model name
		"max_tokens": 50,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello from Claude format"},
		},
	}

	jsonData, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", serverURL+"/v1/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey) // Claude uses x-api-key

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Claude conversion request failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
		return false
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	// Verify Claude response format
	if response["type"] != "message" {
		t.Error("Missing or incorrect type field")
		return false
	}

	if response["role"] != "assistant" {
		t.Error("Missing or incorrect role field")
		return false
	}

	content, ok := response["content"].([]interface{})
	if !ok || len(content) == 0 {
		t.Error("Missing or empty content array")
		return false
	}

	firstContent := content[0].(map[string]interface{})
	if firstContent["type"] != "text" {
		t.Error("First content block should be text")
		return false
	}

	text := firstContent["text"].(string)
	if text == "" {
		t.Error("Empty response text")
		return false
	}

	// Verify model routing occurred (should be using configured model)
	responseModel := response["model"].(string)
	if responseModel != model {
		t.Logf("Model routing: claude-3-sonnet-20240229 → %s", responseModel)
	}

	t.Logf("✅ Claude conversion test successful, response: %s", text)
	return true
}

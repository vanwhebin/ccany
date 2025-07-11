package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// setup test environment
	if err := os.Setenv("GIN_MODE", "test"); err != nil {
		panic(err)
	}
	if err := os.Setenv("LOG_LEVEL", "ERROR"); err != nil {
		panic(err)
	}
	if err := os.Setenv("OPENAI_API_KEY", "test-key"); err != nil {
		panic(err)
	}
	if err := os.Setenv("OPENAI_BASE_URL", "https://api.openai.com/v1"); err != nil {
		panic(err)
	}

	// run tests
	code := m.Run()

	// cleanup
	os.Exit(code)
}

// test basic health check endpoint
func TestHealthCheck(t *testing.T) {
	// create a simple HTTP client to test the actual running server
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// test health check endpoint
	resp, err := client.Get("http://localhost:8082/health")
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// test count tokens endpoint
func TestCountTokens(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	requestData := map[string]interface{}{
		"model": "claude-3-5-sonnet-20241022",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Hello, how are you?",
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	resp, err := client.Post("http://localhost:8082/v1/messages/count_tokens", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "input_tokens")
}

// test basic message endpoint
func TestBasicMessage(t *testing.T) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	requestData := map[string]interface{}{
		"model":      "claude-3-5-haiku-20241022",
		"max_tokens": 100,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Say hello",
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	resp, err := client.Post("http://localhost:8082/v1/messages", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// without valid API key, we expect an error but not a 500 error
	assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
}

// test streaming message endpoint
func TestStreamingMessage(t *testing.T) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	requestData := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 50,
		"stream":     true,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Count to 3",
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	resp, err := client.Post("http://localhost:8082/v1/messages", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// verify streaming response Content-Type
	contentType := resp.Header.Get("Content-Type")
	if resp.StatusCode == http.StatusOK {
		assert.Contains(t, contentType, "text/event-stream")
	}
}

// test tool use functionality
func TestToolUse(t *testing.T) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	requestData := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 200,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "What's the weather like?",
			},
		},
		"tools": []map[string]interface{}{
			{
				"name":        "get_weather",
				"description": "Get the current weather for a location",
				"input_schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	resp, err := client.Post("http://localhost:8082/v1/messages", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// verify request format is correct
	assert.NotEqual(t, http.StatusBadRequest, resp.StatusCode)
}

// test multimodal input
func TestMultimodalInput(t *testing.T) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 1x1 pixel transparent PNG base64 encoded
	sampleImage := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAI9jU8PJAAAAASUVORK5CYII="

	requestData := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 100,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "What do you see in this image?",
					},
					{
						"type": "image",
						"source": map[string]interface{}{
							"type":       "base64",
							"media_type": "image/png",
							"data":       sampleImage,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	resp, err := client.Post("http://localhost:8082/v1/messages", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// verify request format is correct
	assert.NotEqual(t, http.StatusBadRequest, resp.StatusCode)
}

// test system message
func TestSystemMessage(t *testing.T) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	requestData := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 100,
		"system":     "You are a helpful assistant that responds in haiku format.",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Explain AI",
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	resp, err := client.Post("http://localhost:8082/v1/messages", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// verify request format is correct
	assert.NotEqual(t, http.StatusBadRequest, resp.StatusCode)
}

// test invalid request
func TestInvalidRequest(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// send invalid JSON
	resp, err := client.Post("http://localhost:8082/v1/messages", "application/json", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// test missing required fields
func TestMissingRequiredFields(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// send request without model field
	requestData := map[string]interface{}{
		"max_tokens": 100,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Hello",
			},
		},
	}

	jsonData, err := json.Marshal(requestData)
	require.NoError(t, err)

	resp, err := client.Post("http://localhost:8082/v1/messages", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Skip("server not running, skipping integration test")
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// benchmark tests
func BenchmarkHealthCheck(b *testing.B) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get("http://localhost:8082/health")
		if err != nil {
			b.Skip("server not running, skipping benchmark test")
			return
		}
		if err := resp.Body.Close(); err != nil {
			b.Logf("Failed to close response body: %v", err)
		}
	}
}

func BenchmarkCountTokens(b *testing.B) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	requestData := map[string]interface{}{
		"model": "claude-3-5-sonnet-20241022",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Hello, how are you?",
			},
		},
	}

	jsonData, _ := json.Marshal(requestData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Post("http://localhost:8082/v1/messages/count_tokens", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			b.Skip("server not running, skipping benchmark test")
			return
		}
		if err := resp.Body.Close(); err != nil {
			b.Logf("Failed to close response body: %v", err)
		}
	}
}

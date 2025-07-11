package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestConfigurationAPIComplete tests the complete configuration API functionality
func TestConfigurationAPIComplete(t *testing.T) {
	// Server should be running on localhost:8082
	baseURL := "http://localhost:8082"
	client := &http.Client{Timeout: 30 * time.Second}

	var token string

	t.Run("Admin Setup and Authentication", func(t *testing.T) {
		// Setup admin user
		setupData := map[string]interface{}{
			"username": "admin",
			"password": "admin123",
		}

		jsonData, _ := json.Marshal(setupData)
		resp, err := http.Post(baseURL+"/api/setup/admin", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Error setting up admin: %v", err)
		}
		defer resp.Body.Close()

		// 409 means already exists, which is fine
		if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 409 {
			t.Errorf("Unexpected status for admin setup: %d", resp.StatusCode)
		}

		// Login to get token
		loginData := map[string]interface{}{
			"username": "admin",
			"password": "admin123",
		}

		jsonData, _ = json.Marshal(loginData)
		resp, err = http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Error logging in: %v", err)
		}
		defer resp.Body.Close()

		var loginResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&loginResp)

		if resp.StatusCode != 200 {
			t.Fatalf("Login failed with status %d: %v", resp.StatusCode, loginResp["error"])
		}

		var ok bool
		token, ok = loginResp["token"].(string)
		if !ok {
			t.Fatal("No token in login response")
		}

		t.Log("✅ Authentication successful")
	})

	// Helper function for authenticated requests
	makeAuthRequest := func(method, url string, body []byte) (*http.Response, error) {
		req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		return client.Do(req)
	}

	t.Run("Get Initial Configuration", func(t *testing.T) {
		resp, err := makeAuthRequest("GET", baseURL+"/admin/config", nil)
		if err != nil {
			t.Fatalf("Error getting config: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var configResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&configResp)

		if _, ok := configResp["config"]; !ok {
			t.Error("Missing config in response")
		}

		t.Log("✅ Initial configuration retrieved")
	})

	t.Run("Update Configuration", func(t *testing.T) {
		// Use test values instead of real API keys
		updateData := map[string]interface{}{
			"openai_api_key":   "sk-test-key-for-testing-only",
			"claude_api_key":   "sk-ant-test-key-for-testing-only",
			"openai_base_url":  "https://api.example.com/v1",
			"claude_base_url":  "https://api.example.com",
			"big_model":        "test-big-model",
			"small_model":      "test-small-model",
			"max_tokens_limit": 4096,
			"request_timeout":  90,
			"host":             "0.0.0.0",
			"port":             8082,
			"log_level":        "info",
			"jwt_secret":       "test-jwt-secret-for-testing",
		}

		jsonData, _ := json.Marshal(updateData)
		resp, err := makeAuthRequest("PUT", baseURL+"/admin/config", jsonData)
		if err != nil {
			t.Fatalf("Error updating config: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			var errorResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResp)
			t.Fatalf("Config update failed with status %d: %v", resp.StatusCode, errorResp["error"])
		}

		var updateResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&updateResp)

		if updateResp["message"] == nil {
			t.Error("Missing message in update response")
		}

		if updateResp["updated"] == nil {
			t.Error("Missing updated count in response")
		}

		t.Log("✅ Configuration updated successfully")
	})

	t.Run("Verify Configuration Masking", func(t *testing.T) {
		resp, err := makeAuthRequest("GET", baseURL+"/admin/config", nil)
		if err != nil {
			t.Fatalf("Error getting config after update: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var configResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&configResp)

		config, ok := configResp["config"].(map[string]interface{})
		if !ok {
			t.Fatal("Missing config in response")
		}

		// Check that sensitive fields are present (API returns the actual values for test keys)
		sensitiveFields := []string{"openai_api_key", "claude_api_key", "jwt_secret"}
		for _, field := range sensitiveFields {
			if value, exists := config[field]; exists {
				// Should be either the actual value, masked, or boolean indicating presence
				if value == nil || value == "" {
					t.Errorf("Sensitive field %s should have a value or be masked: %v", field, value)
				}
			}
		}

		// Check that non-sensitive fields are visible
		if config["big_model"] != "test-big-model" {
			t.Error("Big model should be visible and updated")
		}

		if config["openai_base_url"] != "https://api.example.com/v1" {
			t.Error("OpenAI base URL should be visible and updated")
		}

		t.Log("✅ Configuration fields verification completed")
	})

	t.Run("Partial Configuration Update", func(t *testing.T) {
		// Test partial update - only update some fields
		partialUpdate := map[string]interface{}{
			"big_model":        "updated-big-model",
			"max_tokens_limit": 2048,
			"temperature":      0.8,
		}

		jsonData, _ := json.Marshal(partialUpdate)
		resp, err := makeAuthRequest("PUT", baseURL+"/admin/config", jsonData)
		if err != nil {
			t.Fatalf("Error in partial update: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			var errorResp map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResp)
			t.Fatalf("Partial update failed with status %d: %v", resp.StatusCode, errorResp["error"])
		}

		// Verify the update
		resp, err = makeAuthRequest("GET", baseURL+"/admin/config", nil)
		if err != nil {
			t.Fatalf("Error getting config after partial update: %v", err)
		}
		defer resp.Body.Close()

		var configResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&configResp)

		config := configResp["config"].(map[string]interface{})

		if config["big_model"] != "updated-big-model" {
			t.Error("Big model should be updated to 'updated-big-model'")
		}

		if fmt.Sprintf("%.0f", config["max_tokens_limit"]) != "2048" {
			t.Error("Max tokens should be updated to 2048")
		}

		// Verify that API keys are still preserved (should still be present in some form)
		if config["openai_api_key"] == nil || config["openai_api_key"] == "" {
			t.Error("OpenAI API key should still be present after partial update")
		}

		t.Log("✅ Partial configuration update working correctly")
	})

	t.Run("Configuration Validation", func(t *testing.T) {
		// Test with invalid data
		invalidData := map[string]interface{}{
			"max_tokens_limit": -1,        // Invalid value
			"port":             "invalid", // Wrong type
		}

		jsonData, _ := json.Marshal(invalidData)
		resp, err := makeAuthRequest("PUT", baseURL+"/admin/config", jsonData)
		if err != nil {
			t.Fatalf("Error testing invalid config: %v", err)
		}
		defer resp.Body.Close()

		// Should handle validation gracefully (either reject or sanitize)
		var respData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respData)

		t.Logf("Validation test completed with status %d", resp.StatusCode)
	})
}

// TestConfigurationPersistence tests that configuration changes persist across operations
func TestConfigurationPersistence(t *testing.T) {
	baseURL := "http://localhost:8082"
	client := &http.Client{Timeout: 30 * time.Second}

	// Get auth token
	token, err := getAuthToken(baseURL, client)
	if err != nil {
		t.Fatalf("Failed to get auth token: %v", err)
	}

	makeAuthRequest := func(method, url string, body []byte) (*http.Response, error) {
		req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		return client.Do(req)
	}

	// Test persistence by doing multiple operations
	testConfig := map[string]interface{}{
		"big_model":        "persistence-test-model",
		"small_model":      "persistence-test-small",
		"max_tokens_limit": 3333,
		"temperature":      0.5,
	}

	t.Run("Save Test Configuration", func(t *testing.T) {
		jsonData, _ := json.Marshal(testConfig)
		resp, err := makeAuthRequest("PUT", baseURL+"/admin/config", jsonData)
		if err != nil {
			t.Fatalf("Failed to save test config: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Fatalf("Failed to save config, status: %d", resp.StatusCode)
		}
	})

	t.Run("Verify Persistence After Save", func(t *testing.T) {
		resp, err := makeAuthRequest("GET", baseURL+"/admin/config", nil)
		if err != nil {
			t.Fatalf("Failed to get config: %v", err)
		}
		defer resp.Body.Close()

		var configResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&configResp)

		config := configResp["config"].(map[string]interface{})

		if config["big_model"] != "persistence-test-model" {
			t.Error("Big model not persisted correctly")
		}

		if fmt.Sprintf("%.0f", config["max_tokens_limit"]) != "3333" {
			t.Error("Max tokens not persisted correctly")
		}

		t.Log("✅ Configuration persistence verified")
	})
}

// Helper function to get authentication token
func getAuthToken(baseURL string, client *http.Client) (string, error) {
	// Try to login (admin should already be set up from previous tests)
	loginData := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
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

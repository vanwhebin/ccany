package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ccany/internal/models"

	"github.com/sirupsen/logrus"
)

// OpenAIClient handles communication with OpenAI API
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, baseURL string, timeout int, logger *logrus.Logger) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		timeout: time.Duration(timeout) * time.Second,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		logger: logger,
	}
}

// CreateChatCompletion sends a chat completion request to OpenAI
func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, req *models.OpenAIChatCompletionRequest) (*models.OpenAIChatCompletionResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"model":      req.Model,
		"stream":     req.Stream,
		"max_tokens": req.MaxTokens,
	}).Debug("Sending OpenAI chat completion request")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("Failed to close response body")
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp models.OpenAIErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			return nil, fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var openaiResp models.OpenAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"response_id":       openaiResp.ID,
		"model":             openaiResp.Model,
		"prompt_tokens":     openaiResp.Usage.PromptTokens,
		"completion_tokens": openaiResp.Usage.CompletionTokens,
	}).Debug("Received OpenAI chat completion response")

	return &openaiResp, nil
}

// CreateChatCompletionStream sends a streaming chat completion request to OpenAI
func (c *OpenAIClient) CreateChatCompletionStream(ctx context.Context, req *models.OpenAIChatCompletionRequest) (<-chan StreamChunk, error) {
	req.Stream = true

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"model":      req.Model,
		"stream":     true,
		"max_tokens": req.MaxTokens,
	}).Debug("Sending OpenAI streaming chat completion request")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.logger.WithError(err).Warn("Failed to close response body")
			}
		}()
		respBody, _ := io.ReadAll(resp.Body)
		var errorResp models.OpenAIErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			return nil, fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	streamChan := make(chan StreamChunk, 10)

	go func() {
		defer close(streamChan)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.logger.WithError(err).Warn("Failed to close response body")
			}
		}()

		c.processStream(ctx, resp.Body, streamChan)
	}()

	return streamChan, nil
}

// StreamChunk represents a chunk from the streaming response
type StreamChunk struct {
	Data  *models.OpenAIStreamResponse
	Error error
	Done  bool
}

// processStream processes the SSE stream from OpenAI
func (c *OpenAIClient) processStream(ctx context.Context, body io.Reader, streamChan chan<- StreamChunk) {
	scanner := NewSSEScanner(body)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			streamChan <- StreamChunk{Error: ctx.Err(), Done: true}
			return
		default:
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				streamChan <- StreamChunk{Done: true}
				return
			}

			var chunk models.OpenAIStreamResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				c.logger.WithError(err).Warn("Failed to unmarshal stream chunk")
				continue
			}

			streamChan <- StreamChunk{Data: &chunk}
		}
	}

	if err := scanner.Err(); err != nil {
		streamChan <- StreamChunk{Error: err, Done: true}
	}
}

// setHeaders sets the required headers for OpenAI API requests
func (c *OpenAIClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "ccany/1.0.0")
}

// ClassifyError classifies OpenAI API errors
func (c *OpenAIClient) ClassifyError(err error) string {
	errStr := err.Error()

	if strings.Contains(errStr, "rate limit") {
		return "rate_limit_error"
	}
	if strings.Contains(errStr, "insufficient_quota") || strings.Contains(errStr, "quota") {
		return "insufficient_quota"
	}
	if strings.Contains(errStr, "invalid_api_key") || strings.Contains(errStr, "authentication") {
		return "authentication_error"
	}
	if strings.Contains(errStr, "model_not_found") || strings.Contains(errStr, "model") {
		return "not_found_error"
	}
	if strings.Contains(errStr, "context_length") || strings.Contains(errStr, "too long") {
		return "invalid_request_error"
	}

	return "api_error"
}

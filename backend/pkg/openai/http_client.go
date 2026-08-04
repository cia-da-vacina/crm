package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.openai.com/v1"

// HTTPClient chama a Chat Completions API de verdade. BaseURL é
// override-ável só pra teste (aponta pra um httptest.Server local) — em
// produção sempre é defaultBaseURL.
type HTTPClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(apiKey string) *HTTPClient {
	return NewHTTPClientWithBaseURL(apiKey, defaultBaseURL)
}

func NewHTTPClientWithBaseURL(apiKey, baseURL string) *HTTPClient {
	return &HTTPClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *HTTPClient) Complete(ctx context.Context, req CompletionRequest) (CompletionResult, error) {
	body := chatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
	}
	if req.JSONResponse {
		body.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return CompletionResult{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return CompletionResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return CompletionResult{}, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CompletionResult{}, err
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return CompletionResult{}, fmt.Errorf("failed to parse openai response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("openai returned status %d", resp.StatusCode)
		if parsed.Error != nil {
			msg = parsed.Error.Message
		}
		return CompletionResult{}, fmt.Errorf("%s", msg)
	}

	if len(parsed.Choices) == 0 {
		return CompletionResult{}, fmt.Errorf("openai response has no choices")
	}

	return CompletionResult{Content: parsed.Choices[0].Message.Content}, nil
}

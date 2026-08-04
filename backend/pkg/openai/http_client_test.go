package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClient_Complete_Success(t *testing.T) {
	var capturedBody chatRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("unexpected Authorization header: %s", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{"message": {"content": "{\"intent\":\"agendar\"}"}}]
		}`))
	}))
	defer srv.Close()

	client := NewHTTPClientWithBaseURL("test-key", srv.URL)
	result, err := client.Complete(context.Background(), CompletionRequest{
		Model:        "gpt-4o-mini",
		Messages:     []Message{{Role: "system", Content: "you are a triage bot"}, {Role: "user", Content: "oi"}},
		JSONResponse: true,
	})
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if result.Content != `{"intent":"agendar"}` {
		t.Fatalf("unexpected content: %q", result.Content)
	}

	if capturedBody.Model != "gpt-4o-mini" {
		t.Errorf("unexpected model sent: %s", capturedBody.Model)
	}
	if len(capturedBody.Messages) != 2 {
		t.Errorf("expected 2 messages sent, got %d", len(capturedBody.Messages))
	}
	if capturedBody.ResponseFormat == nil || capturedBody.ResponseFormat.Type != "json_object" {
		t.Errorf("expected response_format json_object to be set")
	}
}

func TestHTTPClient_Complete_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": {"message": "invalid api key"}}`))
	}))
	defer srv.Close()

	client := NewHTTPClientWithBaseURL("bad-key", srv.URL)
	_, err := client.Complete(context.Background(), CompletionRequest{Model: "gpt-4o-mini"})
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}

func TestHTTPClient_Complete_NoChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices": []}`))
	}))
	defer srv.Close()

	client := NewHTTPClientWithBaseURL("test-key", srv.URL)
	_, err := client.Complete(context.Background(), CompletionRequest{Model: "gpt-4o-mini"})
	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}
}

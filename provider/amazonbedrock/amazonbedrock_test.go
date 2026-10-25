package amazonbedrock

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestComplete_Success(t *testing.T) {
	// Mock Bedrock API server
	mockResponse := `{
		"id": "test-id",
		"object": "text_completion",
		"created": 1234567890,
		"choices": [
			{
				"text": "Hello from Bedrock!",
				"index": 0,
				"finish_reason": "stop"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("missing or wrong API key header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	resp, err := client.Complete(context.Background(), CompletionRequest{
		Model:  "bedrock-model",
		Prompt: "Hello",
	})
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	expected := "Hello from Bedrock!"
	if resp != expected {
		t.Errorf("expected %q, got %q", expected, resp)
	}
}

func TestComplete_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	_, err := client.Complete(context.Background(), CompletionRequest{
		Model:  "bedrock-model",
		Prompt: "Hello",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestComplete_NoChoices(t *testing.T) {
	mockResponse := `{
		"id": "test-id",
		"object": "text_completion",
		"created": 1234567890,
		"choices": []
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	_, err := client.Complete(context.Background(), CompletionRequest{
		Model:  "bedrock-model",
		Prompt: "Hello",
	})
	if err == nil {
		t.Fatal("expected error for no choices, got nil")
	}
}

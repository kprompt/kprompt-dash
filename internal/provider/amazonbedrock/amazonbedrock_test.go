package amazonbedrock

import (
	"context"
	"errors"
	"testing"
)

type mockClient struct {
	response string
	err      error
}

func (m *mockClient) InvokeModel(ctx context.Context, modelID string, input string) (string, error) {
	return m.response, m.err
}

func TestProvider_Generate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		modelID   string
		prompt    string
		client    Client
		want      string
		wantError bool
	}{
		{
			name:    "success",
			modelID: "test-model",
			prompt:  "hello",
			client:  &mockClient{response: "response", err: nil},
			want:    "response",
		},
		{
			name:      "empty modelID",
			modelID:   "",
			prompt:    "hello",
			client:    &mockClient{},
			wantError: true,
		},
		{
			name:      "empty prompt",
			modelID:   "model",
			prompt:    "",
			client:    &mockClient{},
			wantError: true,
		},
		{
			name:      "client error",
			modelID:   "model",
			prompt:    "hello",
			client:    &mockClient{err: errors.New("fail")},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.client)
			got, err := p.Generate(ctx, tt.modelID, tt.prompt)
			if (err != nil) != tt.wantError {
				t.Fatalf("Generate() error = %v, wantError %v", err, tt.wantError)
			}
			if got != tt.want {
				t.Errorf("Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}

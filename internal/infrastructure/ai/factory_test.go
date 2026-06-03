package ai

import (
	"strings"
	"testing"

	"ai-werewolf-go/internal/config"
)

func TestBuildProviderReturnsFallbackProvider(t *testing.T) {
	provider, err := BuildProvider(config.AIConfig{Provider: "fallback", Concurrency: 1})
	if err != nil {
		t.Fatalf("BuildProvider() error = %v", err)
	}
	if _, ok := provider.(FallbackProvider); !ok {
		t.Fatalf("BuildProvider() type = %T, want ai.FallbackProvider", provider)
	}
}

func TestBuildProviderRejectsUnknownProvider(t *testing.T) {
	_, err := BuildProvider(config.AIConfig{Provider: "unknown", Concurrency: 1})
	if err == nil {
		t.Fatal("BuildProvider() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "unsupported ai.provider") {
		t.Fatalf("BuildProvider() error = %v, want unsupported ai.provider", err)
	}
}

func TestBuildProviderReturnsEinoProvider(t *testing.T) {
	provider, err := BuildProvider(config.AIConfig{
		Provider:    "eino",
		BaseURL:     "https://api.example.com/v1",
		Model:       "deepseek-chat",
		APIKey:      "test-key",
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("BuildProvider() error = %v", err)
	}
	if _, ok := provider.(*EinoProvider); !ok {
		t.Fatalf("BuildProvider() type = %T, want *ai.EinoProvider", provider)
	}
}

func TestBuildProviderRejectsIncompleteEinoConfig(t *testing.T) {
	_, err := BuildProvider(config.AIConfig{Provider: "eino", Concurrency: 1})
	if err == nil {
		t.Fatal("BuildProvider() error = nil, want error")
	}
}

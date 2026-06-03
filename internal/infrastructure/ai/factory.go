package ai

import (
	"context"
	"fmt"

	"ai-werewolf-go/internal/config"
	"ai-werewolf-go/internal/domain"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
)

func BuildProvider(cfg config.AIConfig) (domain.DecisionProvider, error) {
	switch cfg.Provider {
	case "fallback":
		return FallbackProvider{}, nil
	case "eino":
		if cfg.Model == "" {
			return nil, fmt.Errorf("build eino provider: ai.model is required")
		}
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("build eino provider: ai.api_key is required")
		}

		chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
			APIKey:  cfg.APIKey,
			Model:   cfg.Model,
			BaseURL: cfg.BaseURL,
		})
		if err != nil {
			return nil, fmt.Errorf("build eino provider: %w", err)
		}
		return NewEinoProvider(chatModel), nil
	default:
		return nil, fmt.Errorf("unsupported ai.provider: %s", cfg.Provider)
	}
}

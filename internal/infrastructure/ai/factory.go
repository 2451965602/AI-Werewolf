package ai

import (
	"context"
	"fmt"
	"time"

	"ai-werewolf-go/internal/config"
	"ai-werewolf-go/internal/domain"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
)

func BuildProvider(cfg config.AIConfig) (domain.DecisionProvider, error) {
	primary, err := buildSingleProvider(cfg.Provider, cfg.BaseURL, cfg.Model, cfg.APIKey, cfg.TimeoutMS)
	if err != nil {
		return nil, err
	}

	if cfg.Fallback == nil {
		return primary, nil
	}

	fallback, err := buildSingleProvider(cfg.Fallback.Provider, cfg.Fallback.BaseURL, cfg.Fallback.Model, cfg.Fallback.APIKey, cfg.Fallback.TimeoutMS)
	if err != nil {
		return nil, fmt.Errorf("build fallback provider: %w", err)
	}

	return NewFailoverProvider(primary, fallback), nil
}

func buildSingleProvider(provider string, baseURL string, modelName string, apiKey string, timeoutMS int) (domain.DecisionProvider, error) {
	timeout := time.Duration(timeoutMS) * time.Millisecond

	switch provider {
	case "fallback":
		return FallbackProvider{}, nil
	case "eino":
		if modelName == "" {
			return nil, fmt.Errorf("build eino provider: ai.model is required")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("build eino provider: ai.api_key is required")
		}

		chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
			APIKey:  apiKey,
			Model:   modelName,
			BaseURL: baseURL,
			Timeout: timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("build eino provider: %w", err)
		}
		return NewEinoProvider(chatModel, timeout), nil
	default:
		return nil, fmt.Errorf("unsupported ai.provider: %s", provider)
	}
}

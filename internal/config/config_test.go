package config

import (
	"os"
	"path/filepath"
	"testing"
)

func clearRelevantEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"CONFIG_PATH",
		"OPENAI_API_KEY",
		"ALT_API_KEY",
		"MY_API_KEY",
		"MY_CUSTOM_KEY",
		"MISSING_KEY_VAR",
		"WEREWOLF_SERVER_ADDR",
		"WEREWOLF_STORAGE_STATE_PATH",
		"WEREWOLF_AI_PROVIDER",
		"WEREWOLF_AI_BASE_URL",
		"WEREWOLF_AI_MODEL",
		"WEREWOLF_AI_CONCURRENCY",
		"WEREWOLF_AI_API_KEY",
		"WEREWOLF_AI_API_KEY_ENV",
		"WEREWOLF_AI_TIMEOUT_MS",
		"WEREWOLF_AI_FALLBACK_PROVIDER",
		"WEREWOLF_AI_FALLBACK_BASE_URL",
		"WEREWOLF_AI_FALLBACK_MODEL",
		"WEREWOLF_AI_FALLBACK_API_KEY",
		"WEREWOLF_AI_FALLBACK_API_KEY_ENV",
		"WEREWOLF_AI_FALLBACK_TIMEOUT_MS",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadReadsAIFallbackConfig(t *testing.T) {
	clearRelevantEnv(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
ai:
  provider: "eino"
  base_url: "http://primary/v1"
  model: "primary-model"
  api_key: "primary-key"
  concurrency: 1
  timeout_ms: 5000
  fallback:
    provider: "fallback"
    timeout_ms: 1000
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath() returned error: %v", err)
	}

	if cfg.AI.TimeoutMS != 5000 {
		t.Fatalf("AI.TimeoutMS = %d, want 5000", cfg.AI.TimeoutMS)
	}
	if cfg.AI.Fallback == nil {
		t.Fatal("AI.Fallback = nil, want fallback config")
	}
	if cfg.AI.Fallback.Provider != "fallback" {
		t.Fatalf("AI.Fallback.Provider = %q, want fallback", cfg.AI.Fallback.Provider)
	}
	if cfg.AI.Fallback.TimeoutMS != 1000 {
		t.Fatalf("AI.Fallback.TimeoutMS = %d, want 1000", cfg.AI.Fallback.TimeoutMS)
	}
}

func TestLoadDefaultsWhenDefaultConfigMissing(t *testing.T) {
	clearRelevantEnv(t)

	// 使用临时目录确保无 config.yaml
	dir := t.TempDir()
	t.Setenv("CONFIG_PATH", "")
	// 切换到临时目录以确保没有默认 config.yaml
	originalDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(originalDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// 默认值校验
	if cfg.Server.Addr != ":8080" {
		t.Errorf("Server.Addr = %q, want :8080", cfg.Server.Addr)
	}
	if cfg.Storage.StatePath != "data/world_state.json" {
		t.Errorf("Storage.StatePath = %q, want data/world_state.json", cfg.Storage.StatePath)
	}
	if cfg.AI.Provider != "fallback" {
		t.Errorf("AI.Provider = %q, want fallback", cfg.AI.Provider)
	}
	if cfg.AI.APIKeyEnv != "OPENAI_API_KEY" {
		t.Errorf("AI.APIKeyEnv = %q, want OPENAI_API_KEY", cfg.AI.APIKeyEnv)
	}
	if cfg.AI.APIKey != "" {
		t.Errorf("AI.APIKey = %q, want empty", cfg.AI.APIKey)
	}
	if cfg.AI.BaseURL != "" {
		t.Errorf("AI.BaseURL = %q, want empty", cfg.AI.BaseURL)
	}
	if cfg.AI.Model != "" {
		t.Errorf("AI.Model = %q, want empty", cfg.AI.Model)
	}
	if cfg.AI.Concurrency != 1 {
		t.Errorf("AI.Concurrency = %d, want 1", cfg.AI.Concurrency)
	}
}

func TestLoadConfigFileOverridesDefaults(t *testing.T) {
	clearRelevantEnv(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  addr: ":9090"
storage:
  state_path: "data/from-file.json"
ai:
  provider: "openai"
  base_url: "https://api.openai.com"
  model: "gpt-4"
  api_key: "file-secret"
  api_key_env: "MY_API_KEY"
  concurrency: 3
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath() returned error: %v", err)
	}

	if cfg.Server.Addr != ":9090" {
		t.Errorf("Server.Addr = %q, want :9090", cfg.Server.Addr)
	}
	if cfg.Storage.StatePath != "data/from-file.json" {
		t.Errorf("Storage.StatePath = %q, want data/from-file.json", cfg.Storage.StatePath)
	}
	if cfg.AI.Provider != "openai" {
		t.Errorf("AI.Provider = %q, want openai", cfg.AI.Provider)
	}
	if cfg.AI.BaseURL != "https://api.openai.com" {
		t.Errorf("AI.BaseURL = %q, want https://api.openai.com", cfg.AI.BaseURL)
	}
	if cfg.AI.Model != "gpt-4" {
		t.Errorf("AI.Model = %q, want gpt-4", cfg.AI.Model)
	}
	if cfg.AI.APIKey != "file-secret" {
		t.Errorf("AI.APIKey = %q, want file-secret", cfg.AI.APIKey)
	}
	if cfg.AI.APIKeyEnv != "MY_API_KEY" {
		t.Errorf("AI.APIKeyEnv = %q, want MY_API_KEY", cfg.AI.APIKeyEnv)
	}
	if cfg.AI.Concurrency != 3 {
		t.Errorf("AI.Concurrency = %d, want 3", cfg.AI.Concurrency)
	}
}

func TestLoadConfigFileReadsAIConcurrency(t *testing.T) {
	clearRelevantEnv(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
ai:
  provider: "fallback"
  concurrency: 1
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath() returned error: %v", err)
	}

	if cfg.AI.Concurrency != 1 {
		t.Errorf("AI.Concurrency = %d, want 1", cfg.AI.Concurrency)
	}
}

func TestLoadRejectsNonPositiveAIConcurrency(t *testing.T) {
	clearRelevantEnv(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
ai:
  provider: "fallback"
  concurrency: 0
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadFromPath(configPath)
	if err == nil {
		t.Fatal("LoadFromPath() with ai.concurrency=0 should return error, got nil")
	}
}

func TestEnvironmentOverridesConfigFile(t *testing.T) {
	clearRelevantEnv(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  addr: ":9090"
storage:
  state_path: "data/from-file.json"
ai:
  provider: "file-provider"
  base_url: "https://file.example.com"
  model: "file-model"
  api_key_env: "MY_API_KEY"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("WEREWOLF_SERVER_ADDR", ":7070")
	t.Setenv("WEREWOLF_STORAGE_STATE_PATH", "data/from-env.json")
	t.Setenv("WEREWOLF_AI_PROVIDER", "env-provider")
	t.Setenv("WEREWOLF_AI_BASE_URL", "https://env.example.com")
	t.Setenv("WEREWOLF_AI_MODEL", "env-model")
	t.Setenv("WEREWOLF_AI_CONCURRENCY", "2")
	t.Setenv("WEREWOLF_AI_API_KEY_ENV", "ALT_API_KEY")
	t.Setenv("ALT_API_KEY", "env-named-secret")

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadFromPath() returned error: %v", err)
	}

	if cfg.Server.Addr != ":7070" {
		t.Errorf("Server.Addr = %q, want :7070", cfg.Server.Addr)
	}
	if cfg.Storage.StatePath != "data/from-env.json" {
		t.Errorf("Storage.StatePath = %q, want data/from-env.json", cfg.Storage.StatePath)
	}
	if cfg.AI.Provider != "env-provider" {
		t.Errorf("AI.Provider = %q, want env-provider", cfg.AI.Provider)
	}
	if cfg.AI.BaseURL != "https://env.example.com" {
		t.Errorf("AI.BaseURL = %q, want https://env.example.com", cfg.AI.BaseURL)
	}
	if cfg.AI.Model != "env-model" {
		t.Errorf("AI.Model = %q, want env-model", cfg.AI.Model)
	}
	if cfg.AI.Concurrency != 2 {
		t.Errorf("AI.Concurrency = %d, want 2", cfg.AI.Concurrency)
	}
	if cfg.AI.APIKeyEnv != "ALT_API_KEY" {
		t.Errorf("AI.APIKeyEnv = %q, want ALT_API_KEY", cfg.AI.APIKeyEnv)
	}
	if cfg.AI.APIKey != "env-named-secret" {
		t.Errorf("AI.APIKey = %q, want env-named-secret", cfg.AI.APIKey)
	}
}

func TestExplicitConfigPathMissingReturnsError(t *testing.T) {
	clearRelevantEnv(t)
	missingPath := filepath.Join(t.TempDir(), "missing-config.yaml")

	// LoadFromPath 指向不存在的文件
	_, err := LoadFromPath(missingPath)
	if err == nil {
		t.Error("LoadFromPath() with missing file should return error, got nil")
	}

	// Load() 带 CONFIG_PATH 指向不存在的文件
	t.Setenv("CONFIG_PATH", missingPath)
	_, err = Load()
	if err == nil {
		t.Error("Load() with missing CONFIG_PATH should return error, got nil")
	}
}

func TestDefaultConfigDirectoryReturnsError(t *testing.T) {
	clearRelevantEnv(t)

	dir := t.TempDir()
	configDir := filepath.Join(dir, "config.yaml")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	originalDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(originalDir)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() with config.yaml directory should return error, got nil")
	}
}

func TestInvalidYAMLReturnsError(t *testing.T) {
	clearRelevantEnv(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  addr: ":9090"
  invalid: [missing bracket
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadFromPath(configPath)
	if err == nil {
		t.Error("LoadFromPath() with invalid YAML should return error, got nil")
	}
}

func TestAIAPIKeyResolutionOrder(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	t.Run("WEREWOLF_AI_API_KEY overrides config file ai.api_key", func(t *testing.T) {
		clearRelevantEnv(t)

		content := `
ai:
  api_key: file-secret
  api_key_env: OPENAI_API_KEY
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		t.Setenv("WEREWOLF_AI_API_KEY", "env-direct-secret")

		cfg, err := LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadFromPath() returned error: %v", err)
		}
		if cfg.AI.APIKey != "env-direct-secret" {
			t.Errorf("AI.APIKey = %q, want env-direct-secret", cfg.AI.APIKey)
		}
	})

	t.Run("config file ai.api_key used when no WEREWOLF_AI_API_KEY", func(t *testing.T) {
		clearRelevantEnv(t)

		content := `
ai:
  api_key: file-secret
  api_key_env: OPENAI_API_KEY
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		cfg, err := LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadFromPath() returned error: %v", err)
		}
		if cfg.AI.APIKey != "file-secret" {
			t.Errorf("AI.APIKey = %q, want file-secret", cfg.AI.APIKey)
		}
	})

	t.Run("ai.api_key_env environment variable as fallback", func(t *testing.T) {
		clearRelevantEnv(t)

		content := `
ai:
  api_key_env: MY_CUSTOM_KEY
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		t.Setenv("MY_CUSTOM_KEY", "env-named-secret")

		cfg, err := LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadFromPath() returned error: %v", err)
		}
		if cfg.AI.APIKey != "env-named-secret" {
			t.Errorf("AI.APIKey = %q, want env-named-secret", cfg.AI.APIKey)
		}
	})

	t.Run("empty API key when nothing configured", func(t *testing.T) {
		clearRelevantEnv(t)

		content := `
ai:
  api_key_env: MISSING_KEY_VAR
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		cfg, err := LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadFromPath() returned error: %v", err)
		}
		if cfg.AI.APIKey != "" {
			t.Errorf("AI.APIKey = %q, want empty", cfg.AI.APIKey)
		}
	})
}

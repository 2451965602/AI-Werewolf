package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 顶层运行时配置结构
type Config struct {
	Server  ServerConfig
	Storage StorageConfig
	AI      AIConfig
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

// StorageConfig 持久化存储配置
type StorageConfig struct {
	StatePath string `mapstructure:"state_path"`
}

// AIConfig AI 服务配置
type AIConfig struct {
	Provider    string `mapstructure:"provider"`
	BaseURL     string `mapstructure:"base_url"`
	Model       string `mapstructure:"model"`
	Concurrency int    `mapstructure:"concurrency"`
	APIKey      string `mapstructure:"api_key"`
	APIKeyEnv   string `mapstructure:"api_key_env"`
	TimeoutMS   int    `mapstructure:"timeout_ms"`
	Fallback    *FallbackAIConfig
}

type FallbackAIConfig struct {
	Provider  string `mapstructure:"provider"`
	BaseURL   string `mapstructure:"base_url"`
	Model     string `mapstructure:"model"`
	APIKey    string `mapstructure:"api_key"`
	APIKeyEnv string `mapstructure:"api_key_env"`
	TimeoutMS int    `mapstructure:"timeout_ms"`
}

// Load 从默认路径加载配置。
// 读取 CONFIG_PATH 环境变量；为空时尝试默认 config.yaml。
// 默认 config.yaml 不存在时允许，使用默认值+环境变量。
// CONFIG_PATH 显式指定但文件不存在时返回错误。
func Load() (Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return loadFromDefaultPath()
	}
	return LoadFromPath(configPath)
}

// LoadFromPath 从指定路径加载配置文件。
// 文件不存在时返回错误。
func LoadFromPath(path string) (Config, error) {
	return loadWithPath(path, true)
}

// loadFromDefaultPath 尝试加载默认 config.yaml，不存在不报错
func loadFromDefaultPath() (Config, error) {
	return loadWithPath("config.yaml", false)
}

// loadWithPath 核心加载逻辑
// required=true 时，文件必须存在，否则返回错误
// required=false 时，文件不存在不报错
func loadWithPath(path string, required bool) (Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 环境变量配置
	v.SetEnvPrefix("WEREWOLF")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// 设置配置文件
	v.SetConfigFile(path)

	err := v.ReadInConfig()
	if err != nil {
		if isFileNotFoundError(err) {
			if required {
				return Config{}, fmt.Errorf("config file not found: %s: %w", path, err)
			}
			// 默认配置文件不存在，允许，继续使用默认值+环境变量
		} else {
			// 其他错误（如 YAML 格式错误）始终返回
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := Config{
		Server: ServerConfig{
			Addr: v.GetString("server.addr"),
		},
		Storage: StorageConfig{
			StatePath: v.GetString("storage.state_path"),
		},
		AI: readAIConfig(v, "ai"),
	}

	// AI key 解析：WEREWOLF_AI_API_KEY > ai.api_key > ai.api_key_env 指向的环境变量 > 空值
	resolveAIKey(v, &cfg)

	// 校验必填字段
	if cfg.Server.Addr == "" {
		return Config{}, fmt.Errorf("server.addr is required")
	}
	if cfg.Storage.StatePath == "" {
		return Config{}, fmt.Errorf("storage.state_path is required")
	}
	if cfg.AI.Concurrency <= 0 {
		return Config{}, fmt.Errorf("ai.concurrency must be greater than 0")
	}

	return cfg, nil
}

// isFileNotFoundError 判断错误是否为文件不存在
// 仅在明确 not exist 时放行，避免把权限或目录错误误判为可忽略缺失。
func isFileNotFoundError(err error) bool {
	var cfgNotFound viper.ConfigFileNotFoundError
	if errors.As(err, &cfgNotFound) {
		return true
	}
	return errors.Is(err, os.ErrNotExist)
}

// setDefaults 设置所有内置默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("storage.state_path", "data/world_state.json")
	v.SetDefault("ai.provider", "fallback")
	v.SetDefault("ai.base_url", "")
	v.SetDefault("ai.model", "")
	v.SetDefault("ai.concurrency", 1)
	v.SetDefault("ai.api_key", "")
	v.SetDefault("ai.api_key_env", "OPENAI_API_KEY")
	v.SetDefault("ai.timeout_ms", 0)
}

func readAIConfig(v *viper.Viper, prefix string) AIConfig {
	config := AIConfig{
		Provider:    v.GetString(prefix + ".provider"),
		BaseURL:     v.GetString(prefix + ".base_url"),
		Model:       v.GetString(prefix + ".model"),
		Concurrency: v.GetInt(prefix + ".concurrency"),
		APIKeyEnv:   v.GetString(prefix + ".api_key_env"),
		TimeoutMS:   v.GetInt(prefix + ".timeout_ms"),
	}

	fallback := FallbackAIConfig{
		Provider:  v.GetString(prefix + ".fallback.provider"),
		BaseURL:   v.GetString(prefix + ".fallback.base_url"),
		Model:     v.GetString(prefix + ".fallback.model"),
		APIKeyEnv: v.GetString(prefix + ".fallback.api_key_env"),
		TimeoutMS: v.GetInt(prefix + ".fallback.timeout_ms"),
	}
	if hasFallbackConfig(fallback) {
		config.Fallback = &fallback
	}

	return config
}

func hasFallbackConfig(cfg FallbackAIConfig) bool {
	return cfg.Provider != "" || cfg.BaseURL != "" || cfg.Model != "" || cfg.APIKey != "" || cfg.APIKeyEnv != "" || cfg.TimeoutMS != 0
}

// resolveAIKey 按 WEREWOLF_AI_API_KEY > ai.api_key > ai.api_key_env 指向的环境变量 > 空值 顺序解析
func resolveAIKey(v *viper.Viper, cfg *Config) {
	resolveAIConfigKey(v, &cfg.AI, "ai")
	if cfg.AI.Fallback != nil {
		resolveFallbackAIConfigKey(v, cfg.AI.Fallback, "ai.fallback")
	}
}

func resolveAIConfigKey(v *viper.Viper, cfg *AIConfig, keyPath string) {
	cfg.APIKey = resolveAPIKeyValue(v, keyPath, cfg.APIKeyEnv)
}

func resolveFallbackAIConfigKey(v *viper.Viper, cfg *FallbackAIConfig, keyPath string) {
	cfg.APIKey = resolveAPIKeyValue(v, keyPath, cfg.APIKeyEnv)
}

func resolveAPIKeyValue(v *viper.Viper, keyPath string, apiKeyEnv string) string {
	directEnvKey := os.Getenv(envKeyFor(keyPath + ".api_key"))
	if directEnvKey != "" {
		return directEnvKey
	}

	fileKey := v.GetString(keyPath + ".api_key")
	if fileKey != "" {
		return fileKey
	}

	if apiKeyEnv != "" {
		if val, ok := os.LookupEnv(apiKeyEnv); ok {
			return val
		}
	}

	return ""
}

func envKeyFor(keyPath string) string {
	clean := strings.TrimPrefix(filepath.ToSlash(keyPath), "/")
	clean = strings.ReplaceAll(clean, "/", ".")
	clean = strings.ReplaceAll(clean, ".", "_")
	clean = strings.ReplaceAll(clean, "-", "_")
	return "WEREWOLF_" + strings.ToUpper(clean)
}

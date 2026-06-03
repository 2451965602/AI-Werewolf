---
comet_change: configure-env-file-priority
role: technical-design
canonical_spec: openspec
---

## 目标

为服务启动增加统一运行时配置能力。配置解析必须满足 OpenSpec 的优先级要求：环境变量优先，其次配置文件，最后使用默认值。HTTP API 和业务行为不改变。

## 采用方案

使用 `internal/config` 包封装 Viper。外部调用方只接触稳定的 `Config` 结构，不直接依赖 Viper。

核心结构：

```go
type Config struct {
    Server  ServerConfig
    Storage StorageConfig
    AI      AIConfig
}
```

加载流程：

```text
Viper SetDefault
  ↓
ReadInConfig(config.yaml 或 CONFIG_PATH)
  ↓
AutomaticEnv(WEREWOLF_* 覆盖)
  ↓
Unmarshal / GetString 生成 Config
  ↓
main 注入 Store / Router / AI 选择逻辑
```

## Viper 使用方式

- 使用 `viper.New()` 创建局部实例，避免全局状态污染测试。
- 用 `SetDefault` 设置所有默认值。
- `CONFIG_PATH` 指定完整配置文件路径；未设置时默认 `config.yaml`。
- `SetEnvPrefix("WEREWOLF")` 绑定业务环境变量前缀。
- `SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))` 将 `storage.state_path` 映射为 `WEREWOLF_STORAGE_STATE_PATH`。
- `AutomaticEnv()` 让每次取值时读取当前进程环境变量。

有效优先级为：

```text
WEREWOLF_* 环境变量 > 配置文件 > 默认值
```

本方案不使用 Viper 的 `Set()`、flags 或远程配置源，避免引入高于环境变量的额外优先级。

## 配置文件与环境变量

默认配置文件：`config.yaml`。

支持环境变量：

- `CONFIG_PATH`
- `WEREWOLF_SERVER_ADDR`
- `WEREWOLF_STORAGE_STATE_PATH`
- `WEREWOLF_AI_PROVIDER`
- `WEREWOLF_AI_BASE_URL`
- `WEREWOLF_AI_MODEL`
- `WEREWOLF_AI_API_KEY`
- `WEREWOLF_AI_API_KEY_ENV`

AI 密钥支持直接配置和环境变量覆盖。有效解析顺序为：

```text
WEREWOLF_AI_API_KEY > 配置文件 ai.api_key > ai.api_key_env 指向的环境变量 > 空值
```

`ai.api_key_env` 可继续保存密钥环境变量名，例如 `OPENAI_API_KEY`。无论密钥来自哪里，程序都不得把密钥值写入日志。

## 接入点

- `cmd/server/main.go`：启动时调用配置加载；失败则退出并打印错误。
- `internal/transport/http/router.go`：新增可注入地址的构造方式，例如 `NewRouter(service, addr)` 或 option 参数。
- `internal/infrastructure/store/json_store.go`：继续接收路径，无需改动持久化实现。
- AI provider：本次解析 AI 配置和 API key；未配置真实模型或密钥时保持 fallback provider，不引入完整 Eino 模型工厂，避免扩大范围。

## 错误处理

- 默认配置文件不存在：不报错，继续使用默认值和环境变量。
- `CONFIG_PATH` 指定文件不存在：返回错误，避免部署配置拼写错误被静默忽略。
- YAML 格式错误：返回错误。
- `server.addr` 或 `storage.state_path` 解析为空：返回错误。

## 测试策略

- 配置包单元测试：默认值、文件覆盖、环境变量覆盖文件、缺失默认配置文件、`CONFIG_PATH` 指向不存在文件、非法 YAML。
- Router 测试：确认自定义地址不会破坏路由注册和现有 handler 行为。
- 全量验证：`go test ./... -timeout 60s`。

## 取舍与风险

- 引入 Viper 比手写 YAML 解析更重，但可减少配置优先级和环境变量映射的自研逻辑。
- 不做配置文件热重载；HTTP 监听地址和存储路径属于启动期配置，运行中切换需要额外生命周期管理，超出本变更范围。
- Viper 的环境变量不会缓存，但外部环境变量通常不会在进程运行中变化；本变更仍按启动期配置使用。
- 允许配置文件保存 AI 密钥会降低集中密钥管理的安全性；部署时仍建议优先使用环境变量覆盖。

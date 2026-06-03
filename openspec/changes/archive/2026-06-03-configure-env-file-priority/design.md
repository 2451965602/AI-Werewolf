## Overview

引入一个内部配置层作为服务启动时的唯一配置来源。配置层基于 Viper 统一处理默认值、YAML 配置文件和环境变量覆盖，得到不可变的 `Config` 值传入各基础设施组件。

## Architecture Decisions

### 单次加载，不做环境变量热监听

Go 进程无法被动监听外部 shell 环境变量变化；环境变量通常是进程启动时注入的部署契约。因此本变更将“环境变量监听”定义为启动时读取并覆盖配置文件。配置变更通过重启进程生效，避免轮询环境变量造成不可预测行为。

### 使用 Viper 统一解析配置

新增依赖 `github.com/spf13/viper`。配置包内部创建独立 `viper.New()` 实例，不使用全局单例，避免测试和未来多实例场景互相污染。

配置加载顺序由 Viper 保证：默认值、配置文件、环境变量。实现时不使用 `Set()` 或命令行 flags，确保本变更要求的有效优先级为 `环境变量 > 配置文件 > 默认值`。

### 内部配置包集中解析

新增 `internal/config` 包，负责：

- 定义配置结构：`Server`、`Storage`、`AI`。
- 提供默认值。
- 使用 Viper 读取 YAML 配置文件。
- 使用 Viper `SetEnvPrefix("WEREWOLF")`、`SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))` 和 `AutomaticEnv()` 应用环境变量覆盖。
- 验证关键字段不为空。

启动入口只消费解析后的配置，不散落 `os.Getenv` 或 YAML 解析逻辑。

### 配置文件路径

默认尝试读取 `config.yaml`。如果文件不存在，则只使用默认值和环境变量。可通过 `CONFIG_PATH` 指定配置文件路径，方便不同环境挂载配置。

### 环境变量命名

采用服务前缀，避免与通用变量冲突：

- `CONFIG_PATH`
- `WEREWOLF_SERVER_ADDR`
- `WEREWOLF_STORAGE_STATE_PATH`
- `WEREWOLF_AI_PROVIDER`
- `WEREWOLF_AI_BASE_URL`
- `WEREWOLF_AI_MODEL`
- `WEREWOLF_AI_API_KEY`
- `WEREWOLF_AI_API_KEY_ENV`

AI API key 的真实值支持三种来源：`WEREWOLF_AI_API_KEY` 直接覆盖、配置文件 `ai.api_key`、以及 `api_key_env` 指向的外部环境变量。解析后的密钥不得写入日志。

## Data Flow

```text
Viper 默认值
  ↓
可选 config.yaml / CONFIG_PATH 指定文件
  ↓
Viper AutomaticEnv 读取 WEREWOLF_* 环境变量覆盖
  ↓
cmd/server/main.go 注入 Store / Router / AI 初始化
```

## Error Handling

- 配置文件不存在时不报错，继续使用默认值和环境变量。
- 配置文件存在但无法读取或 YAML 格式错误时启动失败，并返回明确错误。
- 必填运行字段为空时启动失败。
- 未配置真实 AI 模型或密钥时保持当前 fallback provider 行为，避免阻塞本地启动。

## Testing Strategy

- 为配置加载包添加单元测试，覆盖默认值、配置文件覆盖、环境变量覆盖、缺失文件和非法 YAML。
- 更新路由测试以覆盖自定义监听地址初始化（如现有测试受影响）。
- 运行 `go test ./... -timeout 60s` 验证全项目。

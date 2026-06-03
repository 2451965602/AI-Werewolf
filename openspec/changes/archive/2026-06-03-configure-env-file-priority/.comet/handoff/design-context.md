# Comet Design Handoff

- Change: configure-env-file-priority
- Phase: design
- Mode: compact
- Context hash: 3ccb10da080eda74474c2c9e9e1a38e38307c97357e980814b4b33bdcf2820ac

Generated-by: comet-handoff.sh

OpenSpec remains the canonical capability spec. This handoff is a deterministic, source-traceable context pack, not an agent-authored summary.

## openspec/changes/configure-env-file-priority/proposal.md

- Source: openspec/changes/configure-env-file-priority/proposal.md
- Lines: 1-29
- SHA256: 4c26a9333b9720acd035f80218eedc5acb996130fd40a405bda3ef667c4a84fc

```md
## Why

当前服务入口将状态文件路径、HTTP 监听地址和 AI 相关配置分散在代码或示例文件中，实际运行时无法通过配置文件或环境变量统一覆盖。部署到不同环境时必须修改代码或依赖默认行为，容易造成配置不可追踪、密钥处理不一致和环境差异问题。

## What Changes

- 新增运行时配置能力，支持从配置文件读取服务、存储和 AI 配置。
- 新增环境变量覆盖规则，统一遵循 `环境变量 > 配置文件 > 默认值`。
- 默认读取仓库现有 `config.example.yaml` 对应的配置结构，并允许缺失配置时使用安全默认值。
- 将现有硬编码状态文件路径和 HTTP 地址迁移到配置解析结果。
- AI 密钥可通过配置文件 `ai.api_key` 保存，也可通过环境变量覆盖；同时支持 `api_key_env` 指定外部密钥环境变量名作为兜底。
- 不引入破坏性 API 变更；HTTP 路由路径和响应结构保持不变。

## Capabilities

### New Capabilities

- `runtime-configuration`: 服务可从配置文件、环境变量和默认值合并得到运行配置，并按确定优先级使用。

### Modified Capabilities

- 无。

## Impact

- 影响服务启动入口、HTTP 服务初始化、存储路径初始化和 AI 配置读取。
- 可能新增内部配置包与对应单元测试。
- 新增 `github.com/spf13/viper` 作为配置解析依赖。
- 运行命令和 HTTP API 保持兼容。
```

## openspec/changes/configure-env-file-priority/design.md

- Source: openspec/changes/configure-env-file-priority/design.md
- Lines: 1-71
- SHA256: 527725c644c34d233e558c37813a1169d4947091f73d682e3e4d0bda80e37405

```md
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
```

## openspec/changes/configure-env-file-priority/tasks.md

- Source: openspec/changes/configure-env-file-priority/tasks.md
- Lines: 1-3
- SHA256: 612e48de8b0dd34cbdc6d6d74e6519996e95eed919a892a46d0da3ac4ba844a2

```md
- [ ] 1. 新增内部配置包，定义默认值、YAML 文件加载、环境变量覆盖和基础校验。
- [ ] 2. 将服务启动入口和 HTTP router 初始化改为使用解析后的配置。
- [ ] 3. 为配置优先级和错误路径添加单元测试，并运行全项目测试。
```

## openspec/changes/configure-env-file-priority/specs/runtime-configuration/spec.md

- Source: openspec/changes/configure-env-file-priority/specs/runtime-configuration/spec.md
- Lines: 1-72
- SHA256: de00bdfae06c3a5b143142e9ac63b7294cabff8b71b7a7744d28514890123db4

```md
## ADDED Requirements

### Requirement: Runtime configuration source priority

The application SHALL resolve runtime configuration using the priority order: environment variables, then configuration file values, then built-in defaults.

#### Scenario: Environment variable overrides configuration file

- **GIVEN** a configuration file defines `storage.state_path` as `data/from-file.json`
- **AND** `WEREWOLF_STORAGE_STATE_PATH` is set to `data/from-env.json`
- **WHEN** the application loads configuration
- **THEN** the resolved storage state path is `data/from-env.json`

#### Scenario: Configuration file overrides default value

- **GIVEN** no storage path environment variable is set
- **AND** a configuration file defines `storage.state_path` as `data/from-file.json`
- **WHEN** the application loads configuration
- **THEN** the resolved storage state path is `data/from-file.json`

#### Scenario: Defaults apply when no external configuration exists

- **GIVEN** no configuration file exists at the default path
- **AND** no related environment variables are set
- **WHEN** the application loads configuration
- **THEN** the resolved server address and storage path use built-in defaults

### Requirement: Configurable HTTP server address

The application SHALL use the resolved server address when constructing the HTTP server.

#### Scenario: Custom server address is configured

- **GIVEN** `WEREWOLF_SERVER_ADDR` is set to `:9090`
- **WHEN** the application starts
- **THEN** the HTTP server is configured to listen on `:9090`

### Requirement: Configurable persistent state path

The application SHALL use the resolved storage state path for JSON state persistence.

#### Scenario: Custom state path is configured

- **GIVEN** `WEREWOLF_STORAGE_STATE_PATH` is set to `tmp/state.json`
- **WHEN** the application creates its JSON store
- **THEN** the store uses `tmp/state.json` as the persistence path

### Requirement: AI secret configuration

The application SHALL allow the AI API key to be provided by direct environment variable, configuration file value, or an environment variable name configured in `ai.api_key_env`.

#### Scenario: API key direct environment variable overrides configuration file

- **GIVEN** the configuration file defines `ai.api_key` as `file-secret`
- **AND** `WEREWOLF_AI_API_KEY` is set to `env-secret`
- **WHEN** the application loads AI configuration
- **THEN** the resolved AI API key is `env-secret`

#### Scenario: API key can be stored in configuration file

- **GIVEN** no direct AI API key environment variable is set
- **AND** the configuration file defines `ai.api_key` as `file-secret`
- **WHEN** the application loads AI configuration
- **THEN** the resolved AI API key is `file-secret`

#### Scenario: API key environment variable name is configured as fallback

- **GIVEN** the configuration file defines `ai.api_key_env` as `OPENAI_API_KEY`
- **AND** `OPENAI_API_KEY` is set in the process environment
- **AND** no direct AI API key value is configured
- **WHEN** the application loads AI configuration
- **THEN** the resolved AI API key is read from `OPENAI_API_KEY`
```


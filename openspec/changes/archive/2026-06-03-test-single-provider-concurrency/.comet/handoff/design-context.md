# Comet Design Handoff

- Change: test-single-provider-concurrency
- Phase: design
- Mode: compact
- Context hash: 1a6e8b096161c27e0d76b1272ddcacf81853f98b8ca4269198b15546e04c87e7

Generated-by: comet-handoff.sh

OpenSpec remains the canonical capability spec. This handoff is a deterministic, source-traceable context pack, not an agent-authored summary.

## openspec/changes/test-single-provider-concurrency/proposal.md

- Source: openspec/changes/test-single-provider-concurrency/proposal.md
- Lines: 1-27
- SHA256: 8538d52acc5cfb2744e6cc11b0dcea2cc452a3912bf20b1a1113404184e1f817

```md
## Why

当前仓库已经引入 AI provider 相关运行时配置，但服务启动入口仍硬编码 `FallbackProvider`，导致配置文件中的 `ai.provider` 无法真正生效。同时仓库尚未支持 `ai.concurrency`，无法对真实 provider 做串行限流。需要把配置、provider 选择和并发控制真正接入运行时，并补齐测试，确保 `ai.concurrency = 1` 时行为可验证。

## What Changes

- 让服务启动入口按 `ai.provider` 配置构造实际 provider，不再硬编码 `FallbackProvider`。
- 新增 `ai.concurrency` 配置，并支持对 provider 调用做全局串行排队控制。
- 采用通用 provider 装饰器实现限流，避免把并发控制写死到具体 provider 内部。
- 为配置解析、provider 选择和串行限流行为补充自动化测试。
- 保持现有 HTTP API 不变，不新增对外接口。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `runtime-configuration`: 扩展 AI 运行时配置，支持 provider 选择与并发限制字段，并要求非法配置显式失败。

## Impact

- 影响 `internal/config`、`cmd/server` 和 `internal/infrastructure/ai` 的初始化路径。
- 需要新增或修改配置测试、provider 选择测试和并发控制测试。
- 不引入新的 HTTP API 或存储结构变更。
```

## openspec/changes/test-single-provider-concurrency/design.md

- Source: openspec/changes/test-single-provider-concurrency/design.md
- Lines: 1-53
- SHA256: 516994419d07f3b601388d47b42087c40e349b60e02496ec1927d16223421190

```md
## Overview

本次变更把 AI 运行时配置真正接入服务启动入口，并补充 provider 级并发控制。实现上通过“底层 provider + 通用限流装饰器”的组合，既让 `ai.provider` 真正决定实际 provider，又让 `ai.concurrency = 1` 时能在进程内全局串行排队执行 AI 调用。

## Architecture Decisions

### Provider 选择由启动入口统一负责

`cmd/server/main.go` 不再直接实例化 `FallbackProvider`。启动入口读取 `cfg.AI.Provider` 后，通过统一构造逻辑选择具体 provider。若 provider 名称无法识别，启动立即失败，避免“配置写了但实际没生效”的静默偏差。

### 并发控制通过装饰器而非嵌入具体 provider

新增一个 provider 装饰器包装真实 provider。装饰器在调用前获取并发令牌，在调用结束后释放。这样并发控制与具体 provider 解耦，后续无论是 `fallback`、`eino` 还是其他 provider，都可复用同一套限流逻辑。

### `ai.concurrency = 1` 语义为全局串行排队

并发限制作用域为当前进程内的该 provider 实例。配置值为 1 时，同一时刻只允许一个 AI 调用执行，后续调用进入等待队列，不直接拒绝。该语义符合“单 provider 串行化”预期，也便于通过并发测试验证。

### 非法配置显式失败

若 `ai.provider` 不是受支持的 provider 名称，或 `ai.concurrency <= 0`，启动阶段直接返回错误。相比静默回退或自动修正，这种策略更利于暴露配置错误并保证测试结论可信。

## Data Flow

```text
config.yaml / 环境变量
  ↓
internal/config.Load()
  ↓
解析 ai.provider / ai.concurrency
  ↓
cmd/server/main.go 构造底层 provider
  ↓
LimitedProvider 包装底层 provider
  ↓
application/service 发起 AI 调用
  ↓
并发令牌控制（concurrency=1 时串行排队）
```

## Error Handling

- `ai.provider` 无法识别时启动失败。
- `ai.concurrency <= 0` 时启动失败。
- 底层 provider 调用错误原样返回，限流装饰器不吞错也不改写业务语义。
- 并发控制必须在调用结束时释放令牌，避免异常路径导致后续请求永久阻塞。

## Testing Strategy

- 为配置加载新增 `ai.concurrency` 正常与非法值测试。
- 为 provider 构造路径新增按配置选择 provider 的测试，以及未知 provider 错误测试。
- 为限流装饰器新增并发测试，验证 `concurrency=1` 时多个调用按串行顺序完成。
- 运行 `go test ./... -timeout 60s` 做全量回归。
```

## openspec/changes/test-single-provider-concurrency/tasks.md

- Source: openspec/changes/test-single-provider-concurrency/tasks.md
- Lines: 1-4
- SHA256: 71582669d715c34e01229274f0ef867745510acfbc3066ebb0bd115e30aeccc0

```md
- [ ] 1. 扩展运行时配置，新增 `ai.concurrency` 读取与校验，并补充对应测试。
- [ ] 2. 重构 AI provider 构造路径，按 `ai.provider` 选择实际 provider，并对非法 provider 显式失败。
- [ ] 3. 新增通用 provider 限流装饰器，实现 `ai.concurrency = 1` 时的全局串行排队。
- [ ] 4. 为 provider 选择与并发串行化补充自动化测试，并运行 `go test ./... -timeout 60s`。
```

## openspec/changes/test-single-provider-concurrency/specs/runtime-configuration/spec.md

- Source: openspec/changes/test-single-provider-concurrency/specs/runtime-configuration/spec.md
- Lines: 1-33
- SHA256: 9604a963f04c7b6778aeff3be11e12f468d0d3cf34005359e9bc71d37baca08f

```md
## MODIFIED Requirements

### Requirement: AI provider runtime selection

The application SHALL construct the AI provider from resolved runtime configuration instead of hardcoding a fallback provider.

#### Scenario: Configured provider is selected at startup

- **GIVEN** runtime configuration resolves `ai.provider` to a supported provider name
- **WHEN** the application starts
- **THEN** the service uses the configured provider implementation for AI calls

#### Scenario: Unknown provider fails fast

- **GIVEN** runtime configuration resolves `ai.provider` to an unsupported provider name
- **WHEN** the application starts
- **THEN** startup fails with a clear configuration error

### Requirement: AI concurrency configuration

The application SHALL support `ai.concurrency` as the maximum number of concurrent calls allowed for the configured AI provider instance.

#### Scenario: Concurrency one serializes provider calls

- **GIVEN** runtime configuration resolves `ai.concurrency` to `1`
- **WHEN** multiple AI calls are made concurrently through the same provider instance
- **THEN** the calls execute one at a time in serial order

#### Scenario: Invalid concurrency fails fast

- **GIVEN** runtime configuration resolves `ai.concurrency` to `0` or a negative number
- **WHEN** the application starts
- **THEN** startup fails with a clear configuration error
```


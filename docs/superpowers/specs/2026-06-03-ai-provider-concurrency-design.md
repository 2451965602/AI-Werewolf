---
comet_change: test-single-provider-concurrency
role: technical-design
canonical_spec: openspec
archived-with: 2026-06-03-test-single-provider-concurrency
status: final
---

## Overview

本设计让 AI 运行时配置真正控制服务行为，并为 provider 调用补充可验证的并发限制。目标是消除 `cmd/server/main.go` 中对 `FallbackProvider` 的硬编码，让 `ai.provider` 决定实际 provider，同时新增 `ai.concurrency`，使同一 provider 实例在 `concurrency=1` 时全局串行排队执行 AI 调用。

## Goals

- 启动时按配置选择真实 AI provider。
- 支持 `ai.concurrency`，并在值为 1 时实现全局串行排队。
- 对非法 provider 和非法并发值显式失败。
- 通过自动化测试验证配置解析、provider 选择和串行化行为。

## Non-Goals

- 不新增 HTTP API。
- 不引入跨进程分布式限流。
- 不改变业务层调用 AI 的接口语义。

## Architecture

### 1. 配置层扩展

在 `internal/config.AIConfig` 中新增 `Concurrency int` 字段，对应 `ai.concurrency`。

配置规则：

- 继续沿用 `环境变量 > 配置文件 > 默认值`
- 默认值建议设为 `1`
- 解析完成后执行校验：`concurrency <= 0` 直接返回错误

对应环境变量命名延续现有规则，例如 `WEREWOLF_AI_CONCURRENCY`。

### 2. Provider 构造路径集中化

服务启动入口不再直接写死 provider 类型，而是通过统一构造逻辑完成：

```text
cfg.AI.Provider
  ↓
buildProvider(cfg.AI)
  ↓
concrete provider
  ↓
wrap with LimitedProvider
```

这样可以把 provider 选择、配置校验和装饰器包装集中在一处，避免未来在多个入口复制逻辑。

若 provider 名称未知，则返回明确错误并终止启动。

### 3. 通用限流装饰器

在 `internal/infrastructure/ai` 新增通用 provider 装饰器，例如：

- `LimitedProvider`

职责：

- 包装任意底层 provider
- 在每次 AI 调用前获取令牌
- 调用结束后释放令牌
- 不改变底层返回值和错误

实现建议使用容量为 `concurrency` 的缓冲 channel，或等价的信号量机制。对于 `concurrency=1`，天然得到串行排队效果。

### 4. 错误处理边界

- provider 名称非法：启动失败
- concurrency 非法：启动失败
- provider 运行错误：原样上抛
- 装饰器必须用可靠的释放路径保证不会死锁

## Data Flow

```text
config.yaml / env
  ↓
config.Load()
  ↓
resolve provider + concurrency
  ↓
build concrete provider
  ↓
wrap with LimitedProvider
  ↓
service / engine issues AI calls
  ↓
token acquisition and serialized execution
```

## Testing Strategy

### 配置测试

- `ai.concurrency` 从配置文件和环境变量正确读取
- `ai.concurrency <= 0` 返回错误

### Provider 选择测试

- `provider=fallback` 时创建 fallback provider
- 受支持的真实 provider 名称时创建对应 provider
- 未知 provider 返回错误

### 并发测试

使用 mock provider：

- 在调用开始和结束时记录顺序
- 并发发起多个调用
- 验证 `concurrency=1` 时不存在重叠执行区间

### 回归测试

- `go test ./... -timeout 60s`

## Risks and Mitigations

### 风险：限流器释放不完整导致阻塞

缓解：在调用结束路径统一释放令牌，测试覆盖错误返回路径。

### 风险：provider 构造逻辑继续散落在入口代码

缓解：将 provider 选择收敛到单一构造函数，避免未来再出现“配置存在但未生效”的问题。

## Acceptance Mapping

- OpenSpec `runtime-configuration` delta spec 定义 provider 选择与并发限制的行为契约
- 本设计文档定义具体实现方式：集中构造 + 通用限流装饰器 + fail-fast 校验

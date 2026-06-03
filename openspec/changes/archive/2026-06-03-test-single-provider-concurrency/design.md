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

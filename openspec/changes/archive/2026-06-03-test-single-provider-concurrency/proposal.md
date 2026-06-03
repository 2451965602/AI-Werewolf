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

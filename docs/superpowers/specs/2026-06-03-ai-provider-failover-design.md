---
comet_change: fix-ai-provider-next-hang
role: technical-design
canonical_spec: openspec
---

## Context

当前后端 `POST /api/game/next` 在首轮白天阶段会串行触发 10 次 AI 发言生成。实测外部 AI 平台在真实提示词下单次调用约为 6 秒，因此单次阶段推进被放大到 60 秒级，前端联调时稳定超时。与此同时，现有 Eino provider 使用 `context.Background()` 且没有 HTTP 超时，导致外部依赖一旦变慢，整个阶段推进会被无限放大。

## Goals

- 让 `POST /api/game/next` 在 AI 平台变慢时仍能在可接受时间内返回。
- 在主 AI 提供商超时、空响应或解析失败时，自动回退到第二个提供商。
- 保持现有 HTTP API 与游戏规则接口不变。

## Non-Goals

- 不把 AI provider 体系升级为任意长度的 provider chain。
- 不调整狼人杀游戏规则本身。
- 不修改前端协议。

## Design Summary

采用“主提供商 + 单一 fallback 提供商”的最小扩展方案：

1. 保留当前 `ai` 作为主提供商配置。
2. 新增 `ai.timeout_ms`，为主提供商设置请求超时。
3. 新增 `ai.fallback` 配置块，允许指定第二个提供商及其独立超时。
4. 新增 `FailoverProvider`，主提供商失败时自动切换到 fallback 提供商。
5. 把 Eino provider 的 `Generate` 调用改为带超时的 context。
6. 将“空内容”视为失败条件，触发 fallback。
7. 将服务层 `NextPhase` 的整阶段 AI 重试由 `2` 次降为 `1` 次，因为 provider 内部已承担 failover 责任。

## Configuration Changes

扩展当前 `AIConfig`，增加：

- `TimeoutMS int`
- `Fallback *AIConfig` 或等价的 `FallbackAIConfig`

推荐配置形态：

```yaml
ai:
  provider: eino
  base_url: http://primary.example/v1
  model: qwen3.6-plus
  api_key_env: OPENAI_API_KEY
  timeout_ms: 5000
  fallback:
    provider: fallback
    timeout_ms: 1000
```

这里故意只支持一个 fallback provider，而不是 `providers[]` 通用链。原因是本次问题只需要“一主一备”，不应为未来可能的多级回退增加无必要复杂度。

## Provider Layer Changes

### EinoProvider

- `generate()` 不再直接使用 `context.Background()`。
- 改为使用 `context.WithTimeout()`，超时值来自配置。
- 若 `message.Content == ""`，返回显式错误，避免把空内容视为成功响应。

### FailoverProvider

新增一个包装型 provider：

- 所有 `Speak / VoteTarget / WerewolfTarget / SeerTarget / WitchAction` 调用都先走主 provider。
- 命中以下任一条件时转用 fallback provider：
  - request timeout
  - provider error
  - 空响应
  - 目标编号解析失败

这样可以把“外部平台抖动”限制在单次 provider 决策层，而不是拖死整个 `AdvancePhase`。

## Service Layer Changes

当前 `NextPhase()` 会在 provider 失败时重试整轮 `AdvancePhase` 两次，然后再 fallback 到 `nil` provider。这个策略在外部依赖很慢时会放大总体时长。

设计调整：

- 保留“失败后可退化到 `nil` provider”的机制。
- 将整阶段 AI 重试从 `2` 次降为 `1` 次。

原因：
- provider 内部已经有主备切换，再做整阶段重试收益很低。
- 对首轮白天的 10 次 AI 发言路径，额外重试只会线性扩大总耗时。

## Error Semantics

超时和空响应属于“可退化错误”，优先触发 fallback provider。

只有当：
- 主 provider 失败
- fallback provider 失败
- 或后续 fallback 到 `nil` provider 也失败

才把错误继续向上抛给 HTTP 层。

## Risks

### 1. 配置复杂度上升

新增 fallback 配置后，运行时配置更复杂。

缓解：只支持一个 fallback provider，避免通用链式配置。

### 2. 主 provider 结果不再“无限等待”

超时后会直接切到 fallback，这意味着可能拿不到主 provider 稍后才会返回的内容。

缓解：本次优先保障 `next` 的可返回性，而不是追求单次最优文案质量。

### 3. 回退 provider 行为更保守

若 fallback 设为 `fallback` 本地策略，返回内容会更朴素。

缓解：这比 `next` 长时间挂起更符合当前系统目标。

## Testing Strategy

- `config` 单测：验证主 provider / fallback provider / timeout 字段解析。
- `ai` 单测：验证主 provider 超时、空内容、解析失败时 fallback 被调用。
- `service` 单测：验证 `NextPhase` 在主 provider 很慢时仍能返回，且不再做多轮放大重试。
- 集成验证：
  - `POST /api/game/start`
  - `POST /api/game/next`
  - `GET /api/game/state`
  - `GET /api/game/messages`

## Spec Impact

本次实现会修改两类 OpenSpec：

- `runtime-configuration`：新增 timeout 与 fallback provider 配置要求
- `werewolf-ai-decision`：新增主 provider 失败后的自动降级行为要求

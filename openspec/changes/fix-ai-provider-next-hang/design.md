## Fix Scope

这是一次后端 hotfix，不新增能力，不改 API 契约，只修复 `POST /api/game/next` 在 AI 决策路径上的长时间挂起问题。

## Root Cause Hypothesis

当前 `next` 的首个白天阶段会串行触发 10 次 `Speak` 调用；若外部 AI 平台响应较慢或单次调用异常，整个推进请求会被放大到不可接受的时长。与此同时，服务层还会重试 `AdvancePhase`，进一步扩大等待时间。

## Fix Direction

- 先验证 AI 平台在真实提示词下能否稳定返回文本。
- 为 AI 调用链路增加明确的超时/退化控制，避免整个阶段推进被单次外部依赖拖死。
- 如单次 AI 调用失败或超时，优先退化到已有 fallback 行为，而不是继续无限等待。

## Constraints

- 不修改现有 HTTP 接口。
- 不改变狼人杀规则引擎的业务语义。
- 只允许在 AI 调用与服务退化层做最小修复。

## Why

当前后端 `POST /api/game/next` 在本地联调中长时间不返回，直接阻塞了前端“下一阶段”演示路径。已知 AI 平台可达，但推进阶段依赖的 AI 调用链路存在高延迟或失败放大，需要先定位根因，再修复为可预期的返回行为。

## What Changes

- 验证当前 AI 平台在后端真实提示词下是否正常返回。
- 复现并定位 `POST /api/game/next` 长时间挂起的具体原因。
- 修复后端推进阶段中的 AI 调用问题，确保 `next` 在可接受时间内返回，必要时回退到可用策略。

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `werewolf-ai-decision`: 修正 AI 决策调用在阶段推进中的超时/退化行为。

## Impact

- Affected code: `internal/infrastructure/ai/**`、`internal/application/service.go`、可能涉及 `internal/domain/engine.go`
- Affected runtime behavior: `POST /api/game/next` 的响应时长和失败退化策略
- Affected external dependency: 当前配置的 AI 平台 `http://100.122.143.56:5002/v1`

# Comet Design Handoff

- Change: fix-ai-provider-next-hang
- Phase: design
- Mode: compact
- Context hash: 904d7b8c05d665e215f8175474bc5098d8e8d095b5d7dae921fd8c4db81161af

Generated-by: comet-handoff.sh

OpenSpec remains the canonical capability spec. This handoff is a deterministic, source-traceable context pack, not an agent-authored summary.

## openspec/changes/fix-ai-provider-next-hang/proposal.md

- Source: openspec/changes/fix-ai-provider-next-hang/proposal.md
- Lines: 1-23
- SHA256: d2cf30d552bf7f945a3ca60c570b9358a3910e75a21010347effc85a33af3325

```md
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
```

## openspec/changes/fix-ai-provider-next-hang/design.md

- Source: openspec/changes/fix-ai-provider-next-hang/design.md
- Lines: 1-19
- SHA256: 720b8dce52e9a566627f093adeea0b67fb46eac42b4be656a6c26eb2f29e7a11

```md
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
```

## openspec/changes/fix-ai-provider-next-hang/tasks.md

- Source: openspec/changes/fix-ai-provider-next-hang/tasks.md
- Lines: 1-4
- SHA256: e738faf000f1c547f91a4aa3e4446868b4ddfd2c25025d3983ce9d2a495561b9

```md
- [ ] 验证 AI 平台在真实提示词和当前模型配置下是否正常返回内容。
- [ ] 复现 `POST /api/game/next` 的长时间挂起，并记录耗时和调用路径证据。
- [ ] 在 AI provider / service 层实现最小修复，避免单次外部调用把整次阶段推进拖死。
- [ ] 验证 `start`、`state`、`messages`、`next` 的核心链路，确认 hotfix 生效。
```

## openspec/changes/fix-ai-provider-next-hang/specs/runtime-configuration/spec.md

- Source: openspec/changes/fix-ai-provider-next-hang/specs/runtime-configuration/spec.md
- Lines: 1-13
- SHA256: 2cb7534b36b94c78c2424cace777bf5386c5b2045f868704d4fe90cfd3382b65

```md
## MODIFIED Requirements

### Requirement: AI runtime configuration
The system MUST support configuring AI provider runtime behavior, including request timeout and a single fallback provider.

#### Scenario: Primary provider timeout configuration
- **WHEN** the runtime loads AI configuration
- **THEN** it MUST accept a primary provider timeout setting
- **AND** the AI provider MUST apply that timeout to outbound model requests

#### Scenario: Fallback provider configuration
- **WHEN** the runtime loads AI configuration
- **THEN** it MUST allow configuring one fallback AI provider with its own provider type, endpoint, model, credentials, and timeout
```

## openspec/changes/fix-ai-provider-next-hang/specs/werewolf-ai-decision/spec.md

- Source: openspec/changes/fix-ai-provider-next-hang/specs/werewolf-ai-decision/spec.md
- Lines: 1-17
- SHA256: f6e84c2d9fe478bf51ec72d37cc2f09045fb068a1ee3f27f2cf9b0b15335276a

```md
## MODIFIED Requirements

### Requirement: AI decision generation
The system MUST generate AI decisions without allowing a single slow or malformed provider response to indefinitely block game progression.

#### Scenario: Primary provider times out
- **WHEN** the primary AI provider exceeds its configured timeout during decision generation
- **THEN** the system MUST attempt the same decision through the configured fallback provider

#### Scenario: Primary provider returns unusable content
- **WHEN** the primary AI provider returns empty content or a response that cannot be parsed into the required decision
- **THEN** the system MUST treat that response as a provider failure
- **AND** it MUST attempt the configured fallback provider before giving up

#### Scenario: All configured providers fail
- **WHEN** both the primary and fallback AI providers fail for a decision
- **THEN** the system MUST fall back to the existing non-AI deterministic decision path instead of blocking indefinitely
```


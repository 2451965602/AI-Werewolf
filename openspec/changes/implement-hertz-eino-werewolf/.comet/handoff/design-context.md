# Comet Design Handoff

- Change: implement-hertz-eino-werewolf
- Phase: design
- Mode: compact
- Context hash: 34add18564d36334bba9c0a6d9fa74db0dc8671e96684939f1ebd11b340b48b6

Generated-by: comet-handoff.sh

OpenSpec remains the canonical capability spec. This handoff is a deterministic, source-traceable context pack, not an agent-authored summary.

## openspec/changes/implement-hertz-eino-werewolf/proposal.md

- Source: openspec/changes/implement-hertz-eino-werewolf/proposal.md
- Lines: 1-60
- SHA256: c9dea9a1e48d0d740aa434c692890f24e72dc55d8934d380e54d006ec04d5641

```md
## Why

当前参考实现 `AI-Werewolf` 以 `SpringBoot + Vue` 为主，尚未提供 `Golang + Hertz + Eino` 的等价后端实现。为了在 Go 技术栈中复用同类能力（回合引擎、AI 决策、状态持久化、可观测 API），需要建立一个可运行且可扩展的 Go 版本实现基线。

## 问题背景

- 现有参考项目的核心玩法与规则约束（首日时序、角色私有信息、夜晚技能执行）已经较完整，但与目标技术栈不一致。
- 目标环境希望统一采用 `Golang + Hertz` 提供 API，并使用 `Eino` 构建 AI 调用链路与决策执行流。
- 当前仓库尚无该能力对应的 OpenSpec 变更定义，缺少可追踪的设计与任务拆分。

## 目标

- 提供一个基于 `Hertz` 的狼人杀后端服务骨架，覆盖游戏启动、推进阶段、查询状态、查询消息等核心接口。
- 基于 `Eino` 建立 AI 决策流程（发言、投票、狼人刀人、预言家查验、女巫用药）的可替换执行链。
- 明确并实现与参考项目一致的关键规则约束：
  - 第 1 天白天仅自我介绍与讨论，不进行放逐投票
  - 第 1 夜开始神职与狼人行动
  - 狼人清零后立即结算
- 输出可持久化的游戏状态结构（JSON 文件）与可测试的核心服务边界。

## 范围

- `Golang` 后端：领域模型、游戏引擎、AI 决策入口、REST API。
- `Hertz`：路由与 HTTP 接口层。
- `Eino`：AI 交互与决策链封装。
- 本地 JSON 持久化：游戏状态与基础配置（角色/人物设定）。

## 非目标

- 不包含完整前端重构（如 Vue/React 页面开发）。
- 不包含生产级分布式部署、水平扩缩容与多房间高并发能力。
- 不在本阶段实现复杂反作弊、观战系统、账号体系。

## What Changes

- 新增基于 `Hertz` 的 API 服务入口与路由层。
- 新增狼人杀核心领域模型与游戏状态机实现。
- 新增 `Eino` 驱动的 AI 决策能力抽象及默认实现。
- 新增游戏状态持久化与恢复能力（JSON）。
- 新增规则约束与最小测试覆盖（核心流程与关键边界条件）。

## Capabilities

### New Capabilities

- `werewolf-game-engine`: 回合推进、昼夜阶段、胜负判定、角色行动编排。
- `werewolf-ai-decision`: 基于 Eino 的发言与行动决策链。
- `werewolf-hertz-api`: 提供游戏生命周期与状态查询 REST API。
- `werewolf-state-storage`: 游戏状态 JSON 存储与加载。

### Modified Capabilities

- 无（当前仓库未发现可复用的同名现有 capability）。

## Impact

- 代码影响：新增 Go 服务代码目录、配置、模型、服务、路由与测试。
- 依赖影响：引入 `Hertz` 与 `Eino` 相关依赖。
- 接口影响：新增 `/api/game/*` 风格端点，与参考项目保持语义一致。
- 运行影响：需要配置可用 LLM API（兼容 OpenAI Chat Completions）以启用 AI 决策。
```

## openspec/changes/implement-hertz-eino-werewolf/design.md

- Source: openspec/changes/implement-hertz-eino-werewolf/design.md
- Lines: 1-94
- SHA256: f4f4847b7cdd3820abe532aa515ab7f5c9cab3423cb25d16c1f4333ca3ee1ebd

[TRUNCATED]

```md
## 设计目标

在保持参考项目核心玩法与规则约束的前提下，使用 `Golang + Hertz + Eino` 搭建清晰分层、可替换 AI 实现、可测试的后端架构。

## 高层架构决策

1. 分层架构
- `transport`（Hertz 路由/Handler）
- `application`（用例编排）
- `domain`（游戏规则与状态机）
- `infrastructure`（Eino AI 客户端、JSON 存储）

2. 规则优先、AI 次之
- 游戏规则由 `domain` 强约束，AI 只负责“在规则约束下给出候选决策”。
- AI 输出必须经过 domain 校验；不合法则走启发式回退，而非直接随机破坏规则。

3. 明确接口边界
- `AIDecisionProvider`：发言/投票/夜间行动接口
- `StateRepository`：保存与加载 `GameState`
- `GameEngine`：统一推进阶段并发出事件

## 方案选型

### 为什么是 Hertz

- 轻量、Go 原生生态友好、路由/中间件模型清晰。
- 便于快速构建 REST API，并保留后续接入 SSE/WebSocket 的扩展位。

### 为什么是 Eino

- 可以将 prompt 组装、模型调用、解析和重试逻辑纳入同一链路抽象。
- 便于后续替换不同 LLM 提供商而不改动 domain。

### 存储为何选 JSON（阶段 1）

- 与参考实现一致，快速落地、可读性高。
- 作为最小可运行基线，后续可在不改 domain 的前提下替换为 DB。

## 关键数据流

```text
POST /api/game/start
  -> GameAppService.StartGame
  -> GameEngine.Initialize
  -> StateRepository.Save
  -> 返回初始 GameState

POST /api/game/next
  -> GameAppService.NextPhase
  -> GameEngine.AdvancePhase
      -> 夜晚: 狼人/预言家/女巫决策链（Eino）
      -> 白天: 公告 -> 讨论 -> (非首日) 投票
      -> 胜负判定
  -> StateRepository.Save
  -> 返回更新后 GameState

GET /api/game/state
  -> StateRepository.Load (或内存快照)
  -> 返回当前状态
```

## 核心规则实现口径

- 第 1 天白天：仅自我介绍+自由讨论，不执行放逐投票。
- 第 1 夜起：狼人刀人、预言家查验、女巫用药按顺序执行。
- 角色信息隔离：
  - 狼人可见狼队信息
  - 预言家仅可见自己的查验结果
  - 女巫仅可见可用药状态与昨夜死亡信息
- 终局判断优先：若存活狼人为 0，立即结算并终止后续阶段动作。

## 可靠性与错误处理

- AI 调用失败：
  - 第一步重试（有限次数）
  - 第二步启发式回退（基于规则和最小策略）
- 状态落盘失败：返回错误并保留内存态，避免 silent failure。
- 非法状态迁移：拒绝推进并返回可诊断错误信息。

## 可测试性设计
```

Full source: openspec/changes/implement-hertz-eino-werewolf/design.md

## openspec/changes/implement-hertz-eino-werewolf/tasks.md

- Source: openspec/changes/implement-hertz-eino-werewolf/tasks.md
- Lines: 1-36
- SHA256: d354f47547bce19dee402b60f32f6795cc8aa52646ee1decf3f5e071beec831c

```md
- [ ] 1. 初始化 Go 服务与基础目录结构
  - [ ] 1.1 建立 `cmd`、`internal/domain`、`internal/application`、`internal/infrastructure`、`internal/transport` 目录
  - [ ] 1.2 接入 Hertz 并完成基础启动配置

- [ ] 2. 建模核心领域对象
  - [ ] 2.1 定义 `Player`、`GameState`、`Message`、`Vote` 等核心结构
  - [ ] 2.2 定义阶段、角色、胜负判定相关枚举与规则常量

- [ ] 3. 实现游戏引擎（domain）
  - [ ] 3.1 实现初始化流程（10 人局默认配置与角色分配）
  - [ ] 3.2 实现昼夜阶段推进与行动编排
  - [ ] 3.3 实现终局判定与规则守卫（含首日规则）

- [ ] 4. 实现 AI 决策抽象与 Eino 适配
  - [ ] 4.1 设计 `AIDecisionProvider` 接口
  - [ ] 4.2 基于 Eino 实现发言、投票、狼人/预言家/女巫决策链
  - [ ] 4.3 增加 AI 失败重试与启发式回退

- [ ] 5. 实现状态存储与恢复
  - [ ] 5.1 设计 `StateRepository` 接口
  - [ ] 5.2 实现 JSON 持久化（save/load）
  - [ ] 5.3 处理状态文件不存在、损坏、写入失败等异常路径

- [ ] 6. 实现 Hertz API 层
  - [ ] 6.1 提供 `POST /api/game/start`、`POST /api/game/next`
  - [ ] 6.2 提供 `GET /api/game/state`、`GET /api/game/messages`、`GET /api/game/health`
  - [ ] 6.3 统一错误响应格式与基础日志

- [ ] 7. 编写测试与最小验收
  - [ ] 7.1 编写 domain 规则测试（首日规则、技能约束、终局判定）
  - [ ] 7.2 编写 application/transport 关键流程测试
  - [ ] 7.3 执行 `go test ./...` 并修复阻塞问题

- [ ] 8. 配置与文档
  - [ ] 8.1 补充运行配置示例（LLM API、端口、存储路径）
  - [ ] 8.2 补充 README 的启动与接口说明
```

## openspec/changes/implement-hertz-eino-werewolf/specs/werewolf-ai-decision/spec.md

- Source: openspec/changes/implement-hertz-eino-werewolf/specs/werewolf-ai-decision/spec.md
- Lines: 1-22
- SHA256: 8023dc7489000ad47722196a9e1b5e9e2f9c86fa05660726929836b931a7918d

```md
## ADDED Requirements

### Requirement: Rule-constrained AI decisions
The system MUST validate all AI-generated decisions against domain rules before applying them to game state.

#### Scenario: Invalid target is rejected
- **WHEN** AI returns an action with an invalid target (dead player, self-forbidden target, or role-forbidden target)
- **THEN** the action MUST be rejected and MUST not mutate game state

### Requirement: Decision fallback on model failure
The system MUST execute bounded retry for AI invocation failures and MUST use deterministic heuristic fallback when retries are exhausted.

#### Scenario: Retry then fallback
- **WHEN** model invocation fails for all configured retries
- **THEN** the system MUST apply a heuristic fallback decision valid under current game rules

### Requirement: Role-scoped private context
The system MUST provide role-scoped private context to AI decisions and MUST prevent leakage of non-permitted private information.

#### Scenario: Seer sees only seer-private knowledge
- **WHEN** generating a seer decision prompt
- **THEN** prompt context MUST include seer inspection history and MUST exclude werewolf team-private data
```

## openspec/changes/implement-hertz-eino-werewolf/specs/werewolf-game-engine/spec.md

- Source: openspec/changes/implement-hertz-eino-werewolf/specs/werewolf-game-engine/spec.md
- Lines: 1-26
- SHA256: 21262f8d676b850e5c5a43934b3443f7061b10f7816d03227c9598caa9de8138

```md
## ADDED Requirements

### Requirement: Deterministic phase progression
The system MUST progress game phases through a deterministic state machine and MUST enforce the sequence of initialization, day, and night transitions.

#### Scenario: Start game enters day one
- **WHEN** a new game is started
- **THEN** the game phase MUST be `day` and round MUST be `1`

#### Scenario: Next phase transitions by current phase
- **WHEN** the current phase is `day` and the game is not ended
- **THEN** the next phase MUST become `night` in the same round or the next round according to the engine rule set

### Requirement: First-day voting constraint
The system MUST forbid exile voting on day one and MUST only allow self-introduction and free discussion on the first day phase.

#### Scenario: Day one does not create exile result
- **WHEN** the game engine executes day phase for round `1`
- **THEN** no exile vote result MUST be produced

### Requirement: Immediate win condition check
The system MUST evaluate win conditions after each actionable step and MUST terminate remaining actions if a terminal condition is met.

#### Scenario: Wolves eliminated causes immediate settlement
- **WHEN** alive werewolf count becomes `0`
- **THEN** the game MUST be marked ended immediately and remaining phase actions MUST be skipped
```

## openspec/changes/implement-hertz-eino-werewolf/specs/werewolf-hertz-api/spec.md

- Source: openspec/changes/implement-hertz-eino-werewolf/specs/werewolf-hertz-api/spec.md
- Lines: 1-30
- SHA256: 3b7c402be607b08490a302ae7f5e6c3250c4129259f1fc56a9bd79500173ed03

```md
## ADDED Requirements

### Requirement: Game lifecycle API
The system MUST expose Hertz HTTP endpoints for starting a game, advancing the game phase, and reading game state.

#### Scenario: Start game endpoint
- **WHEN** a client sends `POST /api/game/start`
- **THEN** the system MUST initialize a new game and return the current game state

#### Scenario: Advance phase endpoint
- **WHEN** a client sends `POST /api/game/next` for a non-ended game
- **THEN** the system MUST advance the game through the domain engine and return the updated game state

### Requirement: Message and health API
The system MUST expose endpoints for reading game messages and checking service health.

#### Scenario: Read messages endpoint
- **WHEN** a client sends `GET /api/game/messages`
- **THEN** the system MUST return the ordered message history for the current game

#### Scenario: Health endpoint
- **WHEN** a client sends `GET /api/game/health`
- **THEN** the system MUST return a successful health response without mutating game state

### Requirement: Consistent error responses
The system MUST return consistent error response structures for invalid requests, invalid state transitions, and infrastructure failures.

#### Scenario: Invalid transition response
- **WHEN** a client requests a phase transition that violates domain rules
- **THEN** the API MUST return a non-success status and a diagnostic error body
```

## openspec/changes/implement-hertz-eino-werewolf/specs/werewolf-state-storage/spec.md

- Source: openspec/changes/implement-hertz-eino-werewolf/specs/werewolf-state-storage/spec.md
- Lines: 1-30
- SHA256: 1eb741c491884afa3bdffc26f15f2a89ce325ca4af69a9b91cfa13accdff8023

```md
## ADDED Requirements

### Requirement: JSON game state persistence
The system MUST persist the current game state to JSON after initialization and after each successful phase transition.

#### Scenario: Save after game start
- **WHEN** a new game is started successfully
- **THEN** the initialized game state MUST be saved to the configured JSON storage path

#### Scenario: Save after phase advance
- **WHEN** a phase transition completes successfully
- **THEN** the updated game state MUST be saved to the configured JSON storage path

### Requirement: Game state loading
The system MUST load an existing game state from JSON storage when requested and MUST report missing or invalid files explicitly.

#### Scenario: Load existing state
- **WHEN** persisted state exists and is valid JSON
- **THEN** the system MUST restore the game state from storage

#### Scenario: Invalid state file
- **WHEN** persisted state exists but cannot be decoded
- **THEN** the system MUST return a storage error and MUST not replace current in-memory state with partial data

### Requirement: Atomic state writes
The system MUST avoid corrupting the existing state file when a write fails.

#### Scenario: Failed write preserves previous state
- **WHEN** writing the next state fails
- **THEN** the previous persisted state MUST remain readable
```


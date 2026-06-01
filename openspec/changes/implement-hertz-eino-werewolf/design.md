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

- `domain` 层纯逻辑优先，避免直接依赖 HTTP/外部 AI。
- 通过 `AIDecisionProvider` mock 覆盖关键规则与边界测试：
  - 首日不投票
  - 女巫药水约束
  - 狼人清零立即结束

## 风险与缓解

- 风险：AI 输出不稳定导致流程抖动
- 缓解：结构化输出约束 + 解析校验 + 启发式回退

- 风险：状态文件并发写入冲突
- 缓解：阶段 1 采用单实例串行推进；后续再引入锁或事务化存储

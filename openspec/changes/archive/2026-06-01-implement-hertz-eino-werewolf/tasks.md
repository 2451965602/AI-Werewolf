- [x] 1. 初始化 Go 服务与基础目录结构
  - [x] 1.1 建立 `cmd`、`internal/domain`、`internal/application`、`internal/infrastructure`、`internal/transport` 目录
  - [x] 1.2 接入 Hertz 并完成基础启动配置

- [x] 2. 建模核心领域对象
  - [x] 2.1 定义 `Player`、`GameState`、`Message`、`Vote` 等核心结构
  - [x] 2.2 定义阶段、角色、胜负判定相关枚举与规则常量

- [x] 3. 实现游戏引擎（domain）
  - [x] 3.1 实现初始化流程（10 人局默认配置与角色分配）
  - [x] 3.2 实现昼夜阶段推进与行动编排
  - [x] 3.3 实现终局判定与规则守卫（含首日规则）

- [x] 4. 实现 AI 决策抽象与 Eino 适配
  - [x] 4.1 设计 `AIDecisionProvider` 接口
  - [x] 4.2 基于 Eino 实现发言、投票、狼人/预言家/女巫决策链
  - [x] 4.3 增加 AI 失败重试与启发式回退

- [x] 5. 实现状态存储与恢复
  - [x] 5.1 设计 `StateRepository` 接口
  - [x] 5.2 实现 JSON 持久化（save/load）
  - [x] 5.3 处理状态文件不存在、损坏、写入失败等异常路径

- [x] 6. 实现 Hertz API 层
  - [x] 6.1 提供 `POST /api/game/start`、`POST /api/game/next`
  - [x] 6.2 提供 `GET /api/game/state`、`GET /api/game/messages`、`GET /api/game/health`
  - [x] 6.3 统一错误响应格式与基础日志

- [x] 7. 编写测试与最小验收
  - [x] 7.1 编写 domain 规则测试（首日规则、技能约束、终局判定）
  - [x] 7.2 编写 application/transport 关键流程测试
  - [x] 7.3 执行 `go test ./...` 并修复阻塞问题

- [x] 8. 配置与文档
  - [x] 8.1 补充运行配置示例（LLM API、端口、存储路径）
  - [x] 8.2 补充 README 的启动与接口说明

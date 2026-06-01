# AI Werewolf Go Backend

基于 `Golang + Hertz + Eino` 的 AI 狼人杀后端基线实现。

## 当前能力

- 10 人局默认游戏初始化
- 昼夜阶段推进
- 第 1 天白天不投票
- 狼人清零立即结算
- AI 决策边界与 Eino `BaseChatModel` 适配器
- JSON 状态保存与加载
- Hertz REST API

## 本地运行

```bash
go mod tidy
go test ./...
go run ./cmd/server
```

默认状态文件路径：`data/world_state.json`。

## API

### 健康检查

```bash
curl http://localhost:8080/api/game/health
```

### 开始新游戏

```bash
curl -X POST http://localhost:8080/api/game/start
```

### 推进阶段

```bash
curl -X POST http://localhost:8080/api/game/next
```

### 查询状态

```bash
curl http://localhost:8080/api/game/state
```

### 查询消息

```bash
curl http://localhost:8080/api/game/messages
```

## 配置示例

见 `config.example.yaml`。

当前 `cmd/server` 使用 fallback AI provider 以保证无 API key 时可运行；生产接入时可通过 Eino provider 注入真实 `model.BaseChatModel`。

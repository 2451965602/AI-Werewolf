# AI Provider Failover Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 `POST /api/game/next` 因主 AI 提供商慢响应导致的长时间挂起，并在超时后自动回退到第二个提供商。

**Architecture:** 扩展运行时配置以支持主提供商超时与单一 fallback 提供商；在 AI provider 层实现 failover；在服务层降低整阶段重试放大效应。保持 HTTP API 与游戏规则接口不变。

**Tech Stack:** Go 1.25, Hertz, Viper, CloudWeGo Eino OpenAI adapter, Go testing

---
change: fix-ai-provider-next-hang
design-doc: docs/superpowers/specs/2026-06-03-ai-provider-failover-design.md
base-ref: a57fd1af411a95bf8e3b84a9b30b2ca387c67ebb
---

## File Structure

- Modify: `internal/config/config.go`
  扩展 `AIConfig`，支持 `timeout_ms` 与 `fallback` 配置。
- Modify: `internal/config/config_test.go`
  验证新配置字段与环境变量解析行为。
- Create: `internal/infrastructure/ai/failover_provider.go`
  实现主 provider 失败后回退第二 provider 的包装器。
- Modify: `internal/infrastructure/ai/factory.go`
  构建带 timeout 的主 provider 与 fallback provider。
- Modify: `internal/infrastructure/ai/provider.go`
  给 Eino `Generate` 加超时，并把空响应视为错误。
- Modify: `internal/application/service.go`
  降低 `NextPhase` 的整阶段重试放大。
- Modify: `internal/application/service_test.go`
  验证新的退化行为。
- Modify: `openspec/changes/fix-ai-provider-next-hang/tasks.md`
  勾选完成项。

### Task 1: 配置与 Provider Failover

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Create: `internal/infrastructure/ai/failover_provider.go`
- Modify: `internal/infrastructure/ai/factory.go`
- Modify: `internal/infrastructure/ai/provider.go`

- [ ] **Step 1: 先写配置解析失败测试**

```go
func TestLoadReadsAIFallbackConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := `
ai:
  provider: "eino"
  base_url: "http://primary/v1"
  model: "primary-model"
  api_key: "primary-key"
  concurrency: 1
  timeout_ms: 5000
  fallback:
    provider: "fallback"
    timeout_ms: 1000
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.AI.TimeoutMS != 5000 {
		t.Fatalf("TimeoutMS = %d, want 5000", cfg.AI.TimeoutMS)
	}
	if cfg.AI.Fallback == nil || cfg.AI.Fallback.Provider != "fallback" {
		t.Fatalf("fallback provider not loaded: %#v", cfg.AI.Fallback)
	}
}
```

- [ ] **Step 2: 运行配置测试确认失败**

Run: `go test ./internal/config -run TestLoadReadsAIFallbackConfig -count=1`

Expected: `FAIL`，提示 `TimeoutMS` 或 `Fallback` 字段不存在。

- [ ] **Step 3: 扩展配置结构并通过测试**

```go
type AIConfig struct {
	Provider    string    `mapstructure:"provider"`
	BaseURL     string    `mapstructure:"base_url"`
	Model       string    `mapstructure:"model"`
	Concurrency int       `mapstructure:"concurrency"`
	APIKey      string    `mapstructure:"api_key"`
	APIKeyEnv   string    `mapstructure:"api_key_env"`
	TimeoutMS   int       `mapstructure:"timeout_ms"`
	Fallback    *AIConfig `mapstructure:"fallback"`
}
```

- [ ] **Step 4: 为 failover provider 写失败测试**

```go
func TestFailoverProviderFallsBackOnError(t *testing.T) {
	primary := fakeDecisionProvider{err: errors.New("timeout")}
	secondary := fakeDecisionProvider{speech: "后备发言"}
	provider := NewFailoverProvider(primary, secondary)

	msg, err := provider.Speak(domain.Player{ID: 1, Name: "李明"}, domain.DecisionContext{})
	if err != nil {
		t.Fatal(err)
	}
	if msg != "后备发言" {
		t.Fatalf("speech = %q, want fallback output", msg)
	}
}
```

- [ ] **Step 5: 实现最小 failover 与超时控制**

```go
func (p *EinoProvider) generate(instruction string, view domain.DecisionContext) (*schema.Message, error) {
	ctx := context.Background()
	if p.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}
	msg, err := p.model.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(msg.Content) == "" {
		return nil, fmt.Errorf("model returned empty content")
	}
	return msg, nil
}
```

- [ ] **Step 6: 跑配置与 provider 测试确认通过**

Run:
- `go test ./internal/config -count=1`
- `go test ./internal/infrastructure/ai -count=1`

Expected: `PASS`。

- [ ] **Step 7: 提交配置与 failover 层**

```bash
git add internal/config/config.go internal/config/config_test.go internal/infrastructure/ai/failover_provider.go internal/infrastructure/ai/factory.go internal/infrastructure/ai/provider.go
git commit -m "fix: add ai provider timeout and failover"
```

### Task 2: 服务层退化与阶段推进验证

**Files:**
- Modify: `internal/application/service.go`
- Modify: `internal/application/service_test.go`
- Modify: `openspec/changes/fix-ai-provider-next-hang/tasks.md`

- [ ] **Step 1: 写服务层失败测试**

```go
func TestNextPhaseFallsBackWithoutDoubleAmplifyingRetries(t *testing.T) {
	repository := &fakeRepository{state: domain.NewGame()}
	ai := &fakeAI{err: errors.New("timeout"), failures: 99}
	service := NewService(repository, ai)

	state, err := service.NextPhase(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if state.Phase != domain.PhaseNight {
		t.Fatalf("expected fallback to advance phase, got %s", state.Phase)
	}
	if ai.speakCall > len(state.Players)+1 {
		t.Fatalf("unexpected repeated phase amplification, got %d calls", ai.speakCall)
	}
}
```

- [ ] **Step 2: 运行服务测试确认失败**

Run: `go test ./internal/application -run TestNextPhaseFallsBackWithoutDoubleAmplifyingRetries -count=1`

Expected: `FAIL`，因为当前 `maxAIAttempts` 会放大调用次数。

- [ ] **Step 3: 实现最小服务层修复**

```go
const maxAIAttempts = 1
```

保留 `err != nil` 时回退到 `domain.AdvancePhase(state, nil)` 的现有逻辑，不引入新的服务层分支。

- [ ] **Step 4: 跑服务测试与关键链路测试**

Run:
- `go test ./internal/application ./internal/domain -count=1`
- `go test ./... -count=1`

Expected: `PASS`。

- [ ] **Step 5: 手工验证后端链路**

Run:
- `go run ./cmd/server`
- `curl -X POST http://127.0.0.1:8080/api/game/start`
- `curl -X POST http://127.0.0.1:8080/api/game/next`
- `curl http://127.0.0.1:8080/api/game/state`
- `curl http://127.0.0.1:8080/api/game/messages`

Expected:
- `next` 在超时预算内返回
- `state` 推进到下一阶段
- `messages` 包含新增日志或 fallback 输出

- [ ] **Step 6: 勾选 OpenSpec tasks**

```md
- [x] 验证 AI 平台在真实提示词和当前模型配置下是否正常返回内容。
- [x] 复现 `POST /api/game/next` 的长时间挂起，并记录耗时和调用路径证据。
- [x] 在 AI provider / service 层实现最小修复，避免单次外部调用把整次阶段推进拖死。
- [x] 验证 `start`、`state`、`messages`、`next` 的核心链路，确认 hotfix 生效。
```

- [ ] **Step 7: 提交服务层修复与验证状态**

```bash
git add internal/application/service.go internal/application/service_test.go openspec/changes/fix-ai-provider-next-hang/tasks.md
git commit -m "fix: prevent next phase from hanging on ai latency"
```

## Self-Review Checklist

- Spec coverage: timeout、fallback provider、空响应失败、服务层退化、关键链路验证都已映射到任务。
- Placeholder scan: 无 `TBD` / `TODO` / “类似前一任务”。
- Type consistency: 统一使用 `AIConfig.Fallback`、`TimeoutMS`、`FailoverProvider`、`maxAIAttempts = 1` 这些名称。

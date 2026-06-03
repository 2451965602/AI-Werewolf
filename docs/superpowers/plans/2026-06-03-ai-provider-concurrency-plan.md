---
archived-with: 2026-06-03-test-single-provider-concurrency
status: final
---
# AI Provider Concurrency Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让服务按 `ai.provider` 选择真实 provider，并支持 `ai.concurrency=1` 时的全局串行排队。

**Architecture:** 扩展 `internal/config` 读取并校验 `ai.concurrency`；把 provider 构造收敛到 `internal/infrastructure/ai` 的集中工厂；再用通用 `LimitedProvider` 装饰器包装底层 provider 实现并发控制。测试覆盖配置解析、provider 选择与串行化行为。

**Tech Stack:** Go, Viper, Eino, Go testing

---

---
change: test-single-provider-concurrency
design-doc: docs/superpowers/specs/2026-06-03-ai-provider-concurrency-design.md
base-ref: 2f7686c8c071cd93e0b385b6cff24f741833008e
---

### Task 1: 扩展配置模型与校验

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

- [ ] **Step 1: 先写配置测试**

在 `internal/config/config_test.go` 增加以下覆盖：

```go
func TestLoadConfigFileReadsAIConcurrency(t *testing.T) {
	content := `
ai:
	provider: "fallback"
	concurrency: 1
`
	// 断言 cfg.AI.Concurrency == 1
}

func TestLoadRejectsNonPositiveAIConcurrency(t *testing.T) {
	content := `
ai:
	provider: "fallback"
	concurrency: 0
`
	// 断言 LoadFromPath 返回错误
}
```

- [ ] **Step 2: 运行配置测试确认先失败**

Run: `go test ./internal/config -run TestLoad -timeout 60s`

Expected: 因 `AIConfig` 尚无 `Concurrency` 字段或缺少校验而失败。

- [ ] **Step 3: 最小实现配置字段与校验**

在 `internal/config/config.go` 中：

```go
type AIConfig struct {
	Provider    string `mapstructure:"provider"`
	BaseURL     string `mapstructure:"base_url"`
	Model       string `mapstructure:"model"`
	APIKey      string `mapstructure:"api_key"`
	APIKeyEnv   string `mapstructure:"api_key_env"`
	Concurrency int    `mapstructure:"concurrency"`
}
```

并补充：

```go
v.SetDefault("ai.concurrency", 1)
```

读取后做校验：

```go
if cfg.AI.Concurrency <= 0 {
	return Config{}, fmt.Errorf("ai.concurrency must be greater than 0")
}
```

- [ ] **Step 4: 重新运行配置测试**

Run: `go test ./internal/config -timeout 60s`

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: add ai concurrency config"
```

### Task 2: 集中 provider 构造逻辑

**Files:**
- Modify: `internal/infrastructure/ai/provider.go`
- Create: `internal/infrastructure/ai/factory.go`
- Modify: `cmd/server/main.go`
- Create: `internal/infrastructure/ai/factory_test.go`

- [ ] **Step 1: 先写 provider 选择测试**

在 `internal/infrastructure/ai/factory_test.go` 增加：

```go
func TestBuildProviderReturnsFallbackProvider(t *testing.T) {
	provider, err := BuildProvider(config.AIConfig{Provider: "fallback", Concurrency: 1})
	if err != nil { t.Fatalf("BuildProvider() error = %v", err) }
	if provider == nil { t.Fatal("provider is nil") }
}

func TestBuildProviderRejectsUnknownProvider(t *testing.T) {
	_, err := BuildProvider(config.AIConfig{Provider: "unknown", Concurrency: 1})
	if err == nil { t.Fatal("expected error for unknown provider") }
}
```

- [ ] **Step 2: 运行 provider 测试确认先失败**

Run: `go test ./internal/infrastructure/ai -run TestBuildProvider -timeout 60s`

Expected: 因 `BuildProvider` 不存在而失败。

- [ ] **Step 3: 实现集中构造函数**

新增 `internal/infrastructure/ai/factory.go`，提供类似结构：

```go
func BuildProvider(cfg config.AIConfig) (domain.DecisionProvider, error) {
	var base domain.DecisionProvider
	switch cfg.Provider {
	case "fallback":
		base = FallbackProvider{}
	default:
		return nil, fmt.Errorf("unsupported ai.provider: %s", cfg.Provider)
	}
	return WrapWithConcurrencyLimit(base, cfg.Concurrency), nil
}
```

同时把 `cmd/server/main.go` 改成：

```go
aiProvider, err := ai.BuildProvider(cfg.AI)
if err != nil {
	log.Fatalf("build ai provider: %v", err)
}
```

- [ ] **Step 4: 重新运行 provider 测试**

Run: `go test ./internal/infrastructure/ai -run TestBuildProvider -timeout 60s`

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/infrastructure/ai/provider.go internal/infrastructure/ai/factory.go internal/infrastructure/ai/factory_test.go cmd/server/main.go
git commit -m "feat: build ai provider from config"
```

### Task 3: 实现通用限流装饰器

**Files:**
- Create: `internal/infrastructure/ai/limited_provider.go`
- Create: `internal/infrastructure/ai/limited_provider_test.go`

- [ ] **Step 1: 先写并发串行化测试**

在 `internal/infrastructure/ai/limited_provider_test.go` 中创建一个可观测 mock provider：

```go
type stubProvider struct {
	mu      sync.Mutex
	active  int
	maxSeen int
}

func (s *stubProvider) Speak(player domain.Player, ctx domain.DecisionContext) (string, error) {
	s.mu.Lock()
	s.active++
	if s.active > s.maxSeen { s.maxSeen = s.active }
	s.mu.Unlock()
	time.Sleep(50 * time.Millisecond)
	s.mu.Lock()
	s.active--
	s.mu.Unlock()
	return "ok", nil
}
```

核心断言：

```go
func TestLimitedProviderSerializesCallsWhenConcurrencyIsOne(t *testing.T) {
	base := &stubProvider{}
	provider := WrapWithConcurrencyLimit(base, 1)
	// 并发启动多个 Speak 调用
	// 断言 base.maxSeen == 1
}
```

- [ ] **Step 2: 运行串行化测试确认先失败**

Run: `go test ./internal/infrastructure/ai -run TestLimitedProviderSerializesCallsWhenConcurrencyIsOne -timeout 60s`

Expected: 因 `WrapWithConcurrencyLimit` 不存在而失败。

- [ ] **Step 3: 实现限流装饰器**

新增 `internal/infrastructure/ai/limited_provider.go`：

```go
type LimitedProvider struct {
	base   domain.DecisionProvider
	tokens chan struct{}
}

func WrapWithConcurrencyLimit(base domain.DecisionProvider, concurrency int) domain.DecisionProvider {
	if concurrency <= 1 {
		return &LimitedProvider{base: base, tokens: make(chan struct{}, 1)}
	}
	return &LimitedProvider{base: base, tokens: make(chan struct{}, concurrency)}
}
```

并在每个方法中包装调用：

```go
func (p *LimitedProvider) withPermit(fn func() error) error {
	p.tokens <- struct{}{}
	defer func() { <-p.tokens }()
	return fn()
}
```

每个 provider 方法通过 `withPermit` 调底层 `base`。

- [ ] **Step 4: 重新运行限流测试**

Run: `go test ./internal/infrastructure/ai -run TestLimitedProvider -timeout 60s`

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/infrastructure/ai/limited_provider.go internal/infrastructure/ai/limited_provider_test.go
git commit -m "feat: serialize ai provider calls"
```

### Task 4: 全量回归并同步任务状态

**Files:**
- Modify: `openspec/changes/test-single-provider-concurrency/tasks.md`

- [ ] **Step 1: 运行目标包测试**

Run:

```bash
go test ./internal/config ./internal/infrastructure/ai -timeout 60s
```

Expected: PASS

- [ ] **Step 2: 运行全量回归测试**

Run:

```bash
go test ./... -timeout 60s
```

Expected: PASS

- [ ] **Step 3: 勾选 tasks.md 全部任务**

将：

```markdown
- [ ] 1. 扩展运行时配置...
```

更新为：

```markdown
- [x] 1. 扩展运行时配置...
```

其余任务同理全部勾选。

- [ ] **Step 4: 提交最终实现**

```bash
git add openspec/changes/test-single-provider-concurrency/tasks.md
git commit -m "test: cover ai provider concurrency flow"
```

## Self-Review Checklist

- `runtime-configuration` delta spec 中的 provider 选择与 concurrency 场景都有任务覆盖。
- 计划中的所有改动都集中在配置层、provider 工厂、限流装饰器和对应测试，没有引入不必要的接口改造。
- 全量验证命令固定为 `go test ./... -timeout 60s`。

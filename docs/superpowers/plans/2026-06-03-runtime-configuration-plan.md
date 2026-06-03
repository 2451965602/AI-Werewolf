---
change: configure-env-file-priority
design-doc: docs/superpowers/specs/2026-06-03-runtime-configuration-design.md
base-ref: be6594ffbf971f31c91d01ef8824988677f7f55b
archived-with: 2026-06-03-configure-env-file-priority
---

# Runtime Configuration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add runtime configuration with priority `environment variables > config file > defaults`.

**Architecture:** Encapsulate Viper inside `internal/config`, returning a typed `Config` to startup code. `cmd/server` injects resolved values into infrastructure components; HTTP API behavior stays unchanged.

**Tech Stack:** Go 1.22, Viper, Hertz, existing JSON store, `go test ./... -timeout 60s`.

archived-with: 2026-06-03-configure-env-file-priority
---

## File Structure

- Create `internal/config/config.go`: typed config structs, defaults, Viper loader, validation, AI key resolution.
- Create `internal/config/config_test.go`: tests for defaults, file override, env override, missing config, invalid YAML, AI key resolution.
- Modify `cmd/server/main.go`: load config on startup and inject resolved storage path and server addr.
- Modify `internal/transport/http/router.go`: allow the router to receive a listen address.
- Modify `internal/transport/http/router_test.go`: update constructor usage if required.
- Modify `config.example.yaml`: add `ai.api_key` example with warning comment.
- Modify `go.mod` / `go.sum`: add Viper dependency via `go get github.com/spf13/viper`.
- Modify `openspec/changes/configure-env-file-priority/tasks.md`: check tasks as completed.

## Task 1: Configuration Package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Add Viper dependency**

Run:

```bash
go get github.com/spf13/viper@latest
```

Expected: `go.mod` includes `github.com/spf13/viper`; `go.sum` is updated.

- [ ] **Step 2: Write failing config tests**

Create `internal/config/config_test.go` with tests for:

```go
func TestLoadDefaultsWhenDefaultConfigMissing(t *testing.T)
func TestLoadConfigFileOverridesDefaults(t *testing.T)
func TestEnvironmentOverridesConfigFile(t *testing.T)
func TestExplicitConfigPathMissingReturnsError(t *testing.T)
func TestInvalidYAMLReturnsError(t *testing.T)
func TestAIAPIKeyResolutionOrder(t *testing.T)
```

Use `t.TempDir()` for config files and `t.Setenv()` for env isolation. Expected initial result: tests fail because `Load` and config types do not exist.

- [ ] **Step 3: Implement typed config loader**

Create `internal/config/config.go` with these exported shapes:

```go
type Config struct {
    Server  ServerConfig
    Storage StorageConfig
    AI      AIConfig
}

type ServerConfig struct {
    Addr string `mapstructure:"addr"`
}

type StorageConfig struct {
    StatePath string `mapstructure:"state_path"`
}

type AIConfig struct {
    Provider  string `mapstructure:"provider"`
    BaseURL   string `mapstructure:"base_url"`
    Model     string `mapstructure:"model"`
    APIKey    string `mapstructure:"api_key"`
    APIKeyEnv string `mapstructure:"api_key_env"`
}
```

Implement:

```go
func Load() (Config, error)
func LoadFromPath(path string) (Config, error)
```

Rules:

- Defaults: `server.addr=:8080`, `storage.state_path=data/world_state.json`, `ai.provider=fallback`, `ai.api_key_env=OPENAI_API_KEY`.
- `Load()` reads `CONFIG_PATH`; if empty, attempts default `config.yaml`.
- Missing default `config.yaml` is allowed.
- Missing explicit `CONFIG_PATH` / `LoadFromPath(path)` file is an error.
- Configure Viper with `SetEnvPrefix("WEREWOLF")`, `SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))`, and `AutomaticEnv()`.
- Resolve API key as `WEREWOLF_AI_API_KEY > ai.api_key > os.Getenv(ai.api_key_env)`.
- Return an error if `server.addr` or `storage.state_path` is empty.

- [ ] **Step 4: Run package tests**

Run:

```bash
go test ./internal/config -timeout 60s
```

Expected: PASS.

- [ ] **Step 5: Commit Task 1**

```bash
git add go.mod go.sum internal/config/config.go internal/config/config_test.go
git commit -m "feat: add runtime configuration loader"
```

## Task 2: Startup and Router Integration

**Files:**
- Modify: `cmd/server/main.go`
- Modify: `internal/transport/http/router.go`
- Modify: `internal/transport/http/router_test.go`
- Modify: `config.example.yaml`

- [ ] **Step 1: Update router constructor test expectations**

Update router tests to call the new constructor shape. If tests only need routes, use:

```go
router := NewRouter(service, ":0")
```

Expected initial result: compile fails until router constructor is updated.

- [ ] **Step 2: Update router constructor**

Change `NewRouter` to accept an address and construct Hertz with it:

```go
func NewRouter(service GameService, addr string) *server.Hertz {
    h := server.Default(server.WithHostPorts(addr))
    // existing routes stay unchanged
    return h
}
```

- [ ] **Step 3: Update server startup**

Change `cmd/server/main.go` to load config and inject values:

```go
cfg, err := config.Load()
if err != nil {
    log.Fatalf("load config: %v", err)
}

repository := store.NewJSONStore(cfg.Storage.StatePath)
aiProvider := ai.FallbackProvider{}
service := application.NewService(repository, aiProvider)
router := transporthttp.NewRouter(service, cfg.Server.Addr)
router.Spin()
```

- [ ] **Step 4: Update config example**

Add `ai.api_key` to `config.example.yaml` with a comment that environment variables are recommended for production secrets.

- [ ] **Step 5: Run affected tests**

Run:

```bash
go test ./internal/transport/http ./cmd/server -timeout 60s
```

Expected: PASS or `cmd/server` reports no test files while compiling successfully.

- [ ] **Step 6: Commit Task 2**

```bash
git add cmd/server/main.go internal/transport/http/router.go internal/transport/http/router_test.go config.example.yaml
git commit -m "feat: wire runtime configuration into server startup"
```

## Task 3: Full Verification and OpenSpec Task Sync


**Files:**
- Modify: `openspec/changes/configure-env-file-priority/tasks.md`

- [ ] **Step 1: Format code**

Run:

```bash
go fmt ./...
```

Expected: command exits 0.

- [ ] **Step 2: Run all tests**

Run:

```bash
go test ./... -timeout 60s
```

Expected: PASS.

- [ ] **Step 3: Mark OpenSpec tasks complete**

Update `openspec/changes/configure-env-file-priority/tasks.md` so all three items use `- [x]`.

- [ ] **Step 4: Commit Task 3**

```bash
git add openspec/changes/configure-env-file-priority/tasks.md
git commit -m "chore: complete runtime configuration tasks"
```

## Self-Review

- Spec coverage: priority order, server addr, state path, AI key direct/config/env-name resolution are covered by Tasks 1 and 2.
- Placeholder scan: no placeholder steps remain.
- Type consistency: plan uses `Config`, `ServerConfig`, `StorageConfig`, `AIConfig`, `Load`, and `LoadFromPath` consistently.

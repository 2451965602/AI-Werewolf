---
change: implement-hertz-eino-werewolf
design-doc: docs/superpowers/specs/2026-06-01-implement-hertz-eino-werewolf-design.md
base-ref: none-not-git-repository
archived-with: 2026-06-01-implement-hertz-eino-werewolf
---

# Implement Hertz Eino Werewolf Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go backend for AI Werewolf using Hertz for HTTP APIs and Eino-style AI decision boundaries.

**Architecture:** Use a deterministic domain engine behind application services and Hertz handlers. AI and JSON storage are infrastructure adapters behind interfaces, so rules remain testable without external services.

**Tech Stack:** Go 1.22+, Hertz, Eino-compatible AI adapter boundary, JSON file storage, Go standard testing.

archived-with: 2026-06-01-implement-hertz-eino-werewolf
---

## File Structure

- Create: `go.mod` — module and dependencies.
- Create: `cmd/server/main.go` — process entrypoint and server bootstrap.
- Create: `internal/domain/types.go` — core enums and data structures.
- Create: `internal/domain/engine.go` — deterministic game state machine.
- Create: `internal/domain/engine_test.go` — domain rule tests.
- Create: `internal/application/service.go` — use-case orchestration.
- Create: `internal/application/service_test.go` — fake AI/store application tests.
- Create: `internal/infrastructure/ai/provider.go` — AI decision provider interface implementation boundary and fallback provider.
- Create: `internal/infrastructure/store/json_store.go` — atomic JSON state repository.
- Create: `internal/infrastructure/store/json_store_test.go` — persistence tests.
- Create: `internal/transport/http/router.go` — Hertz routes and handlers.
- Create: `internal/transport/http/router_test.go` — endpoint contract tests.
- Create: `config.example.yaml` — runtime configuration example.
- Modify: `README.md` — startup and API usage notes if README exists; otherwise create it.
- Modify: `openspec/changes/implement-hertz-eino-werewolf/tasks.md` — check off completed OpenSpec tasks as implementation progresses.

## Task 1: Initialize Go module and package skeleton

**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `internal/domain/types.go`

- [ ] **Step 1: Create module metadata**

```go
module ai-werewolf-go

go 1.22

require github.com/cloudwego/hertz v0.9.6
```

- [ ] **Step 2: Add initial domain types**

```go
package domain

type Phase string
const (
    PhaseDay Phase = "day"
    PhaseNight Phase = "night"
    PhaseEnded Phase = "ended"
)

type Role string
const (
    RoleWerewolf Role = "werewolf"
    RoleVillager Role = "villager"
    RoleSeer Role = "seer"
    RoleWitch Role = "witch"
    RoleHunter Role = "hunter"
)

type Player struct {
    ID int `json:"id"`
    Name string `json:"name"`
    Role Role `json:"role"`
    Alive bool `json:"alive"`
    Team string `json:"team"`
}

type Message struct {
    SpeakerID int `json:"speakerId"`
    Speaker string `json:"speaker"`
    Content string `json:"content"`
    Phase Phase `json:"phase"`
    Round int `json:"round"`
    Type string `json:"type"`
}

type GameState struct {
    Round int `json:"round"`
    Phase Phase `json:"phase"`
    Ended bool `json:"ended"`
    Winner string `json:"winner,omitempty"`
    Players []Player `json:"players"`
    Messages []Message `json:"messages"`
    LastNightKilled int `json:"lastNightKilled,omitempty"`
}
```

- [ ] **Step 3: Add server bootstrap**

```go
package main

func main() {
    // Router wiring is added after transport and application services exist.
}
```

- [ ] **Step 4: Verify module loads**

Run: `go test ./...`
Expected: packages compile or report no tests for skeleton packages.

## Task 2: Implement deterministic domain engine with tests

**Files:**
- Create: `internal/domain/engine.go`
- Create: `internal/domain/engine_test.go`

- [ ] **Step 1: Write failing domain tests**

```go
func TestStartGameEntersDayOne(t *testing.T) {
    state := NewGame()
    if state.Round != 1 || state.Phase != PhaseDay {
        t.Fatalf("expected day 1, got round=%d phase=%s", state.Round, state.Phase)
    }
}

func TestDayOneDoesNotVote(t *testing.T) {
    state := NewGame()
    provider := StaticDecisionProvider{}
    next, err := AdvancePhase(state, provider)
    if err != nil { t.Fatal(err) }
    if next.Ended { t.Fatal("game should not end on day one") }
    for _, msg := range next.Messages {
        if msg.Type == "vote" { t.Fatal("day one must not vote") }
    }
}
```

- [ ] **Step 2: Implement `NewGame` and `AdvancePhase`**

Create deterministic role assignment for 10 players and enforce day-one discussion-only behavior.

- [ ] **Step 3: Add win-condition test**

```go
func TestWolvesEliminatedEndsImmediately(t *testing.T) {
    state := NewGame()
    for i := range state.Players {
        if state.Players[i].Role == RoleWerewolf { state.Players[i].Alive = false }
    }
    ended := CheckGameEnd(&state)
    if !ended || !state.Ended || state.Winner != "village" {
        t.Fatalf("expected village win, got ended=%v winner=%q", state.Ended, state.Winner)
    }
}
```

- [ ] **Step 4: Run domain tests**

Run: `go test ./internal/domain -v`
Expected: all domain tests pass.

## Task 3: Implement application service and AI boundary

**Files:**
- Create: `internal/application/service.go`
- Create: `internal/application/service_test.go`
- Create: `internal/infrastructure/ai/provider.go`

- [ ] **Step 1: Define interfaces**

```go
type AIDecisionProvider interface {
    Speak(player domain.Player, state domain.GameState) (string, error)
    Vote(player domain.Player, state domain.GameState) (int, error)
    WerewolfTarget(player domain.Player, state domain.GameState) (int, error)
}

type StateRepository interface {
    Save(context.Context, domain.GameState) error
    Load(context.Context) (domain.GameState, error)
}
```

- [ ] **Step 2: Implement service methods**

Add `StartGame`, `NextPhase`, `GetState`, and `GetMessages`. Each successful mutation saves state through `StateRepository`.

- [ ] **Step 3: Test save behavior and AI fallback path**

Use fake repository and fake AI provider that can return errors. Verify errors do not silently disappear and legal fallback is used.

- [ ] **Step 4: Run application tests**

Run: `go test ./internal/application -v`
Expected: all application tests pass.

## Task 4: Implement JSON state repository

**Files:**
- Create: `internal/infrastructure/store/json_store.go`
- Create: `internal/infrastructure/store/json_store_test.go`

- [ ] **Step 1: Write persistence tests**

Test save/load round trip, invalid JSON error, and failed write preserving previous state.

- [ ] **Step 2: Implement atomic write**

Use temp file in the same directory, encode JSON, close, then rename to target path.

- [ ] **Step 3: Run store tests**

Run: `go test ./internal/infrastructure/store -v`
Expected: all storage tests pass.

## Task 5: Implement Hertz HTTP transport

**Files:**
- Create: `internal/transport/http/router.go`
- Create: `internal/transport/http/router_test.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Add route contract tests**

Test `POST /api/game/start`, `POST /api/game/next`, `GET /api/game/state`, `GET /api/game/messages`, and `GET /api/game/health`.

- [ ] **Step 2: Implement handlers**

Handlers call application service only. They must not contain game rules.

- [ ] **Step 3: Wire server entrypoint**

Create repository, AI provider, application service, and Hertz router in `cmd/server/main.go`.

- [ ] **Step 4: Run transport tests**

Run: `go test ./internal/transport/http -v`
Expected: endpoint tests pass.

## Task 6: Configuration and documentation

**Files:**
- Create: `config.example.yaml`
- Create or modify: `README.md`
- Modify: `openspec/changes/implement-hertz-eino-werewolf/tasks.md`

- [ ] **Step 1: Add config example**

```yaml
server:
  addr: ":8080"
storage:
  state_path: "data/world_state.json"
ai:
  provider: "openai-compatible"
  base_url: "https://api.example.com/v1"
  model: "deepseek-chat"
```

- [ ] **Step 2: Document startup and APIs**

README must include:
- `go mod tidy`
- `go test ./...`
- `go run ./cmd/server`
- API endpoint list and example curl commands

- [ ] **Step 3: Check off OpenSpec tasks**

Update `openspec/changes/implement-hertz-eino-werewolf/tasks.md` as each implemented area completes.

- [ ] **Step 4: Run full verification**

Run: `go test ./...`
Expected: all tests pass.

## Self-Review

- Spec coverage: game engine, AI decision, Hertz API, and state storage are each mapped to tasks and tests.
- Placeholder scan: no TBD/TODO/fill-in-later instructions remain.
- Type consistency: package names and interfaces align across domain, application, infrastructure, and transport tasks.

---
comet_change: implement-hertz-eino-werewolf
role: technical-design
canonical_spec: openspec
---

# Implement Hertz Eino Werewolf — Technical Design

## Context

This design implements the OpenSpec change `implement-hertz-eino-werewolf`.
OpenSpec remains the canonical source for requirements. This document defines HOW the Go implementation should be structured, tested, and evolved.

The goal is a Go backend equivalent to the reference `AI-Werewolf` project, using Hertz for HTTP transport and Eino for AI decision chains.

## Recommended Approach

Use a single-process deterministic game engine with clear package boundaries:

```text
cmd/server
  -> internal/transport/http      Hertz routes and handlers
  -> internal/application         use-case orchestration
  -> internal/domain              game state, rules, phase engine
  -> internal/infrastructure/ai   Eino-backed AI decision provider
  -> internal/infrastructure/store JSON state repository
```

This keeps the first implementation small enough to verify while preserving extension points for future storage, streaming, and multi-room support.

## Architecture Decisions

### Rules before AI

AI is an advisor, not the authority. The domain layer owns all legality checks.

Examples:
- Werewolves cannot kill wolf teammates.
- The seer cannot inspect invalid targets.
- The witch cannot rescue anyone except the last-night victim.
- Day one cannot produce exile voting.

The AI provider returns candidate actions; the engine validates candidates before state mutation. Model invocation or parse failure uses bounded retry followed by deterministic heuristic fallback; a successful model response with an illegal target is rejected without mutating game state.

### Synchronous phase advancement

`POST /api/game/next` performs the full current phase transition synchronously.

This is intentional for the MVP:
- predictable tests
- no background worker lifecycle
- no partial phase state exposed to clients

If AI latency becomes a UX issue later, the same application service can be wrapped with async jobs or streaming events without changing domain rules.

### JSON storage for phase one

State persistence uses a JSON repository.

Write flow:
1. encode full `GameState`
2. write to a temporary file
3. atomically replace the target file

This prevents partial writes from corrupting the last valid save.

## Core Components

### Domain

Responsibilities:
- define `GameState`, `Player`, `Message`, `Vote`, phase and role types
- initialize the 10-player default game
- progress day/night phases
- apply role actions
- check win conditions after every actionable step

Domain MUST not import Hertz or Eino.

### Application

Responsibilities:
- expose `StartGame`, `NextPhase`, `GetState`, `GetMessages`
- coordinate engine, AI provider, and state repository
- convert domain errors into transport-neutral application errors

### Transport

Responsibilities:
- bind Hertz routes
- parse input and serialize output
- map application errors to HTTP status and stable error response bodies

No game rules should live in handlers.

### AI Infrastructure

Responsibilities:
- implement `AIDecisionProvider`
- build role-scoped prompts
- call Eino chains
- parse structured model output
- retry bounded failures

The parser should prefer structured JSON-like output. Free text can be used for speech, but actions should have explicit target IDs and reasons.

### Storage Infrastructure

Responsibilities:
- save and load `GameState`
- handle missing, invalid, and write-failure cases explicitly
- preserve previous valid state on failed write

## Key Data Flow

```text
POST /api/game/start
  -> handler
  -> application.StartGame
  -> domain.InitializeGame
  -> repository.Save
  -> response GameState

POST /api/game/next
  -> handler
  -> application.NextPhase
  -> domain.AdvancePhase(aiProvider)
       -> AI candidate decision
       -> domain validation
       -> state mutation
       -> terminal check
  -> repository.Save
  -> response GameState
```

## Error Handling

- Invalid state transition: return a domain/application error and do not save mutated state.
- AI call failure: retry, then use legal heuristic fallback.
- AI parse failure: retry, then use legal heuristic fallback.
- Illegal target in a successful model response: reject that action without mutating game state.
- Storage write failure: return infrastructure error and keep previous persisted state intact.
- Storage decode failure: return infrastructure error and do not overwrite in-memory state with partial data.

## Testing Strategy

### Domain tests

Required scenarios:
- game start enters day one
- day one has introductions/discussion but no exile vote
- wolves eliminated immediately ends the game
- witch rescue only applies to last-night killed player
- invalid AI target does not mutate state

### Application tests

Use fake AI and fake repository implementations.

Required scenarios:
- start saves initialized game
- next phase saves updated game
- AI failure falls back to a legal action
- repository failure propagates a visible error

### Transport tests

Use Hertz test utilities or HTTP-level tests.

Required scenarios:
- `POST /api/game/start` returns game state
- `POST /api/game/next` advances state
- `GET /api/game/state` returns current state
- invalid transition returns stable error response

## Risks

### LLM instability

Model outputs can be invalid or inconsistent.

Mitigation:
- structured action output
- strict target validation
- bounded retries
- deterministic fallback

### Long synchronous requests

Multiple AI calls in one phase can increase response time.

Mitigation:
- keep phase one synchronous for correctness
- add logging and timing metrics
- defer async/streaming to later change if needed

### State file corruption

Direct writes can corrupt JSON state on interruption.

Mitigation:
- temp-file write and atomic replace
- explicit decode errors
- tests for failed writes

## Non-Goals

- No full frontend rewrite.
- No multi-room or distributed game engine.
- No account system.
- No production database in this phase.

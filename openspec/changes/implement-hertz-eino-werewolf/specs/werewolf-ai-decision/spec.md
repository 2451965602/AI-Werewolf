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

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

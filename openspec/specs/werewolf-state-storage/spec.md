## ADDED Requirements

### Requirement: JSON game state persistence
The system MUST persist the current game state to JSON after initialization and after each successful phase transition.

#### Scenario: Save after game start
- **WHEN** a new game is started successfully
- **THEN** the initialized game state MUST be saved to the configured JSON storage path

#### Scenario: Save after phase advance
- **WHEN** a phase transition completes successfully
- **THEN** the updated game state MUST be saved to the configured JSON storage path

### Requirement: Game state loading
The system MUST load an existing game state from JSON storage when requested and MUST report missing or invalid files explicitly.

#### Scenario: Load existing state
- **WHEN** persisted state exists and is valid JSON
- **THEN** the system MUST restore the game state from storage

#### Scenario: Invalid state file
- **WHEN** persisted state exists but cannot be decoded
- **THEN** the system MUST return a storage error and MUST not replace current in-memory state with partial data

### Requirement: Atomic state writes
The system MUST avoid corrupting the existing state file when a write fails.

#### Scenario: Failed write preserves previous state
- **WHEN** writing the next state fails
- **THEN** the previous persisted state MUST remain readable

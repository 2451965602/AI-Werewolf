## ADDED Requirements

### Requirement: Game lifecycle API
The system MUST expose Hertz HTTP endpoints for starting a game, advancing the game phase, and reading game state.

#### Scenario: Start game endpoint
- **WHEN** a client sends `POST /api/game/start`
- **THEN** the system MUST initialize a new game and return the current game state

#### Scenario: Advance phase endpoint
- **WHEN** a client sends `POST /api/game/next` for a non-ended game
- **THEN** the system MUST advance the game through the domain engine and return the updated game state

### Requirement: Message and health API
The system MUST expose endpoints for reading game messages and checking service health.

#### Scenario: Read messages endpoint
- **WHEN** a client sends `GET /api/game/messages`
- **THEN** the system MUST return the ordered message history for the current game

#### Scenario: Health endpoint
- **WHEN** a client sends `GET /api/game/health`
- **THEN** the system MUST return a successful health response without mutating game state

### Requirement: Consistent error responses
The system MUST return consistent error response structures for invalid requests, invalid state transitions, and infrastructure failures.

#### Scenario: Invalid transition response
- **WHEN** a client requests a phase transition that violates domain rules
- **THEN** the API MUST return a non-success status and a diagnostic error body

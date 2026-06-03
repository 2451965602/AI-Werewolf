## MODIFIED Requirements

### Requirement: AI runtime configuration
The system MUST support configuring AI provider runtime behavior, including request timeout and a single fallback provider.

#### Scenario: Primary provider timeout configuration
- **WHEN** the runtime loads AI configuration
- **THEN** it MUST accept a primary provider timeout setting
- **AND** the AI provider MUST apply that timeout to outbound model requests

#### Scenario: Fallback provider configuration
- **WHEN** the runtime loads AI configuration
- **THEN** it MUST allow configuring one fallback AI provider with its own provider type, endpoint, model, credentials, and timeout
